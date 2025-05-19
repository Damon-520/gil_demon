package server

import (
	"context"
	"fmt"
	"os"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
)

type Server struct {
	grpcServer *GRPCServer
	httpServer *http.Server
	log        *logger.ContextLogger
}

func NewServer(
	cnf *conf.Conf,
	grpcServer *GRPCServer,
	httpServer *http.Server,
	log *logger.ContextLogger,
) *kratos.App {
	id, _ := os.Hostname()
	ctx := context.Background()
	var servers []transport.Server

	switch cnf.App.Mode {
	case consts.StartModeAll:
		gs, err := grpcServer.Register(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("start HTTP server %s\n", cnf.Server.Http.Addr)
		fmt.Printf("start gRPC server %s\n", cnf.Server.Grpc.Addr)
		servers = append(servers, httpServer, gs)
	case consts.StartModeHttp:
		fmt.Printf("start HTTP server %s\n", cnf.Server.Http.Addr)
		servers = append(servers, httpServer)
	case consts.StartModeGrpc:
		gs, err := grpcServer.Register(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("start gRPC server %s\n", cnf.Server.Grpc.Addr)
		servers = append(servers, gs)
	default:
		panic("invalid mode")
	}
	return kratos.New(
		kratos.ID(id),
		kratos.Name(cnf.App.Name),
		kratos.Version(cnf.App.Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(log),
		kratos.Server(
			servers...,
		),
	)
}
