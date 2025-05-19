package zipkinx

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"

	"github.com/gin-gonic/gin"
	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

type Tracer struct {
	appName      string
	reporter     reporter.Reporter
	httpEndpoint *model.Endpoint
	grpcEndpoint *model.Endpoint
}

func NewTracer(cfg *conf.Conf) (*Tracer, func(), error) {
	httpEndpoint, err := zipkin.NewEndpoint(cfg.App.Name, cfg.Server.Http.Addr)
	if err != nil {
		return nil, nil, err
	}
	grpcEndpoint, err := zipkin.NewEndpoint(cfg.App.Name, cfg.Server.Grpc.Addr)
	if err != nil {
		return nil, nil, err
	}
	reporter := zipkinhttp.NewReporter(cfg.ZipKin.Url)
	return &Tracer{
			appName:      cfg.App.Name,
			reporter:     reporter,
			httpEndpoint: httpEndpoint,
			grpcEndpoint: grpcEndpoint,
		}, func() {
			reporter.Close()
		}, nil
}

// 获取http tracer
func (t *Tracer) GetHttpTracer(c *gin.Context) (zipkin.Span, error) {
	tracer, err := zipkin.NewTracer(t.reporter, zipkin.WithLocalEndpoint(t.httpEndpoint))
	if err != nil {
		return nil, err
	}

	// 从请求头中提取现有的 trace 上下文（如果有）
	extractor := t.getTraceSpanFunc(c)
	sc := tracer.Extract(extractor)
	span := t.GetSpan(c, tracer, sc)
	return span, nil
}

// 获取grpc tracer
func (t *Tracer) GetGrpcTracer() (*zipkin.Tracer, error) {
	tracer, err := zipkin.NewTracer(t.reporter, zipkin.WithLocalEndpoint(t.grpcEndpoint))
	if err != nil {
		return nil, err
	}
	return tracer, nil
}

// 获取http span
func (t *Tracer) GetSpan(c *gin.Context, tracer *zipkin.Tracer, sc model.SpanContext) zipkin.Span {
	// 创建 span，如果有父上下文则作为子 span
	var span zipkin.Span
	if sc.Err != nil {
		// 没有父上下文，创建新的根 span，但使用不同的 traceID 和 spanID
		traceID := model.TraceID{
			High: 0,
			Low:  uint64(time.Now().UnixNano()),
		}
		spanID := model.ID(uint64(rand.Int63()))

		sampled := true
		customContext := model.SpanContext{
			TraceID: traceID,
			ID:      spanID,
			Sampled: &sampled, // 设置为 true
		}

		span = tracer.StartSpan(
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			zipkin.Parent(customContext),
		)
	} else {
		// 有父上下文，创建子 span
		span = tracer.StartSpan(
			fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			zipkin.Parent(sc),
		)
	}

	return span
}

func (t *Tracer) getTraceSpanFunc(c *gin.Context) func() (*model.SpanContext, error) {
	return func() (*model.SpanContext, error) {
		var sc model.SpanContext
		if traceID := c.Request.Header.Get(consts.ContextTraceID); traceID != "" {
			if id, err := model.TraceIDFromHex(traceID); err == nil {
				sc.TraceID = id
			}
		}
		if spanID := c.Request.Header.Get(consts.ContextSpanID); spanID != "" {
			if id, err := strconv.ParseUint(spanID, 16, 64); err == nil {
				sc.ID = model.ID(id)
			}
		}
		return &sc, nil
	}
}
