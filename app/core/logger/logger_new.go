package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gil_teacher/app/consts"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/google/wire"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var ProviderSet = wire.NewSet(NewContextLogger)

// ContextLogger 封装 kratos log，支持从 context 中获取额外信息
type ContextLogger struct {
	log.Logger
	logLevel gormLog.LogLevel
}

// NewContextLogger 创建一个新的 ContextLogger
func NewContextLogger(kLog log.Logger) *ContextLogger {
	return &ContextLogger{
		Logger:   kLog,
		logLevel: gormLog.Info,
	}
}

// LogMode log mode
func (l *ContextLogger) LogMode(level gormLog.LogLevel) gormLog.Interface {
	newlogger := *l
	newlogger.logLevel = level
	return &newlogger
}

func (l *ContextLogger) Log(level log.Level, keyvals ...interface{}) error {
	return l.Logger.Log(level, keyvals...)
}

// Log 实现带 context 的日志记录
func (l *ContextLogger) LogWithContent(ctx context.Context, level log.Level, keyvals ...interface{}) error {
	// 从 context 中获取额外信息
	extraFields := extractContextFields(ctx)

	// 合并原有的 keyvals 和从 context 中提取的字段
	newKeyvals := make([]interface{}, 0, len(keyvals)+len(extraFields)*2)
	newKeyvals = append(newKeyvals, keyvals...)
	for k, v := range extraFields {
		newKeyvals = append(newKeyvals, k, v)
	}

	return l.Logger.Log(level, newKeyvals...)
}

// extractContextFields 从 context 中提取需要记录的字段
func extractContextFields(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})

	if ginCtx, ok := ctx.(*gin.Context); ok {
		traceID, ok := ginCtx.Get(consts.ContextTraceID)
		if ok {
			fields[consts.ContextTraceID] = traceID
			return fields
		}
		spanID, ok := ginCtx.Get(consts.ContextSpanID)
		if ok {
			fields[consts.ContextSpanID] = spanID
		}
	}

	traceID, ok := ctx.Value(consts.TraceIDKey{}).(string)
	if ok {
		fields[consts.ContextTraceID] = traceID
		return fields
	}

	md, ok := metadata.FromServerContext(ctx)
	if ok {
		if traceID := md.Get(consts.ContextTraceID); traceID != "" {
			fields[consts.ContextTraceID] = traceID
		}
	}

	return fields
}

// Debug 日志方法
func (l *ContextLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	l.LogWithContent(ctx, log.LevelDebug, "msg", fmt.Sprintf(format, args...))
}

// Info 日志方法
func (l *ContextLogger) Info(ctx context.Context, format string, args ...interface{}) {
	l.LogWithContent(ctx, log.LevelInfo, "msg", fmt.Sprintf(format, args...))
}

// Warn 日志方法
func (l *ContextLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	l.LogWithContent(ctx, log.LevelWarn, "msg", fmt.Sprintf(format, args...))
}

// Error 日志方法
func (l *ContextLogger) Error(ctx context.Context, format string, args ...interface{}) {
	l.LogWithContent(ctx, log.LevelError, "msg", fmt.Sprintf(format, args...))
}

// Fatal 日志方法
func (l *ContextLogger) Fatal(ctx context.Context, format string, args ...interface{}) {
	l.LogWithContent(ctx, log.LevelFatal, "msg", fmt.Sprintf(format, args...))
}

// Trace print sql message
func (l *ContextLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && !errors.Is(err, gormLog.ErrRecordNotFound):
		sql, rows := fc()
		if rows == -1 {
			l.LogWithContent(ctx, log.LevelError, "x_file", utils.FileWithLineNum(), "x_error", err, "x_duration", elapsed.Seconds(), "x_rows", "-", "x_action", sql)
		} else {
			l.LogWithContent(ctx, log.LevelError, "x_file", utils.FileWithLineNum(), "x_error", err, "x_duration", elapsed.Seconds(), "x_rows", rows, "x_action", sql)
		}
	default:
		sql, rows := fc()
		if rows == -1 {
			l.LogWithContent(ctx, log.LevelInfo, "x_file", utils.FileWithLineNum(), "x_duration", elapsed.Seconds(), "x_rows", "-", "x_action", sql)
		} else {
			l.LogWithContent(ctx, log.LevelInfo, "x_file", utils.FileWithLineNum(), "x_duration", elapsed.Seconds(), "x_rows", rows, "x_action", sql)
		}
	}
}
