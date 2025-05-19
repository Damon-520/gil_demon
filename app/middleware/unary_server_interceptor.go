package middleware

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"gil_teacher/app/consts"
	"gil_teacher/app/third_party/middlewares/trace"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"

	"google.golang.org/grpc"
)

var ErrUnknownRequest = errors.InternalServer("UNKNOWN", "unknown request error")

/*
本文件以 grpc.UnaryServerInterceptor func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) 的方式
实现中间件
http转grpc后的grpc方法在 info.FullMethod 中
*/

func (m *Middleware) RequestLog(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()
	res, err := handler(ctx, req)
	spendTime := time.Since(startTime).Microseconds()

	if err != nil {
		m.log.Error(ctx, "Request full_method: %s, spend_time_us: %d, error: %s", info.FullMethod, spendTime, err.Error())
	} else {
		m.log.Info(ctx, "Request full_method: %s, spend_time_us: %d", info.FullMethod, spendTime)
	}
	return res, err
}

func (m *Middleware) Auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	//token := grpc2.HeaderFromContext(ctx, "token")
	//fmt.Println("token", token)
	return handler(ctx, req)
}

func (m *Middleware) ParseHeader(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 解析transport
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return handler(ctx, req)
	}
	// 获取请求头
	header := tr.RequestHeader()

	if userId := header.Get("uid"); userId != "" {
		if uId, err := strconv.ParseInt(userId, 10, 32); err == nil {
			ctx = context.WithValue(ctx, consts.FormHeaderCtxUserIdKey, uId)
		}
	}

	return handler(ctx, req)
}

// WrapTraceIdForCtx server端中间件获取trace，先取header里的，取不到给默认值
func (m *Middleware) WrapTraceIdForCtx(traceKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		traceId := ""
		header, ok := transport.FromServerContext(ctx)
		if ok {
			traceId = trace.GetTraceIdFromHeader(header)
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
		ctx = context.WithValue(ctx, trace.HeaderTraceId, traceId)
		if ok {
			header.RequestHeader().Set(trace.HeaderTraceId, traceId)
		}
		return handler(ctx, req)
	}
}

type Latency struct{}

func (m *Middleware) Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (reply interface{}, err error) {
		startTime := time.Now()
		defer func() {
			if rerr := recover(); rerr != nil {
				buf := make([]byte, 64<<10) //nolint:mnd
				n := runtime.Stack(buf, false)
				buf = buf[:n]
				m.log.Error(ctx, "%v: %+v\n%s\n", rerr, req, buf)
				ctx = context.WithValue(ctx, Latency{}, time.Since(startTime).Seconds())
				err = errors.InternalServer("INTERNAL_SERVER_ERROR", fmt.Sprintf("%+v", rerr))
			}
		}()
		return handler(ctx, req)
	}
}
