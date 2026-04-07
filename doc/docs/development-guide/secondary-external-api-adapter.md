# Secondary External API Adapter

How to add an external API client adapter to a HexaGo-generated project.

---

## Overview

External API adapters (secondary/outbound) connect your application to third-party services like data providers, payment gateways, or messaging platforms.

---

## Generate with HexaGo

Use the `hexago add adapter` command for secondary adapters:

```bash
hexago add adapter secondary external JokeClient \
  --working_directory /path/to/project
```

This generates:  

- `internal/adapters/secondary/external/joke_client.go`  
- `internal/adapters/secondary/external/joke_client_test.go`  

---

## Generated Structure

After generation, your adapter looks like this:

```go
package external

import (
    "context"
    "net/http"
    "time"

    "myapp/internal/core/services/ports"
)

var _ ports.JokeProvider = (*JokeClient)(nil)

type JokeClient struct {
    client  *http.Client
    baseURL string
}

func New() *JokeClient {
    return &JokeClient{
        client:  &http.Client{Timeout: 30 * time.Second},
        baseURL: "https://api.chucknorris.io",
    }
}
```

---

## Implementing an API Client

### 1. Define the Port (in services)

```go
// internal/core/services/ports/joke_provider.go
package ports

type JokeProvider interface {
    Ping(ctx context.Context) error
    GetRandomJoke(ctx context.Context) (*Joke, error)
    GetJokesByCategory(ctx context.Context, category string) ([]*Joke, error)
}

type Joke struct {
    ID       string
    Value    string
    Category string
}
```

### 2. Implement the Adapter

Example using the Chuck Norris API:
> Note that most of the code from the function is repeated, so you could use a third-party http client like Resty
> or write your own client so code doesn't repeat (DRY)

```go
// internal/adapters/secondary/joke/client.go
package joke

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "myapp/internal/core/domain"
    "myapp/internal/core/services/ports"
)

var _ ports.JokeProvider = (*Client)(nil)

// Response types for Chuck Norris API
type ChuckResponse struct {
    ID        string `json:"id"`
    Value     string `json:"value"`
    Category  string `json:"category"`
}

type CategoryResponse struct {
    Categories []string `json:"categories"`
}

type Client struct {
    client  *http.Client
    baseURL string
}

func New() *Client {
    return &Client{
        client:  &http.Client{Timeout: 30 * time.Second},
        baseURL: "https://api.chucknorris.io/j2",
    }
}

func (c *Client) Ping(ctx context.Context) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/categories", nil)
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

func (c *Client) GetRandomJoke(ctx context.Context) (*domain.Joke, error) {
    url := c.baseURL + "/jokes/random"
    
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
        ID:       chuckResp.ID,
        Value:    chuckResp.Value,
        Category: chuckResp.Category,
    }, nil
}

func (c *Client) GetJokesByCategory(ctx context.Context, category string) ([]*domain.Joke, error) {
    // Chuck Norris API doesn't support listing multiple jokes by category,
    // so we'll just return a single joke for demonstration
    url := fmt.Sprintf("%s/jokes/random?category=%s", c.baseURL, category)
    
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

    return []*domain.Joke{{
        ID:       chuckResp.ID,
        Value:    chuckResp.Value,
        Category: chuckResp.Category,
    }}, nil
}

func (c *Client) GetCategories(ctx context.Context) ([]string, error) {
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

    var catResp CategoryResponse
    if err := json.Unmarshal(body, &catResp); err != nil {
        return nil, fmt.Errorf("joke client get categories: unmarshal: %w", err)
    }

    return catResp.Categories, nil
}
```

---

## Making HTTP Requests

For custom HTTP clients without SDKs, here's the pattern:

```go
type HTTPClient struct {
    client  *http.Client
    baseURL string
}

func (c *HTTPClient) Get(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
    url := c.baseURL + endpoint
    if len(params) > 0 {
        q := url.Values(params).Encode()
        url = url + "?" + q
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Accept", "application/json")

    resp, err := c.client.Do(req)
    if resp != nil {
        defer resp.Body.Close()
    }
    if err != nil {
        return nil, fmt.Errorf("execute request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %s", resp.Status)
    }

    return io.ReadAll(resp.Body)
}

func (c *HTTPClient) Post(ctx context.Context, endpoint string, body interface{}) ([]byte, error) {
    jsonBody, err := json.Marshal(body)
    if err != nil {
        return nil, fmt.Errorf("marshal body: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, bytes.NewReader(jsonBody))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if resp != nil {
        defer resp.Body.Close()
    }
    if err != nil {
        return nil, fmt.Errorf("execute request: %w", err)
    }

    return io.ReadAll(resp.Body)
}
```

