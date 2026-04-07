# Changelog

All notable changes to HexaGo will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## v0.1.3 - 2026-04-06

### Version command included in base generation

All generated projects now include a `version` command with ASCII art splash output:

```shell
$ myapp version
┓┏      ┏┓    Version: 1.0.0
┣┫┏┓┓┏┏┓┃┓┏┓  Build: 2026-04-06T12:00:00Z
┛┗┗ ┛┗┗┻┗┛┗┛  Commit: abc1234
```

The `--simple` flag outputs just the version string for scripting:

```shell
$ myapp version --simple
v1.0.0
```

Version info is injected at build time via Makefile ldflags:

```makefile
build: pkg={{.ModuleName}}/pkg/version
build: ldflags = -X $(pkg).version=$(shell git describe --tags --always --dirty)
build: ldflags += -X $(pkg).commit=$(shell git rev-parse HEAD)
build: ldflags += -X $(pkg).buildDate=$(shell date -Iseconds)
```

### Handler plugin pattern (`Use(ServerHandler) Server`)

The `Server` interface in `pkg/server/server.go` now exposes a fluent `Use(ServerHandler) Server` method. Any type that implements `ServerHandler` (a single `Configure(Server)` method) can be registered on the server — each handler mounts its own routes when called, enabling self-contained, isolated handler packages.

### `pkg/httpserver` — exported framework server

Framework-specific server implementations moved to `pkg/httpserver/`. Each server exposes its underlying router/engine as a public field so handlers can register routes directly:

| Framework | Public field |
|-----------|-------------|
| chi | `Server.Router chi.Router` |
| echo | `Server.Echo *echo.Echo` |
| gin | `Server.Router *gin.Engine` |
| fiber | `Server.App *fiber.App` |
| stdlib | `Server.Mux *http.ServeMux` |

### Observability integrated into main server

Health checks and Prometheus metrics are now registered as `ServerHandler` instances on the **main HTTP server** — no separate port, no extra process. The dedicated `observability.Server` (and its CLI flags `--observability` / `--observability-addr`) has been removed.

### Isolated route handler packages

Each route group lives in its own sub-package inside `internal/adapters/{inbound}/http/`:

- `ping/` — `/ping`
- `health/` — `/health`, `/health/ready`, `/health/live` (with `--with-observability`)
- `metrics/` — `/metrics` (with `--with-observability`)

A new wiring file `internal/adapters/{inbound}/http/http.go` creates the server and registers all handlers, keeping `cmd/run.go` fully framework-agnostic.

### Template directory restructured

Template paths now mirror the generated project structure for intuitive discovery. The `//go:embed` directive was changed from `templates/**/*.tmpl` to `//go:embed templates` to support deeply nested subdirectories.

```shell
hexago templates list   # shows the updated layout
```

### Route groups with middleware examples in HTTP adapter templates

All five HTTP adapter templates now include a commented `/api/v1` route group with route-scoped middleware examples (request-id injection, logging, panic recovery, authorization):

| Framework | Group mechanism |
|-----------|----------------|
| chi | `router.Route("/api/v1", func(r chi.Router) { r.Use(...) })` |
| echo | `v1 := srv.Echo.Group("/api/v1")` + `v1.Use(...)` |
| fiber | `v1 := srv.App.Group("/api/v1")` + `v1.Use(...)` |
| gin | `v1 := srv.Router.Group("/api/v1")` + `v1.Use(...)` |
| stdlib | nested `http.NewServeMux()` mounted via `http.StripPrefix("/api/v1", ...)` |

For stdlib, per-group middlewares are applied by wrapping the sub-mux before mounting it on the main `ServeMux`.

### Cross-platform embed fix

Embedded FS path lookups in `template_loader.go` changed from `filepath.Join` to `path.Join` — `embed.FS` always uses forward slashes; `filepath.Join` would fail on Windows.

### Template code style (`interface{}` → `any`)

