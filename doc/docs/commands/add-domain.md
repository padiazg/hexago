# hexago add domain

Add domain entities or value objects to an existing project.

## Synopsis

```shell
hexago add domain entity <name> [flags]
hexago add domain valueobject <name> [flags]
```

Must be run from the project root directory.

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
| `--fields` | | string | `""` | Comma-separated `field:type` pairs |

**Field Types:**

Any valid Go type: `string`, `int`, `int64`, `float64`, `bool`, `time.Time`, etc.

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
hexago add domain valueobject <name> [--fields "field:type"]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--fields` | | string | `""` | Comma-separated `field:type` pairs |

**Examples:**

```shell
hexago add domain valueobject Email
hexago add domain valueobject Money --fields "amount:float64,currency:string"
hexago add domain valueobject Address --fields "street:string,city:string,country:string"
```

---

## Generated Files

For `hexago add domain entity User --fields "id:string,name:string,email:string"`:

```
internal/core/domain/
├── user.go           # Entity definition
└── user_test.go      # Test file
```

---

## Generated Code Structure

**Entity:**

```go
package domain

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

**Value Object:**

```go
package domain

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
