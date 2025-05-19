package loggerx

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
)

func IncrRpcId(rpcidKey string) log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil {
			return ""
		}

		if c, ok := ctx.(*gin.Context); ok {
			extra := c.GetHeader(rpcidKey)
			last := strings.LastIndex(extra, ".")
			i, _ := strconv.Atoi(extra[last+1:])
			extra = extra[:last+1] + strconv.Itoa(i+1)
			c.Header(rpcidKey, extra)
			return extra
		}
		return ""
	}
}
