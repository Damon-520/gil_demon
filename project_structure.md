## Project Structure and Layers

This section outlines the directory structure of the "gil_teacher" project and describes its layered architecture, which promotes separation of concerns and modularity.

### Directory Structure Overview

The project is organized into several key top-level directories:

*   `app/`: Contains the core application logic, including configuration, server setup, controllers (HTTP/gRPC handlers), services (business logic), domain entities, data access objects (DAOs), middleware, and various core utilities.
*   `main/`: Houses the entry points for the application's executables. For instance, `main/gil_teacher/` contains the `main.go` for the primary "gil_teacher" service, and `main/gil_teacher_consumer/` likely contains the entry point for a separate Kafka consumer service.
*   `proto/`: Stores Protocol Buffer (`.proto`) definitions. These are used to define gRPC services and their associated message structures. Generated Go code from these definitions is typically found in subdirectories like `proto/gen/go/`.
*   `common/`: Includes shared utility functions and common initialization code that is used across different parts of the application (e.g., base configuration loading).
*   `libs/`: Contains local, project-specific libraries and helper packages.
*   `docs/`: Project-related documentation.
*   `script/`: Contains utility scripts for development, deployment, or other operational tasks.
*   `third_party/`: Includes third-party libraries or integrations that are not managed as standard Go modules, or specific customizations to them.
*   `vendor/`: Contains vendored Go dependencies, managed by Go's module system, ensuring reproducible builds.

### Layered Architecture

The "gil_teacher" project follows a layered architecture, which can be inferred from its directory structure and the dependency wiring observed in `main/gil_teacher/wire_gen.go`. This layering helps in organizing code, managing dependencies, and improving maintainability.

1.  **Main (`main/`)**:
    *   **Purpose:** Serves as the application's entry point. It's responsible for initializing the Kratos application, setting up configurations, and starting the servers.
    *   **Key Files:** `main/gil_teacher/main.go`, `main/gil_teacher/wire.go` (dependency definitions), `main/gil_teacher/wire_gen.go` (generated dependency injection code).

2.  **Configuration (`app/conf/`)**:
    *   **Purpose:** Defines data structures for application configuration, which are typically loaded from files or environment variables.
    *   **Key Files:** `app/conf/conf.go` (Go structs for configuration), `app/conf/conf.proto` (protobuf definitions for configuration, if used with a config management system like Nacos).

3.  **Server (`app/server/`)**:
    *   **Purpose:** Responsible for setting up and managing the gRPC and HTTP servers. It integrates these servers into the Kratos application lifecycle.
    *   **Key Files:** `app/server/server.go` (general server setup, combines HTTP and gRPC), `app/server/http_server.go` (Gin HTTP server setup), `app/server/grpc_server.go` (gRPC server setup).

4.  **Router (`app/router/`)**:
    *   **Purpose:** Handles HTTP request routing for the Gin framework. It maps incoming HTTP requests to the appropriate handler functions in the controller layer.
    *   **Key Files:** `app/router/router.go`.

5.  **Controller/Interfaces (`app/controller/`)**:
    *   **Purpose:** This layer acts as the presentation or interface layer. It receives incoming requests (HTTP or gRPC), validates them, and passes them to the service layer for processing.
    *   **Subdirectories:**
        *   `app/controller/http_server/`: Contains HTTP handlers (often using Gin) that process RESTful API requests.
        *   `app/controller/grpc_server/`: Contains gRPC service implementations that handle RPC calls.
    *   **Example from `wire_gen.go`:** `controller_task.NewTaskController`, `live_http.NewLiveRoomHttp`.

6.  **Service (`app/service/`)**:
    *   **Purpose:** Contains the core business logic orchestration. Services coordinate interactions between DAOs, domain objects, and other services to fulfill application use cases. They are responsible for transaction management and higher-level business rule enforcement.
    *   **Example from `wire_gen.go`:** `task_service.NewTaskService`, `live_service.NewLiveRoomService`.

7.  **Domain (`app/domain/`)**:
    *   **Purpose:** Encapsulates the fundamental business logic, entities, and rules of the application. This layer often includes domain models, value objects, and domain-specific handlers or repositories that are independent of the application's infrastructure.
    *   **Example from `wire_gen.go`:** `task.NewTaskReportHandler`, `behavior2.NewBehaviorHandler`. Some services might directly use DAOs if the domain logic is simple, but complex operations often involve domain handlers.

8.  **Data Access Object (DAO) / Repository (`app/dao/`)**:
    *   **Purpose:** Provides an abstraction layer for data persistence. DAOs are responsible for all interactions with the database (e.g., CRUD operations). They map application-level data structures to database schemas and vice-versa.
    *   **Example from `wire_gen.go`:** `impl.NewLiveRoomDao` (MySQL via GORM), `dao_task.NewTaskDAO` (PostgreSQL via GORM), `behavior.NewBehaviorDAO` (ClickHouse).

9.  **Core (`app/core/`)**:
    *   **Purpose:** Contains common, cross-cutting concerns and utilities that are used throughout the application. This includes custom logging, database client wrappers (e.g., for MySQL, PostgreSQL), Kafka clients, Zipkin tracing setup, etc.
    *   **Example from `wire_gen.go`:** `logger.NewContextLogger`, `zipkinx.NewTracer`, `dao.NewPostgreSQLClient`.

10. **Middleware (`app/middleware/`)**:
    *   **Purpose:** Provides a way to process requests and responses globally or for specific routes/groups. Middleware can handle tasks like authentication, logging, tracing, CORS, and error handling.
    *   **Key Files/Example from `wire_gen.go`:** `middleware.NewMiddleware`, `middleware.NewTeacherMiddleware`.

11. **Proto (`proto/`)**:
    *   **Purpose:** Contains Protocol Buffer (`.proto`) files that define the structure of gRPC services and the messages they exchange. This is crucial for inter-service communication in a microservices architecture.

12. **Common (`common/`)**:
    *   **Purpose:** A utility layer for code that is shared across multiple parts of the application but doesn't fit neatly into other specific layers. This can include helper functions, constants, or base initialization routines.

### Flow of Control

A typical request flow through the "gil_teacher" application would be as follows:

1.  An incoming request (HTTP or gRPC) arrives at the **Server** layer.
2.  For HTTP requests, the **Router** maps the request to a specific handler in the **Controller** layer. For gRPC, the gRPC server directly invokes the corresponding service implementation in the **Controller** layer.
3.  **Middleware** may intercept and process the request before and after it reaches the controller (e.g., for authentication, logging).
4.  The **Controller** validates the request and calls the appropriate method in the **Service** layer.
5.  The **Service** layer orchestrates the business logic, potentially interacting with **Domain** objects/handlers and one or more **DAO**s to retrieve or persist data.
6.  **DAO**s execute database operations.
7.  The results are passed back up through the layers: Service -> Controller -> Server -> Client.
8.  **Core** utilities (like logging, configuration access) and **Common** helpers can be used by any layer as needed.

This layered approach helps to create a decoupled and maintainable system where each layer has a distinct responsibility.I have analyzed the directory structure and `wire_gen.go` to understand the project's layers and their interactions. I have drafted the content for `project_structure.md`.

To verify, I will read the file and then submit the subtask report.
