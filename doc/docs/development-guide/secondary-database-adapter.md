# Secondary Database Adapter

How to add a database repository adapter to a HexaGo-generated project.

---

## Overview

**Secondary adapters** (also known as "driven" or "outbound" adapters) implement interfaces (ports) defined by the core services. In Hexagonal Architecture:

- **Primary/Driver adapters** (inbound) drive the application — HTTP handlers, gRPC servers, CLI commands
- **Secondary/Driven adapters** (outbound) are driven by the application — databases, external APIs, message queues

The core (domain services) never depends on adapters. Instead, adapters implement ports (interfaces) that the core defines. This keeps your business logic framework-agnostic and testable.

In this guide, we'll create a SQLite-based User repository that persists and retrieves User entities.

---

## Initialize Project

Create the project with all required features:

```shell
$ hexago init user-manager \
  --project-type service \
  --explicit-ports \
  --with-migrations \
  --module github.com/padiazg/user-manager

📋 Project Configuration:
  Name:              user-manager
  Module:            github.com/padiazg/user-manager
  Project Type:      service
  Adapter Style:     primary-secondary
  Core Logic:        services
  Docker:            false
  Observability:     false
  Migrations:        true
  Workers:           false
  Example Code:      false

🚀 Generating project user-manager...
📁 Creating directory structure...
📝 Generating files...
📦 Initializing go module...
go: creating new go.mod: module github.com/padiazg/user-manager
go: to add module requirements and sums:
    go mod tidy
📦 Adding dependencies...
🧹 Running go mod tidy...
✨ Formatting code...

✅ Project generated successfully!

📚 Next steps:
  cd user-manager
  go run main.go run

📖 Read the README.md for more information about the project structure.
```

This creates:

- `cmd/` - CLI commands (root.go, run.go)  
- `internal/core/` - Domain and services  
- `internal/adapters/` - Primary and secondary adapters  
- `migrations/` - Database migrations  
- `pkg/` - Reusable packages (logger)  
- `main.go`, `Makefile`, etc.  

---

## Add Database Path to Config

For this example we'll add a database file path to the config. This allows users to customize where the SQLite database is stored. You can skip this step and hardcode the path in the repository `Open` function if you prefer.

```go
// internal/config/config.go
package config

import (
   "fmt"
   "time"

   "github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
   Server    ServerConfig
   LogLevel  string
   LogFormat string
   DBPath    string    // add this
}

...

func setDefaults() {
   // Server defaults
   viper.SetDefault("dbpath", "./user-manager.db") // add this
...
}
```

---

## Add a domain entity

### 1. Generate the entity

```shell
$ hexago add domain entity User \
  --fields "id:string,name:string,email:string"
  
📦 Adding domain entity: User
   Project: user-manager

📝 Creating entity file: internal/core/domain/users/users.go
📝 Creating port file: internal/core/domain/users/port.go
📝 Creating test file: internal/core/domain/users/users_test.go

✅ Domain entity added successfully!

📝 Next steps:
  1. Add business logic methods to the entity
  2. Add validation rules
  3. Write tests for domain logic
```

This generates:

- `internal/core/domain/users/users.go`
- `internal/core/domain/users/port.go`
- `internal/core/domain/users/users_test.go`

### 2. Update the entity code

The generated entity file includes basic structure. Let's update it with proper validation.

```go
// internal/core/domain/users/users.go
package users

import (
    "errors"
)

// User represents a User entity in the domain.
// This is a domain entity with unique identity and business logic.
type User struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
}

// NewUser creates a new User with validation
func NewUser(id, name, email string) (*User, error) {
    entity := &User{
        ID:    id,
        Name:  name,
        Email: email,
    }

    if err := entity.Validate(); err != nil {
        return nil, err
    }

    return entity, nil
}

// Validate ensures the User entity is in a valid state
func (e *User) Validate() error {
    if e.ID == "" {
        return errors.New("id cannot be empty")
    }

    if e.Name == "" {
        return errors.New("name cannot be empty")
    }

    return nil
}
```

