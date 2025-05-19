package prometheus

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration distribution",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func Init() {
	prometheus.MustRegister(RequestCounter, RequestDuration)
}

// PromMiddleware Gin 中间件：收集 Prometheus 指标
func PromMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过 /metrics 端点
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// 记录开始时间
		timer := prometheus.NewTimer(RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		))
		defer timer.ObserveDuration()

		// 处理请求
		c.Next()

		// 记录请求计数
		RequestCounter.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCodeToString(c.Writer.Status()),
		).Inc()
	}
}

// GetHandler 获取指标暴露 Handler
func GetHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

func statusCodeToString(code int) string {
	return fmt.Sprintf("%d", code)
}
