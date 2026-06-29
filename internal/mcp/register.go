package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterMCPTools(s *server.MCPServer) {
	registerInit(s)
	registerAddServices(s)
	registerAddDomainEntity(s)
	registerAddDomainValueObject(s)
	registerAddAdapter(s)
	registerAddWorker(s)
	registerAddMigration(s)
	registerAddTool(s)
	registerValidate(s)
}

func registerInit(s *server.MCPServer) {
	// hexago_init
	s.AddTool(
		mcp.NewTool("hexago_init",
			mcp.WithDescription(`Bootstrap a new Go project with hexagonal architecture.

Creates the full project skeleton: main.go, cmd/ (Cobra CLI), internal/core/, internal/adapters/,
internal/config/, pkg/logger/, go.mod, Makefile, README.

working_directory is the parent folder; the project is created as working_directory/<name>/.
Set in_place=true to generate files directly into working_directory (no <name> subfolder).

Example call:
  working_directory: "/home/user/projects"
  name: "my-api"
  module: "github.com/user/my-api"
  project_type: "http-server"
  framework: "echo"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the parent directory. The project is created as <working_directory>/<name>/ unless in_place=true."),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Project name (used as directory name and binary name). E.g. my-api, user-service."),
				mcp.Required(),
			),
			mcp.WithString("module",
				mcp.Description("Go module path. E.g. github.com/user/my-api. Defaults to the project name if omitted."),
			),
			mcp.WithString("project_type",
				mcp.Description(`Project type:
  http-server — HTTP API server with a web framework (default)
  service     — long-running daemon with no HTTP layer`),
				mcp.Enum("http-server", "service"),
			),
			mcp.WithString("framework",
				mcp.Description("Web framework. Only relevant when project_type=http-server. Default: stdlib."),
				mcp.Enum("echo", "gin", "chi", "fiber", "stdlib"),
			),
			mcp.WithString("adapter_style",
				mcp.Description(`Naming convention for adapters:
  primary-secondary — adapters/primary/ and adapters/secondary/ (DDD style, default)
  driver-driven     — adapters/driver/ and adapters/driven/ (ports & adapters terminology)`),
				mcp.Enum("primary-secondary", "driver-driven"),
			),
			mcp.WithString("core_logic",
				mcp.Description(`Directory name for business logic inside internal/core/:
  services  — internal/core/services/ (default)
  usecases  — internal/core/usecases/`),
				mcp.Enum("services", "usecases"),
			),
			mcp.WithBoolean("with_docker",
				mcp.Description("Generate a multi-stage Dockerfile and docker-compose.yml."),
			),
			mcp.WithBoolean("with_observability",
				mcp.Description("Add internal/observability/ with health-check and Prometheus metrics endpoints."),
			),
			mcp.WithBoolean("with_migrations",
				mcp.Description("Add migrations/ directory and golang-migrate wiring in cmd/migrate.go."),
			),
			mcp.WithBoolean("with_workers",
				mcp.Description("Add background worker scaffolding (manager + example worker)."),
			),
			mcp.WithBoolean("with_metrics",
				mcp.Description("Add Prometheus metrics (implies with_observability)."),
			),
			mcp.WithBoolean("with_tests",
				mcp.Description("Enable go-testgen test generation for add commands (requires go-testgen ≥ v0.1.0)."),
			),
			mcp.WithBoolean("with_example",
				mcp.Description("Include example service, entity, and adapter to illustrate the architecture."),
			),
			mcp.WithBoolean("explicit_ports",
				mcp.Description("Create an explicit internal/core/ports/ directory with port interfaces."),
			),
			mcp.WithBoolean("in_place",
				mcp.Description("Generate files directly into working_directory instead of creating a <name> subdirectory. Use this when working_directory is already the intended project root."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "init", name}
			if v, _ := args["module"].(string); v != "" {
				cliArgs = append(cliArgs, "--module", v)
			}
			if v, _ := args["project_type"].(string); v != "" {
				cliArgs = append(cliArgs, "--project-type", v)
			}
			if v, _ := args["framework"].(string); v != "" {
				cliArgs = append(cliArgs, "--framework", v)
			}
			if v, _ := args["adapter_style"].(string); v != "" {
				cliArgs = append(cliArgs, "--adapter-style", v)
			}
			if v, _ := args["core_logic"].(string); v != "" {
				cliArgs = append(cliArgs, "--core-logic", v)
			}
			if v, _ := args["with_docker"].(bool); v {
				cliArgs = append(cliArgs, "--with-docker")
			}
			if v, _ := args["with_observability"].(bool); v {
				cliArgs = append(cliArgs, "--with-observability")
			}
			if v, _ := args["with_migrations"].(bool); v {
				cliArgs = append(cliArgs, "--with-migrations")
			}
			if v, _ := args["with_workers"].(bool); v {
				cliArgs = append(cliArgs, "--with-workers")
			}
			if v, _ := args["with_metrics"].(bool); v {
				cliArgs = append(cliArgs, "--with-metrics")
			}
			if v, _ := args["with_tests"].(bool); v {
				cliArgs = append(cliArgs, "--with-tests")
			}
			if v, _ := args["with_example"].(bool); v {
				cliArgs = append(cliArgs, "--with-example")
			}
			if v, _ := args["explicit_ports"].(bool); v {
				cliArgs = append(cliArgs, "--explicit-ports")
			}
			if v, _ := args["in_place"].(bool); v {
				cliArgs = append(cliArgs, "--in-place")
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerAddServices(s *server.MCPServer) {
	// hexago_add_service
	s.AddTool(
		mcp.NewTool("hexago_add_service",
			mcp.WithDescription(`Add a business-logic service (use case) to internal/core/services/ (or usecases/).

Each service lives in its own sub-package with named CRUD methods wired to the domain repository port.

Generates:
  - internal/core/services/<pkg>/<pkg>.go      — Create/GetByID/Update/List methods, service struct
  - internal/core/services/<pkg>/<pkg>_test.go — test skeleton

The generated code belongs to the core layer and must not import from adapters/.

Example call (with entity):
  working_directory: "/home/user/projects/my-api"
  name: "Category"
  entity: "Category"
  description: "manages category records"

Example call (no entity):
  working_directory: "/home/user/projects/my-api"
  name: "Notification"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Service name in PascalCase. E.g. Category, Order, Product."),
				mcp.Required(),
			),
			mcp.WithString("entity",
				mcp.Description("Domain entity this service manages (PascalCase). Determines sub-package name. E.g. Category → categories/."),
			),
			mcp.WithString("description",
				mcp.Description("One-line description embedded as a comment in the generated file."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "service", name}
			if v, _ := args["entity"].(string); v != "" {
				cliArgs = append(cliArgs, "--entity", v)
			}
			if v, _ := args["description"].(string); v != "" {
				cliArgs = append(cliArgs, "--description", v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerAddDomainEntity(s *server.MCPServer) {
	// hexago_add_domain_entity
	s.AddTool(
		mcp.NewTool("hexago_add_domain_entity",
			mcp.WithDescription(`Add a domain entity to internal/core/domain/.

An entity is an object with a unique identity that persists through time (e.g. User, Order, Product).
It contains business logic and validation rules and belongs entirely to the core layer.

Generates:
  - internal/core/domain/<name>.go      — struct with fields, constructor, validation
  - internal/core/domain/<name>_test.go — test skeleton

Fields format: comma-separated name:type pairs.
  "id:string,name:string,email:string,createdAt:time.Time"
  "id:uuid.UUID,amount:float64,currency:string"
Field names are converted to PascalCase automatically.

Example call:
  working_directory: "/home/user/projects/my-api"
  name: "User"
  fields: "id:string,name:string,email:string,createdAt:time.Time"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Entity name in PascalCase. E.g. User, Order, Product, Invoice."),
				mcp.Required(),
			),
			mcp.WithString("fields",
				mcp.Description(`Comma-separated field definitions as name:type pairs.
E.g. "id:string,name:string,email:string,createdAt:time.Time"
Supported types: any valid Go type (string, int, int64, float64, bool, time.Time, uuid.UUID, etc.)`),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "domain", "entity", name}
			if v, _ := args["fields"].(string); v != "" {
				cliArgs = append(cliArgs, "--fields", v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerAddDomainValueObject(s *server.MCPServer) {
	// hexago_add_domain_valueobject
	s.AddTool(
		mcp.NewTool("hexago_add_domain_valueobject",
			mcp.WithDescription(`Add a domain value object to internal/core/domain/.

A value object is an immutable object defined only by its attributes, with no unique identity
(e.g. Email, Money, Address, PhoneNumber). It is compared by value, not by reference.

Standalone (no entity):
  - internal/core/domain/<name>/<name>.go      — own sub-package
  - internal/core/domain/<name>/<name>_test.go

Entity-bound (with entity):
  - internal/core/domain/<entities>/<snake_name>.go  — co-located with entity
  - internal/core/domain/<entities>/<snake_name>_test.go

Fields format: comma-separated name:type pairs.
  "value:string"
  "amount:float64,currency:string"

Example call (standalone):
  working_directory: "/home/user/projects/my-api"
  name: "Email"
  fields: "value:string"

Example call (entity-bound):
  working_directory: "/home/user/projects/my-api"
  name: "StockLevel"
  fields: "value:float64"
  entity: "Product"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Value object name in PascalCase. E.g. Email, Money, Address, PhoneNumber."),
				mcp.Required(),
			),
			mcp.WithString("fields",
				mcp.Description(`Comma-separated field definitions as name:type pairs.
E.g. "value:string" or "amount:float64,currency:string"`),
			),
			mcp.WithString("entity",
				mcp.Description("Entity name (PascalCase) to co-locate this VO with. Omit for a standalone sub-package."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "domain", "valueobject", name}
			if v, _ := args["fields"].(string); v != "" {
				cliArgs = append(cliArgs, "--fields", v)
			}
			if v, _ := args["entity"].(string); v != "" {
				cliArgs = append(cliArgs, "--entity", v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerAddAdapter(s *server.MCPServer) {
	// hexago_add_adapter
	s.AddTool(
		mcp.NewTool("hexago_add_adapter",
			mcp.WithDescription(`Add an adapter to the project.

Adapters connect the core to the outside world. Two directions:

  primary   (inbound)  — drives the application; receives requests from external actors.
                         Lives in internal/adapters/primary/<adapter_type>/.
                         Types: http, grpc, queue
                         E.g. UserHandler (HTTP), OrderConsumer (queue)

  secondary (outbound) — driven by the application; talks to external systems.
                         Lives in internal/adapters/secondary/<adapter_type>/.
                         Types: database, external, cache
                         E.g. UserRepository (database), EmailService (external)

Generates:
  - internal/adapters/<direction>/<adapter_type>/<name>.go
  - internal/adapters/<direction>/<adapter_type>/<name>_test.go

Example calls:
  direction: "primary",   adapter_type: "http",     name: "UserHandler"
  direction: "secondary", adapter_type: "database",  name: "UserRepository"
  direction: "primary",   adapter_type: "grpc",     name: "OrderService"
  direction: "secondary", adapter_type: "external",  name: "EmailClient"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
			mcp.WithString("direction",
				mcp.Description(`Adapter direction:
  primary   — inbound adapter (HTTP handler, gRPC server, message queue consumer)
  secondary — outbound adapter (database repository, external API client, cache)`),
				mcp.Required(),
				mcp.Enum("primary", "secondary"),
			),
			mcp.WithString("adapter_type",
				mcp.Description(`Implementation technology:
  For primary:   http, grpc, queue
  For secondary: database, external, cache`),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Adapter name in PascalCase. E.g. UserHandler, CategoryRepository, EmailClient."),
				mcp.Required(),
			),
			mcp.WithString("entity",
				mcp.Description(`Domain entity this adapter serves (PascalCase). Behaviour by direction:
  primary   http — generates sub-package with two files: <snake_entity>.go (Config/DTOs) + handlers.go (List/Create/GetByID/Update)
  secondary database — generates sub-package implementing the entity's Repository port`),
			),
			mcp.WithString("from-port",
				mcp.Description("Port interface name to implement (only used with explicit_ports projects). E.g. UserRepository, EmailSender."),
			),
			mcp.WithString("test",
				mcp.Description(`Test generation for this component: "with-test" (force on) or "no-test" (force off). Overrides project config.`),
				mcp.Enum("with-test", "no-test"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			direction, _ := args["direction"].(string)
			adapterType, _ := args["adapter_type"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "adapter", direction, adapterType, name}
			if v, _ := args["entity"].(string); v != "" {
				cliArgs = append(cliArgs, "--entity", v)
			}
			if v, _ := args["from-port"].(string); v != "" {
				cliArgs = append(cliArgs, "--from-port", v)
			}
			if v, _ := args["test"].(string); v != "" {
				cliArgs = append(cliArgs, "--"+v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerAddWorker(s *server.MCPServer) {
	// hexago_add_worker
	s.AddTool(
		mcp.NewTool("hexago_add_worker",
			mcp.WithDescription(`Add a background worker to the project.

Workers run concurrently using goroutines and channels with graceful shutdown via context.
Generated file: internal/workers/<name>.go

Worker types:
  queue    — pool of N goroutines consuming jobs from a buffered channel (default)
             Params: workers (concurrency), queue_size (buffer)
  periodic — single goroutine that ticks at a fixed interval
             Params: interval (e.g. "5m", "1h", "30s")
  event    — goroutine that reacts to external events via a channel

Example calls:
  name: "EmailWorker",        worker_type: "queue",    workers: 5, queue_size: 100
  name: "HealthCheckWorker",  worker_type: "periodic",  interval: "1m"
  name: "NotificationWorker", worker_type: "event"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Worker name in PascalCase. E.g. EmailWorker, ReportWorker, CleanupWorker."),
				mcp.Required(),
			),
			mcp.WithString("worker_type",
				mcp.Description("queue (default) | periodic | event"),
				mcp.Enum("queue", "periodic", "event"),
			),
			mcp.WithString("interval",
				mcp.Description("Tick interval for periodic workers. Go duration string: 30s, 5m, 1h. Default: 5m."),
			),
			mcp.WithNumber("workers",
				mcp.Description("Number of concurrent goroutines for queue workers. Default: 5."),
			),
			mcp.WithNumber("queue_size",
				mcp.Description("Buffered channel size for queue workers. Default: 100."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "worker", name}
			if v, _ := args["worker_type"].(string); v != "" {
				cliArgs = append(cliArgs, "--type", v)
			}
			if v, _ := args["interval"].(string); v != "" {
				cliArgs = append(cliArgs, "--interval", v)
			}
			if v, _ := args["workers"].(float64); v > 0 {
				cliArgs = append(cliArgs, "--workers", fmt.Sprintf("%d", int(v)))
			}
			if v, _ := args["queue_size"].(float64); v > 0 {
				cliArgs = append(cliArgs, "--queue-size", fmt.Sprintf("%d", int(v)))
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerAddMigration(s *server.MCPServer) {
	// hexago_add_migration
	s.AddTool(
		mcp.NewTool("hexago_add_migration",
			mcp.WithDescription(`Add a database migration file pair (up + down) using golang-migrate format.

Files are created with sequential numbering:
  migrations/000001_<name>.up.sql
  migrations/000001_<name>.down.sql

The number is automatically incremented based on existing migrations in the directory.
Use snake_case for the migration name to describe the schema change.

Example calls:
  name: "create_users_table"
  name: "add_email_index_to_users"
  name: "alter_orders_add_status_column"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and migrations/)."),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Migration name in snake_case describing the schema change. E.g. create_users_table, add_email_index."),
				mcp.Required(),
			),
			mcp.WithString("migration_type",
				mcp.Description("Migration format: sql (default) or go."),
				mcp.Enum("sql", "go"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "migration", name}
			if v, _ := args["migration_type"].(string); v != "" {
				cliArgs = append(cliArgs, "--type", v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerAddTool(s *server.MCPServer) {
	// hexago_add_tool
	s.AddTool(
		mcp.NewTool("hexago_add_tool",
			mcp.WithDescription(`Add an infrastructure utility to internal/infrastructure/<tool_type>/.

Use for cross-cutting concerns that don't belong to core or adapters.

Tool types:
  logger     — structured logger implementation (e.g. zerolog, zap wrapper)
  validator  — input validation utilities (e.g. request validation)
  mapper     — DTO ↔ domain mapping helpers
  middleware — HTTP middleware (auth, rate limiting, logging, CORS, etc.)

Generates:
  - internal/infrastructure/<tool_type>/<name>.go
  - internal/infrastructure/<tool_type>/<name>_test.go

Example calls:
  tool_type: "logger",     name: "ZerologLogger"
  tool_type: "validator",  name: "RequestValidator"
  tool_type: "mapper",     name: "UserMapper"
  tool_type: "middleware",  name: "AuthMiddleware"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
			mcp.WithString("tool_type",
				mcp.Description("logger | validator | mapper | middleware"),
				mcp.Required(),
				mcp.Enum("logger", "validator", "mapper", "middleware"),
			),
			mcp.WithString("name",
				mcp.Description("Tool name in PascalCase. E.g. ZerologLogger, RequestValidator, UserMapper, AuthMiddleware."),
				mcp.Required(),
			),
			mcp.WithString("description",
				mcp.Description("One-line description embedded as a comment in the generated file."),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			toolType, _ := args["tool_type"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "tool", toolType, name}
			if v, _ := args["description"].(string); v != "" {
				cliArgs = append(cliArgs, "--description", v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}

func registerValidate(s *server.MCPServer) {
	// hexago_validate
	s.AddTool(
		mcp.NewTool("hexago_validate",
			mcp.WithDescription(`Validate that the project follows hexagonal architecture rules.

Checks performed:
  ✓ Core domain has no external dependencies
  ✓ Services/use cases only depend on domain and ports
  ✓ Adapters don't import from other adapters
  ✓ Dependency direction is always inward (adapters → core, never core → adapters)
  ✓ Proper package organization and naming conventions

Returns a structured report with passed checks, warnings, and errors.
Call this after every hexago_add_* operation to catch violations early.

Example call:
  working_directory: "/home/user/projects/my-api"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			cliArgs := []string{"--working-directory", wd, "validate"}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)
}