Now update `port.go` to define the repository port interface.

```go
// internal/core/domain/users/port.go
package users

import "context"

// UserRepository defines the secondary port for User persistence.
type UserRepository interface {
    CreateUser(ctx context.Context, user *UserCreateRequest) (*User, error)
    FindByID(ctx context.Context, id string) (*User, error)
    UpdateEmail(ctx context.Context, id, email string) error
    List(ctx context.Context, limit int) ([]*User, error)
}

type UserCreateRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

---

## Add a Secondary Adapter

A **secondary adapter** implements the port interface we defined earlier (`UserRepository`). It contains the actual logic for interacting with the database. The key principle of Hexagonal Architecture is that the domain and services remain agnostic to how the database works—they only know about the port interface.

### 1. Generate the adapter

```shell
$ hexago add adapter secondary database UserRepository \
  --entity User

📦 Adding secondary adapter: UserRepository (database)
   Project: user-manager
   Adapter dir: secondary

📝 Creating adapter file: internal/adapters/secondary/database/users/users.go
📝 Creating test file: internal/adapters/secondary/database/users/users_test.go

✅ Secondary adapter added successfully!

📝 Next steps:
  1. Implement the port interface methods
  2. Add database queries or external API calls
  3. Wire up dependencies in the DI container
```

This generates:

- `internal/adapters/secondary/database/users/users.go`
- `internal/adapters/secondary/database/users/users_test.go`

### 2. Update the adapter code

The generated adapter implements our `UserRepository` port using SQLite. It includes a compile-time check to ensure the interface is satisfied.

```go
// internal/adapters/secondary/database/users/users.go
package users

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/google/uuid"
    usersDomain "github.com/padiazg/user-manager/internal/core/domain/users"
)

// UserRepository implements usersDomain.UserRepository using SQLite.
type UserRepository struct {
    db *sql.DB
}

// compile-time check that UserRepository satisfies the port.
var _ usersDomain.UserRepository = (*UserRepository)(nil)

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Open opens a SQLite database at the given path.
func Open(path string) (*sql.DB, error) {
    if path == "" {
        return nil, fmt.Errorf("open database: must provide a path")
    }

    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }
    db.SetMaxOpenConns(1) // SQLite single-writer
    return db, nil
}

// Create inserts a new User.
func (r *UserRepository) CreateUser(ctx context.Context, req *usersDomain.UserCreateRequest) (*usersDomain.User, error) {
    const q = `INSERT INTO users (id, email, name, created_at) VALUES (?, ?, ?, ?)`

    res := &usersDomain.User{
        ID:        uuid.New().String(),
        Name:      req.Name,
        Email:     req.Email,
        CreatedAt: time.Now().Format(time.RFC3339),
    }

    _, err := r.db.ExecContext(ctx, q,
        res.ID,
        res.Email,
        res.Name,
        res.CreatedAt,
    )

    if err != nil {
        return nil, fmt.Errorf("repository save user: %w", err)
    }

    return res, nil
}

// FindByID retrieves a User by its ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*usersDomain.User, error) {
    const q = `SELECT id, email, name, created_at FROM users WHERE id = ?`
    var user usersDomain.User

    row := r.db.QueryRowContext(ctx, q, id)
    err := row.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("repository get user: %w", err)
    }

    return &user, nil
}

// Update saves updated User fields.
func (r *UserRepository) UpdateEmail(ctx context.Context, id, email string) error {
    const q = `UPDATE users SET email=? WHERE id = ?`

    if _, err := r.db.ExecContext(ctx, q, email, id); err != nil {
        return fmt.Errorf("repository update user: %w", err)
    }

    return nil
}

