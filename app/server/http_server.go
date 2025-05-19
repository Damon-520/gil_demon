package server

import (
	"fmt"
	"io"

	"gil_teacher/app/consts"

	"gil_teacher/app/conf"
	"gil_teacher/app/controller/http_server/route"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/middleware"
	"gil_teacher/app/third_party/time"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func NewGinHttpServer(
	c *conf.Conf,
	logger *logger.ContextLogger,
	ginRoute *route.HttpRouter,
	mid *middleware.Middleware,
) *http.Server {

	var opts []http.ServerOption
	if c.Server.Http.Network != "" {
		opts = append(opts, http.Network(c.Server.Http.Network))
	}
	if c.Server.Http.Addr != "" {
		opts = append(opts, http.Address(c.Server.Http.Addr))
	}
	if c.Server.Http.Timeout != "" {
		opts = append(opts, http.Timeout(time.ParseDuration(c.Server.Http.Timeout)))
	}

	srv := http.NewServer(opts...)
	ginR := GinGlobalMiddleware(logger, mid)
	ginR = ginRoute.InitRouter(ginR)
	srv.HandlePrefix("/", ginR)

	if c.Config.Env == consts.LocalEnv && (c.App.Mode == consts.StartModeAll || c.App.Mode == consts.StartModeHttp) {
		fmt.Println("[GIN-debug] Routes:")
		for _, r := range ginR.Routes() {
			fmt.Printf("[GIN-debug] %-6s %-25s --> %s\n", r.Method, r.Path, r.Handler)
		}
	}

	return srv
}

// GinGlobalMiddleware 全局路由
func GinGlobalMiddleware(logger *logger.ContextLogger, mid *middleware.Middleware) *gin.Engine {
	//gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	// Default 默认带了 Logger 和 Recovery 中间件
	router_ := gin.Default()

	router_.Use(
		mid.HttpTrace(),
		mid.GinRequestLogger(logger),
		mid.CORS(),
	)

	return router_
}
