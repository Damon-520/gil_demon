package trace

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cast"
)

const (
	HeaderRpcId              = "x-rpcid"             // rpcId
	HeaderTraceId            = "x-traceid"           // traceId
	HeaderBffTraceId         = "x-aio-trace-id"      // bff 传递的traceId
	HeaderStressTestingKey   = "xes-request-type"    // 压测标志的key
	HeaderStressTestingValue = "performance-testing" // 压测标志真正的value
	XesTracePtsPrefix        = "pts_"                // 压测traceID前缀
)

func NewTraceId() string {
	return "gil_teacher-" + uuid.NewV4().String()
}

// IsStressTest 判断是否是压测
func IsStressTest(ctx *gin.Context) bool {
	// 1、指定的 header  k=>v
	stressTestString := ctx.Value(HeaderStressTestingKey)
	if stressTestString != nil && cast.ToString(stressTestString) == HeaderStressTestingValue {
		return true
	}
	// 2、traceId 包含压测前缀 XES_TRACE_PTS_PREFIX
	if md, ok := metadata.FromServerContext(ctx); ok {
		traceId := md.Get(HeaderTraceId)
		if traceId != "" && strings.HasPrefix(traceId, XesTracePtsPrefix) {
			return true
		}
	}

	return false
}

// GetTraceIdFromHeader 获取traceId
func GetTraceIdFromHeader(transport transport.Transporter) (traceId string) {
	traceId = transport.RequestHeader().Get(HeaderBffTraceId)
	if traceId == "" {
		traceId = transport.RequestHeader().Get(HeaderTraceId)
	}
	if traceId == "" {
		traceId = NewTraceId()
	}
	return
}
