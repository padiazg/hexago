# hexago validate

Validate hexagonal architecture compliance for the current project.

## Synopsis

```shell
hexago validate
```

Must be run from the project root directory.

---

## Description

`hexago validate` analyzes your project's import graph to ensure the **Dependency Rule** is respected: dependencies must flow inward (adapters → services → domain), never outward.

---

## Checks Performed

| Check | Description |
|-------|-------------|
| **Project structure** | Verifies required directories exist |
| **Core domain dependencies** | Domain packages must not import adapters or infrastructure |
| **Service/UseCase dependencies** | Services must not import adapter packages |
| **Adapter dependencies** | Adapters must not import other adapter packages |
| **Naming conventions** | Validates package and file naming |

---

## Example Output

**Valid project:**

```
Validating hexagonal architecture...

  ✓ Project structure
  ✓ Core domain dependencies
  ✓ Service/UseCase dependencies
  ✓ Adapter dependencies
  ✓ Naming conventions

Architecture is valid!
```

**Invalid project (architecture violation):**

```
Validating hexagonal architecture...

  ✓ Project structure
  ✗ Core domain dependencies
    internal/core/domain/user.go imports "internal/adapters/secondary/database"
    Domain layer must not depend on adapters

  ✓ Service/UseCase dependencies
  ✓ Adapter dependencies
  ✓ Naming conventions

Found 1 architecture violation(s).
```

---

## Usage in CI/CD

Add architecture validation to your pipeline to catch violations early:

```yaml
# GitHub Actions example
- name: Validate architecture
  run: hexago validate
```

Or add to your Makefile:

```makefile
lint: fmt
    go vet ./...
    hexago validate
```

---

## When to Run

- After adding new components
- Before committing code
- In CI/CD pipelines
- During code review

---

## Common Violations

### Domain importing adapter

```
internal/core/domain/user.go imports "internal/adapters/secondary/database"
```

**Fix:** Move the database import to a service or repository. Domain objects should only know about other domain objects.

### Service importing adapter directly

```
internal/core/services/create_user.go imports "internal/adapters/secondary/database"
```

**Fix:** Define a port interface in the service, and inject the repository through the interface.

### Adapter importing another adapter

```
internal/adapters/primary/http/user_handler.go imports "internal/adapters/secondary/database"
```

**Fix:** HTTP handlers should call services, not repositories directly. Services own the database interaction through port interfaces.
