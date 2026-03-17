# Plan: Align generated code with preferred style

Reference project: `/Users/pato/go/src/github.com/padiazg/tmp/hexago-api-chi`

This document describes the changes needed in the hexago generator so that the
skeleton it produces matches the modular/decoupled style used in that reference
project.  Each task is intentionally small so it can be executed in a single
context window without saturating it.

---

## Quick summary of differences

| Area | Currently generated | Desired style |
|---|---|---|
| Domain entity | `internal/core/domain/category.go`, `package domain` | `internal/core/domain/categories/categories.go`, `package categories` |
| Domain port | Not generated | `internal/core/domain/categories/port.go` with repository interface |
| Value object (standalone) | `internal/core/domain/ean13.go`, `package domain` | `internal/core/domain/ean13/ean13.go`, `package ean13` (own sub-package) |
| Value object (entity-bound) | same flat file | `internal/core/domain/products/stock_level.go`, `package products` (co-located with entity) |
| Service | `internal/core/services/manage_category.go`, `package services`, single `Execute()` | `internal/core/services/categories/categories.go`, `package categories`, named methods (Create/Update/GetByID/List) |
| Services aggregator | Not generated | `internal/core/services/services.go` with `Config` + `Services` structs |
| DB repository | `internal/adapters/secondary/database/category_repository.go`, `package database` | `internal/adapters/secondary/database/categories/categories.go`, `package categories` |
| HTTP handler | Single file, `package http`, generic HandleGet/HandlePost | Two files per entity in own subpackage: `category.go` (Config/DTOs) + `handlers.go` (methods) |
| HTTP wiring | `func New(cfg) server.Server` — no services param | `func New(cfg, services) server.Server` — receives `*services.Services` |
| cmd/run.go | No database wiring | Full DB open → repos → services → http.New(cfg, services) chain |

---

## Task 1 — Domain entity: sub-package structure

**Files to change:**
- `internal/generator/domain.go` — `GenerateEntity()`
- `internal/generator/templates/domain/entity.go.tmpl`
- New template: `internal/generator/templates/domain/entity_test.go.tmpl` (update package reference)

**What to do:**

1. In `domain.go` → `GenerateEntity()`:
   - Compute `pkgName = strings.ToLower(entityName) + "s"` (e.g. `Category` → `categories`).
     Add a helper `utils.ToPlural(s string) string` in `pkg/utils` or inline.
   - Change `domainDir` to `filepath.Join("internal", "core", "domain", pkgName)`.
   - Create the subdirectory with `utils.CreateDir(domainDir)`.
   - Pass `PackageName: pkgName` in the template data map.

2. In `entity.go.tmpl` line 1: change `package domain` → `package {{.PackageName}}`.

3. In `entity_test.go.tmpl`: update import path to
   `"{{.ModuleName}}/internal/core/domain/{{.PackageName}}"` and change
   `package domain` → `package {{.PackageName}}_test`.

**Expected output for `hexago_add_domain_entity name: Category fields: id:string,...`:**
```
internal/core/domain/categories/
  categories.go   → package categories
  categories_test.go
```

---

## Task 2 — Domain entity: generate port.go

**Files to change:**
- `internal/generator/domain.go` — add `generatePortFile()` call inside `GenerateEntity()`
- New template: `internal/generator/templates/domain/port.go.tmpl`

**What to do:**

1. Add new template `domain/port.go.tmpl`:

```go
package {{.PackageName}}

import "context"

// {{.EntityName}}Repository defines the secondary port for {{.EntityName}} persistence.
type {{.EntityName}}Repository interface {
	Create(ctx context.Context, entity *{{.EntityName}}) error
	FindByID(ctx context.Context, id string) (*{{.EntityName}}, error)
	Update(ctx context.Context, entity *{{.EntityName}}) error
	List(ctx context.Context) ([]*{{.EntityName}}, error)
}
```

2. In `domain.go` → `GenerateEntity()`, after writing the entity file, call:
   ```go
   g.generatePortFile(filepath.Join(domainDir, "port.go"), entityName, pkgName)
   ```

