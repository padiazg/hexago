# Secondary Database Adapter

How to add a database repository adapter to a HexaGo-generated project.

---

## Overview

Secondary adapters (outbound) implement interfaces defined by the core services. Database adapters persist and retrieve domain entities.

---

## Generate with HexaGo

Use the `hexago add adapter` command:

```shell
hexago add adapter secondary database UserRepository \
  --entity User
```

This generates:

- `internal/adapters/secondary/database/user_repository.go`
- `internal/adapters/secondary/database/user_repository_test.go`

---

## Generated Structure

After generation, your adapter looks like this:

```go
package database

import (
    "context"
    "database/sql"

    "myapp/internal/core/domain"
    "myapp/internal/core/services/ports"
)

var _ ports.Store = (*UserRepository)(nil)

type UserRepository struct {
    db *sql.DB
}

func New(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}
```

---

## Implementing the Port

First, ensure your service defines the port interface. Then implement it:

### 1. Define the Port (in services)

```go
// internal/core/services/ports/user_store.go
package ports

type UserStore interface {
    GetUser(ctx context.Context, id string) (*domain.User, error)
    SaveUser(ctx context.Context, user *domain.User) error
    ListUsers(ctx context.Context) ([]*domain.User, error)
    DeleteUser(ctx context.Context, id string) error
}
```

### 2. Implement the Adapter

```go
// internal/adapters/secondary/database/user_repository.go
package database

import (
    "context"
    "fmt"

    "myapp/internal/core/domain"
    "myapp/internal/core/services/ports"
)

type UserRepository struct {
    db *sql.DB
}

func New(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
    const q = `SELECT id, email, name, created_at FROM users WHERE id = ?`
    row := r.db.QueryRowContext(ctx, q, id)

    var user domain.User
    var createdAt string
    err := row.Scan(&user.ID, &user.Email, &user.Name, &createdAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("repository get user: %w", err)
    }

    user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
    return &user, nil
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
    const q = `INSERT INTO users (id, email, name, created_at) VALUES (?, ?, ?, ?)`
    _, err := r.db.ExecContext(ctx, q,
        user.ID,
        user.Email,
        user.Name,
        user.CreatedAt.Format(time.RFC3339),
    )
    if err != nil {
        return fmt.Errorf("repository save user: %w", err)
    }
    return nil
}
```

---

## Migrations

Add database migrations for your tables:

```shell
hexago add migration create_users
```

This creates:

- `migrations/000001_create_users.up.sql`
- `migrations/000001_create_users.down.sql`

### Example Migration

```sql
-- migrations/000001_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- migrations/000001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

---

## Database Connection

### Opening a Database

```go
// internal/adapters/secondary/database/postgres.go
package database

import (
    "database/sql"
    "fmt"

    _ "github.com/lib/pq"
)

func Open(cfg Config) (*sql.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
    )

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return db, nil
}
```

### Using SQLite (No CGO)

```go
import _ "modernc.org/sqlite"

func Open(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(1) // SQLite single-writer
    return db, nil
}
```

---

## Wire-Up in cmd/run.go

Connect the repository to the service in your main command:

```go
// cmd/run.go
func run() error {
    // Open database
    db, err := database.Open(cfg.Database)
    if err != nil {
        return err
    }
    defer db.Close()

    // Run migrations
    if err := database.Migrate(db); err != nil {
        return err
    }

    // Create adapter
    userRepo := database.NewUser(db)

    // Create service with adapter
    userSvc := services.NewUserService(userRepo, logger)

    // Create handler
    handler := http.NewUserHandler(userSvc)

    // Start server
    return httpServer.Serve(handler.Routes())
}
```

---

## Testing

### Unit Test with Mock

```go
// internal/adapters/secondary/database/user_repository_test.go
package database

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockDB struct {
    mock.Mock
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
    return nil // simplified
}

func TestUserRepository_GetUser(t *testing.T) {
    // Arrange
    mockDB := new(MockDB)
    repo := New(mockDB)

    // Act & Assert
    user, err := repo.GetUser(context.Background(), "123")
    assert.NoError(t, err)
    assert.Nil(t, user)
}
```

### Integration Test

```go
// +build integration

package database

import (
    "testing"

    "github.com/stretchr/testify/require"
)

func TestIntegration_UserRepository(t *testing.T) {
    db, err := Open(":memory:")
    require.NoError(t, err)
    defer db.Close()

    repo := New(db)
    // Test actual DB operations
}
```

---

## Best Practices

| Practice | Description |
| --- | --- |
| **Compile-time interface check** | Add `var _ ports.UserStore = (*UserRepository)(nil)` |
| **Context propagation** | All DB methods take `context.Context` |
| **Error wrapping** | Wrap errors with operation context |
| **Close defer** | Always defer `db.Close()` |
| **Connection pooling** | Configure `SetMaxOpenConns`, `SetMaxIdleConns` |
| **Parameterized queries** | Never concatenate strings into SQL |
