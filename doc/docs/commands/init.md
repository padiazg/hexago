# hexago init

Create a new hexagonal architecture project.

## Synopsis

```shell
hexago init <name> [flags]
```

`<name>` is the project directory name that will be created.

---

## Flags

### Global flags (inherited from root)

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--working-directory` | `-w` | string | *(current dir)* | Parent directory where the project will be created. Defaults to `os.Getwd()` when not supplied. |

### Init-specific flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--module` | `-m` | string | *(project name)* | Go module name (e.g. `github.com/user/my-app`). Defaults to the project name if omitted. |
| `--project-type` | `-t` | string | `http-server` | Project type: `http-server` or `service` |
| `--framework` | `-f` | string | `stdlib` | Web framework for `http-server`: `echo`, `gin`, `chi`, `fiber`, or `stdlib` |
| `--adapter-style` | | string | `primary-secondary` | Adapter naming: `primary-secondary` or `driver-driven` |
| `--core-logic` | | string | `services` | Business logic directory: `services` or `usecases` |
| `--in-place` | | bool | `false` | Generate files directly into `working_directory` — no `<name>` subdirectory is created. |
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

## Project Types

| Value | Description |
|-------|-------------|
| `http-server` | HTTP API server with a web framework (default) |
| `service` | Long-running daemon or background service (no web framework required) |

The `--framework` flag is only relevant for `http-server` projects. Specifying `--framework` with `--project-type service` emits a warning and is ignored.

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

### Long-running service (no HTTP)

```shell
hexago init email-service \
  --module github.com/company/email-service \
  --project-type service \
  --with-workers \
  --with-migrations
```

### Scaffold into a specific parent directory (no `cd` required)

```shell
hexago init my-api \
  --module github.com/user/my-api \
  --working-directory /home/user/projects
# → creates /home/user/projects/my-api/
```

### Scaffold into an existing directory (`--in-place`)

Use `--in-place` when you are already inside the intended project root, or when
targeting a pre-existing directory (e.g. a freshly cloned empty repo):

```shell
# Current directory becomes the project root — no subfolder created
hexago init my-api --module github.com/user/my-api --in-place

# Remote directory, in-place
hexago init my-api \
  --module github.com/user/my-api \
  --working-directory /home/user/projects/my-api \
  --in-place
# → files go directly into /home/user/projects/my-api/
```

!!! tip
    Without `--in-place`, the project is always placed at `<working-directory>/<name>/`.
    With `--in-place`, it is placed directly at `<working-directory>/`.

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
│   │   │       └── server.go  # Framework-specific lifecycle
│   │   └── secondary/
│   │       └── database/
│   ├── config/
│   └── observability/     # (with --with-observability)
├── pkg/
│   ├── logger/
│   └── server/
│       └── server.go      # Shared Server interface
├── main.go
├── Makefile
├── Dockerfile             # (with --with-docker)
├── compose.yaml           # (with --with-docker)
├── .hexago.yaml           # HexaGo project configuration
└── README.md
```

---

## Project Configuration File (`.hexago.yaml`)

After scaffolding, `hexago init` writes a `.hexago.yaml` file into the project root. This file persists all settings chosen at init time — framework, adapter style, feature flags, etc.

```yaml
# .hexago.yaml (example)
project:
  name: my-app
  module: github.com/user/my-app
  type: http-server
  framework: echo
structure:
  adapter_style: primary-secondary
  core_logic: services
  explicit_ports: false
features:
  docker: false
  observability: false
  migrations: false
  workers: false
```

All `hexago add *` commands read this file automatically — you do not need to pass framework or convention flags on every invocation.

The file also acts as a **defaults layer** when running `hexago init` in a directory that already contains one:

```
flag value  >  .hexago.yaml  >  hardcoded defaults
```

This makes it easy to keep a personal or team-wide preferences file without forcing every flag on every invocation.
