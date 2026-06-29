# Template Customization

HexaGo supports template customization, allowing you to modify the generated code to match your team's preferences, coding standards, and branding.

---

## Template Sources

HexaGo searches for templates in the following order (highest priority first):

| Priority | Location | Use Case |
|----------|----------|----------|
| **1 — Highest** | `<binary-dir>/templates/` | Binary-local overrides |
| **2** | `./.hexago/templates/` | Per-project customization |
| **3** | `~/.hexago/templates/` | User-wide defaults |
| **4 — Fallback** | Embedded in binary | Default templates |

When you provide a custom template, HexaGo uses it instead of the built-in default.

---

## Template Structure

The template directory mirrors the generated project structure — the path of a template tells you exactly where its output lands.

```
templates/
├── cmd/                        # CLI commands
│   ├── root.go.tmpl            → cmd/root.go
│   ├── run.go.tmpl             → cmd/run.go  (base)
│   ├── run_http_server.go.tmpl → cmd/run.go  (http-server type)
│   ├── run_service.go.tmpl     → cmd/run.go  (service type)
│   └── version.go.tmpl         → cmd/version.go
│
├── pkg/
│   ├── server/
│   │   └── server_interface.go.tmpl  → pkg/server/server.go
│   ├── httpserver/             # Framework-specific server + handler base
│   │   ├── http_server_chi.go.tmpl    → pkg/httpserver/server.go
│   │   ├── http_server_echo.go.tmpl
│   │   ├── http_server_gin.go.tmpl
│   │   ├── http_server_fiber.go.tmpl
│   │   └── http_server_stdlib.go.tmpl
│   ├── logger/
│   │   └── logger.go.tmpl      → pkg/logger/logger.go
│   └── version/                # Version info + splash
│       ├── version.go.tmpl     → pkg/version/version.go
│       ├── splash.go.tmpl      → pkg/version/splash.go
│       └── version_test.go.tmpl→ pkg/version/version_test.go
│
├── adapter/                    # Adapter root templates
│   ├── adapter_test.go.tmpl    → <adapter>_test.go
│   ├── http.go.tmpl            → primary/http/<name>.go (generic)
│   ├── grpc.go.tmpl            → primary/grpc/<name>.go
│   ├── queue.go.tmpl           → primary/queue/<name>.go
│   ├── database.go.tmpl        → secondary/database/<name>.go
│   ├── external.go.tmpl        → secondary/external/<name>.go
│   ├── cache.go.tmpl           → secondary/cache/<name>.go
│   └── primary/
│       └── http/               # One sub-directory per framework
│           ├── chi/
│           │   ├── http_adapter.go.tmpl  → internal/adapters/{inbound}/http/http.go
│           │   ├── http_ping.go.tmpl     → .../http/ping/ping.go
│           │   ├── http_health.go.tmpl   → .../http/health/health.go
│           │   └── http_metrics.go.tmpl  → .../http/metrics/metrics.go
│           ├── echo/   (same set)
│           ├── gin/    (same set)
│           ├── fiber/  (same set)
│           └── stdlib/ (same set)
│
├── project/                    # Project-root files
│   ├── main.go.tmpl            → main.go
│   └── config.go.tmpl          → internal/config/config.go
│
├── misc/                       # Project support files
│   ├── makefile.tmpl           → Makefile
│   ├── readme.md.tmpl          → README.md
│   ├── dockerfile.tmpl         → Dockerfile
│   ├── compose.yaml.tmpl       → compose.yaml
│   └── gitignore.tmpl          → .gitignore
│
├── observability/              # Health / metrics helpers
│   ├── health.go.tmpl          → internal/observability/health.go
│   └── metrics.go.tmpl         → internal/observability/metrics.go
│
├── service/                    # Business logic templates
│   ├── service.go.tmpl         → service implementation
│   ├── service_test.go.tmpl    → service tests
│   ├── services_aggregator.go.tmpl → services.go aggregator
│   ├── services_stub.go.tmpl   → services stub (initial)
│   └── processor.go.tmpl       → processor (service type)
├── domain/                     # Domain entity templates
│   ├── entity.go.tmpl          → entity definition
│   ├── entity_test.go.tmpl     → entity tests
│   ├── errors.go.tmpl          → domain errors
│   ├── port.go.tmpl            → repository port interface
│   ├── value_object.go.tmpl    → value object definition
│   └── value_object_test.go.tmpl → value object tests
├── worker/                     # Background worker templates
│   ├── queue.go.tmpl           → queue worker
│   ├── periodic.go.tmpl        → periodic worker
│   ├── event.go.tmpl           → event worker
│   ├── manager.go.tmpl         → worker manager
│   └── worker_test.go.tmpl     → worker tests
├── migration/                  # Database migration templates
│   ├── up.sql.tmpl             → up migration
│   ├── down.sql.tmpl           → down migration
│   └── migrator.go.tmpl        → migrator command
└── tool/                       # Infrastructure tool templates
    ├── logger.go.tmpl          → logger implementation
    ├── logger_test.go.tmpl
    ├── validator.go.tmpl       → validator
    ├── validator_test.go.tmpl
    ├── mapper.go.tmpl          → DTO mapper
    ├── mapper_test.go.tmpl
    ├── middleware.go.tmpl      → HTTP middleware
    ├── middleware_test.go.tmpl
    └── generic_test.go.tmpl    → shared test utilities
```

