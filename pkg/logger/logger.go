package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// This provide dependency injection
type Field = zapcore.Field

// We need only these, Panic and Fatal was avoided here
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

const (
	LEVEL_DEBUG = zapcore.DebugLevel
	LEVEL_INFO  = zapcore.InfoLevel
	LEVEL_WARN  = zapcore.WarnLevel
	LEVEL_ERROR = zapcore.ErrorLevel
)

// New provides the zap logger and overrides the standard log if used
func New(level zapcore.Level) Logger {
	consoleErrors := zapcore.Lock(os.Stdout)

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeDuration = zapcore.MillisDurationEncoder
	config.EncodeName = zapcore.FullNameEncoder

	logLevel := zap.NewAtomicLevel()
	logLevel.SetLevel(level)

	consoleEncoder := zapcore.NewJSONEncoder(config)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, logLevel),
	)

	caller := zap.AddCaller()
	dev := zap.Development()

	log := zap.New(core, caller, dev)

	return log
}
