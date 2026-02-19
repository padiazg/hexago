# hexago add service

Add a business logic service (use case) to an existing project.

## Synopsis

```shell
hexago add service <name> [flags]
```

Must be run from the project root directory.

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--description` | | string | `""` | Description of what the service does |

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

For `hexago add service CreateUser`:

```
internal/core/services/
├── create_user.go         # Service implementation
└── create_user_test.go    # Test file
```

!!! note
    If your project uses `--core-logic usecases`, files are placed in `internal/core/usecases/` instead.

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
