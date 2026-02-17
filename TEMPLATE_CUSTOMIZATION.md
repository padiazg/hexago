# Template Customization Guide

## Overview

HexaGo now supports template customization, allowing you to modify the generated code to match your team's preferences and standards.

## Template System

### Template Sources (Priority Order)

HexaGo searches for templates in the following order:

1. **Project-local overrides** (`./.hexago/templates/`) - Highest priority
2. **User-global overrides** (`~/.hexago/templates/`) - Medium priority
3. **Embedded templates** (built into hexago binary) - Fallback

When you customize a template, hexago will use your version instead of the default.

### Template Structure

```
templates/
├── project/                    # Project initialization
│   ├── main.go.tmpl
│   ├── root_cmd.go.tmpl
│   ├── run_cmd.go.tmpl
│   ├── config.go.tmpl
│   └── logger.go.tmpl
├── misc/                       # Project files
│   ├── makefile.tmpl
│   ├── readme.md.tmpl
│   ├── dockerfile.tmpl
│   ├── compose.yaml.tmpl
│   └── gitignore.tmpl
├── service/                    # Business logic
│   ├── service.go.tmpl
│   └── service_test.go.tmpl
├── domain/                     # Domain entities
│   ├── entity.go.tmpl
│   ├── entity_test.go.tmpl
│   ├── value_object.go.tmpl
│   └── value_object_test.go.tmpl
├── adapter/                    # Adapters
│   ├── primary/
│   │   ├── http_handler.go.tmpl
│   │   ├── grpc_handler.go.tmpl
│   │   └── queue_consumer.go.tmpl
│   └── secondary/
│       ├── database_repo.go.tmpl
│       ├── external_service.go.tmpl
│       └── cache_adapter.go.tmpl
├── worker/                     # Background workers
│   ├── queue_worker.go.tmpl
│   ├── periodic_worker.go.tmpl
│   ├── event_worker.go.tmpl
│   └── manager.go.tmpl
└── migration/                  # Database migrations
    ├── migration.up.sql.tmpl
    ├── migration.down.sql.tmpl
    └── migrator.go.tmpl
```

## How to Customize Templates

### Method 1: Export and Modify

```bash
# Export a template to project-local override
hexago templates export project/main.go.tmpl

# Or export to global override (affects all projects)
hexago templates export project/main.go.tmpl --global

# Edit the exported template
vim .hexago/templates/project/main.go.tmpl

# Generate code - will use your custom template
hexago init my-app --module github.com/me/my-app
```

### Method 2: Manual Creation

Create template files manually in the override locations:

```bash
# Project-local override
mkdir -p .hexago/templates/service
cat > .hexago/templates/service/service.go.tmpl << 'EOF'
{{/*
Template: service/service.go
Custom template for my team
*/}}
package {{.CoreLogic}}
// ... your custom template ...
EOF

# Now generate with custom template
hexago add service CreateUser
```

## Template Commands

### List Available Templates

```bash
hexago templates list
```

Shows all available templates.

### Check Which Template Will Be Used

```bash
hexago templates which project/main.go.tmpl
```

Shows the source (project-local, user-global, or embedded) that will be used.

### Export Template

```bash
# Export to project-local (./.hexago/templates/)
hexago templates export project/main.go.tmpl

# Export to user-global (~/.hexago/templates/)
hexago templates export project/main.go.tmpl --global

# Export to custom location
hexago templates export project/main.go.tmpl --output /path/to/template
```

### Validate Template Syntax

```bash
hexago templates validate .hexago/templates/project/main.go.tmpl
```

Checks if your custom template has valid Go template syntax.

### Reset Template

```bash
# Remove project-local override
hexago templates reset project/main.go.tmpl

# Remove user-global override
hexago templates reset project/main.go.tmpl --global
```

## Template Syntax

Templates use Go's `text/template` syntax with custom functions.

### Available Variables

Variables depend on the template. Check the header comment in each template file:

