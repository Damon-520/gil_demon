package trace

import (
	"context"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/spf13/cast"
)

// WrapTraceIdForCtx server端中间件获取trace，先取header里的，取不到给默认值
func WrapTraceIdForCtx(traceKey string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			traceId := ""
			header, ok := transport.FromServerContext(ctx)
			if ok {
				traceId = GetTraceIdFromHeader(header)
			}

			if md, ok := metadata.FromServerContext(ctx); ok {
				mTraceId := md.Get(traceKey)
				if mTraceId == "" {
					md.Set(traceKey, traceId)
					ctx = metadata.NewServerContext(ctx, md)
				} else {
					traceId = mTraceId
				}
			}
			ctx = context.WithValue(ctx, HeaderTraceId, traceId)
			if ok {
				header.RequestHeader().Set(HeaderTraceId, traceId)
			}
			return handler(ctx, req)
		}
	}
}

// WrapRpcIdForCtxClient client端
func WrapRpcIdForCtxClient(rpcIdKey string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				header := tr.RequestHeader()
				rpcId := header.Get(rpcIdKey)
				if rpcId != "" {
					header.Set(rpcIdKey, rpcId+".0")
				}
			}

			return handler(ctx, req)
		}
	}
}

// WrapTraceIdForCtxClient client端
func WrapTraceIdForCtxClient(traceIdKey string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				header := tr.RequestHeader()
				traceId := header.Get(traceIdKey)
				if traceId == "" {
					traceId = NewTraceId()
				}
				header.Set(traceIdKey, traceId)
			}

			return handler(ctx, req)
		}
	}
}

// WrapStressTestingForCtx  增加压力测试标识
func WrapStressTestingForCtx() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if header, ok := transport.FromServerContext(ctx); ok {
				stressTestingValue := ""
				stressTesting := header.RequestHeader().Get(HeaderStressTestingKey)
				if cast.ToString(stressTesting) != "" {
					stressTestingValue = cast.ToString(stressTesting)
				}
				if md, ok := metadata.FromServerContext(ctx); ok {
					md.Set(HeaderStressTestingKey, stressTestingValue)
					ctx = metadata.NewServerContext(ctx, md)
				}
				ctx = context.WithValue(ctx, HeaderStressTestingKey, stressTestingValue)
				return handler(ctx, req)
			}
			return handler(ctx, req)
		}
	}
}
