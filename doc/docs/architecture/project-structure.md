# Project Structure

HexaGo generates a well-organized directory structure that enforces hexagonal architecture boundaries.

---

## Default Structure

```
my-app/
├── cmd/                          # CLI commands (Cobra)
│   ├── root.go                   # Root command + Viper configuration
│   └── run.go                    # Server startup with graceful shutdown
│
├── internal/                     # Private application code
│   ├── core/                     # 🎯 CORE — No external dependencies
│   │   ├── domain/               # Business entities and value objects
│   │   │   └── .gitkeep
│   │   └── services/             # Business logic and use cases
│   │       └── .gitkeep
│   │
│   ├── adapters/                 # 🔌 ADAPTERS — External interfaces
│   │   ├── primary/              # Inbound (drives the application)
│   │   │   └── http/             # HTTP handlers
│   │   │       └── .gitkeep
│   │   └── secondary/            # Outbound (driven by the application)
│   │       └── database/         # Database repositories
│   │           └── .gitkeep
│   │
│   ├── config/                   # Configuration management
│   │   └── config.go             # Viper config struct and loader
│   │
│   └── observability/            # (with --with-observability)
│       ├── health.go             # Health check endpoints
│       └── metrics.go            # Prometheus metrics
│
├── pkg/                          # Reusable packages (safe to import by others)
│   └── logger/
│       └── logger.go             # Structured logger interface + implementation
│
├── migrations/                   # (with --with-migrations)
│   └── .gitkeep
│
├── main.go                       # Minimal entry point — calls cmd.Execute()
├── Makefile                      # Build, test, lint, docker targets
├── Dockerfile                    # (with --with-docker) Multi-stage build
├── compose.yaml                  # (with --with-docker) Docker Compose services
├── .gitignore
└── README.md                     # Architecture documentation for the project
```

---

## Key Directories Explained

### `cmd/`

Contains [Cobra](https://github.com/spf13/cobra) commands:

- **`root.go`** — Initializes Viper configuration, reads `.my-app.yaml` and environment variables
- **`run.go`** — Starts the HTTP server with context-based graceful shutdown. Listens for SIGINT/SIGTERM.

### `internal/core/`

The innermost layer — **zero external dependencies allowed**.

| Directory | Contents |
|-----------|----------|
| `domain/` | Entities (`User`, `Order`), value objects (`Email`, `Money`), domain errors |
| `services/` | Use cases (`CreateUser`, `ProcessOrder`), port interface definitions |

### `internal/adapters/`

Connects the core to the outside world.

| Directory | Contents |
|-----------|----------|
| `primary/http/` | HTTP handlers (framework-specific) |
| `primary/grpc/` | gRPC handlers |
| `primary/queue/` | Message queue consumers |
| `secondary/database/` | Database repositories |
| `secondary/external/` | External API clients |
| `secondary/cache/` | Cache adapters |

### `pkg/`

Reusable packages that are safe to import by external projects. Currently contains the logger.

### `migrations/`

SQL migration files (when `--with-migrations` is used). Files follow the `golang-migrate` naming convention:

```
000001_create_users_table.up.sql
000001_create_users_table.down.sql
000002_add_email_index.up.sql
000002_add_email_index.down.sql
```

---

## DDD / Driver-Driven Naming Variant

When using `--adapter-style driver-driven --core-logic usecases --explicit-ports`:

```
my-app/
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   ├── usecases/             # (instead of services/)
│   │   └── ports/                # (with --explicit-ports) Interface definitions
│   └── adapters/
│       ├── driver/               # (instead of primary/)
│       └── driven/               # (instead of secondary/)
```

---

## After Adding Components

Running `hexago add` commands populates the structure:

```
my-app/
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── user.go           # hexago add domain entity User
│   │   │   ├── user_test.go
│   │   │   ├── email.go          # hexago add domain valueobject Email
│   │   │   └── email_test.go
│   │   └── services/
│   │       ├── create_user.go    # hexago add service CreateUser
│   │       └── create_user_test.go
│   └── adapters/
│       ├── primary/http/
│       │   └── user_handler.go   # hexago add adapter primary http UserHandler
│       └── secondary/database/
│           └── user_repository.go # hexago add adapter secondary database UserRepository
└── migrations/
    ├── 000001_create_users.up.sql    # hexago add migration create_users
    └── 000001_create_users.down.sql
```

---

## Configuration File

The generated project reads configuration from `.my-app.yaml` (where `my-app` is your project name):

```yaml
server:
  port: 8080
  readtimeout: 15s
  writetimeout: 15s
  shutdowntimeout: 30s

loglevel: info     # debug, info, warn, error
logformat: json    # json, text
```

All config values can be overridden with environment variables using the `MY_APP_` prefix:

```shell
export MY_APP_SERVER_PORT=9000
export MY_APP_LOGLEVEL=debug
```
