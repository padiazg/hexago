# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HexaGo is a CLI tool written in Go that generates scaffolding for applications following the Hexagonal Architecture (Ports & Adapters) pattern. It's an opinionated code generator that enforces architectural boundaries and helps developers avoid conceptual confusion when building Go applications.

**Key Purpose**: Generate production-ready Go projects with proper hexagonal architecture separation, including:
- Framework support (Echo, Gin, Chi, Fiber, stdlib)
- Docker setup with multi-stage builds
- Graceful shutdown with context-based cancellation
- Background workers using goroutines and channels
- Database migrations (golang-migrate)
- OpenAPI/Swagger documentation

## Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra (github.com/spf13/cobra) - used as library, not CLI tool
- **Configuration**: Viper (github.com/spf13/viper)
- **Template Engine**: Go's `text/template` with `go:embed` for portability
- **Testing**: Go standard testing + testify

## Code Generation Strategy

**Important**: HexaGo generates all files directly using embedded templates. It does NOT use the `cobra-cli` tool.

**Why not use cobra-cli?**
- cobra-cli is rarely installed by Go developers
- We need custom hexagonal architecture from the start
- Cobra boilerplate is simple enough to template ourselves
- Provides better user experience (single command, no prerequisites)
- Full control over generated structure
- Easier to test and maintain

See `IMPLEMENTATION_STRATEGY.md` for detailed implementation approach.

## Project Structure (HexaGo Tool Itself)

**Important: Cobra CLI Pattern**
- `main.go` - Minimal entry point that only calls `cmd.Execute()`
- `cmd/` directory contains all Cobra command implementations
- Each command is registered via `init()` function with `rootCmd.AddCommand()`

```
hexago/
├── main.go                 # Entry point - calls cmd.Execute() only
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command, Execute() func, config init
│   ├── init.go            # Init command (project initialization)
│   ├── add.go             # Add command (parent command)
│   ├── add_usecase.go     # Add usecase subcommand
│   ├── add_domain.go      # Add domain subcommand
│   ├── add_driven.go      # Add driven adapter subcommand
│   ├── add_driver.go      # Add driver adapter subcommand
│   ├── add_worker.go      # Add worker subcommand
│   ├── add_migration.go   # Add migration subcommand
│   ├── validate.go        # Validate command
│   ├── diagram.go         # Diagram command
│   └── version.go         # Version command
├── internal/
│   ├── generator/         # Code generation logic
│   │   ├── generator.go  # Main generator interface
│   │   ├── project.go    # Project initialization
│   │   ├── usecase.go    # Use case generation
│   │   ├── driven.go     # Driven adapter generation
│   │   ├── driver.go     # Driver adapter generation
│   │   ├── domain.go     # Domain entity generation
│   │   ├── worker.go     # Worker generation
│   │   └── tool.go       # Tool/utility generation
│   ├── templates/        # Embedded templates (go:embed)
│   │   ├── project/      # Initial project templates
│   │   ├── cobra/        # Cobra CLI templates
│   │   ├── usecase/      # Use case templates
│   │   ├── driven/       # Driven adapter templates
│   │   ├── driver/       # Driver adapter templates
│   │   ├── domain/       # Domain entity templates
│   │   ├── worker/       # Worker templates
│   │   └── tool/         # Tool templates
│   ├── validator/        # Structure validation
│   │   └── validator.go
│   └── config/           # Configuration management
│       └── config.go
├── pkg/
│   └── fileutil/         # File system utilities
│       └── fileutil.go
├── go.mod
└── README.md
```

## Generated Project Structure

HexaGo generates projects with strict hexagonal architecture following the Cobra CLI pattern:

**Critical: Entry Point Pattern**
- `main.go` - Minimal, only calls `cmd.Execute()`
- `cmd/root.go` - Defines root command, configuration loading, and Execute() function
- `cmd/run.go` - Contains actual server/application logic with graceful shutdown
- Other `cmd/*.go` - Additional subcommands (migrate, seed, etc.)

