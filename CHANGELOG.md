# Changelog

All notable changes to HexaGo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - v0.0.2

### Added

#### Template Management Commands (`hexago templates`)
- **`hexago templates list`**: lists all embedded templates grouped by directory; annotates overrides with `← project-local` or `← user-global`
- **`hexago templates which <name>`**: shows the winning source (embedded, project-local, user-global, or binary-local) with its full path
- **`hexago templates export <name> [--global]`**: copies a built-in template to `.hexago/templates/<name>` or `~/.hexago/templates/<name>` for customization
- **`hexago templates export-all [--global] [--force]`**: bulk-exports every embedded template at once; skips templates that already have an override unless `--force` is passed
- **`hexago templates validate <path>`**: parses a template file and reports `text/template` syntax errors — prints `✓` on success, `✗ <error>` on failure
- **`hexago templates reset <name> [--global]`**: removes a custom override, reverting to the next-priority source; errors clearly when no override exists
- `TemplateLoader.Validate(path string) error` and `TemplateLoader.Reset(name string, global bool) error` added to `internal/generator/template_loader.go`

#### `.hexago.yaml` Project Configuration File
- **`internal/generator/hexago_config.go`** (new): typed YAML structs (`HexagoConfig`, `HexagoProjectConfig`, `HexagoStructureConfig`, `HexagoFeaturesConfig`) plus four helpers:
  - `HexagoConfigFromProject(cfg)` — maps `ProjectConfig` → YAML struct
  - `(h) ToProjectConfig()` — maps YAML struct → `ProjectConfig`
  - `LoadHexagoConfig(dir)` — reads `{dir}/.hexago.yaml` with `gopkg.in/yaml.v3`
  - `SaveHexagoConfig(dir, cfg)` — writes `{dir}/.hexago.yaml` with a comment header
- **`hexago init` writes `.hexago.yaml`** into the generated project root after scaffolding, persisting all init-time settings (framework, adapter style, features, etc.) that could not be recovered from the filesystem alone
- **`hexago add *` reads `.hexago.yaml` first**: `DetectConfig()` in `detector.go` now tries `LoadHexagoConfig` before falling back to filesystem heuristics — giving every `add` command access to the full original config (including `Framework`, `ProjectType`, `Author`, `GoVersion`, feature flags)
- **`hexago init` honours `.hexago.yaml` as a defaults layer**: priority is `flags > .hexago.yaml > hardcoded defaults`. Any flag not explicitly passed on the command line is filled from a `.hexago.yaml` found in the current working directory, enabling a personal or team-wide preferences file without forcing every flag on every invocation. Uses Cobra's `cmd.Flags().Changed()` to distinguish user-supplied flags from default values
- `gopkg.in/yaml.v3` added as a direct dependency

#### HTTP Server Interface Pattern
- **Shared `Server` interface** in `pkg/server/server.go`: a single `Run(errChan chan<- error)` / `Stop(ctx context.Context) error` contract lives in a public, framework-agnostic package instead of being re-declared in every adapter
- **`http_server_interface.go.tmpl`**: new template that generates `pkg/server/server.go` for every `http-server` project
- **Compile-time interface guard** in every framework adapter: `var _ srv.Server = (*server)(nil)` catches implementation drift at build time, not at runtime

#### HTTP Server Adapter Refactoring
- **Framework-specific `server.go`** files extracted from `cmd/run.go` into `internal/adapters/{primary|driver}/http/server.go` for all five supported frameworks (Echo, Gin, Chi, Fiber, stdlib):
  - Framework instance creation, middleware wiring, and `http.Server` configuration are now encapsulated inside each adapter
  - `setupRoutes` promoted from a package-level function to a method on `*server`, giving it direct access to the framework instance without parameter passing
  - Each adapter's `New()` constructor returns `srv.Server` (the shared interface), hiding all framework types behind the abstraction boundary
- **Thin `cmd/run.go` orchestrator**: the run command is now completely framework-agnostic — it only calls `httpserver.New()`, `srv.Run()`, and `srv.Stop()`. No framework imports, no repeated signal/shutdown boilerplate per framework

### Changed
- `cmd/run.go` (generated) no longer contains `setupRoutes` or any web-framework imports
- `internal/adapters/{inbound}/http/server.go` (generated) now owns all framework-specific lifecycle code
- `pkg/server/server.go` (generated, new) is the single source of truth for the `Server` interface contract

### Refactored (internal — no generated-code change)

#### Remove global template loader singleton
- **`globalTemplateLoader` package-level variable and `init()` removed** from `internal/generator/templates.go`
- `TemplateLoader` is now a field (`templateLoader *TemplateLoader`) on `ProjectConfig`, initialized in `NewProjectConfig()`
- All generator methods that previously called `globalTemplateLoader.Render(...)` now call `g.config.templateLoader.Render(...)` — scoping the loader to its owning config and making generators straightforward to test in isolation

#### New `pkg/utils` package
- `pkg/utils/case.go` added with two exported helpers:
  - `ToSnakeCase(s string) string` — converts CamelCase identifiers to snake_case file names
  - `ToTitleCase(s string) string` — uppercases the first letter of a string
- Eliminates at least three identical local `toSnakeCase` copies that existed independently in `service.go`, `tool.go`, `worker.go`, `domain.go`, `adapter.go`, and `cmd/add_tool.go`
- `createTemplateFuncMap()` in `template_loader.go` now references `utils.ToSnakeCase` and `utils.ToTitleCase` for the `"snake"` and `"title"` template functions

