## Observability

The "gil_teacher" application incorporates several mechanisms to ensure system observability, allowing developers and operators to monitor its behavior, diagnose issues, and understand its performance. These mechanisms include logging, distributed tracing, and metrics collection.

### Logging

Comprehensive logging is implemented to provide insights into the application's operations and to aid in debugging.

*   **Custom Logger & Initialization:**
    *   The core logging functionality is defined in `app/core/logger/logger.go`. This custom logger utilizes the `logrus` library and provides a Kratos-compatible `log.Logger` interface.
    *   It is initialized in `common/init.go` via the `logInit` function. This function configures the logger based on settings from `app/conf/conf.go` (specifically the `Log` struct, which defines `Path`, `Level`, `RotationTime`, and `MaxAge`).
*   **Structured Logging:**
    *   The logger is set up to produce structured logs, typically in JSON format, as defined by the `logFormatter` in `app/core/logger/logger.go`. This formatter includes standard fields like `timestamp`, `level`, and `host`.
    *   When initialized through `common/init.go:logInit`, the logger is wrapped with `log.With(..., "service_name", appName)`, ensuring that a `service_name` field is consistently included in log entries. This aligns with Kratos's structured logging approach, making logs easier to parse, search, and analyze, especially in a microservices environment.
*   **Log Rotation:**
    *   To manage log file sizes and retention, the logger integrates with `github.com/lestrrat-go/file-rotatelogs` (as seen in `app/core/logger/logger.go` and its dependency in `go.mod`).
    *   Configuration for rotation (e.g., `RotationTime`, `MaxAge`) is provided via the `Log` struct in `app/conf/conf.go`. This ensures that log files are rotated periodically and old logs are automatically cleaned up.
*   **Output:** Logs are written to both standard output and to files if a `Path` is specified in the configuration.

### Distributed Tracing (Zipkin)

To understand the flow of requests across different services (if applicable) and within the application's components, distributed tracing is implemented using Zipkin.

*   **Tracer Implementation:**
    *   The Zipkin tracer is set up in `app/core/zipkinx/zipkinx.go`. The `NewTracer` function initializes a Zipkin tracer, configuring it with a reporter (`zipkinhttp.NewReporter`) that sends trace data to the Zipkin server specified by `cfg.ZipKin.Url` (from `app/conf/conf.go`).
    *   It also creates local endpoints for both HTTP and gRPC services (`httpEndpoint`, `grpcEndpoint`) to correctly identify the origin of spans.
*   **Initialization & Integration:**
    *   The Zipkin tracer is initialized at the application startup within `main/gil_teacher/wire_gen.go` by calling `zipkinx.NewTracer(cnf)`.
    *   The created `tracer` instance is then injected into `middleware.NewMiddleware`. This middleware is subsequently applied to both HTTP (Gin) and gRPC servers, allowing it to intercept requests/responses and manage span creation, propagation (extracting parent span context and injecting current span context for downstream calls), and reporting.
*   **Purpose:** Zipkin helps in visualizing request latency, understanding service dependencies, and pinpointing bottlenecks in a distributed system or complex monolithic application.
*   **Client Library:** The application uses `github.com/openzipkin/zipkin-go v0.4.3` (as listed in `go.mod`).

### Metrics (Prometheus)

The application exposes metrics for monitoring its performance and health using Prometheus.

*   **Metrics Definition & Collection:**
    *   `app/utils/prometheus/prometheus.go` defines custom Prometheus metrics, such as `RequestCounter` (a counter for total HTTP requests) and `RequestDuration` (a histogram for request latency), both including labels like method, path, and status.
    *   It provides a Gin middleware `PromMiddleware()` that intercepts HTTP requests to record these metrics.
*   **Initialization:**
    *   The Prometheus metrics are registered with the Prometheus client library during application startup by calling `prometheus.Init()` within `common/init.go:InitBase`.
*   **Exposure:**
    *   An HTTP endpoint (typically `/metrics`) is exposed for Prometheus to scrape these metrics. The `GetHandler()` function in `app/utils/prometheus/prometheus.go` provides a Gin handler (`gin.WrapH(promhttp.Handler())`) for this purpose. This handler would be registered in the Gin router setup.
*   **Purpose:** Prometheus metrics allow for real-time monitoring of key application indicators, alerting on anomalies, and understanding resource utilization and performance trends.
*   **Client Library:** The `github.com/prometheus/client_golang v1.22.0` library (from `go.mod`) is used to instrument the application code and expose metrics.

Together, these logging, tracing, and metrics capabilities provide a robust observability solution for the "gil_teacher" application, crucial for maintaining its reliability and performance.
