# AGENTS.md - HexaGo Development Guide

## Quick Commands

```bash
# Build
go build -o hexago main.go
make build  # includes version info from git tags

# Run directly
go run main.go [command]

# Test
go test ./...

# Lint (if installed)
golangci-lint run ./...
```

## Key Architectural Facts

- **HexaGo is itself a code generator** - it generates hexagonal architecture Go projects
- **CLI framework**: Cobra used as library, not CLI tool
- **Entry point**: `main.go` only calls `cmd.Execute()` - real logic lives in `cmd/*.go`
- **Command registration**: Each subcommand registers itself via `init()` with `rootCmd.AddCommand()`
- **Templates**: Embedded via `go:embed` in `internal/templates/`

## Generated Project Patterns (For Context)

When HexaGo generates a project, it creates:
- `main.go` → minimal, calls `cmd.Execute()`
- `cmd/root.go` → config + `Execute()` function
- `cmd/run.go` → server logic with graceful shutdown
- `internal/core/` → domain + services (no external deps)
- `internal/adapters/` → primary/secondary adapters

## MCP Integration

HexaGo includes a built-in MCP server (`hexago mcp`) - used by Claude Code, Cursor, etc. The MCP tools are:
- `hexago_init` - initialize projects
- `hexago_add_service` - add business logic
- `hexago_add_domain_entity` - add domain entities
- `hexago_add_adapter` - add HTTP handlers, repositories
- `hexago_add_worker` - add background workers
- `hexago_validate` - verify architecture compliance

## Important Files

- `.hexago.yaml` - project config (framework, adapter style, features)
- `go.mod` - current Go version is 1.25.5
- `CLAUDE.md` - detailed guidance for Claude Code (read this for implementation details)

## Common Gotchas

- Run `hexago validate` after any `add` command to check architecture compliance
- MCP tools require `--working_directory` parameter (project root)
- Generated projects use Viper config with YAML + env vars (prefix: `MYAPP_`)