3. Implement `generatePortFile()` passing `EntityName` and `PackageName` to the template.

**Expected output:**
```
internal/core/domain/categories/
  categories.go
  port.go         → CategoryRepository interface
  categories_test.go
```

---

## Task 3 — Value object: co-located or own sub-package

**Files to change:**
- `internal/generator/domain.go` — `GenerateValueObject()`
- `internal/generator/templates/domain/value_object.go.tmpl`
- `internal/generator/templates/domain/value_object_test.go.tmpl`
- CLI command (`cmd/add.go` or equivalent) — add `--entity` flag to the value
  object sub-command

**Background:**

Value objects are NOT all placed in their own sub-package.  The reference
project uses two distinct placements:

| Value object | Placement | Rule |
|---|---|---|
| `StockLevel` | `domain/products/stock_level.go`, `package products` | **Entity-bound**: only meaningful inside one entity's aggregate |
| `MovementType` | `domain/movements/movemente_type.go`, `package movements` | **Entity-bound**: same |
| `EAN13` | `domain/ean13/ean13.go`, `package ean13` | **Standalone**: cross-cutting concept, reusable across aggregates |

Value objects do NOT get a `port.go`.

**What to do:**

1. Add an optional `--entity <EntityName>` flag to `hexago_add_domain_valueobject`
   (and to the `hexago_add_domain_valueobject` MCP tool parameter list).

2. In `domain.go` → `GenerateValueObject()`, branch on whether `entityName` was
   supplied:

   **With `--entity Category` (entity-bound):**
   - `pkgName = utils.ToPlural(strings.ToLower(entityName))` — e.g. `categories`
   - `voDir = filepath.Join("internal", "core", "domain", pkgName)`
   - The directory must already exist (entity must have been created first);
     return an error if it doesn't.
   - File: `voDir/<snake_case_voName>.go`, same `package <pkgName>` as the entity.

   **Without `--entity` (standalone):**
   - `pkgName = strings.ToLower(voName)` — e.g. `ean13`
   - `voDir = filepath.Join("internal", "core", "domain", pkgName)`
   - Create the subdirectory.
   - File: `voDir/<snake_case_voName>.go`, `package <pkgName>`.

3. Pass `PackageName: pkgName` in the template data map (same variable either way).

4. In `value_object.go.tmpl` line 1: change `package domain` → `package {{.PackageName}}`.

5. In `value_object_test.go.tmpl`: update import and change
   `package domain` → `package {{.PackageName}}_test`.

**Expected output examples:**

```
# hexago_add_domain_valueobject name: StockLevel fields: value:float64 --entity Product
internal/core/domain/products/
  product.go          (existing)
  port.go             (existing)
  stock_level.go  ←   new, package products

# hexago_add_domain_valueobject name: EAN13 fields: value:string
internal/core/domain/ean13/
  ean13.go        ←   new, package ean13
```

---

## Task 4 — Service: sub-package + named methods

**Files to change:**
- `internal/generator/service.go` — `Generate()`
- `internal/generator/templates/service/service.go.tmpl`
- `internal/generator/templates/service/service_test.go.tmpl`

**What to do:**

1. In `service.go` → `Generate()`:
   - Derive `pkgName = strings.ToLower(serviceName) + "s"` — wait, actually use the
     domain entity the service belongs to.  Since the CLI currently only takes a
     service name (e.g. `ManageCategory`), derive the pkg name from a new optional
     `--entity` flag (e.g. `--entity category` → subdir `categories`), or use a
     simple heuristic: strip known prefixes (`Manage`, `Get`, `Record`) and
     pluralize the rest.
   - Alternatively (simpler): add a `--package` flag that defaults to
     `strings.ToLower(serviceName)`.
   - Change `serviceDir` to `filepath.Join("internal", "core", g.config.CoreLogicDir(), pkgName)`.
   - Create subdirectory.

