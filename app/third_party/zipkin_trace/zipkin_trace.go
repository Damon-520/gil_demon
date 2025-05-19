package zipkin_trace

import (
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

// newZipkinTracer creates a new Zipkin tracer.
func NewZipkinTracer() (*zipkin.Tracer, func(), error) {
	zipkinURL := "http://zipkin.local.xiaoluxue.cn:9411/zipkin/api/v2/spans"
	reporter := zipkinhttp.NewReporter(zipkinURL)
	endpoint, _ := zipkin.NewEndpoint("demo-api", "localhost:8080")
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		reporter.Close()
	}
	return tracer, cleanup, nil
}
