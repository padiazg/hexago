# hexago init

Create a new hexagonal architecture project.

## Synopsis

```shell
hexago init <name> [flags]
```

`<name>` is the project directory name that will be created.

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--module` | `-m` | string | *(required)* | Go module name (e.g. `github.com/user/my-app`) |
| `--framework` | `-f` | string | `stdlib` | Web framework: `echo`, `gin`, `chi`, `fiber`, or `stdlib` |
| `--adapter-style` | | string | `primary-secondary` | Adapter naming: `primary-secondary` or `driver-driven` |
| `--core-logic` | | string | `services` | Business logic directory: `services` or `usecases` |
| `--with-docker` | | bool | `false` | Generate Dockerfile and docker-compose |
| `--with-observability` | | bool | `false` | Include health checks and Prometheus metrics |
| `--with-migrations` | | bool | `false` | Include database migration setup |
| `--with-workers` | | bool | `false` | Include background worker pattern |
| `--with-metrics` | | bool | `false` | Include Prometheus metrics *(deprecated — use `--with-observability`)* |
| `--with-example` | | bool | `false` | Include example code |
| `--explicit-ports` | | bool | `false` | Create an explicit `ports/` directory |

!!! note
    All `--with-*` flags default to `false` (opt-in). This keeps generated projects lean — only include what you need.

---

## Framework Options

| Value | Handler Type |
|-------|-------------|
| `stdlib` | `http.Handler` (default) |
| `echo` | `func(echo.Context) error` |
| `gin` | `func(*gin.Context)` |
| `chi` | Standard library with chi router |
| `fiber` | `func(*fiber.Ctx) error` |

---

## Naming Convention Options

### `--adapter-style`

Controls how inbound and outbound adapter directories are named:

| Value | Inbound | Outbound |
|-------|---------|----------|
| `primary-secondary` | `adapters/primary/` | `adapters/secondary/` |
| `driver-driven` | `adapters/driver/` | `adapters/driven/` |

### `--core-logic`

Controls how the business logic directory is named:

| Value | Directory |
|-------|-----------|
| `services` | `internal/core/services/` |
| `usecases` | `internal/core/usecases/` |

---

## Examples

### Basic project with stdlib

```shell
hexago init my-app --module github.com/user/my-app
```

### With Echo framework

```shell
hexago init api-server --module github.com/user/api-server --framework echo
```

### Full-featured project

```shell
hexago init platform \
  --module github.com/company/platform \
  --framework gin \
  --with-docker \
  --with-observability \
  --with-migrations \
  --with-workers
```

### DDD / Ports-and-Adapters style naming

```shell
hexago init ordering \
  --module github.com/shop/ordering \
  --adapter-style driver-driven \
  --core-logic usecases \
  --explicit-ports
```

This creates `internal/core/usecases/`, `internal/adapters/driver/`, `internal/adapters/driven/`, and `internal/core/ports/`.

### Microservice (no HTTP, long-running daemon)

```shell
hexago init email-service \
  --module github.com/company/email-service \
  --with-workers \
  --with-migrations
```

---

## Generated Files

```
my-app/
├── cmd/
│   ├── root.go            # Root command + Viper config
│   └── run.go             # Server with graceful shutdown
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   └── services/
│   ├── adapters/
│   │   ├── primary/
│   │   │   └── http/
│   │   └── secondary/
│   │       └── database/
│   ├── config/
│   └── observability/     # (with --with-observability)
├── pkg/
│   └── logger/
├── main.go
├── Makefile
├── Dockerfile             # (with --with-docker)
├── compose.yaml           # (with --with-docker)
└── README.md
```