---

## Wire-Up in cmd/run.go

```go
// cmd/run.go
func run() error {
    // Create external client
    jokeClient := joke.New()

    // Test connection with Ping
    if err := jokeClient.Ping(ctx); err != nil {
        return fmt.Errorf("joke client ping failed: %w", err)
    }

    // Get a random joke
    j, err := jokeClient.GetRandomJoke(ctx)
    if err != nil {
        return fmt.Errorf("get random joke: %w", err)
    }

    logger.Info("Got joke", "joke", j.Value)

    // Create service with adapter
    jokeService := services.NewJokeService(jokeClient, logger)

    // Start server with handlers
    return httpServer.Serve(handler.Routes())
}
```

---

## Error Handling

### Typed Errors

```go
var (
    ErrRateLimited = errors.New("rate limited")
    ErrNotFound    = errors.New("resource not found")
    ErrAPIError    = errors.New("API error")
)

func (c *Client) GetJoke(ctx context.Context, id string) (*Joke, error) {
    resp, err := c.doRequest(ctx, "GET", "/jokes/"+id, nil)
    if err != nil {
        return nil, err
    }

    switch resp.StatusCode {
    case http.StatusNotFound:
        return nil, ErrNotFound
    case http.StatusTooManyRequests:
        return nil, ErrRateLimited
    case http.StatusInternalServerError:
        return nil, ErrAPIError
    }
    // ...
}
```

### Retry Logic

```go
func (c *Client) DoWithRetry(ctx context.Context, fn func() error) error {
    maxRetries := 3
    delay := time.Second

    for i := 0; i < maxRetries; i++ {
        if err := fn(); err != nil {
            if errors.Is(err, ErrRateLimited) {
                time.Sleep(delay)
                delay *= 2
                continue
            }
            return err
        }
        return nil
    }
    return fmt.Errorf("max retries exceeded")
}
```

---

## Testing

### Unit Test with Mock

```go
package joke

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockJokeProvider struct {
    mock.Mock
}

func (m *MockJokeProvider) Ping(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockJokeProvider) GetRandomJoke(ctx context.Context) (*domain.Joke, error) {
    args := m.Called(ctx)
    return args.Get(0).(*domain.Joke), args.Error(1)
}

func TestClient_GetRandomJoke(t *testing.T) {
    // This would test the actual implementation
    // by making real HTTP calls (integration test)
}

func TestJokeService_GetRandom(t *testing.T) {
    // Arrange
    mockProvider := new(MockJokeProvider)
    mockProvider.On("GetRandomJoke", mock.Anything).Return(&domain.Joke{
        ID:       "test-123",
        Value:    "Chuck Norris can divide by zero.",
        Category: "dev",
    }, nil)

    service := NewJokeService(mockProvider, nil)

    // Act
    joke, err := service.GetRandom(context.Background())

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Chuck Norris can divide by zero.", joke.Value)
    mockProvider.AssertExpectations(t)
}
```

### Integration Test

```go
// +build integration

package joke

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestIntegration_Client_GetRandomJoke(t *testing.T) {
    client := New()

    joke, err := client.GetRandomJoke(context.Background())
    require.NoError(t, err)
    require.NotEmpty(t, joke.ID)
    require.NotEmpty(t, joke.Value)
}

func TestIntegration_Client_Ping(t *testing.T) {
    client := New()

    err := client.Ping(context.Background())
    require.NoError(t, err)
}
```

---

## Best Practices

| Practice | Description |
|----------|-------------|
| **Compile-time interface check** | Add `var _ ports.JokeProvider = (*Client)(nil)` |
| **Context first** | All methods take `context.Context` as first argument |
| **Timeout configuration** | Set HTTP client timeouts (30s is typical) |
| **Error wrapping** | Wrap errors with operation context |
| **Retry logic** | Implement retry for transient failures |
| **Health checks** | Implement `Ping()` for connection verification |
| **Configuration** | Use config struct, don't hardcode values |
| **Idiomatic naming** | Package name + type: `joke.Client`, not `JokeClient` |