```
generated-project/
├── main.go                      # Entry point - calls cmd.Execute() only!
├── cmd/                         # Cobra commands
│   ├── root.go                 # Root command, Execute(), config init
│   ├── run.go                  # Server/app logic with graceful shutdown
│   ├── migrate.go              # Database migration command
│   └── seed.go                 # Database seeding command (optional)
├── internal/
│   ├── core/                   # CORE - No external dependencies
│   │   ├── domain/            # Business entities
│   │   ├── ports/             # Port interfaces (optional)
│   │   │   ├── driven/        # Outbound ports (repos, services)
│   │   │   └── driver/        # Inbound ports (handlers)
│   │   └── services/          # Business logic (or usecases/)
│   ├── adapters/              # ADAPTERS - Implement ports
│   │   ├── primary/           # Inbound (can use driver/ instead)
│   │   │   ├── http/          # HTTP handlers
│   │   │   ├── grpc/          # gRPC handlers
│   │   │   └── queue/         # Message queue consumers
│   │   └── secondary/         # Outbound (can use driven/ instead)
│   │       ├── database/      # Database repositories
│   │       └── external/      # External service clients
│   ├── config/                # Configuration (Viper)
│   │   └── config.go
│   ├── observability/         # Health checks, metrics
│   │   ├── health.go
│   │   └── metrics.go
│   └── di/                    # Dependency injection
│       └── container.go
├── pkg/                       # Public reusable packages
│   ├── logger/               # Logger package
│   └── utils/                # Utilities
└── migrations/               # Database migrations
```

**Note**: Generated projects can use either terminology:
- `adapters/primary` and `adapters/secondary` (more common in DDD)
- `adapters/driver` and `adapters/driven` (ports & adapters terminology)
- `core/services` or `core/usecases` (both valid)
- Ports can be explicit (in `ports/` directory) or implicit (interfaces where needed)

## Development Commands

### Building and Running HexaGo CLI
```bash
# Build the CLI tool
go build -o hexago main.go
# or with CGO disabled for static binary
CGO_ENABLED=0 go build -o hexago main.go

# Run directly (development)
go run main.go [command]
# Example:
go run main.go init my-app --module github.com/user/my-app --framework echo

# Install locally to $GOPATH/bin
go install

# Run installed version
hexago init my-app --module github.com/user/my-app --framework echo
```

### Running Generated Projects
Generated projects follow the Cobra CLI pattern:
```bash
# Build the generated application
go build -o myapp

# Run the server (main command)
./myapp run

# Run with flags
./myapp run --config .myapp.yaml --observability

# Other commands
./myapp migrate up
./myapp create --config .myapp.yaml  # Create example config
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/generator/...

# Run a specific test
go test ./pkg/utils -run TestSpecificFunction
```

### Code Quality
```bash
# Format code
go fmt ./...
gofmt -s -w .

# Run linter (if golangci-lint is installed)
golangci-lint run ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

### Makefile Pattern for Generated Projects
Generated projects should include a Makefile with common tasks:
```makefile
.PHONY: build clean test build-image

SERVICE_NAME=myapp

build:
	@echo "Building $(SERVICE_NAME)..."
	@CGO_ENABLED=0 go build -o $(SERVICE_NAME)

clean:
	@echo "Cleaning..."
	@rm -f $(SERVICE_NAME) coverage.out

test:
	@echo "Testing..."
	@go test -coverprofile=coverage.out -cover ./...

build-image:
	@echo "Building Docker image..."
	@docker build -t $(SERVICE_NAME):latest .