// List returns all User records.
func (r *UserRepository) List(ctx context.Context, limit int) ([]*usersDomain.User, error) {
    const q = `SELECT id, email, name, created_at FROM users ORDER BY created_at DESC LIMIT ?`

    rows, err := r.db.QueryContext(ctx, q, limit)
    if err != nil {
        return nil, fmt.Errorf("repository user list: %w", err)
    }
    defer rows.Close()

    var res []*usersDomain.User

    for rows.Next() {
        var user usersDomain.User
        if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
            return nil, fmt.Errorf("repository user scan: %w", err)
        }
        res = append(res, &user)
    }

    return res, rows.Err()
}

```

---

## Migrations

### 1. Generate the migrator

Add database migrations for the table:

```shell
$ hexago add migration create_users

📦 Adding migration: create_users
   Project: user-manager
   Type: sql

📝 Creating migration files:
   UP:   migrations/000001_create_users.up.sql
   DOWN: migrations/000001_create_users.down.sql
📝 Creating migration manager: internal/infrastructure/database/migrator.go

ℹ️  Add these commands to your Makefile:

migrate-up: ## Run database migrations
    @migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Rollback last migration
    @migrate -path migrations -database "$(DB_URL)" down 1

migrate-version: ## Show current migration version
    @migrate -path migrations -database "$(DB_URL)" version

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
    @migrate -path migrations -database "$(DB_URL)" force $(VERSION)

# Add DB_URL to your environment or Makefile:
# DB_URL=postgresql://user:password@localhost:5432/dbname?sslmode=disable

✅ Migration added successfully!

📝 Files created:
   - migrations/000001_create_users.up.sql
   - migrations/000001_create_users.down.sql

📝 Next steps:
  1. Edit the .up.sql file with your schema changes
  2. Edit the .down.sql file to reverse those changes
  3. Run migrations:
     make migrate-up
  4. To rollback:
     make migrate-down
```

This creates:

- `migrations/000001_create_users.up.sql`
- `migrations/000001_create_users.down.sql`

### 2. Update the migration files

Create the SQL migration files:

```sql
-- migrations/000001_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

```sql
-- migrations/000001_create_users.down.sql
DROP TABLE IF EXISTS users;
```

### 3. Update the migrator for SQLite

The generated migrator uses PostgreSQL by default. Update it to use SQLite:

```go
// internal/infrastructure/database/migrator.go
package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	sqlitemig "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"

	"github.com/padiazg/user-manager/pkg/logger"
)

// Migrator handles database migrations using golang-migrate
type Migrator struct {
	db     *sql.DB
	logger logger.Logger
}

// MigratorConfig is the configuration data for the migrator
type MigratorConfig struct {
	DB     *sql.DB
	Logger logger.Logger
}

// NewMigrator creates a new migration manager
func NewMigrator(cfg *MigratorConfig) *Migrator {
	return &Migrator{
		db:     cfg.DB,
		logger: cfg.Logger,
	}
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	migration, err := m.getMigration()
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	m.logger.Info("Running migrations...")
	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	m.logger.Info("Migrations completed successfully")
	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	migration, err := m.getMigration()
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	m.logger.Info("Rolling back migration...")
	if err := migration.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	m.logger.Info("Migration rolled back successfully")
	return nil
}

// Version returns the current migration version
func (m *Migrator) Version() (uint, bool, error) {
	migration, err := m.getMigration()
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	version, dirty, err := migration.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}

// getMigration creates a migrate instance
func (m *Migrator) getMigration() (*migrate.Migrate, error) {
	driver, err := sqlitemig.WithInstance(m.db, &sqlitemig.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite",
		driver,
	)
}
```

### 4. Wire-up the migrator to a command

Now we need to wire up the migrator to a command:

