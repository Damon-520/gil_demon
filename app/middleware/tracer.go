package middleware

import (
	"context"

	"gil_teacher/app/consts"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/metadata"
)

// Trace 暂时用于http
func (m *Middleware) HttpTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 HTTP tracer
		span, err := m.tracer.GetHttpTracer(c)
		if err != nil {
			m.log.Error(c, "get http tracer failed %v", err)
			c.Next()
			return
		}
		defer span.Finish()

		// 将 trace 信息注入到请求头和响应头
		traceID := span.Context().TraceID.String()
		spanID := span.Context().ID.String()

		// 设置到请求头
		c.Request.Header.Set(consts.ContextTraceID, traceID)
		c.Request.Header.Set(consts.ContextSpanID, spanID)

		// 设置到 metadata
		md := metadata.New(map[string][]string{
			consts.ContextTraceID: {traceID},
			consts.ContextSpanID:  {spanID},
		})
		ctx := metadata.NewServerContext(c.Request.Context(), md)

		// 同时设置到 context.Value
		ctx = context.WithValue(ctx, consts.TraceIDKey{}, traceID)
		ctx = context.WithValue(ctx, consts.SpanIDKey{}, spanID)
		// 设置到 gin.Context
		c.Set(consts.ContextTraceID, traceID)
		c.Set(consts.ContextSpanID, spanID)
		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 设置到响应头
		c.Header(consts.ContextTraceID, traceID)
		c.Header(consts.ContextSpanID, spanID)

		c.Next()
	}
}