All generated code templates use the `any` type alias (Go 1.18+) instead of `interface{}` — adapter, tool, worker, observability, and project templates updated for consistency with modern Go style.

### Enhanced service generation

Service generation now distinguishes between entity-bound services (requiring repository dependencies) and standalone services. Entity-bound services generate CRUD methods (Create, GetByID, Update, List), while standalone services generate an Execute method for custom business logic. The services aggregator has been updated to correctly handle both service types during initialization.

---

## v0.0.3 - 2026-03-04

### `--working-directory` global flag

- New **`-w` / `--working-directory`** persistent flag on the root command — every subcommand can now target a project in a different directory without `cd`-ing into it first.
- All `hexago add *` and `hexago validate` commands accept the flag and pass it to the project detector. When omitted, the current working directory is used as before.
- `hexago init --working-directory <dir>` creates the project under `<dir>` instead of the current directory.

```shell
# Add a service to a project located elsewhere — no cd required
hexago add service CreateUser --working-directory /home/user/projects/my-api
```

### `--in-place` flag for `hexago init`

- New **`--in-place`** bool flag: generates project files directly into `working_directory` instead of creating a `<name>` subdirectory inside it.
- Useful when the target directory already exists and is the intended project root (e.g. a freshly cloned empty repo or the current working directory).

```shell
# Scaffold into the current directory
hexago init my-api --module github.com/user/my-api --in-place

# Scaffold into an existing remote directory
hexago init my-api --module github.com/user/my-api \
  --working-directory /home/user/projects/my-api \
  --in-place
```

### Built-in MCP Server (`hexago mcp`)

