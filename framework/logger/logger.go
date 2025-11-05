package logger

import (
	"fmt"
	"github.com/hawthorntrees/cronframework/framework/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

var loggerCore zapcore.Core
var baseLogger *zap.Logger

func Init(cfg *config.LoggerConfig) *zap.Logger {
	initCore(cfg)
	createLogger(cfg.Level)
	return baseLogger
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	localTime := t.Local()
	enc.AppendString(localTime.Format("2006-01-02 15:04:05.000"))
}

func initCore(cfg *config.LoggerConfig) {
	exePath, err2 := os.Executable()
	if err2 != nil {
		panic("获取路径失败")
	}
	dir := filepath.Dir(exePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(fmt.Errorf("创建日志目录失败%v", err))
	}

	writer := &lumberjack.Logger{
		Filename:   filepath.Join(dir, cfg.Filename),
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxFiles,
		Compress:   cfg.Compress,
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	loggerCore = zapcore.NewCore(
		jsonEncoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(writer)),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return true
		}),
	)
}

func createLogger(level string) {
	lv := zap.NewAtomicLevel()
	switch level {
	case "debug":
		lv.SetLevel(zapcore.DebugLevel)
	case "info":
		lv.SetLevel(zapcore.InfoLevel)
	case "warn":
		lv.SetLevel(zapcore.WarnLevel)
	case "error":
		lv.SetLevel(zapcore.ErrorLevel)
	default:
		lv.SetLevel(zapcore.InvalidLevel)

	}
	baseLogger = zap.New(loggerCore,
		zap.AddCaller(),
		zap.IncreaseLevel(lv),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}
func GetLoggerCore() zapcore.Core {
	return loggerCore
}

func GetBaseLogger() *zap.Logger {
	return baseLogger
}