2. Rewrite `service.go.tmpl` to match the named-methods pattern:

```go
package {{.PackageName}}

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	{{.EntityImportAlias}} "{{.ModuleName}}/internal/core/domain/{{.EntityPackage}}"
)

// Create{{.EntityName}}Input holds the data needed to create a new {{.EntityName}}.
type Create{{.EntityName}}Input struct {
	// TODO: Add fields
}

// Update{{.EntityName}}Input holds the data needed to update an existing {{.EntityName}}.
type Update{{.EntityName}}Input struct {
	ID string
	// TODO: Add fields
}

// {{.ServiceName}}Service {{.Description}}
type {{.ServiceName}}Service struct {
	repo {{.EntityImportAlias}}.{{.EntityName}}Repository
}

// New{{.ServiceName}}Service creates a new {{.ServiceName}}Service.
func New{{.ServiceName}}Service(repo {{.EntityImportAlias}}.{{.EntityName}}Repository) *{{.ServiceName}}Service {
	return &{{.ServiceName}}Service{repo: repo}
}

// Create validates and persists a new {{.EntityName}}.
func (s *{{.ServiceName}}Service) Create(ctx context.Context, input Create{{.EntityName}}Input) (*{{.EntityImportAlias}}.{{.EntityName}}, error) {
	entity, err := {{.EntityImportAlias}}.New{{.EntityName}}(uuid.NewString() /* TODO: map input fields */)
	if err != nil {
		return nil, fmt.Errorf("invalid {{.EntityName | lower}}: %w", err)
	}
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, fmt.Errorf("creating {{.EntityName | lower}}: %w", err)
	}
	return entity, nil
}

// GetByID returns a single {{.EntityName}} by its ID.
func (s *{{.ServiceName}}Service) GetByID(ctx context.Context, id string) (*{{.EntityImportAlias}}.{{.EntityName}}, error) {
	return s.repo.FindByID(ctx, id)
}

// Update replaces the mutable fields of an existing {{.EntityName}}.
func (s *{{.ServiceName}}Service) Update(ctx context.Context, input Update{{.EntityName}}Input) (*{{.EntityImportAlias}}.{{.EntityName}}, error) {
	entity, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("finding {{.EntityName | lower}}: %w", err)
	}
	// TODO: apply input fields to entity
	if err := entity.Validate(); err != nil {
		return nil, fmt.Errorf("invalid {{.EntityName | lower}}: %w", err)
	}
	if err := s.repo.Update(ctx, entity); err != nil {
		return nil, fmt.Errorf("updating {{.EntityName | lower}}: %w", err)
	}
	return entity, nil
}

// List returns all {{.EntityName}} records.
func (s *{{.ServiceName}}Service) List(ctx context.Context) ([]*{{.EntityImportAlias}}.{{.EntityName}}, error) {
	return s.repo.List(ctx)
}
```

3. New template variables needed:
   - `PackageName` — e.g. `categories`
   - `EntityName` — e.g. `Category`
   - `EntityPackage` — e.g. `categories` (domain subpackage)
   - `EntityImportAlias` — e.g. `categoriesDomain`

---

## Task 5 — Services aggregator (services/services.go)

**Files to change:**
- `internal/generator/service.go` — add/update aggregator
- New template: `internal/generator/templates/service/services_aggregator.go.tmpl`

**What to do:**

The aggregator file `internal/core/services/services.go` must be created on the
first `hexago_add_service` call and updated on subsequent ones.

Strategy: generate or re-generate this file each time a service is added.

1. After generating the service file, call `g.upsertAggregator()`.
2. `upsertAggregator()` reads the existing `services.go` (if any), parses the
   known services, appends the new one, and rewrites the file.
   Simpler alternative: scan `internal/core/services/*/` for subdirectories,
   inspect the `*Service` types found, and regenerate `services.go` from scratch.

3. Template `services_aggregator.go.tmpl`:

