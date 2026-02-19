# hexago add adapter

Add a primary (inbound) or secondary (outbound) adapter to an existing project.

## Synopsis

```shell
hexago add adapter primary <type> <name>
hexago add adapter secondary <type> <name>
```

Must be run from the project root directory.

---

## Primary Adapters (Inbound)

Primary adapters handle **incoming requests** — they drive your application.

```shell
hexago add adapter primary <type> <name>
```

**Available types:**

| Type | Description | Example Use |
|------|-------------|-------------|
| `http` | HTTP request handler | REST API endpoints |
| `grpc` | gRPC service handler | gRPC service endpoints |
| `queue` | Message queue consumer | Kafka, RabbitMQ consumers |

**Examples:**

```shell
hexago add adapter primary http UserHandler
hexago add adapter primary http ProductHandler
hexago add adapter primary grpc OrderService
hexago add adapter primary queue EmailConsumer
```

---

## Secondary Adapters (Outbound)

Secondary adapters handle **outgoing communication** — they are driven by your application.

```shell
hexago add adapter secondary <type> <name>
```

**Available types:**

| Type | Description | Example Use |
|------|-------------|-------------|
| `database` | Database repository | PostgreSQL, MySQL, SQLite |
| `external` | External API client | Payment gateways, third-party APIs |
| `cache` | Cache adapter | Redis, in-memory cache |

**Examples:**

```shell
hexago add adapter secondary database UserRepository
hexago add adapter secondary database ProductRepository
hexago add adapter secondary external EmailService
hexago add adapter secondary external PaymentGateway
hexago add adapter secondary cache UserCache
```

---

## Generated Files

### Primary HTTP adapter

For `hexago add adapter primary http UserHandler`:

```
internal/adapters/primary/http/
└── user_handler.go
```

### Secondary database adapter

For `hexago add adapter secondary database UserRepository`:

```
internal/adapters/secondary/database/
└── user_repository.go
```

---

## Generated Code Structure

**HTTP Handler (Echo):**

```go
package http

import (
    "net/http"
    "github.com/labstack/echo/v4"
)

// UserHandler handles HTTP requests for User
type UserHandler struct {
    // TODO: Add service dependencies
}

// NewUserHandler creates a new UserHandler
func NewUserHandler() *UserHandler {
    return &UserHandler{}
}

// RegisterRoutes registers routes on the given router
func (h *UserHandler) RegisterRoutes(e *echo.Echo) {
    e.GET("/users", h.List)
    e.POST("/users", h.Create)
    e.GET("/users/:id", h.Get)
}

// List handles GET /users
func (h *UserHandler) List(c echo.Context) error {
    // TODO: Implement
    return c.JSON(http.StatusOK, nil)
}
```

**Database Repository:**

```go
package database

import "context"

// UserRepository implements the user storage port
type UserRepository struct {
    // TODO: Add database connection
}

// NewUserRepository creates a new UserRepository
func NewUserRepository() *UserRepository {
    return &UserRepository{}
}

// FindByID retrieves a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (interface{}, error) {
    // TODO: Implement database query
    return nil, nil
}
```

---

## Naming Conventions

If your project uses `--adapter-style driver-driven`, the directories are:

- Primary adapters → `internal/adapters/driver/`
- Secondary adapters → `internal/adapters/driven/`

HexaGo auto-detects your project's naming convention.

---

## Architecture Notes

Adapters belong to the **adapters layer** and must:

- ✅ Implement port interfaces defined in the core layer
- ✅ Handle external communication (HTTP, database, queues)
- ✅ Use frameworks and external libraries
- ❌ Never import other adapter packages
- ❌ Never contain business logic (delegate to services)
