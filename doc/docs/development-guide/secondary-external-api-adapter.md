# Secondary External API Adapter

How to add an external API client adapter to a HexaGo-generated project.

---

## Overview

External API adapters (secondary/outbound) connect your application to third-party services like data providers, payment gateways, or messaging platforms.

---

## Initialize Project

Create the project with all required features:

```shell
$ hexago init chuck-norris \
  --project-type service \
  --explicit-ports \             
  --module github.com/padiazg/chuck-norris

📋 Project Configuration:
  Name:              chuck-norris
  Module:            github.com/padiazg/chuck-norris
  Project Type:      service
  Adapter Style:     primary-secondary
  Core Logic:        services
  Docker:            false
  Observability:     false
  Migrations:        false
  Workers:           false
  Example Code:      false

🚀 Generating project chuck-norris...
📁 Creating directory structure...
📝 Generating files...
📦 Initializing go module...
go: creating new go.mod: module github.com/padiazg/chuck-norris
go: to add module requirements and sums:
    go mod tidy
📦 Adding dependencies...
🧹 Running go mod tidy...
✨ Formatting code...

✅ Project generated successfully!

📚 Next steps:
  cd chuck-norris
  go run main.go run

📖 Read the README.md for more information about the project structure.

# move to the new project folder
$ cd chuck-norris
```

## Add a domain value object

### 1. Generate the value object

```shell
$ hexago add domain valueobject Joke \
  --fields "id:string,url:string,value:string"

📦 Adding value object: Joke
   Project: chuck-norris

📝 Creating value object file: internal/core/domain/joke/joke.go
📝 Creating test file: internal/core/domain/joke/joke_test.go

✅ Value object added successfully!

📝 Next steps:
  1. Ensure immutability (no setter methods)
  2. Implement validation in constructor
  3. Implement Equals method for value comparison
```

This generates:

- `internal/core/domain/joke/joke.go`
- `internal/core/domain/joke/joke_test.go`

### 2. Update the value object code

You can emove the `joke_test.go` for now (you can add tests later). We'll also add a **port** interface in the same package—this defines what the external API client must implement.

```go
// internal/core/domain/joke/joke.go
package joke

// Joke is a value object representing Joke.
// Value objects are immutable and compared by value, not identity.
type Joke struct {
    ID    string `json:"id"`
    URL   string `json:"url"`
    Value string `json:"value"`
}

// String returns string representation
func (v Joke) String() string {
    return v.Value
}
```

```go
// internal/core/domain/joke/port.go
package joke

import "context"

type JokeProvider interface {
    Ping(ctx context.Context) error
    GetRandom(ctx context.Context) (*Joke, error)
    GetByCategory(ctx context.Context, category string) (*Joke, error)
    ListCategories(ctx context.Context) ([]string, error)
    Search(ctx context.Context, query string) ([]string, error)
}
```

## Add a secondary adapter

A **secondary adapter** (also called "driven" or "outbound" adapter) implements the port interface we defined earlier (`JokeProvider`). It contains the actual logic for making HTTP calls to the external Chuck Norris API. The domain and services remain agnostic to how the external API is called—they only know about the port interface.

### 1. Generate the adapter

```shell
$ hexago add adapter secondary external JokeClient --port JokeProvider           

📦 Adding secondary adapter: JokeClient (external)
   Project: chuck-norris
   Adapter dir: secondary

📝 Creating adapter file: internal/adapters/secondary/external/joke_client.go
📝 Creating test file: internal/adapters/secondary/external/joke_client_test.go

✅ Secondary adapter added successfully!

📝 Next steps:
  1. Implement the port interface methods
  2. Add database queries or external API calls
  3. Wire up dependencies in the DI container  
```

### 2. Update the adapter code

