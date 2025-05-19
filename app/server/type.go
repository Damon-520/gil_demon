package server

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type (
	// GatewayHandler gRPC gateway 生成的 handler 代码
	GatewayHandler func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

	// HTTPErrorHandleFunc 错误返回处理
	// 如果指定了该 handler, 需要自己序列化返回的 body 体, 以及指定 HTTP status code
	// 如果内部序列化失败,建议直接返回一个 fallabck 的结果.
	HTTPErrorHandleFunc func(sts *status.Status) (_ []byte, statusCode int)

	// GRPCErrorHandleFunc gRPC 层的错误转换, 会将放回的错误信息放置于  status.Status.details 中
	GRPCErrorHandleFunc func(err error) proto.Message

	// 权限认证, 可以返回新的 context.Context 进行传递
	AuthHandleFunc func(ctx context.Context, info *grpc.UnaryServerInfo) (context.Context, error)
)

type ServerRegister interface {
	// 注册到 gRPC 服务
	RegisterGRPC(server grpc.ServiceRegistrar)
	// 注册到 controller 服务
	RegisterHTTP(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
}
