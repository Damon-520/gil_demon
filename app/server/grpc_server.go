package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	go_runtime "runtime"
	"strings"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/middleware"
	uTime "gil_teacher/app/third_party/time"
	pb "gil_teacher/proto/gen/go/proto/gil_teacher/base"

	"github.com/go-kratos/kratos/v2/log"
	kHttp "github.com/go-kratos/kratos/v2/transport/http"
	grpcMid "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server 服务器结构体
type GRPCServer struct {
	Cfg             *conf.Server
	serverRegisters []ServerRegister // 服务注册器列表，用于注册 gRPC 和 HTTP 服务
	logger          log.Logger       // 日志记录器
	mid             *middleware.Middleware

	// 允许透进来的 HTTP headers
	IncomingHeaderWhiteList []string
}

func NewGRPCServer(
	cfg *conf.Server,
	logger log.Logger, // 日志记录器
	mid *middleware.Middleware,
	serverRegisters []ServerRegister, // 服务注册器列表
) (*GRPCServer, error) {
	if cfg.Grpc.Addr == "" {
		return nil, errors.New("addr is empty")
	}
	if len(serverRegisters) == 0 {
		return nil, errors.New("serverRegisters is empty")
	}
	svr := &GRPCServer{
		mid:             mid,
		Cfg:             cfg,
		logger:          logger,
		serverRegisters: serverRegisters,
		IncomingHeaderWhiteList: []string{
			// 允许透传进来的自定义token
			"token",
			"x-token",
			"trace_id",
		},
	}

	return svr, nil
}

// Register 启动服务器
// 该方法实现了：
// 1. 建立 gRPC 连接
// 2. 配置 JSON 序列化选项
// 3. 设置 gRPC-Gateway 多路复用器
// 4. 注册 HTTP 和 gRPC 服务
// 5. 配置中间件和拦截器
// 6. 启动服务器并处理优雅关闭
func (s *GRPCServer) Register(ctx context.Context) (*kHttp.Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// grpc地址
	grpcConn, err := dail("tcp", s.Cfg.Grpc.Addr)
	if err != nil {
		return nil, err
	}

	var options []runtime.ServeMuxOption
	options = append(options, runtime.WithErrorHandler(runtimeHTTPErrorHandler(DefaultHTTPErrorHandleFunc)))

	// proto解析规则
	jsonPb := &runtime.JSONPb{}
	jsonPb.MarshalOptions.UseProtoNames = true
	jsonPb.MarshalOptions.UseEnumNumbers = true
	jsonPb.MarshalOptions.EmitUnpopulated = true
	options = append(options, runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonPb))
	options = append(options, runtime.WithIncomingHeaderMatcher(s.incomingHeaderMatcher()))
	gw := runtime.NewServeMux(options...)
	// 注册 HTTP 服务
	for _, reg := range s.serverRegisters {
		if err := reg.RegisterHTTP(ctx, gw, grpcConn); err != nil {
			return nil, fmt.Errorf("failed to register HTTP service: %v", err)
		}
	}

	// 通用系统中间件
	hs := NewWithInterceptors()

	// http路由
	hs.AddRoute("/healthz", healthzServer(grpcConn))

	// interceptor = func(controller.Handler) controller.Handler
	global := []interceptor{
		// func(h controller.Handler) controller.Handler
		s.mid.GetHeader,
		//s.mid.ZipKin, TODO 未安装zipkin，本地无法启动
	}
	hs.AddRouteWith("/", gw, global...)

	grpcOptions := []grpc.ServerOption{}

	grpcOptions = append(
		grpcOptions,
		grpc.UnaryInterceptor(grpcMid.ChainUnaryServer(
			s.mid.RequestLog, // 请求日志
			s.mid.Auth,
			s.mid.ParseHeader,
			s.mid.WrapTraceIdForCtx("trace_id"), // TODO 常量
			s.mid.Recovery(),
		)),
	)
	grpcServer := grpc.NewServer(grpcOptions...)
	// 注册 gRPC 服务
	for _, reg := range s.serverRegisters {
		reg.RegisterGRPC(grpcServer)
	}

	reflection.Register(grpcServer)

	// 创建支持 gRPC 和 HTTP 的多路复用处理器
	handler := grpcHandlerFunc(grpcServer, hs)

	srv := kHttp.NewServer([]kHttp.ServerOption{
		kHttp.Network(s.Cfg.Grpc.Network),
		kHttp.Address(s.Cfg.Grpc.Addr),
		kHttp.Timeout(uTime.ParseDuration(s.Cfg.Grpc.Timeout)),
	}...)
	srv.Handler = handler

	return srv, nil

	// 启动 HTTP/gRPC 服务器
	//srv := &controller.Server{
	//	Addr:    s.Cfg.Grpc.Addr,
	//	Handler: handler,
	//}

	// 在新的 goroutine 中启动服务器
	go func() {
		fmt.Printf("Starting HTTP/gRPC server on %s\n", s.Cfg.Grpc.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// 等待上下文取消或服务器错误
	<-ctx.Done()
	fmt.Println("Shutting down server...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return nil, fmt.Errorf("server shutdown error: %v", err)
	}

	return nil, nil
}

// dail 创建 gRPC 客户端连接
// 支持 TCP 和 Unix 域套接字两种连接方式
func dail(network, addr string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	switch network {
	case "tcp":
	case "unix":
		d := func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}
		opts = append(opts, grpc.WithContextDialer(d))
	default:
		return nil, fmt.Errorf("unsupported network type: %q", network)
	}
	return grpc.NewClient(addr, opts...)
}