HexaGo now ships with a built-in [Model Context Protocol](https://modelcontextprotocol.io/) server, letting AI assistants scaffold hexagonal architecture projects without leaving the conversation.

```shell
hexago mcp   # start the stdio MCP server
```

Register with Claude Code:

```shell
claude mcp add --scope project hexago -- hexago mcp
```

**Nine tools** are available — each delegates to the hexago CLI with `--working-directory`:

| Tool | What it does |
|------|-------------|
| `hexago_init` | Bootstrap a new project |
| `hexago_add_service` | Add a business-logic service |
| `hexago_add_domain_entity` | Add a domain entity |
| `hexago_add_domain_valueobject` | Add a domain value object |
| `hexago_add_adapter` | Add a primary or secondary adapter |
| `hexago_add_worker` | Add a background worker |
| `hexago_add_migration` | Add a database migration |
| `hexago_add_tool` | Add an infrastructure utility |
| `hexago_validate` | Validate architecture compliance |

The server delivers comprehensive usage instructions on every `initialize` handshake, covering all parameter names, valid enum values, defaults, field format, and a directive that prevents AI agents from falling back to raw CLI shell calls.

See [`hexago mcp`](commands/mcp.md) for client configuration examples (Claude Code, Claude Desktop, VS Code, Cursor, Windsurf, Zed).

### Changed

- `GetCurrentProjectConfig()` signature updated to `GetCurrentProjectConfig(dir string)` — empty string falls back to `os.Getwd()`. All call sites updated.
- `cmd/init.go` resolves `OutputDir` from the `--working-directory` flag value with an `os.Getwd()` fallback.
- `internal/generator/project.go` and `detector.go` migrated from `pkg/fileutil` to `pkg/utils` for file-system helpers (internal refactor, no behaviour change).

---

## [0.0.2] — 2026-02-26

!!! success "Release Highlights"
    Template management commands, project config file, and cleaner generated HTTP server architecture.

### Template Management (`hexago templates`)

Full control over the templates HexaGo uses to generate your projects:

- **`hexago templates list`** — Lists all built-in templates grouped by directory. Templates with an active override are annotated with `← project-local` or `← user-global`.
- **`hexago templates which <name>`** — Shows which source wins for a given template (embedded, project-local, user-global, or binary-local) with its full path.
- **`hexago templates export <name> [--global]`** — Copies a built-in template to `.hexago/templates/<name>` (project-local) or `~/.hexago/templates/<name>` (user-global) for customization.
- **`hexago templates export-all [--global] [--force]`** — Bulk-exports every embedded template at once; skips templates that already have an override unless `--force` is passed.
- **`hexago templates validate <path>`** — Parses a template file and reports `text/template` syntax errors. Prints `✓` on success, `✗ <error>` on failure.
- **`hexago templates reset <name> [--global]`** — Removes a custom override, reverting to the next-priority source.

See [Template Customization](customization/templates.md) for full details.

### `.hexago.yaml` Project Configuration File

- **`hexago init` now writes `.hexago.yaml`** into the generated project root after scaffolding, persisting all init-time settings (framework, adapter style, features, etc.).
- **`hexago add *` reads `.hexago.yaml` automatically** — no need to repeat flags on every invocation. Settings detected from the config file supplement filesystem heuristics.
- **Acts as a defaults layer** — priority is `flags > .hexago.yaml > hardcoded defaults`. Any flag not explicitly passed is filled from `.hexago.yaml`, enabling personal or team-wide preferences.
- Useful for sharing consistent conventions across a team without enforcing every flag.

### HTTP Server Architecture (Generated Code)

- **Shared `Server` interface** in `pkg/server/server.go` — a single `Run(errChan chan<- error)` / `Stop(ctx context.Context) error` contract shared across all adapters.
- **Framework-specific `server.go`** extracted into `internal/adapters/{primary|driver}/http/server.go` for all five supported frameworks (Echo, Gin, Chi, Fiber, stdlib). Each adapter's `New()` constructor returns the shared `srv.Server` interface, hiding all framework types behind the abstraction boundary.
- **Thin `cmd/run.go` orchestrator** — now completely framework-agnostic: no framework imports, no repeated signal/shutdown boilerplate. Just calls `httpserver.New()`, `srv.Run()`, and `srv.Stop()`.
- **Compile-time interface guards** (`var _ srv.Server = (*server)(nil)`) catch implementation drift at build time.

### Refactored (Internal — No Generated-Code Change)

- **Removed global template loader singleton** — `TemplateLoader` is now a field on `ProjectConfig`, scoping it to its owning config and making generators straightforward to test in isolation.
- **New `pkg/utils` package** — `ToSnakeCase` and `ToTitleCase` helpers replace multiple identical local copies across the generator.
- **Observability templates moved** — `misc/health.go.tmpl`, `misc/metrics.go.tmpl`, and `misc/server.go.tmpl` relocated to `observability/` to match the generated `internal/observability/` package structure.
- **Extended `pkg/fileutil`** — `HomeDir()` and `BinaryDir()` migrated from the generator package into `pkg/fileutil`, removing private helpers from the generator.

---

## [0.0.1] — 2026-02-17

!!! success "MVP Release"
    Initial public release of HexaGo.

### Core Features

- **Project Type Support**: Generate projects with different architectural patterns
    - `http-server` — HTTP API server with framework support (Echo, Gin, Chi, Fiber, stdlib)
    - `service` — Long-running daemon/service with no web framework for main logic
- **Hexagonal Architecture** — Strict separation of concerns with core/adapters structure
- **Framework Support** — Echo, Gin, Chi, Fiber, and Go stdlib for HTTP servers
- **Graceful Shutdown** — Context-based cancellation with signal handling for all project types
- **Configuration Management** — Viper-based config with YAML files and environment variable support
- **Structured Logging** — Logger package with configurable levels and formats

### Observability

- **Health Checks**:
    - `/health` — Complete health report with component status
    - `/health/ready` — Kubernetes readiness probe
    - `/health/live` — Kubernetes liveness probe
- **Prometheus Metrics** — Request counters, latency histograms, active operations gauge
- **Separate Observability Server** — Runs on independent port (default: 8080)
- **Component Registration** — Register custom health checks for databases, queues, etc.

### Service Pattern (Long-Running Daemon)

- **Processor Pattern** — Main business logic in `Processor.Start(ctx)` method
- **Context-Based Shutdown** — Clean cancellation and resource cleanup
- **Background Processing** — Example implementations for queues, schedulers, file watchers
- **Signal Handling** — SIGINT, SIGTERM, SIGQUIT support
- **Configurable Timeouts** — Grace period for shutdown operations

### Template System

- **Externalized Templates** — All code templates can be customized
- **Multi-Source Loading**:
    - Binary-local: `templates/` (next to executable)
    - Project-local: `.hexago/templates/` (per-project customization)
    - User-global: `~/.hexago/templates/` (user-wide customization)
    - Embedded: Fallback templates compiled into binary
- **Company Branding** — Easy to customize headers, comments, and code style
- **Version Control** — Share custom templates across teams

### Code Generation

- **Component Generators**:
    - Services/UseCases — Business logic layer
    - Domain Entities — Core domain objects with fields
    - Value Objects — Immutable domain values
    - HTTP Adapters — Framework-specific handlers
    - Database Adapters — Repository implementations
    - External Service Adapters — API client wrappers
    - Cache Adapters — Redis/memory cache implementations
    - Queue Adapters — Message queue consumers
- **Background Workers** — Queue-based, periodic, and event-driven patterns
- **Database Migrations** — Sequential numbered migrations with golang-migrate support
- **Infrastructure Tools** — Loggers, validators, mappers, middleware

### Project Flexibility

- **Optional Features** — All features opt-in via flags (default: false)
    - `--with-docker` — Docker files (Dockerfile, compose.yaml)
    - `--with-observability` — Health checks and metrics
    - `--with-migrations` — Database migration setup
    - `--with-workers` — Background worker pattern
    - `--with-metrics` — Prometheus metrics *(deprecated, use `--with-observability`)*
    - `--with-example` — Example code
    - `--explicit-ports` — Explicit ports/ directory structure
- **Naming Conventions**:
    - Adapter style: `primary-secondary` or `driver-driven`
    - Core logic: `services` or `usecases`
- **Architecture Validation** — Auto-detection of existing project conventions

### Developer Experience

- **Cobra CLI** — Command structure with subcommands
- **Auto-Detection** — Respects existing project structure and conventions
- **Smart Defaults** — Sensible defaults with override options
- **Helpful Messages** — Clear error messages and configuration summaries
- **Educational Comments** — Generated code includes architecture guidance

### Build & Release

- **GoReleaser Integration** — Automated multi-platform builds
- **GitHub Actions** — CI/CD workflow for releases
- **Platform Support**:
    - Linux: x86_64, arm64
    - macOS: x86_64 (Intel), arm64 (Apple Silicon)
- **Static Binaries** — `CGO_ENABLED=0` for portability
- **Homebrew Support** — Ready for homebrew-tap publication

### Documentation

- Comprehensive README with examples
- Quick start guide
- Architecture documentation
- Template customization guide
- Project type comparison

### Project Types Use Cases

**HTTP Server (`http-server`)** — Perfect for:

- REST APIs
- GraphQL servers
- Microservices with HTTP interfaces
- Web applications with API backends

**Service (`service`)** — Perfect for:

- MQTT/Kafka message consumers
- File system watchers
- Background job processors
- Event stream processors
- Periodic task schedulers
- Data pipeline processors

### Security

- No external dependencies in core (stdlib only)
- Static binary compilation
- No code execution from templates (`text/template`, not `html/template`)

---

## How to Update

```shell
go install github.com/padiazg/hexago@v0.0.2
```

Or download binaries from [GitHub Releases](https://github.com/padiazg/hexago/releases/tag/v0.0.2).

[0.0.2]: https://github.com/padiazg/hexago/releases/tag/v0.0.2
[0.0.1]: https://github.com/padiazg/hexago/releases/tag/v0.0.1
