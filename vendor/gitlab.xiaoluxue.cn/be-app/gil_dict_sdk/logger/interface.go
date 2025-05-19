package logger

import "context"

type Logger interface {
	// Debug 日志方法
	Debug(ctx context.Context, format string, args ...interface{})

	// Info 日志方法
	Info(ctx context.Context, format string, args ...interface{})

	// Warn 日志方法
	Warn(ctx context.Context, format string, args ...interface{})

	Error(ctx context.Context, format string, args ...interface{})

	Fatal(ctx context.Context, format string, args ...interface{})
}