// DefaultHTTPErrorHandleFunc 默认的 HTTP 错误处理函数
// 将 gRPC 状态错误转换为统一的 HTTP 响应格式
func DefaultHTTPErrorHandleFunc(sts *status.Status) (_ []byte, statusCode int) {
	// 暂时统一为 200.
	statusCode = 200

	er := &pb.Error{
		Code:    10000, // TODO 常量
		Message: sts.Err().Error(),
	}
	if details := sts.Details(); len(details) > 0 {
		if err, ok := details[0].(*pb.Error); ok {
			er.Code = err.Code
			er.Message = err.Message
		}
	}
	b, err1 := json.Marshal(er)
	if err1 != nil {
		// 序列化错误,使用 fallback
		return []byte(`{code": 10000, "message": "系统错误: 序列化结果失败"}`), statusCode
	}
	return b, statusCode
}

// incomingHeaderMatcher 请求头匹配器
// 用于将 HTTP 请求头转发到 gRPC 服务
func (s *GRPCServer) incomingHeaderMatcher() func(originHttpHeaderKey string) (string, bool) {
	allow := make(map[string]bool)
	for _, v := range s.IncomingHeaderWhiteList {
		allow[strings.ToLower(v)] = true
	}
	return func(originHttpHeaderKey string) (string, bool) {
		v, ok := runtime.DefaultHeaderMatcher(originHttpHeaderKey) // 先用标注的处理
		if ok {
			return v, ok
		}
		// 处理我们自己认识的 header, 最好变成小写
		v = strings.ToLower(originHttpHeaderKey)
		if allow[v] {
			return v, true
		}
		return "", false
	}
}

// printPanic panic 处理函数
// 捕获 panic，打印堆栈信息，并返回统一的错误响应
func printPanic(p interface{}) (err error) {
	var buf [4096]byte
	n := go_runtime.Stack(buf[:], false)

	msgMap := make(map[string]string)
	msgMap["level"] = "error"
	msgMap["msg"] = fmt.Sprintf("panic: %v\r\n%s", p, string(buf[:n]))
	msgJson, _ := json.Marshal(msgMap)
	fmt.Printf("%s\n", msgJson)
	return status.Errorf(codes.Internal, "系统异常")
}

// isValidToken token 验证函数
// TODO: 需要实现具体的 token 验证逻辑
func isValidToken(token string) bool {
	// TODO: 实现具体的 token 验证逻辑
	return token != ""
}

// grpcHandlerFunc 多协议处理函数"
// 根据请求协议版本和内容类型，将请求分发到 gRPC 或 HTTP 处理器
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

// allowCORS CORS 中间件
// 处理跨域请求，设置相关的 CORS 响应头
// 支持自定义允许的请求头和方法
func (s *GRPCServer) allowCORS(h http.Handler) http.Handler {
	allowHeaders := strings.Join(append([]string{"Content-Type", "Accept", "Authorization"}, s.IncomingHeaderWhiteList...), ",")
	allowMethods := strings.Join([]string{"GET", "HEAD", "POST", "PUT", "DELETE"}, ",")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("into allowCORS")
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// exposeHeaders 响应头暴露中间件
// 允许前端访问特定的响应头
// 主要用于暴露 traceid 相关的头信息
func exposeHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 允许前端获取 traceid
		traceHeader := "traceid"
		exposeHeaders := strings.Join([]string{traceHeader, "Grpc-Metadata-" + traceHeader}, ",")
		w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
		h.ServeHTTP(w, r)
	})
}

func healthzServer(conn *grpc.ClientConn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if s := conn.GetState(); s != connectivity.Ready && s != connectivity.Idle {
			http.Error(w, fmt.Sprintf("grpc server state is %s", s), http.StatusBadGateway)
			return
		}
		fmt.Fprintln(w, "ok")
	}
}

func (s *GRPCServer) healthCheckInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 跳过健康检查接口的健康检查
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		// 检查服务健康状态
		if !s.isHealthy() {
			return nil, status.Error(codes.Unavailable, "服务暂时不可用")
		}

		return handler(ctx, req)
	}
}

// isHealthy 检查服务是否健康
func (s *GRPCServer) isHealthy() bool {
	// TODO: 实现具体的健康检查逻辑
	// 可以检查：
	// 1. 数据库连接
	// 2. 缓存服务
	// 3. 依赖的外部服务
	// 4. 系统资源（内存、磁盘等）
	return true
}
