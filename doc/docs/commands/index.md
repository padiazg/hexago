# Commands Reference

HexaGo provides a set of commands to scaffold and manage hexagonal architecture projects.

---

## Command Overview

| Command | Description |
|---------|-------------|
| [`hexago init`](init.md) | Create a new hexagonal architecture project |
| [`hexago add service`](add-service.md) | Add a business logic service/use case |
| [`hexago add domain`](add-domain.md) | Add a domain entity or value object |
| [`hexago add adapter`](add-adapter.md) | Add a primary or secondary adapter |
| [`hexago add worker`](add-worker.md) | Add a background worker |
| [`hexago add migration`](add-migration.md) | Add a database migration |
| [`hexago add tool`](add-tool.md) | Add an infrastructure tool |
| [`hexago validate`](validate.md) | Validate architecture compliance |
| [`hexago mcp`](mcp.md) | Start the built-in MCP server for AI assistants |
| [`hexago version`](version.md) | Print version and build information |
| [`hexago templates`](../customization/templates.md) | Manage and customize code generation templates |

---

## Global Help

```shell
hexago --help              # General help
hexago init --help         # Help for a specific command
hexago add --help          # Help for add subcommands
```

---

## Where to Run Commands

All `hexago add` and `hexago validate` commands operate on the **project root** — the directory that contains `go.mod` and `internal/`.

You can point to it in two ways:

=== "Navigate first (classic)"

    ```shell
    cd my-project
    hexago add service CreateUser
    ```

=== "--working-directory flag (no cd)"

    ```shell
    hexago add service CreateUser --working-directory /home/user/projects/my-project
    ```

The `-w` short form also works:

```shell
hexago validate -w /home/user/projects/my-project
```

!!! warning
    Running without a valid hexagonal architecture project root produces:
    ```
    Error: not a hexagonal architecture project
    ```
