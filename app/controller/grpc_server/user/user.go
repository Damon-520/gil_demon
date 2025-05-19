package user

import (
	"context"

	"gil_teacher/app/core/logger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pb "gil_teacher/proto/gen/go/proto/gil_teacher/api" // 替换成你的proto包路径
	userpb "gil_teacher/proto/gen/go/proto/gil_teacher/user"
)

// 服务实现
type UserServer struct {
	log                                  *logger.ContextLogger
	pb.UnimplementedApiUserServiceServer // 替换成你的proto service名称
}

func NewUserServer(log *logger.ContextLogger) *UserServer {
	return &UserServer{log: log}
}

// 注册gRPC服务
func (s *UserServer) RegisterGRPC(server grpc.ServiceRegistrar) {
	pb.RegisterApiUserServiceServer(server, s)
}

// 注册HTTP服务
func (s *UserServer) RegisterHTTP(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pb.RegisterApiUserServiceHandler(ctx, mux, conn)
}

// 用户列表
func (s *UserServer) List(ctx context.Context, req *pb.ListReq) (*pb.ListRes, error) {
	s.log.Info(ctx, "infolog")
	return &pb.ListRes{
		Code:    200,
		Message: "success",
		Data: &pb.ListRes_Data{
			Lists: []*userpb.User{
				{
					Id:    2,
					Name:  "张三",
					Email: "zhangsan@example.com",
				},
			},
		},
	}, nil
}