```go
// cmd/migrate.go
package cmd

import (
	"fmt"

	usersRepo "github.com/padiazg/user-manager/internal/adapters/secondary/database/users"
	"github.com/padiazg/user-manager/internal/infrastructure/database"
	"github.com/padiazg/user-manager/pkg/logger"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
   Use:   "migrate",
   Short: "Database migration management",
   Long:  `Run, rollback or inspect database migrations.`,
}

var migrateUpCmd = &cobra.Command{
   Use:   "up",
   Short: "Apply all pending migrations",
   RunE: func(cmd *cobra.Command, args []string) error {
      return withMigrator(func(m *database.Migrator, _ logger.Logger) error {
         return m.Up()
      })
   },
}

var migrateDownCmd = &cobra.Command{
   Use:   "down",
   Short: "Roll back the last applied migration",
   RunE: func(cmd *cobra.Command, args []string) error {
      return withMigrator(func(m *database.Migrator, _ logger.Logger) error {
         return m.Down()
      })
   },
}

var migrateVersionCmd = &cobra.Command{
   Use:   "version",
   Short: "Show the current migration version",
   RunE: func(cmd *cobra.Command, args []string) error {
      return withMigrator(func(m *database.Migrator, log logger.Logger) error {
         version, dirty, err := m.Version()
         if err != nil {
            return err
         }
         dirtyFlag := ""
         if dirty {
            dirtyFlag = " (dirty)"
         }
         log.Info("Current migration version: %d%s\n", version, dirtyFlag)
         return nil
      })
   },
}

// withMigrator opens the DB, creates a Migrator and calls fn, then closes the DB.
func withMigrator(fn func(*database.Migrator, logger.Logger) error) error {
   cfg := GetConfig()
   log := logger.New(&logger.Config{
      Level:  cfg.LogLevel,
      Format: cfg.LogFormat,
   })

   db, err := usersRepo.Open(cfg.DBPath)
   if err != nil {
      return fmt.Errorf("opening database: %w", err)
   }
   defer db.Close()

   if err := db.Ping(); err != nil {
      return fmt.Errorf("connecting to database: %w", err)
   }

   return fn(database.NewMigrator(&database.MigratorConfig{
      DB:     db,
      Logger: log,
   }), log)
}

func init() {
   migrateCmd.AddCommand(migrateUpCmd)
   migrateCmd.AddCommand(migrateDownCmd)
   migrateCmd.AddCommand(migrateVersionCmd)
   rootCmd.AddCommand(migrateCmd)
}
```

Let's test it before we go further:

```shell
$ go run main.go migrate up
2026/04/09 23:11:03 [INFO] Running migrations...
2026/04/09 23:11:03 [INFO] Migrations completed successfully

$ ls -l *.db
-rw-r--r-- 1 pato pato 28672 Apr  9 23:11 user-manager.db
```

The database file is created automatically when migrations run.

---

## Add a Service

### 1. Generate the service

```shell
$ hexago add service User 

📦 Adding service: User
   Project: user-manager
   Module: github.com/padiazg/user-manager
   Logic dir: services

📝 Creating service file: internal/core/services/user/user.go
📝 Creating test file: internal/core/services/user/user_test.go
📝 Updating services aggregator: internal/core/services/services.go

✅ Service added successfully!

📝 Next steps:
  1. Implement the business logic in the Execute method
  2. Add any required dependencies to the constructor
  3. Write tests in the generated test file
```

This creates:

- `internal/core/services/user/user.go`
- `internal/core/services/user/user_test.go`
- `internal/core/services/services.go`

### 2. Update service code

