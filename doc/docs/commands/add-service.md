# hexago add service

Add a business logic service (use case) to an existing project.

## Synopsis

```shell
hexago add service <name> [flags]
```

Operates on the project root — use `--working-directory` (`-w`) to target a project without changing directories.

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--description` | `-d` | string | `""` | Description of what the service does |
| `--entity` | `-e` | string | `""` | Domain entity this service manages (PascalCase); determines sub-package name |
| `--from-port` | | string | `""` | Port interface name to infer method signatures from |
| `--infer-tests` | | bool | `false` | Generate tests with method signatures from port |

---

## Examples

```shell
hexago add service CreateUser
hexago add service GetUser
hexago add service SendEmail --description "Sends email notifications"
hexago add service ProcessOrder --description "Handles order processing"
```

---

## Generated Files

Services are placed in their own sub-package under the core logic directory:

For `hexago add service CreateUser --entity User`:

```
internal/core/services/create_user/
├── create_user.go         # Service implementation
├── create_user_test.go    # Test file
└── services.go            # Service aggregator (updated on each add)
```

!!! note
    If your project uses `--core-logic usecases`, files are placed in `internal/core/usecases/` instead.
    When `--entity` is provided, the sub-package name is derived from the entity (e.g. `categories/` for `Category`).

---

## Generated Code Structure

The generated service file provides a scaffold with:

- Input and output types
- Service struct with dependency placeholders
- Constructor function
- `Execute` method with `// TODO` comments

```go
package services

import (
    "context"
    "fmt"
)

// CreateUserInput represents the input for CreateUser
type CreateUserInput struct {
    // TODO: Add input fields
}

// CreateUserOutput represents the output for CreateUser
type CreateUserOutput struct {
    // TODO: Add output fields
}

// CreateUserService Creates a new user
type CreateUserService struct {
    // TODO: Add dependencies (repositories, external services)
}

// NewCreateUserService creates a new instance
func NewCreateUserService() *CreateUserService {
    return &CreateUserService{}
}

// Execute runs the CreateUser use case
func (s *CreateUserService) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
    // TODO: Implement business logic
    return nil, fmt.Errorf("not implemented")
}
```

---

## Architecture Notes

Services belong to the **core layer** and must:

- ✅ Define business logic
- ✅ Define port interfaces (for repositories, external services)
- ✅ Be framework-agnostic
- ❌ Never import adapter packages
- ❌ Never import infrastructure packages directly