```go
{{/*
Template: project/main.go
Variables:
  - ProjectName: string - Project name
  - ModuleName: string - Go module name
  - Year: string - Current year
  - Author: string - Author name
*/}}
```

### Custom Template Functions

HexaGo provides additional template functions:

```go
{{.ProjectName | upper}}     // Uppercase: MY-APP
{{.ProjectName | lower}}     // Lowercase: my-app
{{.ProjectName | title}}     // Title case: My-App
{{.ServiceName | snake}}     // Snake case: create_user
```

Functions:
- `upper` - Convert to uppercase
- `lower` - Convert to lowercase
- `title` - Title case (capitalize first letter)
- `snake` - Convert to snake_case

## Example Customizations

### Example 1: Add Company Header to All Files

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

### Example 2: Custom Service Template with Logging

Create `.hexago/templates/service/service.go.tmpl`:

```go
package {{.CoreLogic}}

import (
	"context"
	"fmt"

	"{{.ModuleName}}/internal/core/domain"
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

### Example 3: Custom HTTP Handler with Metrics

Create `.hexago/templates/adapter/primary/http_handler.go.tmpl` with your preferred HTTP handler structure including metrics, request tracing, etc.

## Best Practices

### 1. Document Your Templates

Always include a header comment with:
- Template name
- Description
- Required variables
- Optional variables

### 2. Version Your Templates

If your project uses custom templates, commit them to version control:

```bash
git add .hexago/templates/
git commit -m "Add custom hexago templates"
```

This ensures all team members use the same templates.

### 3. Test Your Templates

After customizing, generate a test project to verify:

```bash
# Test project generation
hexago init test-app --module github.com/test/app

# Test service generation
cd test-app
hexago add service TestService

# Verify code compiles
go build
```

### 4. Keep Templates Updated

When upgrading hexago, check if default templates have changed:

```bash
hexago templates diff project/main.go.tmpl
```

(Feature coming soon)

### 5. Share Templates

You can share your custom templates with your team:

```bash
# Export all custom templates
tar czf hexago-templates.tar.gz .hexago/templates/

# On another machine
tar xzf hexago-templates.tar.gz
```

Or use user-global templates for consistent setup across all your projects.

## Troubleshooting

### Template Not Found Error

```
Error: template not found: project/main.go.tmpl
```

**Solution:** The template doesn't exist. Check available templates with `hexago templates list`.

### Template Syntax Error

```
Error: failed to parse template: unclosed action
```

**Solution:** Your custom template has invalid Go template syntax. Validate it:

```bash
hexago templates validate .hexago/templates/project/main.go.tmpl
```

### Wrong Template Being Used

```bash
# Check which template will be used
hexago templates which project/main.go.tmpl
```

Remember the priority order:
1. Project-local (`./.hexago/templates/`)
2. User-global (`~/.hexago/templates/`)
3. Embedded (built-in)

### Variables Not Available

If you get errors about missing variables, check the template header comment to see which variables are provided.

## Migration Status

**Note:** Template externalization is currently in progress. The following templates are available as separate `.tmpl` files:

✅ Available for customization:
- project/main.go.tmpl
- project/root_cmd.go.tmpl
- project/run_cmd.go.tmpl
- project/config.go.tmpl
- project/logger.go.tmpl
- misc/makefile.tmpl
- misc/readme.md.tmpl
- misc/dockerfile.tmpl
- misc/compose.yaml.tmpl
- misc/gitignore.tmpl
- service/service.go.tmpl
- service/service_test.go.tmpl

🚧 Coming soon:
- domain templates
- adapter templates
- worker templates
- migration templates

The remaining templates are still embedded in the code but will be externalized in future updates.

## Contributing Custom Templates

If you create useful custom templates, consider contributing them back to the project!

1. Create a template variant (e.g., `service_with_tracing.go.tmpl`)
2. Add documentation
3. Submit a pull request

## Support

If you have questions or issues with template customization:

- Open an issue: https://github.com/padiazg/hexago/issues
- Discussions: https://github.com/padiazg/hexago/discussions

---

Happy customizing! 🎨
