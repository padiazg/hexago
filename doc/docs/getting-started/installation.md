# Installation

HexaGo requires **Go 1.21 or later**. Choose your preferred installation method below.

---

## Using Go Install

The simplest method — installs the latest release directly from the module proxy:

```shell
go install github.com/padiazg/hexago@latest
```

To install a specific version:

```shell
go install github.com/padiazg/hexago@v0.0.1
```

!!! tip
    Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`:
    ```shell
    export PATH="$PATH:$(go env GOPATH)/bin"
    ```

---

## Using Homebrew

```shell
brew tap padiazg/hexago
brew install hexago
```

---

## Build from Source

```shell
git clone https://github.com/padiazg/hexago.git
cd hexago
go build -o hexago
```

Then move the binary somewhere on your `PATH`:

```shell
mv hexago /usr/local/bin/hexago
# or
mv hexago ~/go/bin/hexago
```

---

## Verify Installation

```shell
hexago --help
```

You should see output like:

```
HexaGo - Hexagonal Architecture Scaffolding CLI

Usage:
  hexago [command]

Available Commands:
  init        Initialize a new hexagonal architecture project
  add         Add components to an existing project
  validate    Validate hexagonal architecture compliance
  templates   Manage code generation templates

Flags:
  -h, --help   help for hexago
```

---

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | Required to install and build |
| Git | Any | For cloning and version control |

!!! note
    HexaGo generates **static binaries** (`CGO_ENABLED=0`), so generated projects can be deployed without external dependencies.

---

## Platform Support

Pre-built binaries are available for:

- Linux x86_64
- Linux arm64
- macOS x86_64 (Intel)
- macOS arm64 (Apple Silicon)

Download from [GitHub Releases](https://github.com/padiazg/hexago/releases).

---

## Next Steps

Once installed, head to the [Quick Start guide](quickstart.md) to create your first project.
