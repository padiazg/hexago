# hexago add adapter

Add a primary (inbound) or secondary (outbound) adapter to an existing project.

## Synopsis

```shell
hexago add adapter primary <type> <name>
hexago add adapter secondary <type> <name>
```

Operates on the project root — use `--working-directory` (`-w`) to target a project without changing directories.

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
├── user_handler.go
└── user_handler_test.go
```

### Secondary database adapter

For `hexago add adapter secondary database UserRepository`:

```
internal/adapters/secondary/database/userrepository/
├── userrepository.go
└── userrepository_test.go
```

When `--entity` is provided, database adapters use the entity name as sub-package:

```
internal/adapters/secondary/database/user_repository/
├── user_repository.go
└── user_repository_test.go
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
func (r *UserRepository) FindByID(ctx context.Context, id string) (any, error) {
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

## Flags

### Primary adapter flags

| Flag | Short | Description |
|------|-------|-------------|
| `--entity` | `-e` | Domain entity this handler serves (PascalCase). Generates a sub-package with config and handler files. |
| `--with-test` | | Generate tests for this component (overrides project config). |
| `--no-test` | | Skip test generation for this component (overrides project config). |
| `--working-directory` | `-w` | Project root (defaults to the current directory). |

### Secondary adapter flags

| Flag | Short | Description |
|------|-------|-------------|
| `--entity` | `-e` | Domain entity this adapter implements (PascalCase). Determines the sub-package for database adapters. |
| `--from-port` | `-p` | Port interface name to infer method signatures from (semantic code generation). |
| `--with-test` | | Generate tests for this component using go-testgen (overrides project config). |
| `--no-test` | | Skip test generation for this component (overrides project config). |
| `--working-directory` | `-w` | Project root (defaults to the current directory). |

### `--from-port` — semantic method generation

When `--from-port` is specified, HexaGo loads the named port interface from the project
and generates adapter methods with the correct signatures:

```shell
hexago add adapter secondary database UserRepository --from-port UserRepository
hexago add adapter secondary external EmailService   --from-port EmailSender
```

If the interface cannot be loaded, generation falls back to generic stubs.

---

## Test Generation

HexaGo can scaffold tests for generated adapters using [go-testgen](https://padiazg.github.io/go-testgen/).

**Prerequisites:** install go-testgen ≥ v0.1.0:

```shell
go install github.com/padiazg/go-testgen@latest
```

**Per-command flags** (`--with-test` / `--no-test`) override the project-level default set by `hexago init --with-tests`.

```shell
# enable for this run (regardless of config)
hexago add adapter secondary database UserRepository --with-test

# disable for this run (regardless of config)
hexago add adapter primary http UserHandler --no-test
```

After the adapter files are written, HexaGo runs:

```shell
go-testgen report <adapter-pkg> --format json   # discover untested functions
go-testgen gen <adapter-pkg> <FuncSpec> ...     # generate each missing test
```

If go-testgen is missing, outdated, or the report fails, a warning is printed and the adapter
is still generated normally.

### `--entity` — domain entity binding

For primary HTTP adapters, `--entity` generates a sub-package with a config file and
CRUD handlers (List, Create, GetByID, Update):

```shell
hexago add adapter primary http UserHandler --entity User
```

For secondary database adapters, `--entity` creates the sub-package named after the
entity and generates repository methods:

```shell
hexago add adapter secondary database UserRepository --entity User
```

---

## Architecture Notes

Adapters belong to the **adapters layer** and must:

- ✅ Implement port interfaces defined in the core layer
- ✅ Handle external communication (HTTP, database, queues)
- ✅ Use frameworks and external libraries
- ❌ Never import other adapter packages
- ❌ Never contain business logic (delegate to services)
