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
| `hexago templates` | Manage and customize code generation templates |

---

## Global Help

```shell
hexago --help              # General help
hexago init --help         # Help for a specific command
hexago add --help          # Help for add subcommands
```

---

## Where to Run Commands

All `hexago add` and `hexago validate` commands must be run from the **project root directory** — the directory containing `go.mod`.

```shell
cd my-project       # navigate to project root
hexago add service CreateUser
```

!!! warning
    Running `hexago add` outside a valid hexagonal architecture project root will produce an error:
    ```
    Error: not a hexagonal architecture project
    ```