```go
// internal/core/services/user/user.go
package user

import (
    "context"
    "fmt"

    userDomain "github.com/padiazg/user-manager/internal/core/domain/users"
)

// UserService implements User logic
type Service struct {
    repository userDomain.UserRepository
}

type Config struct {
    Repository userDomain.UserRepository
}

// NewUserService creates a new UserService.
func New(cfg *Config) *Service {
    return &Service{
        repository: cfg.Repository,
    }
}

func (s *Service) CreateUser(ctx context.Context, req *userDomain.UserCreateRequest) (*userDomain.User, error) {
    if req.Name == "" {
        return nil, fmt.Errorf("must provide a name")
    }

    return s.repository.CreateUser(ctx, req)
}

func (s *Service) FindByID(ctx context.Context, id string) (*userDomain.User, error) {
    return s.repository.FindByID(ctx, id)
}

func (s *Service) UpdateEmail(ctx context.Context, id, email string) error {
    return s.repository.UpdateEmail(ctx, id, email)
}

func (s *Service) List(ctx context.Context, limit int) ([]*userDomain.User, error) {
    return s.repository.List(ctx, limit)
}
```

```go
// internal/core/services/services.go 
package services

import (
    userDomain "github.com/padiazg/user-manager/internal/core/domain/users"
    userSvc "github.com/padiazg/user-manager/internal/core/services/user"
)

// Config holds the repository dependencies required to initialise entity-bound services.
type Config struct {
    UserRepository userDomain.UserRepository
}

// Services aggregates all domain services.
type Services struct {
    User *userSvc.Service
}

// New wires all services using the provided repository config.
func New(config *Config) *Services {
    return &Services{
        User: userSvc.New(&userSvc.Config{
            Repository: config.UserRepository,
        }),
    }
}
```

---

## Wire-up in `cmd`

We won't use the original `cmd/run.go` command to start a server. Instead, we'll implement our own one-time run commands (like a CLI tool). This approach is useful for CLI applications that need to perform specific tasks rather than running a long-lived server.

> It's safe to remove `cmd/run.go` and `internal/core/services/processor.go` if they're not needed.

All the commands use timeout contexts and OS signals for graceful shutdown.

### 1. Add command

```go
package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    userRepository "github.com/padiazg/user-manager/internal/adapters/secondary/database/users"
    userDomain "github.com/padiazg/user-manager/internal/core/domain/users"
    "github.com/padiazg/user-manager/internal/core/services"
    "github.com/padiazg/user-manager/pkg/logger"
    "github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
    Use:   "add <name>",
    Short: "Add a user",
    Long: `Add a user to the database.
You can optionally specify an email address.

Examples:
    user-manager add "John Doe"
    user-manager add "Jane Doe" --email "jane.doe@foo.bar"
`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        if name == "" {
            return fmt.Errorf("add: must provide name")
        }

        cfg := GetConfig()

        // Initialize logger from config
        log := logger.New(&logger.Config{
            Level:  cfg.LogLevel,
            Format: cfg.LogFormat,
        })

        // ── Open database ─────────────────────────────────────────────────
        db, err := userRepository.Open(cfg.DBPath)
        if err != nil {
            return fmt.Errorf("opening database: %w", err)
        }

        // ── Secondary Adapters ────────────────────────────────────────────
        repository := userRepository.NewUserRepository(db)

        // ── Services (core) ───────────────────────────────────────────────
        service := services.New(&services.Config{
            UserRepository: repository,
        })

        // Configure context with cancellation for graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        // Channel to capture OS signals
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

        // Channel for processor errors
        errChan := make(chan error, 1)
        resChan := make(chan *userDomain.User)

        // build request
        email, _ := cmd.Flags().GetString("email")

        req := &userDomain.UserCreateRequest{
            Name:  name,
            Email: email,
        }

        go func() {
            user, err := service.User.CreateUser(ctx, req)
            if err != nil {
                errChan <- fmt.Errorf("add: %w", err)
            }

            resChan <- user
            close(resChan)
        }()

        // Wait for result, signal or error
        select {
        case user := <-resChan:
            bytes, err := json.Marshal(user)
            if err != nil {
                return fmt.Errorf("marshaling user: %w", err)
            }
            fmt.Printf("%s", string(bytes))
        case sig := <-sigChan:
            log.Info("Received signal %v, initiating graceful shutdown...", sig)
            cancel()
        case err := <-errChan:
            cancel()
            return fmt.Errorf("add: %w", err)
        case <-ctx.Done():
            log.Warn("Timeout, forcing exit")
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(addCmd)
    addCmd.Flags().StringP("email", "e", "", "Email")

}
```

