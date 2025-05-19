package healthx

import (
	"fmt"
	"gil_teacher/app/core/envx"
	"net/http"
)

type HealthServer struct {
	Env  string
	Port int64
}

func NewHealthServer(env string, port int64) *HealthServer {
	return &HealthServer{
		Env:  env,
		Port: port,
	}
}

func (healthServer *HealthServer) Run() {
	if healthServer.Env == envx.ENV_TEST || healthServer.Env == envx.ENV_PROD {
		go func() {
			// 注册处理函数
			http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				_, err := fmt.Fprintf(w, "Hello, World!")
				if err != nil {
					return
				}
			})
			// 启动服务器
			fmt.Printf("Health server listening on port %d\n", healthServer.Port)

			addr := fmt.Sprintf(":%d", healthServer.Port)
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				fmt.Println("Error starting server:", err)
				panic(err)
			}
		}()
	}
}
