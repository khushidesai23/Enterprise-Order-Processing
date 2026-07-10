package logger

import (
	"strings"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init() error {

	level := zapcore.InfoLevel

	switch strings.ToLower(config.Get().Log.Level) {

	case "debug":
		level = zapcore.DebugLevel

	case "warn":
		level = zapcore.WarnLevel

	case "error":
		level = zapcore.ErrorLevel
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      config.Get().App.Env == "development",
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},

		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
	}

	var err error

	log, err = cfg.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return err
	}

	return nil
}

func L() *zap.Logger {
	return log
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}