```go
// cmd/list.go
package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    userRepository "github.com/padiazg/user-manager/internal/adapters/secondary/database/users"
    userDomain "github.com/padiazg/user-manager/internal/core/domain/users"
    "github.com/padiazg/user-manager/internal/core/services"
    "github.com/padiazg/user-manager/pkg/logger"
    "github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List users",
    Long: `List registered users
Use --id to filter for a single id

Examples:
    user-manager list
    user-manager list --id "40bd1e44-c7a1-4f93-91c0-4449d6f69643"
`,
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg := GetConfig()

        // Initialize logger from config
        log := logger.New(&logger.Config{
            Level:  cfg.LogLevel,
            Format: cfg.LogFormat,
        })

        // ── Open database ─────────────────────────────────────────────────
        db, err := userRepository.Open(cfg.DBPath)
        if err != nil {
            return fmt.Errorf("opening database: %w", err)
        }

        // ── Secondary Adapters ────────────────────────────────────────────
        repository := userRepository.NewUserRepository(db)

        // ── Services (core) ───────────────────────────────────────────────
        service := services.New(&services.Config{
            UserRepository: repository,
        })

        // Configure context with cancellation for graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        // Channel to capture OS signals
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

        // Channel for processor errors
        errChan := make(chan error, 1)
        resChan := make(chan []*userDomain.User)

        id, _ := cmd.Flags().GetString("id")
        limit, _ := cmd.Flags().GetInt("limit")

        go func() {
            var (
                res []*userDomain.User
                err error
            )

            if id == "" {
               res, err = service.User.List(ctx, limit)
            } else {
               limit = 1
               var user *userDomain.User
                user, err = service.User.FindByID(ctx, id)
                if user != nil {
                    res = append(res, user)
                }
            }

            if err != nil {
                errChan <- fmt.Errorf("list: %w", err)
            }

            resChan <- res
            close(resChan)
        }()

        // Wait for result, signal or error
        select {
        case user := <-resChan:
            bytes, err := json.Marshal(user)
            if err != nil {
                return fmt.Errorf("marshaling user: %w", err)
            }
            fmt.Printf("%s", string(bytes))
        case sig := <-sigChan:
            log.Info("Received signal %v, initiating graceful shutdown...", sig)
            cancel()
        case err := <-errChan:
            cancel()
            return fmt.Errorf("list: %w", err)
        case <-ctx.Done():
            log.Warn("Timeout, forcing exit")
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(listCmd)
    listCmd.Flags().StringP("id", "i", "", "user ID")
    listCmd.Flags().IntP("limit", "l", 10, "set results count limit")
}
```

### 3. Update command

```go
// cmd/update.go
package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    userRepository "github.com/padiazg/user-manager/internal/adapters/secondary/database/users"
    "github.com/padiazg/user-manager/internal/core/services"
    "github.com/padiazg/user-manager/pkg/logger"
    "github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
    Use:   "update <id> <new email>",
    Short: "Update user email",
    Long: `Update the email for a user
