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
	TaskName string
}

func NewTaskManager(taskCfg *config.TaskConfig) *TaskManager {
	initLogger(taskCfg.LogLevel)
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
			if entryInfo.CronExpr == task.CronExpr && entryInfo.TaskName == task.Name {
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
			TaskName: task.Name,
		}

		m.logger.Debugf("添加/更新任务调度[%v-%v-%v]", task.ID, task.Name, task.CronExpr)
	}

	for taskID, entryInfo := range m.taskEntries {
		if !keepTaskIDs[taskID] {
			m.cron.Remove(entryInfo.EntryID)
			delete(m.taskEntries, taskID)
			m.logger.Debugf("移除已删除的任务[%v-%v]", taskID, entryInfo.TaskName)
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

// 执行过程，最难的地方就是如何登记 任务执行表
// 1. 我们纠结，任务成功了，要不要登记，我们纠结的是 记录太多，所以纠结，但是最为一个成熟的框架，必须登记
// 所以我们还是 按照传统登记吧,成功失败都要登记，跳过就跳过了，里面不能panic，不捕获panic
var (
	stateFiled     = "failed"
	stateSuccess   = "success"
	lockTaskFailed = "任务抢占失败"
	noFunc         = "任务函数未注册"
)

func (m *TaskManager) executeTask(task *model.Hawthorn_task) {

	traceID, err := utils.GenerateTraceID()
	if err != nil {
		traceID = "traceErr"
	}
	lg := taskLogger.With(zap.String("traceID", traceID))

	execution := &model.Hawthorn_task_execution{
		TaskID:      task.ID,
		NodeID:      m.nodeID,
		TraceID:     traceID,
		StartTime:   time.Now(),
		CreatedDate: time.Now(),
	}
	var finalErr error

	defer func() {
		err := recover()
		if err != nil {
			execution.Error = fmt.Sprintf("未知异常：%v", err)
			execution.Status = stateFiled
		} else if finalErr != nil {
			execution.Error = fmt.Sprintf("任务失败：%v", finalErr)
			execution.Status = stateFiled
		} else if execution.Status == "" {
			return
		}
		end := time.Now()
		execution.EndTime = &end
		if err := m.repo.CreateExecution(m.ctx, execution); err != nil {
			panic(fmt.Errorf("登记执行记录失败: %v", err))
		}
	}()

	lockErr := m.repo.TryLockTask(m.ctx, task.ID, 2*time.Second)
	if lockErr != nil {
		if errors.Is(lockErr, gorm.ErrRecordNotFound) {
			return
		}
		execution.Status = stateFiled
		execution.Error = lockTaskFailed
		lg.Sugar().Errorw("任务%d-%s抢占失败:%w", task.ID, task.Name, lockErr)
		return
	}

	ctx := context.WithValue(context.Background(), "traceID", traceID)
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