#### Extended `pkg/fileutil`
- `HomeDir() string` and `BinaryDir() string` migrated from `internal/generator/template_loader.go` into `pkg/fileutil/fileutil.go`
- `template_loader.go` now uses `fileutil.HomeDir()`, `fileutil.BinaryDir()`, and `fileutil.FileExists` — removing three private helper functions from the generator package

---

## [0.0.1] - 2026-02-17

### Added - MVP Release

#### Core Features
- **Project Type Support**: Generate projects with different architectural patterns
  - `http-server`: HTTP API server with framework support (Echo, Gin, Chi, Fiber, stdlib)
  - `service`: Long-running daemon/service with no web framework for main logic
- **Hexagonal Architecture**: Strict separation of concerns with core/adapters structure
- **Framework Support**: Echo, Gin, Chi, Fiber, and Go stdlib for HTTP servers
- **Graceful Shutdown**: Context-based cancellation with signal handling for all project types
- **Configuration Management**: Viper-based config with YAML files and environment variable support
- **Structured Logging**: Logger package with configurable levels and formats

#### Observability (Available for All Project Types)
- **Health Checks**:
  - `/health` - Complete health report with component status
  - `/health/ready` - Kubernetes readiness probe
  - `/health/live` - Kubernetes liveness probe
- **Prometheus Metrics**: Request counters, latency histograms, active operations gauge
- **Separate Observability Server**: Runs on independent port (default: 8080)
- **Component Registration**: Register custom health checks for databases, queues, etc.

#### Service Pattern (Long-Running Daemon)
- **Processor Pattern**: Main business logic in `Processor.Start(ctx)` method
- **Context-Based Shutdown**: Clean cancellation and resource cleanup
- **Background Processing**: Example implementations for queues, schedulers, file watchers
- **Signal Handling**: SIGINT, SIGTERM, SIGQUIT support
- **Configurable Timeouts**: Grace period for shutdown operations

#### Template System
- **Externalized Templates**: All code templates can be customized
- **Multi-Source Loading**:
  - Binary-local: `templates/` (next to executable)
  - Project-local: `.hexago/templates/` (per-project customization)
  - User-global: `~/.hexago/templates/` (user-wide customization)
  - Embedded: Fallback templates compiled into binary
- **Company Branding**: Easy to customize headers, comments, and code style
- **Version Control**: Share custom templates across teams

#### Code Generation
- **Component Generators**:
  - Services/UseCases: Business logic layer
  - Domain Entities: Core domain objects with fields
  - Value Objects: Immutable domain values
  - HTTP Adapters: Framework-specific handlers
  - Database Adapters: Repository implementations
  - External Service Adapters: API client wrappers
  - Cache Adapters: Redis/memory cache implementations
  - Queue Adapters: Message queue consumers
- **Background Workers**: Queue-based, periodic, and event-driven patterns
- **Database Migrations**: Sequential numbered migrations with golang-migrate support
- **Infrastructure Tools**: Loggers, validators, mappers, middleware

#### Project Flexibility
- **Optional Features**: All features opt-in via flags (default: false)
  - `--with-docker`: Docker files (Dockerfile, compose.yaml)
  - `--with-observability`: Health checks and metrics
  - `--with-migrations`: Database migration setup
  - `--with-workers`: Background worker pattern
  - `--with-metrics`: Prometheus metrics (deprecated, use --with-observability)
  - `--with-example`: Example code
  - `--explicit-ports`: Explicit ports/ directory structure
- **Naming Conventions**:
  - Adapter style: `primary-secondary` or `driver-driven`
  - Core logic: `services` or `usecases`
- **Architecture Validation**: Auto-detection of existing project conventions

#### Developer Experience
- **Cobra CLI**: Command structure with subcommands
- **Auto-Detection**: Respects existing project structure and conventions
- **Smart Defaults**: Sensible defaults with override options
- **Helpful Messages**: Clear error messages and configuration summaries
- **Educational Comments**: Generated code includes architecture guidance

#### Build & Release
- **GoReleaser Integration**: Automated multi-platform builds
- **GitHub Actions**: CI/CD workflow for releases
- **Platform Support**:
  - Linux: x86_64, arm64
  - macOS: x86_64 (Intel), arm64 (Apple Silicon)
- **Static Binaries**: CGO_ENABLED=0 for portability
- **Homebrew Support**: Ready for homebrew-tap publication

### Documentation
- Comprehensive README with examples
- Quick start guide
- Architecture documentation
- Template customization guide
- Project type comparison

### Project Types Use Cases

#### HTTP Server (`http-server`)
Perfect for:
- REST APIs
- GraphQL servers
- Microservices with HTTP interfaces
- Web applications with API backends

#### Service (`service`)
Perfect for:
- MQTT/Kafka message consumers
- File system watchers
- Background job processors
- Event stream processors
- Periodic task schedulers
- Data pipeline processors

### Breaking Changes
None (initial release)

### Security
- No external dependencies in core (stdlib only)
- Static binary compilation
- No code execution from templates (text/template, not html/template)

---

## How to Update

```bash
go install github.com/padiazg/hexago@v0.0.1
```

Or download binaries from [GitHub Releases](https://github.com/padiazg/hexago/releases/tag/v0.0.1)

[0.0.1]: https://github.com/padiazg/hexago/releases/tag/v0.0.1