```go
package {{.CoreLogic}}

import (
{{- range .Entries}}
	{{.Alias}} "{{$.ModuleName}}/internal/core/{{$.CoreLogic}}/{{.Package}}"
{{- end}}
)

type Config struct {
{{- range .Entries}}
	{{.RepoField}} {{.DomainAlias}}.{{.RepoInterface}}
{{- end}}
}

type Services struct {
{{- range .Entries}}
	{{.ServiceField}} *{{.Alias}}.{{.ServiceType}}
{{- end}}
}

func New(config *Config) *Services {
	return &Services{
{{- range .Entries}}
		{{.ServiceField}}: {{.Alias}}.New{{.ServiceType}}(config.{{.RepoField}}),
{{- end}}
	}
}
```

Note: This task is the most complex. A simpler first iteration can just append
a `// TODO: add <ServiceName>` comment and ask the developer to wire manually.

---

## Task 6 — Secondary adapter (database repository): sub-package + proper types

**Files to change:**
- `internal/generator/adapter.go` — `GenerateSecondary()` / `generateDatabaseAdapter()`
- `internal/generator/templates/adapter/database.go.tmpl`

**What to do:**

1. In `adapter.go` → `GenerateSecondary()` for type `"database"`:
   - Compute `pkgName = strings.ToLower(entityName)` (singular, matching the domain
     package, e.g. `categories`).
   - Change `adapterDir` to
     `filepath.Join("internal", "adapters", g.config.AdapterOutboundDir(), "database", pkgName)`.
   - Create subdirectory.
   - Pass additional template vars: `PackageName`, `EntityName`, `EntityPackage`,
     `EntityImportAlias`, `DomainImport`.

2. Rewrite `database.go.tmpl`:

```go
package {{.PackageName}}

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"{{.ModuleName}}/internal/core/domain"
	{{.EntityImportAlias}} "{{.ModuleName}}/internal/core/domain/{{.EntityPackage}}"
)

// {{.RepoName}} implements {{.EntityImportAlias}}.{{.EntityName}}Repository using PostgreSQL.
type {{.RepoName}} struct {
	db *sql.DB
}

// compile-time check that {{.RepoName}} satisfies the port.
var _ {{.EntityImportAlias}}.{{.EntityName}}Repository = (*{{.RepoName}})(nil)

// New{{.RepoName}} creates a new {{.RepoName}}.
func New{{.RepoName}}(db *sql.DB) *{{.RepoName}} {
	return &{{.RepoName}}{db: db}
}

// Create inserts a new {{.EntityName}}.
func (r *{{.RepoName}}) Create(ctx context.Context, e *{{.EntityImportAlias}}.{{.EntityName}}) error {
	// TODO: implement INSERT query
	return fmt.Errorf("not implemented")
}

// FindByID retrieves a {{.EntityName}} by its ID.
func (r *{{.RepoName}}) FindByID(ctx context.Context, id string) (*{{.EntityImportAlias}}.{{.EntityName}}, error) {
	// TODO: implement SELECT WHERE id=$1
	// Use: if errors.Is(err, sql.ErrNoRows) { return nil, domain.ErrNotFound }
	_ = errors.Is(nil, sql.ErrNoRows) // remove when implemented
	return nil, domain.ErrNotFound
}

// Update saves updated {{.EntityName}} fields.
func (r *{{.RepoName}}) Update(ctx context.Context, e *{{.EntityImportAlias}}.{{.EntityName}}) error {
	// TODO: implement UPDATE query
	// Check rows affected → return domain.ErrNotFound if 0
	return fmt.Errorf("not implemented")
}

// List returns all {{.EntityName}} records.
func (r *{{.RepoName}}) List(ctx context.Context) ([]*{{.EntityImportAlias}}.{{.EntityName}}, error) {
	// TODO: implement SELECT query
	return nil, fmt.Errorf("not implemented")
}
```

---

## Task 7 — Primary HTTP adapter: sub-package + split files

