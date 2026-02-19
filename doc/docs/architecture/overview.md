# Architecture Overview

HexaGo generates projects that follow **Hexagonal Architecture** (also known as Ports & Adapters), a software design pattern introduced by Alistair Cockburn.

---

## The Core Idea

Your application's business logic lives in an isolated **core** that has no knowledge of how it is invoked or how it persists data. All communication with the outside world happens through **ports** (interfaces) that are implemented by **adapters**.

```
┌─────────────────────────────────────────────────────┐
│                    ADAPTERS                         │
│                                                     │
│  HTTP Handler   gRPC Handler   Message Consumer     │
│       │              │               │              │
│  ─────┼──────────────┼───────────────┼─────────     │
│       ▼              ▼               ▼              │
│                   CORE                              │
│            ┌──────────────────┐                    │
│            │    Services /    │                    │
│            │    Use Cases     │                    │
│            │        │         │                    │
│            │        ▼         │                    │
│            │      Domain      │                    │
│            └──────────────────┘                    │
│       ▲              ▲               ▲              │
│  ─────┼──────────────┼───────────────┼─────────     │
│       │              │               │              │
│  DB Repository   Cache Adapter   Email Service      │
│                                                     │
└─────────────────────────────────────────────────────┘
```

---

## The Dependency Rule

**Dependencies always flow inward.** Outer layers know about inner layers, but inner layers never know about outer layers.

```
Adapters → Services/UseCases → Domain
```

| Layer | Can depend on | Cannot depend on |
|-------|--------------|-----------------|
| Domain | Nothing | Services, Adapters |
| Services | Domain | Adapters, Infrastructure |
| Adapters | Services (via interfaces) | Other adapters |

This rule ensures that your core business logic is **testable in isolation** and can be used with any adapter (HTTP today, gRPC tomorrow) without changing a line of business logic.

---

## Layers

### 1. Domain (`internal/core/domain/`)

The innermost layer. Contains:

- **Entities** — Business objects with identity (`User`, `Order`, `Product`)
- **Value Objects** — Immutable domain concepts (`Email`, `Money`, `Address`)
- **Business Rules** — Validation and invariants within entities

**Rules:**
- ✅ Pure Go — zero external dependencies
- ✅ Can contain business logic and validation
- ❌ No imports from services, adapters, or infrastructure

```go
// Domain entity — no external dependencies
package domain

type User struct {
    ID    string
    Name  string
    Email Email  // value object
}

func (u *User) ChangeName(name string) error {
    if name == "" {
        return ErrNameRequired
    }
    u.Name = name
    return nil
}
```

---

### 2. Services / Use Cases (`internal/core/services/`)

The application logic layer. Contains:

- **Services** — Orchestrate domain objects to fulfill use cases
- **Port Interfaces** — Define what the service needs from the outside world

**Rules:**
- ✅ Orchestrate domain objects
- ✅ Define port interfaces (repository interfaces, external service interfaces)
- ✅ Framework-agnostic
- ❌ No direct imports of adapters
- ❌ No direct database or HTTP dependencies

```go
// Service defines a port interface and uses it
package services

type UserRepository interface {          // Port definition
    FindByID(ctx context.Context, id string) (*domain.User, error)
    Save(ctx context.Context, user *domain.User) error
}

type CreateUserService struct {
    repo UserRepository                  // Depends on interface, not implementation
}

func (s *CreateUserService) Execute(ctx context.Context, input CreateUserInput) (*domain.User, error) {
    user := domain.NewUser(input.ID, input.Name, input.Email)
    return user, s.repo.Save(ctx, user)
}
```

---

### 3. Adapters (`internal/adapters/`)

The outermost layer. Contains:

- **Primary / Driver** — Inbound adapters (HTTP handlers, gRPC servers, message consumers)
- **Secondary / Driven** — Outbound adapters (database repositories, external API clients, caches)

**Rules:**
- ✅ Implement port interfaces defined in the core layer
- ✅ Handle all external communication (HTTP, database, queues)
- ✅ Use frameworks and external libraries here
- ❌ No business logic — delegate to services
- ❌ No cross-adapter imports

```go
// Adapter implements the port interface
package database

type PostgresUserRepository struct {
    db *sql.DB
}

// Implements services.UserRepository interface
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    // Database query here
    row := r.db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = $1", id)
    // ...
}
```

---

### 4. Infrastructure (`internal/config/`, `pkg/`)

Cross-cutting concerns that support all layers:

- **Configuration** — Viper-based config with YAML and environment variables
- **Logging** — Structured logger package
- **Observability** — Health checks and Prometheus metrics

---

## Naming Conventions

HexaGo supports two naming styles, both representing the same concept:

| Concept | DDD Terminology | Ports & Adapters |
|---------|----------------|-----------------|
| Inbound adapters | `adapters/primary/` | `adapters/driver/` |
| Outbound adapters | `adapters/secondary/` | `adapters/driven/` |
| Business logic | `services/` | `usecases/` |

Choose the vocabulary that fits your team. Set during `hexago init` with `--adapter-style` and `--core-logic`.

---

## Further Reading

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) — Alistair Cockburn's original article
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) — Robert C. Martin
- [Ports & Adapters](https://herbertograca.com/2017/11/16/explicit-architecture-01-ddd-hexagonal-onion-clean-cqrs-how-i-put-it-all-together/) — Herberto Graça
