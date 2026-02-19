# hexago add tool

Add an infrastructure tool to an existing project.

## Synopsis

```shell
hexago add tool <type> <name> [flags]
```

Must be run from the project root directory.

---

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--description` | | string | `""` | Description of the tool |

---

## Tool Types

| Type | Description | Typical Use |
|------|-------------|-------------|
| `logger` | Structured logger | Wrapping zap, logrus, or slog |
| `validator` | Input validator | Request validation, business rule validation |
| `mapper` | Object mapper | Converting between domain and DTO types |
| `middleware` | HTTP middleware | Auth, rate limiting, request logging |

---

## Examples

```shell
hexago add tool logger StructuredLogger
hexago add tool validator RequestValidator --description "Validates incoming HTTP requests"
hexago add tool mapper UserMapper --description "Maps between User domain and UserDTO"
hexago add tool middleware AuthMiddleware --description "JWT authentication middleware"
hexago add tool middleware RateLimitMiddleware
hexago add tool middleware CORSMiddleware
```

---

## Generated Files

For `hexago add tool validator PostValidator`:

```
internal/
└── tools/
    └── post_validator.go
```

---

## Generated Code Structure

**Validator:**

```go
package tools

// PostValidator validates Post-related inputs
type PostValidator struct{}

// NewPostValidator creates a new PostValidator
func NewPostValidator() *PostValidator {
    return &PostValidator{}
}

// Validate validates the given input
func (v *PostValidator) Validate(input interface{}) error {
    // TODO: Implement validation logic
    return nil
}
```

**Middleware:**

```go
package tools

import "net/http"

// AuthMiddleware provides JWT authentication
type AuthMiddleware struct {
    // TODO: Add dependencies (token validator, etc.)
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware() *AuthMiddleware {
    return &AuthMiddleware{}
}

// Handle wraps an http.Handler with authentication
func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // TODO: Implement authentication logic
        next.ServeHTTP(w, r)
    })
}
```

**Mapper:**

```go
package tools

// UserMapper converts between domain and DTO types
type UserMapper struct{}

// NewUserMapper creates a new UserMapper
func NewUserMapper() *UserMapper {
    return &UserMapper{}
}

// ToDTO converts a domain User to UserDTO
func (m *UserMapper) ToDTO(user interface{}) interface{} {
    // TODO: Implement mapping
    return nil
}

// ToDomain converts a UserDTO to domain User
func (m *UserMapper) ToDomain(dto interface{}) interface{} {
    // TODO: Implement mapping
    return nil
}
```
