# hexago add domain

Add domain entities or value objects to an existing project.

## Synopsis

```shell
hexago add domain entity <name> [flags]
hexago add domain valueobject <name> [flags]
```

Operates on the project root — use `--working-directory` (`-w`) to target a project without changing directories.

---

## Subcommands

### entity

Add a domain entity — a mutable business object with identity.

```shell
hexago add domain entity <name> [--fields "field:type,field:type"]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--fields` | `-f` | string | `""` | Comma-separated `field:type` pairs |

**Field Types:**

Any valid Go type: `string`, `int`, `int64`, `float64`, `bool`, `time.Time`, etc.

**Reserved keyword handling:**

Field names that lower-case to a Go reserved keyword (`type`, `map`, `range`, `func`, `chan`, etc.) are automatically renamed in the generated constructor by appending `Val`. The struct field itself is unaffected.

```shell
# "type" is a reserved keyword — hexago handles it automatically
hexago add domain entity Movement \
  --fields "productEan13:string,type:string,quantity:float64"
```

```go
// struct field stays "Type"; constructor parameter becomes "typeVal"
func NewMovement(productEan13 string, typeVal string, quantity float64) (*Movement, error) {
    return &Movement{
        ProductEan13: productEan13,
        Type:         typeVal,
        Quantity:     quantity,
    }, nil
}
```

**Examples:**

```shell
hexago add domain entity User --fields "id:string,name:string,email:string"
hexago add domain entity Order --fields "id:string,total:float64,createdAt:time.Time"
hexago add domain entity Product --fields "id:string,name:string,price:float64,stock:int"
```

---

### valueobject

Add a value object — an immutable domain concept defined by its attributes, not identity.

```shell
hexago add domain valueobject <name> [--fields "field:type"] [--entity <entity>]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--fields` | `-f` | string | `""` | Comma-separated `field:type` pairs |
| `--entity` | `-e` | string | `""` | Entity name to co-locate with (entity-bound); omit for standalone sub-package |

**Examples:**

```shell
hexago add domain valueobject Email
hexago add domain valueobject Money --fields "amount:float64,currency:string"
hexago add domain valueobject Address --fields "street:string,city:string,country:string"
hexago add domain valueobject StockLevel --fields "value:float64" --entity Product
```

---

## Generated Files

For `hexago add domain entity User --fields "id:string,name:string,email:string"`:

```
internal/core/domain/users/
├── users.go          # Entity definition
├── port.go           # Repository port interface
└── users_test.go     # Test file
```

Value objects are generated either as a standalone sub-package or co-located with an entity:

```shell
# Standalone — own sub-package
hexago add domain valueobject Email
# → internal/core/domain/email/email.go

# Entity-bound — co-located with entity
hexago add domain valueobject StockLevel --entity Product
# → internal/core/domain/products/stock_level.go
```

---

## Generated Code Structure

**Entity (in `internal/core/domain/users/users.go`):**

```go
package users

// User represents the User entity
type User struct {
    ID    string
    Name  string
    Email string
}

// NewUser creates a new User entity
func NewUser(id string, name string, email string) *User {
    return &User{
        ID:    id,
        Name:  name,
        Email: email,
    }
}
```

**Value Object (standalone, in `internal/core/domain/email/email.go`):**

```go
package email

// Email represents the Email value object
type Email struct {
    // TODO: Add fields
}

// NewEmail creates a new Email value object
func NewEmail() (*Email, error) {
    // TODO: Add validation
    return &Email{}, nil
}
```

---

## Architecture Notes

Domain objects belong to the **core layer** and must:

- ✅ Contain pure business logic and validation
- ✅ Be self-contained with zero external dependencies
- ❌ Never import adapter packages
- ❌ Never import infrastructure packages
- ❌ Never import external libraries (no database drivers, HTTP clients, etc.)
