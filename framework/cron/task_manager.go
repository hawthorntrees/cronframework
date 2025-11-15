package cron

import (
	"context"
	"errors"
	"fmt"
	"github.com/hawthorntrees/cronframework/framework/config"
	"github.com/hawthorntrees/cronframework/framework/utils"
	"gorm.io/gorm"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/hawthorntrees/cronframework/framework/model"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type TaskFunc func(ctx context.Context, params TaskParams) error

type TaskParams struct {
	TaskID     int64
	TaskName   string
	RetryCount int
}

type TaskManager struct {
	nodeID       string
	cron         *cron.Cron
	taskFuncs    map[string]TaskFunc
	taskEntries  map[int64]cronEntryInfo // 任务ID -> 定时任务信息
	repo         *Repository
	syncInterval time.Duration
	logger       *zap.SugaredLogger
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

type cronEntryInfo struct {
	EntryID  cron.EntryID
	CronExpr string
	TimeOut  int
}

var noRecordExecution bool

func NewTaskManager(taskCfg *config.TaskConfig) *TaskManager {
	initLogger(taskCfg.LogLevel)
	noRecordExecution = taskCfg.NotRecordTaskExecution
	lg := taskLogger.With(zap.String("traceID", "task-manager")).Sugar()
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		nodeID:       taskCfg.NodeID,
		cron:         cron.New(cron.WithSeconds()),
		taskFuncs:    make(map[string]TaskFunc),
		taskEntries:  make(map[int64]cronEntryInfo),
		repo:         NewRepository(),
		syncInterval: taskCfg.TaskSyncInterval,
		logger:       lg,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (m *TaskManager) RegisterTask(fn TaskFunc) error {
	pc := reflect.ValueOf(fn).Pointer()
	name := runtime.FuncForPC(pc).Name()

	if _, exists := m.taskFuncs[name]; exists {
		return fmt.Errorf("任务 [%s] 已注册", name)
	}

	m.taskFuncs[name] = fn
	m.logger.Debugf("注册任务成功:%s", name)
	return nil
}

func (m *TaskManager) Start() error {

	if err := m.syncTasks(); err != nil {
		m.logger.Errorf("初始同步任务失败:%v", err)
		return err
	}
	m.cron.Start()

	go m.startSyncLoop()
	m.logger.Debug("任务管理器启动成功")
	return nil
}

func (m *TaskManager) Stop() {
	m.cancel()
	m.cron.Stop()
	m.logger.Debug("任务管理器已停止")
}

func (m *TaskManager) startSyncLoop() {
	ticker := time.NewTicker(m.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.syncTasks(); err != nil {
				m.logger.Errorf("同步任务失败:%v", err)
			}
		}
	}
}

func (m *TaskManager) syncTasks() error {
	m.logger.Debug("开始同步任务配置")
	tasks, err := m.repo.GetEnabledTasks(m.ctx)
	if err != nil {
		return fmt.Errorf("获取任务列表失败: %w", err)
	}

	keepTaskIDs := make(map[int64]bool)

	for _, task := range tasks {
		keepTaskIDs[task.ID] = true

		if entryInfo, exists := m.taskEntries[task.ID]; exists {
			if entryInfo.CronExpr == task.CronExpr && entryInfo.TimeOut == task.Timeout {
				continue
			}
			m.cron.Remove(entryInfo.EntryID)
			m.logger.Debugf("移除变更的任务:%s", task.Name)
		}
		if _, exists := m.taskFuncs[task.Name]; !exists {
			m.logger.Warnf("任务函数[%s]未注册", task.Name)
			continue
		}

		entryID, err := m.addTaskToCron(task)
		if err != nil {
			m.logger.Errorf("添加任务到调度器失败[%v-%v:%v]", task.ID, task.Name, err)
			continue
		}
		m.taskEntries[task.ID] = cronEntryInfo{
			EntryID:  entryID,
			CronExpr: task.CronExpr,
			TimeOut:  task.Timeout,
		}

		m.logger.Debugf("添加/更新任务调度[%v-%v-%v]", task.ID, task.Name, task.CronExpr)
	}

	for taskID, entryInfo := range m.taskEntries {
		if !keepTaskIDs[taskID] {
			m.cron.Remove(entryInfo.EntryID)
			delete(m.taskEntries, taskID)
			m.logger.Debugf("移除已删除的任务[%v]", taskID)
		}
	}

	m.logger.Debug("任务配置同步完成")
	return nil
}

func (m *TaskManager) addTaskToCron(task *model.Hawthorn_task) (cron.EntryID, error) {
	job := func() {
		defer func() {
			err := recover()
			if err != nil {
				m.logger.Error("任务执行框架异常：", zap.Error(fmt.Errorf("%v", err)))
			}
		}()
		m.executeTask(task)
	}

	return m.cron.AddFunc(task.CronExpr, job)
}

var (
	stateFiled     = "failed"
	stateSuccess   = "success"
	lockTaskFailed = "任务抢占失败"
	noFunc         = "任务函数未注册"
)

func createContext() (traceID string, ctx context.Context, lg *zap.Logger) {
	id, err := utils.GenerateTraceID()
	if err != nil {
		id = "traceErr"
	}
	c := context.WithValue(context.Background(), "traceID", id)
	l := taskLogger.With(zap.String("traceID", id))
	return id, c, l
}
func (m *TaskManager) executeTask(task *model.Hawthorn_task) {
	traceID, ctx, lg := createContext()
	nw := time.Now()
	now := nw.Truncate(time.Millisecond)
	expiredAt := nw.Add(time.Duration(task.Timeout) * time.Second).Truncate(time.Millisecond)
	execution := &model.Hawthorn_task_execution{
		TaskID:      task.ID,
		NodeID:      m.nodeID,
		TraceID:     traceID,
		StartTime:   now,
		CreatedDate: now,
	}
	var finalErr error

	defer func() {
		err := recover()
		if err != nil {
			lg.Sugar().Errorw("未知异常:%v", err)
			execution.Error = fmt.Sprintf("未知异常：%v", err)
			execution.Status = stateFiled
		} else if finalErr != nil {
			lg.Sugar().Errorw("任务失败:%v", finalErr)
			execution.Error = fmt.Sprintf("任务失败：%v", finalErr)
			execution.Status = stateFiled
		} else if execution.Status == "" {
			return
		}

		err2 := m.repo.ReleaseLockTask(ctx, task.ID, now, expiredAt, lg)
		if err2 != nil {
			execution.Error = execution.Error + "释放锁失败"
			lg.Sugar().Errorw("释放锁失败：%d,%w", task.ID, err2)
		}
		if !noRecordExecution {
			end := time.Now().Truncate(time.Millisecond)
			execution.EndTime = &end
			if err := m.repo.CreateExecution(ctx, execution); err != nil {
				panic(fmt.Errorf("登记执行记录失败: %v", err))
			}
		}
	}()

	lockErr := m.repo.TryLockTask(ctx, task.ID, now, expiredAt)
	if lockErr != nil {
		if errors.Is(lockErr, gorm.ErrRecordNotFound) {
			return
		}
		execution.Status = stateFiled
		execution.Error = lockTaskFailed
		lg.Sugar().Errorw("任务%d-%s抢占失败:%w", task.ID, task.Name, lockErr)
		return
	}
	m.mu.RLock()
	taskFunc, exists := m.taskFuncs[task.Name]
	m.mu.RUnlock()

	if !exists {
		execution.Status = stateFiled
		execution.Error = noFunc
		return
	}

	for i := 0; i <= task.RetryCount; i++ {
		finalErr = taskFunc(ctx, TaskParams{
			TaskID:     task.ID,
			TaskName:   task.Name,
			RetryCount: i,
		})

		if finalErr == nil {
			execution.Status = stateSuccess
			break
		}

		if i < task.RetryCount {
			time.Sleep(1 * time.Second)
		}
	}
	return
}
