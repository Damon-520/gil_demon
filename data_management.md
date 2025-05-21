## Data Management

This section outlines the "gil_teacher" application's approach to data management, including the various data stores utilized, the Object-Relational Mapper (ORM) and clients for interaction, and the role of Data Access Objects (DAOs) in abstracting data persistence logic.

### Overview of Data Stores

The application employs a polyglot persistence strategy, leveraging several data storage technologies to meet different needs. These are primarily configured in `app/conf/conf.go` within the `Data` struct and their clients are initialized in `main/gil_teacher/wire_gen.go`.

The key data stores include:

*   **MySQL:** Used for transactional data, likely for core application entities. Read/write replicas might be configured (e.g., `ActivityWrite`, `ActivityRead`).
*   **PostgreSQL:** Another relational database used for primary data storage, similar to MySQL. Multiple configurations exist (e.g., `PostgreSQLWrite`, `PostgreSQLRead`, a general `PostgreSQL`).
*   **ClickHouse:** Employed for analytical workloads and potentially storing event or behavioral data. Separate read/write configurations suggest its use for high-volume data.
*   **Redis:** Utilized for caching, session management, and other temporary or fast-access data storage needs.
*   **MongoDB:** Configured for use, suggesting it may store document-based data, though its specific DAOs were not as prominently featured in the analyzed `wire_gen.go` snippet compared to others.
*   **PolarDB:** Also configured, which is a cloud-native database service compatible with MySQL and PostgreSQL, likely used for scalability and reliability.

### SQL Databases (MySQL, PostgreSQL, PolarDB)

*   **Purpose:** These relational databases serve as the primary storage for the application's core structured data. This includes entities related to users, tasks, schedules, live rooms, and other foundational elements of an educational platform.
*   **ORM - GORM:** The application uses GORM (`gorm.io/gorm`, with drivers `gorm.io/driver/mysql` and `gorm.io/driver/postgres` as seen in `go.mod`) as its Object-Relational Mapper for interacting with these SQL databases.
    *   Evidence of GORM usage is found in DAO implementations. For example, `app/dao/task/task.go` defines `taskDAO` with a `*gorm.DB` field and uses GORM methods for database operations (e.g., `d.DB(ctx).Create(task)`).
    *   `main/gil_teacher/wire_gen.go` shows the instantiation of GORM database connections (e.g., `dao.NewActivityDB`, `dao.NewPostgreSQLClient` which then provides a `*gorm.DB` via `providers3.ProvidePostgreSQLDB`) that are subsequently injected into DAOs.

### ClickHouse

*   **Purpose:** ClickHouse is designed for high-performance online analytical processing (OLAP). In "gil_teacher," it is likely used for:
    *   Storing and analyzing large volumes of event data, such as user behavior logs, interaction tracking within live rooms, or task completion events.
    *   Generating reports and statistics that require fast aggregation over large datasets.
    *   The `behavior.NewBehaviorDAO` initialized with a ClickHouse client (`dao.NewClickHouseRWClient`) in `main/gil_teacher/wire_gen.go` strongly supports this use case.
*   **Client:** The application uses the `github.com/ClickHouse/clickhouse-go/v2` client library, as specified in `go.mod`, to interact with the ClickHouse server.

### Redis

*   **Purpose:** Redis is an in-memory data store typically used for:
    *   **Caching:** Storing frequently accessed data to reduce database load and improve response times (e.g., caching user sessions, schedules, or configuration parameters).
    *   **Session Management:** Managing user session information for authenticated users.
    *   **Rate Limiting, Queues (simple cases), Pub/Sub:** While not explicitly shown, these are common Redis use cases.
    *   The instantiation of `dao.NewApiRedisClient` in `main/gil_teacher/wire_gen.go` and its injection into services like `task_service.NewTaskService` and domain handlers like `behavior2.NewBehaviorHandler` and `schedule.NewScheduleCacheService` indicates its active role.
*   **Client:** The `github.com/go-redis/redis/v8` library (from `go.mod`) is used for Redis communication.

### MongoDB

*   **Purpose:** MongoDB is a NoSQL document database. Its presence in `app/conf/conf.go` (as `Mongo *Mongo`) and the `go.mongodb.org/mongo-driver/v2` driver in `go.mod` suggests it's available for storing:
    *   Data with flexible schemas.
    *   Large, complex documents.
    *   Content management or user-generated content that doesn't fit well into relational structures.
*   While specific DAO initializations for MongoDB were not prominent in the `main/gil_teacher/wire_gen.go` snippet reviewed, its configuration implies it is part of the data storage landscape, potentially for specific microservices or modules not covered in detail.

### Data Access Objects (DAOs)

*   **Role:** Data access across all these storage systems is encapsulated within Data Access Objects, primarily located in the `app/dao/` directory and its subdirectories (e.g., `app/dao/task/`, `app/dao/behavior/`, `app/dao/live_room/`).
*   **Abstraction:** DAOs provide a clean API for the service layer to interact with data, abstracting the underlying database-specific queries and logic. This promotes separation of concerns, making the codebase more modular and easier to maintain. Each DAO is typically responsible for a specific entity or a group of related entities within a particular data store.
*   **Dependency Injection:** As seen in `main/gil_teacher/wire_gen.go`, DAOs are instantiated with the appropriate database clients/ORM instances and injected into the service or domain layers that require them.

This multi-faceted approach to data management allows "gil_teacher" to choose the most appropriate data store for different types of data and workloads, optimizing for performance, scalability, and flexibility.