---

## Template Commands

### List available templates

```shell
hexago templates list
```

Shows all 81 built-in templates grouped by directory. Templates with an active override are annotated with `← project-local` or `← user-global`.

### Check which template will be used

```shell
hexago templates which project/main.go.tmpl
```

Shows the winning source (embedded, project-local, user-global, or binary-local) with its full path.

### Export a template for editing

```shell
# Export to project-local (./.hexago/templates/)
hexago templates export project/main.go.tmpl

# Export to user-global (~/.hexago/templates/)
hexago templates export project/main.go.tmpl --global
```

### Export all templates at once

```shell
# Export all templates to project-local
hexago templates export-all

# Export all templates to user-global
hexago templates export-all --global

# Overwrite templates that already have an override
hexago templates export-all --force
```

Templates that already have an override are skipped by default. Use `--force` to overwrite them.

### Validate template syntax

```shell
hexago templates validate .hexago/templates/project/main.go.tmpl
```

Prints `✓ <path> — template syntax is valid` on success, or `✗ <path>` with the error detail on failure.

### Reset to default

```shell
# Remove project-local override
hexago templates reset project/main.go.tmpl

# Remove user-global override
hexago templates reset project/main.go.tmpl --global
```

---

## How to Customize

### Method 1: Export and Edit

```shell
# Export the template you want to customize
hexago templates export service/service.go.tmpl

# Edit the exported file
nano .hexago/templates/service/service.go.tmpl

# Generate code — will use your custom template
hexago add service CreateUser
```

### Method 2: Create from Scratch

```shell
mkdir -p .hexago/templates/service
cat > .hexago/templates/service/service.go.tmpl << 'EOF'
package {{.CoreLogic}}

import (
    "context"
    "fmt"
)

// {{.ServiceName}}Service {{.Description}}
type {{.ServiceName}}Service struct {
    // TODO: Add dependencies
}

func New{{.ServiceName}}Service() *{{.ServiceName}}Service {
    return &{{.ServiceName}}Service{}
}

func (s *{{.ServiceName}}Service) Execute(ctx context.Context) error {
    return fmt.Errorf("not implemented")
}
EOF

hexago add service CreateUser
```

---

## Template Syntax

Templates use Go's `text/template` syntax.

### Available Variables

Variables depend on the template. Check the header comment in each template file:

```
{{/*
Template: project/main.go
Variables:
  - ProjectName: string - Project name
  - ModuleName: string - Go module name
  - Year: string - Current year
  - Author: string - Author name
*/}}
```

Common variables:

| Variable | Available In | Description |
|----------|-------------|-------------|
| `ProjectName` | project templates | Project/app name |
| `ModuleName` | all templates | Go module path |
| `ServiceName` | service templates | Service name (PascalCase) |
| `CoreLogic` | service templates | `services` or `usecases` |
| `Description` | service, adapter | Description string |
| `Year` | project templates | Current year |
| `Author` | project templates | Author name |

### Custom Template Functions

