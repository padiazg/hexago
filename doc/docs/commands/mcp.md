# hexago mcp

Start HexaGo as a [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server over stdio.

## Synopsis

```shell
hexago mcp
```

AI assistants connect to this server and use HexaGo's tools to scaffold hexagonal
architecture projects without leaving the conversation.
Each tool delegates to the hexago binary with `--working-directory`, so all generation
logic is identical to the regular CLI — no duplication.

---

## Available Tools

| Tool | What it does |
|------|-------------|
| `hexago_init` | Bootstrap a new project |
| `hexago_add_service` | Add a business-logic service/use case |
| `hexago_add_domain_entity` | Add a domain entity |
| `hexago_add_domain_valueobject` | Add a domain value object |
| `hexago_add_adapter` | Add a primary or secondary adapter |
| `hexago_add_worker` | Add a background worker |
| `hexago_add_migration` | Add a database migration |
| `hexago_add_tool` | Add an infrastructure utility |
| `hexago_validate` | Validate architecture compliance |

All tools require a `working_directory` absolute path parameter:

- **`hexago_init`** — parent directory; project is created as `<working_directory>/<name>/`.
  Pass `in_place: true` to generate directly into `working_directory`.
- **All other tools** — project root (the directory containing `go.mod` and `internal/`).

---

## Client Configuration

=== "Claude Code"

    ```shell
    # Project scope — stored in .mcp.json, commit it so the whole team gets it
    claude mcp add --scope project hexago -- hexago mcp

    # User scope — available across all your projects
    claude mcp add --scope user hexago -- hexago mcp
    ```

    Verify with `claude mcp list`. Scope precedence (highest → lowest): `local > project > user`.

=== "Claude Desktop"

    Edit the config file and **fully restart** the application.

    | Platform | Config file |
    |----------|-------------|
    | macOS    | `~/Library/Application Support/Claude/claude_desktop_config.json` |
    | Windows  | `%APPDATA%\Claude\claude_desktop_config.json` |

    ```json
    {
      "mcpServers": {
        "hexago": {
          "command": "hexago",
          "args": ["mcp"]
        }
      }
    }
    ```

    !!! tip
        Logs: `~/Library/Logs/Claude/mcp.log` (macOS) · `%APPDATA%\Claude\logs\` (Windows).

=== "VS Code"

    VS Code uses a `"servers"` key and **requires** `"type": "stdio"`.

    **Workspace scope** — commit this file to share with your team:

    `.vscode/mcp.json`
    ```json
    {
      "servers": {
        "hexago": {
          "type": "stdio",
          "command": "hexago",
          "args": ["mcp"]
        }
      }
    }
    ```

    **User scope** paths:

    | Platform | Path |
    |----------|------|
    | macOS    | `~/Library/Application Support/Code/User/mcp.json` |
    | Linux    | `~/.config/Code/User/mcp.json` |
    | Windows  | `%APPDATA%\Code\User\mcp.json` |

    Open via Command Palette: **MCP: Open User Configuration**.

=== "Cursor"

    **Project scope** (`.cursor/mcp.json` in the project root):
    ```json
    {
      "mcpServers": {
        "hexago": {
          "command": "hexago",
          "args": ["mcp"]
        }
      }
    }
    ```

    **Global scope** (`~/.cursor/mcp.json`):
    ```json
    {
      "mcpServers": {
        "hexago": {
          "command": "hexago",
          "args": ["mcp"]
        }
      }
    }
    ```

    After editing, restart the MCP server from Cursor's Settings → MCP panel.

=== "Windsurf"

    | Platform | Config file |
    |----------|-------------|
    | macOS / Linux | `~/.codeium/windsurf/mcp_config.json` |
    | Windows       | `%USERPROFILE%\.codeium\windsurf\mcp_config.json` |

    ```json
    {
      "mcpServers": {
        "hexago": {
          "command": "hexago",
          "args": ["mcp"]
        }
      }
    }
    ```

=== "Zed"

    Add to your Zed settings (`Cmd+,` → JSON view):

    | Platform | Config file |
    |----------|-------------|
    | macOS    | `~/.zed/settings.json` |
    | Linux    | `~/.config/zed/settings.json` |

    ```json
    {
      "context_servers": {
        "hexago": {
          "source": "custom",
          "command": "hexago",
          "args": ["mcp"],
          "env": {}
        }
      }
    }
    ```

    `"source": "custom"` is required. Verify in the Agent Panel — a green dot means active.

---

## Quick Reference

| Client | Config file | Key | `type` field |
|--------|-------------|-----|--------------|
| Claude Code | `~/.claude.json` / `.mcp.json` | `mcpServers` | — |
| Claude Desktop | `claude_desktop_config.json` | `mcpServers` | — |
| VS Code | `.vscode/mcp.json` or `…/Code/User/mcp.json` | `servers` | `"stdio"` required |
| Cursor | `.cursor/mcp.json` or `~/.cursor/mcp.json` | `mcpServers` | — |
| Windsurf | `mcp_config.json` | `mcpServers` | — |
| Zed | `settings.json` | `context_servers` | `source: "custom"` |

!!! tip
    If `hexago` is not on `PATH`, use the full absolute binary path in every config
    (e.g. `/home/user/go/bin/hexago`).

---

## Tool Parameters Reference

### `hexago_init`

| Parameter | Required | Type | Default | Description |
|-----------|----------|------|---------|-------------|
| `working_directory` | ✓ | string | — | Parent directory for the new project |
| `name` | ✓ | string | — | Project name |
| `module` | | string | *(name)* | Go module path |
| `project_type` | | string | `http-server` | `http-server` \| `service` |
| `framework` | | string | `stdlib` | `echo` \| `gin` \| `chi` \| `fiber` \| `stdlib` |
| `adapter_style` | | string | `primary-secondary` | `primary-secondary` \| `driver-driven` |
| `core_logic` | | string | `services` | `services` \| `usecases` |
| `in_place` | | bool | `false` | Generate directly into `working_directory` |
| `with_docker` | | bool | `false` | Dockerfile + docker-compose |
| `with_observability` | | bool | `false` | Health checks + Prometheus |
| `with_migrations` | | bool | `false` | Migration setup |
| `with_workers` | | bool | `false` | Worker scaffolding |
| `with_metrics` | | bool | `false` | Prometheus metrics |
| `with_example` | | bool | `false` | Example code |
| `explicit_ports` | | bool | `false` | Explicit `ports/` directory |

### `hexago_add_service`

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `working_directory` | ✓ | string | Project root |
| `name` | ✓ | string | PascalCase name (e.g. `CreateUser`) |
| `description` | | string | One-line comment in generated file |

### `hexago_add_domain_entity` / `hexago_add_domain_valueobject`

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `working_directory` | ✓ | string | Project root |
| `name` | ✓ | string | PascalCase name (e.g. `User`, `Email`) |
| `fields` | | string | Comma-separated `name:type` pairs. E.g. `"id:string,name:string,createdAt:time.Time"` |

### `hexago_add_adapter`

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `working_directory` | ✓ | string | Project root |
| `direction` | ✓ | string | `primary` (inbound) \| `secondary` (outbound) |
| `adapter_type` | ✓ | string | For primary: `http`, `grpc`, `queue`. For secondary: `database`, `external`, `cache` |
| `name` | ✓ | string | PascalCase name (e.g. `UserHandler`, `UserRepository`) |
| `port` | | string | Port interface name to implement (only for projects with `explicit_ports`). E.g. `UserRepository`, `EmailSender` |

### `hexago_add_worker`

| Parameter | Required | Type | Default | Description |
|-----------|----------|------|---------|-------------|
| `working_directory` | ✓ | string | — | Project root |
| `name` | ✓ | string | — | PascalCase name (e.g. `EmailWorker`) |
| `worker_type` | | string | `queue` | `queue` \| `periodic` \| `event` |
| `interval` | | string | `5m` | Duration for periodic workers (e.g. `30s`, `1h`) |
| `workers` | | number | `5` | Goroutine pool size (queue type) |
| `queue_size` | | number | `100` | Channel buffer size (queue type) |

### `hexago_add_migration`

| Parameter | Required | Type | Default | Description |
|-----------|----------|------|---------|-------------|
| `working_directory` | ✓ | string | — | Project root |
| `name` | ✓ | string | — | snake_case name (e.g. `create_users_table`) |
| `migration_type` | | string | `sql` | `sql` \| `go` |

### `hexago_add_tool`

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `working_directory` | ✓ | string | Project root |
| `tool_type` | ✓ | string | `logger` \| `validator` \| `mapper` \| `middleware` |
| `name` | ✓ | string | PascalCase name (e.g. `AuthMiddleware`) |
| `description` | | string | One-line comment in generated file |

### `hexago_validate`

| Parameter | Required | Type | Description |
|-----------|----------|------|-------------|
| `working_directory` | ✓ | string | Project root |

---

## Updating the MCP After a Binary Upgrade

The MCP server is a **long-running process** started once when the AI client launches.
Replacing the binary on disk does not affect the already-running server — you must
restart it so the new binary is loaded.

=== "Claude Code"

    The MCP server process stays alive for the duration of the Claude Code session.
    The simplest way to reload it is to **exit and restart Claude Code**:

    ```shell
    # Exit the current session (Ctrl+D or /exit), then restart
    claude
    ```

    To verify you are on the new version without restarting, run inside a session:

    ```shell
    ! hexago version
    ```

    If that shows the new version but the MCP tools still behave like the old one,
    restart the session.

    **Tip:** `claude mcp list` shows which servers are registered but not which binary
    version they are running. Always restart after upgrading.

=== "Claude Desktop"

    Fully quit and reopen the application — menu bar icon → Quit, **not** just close the window.

=== "VS Code / Cursor / Windsurf"

    Use the editor's MCP panel to **Stop** and **Start** (or **Restart**) the hexago server,
    or reload the window (`Ctrl+Shift+P` → **Developer: Reload Window**).

=== "Zed"

    Restart the Agent Panel or reload Zed.

---

## Prompting Tips

- **Always give an absolute path** for `working_directory` — relative paths may resolve
  differently depending on how the AI client launches the server.
- **Skip "use hexago"** in your prompt — just describe what you want; the AI will pick
  the right MCP tool automatically.
- **Call `hexago_validate`** after every component is added to catch architecture
  violations early.

**Example prompts:**

```
Add a domain entity named User with fields id:string, name:string, email:string,
createdAt:time.Time to the project at /home/user/projects/my-api
```

```
Initialize a new hexagonal architecture project with Echo framework at /home/user/projects,
named blog-api, with Docker and observability support
```
