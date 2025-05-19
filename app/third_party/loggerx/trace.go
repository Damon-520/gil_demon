package loggerx

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
)

// 获取traceid
func TraceID(traceKey string) log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			return ""
		}

		if c, ok := ctx.(*gin.Context); ok {
			return c.GetHeader(traceKey)
		}

		return ""
	}
}