```go
// internal/adapters/secondary/external/joke_client.go
package external

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    domain "github.com/padiazg/chuck-norris/internal/core/domain/joke"
)

var _ domain.JokeProvider = (*Client)(nil)

// Client implements communication with external service
type Client struct {
    client  *http.Client
    baseURL string
}

// Response types for Chuck Norris API
type ChuckResponse struct {
    IconURL string `json:"icon_url"`
    ID      string `json:"id"`
    URL     string `json:"url"`
    Value   string `json:"value"`
}

type SearchResponse struct {
    Total  int            `json:"total"`
    Result []SearchResult `json:"result"`
}

type SearchResult struct {
    Categories []string `json:"categories"`
    CreatedAt  string   `json:"created_at"`
    IconURL    string   `json:"icon_url"`
    ID         string   `json:"id"`
    UpdatedAt  string   `json:"updated_at"`
    URL        string   `json:"url"`
    Value      string   `json:"value"`
}

// NewClient creates a new Client
func NewClient() *Client {
    return &Client{
        client:  &http.Client{Timeout: 30 * time.Second},
        baseURL: "https://api.chucknorris.io/jokes",
    }
}

func (c *Client) Ping(ctx context.Context) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL+"/categories", nil)
    if err != nil {
        return fmt.Errorf("joke client ping: create request: %w", err)
    }

    resp, err := c.client.Do(req)
    if resp != nil {
        defer resp.Body.Close()
    }
    if err != nil {
        return fmt.Errorf("joke client ping: execute request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("joke client ping: status %s", resp.Status)
    }

    return nil
}

func (c *Client) GetRandom(ctx context.Context) (*domain.Joke, error) {
    url := c.baseURL + "/random"

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("joke client get random: create request: %w", err)
    }

    resp, err := c.client.Do(req)
    if resp != nil {
        defer resp.Body.Close()
    }
    if err != nil {
        return nil, fmt.Errorf("joke client get random: execute request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("joke client get random: status %s", resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("joke client get random: read body: %w", err)
    }

    var chuckResp ChuckResponse
    if err := json.Unmarshal(body, &chuckResp); err != nil {
        return nil, fmt.Errorf("joke client get random: unmarshal: %w", err)
    }

    return &domain.Joke{
        ID:    chuckResp.ID,
        URL:   chuckResp.URL,
        Value: chuckResp.Value,
    }, nil
}

func (c *Client) GetByCategory(ctx context.Context, category string) (*domain.Joke, error) {
    url := fmt.Sprintf("%s/random?category=%s", c.baseURL, category)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("joke client get by category: create request: %w", err)
    }

    resp, err := c.client.Do(req)
    if resp != nil {
        defer resp.Body.Close()
    }
    if err != nil {
        return nil, fmt.Errorf("joke client get by category: execute request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("joke client get by category: status %s", resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("joke client get by category: read body: %w", err)
    }

    var chuckResp ChuckResponse
    if err := json.Unmarshal(body, &chuckResp); err != nil {
        return nil, fmt.Errorf("joke client get by category: unmarshal: %w", err)
    }

    return &domain.Joke{
        ID:    chuckResp.ID,
        URL:   chuckResp.URL,
        Value: chuckResp.Value,
    }, nil
}

func (c *Client) ListCategories(ctx context.Context) ([]string, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/categories", nil)
    if err != nil {
        return nil, fmt.Errorf("joke client get categories: create request: %w", err)
    }

    resp, err := c.client.Do(req)
    if resp != nil {
        defer resp.Body.Close()
    }
    if err != nil {
        return nil, fmt.Errorf("joke client get categories: execute request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("joke client get categories: status %s", resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("joke client get categories: read body: %w", err)
    }

    var catResp []string
    if err := json.Unmarshal(body, &catResp); err != nil {
        return nil, fmt.Errorf("joke client get categories: unmarshal: %w", err)
    }

    return catResp, nil
}

func (c *Client) Search(ctx context.Context, query string) ([]string, error) {
    if query == "" {
        return nil, fmt.Errorf("joke client get search: must provide query")
    }

    url := fmt.Sprintf("%s/search?query=%s", c.baseURL, query)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("joke client get search: create request: %w", err)
    }

    resp, err := c.client.Do(req)
    if resp != nil {
        defer resp.Body.Close()
    }
    if err != nil {
        return nil, fmt.Errorf("joke client get search: execute request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("joke client get search: status %s", resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("joke client get search: read body: %w", err)
    }

    var searchResp SearchResponse
    if err := json.Unmarshal(body, &searchResp); err != nil {
        return nil, fmt.Errorf("joke client get search: unmarshal: %w", err)
    }

    var list []string
    if searchResp.Total > 0 {
        for _, joke := range searchResp.Result {
            list = append(list, joke.Value)
        }
    }

    return list, nil
}
```

## Add a service

### 1. Generate the service

```shell
$ hexago add service Joke

📦 Adding service: Joke
   Project: chuck-norris
   Module: github.com/padiazg/chuck-norris
   Logic dir: services

📝 Creating service file: internal/core/services/joke/joke.go
📝 Creating test file: internal/core/services/joke/joke_test.go
📝 Updating services aggregator: internal/core/services/services.go

✅ Service added successfully!

📝 Next steps:
  1. Implement the business logic in the Execute method
  2. Add any required dependencies to the constructor
  3. Write tests in the generated test file

```

### 2. Update the service code