```

## Key Architectural Principles

### Hexagonal Architecture Rules (For Generated Projects)

1. **Core Never Depends on Adapters**: The `internal/core` package must never import from `internal/adapters`

2. **Dependency Direction**: Always inward
   - Adapters → Ports (interfaces) → Services/UseCases → Domain
   - Never: Domain → Services, Services → Adapters

3. **Adapter Terminology** (both valid, be consistent):
   - **Primary/Secondary** (DDD terminology):
     - `adapters/primary/` - Inbound (HTTP, gRPC, CLI, queue consumers)
     - `adapters/secondary/` - Outbound (database, external APIs)
   - **Driver/Driven** (Ports & Adapters terminology):
     - `adapters/driver/` - Inbound (driving the application)
     - `adapters/driven/` - Outbound (driven by the application)

4. **Ports as Contracts**:
   - Can be explicit (in `core/ports/` directory) OR
   - Can be implicit (interfaces defined where needed)
   - Ports define what the core needs, adapters implement them

5. **Services vs UseCases**:
   - Both terms are valid for `internal/core/` business logic
   - `services/` - More common in DDD
   - `usecases/` - More explicit about use case driven design
   - Choose one and be consistent

6. **Framework Agnostic Core**: Business logic should work without any framework

### Template Design Guidelines

- Use `text/template` for code generation
- Embed templates with `go:embed` directive
- Support template variables: `{{.ProjectName}}`, `{{.ModuleName}}`, `{{.ComponentName}}`, etc.
- Use snake_case for generated file names
- Include educational comments in generated code explaining architectural boundaries

### Code Generation Best Practices

- Always validate inputs before generation
- Check for existing files and prompt before overwriting (unless `--force`)
- Generated code must pass `go vet` and `gofmt`
- Include TODO comments for implementation guidance
- Add inline documentation explaining which layer/component and what dependencies are allowed

## CLI Command Structure

HexaGo itself uses Cobra for its CLI structure. Each command is implemented following the same pattern it generates.

### HexaGo Commands (the tool itself)

```bash
hexago init <project-name>           # Initialize new project
hexago add usecase <name>            # Add use case/service
hexago add driven <type> <name>      # Add driven/secondary port + adapter
hexago add driver <type> <name>      # Add driver/primary adapter
hexago add domain entity <name>      # Add domain entity
hexago add domain valueobject <name> # Add value object
hexago add tool <type> <name>        # Add infrastructure tool
hexago add worker <name>             # Add background worker
hexago add migration <name>          # Add database migration
hexago validate                      # Validate architecture
hexago diagram                       # Generate architecture diagram
hexago version                       # Show version
```

### Generated Project Commands

Projects generated by HexaGo will have commands like:
```bash
myapp run                            # Start the server (main command)
myapp run --config .myapp.yaml       # Start with specific config
myapp run --observability            # Start with observability enabled
myapp migrate up                     # Run database migrations
myapp migrate down                   # Rollback migrations
myapp create --config .myapp.yaml    # Create example config file
myapp seed                           # Seed database (if generated)
myapp version                        # Show version
```

**Important**: Each subcommand is in its own file (`cmd/run.go`, `cmd/migrate.go`, etc.) and registers itself via `init()` function.

### Common Flags

- `--module, -m`: Go module name
- `--framework, -f`: Web framework (echo|gin|chi|fiber|stdlib)
- `--implementation, -i`: Implementation type (postgres|mysql|memory|http|grpc)
- `--with-docker`: Include Docker files (default: true)
- `--with-example`: Include example code
- `--force`: Overwrite existing files

## Implementation Phases

The project follows a phased approach as defined in hexago-project-spec.md:

**Phase 1 (MVP)**: Basic init, use case, repository, HTTP handler, domain entity generation
**Phase 2**: Workers, value objects, tools, framework-specific templates
**Phase 3**: Diagrams, OpenAPI docs, validation, migration utilities
**Phase 4**: Auth scaffolding, rate limiting, circuit breakers, Kubernetes manifests

## Testing Strategy

### For HexaGo Tool
- Unit tests for generators (table-driven tests)
- Integration tests that generate sample projects
- Validation tests ensuring generated code compiles
- Golden file tests comparing generated output

### For Generated Projects
- Each generated component includes test file with examples
- Mock interfaces for driven ports
- Test helpers and utilities

## Important Patterns

### Cobra CLI Entry Point Pattern
**Critical for code generation:**

1. **main.go** - Minimal entry point:
```go
package main

