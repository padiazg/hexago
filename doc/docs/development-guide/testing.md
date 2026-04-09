# Testing

How to write and run tests in HexaGo-generated projects.

---

## Test Types

| Type | Description | How to Run |
| --- | --- | --- |
| **Unit tests** | Test isolated components, no external deps | `go test ./...` |
| **Integration tests** | Test with real external services | `go test -tags=integration ./...` |

Unit tests must pass without any external service (database, API, etc.).

---

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test ./... -run TestUser_Create

# Run tests in specific package
go test ./internal/core/domain/...

# Run with race detector
go test -race ./...
```

### Code Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View HTML coverage report
go tool cover -html=coverage.out

# Show coverage by function
go tool cover -func=coverage.out
```

### Integration Tests

```bash
# Run only integration tests
go test -tags=integration ./...

# Run with verbose output
go test -tags=integration -v ./...
```

Integration tests require API keys or database connections configured in `.env`.

---

## Test Organization

### File Naming

```shell
internal/core/domain/
├── user.go
├── user_test.go      ← unit tests
└── user_integration_test.go  ← integration tests (build tag)
```

### Test Structure

Follow AAA pattern (Arrange, Act, Assert):

```go
func TestUserService_Create(t *testing.T) {
    // Arrange - set up test data and mocks
    mockStore := &MockStore{}
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    service := NewUserService(mockStore, logger)

    req := &CreateUserRequest{
        Email: "test@example.com",
        Name:  "Test User",
    }

    // Act - execute the function under test
    result, err := service.Create(context.Background(), req)

    // Assert - verify the outcome
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", result.Email)
    assert.Equal(t, "Test User", result.Name)
}
```

---

## Test Dependencies

### Testify

Use `testify` for assertions and mocking:

```go
import (
    "github.com/stretchr/testify/assert"   // non-fatal assertions
    "github.com/stretchr/testify/require"   // fatal assertions
    "github.com/stretchr/testify/mock"     // mocking
)
```

| Package | Use Case |
| --- | --- |
| `assert` | Continue on failure, good for multiple checks |
| `require` | Stop on failure, good for prerequisites |
| `mock` | Create mock objects for interfaces |

### Example with Mocks

```go
type MockStore struct {
    mock.Mock
}

func (m *MockStore) GetUser(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func TestGetUser(t *testing.T) {
    mockStore := new(MockStore)

    // Set up expectations
    mockStore.On("GetUser", mock.Anything, "123").Return(&User{ID: "123"}, nil)

    service := NewUserService(mockStore)
    user, err := service.GetUser(context.Background(), "123")

    require.NoError(t, err)
    assert.Equal(t, "123", user.ID)

    // Verify all expectations were met
    mockStore.AssertExpectations(t)
}
```

---

## Build Tags

Use build tags to separate unit and integration tests:

```go
// +build integration

package myapp_test

// Integration tests require real services
func TestRealDatabase(t *testing.T) {
    // This test only runs with -tags=integration
}
```

```go
// Unit tests - no build tag needed
package myapp_test

func TestUnit(t *testing.T) {
    // This test always runs
}
```

---

## Test Utilities

### Test Fixtures

```go
func TestMain(m *testing.M) {
    // Setup before all tests
    os.Exit(m.Run())
    // Teardown after all tests
}
```

### Helper Functions

```go
func mustCreateUser(t *testing.T, email string) *User {
    user := &User{
        ID:    uuid.New(),
        Email: email,
    }
    require.NoError(t, user.Validate())
    return user
}
```

---

## Table-Driven Tests

For testing multiple input combinations:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name  string
        email string
        valid bool
    }{
        {"valid email", "test@example.com", true},
        {"invalid - no @", "testexample.com", false},
        {"invalid - no domain", "test@", false},
        {"empty", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if tt.valid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

---

## Best Practices

1. **Test behavior, not implementation** — Test public interfaces
2. **Name tests descriptively** — `TestService_Create_ValidInput`
3. **One assertion per test is not required** — Group related assertions
4. **Use `require` for prerequisites** — Fail fast on missing preconditions
5. **Use `assert` for actual checks** — Continue to see all failures
6. **No external deps in unit tests** — Mock everything external
7. **Clean up in tests** — Use `t.Cleanup()` for resources
