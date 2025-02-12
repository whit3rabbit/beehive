package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Initialize(logLevel string) error {
	level := zap.InfoLevel
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	config := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			TimeKey:     "time",
			CallerKey:   "caller",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	var err error
	Log, err = config.Build()
	if err != nil {
		return err
	}

	// Replace the global logger with our new logger
	zap.ReplaceGlobals(Log)

	return nil
}

func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

func GetLogger() *zap.Logger {
	if Log == nil {
		// If the logger is not initialized, create a default logger
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
		var err error
		Log, err = config.Build()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
	}
	if Log == nil {
		return zap.NewNop()
	}
	return Log
}

func Debug(message string, fields ...zap.Field) {
	GetLogger().Debug(message, fields...)
}

func Info(message string, fields ...zap.Field) {
	GetLogger().Info(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	GetLogger().Warn(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	GetLogger().Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	GetLogger().Fatal(message, fields...)
	os.Exit(1)
}
