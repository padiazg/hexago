# Changelog

All notable changes to HexaGo will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
go install github.com/padiazg/hexago@v0.0.1
```

Or download binaries from [GitHub Releases](https://github.com/padiazg/hexago/releases/tag/v0.0.1).

[0.0.1]: https://github.com/padiazg/hexago/releases/tag/v0.0.1
