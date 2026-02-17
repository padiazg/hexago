# HexaGo Quick Start Guide

## Installation

```bash
cd /home/pato/go/src/github.com/padiazg/hexago
go install
```

Or build locally:
```bash
go build -o hexago main.go
```

## Generate Your First Project

### Basic Project
```bash
hexago init my-app --module github.com/yourname/my-app
cd my-app
go run main.go run
```

Visit http://localhost:8080/health to see it working!

### With Echo Framework
```bash
hexago init api-server \
  --module github.com/yourname/api-server \
  --framework echo

cd api-server
make run
```

### With Alternative Naming (DDD Style)
```bash
hexago init service \
  --module github.com/company/service \
  --adapter-style driver-driven \
  --core-logic usecases

cd service
go run main.go run
```

## Project Structure

Generated projects follow hexagonal architecture:

```
my-app/
├── cmd/                    # Cobra commands
│   ├── root.go            # Root + config
│   └── run.go             # Server logic
├── internal/
│   ├── core/              # Business logic (NO external deps)
│   │   ├── domain/        # Entities
│   │   └── services/      # Use cases
│   ├── adapters/          # External interfaces
│   │   ├── primary/       # Inbound (HTTP, gRPC)
│   │   └── secondary/     # Outbound (DB, APIs)
│   └── config/            # Configuration
├── pkg/logger/            # Reusable packages
├── main.go                # Entry point
└── Makefile               # Common tasks
```

## Common Commands

### Development
```bash
make run          # Run the application
make build        # Build binary
make test         # Run tests
make fmt          # Format code
```

### Docker
```bash
make build-image  # Build Docker image
make docker-up    # Start with Docker Compose
make docker-down  # Stop services
```

## Configuration

Create `.my-app.yaml`:
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
```bash
export MY_APP_SERVER_PORT=8080
export MY_APP_LOGLEVEL=debug
```

## Available Flags

```bash
hexago init <name> [flags]

Flags:
  -m, --module string          Go module name
  -f, --framework string       echo|gin|chi|fiber|stdlib (default "stdlib")
      --adapter-style string   primary-secondary|driver-driven (default "primary-secondary")
      --core-logic string      services|usecases (default "services")
      --with-docker            Generate Docker files (default false)
      --with-observability     Include health + metrics (default false)
      --with-migrations        Include migration setup (default false)
      --with-workers           Include worker pattern (default false)
      --with-metrics           Include Prometheus metrics (default false)
      --with-example           Include example code (default false)
      --explicit-ports         Create ports/ directory (default false)
```

## Examples

### Simple REST API
```bash
hexago init todo-api --module github.com/me/todo-api --framework gin
cd todo-api
# Add your endpoints in internal/adapters/primary/http/
go run main.go run
```

### Microservice with Workers
```bash
hexago init email-service \
  --module github.com/company/email-service \
  --with-workers \
  --with-migrations

cd email-service
# Add your business logic in internal/core/services/
make run
```

### Domain-Driven Design Style
```bash
hexago init ordering \
  --module github.com/shop/ordering \
  --adapter-style driver-driven \
  --core-logic usecases \
  --explicit-ports

cd ordering
# Your ports will be in internal/core/ports/
go run main.go run
```

## Next Steps

1. Read the generated `README.md` in your project
2. Add domain entities in `internal/core/domain/`
3. Add business logic in `internal/core/services/` (or `usecases/`)
4. Add HTTP handlers in `internal/adapters/primary/http/`
5. Add repositories in `internal/adapters/secondary/database/`

## Architecture Guidelines

### Core Layer Rules
- ✅ Define business logic
- ✅ Define port interfaces
- ✅ Be framework-agnostic
- ❌ Never import adapters
- ❌ Never import infrastructure

### Adapters Layer Rules
- ✅ Implement port interfaces
- ✅ Handle external communication
- ✅ Use frameworks here
- ❌ Never import from other adapters
- ✅ Depend only on core interfaces

## Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/core/services
```

## Graceful Shutdown

All generated projects include graceful shutdown:
- Press Ctrl+C to stop
- Server waits for active requests (30s timeout)
- Clean resource cleanup

## Getting Help

```bash
hexago --help              # General help
hexago init --help         # Init command help
```

## Troubleshooting

### "module not found" error
```bash
cd your-project
go mod tidy
```

### Port already in use
Change port in `.my-app.yaml` or:
```bash
export MY_APP_SERVER_PORT=9000
./my-app run
```

### Build errors
```bash
go mod tidy
go fmt ./...
go build
```

## Learn More

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Ports & Adapters](https://herbertograca.com/2017/11/16/explicit-architecture-01-ddd-hexagonal-onion-clean-cqrs-how-i-put-it-all-together/)

Happy coding! 🚀
