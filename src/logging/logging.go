package logging

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
	configLogger := zap.NewDevelopmentConfig()
	configLogger.EncoderConfig.TimeKey = zapcore.OmitKey
	configLogger.EncoderConfig.CallerKey = zapcore.OmitKey
	configLogger.EncoderConfig.ConsoleSeparator = " | "

	logger, err := configLogger.Build()
	if err != nil {
		panic(fmt.Sprintf("logger build failed: %+v", err))
	}
	return logger
}

var Logger = NewLogger()

func Info(template string, args ...interface{}) {
	Logger.Sugar().Infof(template, args...)
}

func Warn(template string, args ...interface{}) {
	Logger.Sugar().Warnf(template, args...)
}

func Error(template string, args ...interface{}) {
	Logger.Sugar().Errorf(template, args...)
}

func Exception(err error) {
	Logger.Sugar().Error(err)
}

func Debug(template string, args ...interface{}) {
	Logger.Sugar().Debugf(template, args...)
}
