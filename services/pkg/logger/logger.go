package logger

import (
	"context"
)

type contextKey string
type loggerKey string

const (
	LoggerKey    contextKey = "logger"
	RequestIDKey contextKey = "requestID"
	RequestID    loggerKey  = "requestID"
	ServiceName  loggerKey  = "service"
)

type Field struct {
	Key   string
	Value any
}

func String(key, val string) Field  { return Field{Key: key, Value: val} }
func Int(key string, val int) Field { return Field{Key: key, Value: val} }
func Error(val error) Field         { return Field{Key: "error", Value: val} }
func Any(key string, val any) Field { return Field{Key: key, Value: val} }

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)

	With(fields ...Field) Logger
	Sync() error
}

func GetLoggerFromContext(ctx context.Context) (Logger, bool) {
	log, ok := ctx.Value(LoggerKey).(Logger)
	return log, ok
}

func AddLoggerToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, logger)
}
