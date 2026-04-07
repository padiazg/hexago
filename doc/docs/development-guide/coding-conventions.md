# Coding Conventions

Standard patterns and conventions for writing Go code in HexaGo-generated projects.

---

## Context Usage

All I/O operations must accept `context.Context` as the first argument:

```go
// ✓ Correct - context first
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    return s.store.GetUser(ctx, id)
}

// ✗ Wrong - missing context
func (s *Service) GetUser(id string) (*User, error) {
    return s.store.GetUser(id)
}
```

This enables:
- Cancellation propagation
- Timeout management
- Tracing correlation

---

## Error Handling

### Wrapping Errors

Always wrap errors with context:

```go
// ✓ Correct - wrapped with operation context
if err := s.store.Save(ctx, user); err != nil {
    return fmt.Errorf("save user: %w", err)
}

// ✗ Wrong - lost error context
if err := s.store.Save(ctx, user); err != nil {
    return err
}
```

### Error Patterns

```go
// Sentinel errors for expected conditions
var ErrNotFound = errors.New("entity not found")

// Custom error types for rich error information
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Error wrapping in business logic
func (s *Service) Process(ctx context.Context, req *Request) error {
    if err := s.validate(req); err != nil {
        return fmt.Errorf("validate request: %w", err)
    }
    // ...
}
```

---

## Logging

Use the project's Logger interface — never use `fmt.Println` or the standard `log` package:

```go
// ✓ Correct - use structured logger
s.logger.Info("processing request", "request_id", id)
s.logger.Debug("cache hit", "key", key)
s.logger.Error("request failed", "error", err)

// ✗ Wrong - never use these
fmt.Println("processing request")
log.Printf("request failed: %v", err)
```

### Log Levels

| Level | Use For |
|-------|---------|
| `Debug` | Development info, verbose state |
| `Info` | Normal operation events |
| `Warn` | Recoverable issues, degraded behavior |
| `Error` | Failures that need attention |

---

## Constructors — No init()

Avoid `init()` functions. Use explicit constructors:

```go
// ✓ Correct - explicit constructor
type Service struct {
    store Store
    logger Logger
}

func NewService(store Store, logger Logger) *Service {
    return &Service{
        store: store,
        logger: logger,
    }
}

// ✗ Wrong - init() for initialization
var service *Service

func init() {
    service = &Service{...}
}
```

Benefits:
- Dependencies are explicit and testable
- No hidden initialization order
- Easy to create multiple instances

---

## Testing Conventions

### Assertions

Use `testify/assert` for non-fatal assertions and `testify/require` for fatal ones:

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Non-fatal - continue on failure, good for cleanup
assert.Equal(t, expected, actual, "they should be equal")

// Fatal - stop immediately on failure, good for prerequisites
require.NotNil(t, user, "user must exist")
```

### Test Organization

```go
func TestService_Create(t *testing.T) {
    // Arrange
    service := NewService(mockStore, logger)
    req := &CreateRequest{Name: "test"}

    // Act
    result, err := service.Create(context.Background(), req)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "test", result.Name)
}
```

---

## Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, short | `domain`, `services` |
| Interfaces | PascalCase, ending with er | `Store`, `Provider` |
| Structs | PascalCase | `UserService`, `OrderHandler` |
| Methods | PascalCase | `GetByID`, `Create` |
| Variables | camelCase | `userID`, `orderTotal` |
| Constants | PascalCase | `DefaultTimeout`, `MaxRetries` |
| Test functions | `Test<Subject>_<Scenario>` | `TestUser_Create_ValidInput` |

---

## Imports

Organize imports in three groups (go fmt handles this automatically):

1. Standard library
2. Third-party packages
3. Internal packages

```go
import (
    "context"
    "fmt"

    "github.com/stretchr/testify/assert"
    "go.uber.org/zap"

    "myapp/internal/core/domain"
    "myapp/internal/core/services/ports"
)
```

---

## Configuration

Never hardcode configuration values. Use the config system:

```go
type Config struct {
    ServerPort    int           `mapstructure:"server_port"`
    DatabaseURL   string        `mapstructure:"database_url"`
    LogLevel      string        `mapstructure:"log_level"`
    Timeout       time.Duration `mapstructure:"timeout"`
}
```

Configuration is read in priority order:

1. Environment variables (`MY_APP_*`)  
2. Config file (`./.my-app.yaml`)  
3. Home directory (`~/.my-app.yaml`)  
4. Defaults  