| Function | Example | Result |
|----------|---------|--------|
| `upper` | `{{.ProjectName \| upper}}` | `MY-APP` |
| `lower` | `{{.ProjectName \| lower}}` | `my-app` |
| `title` | `{{.ProjectName \| title}}` | `My-App` |
| `snake` | `{{.ServiceName \| snake}}` | `create_user` |
| `lbrace` | `{{lbrace}}` | `{{` |
| `rbrace` | `{{rbrace}}` | `}}` |
| `zeroVal` | `{{"string" \| zeroVal}}` | `""` |
| `firstMethod` | `{{.Methods \| firstMethod}}` | First element of a slice |

---

## Examples

### Example 1: Add Company Header

Create `.hexago/templates/project/main.go.tmpl`:

```go
{{/*
Custom template with company header
*/}}
/*
Copyright © {{.Year}} {{.Author}}

CONFIDENTIAL - My Company Inc.
All Rights Reserved.

This source code is proprietary and confidential.
Unauthorized copying of this file is strictly prohibited.
*/
package main

import (
    "fmt"
    "os"

    "{{.ModuleName}}/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Example 2: Service with Structured Logging

Create `.hexago/templates/service/service.go.tmpl`:

```go
package {{.CoreLogic}}

import (
    "context"
    "fmt"

    "{{.ModuleName}}/pkg/logger"
)

// {{.ServiceName}}Input represents the input for {{.ServiceName}}
type {{.ServiceName}}Input struct {
    // TODO: Add input fields
}

// {{.ServiceName}}Output represents the output for {{.ServiceName}}
type {{.ServiceName}}Output struct {
    // TODO: Add output fields
}

// {{.ServiceName}}Service {{.Description}}
type {{.ServiceName}}Service struct {
    logger logger.Logger
    // TODO: Add other dependencies
}

// New{{.ServiceName}}Service creates a new instance
func New{{.ServiceName}}Service(log logger.Logger) *{{.ServiceName}}Service {
    return &{{.ServiceName}}Service{
        logger: log,
    }
}

// Execute runs the {{.ServiceName}} use case
func (s *{{.ServiceName}}Service) Execute(ctx context.Context, input {{.ServiceName}}Input) (*{{.ServiceName}}Output, error) {
    s.logger.Info("Executing {{.ServiceName}} service")
    defer s.logger.Info("{{.ServiceName}} service completed")

    // TODO: Implement business logic
    return nil, fmt.Errorf("not implemented")
}
```

### Example 3: Custom HTTP Handler with Metrics Startup

Create `.hexago/templates/adapter/primary/http_handler.go.tmpl` with your preferred HTTP handler structure including metrics, request tracing, span creation, etc.

---

## Best Practices

### Document Your Templates

Include a header comment with the template name, required variables, and description:

```
{{/*
Template: service/service.go
Custom template — includes structured logging by default
Variables: ServiceName, CoreLogic, ModuleName, Description
*/}}
```

### Version Your Templates

Commit project-local templates to version control so all team members use the same templates:

```shell
git add .hexago/templates/
git commit -m "Add custom hexago templates with company standards"
```

### Test After Customizing

```shell
# Test project generation
hexago init test-app --module github.com/test/app

# Test service generation
cd test-app
hexago add service TestService

# Verify it compiles
go build
```

### Share Templates Across Projects

For user-wide templates (affect all your projects):

```shell
hexago templates export service/service.go.tmpl --global
nano ~/.hexago/templates/service/service.go.tmpl
```

For team sharing, you can distribute templates as a tarball or a Git submodule:

```shell
# Pack templates
tar czf hexago-templates.tar.gz .hexago/templates/

# Unpack on another machine
tar xzf hexago-templates.tar.gz
```

---

## Currently Available for Customization

All 81 built-in templates are available for customization. Run `hexago templates list` to see the full set. Use `hexago templates export-all` to export every template to your override directory in one step.

---

## Troubleshooting

### Template not found

```
Error: template not found: project/main.go.tmpl
```

Run `hexago templates list` to see all available templates.

### Template syntax error

```
Error: failed to parse template: unclosed action
```

Validate the template:

```shell
hexago templates validate .hexago/templates/project/main.go.tmpl
```

### Wrong template being used

```shell
hexago templates which project/main.go.tmpl
```

Remember priority: project-local → user-global → embedded.
