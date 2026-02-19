# HexaGo

**Generate production-ready Go applications with Hexagonal Architecture**

HexaGo is an opinionated CLI tool to scaffold Go applications following the **Hexagonal Architecture** (Ports & Adapters) pattern. It helps developers maintain proper separation of concerns and build maintainable, testable applications from day one.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/padiazg/hexago/blob/main/LICENSE)

---

## Features

<div class="grid cards" markdown>

-   :rocket: **Project Generation**

    ---

    One command creates a complete project with your chosen web framework, Docker support, graceful shutdown, configuration management, and observability built in.

    [:octicons-arrow-right-24: Quick Start](getting-started/quickstart.md)

-   :puzzle_piece: **Component Generation**

    ---

    Add services, domain entities, value objects, and adapters to existing projects. Auto-detects your project conventions and generates consistent, context-aware code.

    [:octicons-arrow-right-24: Commands Reference](commands/index.md)

-   :zap: **High Value Features**

    ---

    Generate background workers (queue, periodic, event-driven), database migrations with sequential numbering, and validate your architecture compliance.

    [:octicons-arrow-right-24: Architecture Guide](architecture/overview.md)

-   :art: **Template Customization**

    ---

    Modify generated code to match your team's style. Add company headers, enforce coding standards, and share templates across your organization via version control.

    [:octicons-arrow-right-24: Template Customization](customization/templates.md)

</div>

---

## Quick Install

=== "Go"

    ```shell
    go install github.com/padiazg/hexago@latest
    ```

=== "Homebrew"

    ```shell
    brew tap padiazg/hexago
    brew install hexago
    ```

=== "Build from source"

    ```shell
    git clone https://github.com/padiazg/hexago.git
    cd hexago
    go build -o hexago
    ```

---

## Get Started in 60 Seconds

```shell
# Create a new project with Echo framework
hexago init my-api --module github.com/user/my-api --framework echo

cd my-api

# Add domain entities
hexago add domain entity User --fields "id:string,name:string,email:string"

# Add business logic
hexago add service CreateUser

# Add HTTP handler
hexago add adapter primary http UserHandler

# Validate architecture
hexago validate

# Build and run
make run
```

Visit [http://localhost:8080/health](http://localhost:8080/health) to see it working!

[Get Started :octicons-arrow-right-24:](getting-started/installation.md){ .md-button .md-button--primary }
[View Commands :octicons-arrow-right-24:](commands/index.md){ .md-button }

---

## Framework Support

HexaGo generates framework-specific code for:

| Framework | Handler Signature |
|-----------|-------------------|
| **stdlib** | `http.Handler` |
| **Echo** | `func(echo.Context) error` |
| **Gin** | `func(*gin.Context)` |
| **Chi** | Standard library with chi router |
| **Fiber** | `func(*fiber.Ctx) error` |

---

## Why Hexagonal Architecture?

Hexagonal Architecture separates your application into three distinct layers:

```
Adapters (HTTP, DB, Queue) → Services/UseCases → Domain
```

- **Domain** — Pure business entities, zero external dependencies
- **Services** — Application logic, orchestrates domain objects, defines port interfaces
- **Adapters** — Implementations of ports (HTTP handlers, repositories, external services)

This structure makes your application **testable**, **maintainable**, and **framework-agnostic**.

[Learn more about the architecture :octicons-arrow-right-24:](architecture/overview.md)