```go
// internal/core/services/joke/joke.go
package joke

import (
    "context"
    "fmt"

    domain "github.com/padiazg/chuck-norris/internal/core/domain/joke"
)

// JokeService implements Joke logic
type Service struct {
    provider domain.JokeProvider
}

type Config struct {
    Provider domain.JokeProvider
}

// NewJokeService creates a new JokeService.
func New(cfg *Config) *Service {
    return &Service{provider: cfg.Provider}
}

// Execute runs the service logic.
func (s *Service) Random(ctx context.Context) (*domain.Joke, error) {
    return s.provider.GetRandom(ctx)
}

func (s *Service) ByCategory(ctx context.Context, category string) (*domain.Joke, error) {
    if category == "" {
        return nil, fmt.Errorf("must provide category")
    }

    return s.provider.GetByCategory(ctx, category)
}

func (s *Service) ListCategories(ctx context.Context) ([]string, error) {
    return s.provider.ListCategories(ctx)
}

func (s *Service) Search(ctx context.Context, query string) ([]string, error) {
    return s.provider.Search(ctx, query)
}
```

```go
// internal/core/services/services.go
package services

import (
    domain "github.com/padiazg/chuck-norris/internal/core/domain/joke"
    svc "github.com/padiazg/chuck-norris/internal/core/services/joke"
)

// Config holds the repository dependencies required to initialise entity-bound services.
type Config struct {
    JokeProvider domain.JokeProvider
}

// Services aggregates all domain services.
type Services struct {
    Joke *svc.Service
}

// New wires all services using the provided repository config.
func New(config *Config) *Services {
    return &Services{
        Joke: svc.New(&svc.Config{
            Provider: config.JokeProvider,
        }),
    }
}
```
---

## Wire-up in `cmd`

We won't use the original `cmd/run.go` command to start a server. Instead, we will implement our own one-time run commands (like a CLI tool). This approach is useful for CLI applications that need to perform specific tasks rather than running a long-lived server.
> it's safe to remove `cmd/run.go` and `internal/core/services/processor.go`

All the commands uses timeout contexts and os signals

### 1. joke

```go
// cmd/joke.go
package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    client "github.com/padiazg/chuck-norris/internal/adapters/secondary/external"
    "github.com/padiazg/chuck-norris/internal/core/domain/joke"
    services "github.com/padiazg/chuck-norris/internal/core/services"
    "github.com/padiazg/chuck-norris/pkg/logger"
    "github.com/spf13/cobra"
)

// jokeCmd represents the joke command
var jokeCmd = &cobra.Command{
    Use:   "joke",
    Short: "Random Chuck Norris joke",
    Long:  `Displays a random Chuck Norris joke`,
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg := GetConfig()

        // Initialize logger from config
        log := logger.New(&logger.Config{
            Level:  cfg.LogLevel,
            Format: cfg.LogFormat,
        })

        // ── Secondary Adapters ────────────────────────────────────────────
        provider := client.NewClient()

        // ── Services (core) ───────────────────────────────────────────────
        svc := services.New(&services.Config{
            JokeProvider: provider,
        })

        // Configure context with cancellation for graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()

        // Channel to capture OS signals
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

        // Channel for processor errors
        errChan := make(chan error, 1)
        resChan := make(chan string)

        category, _ := cmd.Flags().GetString("category")

        go func() {
            var (
                res *joke.Joke
                err error
            )

            if category != "" {
                res, err = svc.Joke.ByCategory(ctx, category)
            } else {
                res, err = svc.Joke.Random(ctx)
            }

            if err != nil {
                errChan <- fmt.Errorf("joke: %w", err)
            }

            resChan <- res.Value
            close(resChan)
        }()

        // Wait for joke, signal or error
        select {
        case joke := <-resChan:
            fmt.Printf("%s", joke)
        case sig := <-sigChan:
            log.Info("Received signal %v, initiating graceful shutdown...", sig)
            cancel()
        case err := <-errChan:
            cancel()
            return fmt.Errorf("getting joke: %w", err)
        case <-ctx.Done():
            log.Warn("Timeout, forcing exit")
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(jokeCmd)
    jokeCmd.Flags().StringP("category", "c", "", "Filters the joke for a category")
}
```

Build and test

```shell
make build

$ ./chuck-norris joke
Most people have Microwave ovens. Chuck Norris has a Megawave oven.

$ ./chuck-norris joke --category sport
Chuck Norris plays racquetball with a waffle iron and a bowling ball.
```

### 2. categories