import (
    "fmt"
    "os"
    "github.com/user/myapp/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

2. **cmd/root.go** - Root command with Execute():
```go
package cmd

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "Application description",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // Setup persistent flags, config
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}
```

3. **cmd/run.go** - The real application entrypoint:
```go
package cmd

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Start the server",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Application logic here
        // - Load configuration
        // - Initialize services
        // - Start server
        // - Graceful shutdown
        return nil
    },
}

func init() {
    rootCmd.AddCommand(runCmd)
}
```

### Graceful Shutdown Pattern (in cmd/run.go)
Generated projects use context-based cancellation with:
- Signal handling (os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
- Context with cancellation for coordinating goroutines
- Error channel for processor/server errors
- Timeout-based HTTP server shutdown (typically 10-30s)
- Worker coordination with WaitGroups
- Select statement waiting for signals or errors

Example structure:
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
errChan := make(chan error, 1)

// Start server in goroutine
go func() {
    errChan <- server.Start(ctx)
}()

// Wait for signal or error
select {
case sig := <-sigChan:
    cancel()
    // Graceful shutdown with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownCancel()
    return server.Shutdown(shutdownCtx)
case err := <-errChan:
    cancel()
    return err
}
```

### Worker Pattern
Background workers use:
- Goroutines for concurrent processing
- Channels for job queues
- Context for cancellation
- WaitGroups for coordination
- Clean shutdown hooks
- Manager pattern for orchestrating multiple workers

### Configuration Pattern
- Viper for config management (in cmd/root.go or internal/config)
- YAML config file (e.g., `.myapp.yaml`)
- Environment variable override support (with prefixes like `MYAPP_`)
- Sensible defaults via `viper.SetDefault()`
- Type-safe config structs
- Config loading in `cmd/root.go` via `initConfig()` function
- GetConfig() helper function for commands to access initialized config

### Observability Pattern
Modern generated projects should include observability support:
- **Health Checks**: Endpoint exposing component health (MQTT, DB, etc.)
- **Metrics**: Prometheus metrics for monitoring
- **Observability Server**: Separate HTTP server for health and metrics
  - Runs in background goroutine
  - Separate from main application server
  - Graceful shutdown coordinated with main server
- **Structure**: `internal/observability/` package with:
  - `health.go` - Health checker with component registration
  - `metrics.go` - Prometheus metrics (request count, latency, etc.)
  - `server.go` - HTTP server for exposing health and metrics

Example observability endpoints:
- `GET /health` - Overall health status
- `GET /health/{component}` - Specific component health
- `GET /metrics` - Prometheus metrics

## File Naming Conventions

- Use snake_case: `create_user.go`, `user_repository.go`
- Test files: `create_user_test.go`
- Interface files named after interface: `user_repository.go`
- Handler files: `user_handler.go`

## Error Handling

- Validate all inputs before generation
- Provide clear error messages with actionable suggestions
- Check for file conflicts before writing
- Support `--dry-run` for preview (future feature)

## Framework-Specific Generation

When generating HTTP handlers, adapt to the selected framework:
- **stdlib**: `func(w http.ResponseWriter, r *http.Request)`
- **Echo**: `func(c echo.Context) error`
- **Gin**: `func(c *gin.Context)`
- **Chi**: Same as stdlib (chi uses standard http.HandlerFunc)
- **Fiber**: `func(c *fiber.Ctx) error`

## Configuration File Support

Projects can have `.hexago.yaml` for customization:
- Custom template paths
- Naming conventions
- Code generation preferences
- Testing framework selection

## Educational Features

Every generated file includes:
- Comments explaining what layer it belongs to
- Allowed/forbidden dependencies
- Links to architecture documentation
- TODO comments guiding implementation

## Real-World Example Reference

A production implementation at `/home/pato/go/src/github.com/segel-si/iot` demonstrates:

**Structure Used:**
- `main.go` - Minimal, calls `cmd.Execute()`
- `cmd/root.go` - Root command, config loading, `Execute()` function
- `cmd/run.go` - Server logic with graceful shutdown
- `cmd/create.go` - Config file generation
- `internal/core/domain/` - Domain entities
- `internal/core/services/` - Business logic (not "usecases")
- `internal/adapters/primary/queue/` - MQTT inbound adapter
- `internal/adapters/secondary/database/` - InfluxDB outbound adapter
- `internal/config/` - Viper configuration
- `internal/observability/` - Health checks and Prometheus metrics
- `pkg/logger/` - Reusable logger package
- No explicit `ports/` directory - interfaces defined where needed

**Key Patterns:**
- Graceful shutdown with context cancellation and timeout
- Signal handling (os.Interrupt, syscall.SIGTERM)
- Error channels for goroutine communication
- Channel-based concurrent processing pipeline
- Viper config with YAML file and env var overrides
- Observability server running in background goroutine

## Directory Naming Flexibility

HexaGo should support generation with different naming conventions via config:

**Adapter Naming** (`.hexago.yaml`):
```yaml
conventions:
  adapter_style: "primary-secondary"  # or "driver-driven"
  core_logic: "services"              # or "usecases"
  explicit_ports: false               # true = create ports/ dir
```

This allows generating projects that match team preferences while maintaining hexagonal architecture principles.

## MCP Server

HexaGo ships with a built-in MCP server (`hexago mcp`). When using it to work on
hexagonal architecture projects:

- Always use the hexago MCP tools to create components, never create files manually
  inside `internal/core` or `internal/adapters`
- After adding any component, call `hexago_validate` to verify architecture compliance
- If `hexago_validate` fails, fix the issue before proceeding
- Every tool requires a `working_directory` parameter — the project root for `add` /
  `validate`, or the parent directory for `hexago_init`

## Notes

- Start implementation with `init` command as foundation
- Use embedded templates for portability (`go:embed`)
- Support both interactive and non-interactive modes
- Add verbose mode (`-v`) for debugging
- Generate README explaining architecture in target projects
- Templates should handle both naming conventions (primary/secondary vs driver/driven)
- Always generate minimal `main.go` that only calls `cmd.Execute()`
- Place real application logic in `cmd/run.go` (or similar command)
- Include graceful shutdown pattern in all server commands
