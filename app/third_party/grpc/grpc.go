package grpc

import (
	"context"
	"google.golang.org/grpc/metadata"
)

const GRPCHeaderPrefix = "grpcgateway-"

func SetContext(ctx context.Context, key string, val string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, key, val)
}

func HeaderFromContext(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	return getHeader(md, key)
}

func getHeader(md metadata.MD, key string) string {
	if v := md.Get(GRPCHeaderPrefix + key); len(v) > 0 {
		return v[0]
	}
	if v := md.Get(key); len(v) > 0 {
		return v[0]
	}
	return ""
}
