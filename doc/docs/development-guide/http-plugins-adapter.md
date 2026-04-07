---
title: HTTP Adapter Plugins
description: How to extend HTTP adapters with new routes, route groups, and middleware for each supported framework.
weight: 30
---

# HTTP Adapter Plugin Pattern

HexaGo generates HTTP adapters with a plugin-like architecture where each handler registers itself on the server via the `ServerHandler` interface. This guide covers how to add new routes, create route groups, and apply middleware for each supported framework.

## Architecture Overview

The HTTP adapter follows a **wiring pattern** where:

1. **`pkg/httpserver`** - Framework-specific HTTP server implementations (`Echo`, `Gin`, `Chi`, `Fiber`, `stdlib`)
2. **`internal/adapters/primary/http/http_adapter.go`** - Wires all handlers together
3. **Handler packages** - Each handler (ping, health, metrics, custom) implements `ServerHandler`

```
internal/adapters/primary/http/
├── http_adapter.go    # Wires all handlers (main entry point)
├── ping/              # Example handler: GET /ping
├── health/            # Example handler: GET /health
├── metrics/           # Example handler: GET /metrics
└── yourhandler/       # Your custom handlers go here
```

## The ServerHandler Interface

All handlers implement the `ServerHandler` interface from `pkg/server`:

```go
type ServerHandler interface {
    Configure(Server)
}
```

Each framework's `httpserver.Server` exposes its native router via a `Handler` struct:

| Framework | Router Field | Group Method |
|-----------|--------------|--------------|
| Echo      | `Echo *echo.Echo` | `Echo.Group(path)` |
| Gin       | `Router *gin.Engine` | `Router.Group(path)` |
| Chi       | `Router chi.Router` | `Router.Route(path, fn)` |
| Fiber     | `App *fiber.App` | `App.Group(path)` |
| stdlib    | `Mux *http.ServeMux` | Manual via nested `ServeMux` |

## Adding a New Route Handler

### Step 1: Create the Handler Package

Create a new directory under `internal/adapters/primary/http/`:

```bash
mkdir -p internal/adapters/primary/http/users
```

### Step 2: Implement the Handler

Create `internal/adapters/primary/http/users/handler.go`:

```go
package users

import (
    "net/http"

    "github.com/labstack/echo/v4"  // Change based on your framework
    "github.com/myproject/pkg/server"
)

// Config holds handler configuration
type Config struct {
    Path     string
    Echo     *echo.Echo            // Use: Router/Route/Mux based on framework
    Services *myservice.Services   // Your core services
}

type handler struct {
    *Config
}

// New creates a new users handler
func New(config *Config) *handler {
    return &handler{Config: config}
}

// Configure registers routes on the server
func (h *handler) Configure(srv server.Server) {
    // Framework-specific route registration
    h.Echo.GET(h.Path, h.List)
    h.Echo.POST(h.Path, h.Create)
}

// List handles GET /users
func (h *handler) List(c echo.Context) error {
    // Implementation
    return c.JSON(http.StatusOK, map[string]string{
        "message": "users list",
    })
}

// Create handles POST /users
func (h *handler) Create(c echo.Context) error {
    // Implementation
    return c.JSON(http.StatusCreated, map[string]string{
        "message": "user created",
    })
}
```

### Step 3: Register in http_adapter.go

Add your handler to `internal/adapters/primary/http/http_adapter.go`:

```go
import (
    "myproject/internal/adapters/primary/http/users"
)

// In New() function:
func New(cfg *httpsrv.ServerConfig, services *services.Services) server.Server {
    srv := httpsrv.New(cfg).(*httpsrv.Server)

    // Existing handlers...
    srv.Use(ping.New(&ping.Config{...}))

    // Add your handler
    srv.Use(users.New(&users.Config{
        Path:     "/users",
        Echo:     srv.Echo,   // Use: Router/App/Mux based on framework
        Services: services,
    }))

    return srv
}
```

---

## Framework-Specific Patterns

### Echo

**Route Groups:**

```go
// In http_adapter.go
v1 := srv.Echo.Group("/api/v1")

// Apply group-scoped middleware
v1.Use(echomiddleware.RequestID())
v1.Use(echomiddleware.Logger())
v1.Use(echomiddleware.Recover())

// Register handler with group
srv.Use(users.New(&users.Config{
    Group:    v1,  // Pass group instead of srv.Echo
    Services: services,
}))
```

**Handler receives group:**

```go
type Config struct {
    Group    *echo.Group  // Use this instead of Echo
    Services *services.Services
}

func (h *handler) Configure(srv server.Server) {
    h.Group.GET(h.Path, h.List)
    h.Group.POST(h.Path, h.Create)
}
```

---

### Gin

**Route Groups:**

```go
// In http_adapter.go
v1 := srv.Router.Group("/api/v1")

// Apply group-scoped middleware
v1.Use(gin.Logger())
v1.Use(gin.Recovery())

// Register handler
srv.Use(users.New(&users.Config{
    Group:    v1,
    Services: services,
}))
```

**Handler:**

```go
type Config struct {
    Group    *gin.RouterGroup
    Services *services.Services
}

func (h *handler) Configure(srv server.Server) {
    h.Group.GET(h.Path, h.List)
    h.Group.POST(h.Path, h.Create)
}

// Handler method
func (h *handler) List(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "users list"})
}
```

---

### Chi

**Route Groups:**

```go
// In http_adapter.go - chi uses Route() instead of Group()
srv.Router.Route("/api/v1", func(r chi.Router) {
    // Apply middleware to this group
    r.Use(chimiddleware.RequestID)
    r.Use(chimiddleware.Logger)
    r.Use(chimiddleware.Recoverer)

    // Handler receives router
    srv.Use(users.New(&users.Config{
        Router:   r,
        Services: services,
    }))
})
```