`,
    Args: cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        id := args[0]
        if id == "" {
            return fmt.Errorf("update: must provide an id")
        }

        email := args[1]
        if email == "" {
            return fmt.Errorf("update: must provide an email")
        }

        cfg := GetConfig()

        // Initialize logger from config
        log := logger.New(&logger.Config{
            Level:  cfg.LogLevel,
            Format: cfg.LogFormat,
        })

        // ── Open database ─────────────────────────────────────────────────
        db, err := userRepository.Open(cfg.DBPath)
        if err != nil {
            return fmt.Errorf("opening database: %w", err)
        }

        // ── Secondary Adapters ────────────────────────────────────────────
        repository := userRepository.NewUserRepository(db)

        // ── Services (core) ───────────────────────────────────────────────
        service := services.New(&services.Config{
            UserRepository: repository,
        })

        // Configure context with cancellation for graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
        defer cancel()

        // Channel to capture OS signals
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

        // Channel for processor errors
        errChan := make(chan error, 1)
        doneChan := make(chan bool)

        go func() {
            err := service.User.UpdateEmail(ctx, id, email)
            if err != nil {
                errChan <- fmt.Errorf("update: %w", err)
            }

            doneChan <- true
            close(doneChan)
        }()

        // Wait for result, signal or error
        select {
        case <-doneChan:
            fmt.Printf("email updated")
        case sig := <-sigChan:
            log.Info("Received signal %v, initiating graceful shutdown...", sig)
            cancel()
        case err := <-errChan:
            cancel()
            return fmt.Errorf("update: %w", err)
        case <-ctx.Done():
            log.Warn("Timeout, forcing exit")
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(updateCmd)
}
```

## Build and Test

```shell
# build the binary
make build

# add an user
$ ./user-manager add "Patricio Diaz" | jq
{
  "id": "8296016b-219a-47a8-819c-2f77f459cbd0",
  "name": "Patricio Diaz",
  "email": "",
  "created_at": "2026-04-13T14:59:19-03:00"
}

# get the user from the db
./user-manager list --id 8296016b-219a-47a8-819c-2f77f459cbd0 | jq
[
  {
    "id": "8296016b-219a-47a8-819c-2f77f459cbd0",
    "name": "Patricio Diaz",
    "email": "",
    "created_at": "2026-04-13T14:59:19-03:00"
  }
]

# add another user
./user-manager add "John Doe" --email "jhon.doe@foo.bar" | jq
{
  "id": "5b53a659-c46a-44c4-8b40-15d5be682184",
  "name": "John Doe",
  "email": "jhon.doe@foo.bar",
  "created_at": "2026-04-13T15:01:30-03:00"
}

# update first user 
./user-manager update 8296016b-219a-47a8-819c-2f77f459cbd0 "padiazg@gmail.com"
email updated

# list all records
./user-manager list | jq
[
  {
    "id": "5b53a659-c46a-44c4-8b40-15d5be682184",
    "name": "John Doe",
    "email": "jhon.doe@foo.bar",
    "created_at": "2026-04-13T15:01:30-03:00"
  },
  {
    "id": "8296016b-219a-47a8-819c-2f77f459cbd0",
    "name": "Patricio Diaz",
    "email": "padiazg@gmail.com",
    "created_at": "2026-04-13T14:59:19-03:00"
  }
]
```

## Summary

This guide demonstrated how to build a CLI application with database persistence using hexagonal architecture:

| Step | Layer | Component | HexaGo Command |
| --- | --- | --- | --- |
| 1 | - | Project initialization | `hexago init` |
| 2 | Domain | Entity (User) | `hexago add domain entity` |
| 3 | Domain + Ports | Port interface (UserRepository) | Manual in `internal/core/domain/` |
| 4 | Secondary | Database adapter (SQLite) | `hexago add adapter secondary database` |
| 5 | Infrastructure | Migrator + migrations | `hexago add migration` |
| 6 | Core | Service implementation | `hexago add service` |
| 7 | Primary | CLI Commands | Manual in `cmd/` |

Key patterns demonstrated:

- **Domain entities**: Business objects with unique identity and validation
- **Port interfaces**: Define contracts that adapters must implement
- **Secondary adapters**: Implement outbound ports (database repositories)
- **Database migrations**: Schema versioning with golang-migrate
- **Service layer**: Orchestrates domain logic, depends only on ports
- **CLI commands**: One-time run commands using context and signal handling for graceful shutdown
- **Dependency rule**: Adapters → Core (never the other way)
