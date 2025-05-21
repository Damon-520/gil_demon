## Request Flow

This section details the typical flow of requests through the "gil_teacher" application, covering both HTTP and gRPC communication paths. The application's layered architecture ensures a structured and maintainable way of processing requests.

### HTTP Request Flow

The HTTP request flow is primarily managed by the Gin web framework, integrated within the Kratos application structure.

1.  **Incoming Request & Server:** An HTTP request first reaches the Kratos HTTP server, which is configured using Gin. The `app/server/http_server.go` file (specifically `NewGinHttpServer`) shows how the Gin engine is set up and integrated.

2.  **Routing (Gin):** The Gin engine, configured with routes, directs the request to the appropriate handler.
    *   Route definitions are managed in `app/router/router.go` (e.g., `RegisterRouter` function). `main/gil_teacher/wire_gen.go` shows `route.NewHttpRouter` being instantiated with various controller dependencies. This router maps HTTP paths and methods (e.g., `GET /es-search`) to specific functions within HTTP controllers.

3.  **Middleware Execution:** Before the request reaches the controller handler, it passes through a series of Gin middlewares.
    *   **Global Middleware:** Configured in `app/server/http_server.go` within `GinGlobalMiddleware`. This includes general middleware for tracing (`mid.HttpTrace()`), request logging (`mid.GinRequestLogger()`), and CORS (`mid.CORS()`). The `middleware.Middleware` instance providing these is created in `main/gil_teacher/wire_gen.go` (`middleware.NewMiddleware(cnf, tracer, contextLogger)`).
    *   **Route-Specific/Group Middleware:** Some routes or groups might have additional middleware. For example, authentication or authorization middleware like `middleware.NewTeacherMiddleware` (seen in `main/gil_teacher/wire_gen.go` injected into controllers like `controller_task.NewTaskController`) is used. `app/middleware/teacher.go`'s `WithTeacherContext()` is an example that fetches teacher details based on request headers and populates the Gin context.

4.  **Controller (`app/controller/http_server/`):** The matched controller handler processes the request.
    *   Controllers are responsible for parsing request parameters, validating input, and then delegating the core business logic to the service layer.
    *   `main/gil_teacher/wire_gen.go` illustrates this, e.g., `controller_task.NewTaskController` is injected with and calls methods on `taskService`, `taskResourceService`, etc.

5.  **Service Layer (`app/service/`):** This layer contains the primary business logic.
    *   Services orchestrate operations, manage transactions, and coordinate data access by calling methods in the DAO layer.
    *   For example, `task_service.NewTaskService` (from `wire_gen.go`) would handle task-related business operations, using `taskDAO` and `taskAssignDAO`.

6.  **Data Access Object (DAO) Layer (`app/dao/`):** DAOs are responsible for all database interactions.
    *   They execute CRUD operations, abstracting the database details from the service layer. GORM is typically used here.
    *   Examples from `wire_gen.go`: `dao_task.NewTaskDAO(db)` interacts with PostgreSQL, `impl.NewLiveRoomDao(activityDB)` interacts with a MySQL database.

7.  **Response Generation:** The controller receives data or status from the service layer and generates an HTTP response (often JSON), which is then sent back to the client through the Gin engine and Kratos server. The `app/third_party/response/response.go` package likely provides standardized response formatting.

### gRPC Request Flow

The gRPC request flow is managed by the Kratos framework using its gRPC server capabilities.

1.  **Incoming Request & Server:** A gRPC request arrives at the Kratos gRPC server, configured in `app/server/grpc_server.go` (e.g., `NewGRPCServer`). This server handles the underlying gRPC protocol.

2.  **gRPC Interceptors:** Similar to middleware in HTTP, gRPC interceptors process the request before it reaches the service implementation.
    *   `app/server/grpc_server.go` shows the configuration of unary server interceptors: `grpc.UnaryInterceptor(grpcMid.ChainUnaryServer(...))`.
    *   The `middleware.Middleware` instance (created in `main/gil_teacher/wire_gen.go`) provides these interceptors, including `s.mid.RequestLog`, `s.mid.Auth`, `s.mid.ParseHeader`, `s.mid.WrapTraceIdForCtx`, and `s.mid.Recovery()`. These handle concerns like logging, authentication, header parsing, trace ID propagation, and panic recovery for gRPC requests.

3.  **gRPC Service Implementation (Controller/Handler - `app/controller/grpc_server/` or Service methods):** The request is dispatched to the specific gRPC service method implementation.
    *   These implementations are defined based on `.proto` files (located in `proto/`).
    *   `app/server/grpc_server.go` shows that services are registered via a list of `ServerRegister` interfaces. `main/gil_teacher/wire_gen.go` shows `providers.NewServerRegisters(liveRoomHttp, userServer)` where `liveRoomHttp` (an instance of `live_http.NewLiveRoomHttp`) and `userServer` are gRPC service handlers.
    *   These handlers, like `live_http.NewLiveRoomHttp`, are themselves injected with business logic services (e.g., `liveRoomService` as seen in `wire_gen.go`).

4.  **Service Layer (`app/service/`) / Domain Layer (`app/domain/`):** The gRPC service implementation calls methods in the main business service layer or domain handlers.
    *   This is where the core application logic resides, similar to the HTTP flow. For example, `live_service.NewLiveRoomService` (injected into `live_http.LiveRoomHttp`) would contain the business logic for live room operations.

5.  **Data Access Object (DAO) Layer (`app/dao/`):** Services or domain handlers interact with DAOs for data persistence.
    *   This layer functions identically to the HTTP flow, abstracting database operations. For instance, `live_service.NewLiveRoomService` uses `iLiveRoomDao`.

6.  **gRPC Response:** The gRPC service implementation returns a response message (defined in a `.proto` file), which is then sent back to the client via the Kratos gRPC server. Error handling often involves returning gRPC status codes, as seen with `DefaultHTTPErrorHandleFunc` in `app/server/grpc_server.go` which translates gRPC statuses for HTTP gateway purposes.

In both flows, `main/gil_teacher/wire_gen.go` is instrumental in showing how these layers are wired together through dependency injection, ensuring that controllers have access to services, and services have access to DAOs and other necessary components. This layered approach promotes separation of concerns and code reusability.