**Handler:**

```go
type Config struct {
    Router   chi.Router
    Services *services.Services
}

func (h *handler) Configure(srv server.Server) {
    h.Router.Get(h.Path, h.List)
    h.Router.Post(h.Path, h.Create)
}

// Handler method - stdlib http.HandlerFunc signature
func (h *handler) List(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "users list"})
}
```

---

### Fiber

**Route Groups:**

```go
// In http_adapter.go
v1 := srv.App.Group("/api/v1")

// Apply group-scoped middleware
v1.Use(fibermiddleware.RequestID())
v1.Use(fibermiddleware.Logger())
v1.Use(fibermiddleware.Recover())

// Register handler
srv.Use(users.New(&users.Config{
    Group:    v1,
    Services: services,
}))
```

**Handler:**

```go
type Config struct {
    Group    fiber.Router
    Services *services.Services
}

func (h *handler) Configure(srv server.Server) {
    h.Group.Get(h.Path, h.List)
    h.Group.Post(h.Path, h.Create)
}

// Handler method - Fiber context
func (h *handler) List(c *fiber.Ctx) error {
    return c.JSON(map[string]string{"message": "users list"})
}
```

---

### stdlib

**Route Groups:**

stdlib doesn't have built-in groups, so use nested `ServeMux` with middleware wrapping:

```go
// In http_adapter.go
apiMux := http.NewServeMux()

// Middleware wrapper pattern
var apiHandler http.Handler = apiMux
apiHandler = yourlogging.Middleware(apiHandler)
apiHandler = yourauth.Middleware(services)(apiHandler)

// Mount with prefix stripping
srv.Mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiHandler))

// Register handler
srv.Use(users.New(&users.Config{
    Mux:      apiMux,
    Services: services,
}))
```

**Handler:**

```go
type Config struct {
    Mux      *http.ServeMux
    Services *services.Services
}

func (h *handler) Configure(srv server.Server) {
    h.Mux.HandleFunc(h.Path, h.List)
    h.Mux.HandleFunc(h.Path+"/", h.Create)
}

// Handler method - stdlib http.HandlerFunc
func (h *handler) List(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "users list"})
}
```

---

## Middleware Patterns

### Global Middleware (in pkg/httpserver)

Edit the server template in `internal/generator/templates/pkg/httpserver/http_server_<framework>.go.tmpl`:

```go
// Echo example
func New(cfg *ServerConfig) srv.Server {
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(yourauth.GlobalMiddleware())
    // ...
}
```

### Route-Group Middleware

Applied in `http_adapter.go` to a specific group:

```go
// Echo
v1 := srv.Echo.Group("/api/v1")
v1.Use(echomiddleware.Logger())

// Gin
v1 := srv.Router.Group("/api/v1")
v1.Use(gin.Logger())

// Chi - apply in Route() callback
srv.Router.Route("/api/v1", func(r chi.Router) {
    r.Use(chimiddleware.Logger)
})

// Fiber
v1 := srv.App.Group("/api/v1")
v1.Use(fibermiddleware.Logger())

// stdlib - wrap handler
var handler http.Handler = apiMux
handler = loggingMiddleware(handler)
```

### Handler-Level Middleware

Wrap individual routes in your handler:

```go
// Echo
func (h *handler) Configure(srv server.Server) {
    handler := func(c echo.Context) error { ... }
    wrapped := middleware.Logger()(handler)
    h.Echo.GET(h.Path, wrapped)
}

// Gin
func (h *handler) Configure(srv server.Server) {
    handler := func(c *gin.Context) { ... }
    wrapped := gin.Logger().Func()(handler)
    h.Router.GET(h.Path, wrapped)
}
```

---

## Common Handler Structure

Each handler should follow this pattern:

```go
package handler

import (
    "net/http"
    "github.com/myproject/pkg/server"
)

// Config holds all configuration needed by the handler
type Config struct {
    // Common fields
    Path     string
    Services *myservice.Services

    // Framework-specific router
    // Choose ONE based on framework:
    Echo     *echo.Echo
    Router   *gin.Engine
    App      *fiber.App
    Mux      *http.ServeMux
    Group    *echo.Group  // for group-scoped routes
}

// handler embeds Config and implements ServerHandler
type handler struct {
    *Config
}

// New creates a new handler instance
func New(config *Config) *handler {
    return &handler{Config: config}
}

// Configure registers routes on the server
func (h *handler) Configure(srv server.Server) {
    // Register routes using framework-specific router
    // See examples above for each framework
}

// Handler methods...
```

---

## Best Practices

1. **One handler per domain concept** - Keep handlers focused (users, products, auth)
2. **Use the Config pattern** - Makes handlers testable and configurable
3. **Follow framework conventions** - Use the idiomatic router type for each framework
4. **Group related routes** - Use route groups for `/api/v1` style APIs
5. **Apply middleware at the right level**:
   - Global: in `pkg/httpserver` server creation
   - Group: in `http_adapter.go` when creating groups
   - Handler: inside individual handler methods

---

## Troubleshooting

**Handler not registering routes?**

- Ensure you call `srv.Use(yourhandler.New(...))` in `http_adapter.go`
- Verify the handler implements `ServerHandler` interface with `Configure(Server)` method

**Middleware not applying?**

- Check you're using the correct middleware package for your framework
- Ensure middleware is applied to the right level (global/group/handler)

**Router field nil?**

- Verify you're passing the correct framework-specific router from `http_adapter.go`
- Each framework uses a different field name: `Echo`, `Router`, `App`, `Mux`

---

## Related Documentation

- [Architecture Overview](../architecture/overview.md)
- [Project Structure](../architecture/project-structure.md)
- [Add Adapter Command](../commands/add-adapter.md)