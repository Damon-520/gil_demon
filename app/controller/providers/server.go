package providers

import (
	"gil_teacher/app/controller/grpc_server/live_http"
	"gil_teacher/app/controller/grpc_server/user"
	"gil_teacher/app/server"
)

// NewServerRegisters 创建服务注册器
func NewServerRegisters(
	lhh *live_http.LiveRoomHttp,
	uh *user.UserServer,
) []server.ServerRegister {
	return []server.ServerRegister{
		lhh,
		uh,
	}
}
