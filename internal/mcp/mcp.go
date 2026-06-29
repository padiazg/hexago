package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/padiazg/hexago/internal/version"
)

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
  with_tests         — enable go-testgen test generation for add commands
  explicit_ports     — create internal/core/ports/ with explicit port interfaces

────────────────────────────────────────────────────────────────────────────────
## hexago_add_service — add a business-logic use case
────────────────────────────────────────────────────────────────────────────────

Generated: internal/core/<core_logic>/<pkg>/<pkg>.go + <pkg>_test.go

Required:  working_directory, name  (PascalCase, e.g. "Category", "Order")
Optional:
  entity        Domain entity this service manages. Determines sub-package name (e.g. "Category" → categories/).
  description   One-line comment embedded in the generated file.

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

Generated (standalone):  internal/core/domain/<name>/<name>.go + <name>_test.go
Generated (entity-bound): internal/core/domain/<entities>/<snake_name>.go + <snake_name>_test.go
A value object is immutable, has no identity, and is compared by value (e.g. Email, Money).

Required:  working_directory, name  (PascalCase)
Optional:
  fields     Same format as domain entity. E.g. "value:string" or "amount:float64,currency:string"
  entity     Entity name to co-locate with (entity-bound VO). Omit for a standalone sub-package.

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

Optional:
  entity      Domain entity this adapter serves (PascalCase); determines sub-package for database adapters
  from-port   Port interface name to implement (for explicit_ports projects). E.g. "UserRepository"
  test        "with-test" | "no-test" — force on/off test generation (overrides config)

	Example calls:
	  direction=primary,   adapter_type=http,      name=UserHandler
	  direction=primary,   adapter_type=grpc,      name=OrderService
	  direction=primary,   adapter_type=queue,     name=PaymentConsumer
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
  tool_type=logger,      name=ZerologLogger
  tool_type=validator,   name=RequestValidator
  tool_type=mapper,      name=UserMapper
  tool_type=middleware,  name=AuthMiddleware

────────────────────────────────────────────────────────────────────────────────
## hexago_validate — validate architecture compliance
────────────────────────────────────────────────────────────────────────────────

Required:  working_directory

Checks dependency direction (adapters → core, never core → adapters), package organization,
and naming conventions. Returns passed checks, warnings, and errors.`

func Run() error {
	s := server.NewMCPServer("hexago", version.CurrentVersion().String(),
		server.WithToolCapabilities(false),
		server.WithInstructions(mcpInstructions),
	)

	RegisterMCPTools(s)
	return server.ServeStdio(s)
}