**Files to change:**
- `internal/generator/adapter.go` — `GeneratePrimary()` / `generateHTTPAdapter()`
- New template: `internal/generator/templates/adapter/primary/http/chi/handler_config.go.tmpl`
- New template: `internal/generator/templates/adapter/primary/http/chi/handler_methods.go.tmpl`

**What to do:**

1. In `adapter.go` → `GeneratePrimary()` for type `"http"`:
   - Compute `pkgName = strings.ToLower(entityName) + "s"`.
   - Change `adapterDir` to
     `filepath.Join("internal", "adapters", g.config.AdapterInboundDir(), "http", pkgName)`.
   - Create subdirectory.
   - Generate TWO files: `<snake_case>.go` (config/DTOs) and `handlers.go`.

2. Template `handler_config.go.tmpl` (Config, handler struct, New, Configure, DTOs):

```go
package {{.PackageName}}

import (
	"time"

	"github.com/go-chi/chi/v5"
	{{.EntityImportAlias}} "{{.ModuleName}}/internal/core/domain/{{.EntityPackage}}"
	"{{.ModuleName}}/internal/core/{{.CoreLogic}}"
	{{.ServiceImportAlias}} "{{.ModuleName}}/internal/core/{{.CoreLogic}}/{{.ServicePackage}}"
	"{{.ModuleName}}/pkg/server"
)

type Config struct {
	Router   chi.Router
	Services *{{.CoreLogic}}.Services
}

type handler struct {
	*Config
	manage *{{.ServiceImportAlias}}.{{.ServiceName}}Service
}

func New(config *Config) *handler {
	return &handler{
		Config: config,
		manage: config.Services.{{.ServiceField}},
	}
}

func (h *handler) Configure(srv server.Server) {
	h.Router.Route("/{{.RoutePrefix}}", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
	})
}

// --- DTOs ---

type create{{.EntityName}}Request struct {
	// TODO: Add request fields
}

type update{{.EntityName}}Request struct {
	// TODO: Add request fields
}

type {{.EntityVarName}}Response struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	// TODO: Add response fields
}

func to{{.EntityName}}Response(e *{{.EntityImportAlias}}.{{.EntityName}}) {{.EntityVarName}}Response {
	return {{.EntityVarName}}Response{
		// TODO: Map fields
	}
}
```

3. Template `handler_methods.go.tmpl` (HTTP handler methods):

```go
package {{.PackageName}}

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	{{.ServiceImportAlias}} "{{.ModuleName}}/internal/core/{{.CoreLogic}}/{{.ServicePackage}}"
	"{{.ModuleName}}/pkg/httphelpers"
)

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.manage.List(r.Context())
	if err != nil {
		httphelpers.RespondHTTPError(w, err)
		return
	}
	resp := make([]{{.EntityVarName}}Response, 0, len(items))
	for _, item := range items {
		resp = append(resp, to{{.EntityName}}Response(item))
	}
	httphelpers.RespondJSON(w, http.StatusOK, resp)
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var req create{{.EntityName}}Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httphelpers.RespondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	item, err := h.manage.Create(r.Context(), {{.ServiceImportAlias}}.Create{{.EntityName}}Input{
		// TODO: map req fields
	})
	if err != nil {
		httphelpers.RespondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httphelpers.RespondJSON(w, http.StatusCreated, to{{.EntityName}}Response(item))
}

func (h *handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.manage.GetByID(r.Context(), id)
	if err != nil {
		httphelpers.RespondHTTPError(w, err)
		return
	}
	httphelpers.RespondJSON(w, http.StatusOK, to{{.EntityName}}Response(item))
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req update{{.EntityName}}Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httphelpers.RespondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	item, err := h.manage.Update(r.Context(), {{.ServiceImportAlias}}.Update{{.EntityName}}Input{
		ID: id,
		// TODO: map req fields
	})
	if err != nil {
		httphelpers.RespondHTTPError(w, err)
		return
	}
	httphelpers.RespondJSON(w, http.StatusOK, to{{.EntityName}}Response(item))
}
```