```go
// cmd/categories.go
package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    client "github.com/padiazg/chuck-norris/internal/adapters/secondary/external"
    services "github.com/padiazg/chuck-norris/internal/core/services"
    "github.com/padiazg/chuck-norris/pkg/logger"
    "github.com/spf13/cobra"
)

// categoriesCmd represents the categories command
var categoriesCmd = &cobra.Command{
    Use:   "categories",
    Short: "List",
    Long:  `List jokes categories`,
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg := GetConfig()

        // Initialize logger from config
        log := logger.New(&logger.Config{
            Level:  cfg.LogLevel,
            Format: cfg.LogFormat,
        })

        // ── Secondary Adapters ────────────────────────────────────────────
        provider := client.NewClient()

        // ── Services (core) ───────────────────────────────────────────────
        svc := services.New(&services.Config{
            JokeProvider: provider,
        })

        // Configure context with cancellation for graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
        defer cancel()

        // Channel to capture OS signals
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

        // Channel for processor errors
        errChan := make(chan error, 1)
        resChan := make(chan []string)

        go func() {
            res, err := svc.Joke.ListCategories(ctx)
            if err != nil {
                errChan <- fmt.Errorf("list: %w", err)
            }

            resChan <- res
            close(resChan)
        }()

        // Wait for list, signal or error
        select {
        case list := <-resChan:
            bytes, err := json.Marshal(list)
            if err != nil {
                return fmt.Errorf("marshaling list: %w", err)
            }
            fmt.Printf("%s", string(bytes))
        case sig := <-sigChan:
            log.Info("Received signal %v, initiating graceful shutdown...", sig)
            cancel()
        case err := <-errChan:
            cancel()
            return fmt.Errorf("getting list: %w", err)
        case <-ctx.Done():
            log.Warn("Timeout, forcing exit")
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(categoriesCmd)
}
```

Build and test

```shell
make build

$ ./chuck-norris categories | jq
[
  "animal",
  "career",
  "celebrity",
  "dev",
  "explicit",
  "fashion",
  "food",
  "history",
  "money",
  "movie",
  "music",
  "political",
  "religion",
  "science",
  "sport",
  "travel"
]
```

### 3. search

```go
// cmd/search.go
package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    client "github.com/padiazg/chuck-norris/internal/adapters/secondary/external"
    services "github.com/padiazg/chuck-norris/internal/core/services"
    "github.com/padiazg/chuck-norris/pkg/logger"
    "github.com/spf13/cobra"
)

// searchCmd represents the categories command
var searchCmd = &cobra.Command{
    Use:   "search <query>",
    Short: "Search",
    Long: `Search jokes by a given string
Returns a list of jokes

Example:
  chuck-norris search 
`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        query := args[0]
        if query == "" {
            return fmt.Errorf("search: must provide query")
        }

        cfg := GetConfig()

        // Initialize logger from config
        log := logger.New(&logger.Config{
            Level:  cfg.LogLevel,
            Format: cfg.LogFormat,
        })

        // ── Secondary Adapters ────────────────────────────────────────────
        provider := client.NewClient()

        // ── Services (core) ───────────────────────────────────────────────
        svc := services.New(&services.Config{
            JokeProvider: provider,
        })

        // Configure context with cancellation for graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
        defer cancel()

        // Channel to capture OS signals
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

        // Channel for processor errors
        errChan := make(chan error, 1)
        resChan := make(chan []string)

        go func() {
            res, err := svc.Joke.Search(ctx, query)
            if err != nil {
                errChan <- fmt.Errorf("list: %w", err)
            }

            resChan <- res
            close(resChan)
        }()

        // Wait for list, signal or error
        select {
        case list := <-resChan:
            bytes, err := json.Marshal(list)
            if err != nil {
                return fmt.Errorf("marshaling list: %w", err)
            }
            fmt.Printf("%s", string(bytes))
        case sig := <-sigChan:
            log.Info("Received signal %v, initiating graceful shutdown...", sig)
            cancel()
        case err := <-errChan:
            cancel()
            return fmt.Errorf("getting list: %w", err)
        case <-ctx.Done():
            log.Warn("Timeout, forcing exit")
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(searchCmd)

}
```

Build and test

```shell
make build

$ ./chuck-norris search hospital | jq

[
  "Chuck Norris built the hospital where he was born.",
  "When Chuck Norris was born, the whole hospital cried.",
  "Chuck Norris is so hard he jumped from the Eiffel Tower broke both his legs and walked to the hospital",
  ...
]
```

## Summary

This guide demonstrated how to build a CLI application that consumes an external API using hexagonal architecture:

| Step | Layer | Component | HexaGo Command |
| --- | --- | --- | --- |
| 1 | - | Project initialization | `hexago init` |
| 2 | Domain | Value object (Joke) | `hexago add domain valueobject` |
| 3 | Domain + Ports | Port interface (JokeProvider) | Manual in `internal/core/domain/` |
| 4 | Secondary | External API Client | `hexago add adapter secondary external` |
| 5 | Core | Service implementation | `hexago add service` |
| 6 | Primary | CLI Commands | Manual in `cmd/` |

Key patterns demonstrated:

- **Value objects**: Immutable domain types with value semantics
- **Port interfaces**: Define contracts that adapters must implement
- **Secondary adapters**: Implement outbound ports (external API clients)
- **Service layer**: Orchestrates domain logic, depends only on ports
- **CLI commands**: One-time run commands using context and signal handling for graceful shutdown
- **Dependency rule**: Adapters → Core (never the other way)
