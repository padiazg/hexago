# Build and Run

Commands for building, testing, and running your HexaGo-generated application.

---

## Make Targets

The generated `Makefile` provides convenient targets for common operations:

| Target | Description |
| --- | --- |
| `make build` | Compile the application binary |
| `make clean` | Remove build artifacts |
| `make test` | Run unit tests with race detection |
| `make test-integration` | Run integration tests (requires API keys) |
| `make test-coverage` | Generate HTML coverage report |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code with `go fmt` |
| `make mod-tidy` | Tidy go modules |
| `make validate` | Validate hexagonal architecture |
| `make run` | Run the application |
| `make docker-up` | Start Docker Compose services |
| `make docker-down` | Stop Docker Compose services |

---

## Go Commands

### Build

```bash
# Build all packages
go build ./...

# Build specific binary
go build -o my-app main.go
```

### Test

```bash
# Run all tests
go test ./...

# Run specific test
go test ./... -run TestFoo

# Run tests in specific package
go test ./internal/core/domain/...

# Run with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
```

### Run Modes

Some generated projects include multiple run modes:

```bash
# Standard server mode
go run main.go run

# Simulation mode (paper trading with live data)
go run main.go simulate --from 2024-01-01 --wfo

# Replay mode (local CSV, no API key needed)
go run main.go replay --csv testdata/sample.csv
```

### Lint

```bash
# Run linter
golangci-lint run ./...

# Fix auto-fixable issues
golangci-lint run ./... --fix
```

---

## Docker Commands

```bash
# Build Docker image
docker build -t my-app:latest .

# Start services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

---

## Architecture Validation

After any code changes, validate the hexagonal architecture constraints:

```bash
hexago validate
```

This checks:

- Core domain has no external dependencies
- Services only depend on domain and ports
- Adapters don't import from other adapters
- Proper dependency direction (adapters → core)

---

## Troubleshooting

### Build fails

```bash
# Ensure dependencies are downloaded
go mod download

# Tidy modules
go mod tidy
```

### Tests fail

```bash
# Run with verbose output
go test -v ./...

# Run only unit tests (no integration)
go test ./...

# Integration tests require API keys in .env
```

### Linter errors

```bash
# See what the linter expects
golangci-lint run ./... 2>&1 | head -50
```
