package framework

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/config"
	"github.com/hawthorntrees/cronframework/framework/cron"
	"github.com/hawthorntrees/cronframework/framework/dbs"
	"github.com/hawthorntrees/cronframework/framework/logger"
	"github.com/hawthorntrees/cronframework/framework/model"
	"github.com/hawthorntrees/cronframework/framework/route"
	"github.com/hawthorntrees/cronframework/framework/utils"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Framework struct {
	config      *config.Config
	taskManager *cron.TaskManager
	engine      *gin.Engine
	router      *gin.RouterGroup
	ctx         context.Context
	cancel      context.CancelFunc
	server      *http.Server
	running     bool
	mu          sync.Mutex
	wg          sync.WaitGroup
	//db          *gorm.DB
	log *zap.Logger
}

func New(filepath string) (*Framework, error) {
	cfg := config.Init(filepath)
	utils.InitSnowflake(cfg.App.WorkID)
	log := logger.Init(&cfg.Logger).With(zap.String("traceID", "boot-server"))
	log.Debug("日志管理器初始化成功")
	dbs.Init(log, cfg)
	log.Debug("数据源初始化成功")

	var taskManager *cron.TaskManager
	taskManager = cron.NewTaskManager(&cfg.CronTask)
	log.Debug("任务管理器初始化成功")

	engine, router := route.Init(&cfg.Server)
	log.Debug("路由初始化成功")

	ctx, cancel := context.WithCancel(context.Background())
	f := &Framework{
		config:      cfg,
		taskManager: taskManager,
		engine:      engine,
		router:      router,
		ctx:         ctx,
		cancel:      cancel,
		log:         log,
	}
	f.server = &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	log.Debug("框架初始化完成")

	return f, nil

}

func (f *Framework) RegisterTask(fn cron.TaskFunc) error {
	return f.taskManager.RegisterTask(fn)
}

func (f *Framework) Router() *gin.RouterGroup {
	return f.router
}
func (f *Framework) Start() error {
	f.mu.Lock()
	if f.running {
		f.mu.Unlock()
		return errors.New("框架已启动")
	}
	f.running = true
	f.mu.Unlock()
	if err := f.taskManager.Start(); err != nil {
		return err
	}
	f.wg.Add(1)

	go func() {
		defer f.wg.Done()
		f.log.Info("HTTP服务启动")
		if err := f.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			f.log.Error("HTTP服务异常退出", zap.Error(err))
		}
	}()
	f.wg.Add(1)
	go f.gracefulShutdown()
	f.wg.Wait()
	return nil
}

// Stop 停止框架
func (f *Framework) Stop() {
	f.cancel()
	f.taskManager.Stop()
	f.log.Debug("框架已停止")
	f.log.Sync()

}

func (f *Framework) gracefulShutdown() {
	defer f.wg.Done()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		f.log.Debug("收到退出信号:" + sig.String())
	}

	f.log.Debug("开始停止服务")
	f.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := f.server.Shutdown(ctx); err != nil {
		f.log.Warn("HTTP服务关闭超时:" + err.Error())
	} else {
		f.log.Info("已关闭HTTP服务")
	}

	dbs.CloseDBS()
}

func (f *Framework) AutoMigrate() error {
	db := dbs.GetDB()
	err := db.AutoMigrate(&model.Hawthorn_task{}, &model.Hawthorn_task_execution{})
	if err != nil {
		f.log.Warn("数据迁移失败")
		return err
	}
	return nil
}
