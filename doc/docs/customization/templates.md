# Template Customization

HexaGo supports template customization, allowing you to modify the generated code to match your team's preferences, coding standards, and branding.

---

## Template Sources

HexaGo searches for templates in the following order (highest priority first):

| Priority | Location | Use Case |
|----------|----------|----------|
| **1 — Highest** | `./.hexago/templates/` | Per-project customization |
| **2 — Medium** | `~/.hexago/templates/` | User-wide defaults |
| **3 — Fallback** | Embedded in binary | Default templates |

When you provide a custom template, HexaGo uses it instead of the built-in default.

---

## Template Structure

```
templates/
├── project/                    # Project initialization files
│   ├── main.go.tmpl
│   ├── root_cmd.go.tmpl
│   ├── run_cmd.go.tmpl
│   ├── config.go.tmpl
│   └── logger.go.tmpl
├── misc/                       # Project support files
│   ├── makefile.tmpl
│   ├── readme.md.tmpl
│   ├── dockerfile.tmpl
│   ├── compose.yaml.tmpl
│   └── gitignore.tmpl
├── service/                    # Business logic templates
│   ├── service.go.tmpl
│   └── service_test.go.tmpl
├── domain/                     # Domain entity templates
│   ├── entity.go.tmpl
│   ├── entity_test.go.tmpl
│   ├── value_object.go.tmpl
│   └── value_object_test.go.tmpl
├── adapter/                    # Adapter templates
│   ├── primary/
│   │   ├── http_handler.go.tmpl
│   │   ├── grpc_handler.go.tmpl
│   │   └── queue_consumer.go.tmpl
│   └── secondary/
│       ├── database_repo.go.tmpl
│       ├── external_service.go.tmpl
│       └── cache_adapter.go.tmpl
├── worker/                     # Background worker templates
│   ├── queue_worker.go.tmpl
│   ├── periodic_worker.go.tmpl
│   ├── event_worker.go.tmpl
│   └── manager.go.tmpl
└── migration/                  # Database migration templates
    ├── migration.up.sql.tmpl
    ├── migration.down.sql.tmpl
    └── migrator.go.tmpl
```

---

## Template Commands

### List available templates

```shell
hexago templates list
```

### Check which template will be used

```shell
hexago templates which project/main.go.tmpl
```

Shows whether the project-local, user-global, or embedded template will be used.

### Export a template for editing

```shell
# Export to project-local (./.hexago/templates/)
hexago templates export project/main.go.tmpl

# Export to user-global (~/.hexago/templates/)
hexago templates export project/main.go.tmpl --global

# Export to a specific location
hexago templates export project/main.go.tmpl --output /path/to/template
```

### Validate template syntax

```shell
hexago templates validate .hexago/templates/project/main.go.tmpl
```

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
vim .hexago/templates/service/service.go.tmpl

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
vim ~/.hexago/templates/service/service.go.tmpl
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

!!! success "Available now"
    - `project/main.go.tmpl`
    - `project/root_cmd.go.tmpl`
    - `project/run_cmd.go.tmpl`
    - `project/config.go.tmpl`
    - `project/logger.go.tmpl`
    - `misc/makefile.tmpl`
    - `misc/readme.md.tmpl`
    - `misc/dockerfile.tmpl`
    - `misc/compose.yaml.tmpl`
    - `misc/gitignore.tmpl`
    - `service/service.go.tmpl`
    - `service/service_test.go.tmpl`

!!! info "Coming soon"
    - Domain templates
    - Adapter templates
    - Worker templates
    - Migration templates

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