4. New template variables: `PackageName`, `EntityName`, `EntityVarName` (lowercase first
   letter), `EntityPackage`, `EntityImportAlias`, `ServicePackage`, `ServiceImportAlias`,
   `ServiceName`, `ServiceField`, `RoutePrefix`, `CoreLogic`.

---

## Task 8 — HTTP wiring: accept services parameter

**Files to change:**
- `internal/generator/templates/adapter/primary/http/chi/http_adapter.go.tmpl`
  (and echo/gin/fiber/stdlib equivalents)

**What to do:**

1. Change function signature:
   ```go
   // before
   func New(cfg *httpsrv.ServerConfig) server.Server {
   // after
   func New(cfg *httpsrv.ServerConfig, services *{{.CoreLogic}}.Services) server.Server {
   ```

2. Add `{{.CoreLogic}}` import: `"{{.ModuleName}}/internal/core/{{.CoreLogic}}"`.

3. Add the `/api/v1` route group pattern with a `TODO` comment:
   ```go
   srv.Router.Route("/api/v1", func(r chi.Router) {
       // TODO: register your handlers here
       // srv.Use(yourhandler.New(&yourhandler.Config{
       //     Router:   r,
       //     Services: services,
       // }))
   })
   ```

---

## Task 9 — cmd/run.go: add database wiring

**Files to change:**
- `internal/generator/templates/cmd/run_http_server.go.tmpl`

**What to do:**

Add three clearly-delimited sections between logger init and `http.New(...)`:

```go
// ── Database ──────────────────────────────────────────────────────
db, err := sql.Open("postgres", cfg.Database.URL)
if err != nil {
    return fmt.Errorf("opening database: %w", err)
}
defer db.Close()

if err := db.Ping(); err != nil {
    log.Error("Database unreachable: %v", err)
}

// ── Repositories (secondary adapters) ────────────────────────────
// TODO: instantiate repositories
// exampleRepo := exampleRepo.NewExampleRepository(db)

// ── Services (core) ───────────────────────────────────────────────
// TODO: wire services
// services := services.New(&services.Config{
//     ExampleRepository: exampleRepo,
// })
```

And update `http.New(...)` call to pass services:
```go
srv := http.New(&httpsrv.ServerConfig{
    Config: cfg,
    Logger: log,
}, services)
```

Add imports: `"database/sql"`, `"fmt"`, `_ "github.com/lib/pq"`,
`"{{.ModuleName}}/internal/core/{{.CoreLogic}}"`.

---

## Task 10 — Generator variables: add pluralization helper

**Files to change:**
- `pkg/utils/utils.go` (or new file `pkg/utils/strings.go`)

**What to do:**

Add `ToPlural(s string) string` — a simple English pluralizer sufficient for
typical entity names:

```go
func ToPlural(s string) string {
    s = strings.ToLower(s)
    switch {
    case strings.HasSuffix(s, "y"):
        return s[:len(s)-1] + "ies"
    case strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
         strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh"):
        return s + "es"
    default:
        return s + "s"
    }
}
```

This covers: `Category→categories`, `Product→products`, `Movement→movements`,
`Status→statuses`, `Box→boxes`.

---

## Execution order (recommended)

Run these tasks one at a time, testing after each with `go build ./...` in the
hexago project and with a fresh `hexago init` + `hexago_add_*` run.

```
Task 10 → Task 1 → Task 2 → Task 3 → Task 4 → Task 6 → Task 7 → Task 8 → Task 9 → Task 5
```

Tasks 5 (services aggregator) is the most complex; leave it for last or
implement the simple "append TODO comment" version first and iterate.

---

## Files NOT needing changes

- `pkg/httpserver/` — already framework-agnostic, no changes needed.
- `pkg/server/server.go` — interface already correct.
- `internal/observability/` — generated correctly.
- `internal/config/` — generated correctly.
- `pkg/httphelpers/` — must exist in generated project (already there via `pkg/`
  scaffold); no template changes needed.
- Migration templates — no changes needed.
- Worker/tool templates — out of scope for this refactor.
