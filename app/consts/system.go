package consts

// 环境
const (
	LocalEnv  = "local"
	TestEnv   = "test"
	OnlineEnv = "online"
)

const (
	StartModeAll  = 0 // 全部 http & grpc
	StartModeHttp = 1 // 只启动http
	StartModeGrpc = 2 // 只启动grpc
)

const (
	ContextTraceID = "trace_id"
	ContextSpanID  = "span_id"
)

type TraceIDKey struct{}

type SpanIDKey struct{}
