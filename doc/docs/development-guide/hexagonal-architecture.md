# Hexagonal Architecture

Understanding the architecture pattern and dependency rules in HexaGo-generated projects.

---

## Core Principle

The dependency rule: **adapters → services → domain**

```
External World     Adapters      Services       Domain
  (HTTP, DB)    →  (primary/  →  (services/  →  (domain/
                   secondary)    ports/)        entities)
```

- Core has **zero external dependencies**
- Adapters implement interfaces defined by the core
- Dependency direction is always inward

---

## Layer Structure

### Domain Layer (`internal/core/domain/`)

Pure business entities with no external imports:

```go
// No imports from adapters or external packages
type User struct {
    ID        uuid.UUID
    Email     string
    Name      string
    CreatedAt time.Time
}

func (u *User) Validate() error {
    if u.Email == "" {
        return ErrEmailRequired
    }
    return nil
}
```

Contains:

- Entities (objects with unique identity)
- Value objects (immutable, compared by value)
- Domain errors

### Services Layer (`internal/core/services/`)

Business logic and use cases. Defines port interfaces:

```go
type UserService struct {
    store Store
}

type Store interface {
    GetUser(ctx context.Context, id string) (*User, error)
    SaveUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id string) error
}
```

Contains:

- Service structs with business logic
- Port interfaces (what the core needs from outside)
- Use case implementations

### Adapters Layer (`internal/adapters/`)

External interfaces — implements ports defined by services.

| Direction | Type | Examples |
| --- | --- | --- |
| **Primary** (inbound) | Driven by external actors | HTTP handlers, gRPC servers, CLI commands, queue consumers |
| **Secondary** (outbound) | Drives external systems | Database repositories, external API clients, cache adapters, notifiers |

```go
// Primary adapter - HTTP handler
type UserHandler struct {
    service *services.UserService
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    // Adapts HTTP request → service call → HTTP response
}

// Secondary adapter - database repository
type SQLiteUserRepository struct {
    db *sql.DB
}

func (r *SQLiteUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    // Implements the Store port interface
}
```

---

## Port Patterns

### Defining Ports

Ports are interfaces defined in the services layer:

```go
// internal/core/ports/store.go
package ports

type Store interface {
    GetUser(ctx context.Context, id string) (*domain.User, error)
    SaveUser(ctx context.Context, user *domain.User) error
    ListUsers(ctx context.Context) ([]*domain.User, error)
    DeleteUser(ctx context.Context, id string) error
}
```

### Implementing Ports

Adapters implement these interfaces:

```go
// internal/adapters/secondary/database/user_repository.go
package database

type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
    // Implementation
}

var _ ports.Store = (*UserRepository)(nil)  // Compile-time interface check
```

---

## Key Interfaces (Example)

These are common port interfaces in HexaGo projects:

| Interface | Purpose | Methods |
| --- | --- | --- |
| `Store` | Data persistence | Get, Save, List, Delete |
| `FeedProvider` | External data feeds | History, Subscribe, Ping |
| `Notifier` | Notifications | SendSignal, SendUpdate, SendError |
| `Reporter` | Report generation | WriteSimulation, WriteBacktest |

---

## Dependency Injection

Services receive dependencies via constructors:

```go
func NewUserService(store ports.Store, logger Logger) *UserService {
    return &UserService{
        store: store,
        logger: logger,
    }
}
```

The `cmd/run.go` wire-up assembles the application:

```go
// cmd/run.go
func run() error {
    // Create adapters
    repo := database.NewUserRepository(db)
    notifier := telegram.NewClient(apiKey)

    // Create services with adapter implementations
    userSvc := services.NewUserService(repo, logger)

    // Create primary adapters (handlers)
    handler := http.NewUserHandler(userSvc)

    // Start server with handler
    return httpServer.Serve(handler.Routes())
}
```

---

## Validation

Run architecture validation after any code changes:

```bash
hexago validate
```

Checks performed:

- ✓ Core domain has no external dependencies
- ✓ Services only depend on domain and ports
- ✓ Adapters don't import from other adapters
- ✓ Proper dependency direction (adapters → core)
- ✓ Proper package organization and naming

---

## Naming Variants

### Default (primary-secondary)

```shell
internal/
├── adapters/
│   ├── primary/    # Inbound adapters
│   └── secondary/  # Outbound adapters
```

### DDD / Driver-Driven

When using `--adapter-style driver-driven`:

```shell
internal/
├── adapters/
│   ├── driver/     # Inbound (drives the application)
│   └── driven/     # Outbound (driven by the application)
```

---

## Anti-Patterns to Avoid

| Anti-Pattern | Problem | Solution |
| --- | --- | --- |
| Core imports adapters | Violates dependency rule | Define port in domain, implement in adapter |
| Business logic in handlers | Leaky abstraction | Move to services |
| Database types in domain | External dependency in core | Use domain types, map in adapter |
| Direct HTTP calls in services | Tight coupling | Create external client adapter |
