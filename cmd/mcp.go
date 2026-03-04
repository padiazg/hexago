/*
Copyright © 2026 HexaGo Contributors
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/padiazg/hexago/pkg/version"
	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the HexaGo MCP server (stdio)",
	Long: `Start HexaGo as a Model Context Protocol (MCP) server over stdio.

AI assistants (Claude Code, Claude Desktop, etc.) can use this server to scaffold
hexagonal architecture projects without leaving their conversation.

Each MCP tool delegates to the hexago CLI with --working-directory, so all
generation logic is shared with the regular CLI commands.

Register with Claude Code:
  claude mcp add hexago -- hexago mcp

Or scoped to a project:
  claude mcp add --scope project hexago -- hexago mcp`,
	RunE: runMCPServer,
}

const mcpInstructions = `You are connected to HexaGo, a scaffolding tool for Go applications
following Hexagonal Architecture (Ports & Adapters).

## Golden rules

1. NEVER run hexago as a shell command (e.g. "hexago add ..."). Use the MCP tools — they
   handle working_directory and return structured output.
2. working_directory must always be an absolute path.
3. ALWAYS call hexago_validate after adding any component to catch violations early.

## working_directory

- hexago_init    → parent directory; project is created as <working_directory>/<name>/
                   Set in_place=true to generate directly into working_directory instead.
- all other tools → project root: the directory that contains go.mod and internal/

────────────────────────────────────────────────────────────────────────────────
## hexago_init — bootstrap a new project
────────────────────────────────────────────────────────────────────────────────

Required:  working_directory, name
Optional:
  module          Go module path. E.g. "github.com/user/my-api". Defaults to name.
  project_type    "http-server" (default) | "service"
  framework       "stdlib" (default) | "echo" | "gin" | "chi" | "fiber"
                  Only used when project_type=http-server.
  adapter_style   "primary-secondary" (default) | "driver-driven"
                  Controls directory names: adapters/primary+secondary vs adapters/driver+driven.
  core_logic      "services" (default) | "usecases"
                  Controls the directory name inside internal/core/.
  in_place        bool — generate files directly into working_directory (no <name> subfolder).
                  Use when working_directory is already the intended project root.

Feature flags (all bool, default false):
  with_docker        — Dockerfile + docker-compose.yml
  with_observability — internal/observability/ with health-check and Prometheus endpoints
  with_migrations    — migrations/ directory and cmd/migrate.go wiring (golang-migrate)
  with_workers       — background worker scaffolding (manager + example worker)
  with_metrics       — Prometheus metrics (implies with_observability)
  with_example       — example service, entity, and adapter illustrating the architecture
  explicit_ports     — create internal/core/ports/ with explicit port interfaces

────────────────────────────────────────────────────────────────────────────────
## hexago_add_service — add a business-logic use case
────────────────────────────────────────────────────────────────────────────────

Generated: internal/core/<core_logic>/<name>.go + <name>_test.go

Required:  working_directory, name  (PascalCase, e.g. "CreateUser", "GetOrderByID")
Optional:
  description     One-line comment embedded in the generated file.

────────────────────────────────────────────────────────────────────────────────
## hexago_add_domain_entity — add a domain entity
────────────────────────────────────────────────────────────────────────────────

Generated: internal/core/domain/<name>.go + <name>_test.go
An entity has unique identity and contains business logic (e.g. User, Order, Product).

Required:  working_directory, name  (PascalCase)
Optional:
  fields     Comma-separated name:type pairs.
             "id:string,name:string,email:string,createdAt:time.Time"
             "id:uuid.UUID,amount:float64,currency:string,active:bool"
             Any valid Go type is accepted. Names are auto-converted to PascalCase.

────────────────────────────────────────────────────────────────────────────────
## hexago_add_domain_valueobject — add a domain value object
────────────────────────────────────────────────────────────────────────────────

Generated: internal/core/domain/<name>.go + <name>_test.go
A value object is immutable, has no identity, and is compared by value (e.g. Email, Money).

Required:  working_directory, name  (PascalCase)
Optional:
  fields     Same format as domain entity. E.g. "value:string" or "amount:float64,currency:string"

────────────────────────────────────────────────────────────────────────────────
## hexago_add_adapter — add an inbound or outbound adapter
────────────────────────────────────────────────────────────────────────────────

Generated: internal/adapters/<direction>/<adapter_type>/<name>.go + <name>_test.go

Required:  working_directory, direction, adapter_type, name  (PascalCase)

  direction      "primary"   — inbound: receives requests (HTTP handler, gRPC server, queue consumer)
               | "secondary" — outbound: calls external systems (DB repo, API client, cache)

  adapter_type   For primary:   "http" | "grpc" | "queue"
                 For secondary: "database" | "external" | "cache"
                 Any other string is accepted and used as the subdirectory name.

Examples:
  direction=primary,   adapter_type=http,     name=UserHandler
  direction=primary,   adapter_type=grpc,     name=OrderService
  direction=primary,   adapter_type=queue,    name=PaymentConsumer
  direction=secondary, adapter_type=database,  name=UserRepository
  direction=secondary, adapter_type=external,  name=EmailClient
  direction=secondary, adapter_type=cache,     name=SessionCache

────────────────────────────────────────────────────────────────────────────────
## hexago_add_worker — add a background worker
────────────────────────────────────────────────────────────────────────────────

Generated: internal/workers/<name>.go

Required:  working_directory, name  (PascalCase, e.g. "EmailWorker")
Optional:
  worker_type   "queue" (default) | "periodic" | "event"
  interval      Duration string for periodic workers. Default "5m". E.g. "30s", "1h", "15m".
  workers       int — goroutine pool size for queue workers. Default 5.
  queue_size    int — buffered channel size for queue workers. Default 100.

Examples:
  name=EmailWorker,       worker_type=queue,    workers=10, queue_size=200
  name=CleanupWorker,     worker_type=periodic,  interval=1h
  name=AlertWorker,       worker_type=event

────────────────────────────────────────────────────────────────────────────────
## hexago_add_migration — add a database migration file pair
────────────────────────────────────────────────────────────────────────────────

Generated: migrations/<seq>_<name>.up.sql + migrations/<seq>_<name>.down.sql
Sequence number is auto-incremented from existing migrations.

Required:  working_directory, name  (snake_case, e.g. "create_users_table")
Optional:
  migration_type   "sql" (default) | "go"

────────────────────────────────────────────────────────────────────────────────
## hexago_add_tool — add an infrastructure utility
────────────────────────────────────────────────────────────────────────────────

Generated: internal/infrastructure/<tool_type>/<name>.go + <name>_test.go

Required:  working_directory, tool_type, name  (PascalCase)

  tool_type   "logger"     — structured logger implementation
            | "validator"  — input / request validation utilities
            | "mapper"     — DTO ↔ domain mapping helpers
            | "middleware" — HTTP middleware (auth, rate limiting, logging, CORS, etc.)

Optional:
  description   One-line comment embedded in the generated file.

Examples:
  tool_type=logger,     name=ZerologLogger
  tool_type=validator,  name=RequestValidator
  tool_type=mapper,     name=UserMapper
  tool_type=middleware,  name=AuthMiddleware

────────────────────────────────────────────────────────────────────────────────
## hexago_validate — validate architecture compliance
────────────────────────────────────────────────────────────────────────────────

Required:  working_directory

Checks dependency direction (adapters → core, never core → adapters), package organization,
and naming conventions. Returns passed checks, warnings, and errors.`

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	s := server.NewMCPServer("hexago", version.CurrentVersion().String(),
		server.WithToolCapabilities(false),
		server.WithInstructions(mcpInstructions),
	)
	registerMCPTools(s)
	return server.ServeStdio(s)
}

// runSelf executes the current hexago binary with the given arguments and returns
// the combined stdout+stderr output.
func runSelf(ctx context.Context, args ...string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot resolve hexago binary: %w", err)
	}
	c := exec.CommandContext(ctx, self, args...)
	out, err := c.CombinedOutput()
	return string(out), err
}

// toolResult converts runSelf output into an MCP CallToolResult.
// Errors are reported as tool-level errors (IsError=true) so the LLM can see them.
func toolResult(out string, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("ERROR: %v\n\n%s", err, out)), nil
	}
	return mcp.NewToolResultText(out), nil
}

func registerMCPTools(s *server.MCPServer) {
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

	// hexago_add_service
	s.AddTool(
		mcp.NewTool("hexago_add_service",
			mcp.WithDescription(`Add a business-logic service (use case) to internal/core/services/ (or usecases/).

Generates:
  - internal/core/services/<name>.go      — Input/Output structs, service struct, constructor, Execute()
  - internal/core/services/<name>_test.go — test skeleton

The generated code belongs to the core layer and must not import from adapters/.

Example call:
  working_directory: "/home/user/projects/my-api"
  name: "CreateUser"
  description: "Creates a new user account"`),
			mcp.WithString("working_directory",
				mcp.Description("Absolute path to the project root (the directory containing go.mod and internal/)."),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Service name in PascalCase. Describes the use case. E.g. CreateUser, GetOrderByID, SendWelcomeEmail."),
				mcp.Required(),
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
			if v, _ := args["description"].(string); v != "" {
				cliArgs = append(cliArgs, "--description", v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)

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

	// hexago_add_domain_valueobject
	s.AddTool(
		mcp.NewTool("hexago_add_domain_valueobject",
			mcp.WithDescription(`Add a domain value object to internal/core/domain/.

A value object is an immutable object defined only by its attributes, with no unique identity
(e.g. Email, Money, Address, PhoneNumber). It is compared by value, not by reference.

Generates:
  - internal/core/domain/<name>.go      — immutable struct, constructor with validation, Equals()
  - internal/core/domain/<name>_test.go — test skeleton

Fields format: comma-separated name:type pairs.
  "value:string"
  "amount:float64,currency:string"

Example call:
  working_directory: "/home/user/projects/my-api"
  name: "Email"
  fields: "value:string"`),
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
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "domain", "valueobject", name}
			if v, _ := args["fields"].(string); v != "" {
				cliArgs = append(cliArgs, "--fields", v)
			}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)

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
				mcp.Description("Adapter name in PascalCase. E.g. UserHandler, UserRepository, EmailClient."),
				mcp.Required(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			wd, _ := args["working_directory"].(string)
			direction, _ := args["direction"].(string)
			adapterType, _ := args["adapter_type"].(string)
			name, _ := args["name"].(string)
			cliArgs := []string{"--working-directory", wd, "add", "adapter", direction, adapterType, name}
			return toolResult(runSelf(ctx, cliArgs...))
		},
	)

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
