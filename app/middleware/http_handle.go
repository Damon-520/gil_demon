package middleware

import (
	"net/http"

	"gil_teacher/app/conf"
	"gil_teacher/app/core/logger"
	"gil_teacher/app/core/zipkinx"
)

type Middleware struct {
	AppName string
	tracer  *zipkinx.Tracer
	log     *logger.ContextLogger
}

func NewMiddleware(cfg *conf.Conf, tracer *zipkinx.Tracer, log *logger.ContextLogger) *Middleware {
	return &Middleware{
		AppName: cfg.App.Name,
		tracer:  tracer,
		log:     log,
	}
}

func (m *Middleware) GetHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
