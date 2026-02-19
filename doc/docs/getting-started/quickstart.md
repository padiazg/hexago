# Quick Start

This guide walks you through creating your first HexaGo project and adding components to it.

---

## 1. Create a New Project

=== "Basic (stdlib)"

    ```shell
    hexago init my-app --module github.com/yourname/my-app
    cd my-app
    go run main.go run
    ```

=== "With Echo Framework"

    ```shell
    hexago init api-server \
      --module github.com/yourname/api-server \
      --framework echo

    cd api-server
    make run
    ```

=== "DDD Style (driver-driven)"

    ```shell
    hexago init service \
      --module github.com/company/service \
      --adapter-style driver-driven \
      --core-logic usecases

    cd service
    go run main.go run
    ```

Visit [http://localhost:8080/health](http://localhost:8080/health) — you should see the health response immediately.

---

## 2. Explore the Generated Project

```
my-app/
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command + config
│   └── run.go             # Server with graceful shutdown
├── internal/
│   ├── core/              # CORE — No external dependencies
│   │   ├── domain/        # Domain entities
│   │   └── services/      # Business logic
│   ├── adapters/          # ADAPTERS — External interfaces
│   │   ├── primary/       # Inbound (HTTP, gRPC)
│   │   └── secondary/     # Outbound (DB, APIs)
│   └── config/            # Configuration
├── pkg/
│   └── logger/            # Reusable logger package
├── main.go                # Minimal entry point
├── Makefile               # Common tasks
├── Dockerfile             # Multi-stage build
└── compose.yaml           # Docker Compose
```

---

## 3. Add Domain Entities

```shell
cd my-app

hexago add domain entity User --fields "id:string,name:string,email:string"
hexago add domain entity Product --fields "id:string,name:string,price:float64"
hexago add domain valueobject Email
```

---

## 4. Add Business Logic

```shell
hexago add service CreateUser --description "Creates a new user"
hexago add service GetUser
hexago add service ListUsers
```

---

## 5. Add Adapters

```shell
# Database repositories (secondary/outbound)
hexago add adapter secondary database UserRepository
hexago add adapter secondary database ProductRepository

# HTTP handlers (primary/inbound)
hexago add adapter primary http UserHandler
hexago add adapter primary http ProductHandler
```

---

## 6. Implement Your Logic

Open the generated files — they contain `// TODO` comments to guide you:

```shell
# Business logic
vim internal/core/services/create_user.go

# HTTP handler
vim internal/adapters/primary/http/user_handler.go

# Repository
vim internal/adapters/secondary/database/user_repository.go
```

---

## 7. Run and Test

```shell
# Build and run
make run

# Or directly
go run main.go run

# Run tests
make test

# With coverage
make test-coverage
```

---

## 8. Validate Architecture

```shell
hexago validate
```

This checks that your code follows hexagonal architecture rules — no illegal dependencies between layers.

---

## Configuration

Create `.my-app.yaml` in the project root:

```yaml
server:
  port: 8080
  readtimeout: 15s
  writetimeout: 15s
  shutdowntimeout: 30s

loglevel: info
logformat: json
```

Or use environment variables:

```shell
export MY_APP_SERVER_PORT=8080
export MY_APP_LOGLEVEL=debug
```

---

## Available Make Targets

```shell
make build           # Build the application
make run             # Run the application
make test            # Run tests
make test-coverage   # Run tests with coverage
make clean           # Clean build artifacts
make fmt             # Format code
make lint            # Run linter
make docker-build    # Build Docker image
make docker-up       # Start Docker Compose
make docker-down     # Stop Docker Compose
```

---

## Full Example: Blog API

Here's a complete example building a blog API with Gin:

```shell
# 1. Create project
hexago init blog-api --module github.com/me/blog-api --framework gin

cd blog-api

# 2. Add domain
hexago add domain entity Post --fields "id:string,title:string,content:string,authorID:string"
hexago add domain entity Author --fields "id:string,name:string,email:string"
hexago add domain valueobject Email

# 3. Add business logic
hexago add service CreatePost
hexago add service GetPost
hexago add service ListPosts
hexago add service CreateAuthor

# 4. Add repositories
hexago add adapter secondary database PostRepository
hexago add adapter secondary database AuthorRepository

# 5. Add HTTP handlers
hexago add adapter primary http PostHandler
hexago add adapter primary http AuthorHandler

# 6. Add workers
hexago add worker EmailWorker --type queue
hexago add worker CacheWarmer --type periodic --interval 10m

# 7. Add migrations
hexago add migration create_posts_table
hexago add migration create_authors_table

# 8. Add infrastructure tools
hexago add tool validator PostValidator
hexago add tool middleware RateLimitMiddleware

# 9. Validate architecture
hexago validate

# 10. Build and run
make run
```

---

## Troubleshooting

### "not a hexagonal architecture project"

Run commands from the project root directory where `go.mod` exists.

### "module not found" error

```shell
cd your-project
go mod tidy
```

### Port already in use

```shell
export MY_APP_SERVER_PORT=9000
./my-app run
```

### Build errors

```shell
go mod tidy
go fmt ./...
go build
```

---

## Next Steps

- Browse the [Commands Reference](../commands/index.md) for all available flags
- Read about [Architecture Principles](../architecture/overview.md)
- Learn to [customize templates](../customization/templates.md) to match your team's style
