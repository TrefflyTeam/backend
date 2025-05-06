package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(env string) *zap.Logger {
	var zapLogger *zap.Logger
	var err error

	switch env {
	case "production":
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "@timestamp"
		config.EncoderConfig.MessageKey = "message"
		zapLogger, err = config.Build()
	default:
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapLogger, err = config.Build(zap.AddStacktrace(zap.ErrorLevel))
	}

	if err != nil {
		panic(err)
	}

	return zapLogger
}
