package dbs

import (
	"context"
	"fmt"
	logger2 "github.com/hawthorntrees/cronframework/framework/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"time"
)

var zaplog *loggerAdapter

type loggerAdapter struct {
	logger *zap.Logger
	level  zap.AtomicLevel
}

var levelMap = map[logger.LogLevel]zapcore.Level{
	logger.Silent: zapcore.InvalidLevel,
	logger.Info:   zapcore.InfoLevel,
	logger.Warn:   zapcore.WarnLevel,
	logger.Error:  zapcore.ErrorLevel,
}

func initLogger(lel string) {
	var level = zap.NewAtomicLevel()
	core := logger2.GetLoggerCore()
	switch lel {
	case "silent":
		level.SetLevel(zapcore.InvalidLevel)
	case "info":
		level.SetLevel(zapcore.InfoLevel)
	case "warn":
		level.SetLevel(zapcore.WarnLevel)
	case "error":
		level.SetLevel(zapcore.ErrorLevel)
	default:
		level.SetLevel(zapcore.InvalidLevel)
	}
	// 我们需要创建一个
	lg := zap.New(core,
		zap.AddCaller(),
		zap.IncreaseLevel(level),
		zap.AddCallerSkip(1),
	)
	zaplog = &loggerAdapter{logger: lg, level: level}
}

func (l *loggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	if zapLevel, ok := levelMap[level]; ok {
		l.level.SetLevel(zapLevel)
	}
	return l
}

func (l *loggerAdapter) Info(ctx context.Context, s string, i ...interface{}) {
	if !l.level.Enabled(zapcore.InfoLevel) {
		return
	}
	f := toZapFields(i...)
	traceID := ctx.Value("traceID")
	log := l.logger
	if traceID != nil {
		id, ok := traceID.(string)
		if ok {
			log = l.logger.With(zap.String("traceID", id))
		}
	}
	log.Info(s, f...)
}

func (l *loggerAdapter) Warn(ctx context.Context, s string, i ...interface{}) {
	if !l.level.Enabled(zapcore.WarnLevel) {
		return
	}
	f := toZapFields(i...)
	traceID := ctx.Value("traceID")
	log := l.logger
	if traceID != nil {
		id, ok := traceID.(string)
		if ok {
			log = l.logger.With(zap.String("traceID", id))
		}
	}
	log.Warn(s, f...)
}

func (l *loggerAdapter) Error(ctx context.Context, s string, i ...interface{}) {
	if !l.level.Enabled(zapcore.ErrorLevel) {
		return
	}
	f := toZapFields(i...)
	traceID := ctx.Value("traceID")
	log := l.logger
	if traceID != nil {
		id, ok := traceID.(string)
		if ok {
			log = l.logger.With(zap.String("traceID", id))
		}
	}
	log.Error(s, f...)
}

func (l *loggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	traceID := ctx.Value("traceID")
	log := l.logger
	if traceID != nil {
		id, ok := traceID.(string)
		if ok {
			log = l.logger.With(zap.String("traceID", id))
		}
	}
	// 记录SQL执行日志
	sql, rows := fc()
	elapsed := time.Since(begin)

	if err != nil {
		log.Error("SQL执行错误",
			zap.String("sql", sql),
			zap.Int64("rowsAffected", rows),
			zap.Duration("elapsed", elapsed),
			zap.Error(err))
	} else if traceID != nil {
		log.Info("SQL执行",
			zap.String("sql", sql),
			zap.Int64("rowsAffected", rows),
			zap.Duration("elapsed", elapsed))
	}
}

func toZapFields(args ...interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(args))
	for i := 0; i < len(args); i++ {
		if i+1 >= len(args) {
			fields = append(fields, zap.Any(fmt.Sprintf("arg_%d", i), args[i]))
			continue
		}
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("key_%d", i) // 非字符串 key 生成默认名称
		}
		fields = append(fields, zap.Any(key, args[i+1]))
		i++
	}
	return fields
}
