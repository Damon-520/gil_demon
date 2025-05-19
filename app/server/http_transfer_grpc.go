package server

import (
	"net/http"
	"sort"
)

type interceptor = func(http.Handler) http.Handler

// 对 controller server 的包装，封装了 拦截器
type httpServer struct {
	httpMux      *http.ServeMux
	interceptors []interceptor
}

// 初始化，传入定义的拦截器
func NewWithInterceptors(interceptors ...interceptor) *httpServer {
	// 逆序，排前边的拦截器先执行
	sort.Slice(interceptors, func(i, j int) bool {
		return j < i
	})
	return &httpServer{
		httpMux:      http.NewServeMux(),
		interceptors: interceptors,
	}
}

// 添加 controller 路由，配置拦截器
func (s *httpServer) AddRoute(path string, handler http.Handler) {
	h := handler
	for _, c := range s.interceptors {
		h = c(h)
	}
	s.httpMux.Handle(path, h)
}

// 添加 controller 路由，配置自定义的拦截器
func (s *httpServer) AddRouteWith(path string, handler http.Handler, interceptors ...interceptor) {
	sort.Slice(interceptors, func(i, j int) bool {
		return j < i
	})
	h := handler
	for _, c := range interceptors {
		h = c(h)
	}
	s.httpMux.Handle(path, h)
}

// 实现 http.Handler 接口
func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpMux.ServeHTTP(w, r)
}
