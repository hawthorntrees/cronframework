package cron

import (
	"context"
	"github.com/hawthorntrees/cronframework/framework/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var taskLogger *zap.Logger

var taskLevel = zap.NewAtomicLevel()

func initLogger(level string) {
	core := logger.GetLoggerCore()
	switch level {
	case "debug":
		taskLevel.SetLevel(zapcore.DebugLevel)
	case "warn":
		taskLevel.SetLevel(zapcore.WarnLevel)
	case "error":
		taskLevel.SetLevel(zapcore.ErrorLevel)
	case "info":
		taskLevel.SetLevel(zapcore.InfoLevel)
	default:
		taskLevel.SetLevel(zapcore.InvalidLevel)
	}
	taskLogger = zap.New(core,
		zap.AddCaller(),
		zap.IncreaseLevel(taskLevel),
		//zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}
func GetLogger(ctx context.Context) *zap.Logger {
	traceID := ctx.Value("traceID")
	if traceID != nil {
		return taskLogger.With(zap.String("traceID", traceID.(string)))
	} else {
		return taskLogger.With(zap.String("traceID", "traceIDErr"))
	}
}
