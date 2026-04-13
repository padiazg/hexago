# URL Shortener with QR Code - Full Example

Complete walkthrough of building a URL shortener service with HexaGo. This guide demonstrates how to wire together domain entities, port interfaces, secondary adapters (database + external API), core services, and primary adapters.

---

## Step 1: Initialize Project

Create the project with all required features:

```shell
$ hexago init url-shortener \
  --module github.com/myuser/url-shortener \
  --framework chi \
  --with-docker \
  --with-migrations \
  --with-observability

📋 Project Configuration:
  Name:              url-shortener
  Module:            github.com/myuser/url-shortener
  Project Type:      http-server
  Framework:         chi
  Adapter Style:     primary-secondary
  Core Logic:        services
  Docker:            true
  Observability:     true
  Migrations:        true
  Workers:           false
  Example Code:      false

🚀 Generating project url-shortener...
📁 Creating directory structure...
📝 Generating files...
📦 Initializing go module...
go: creating new go.mod: module github.com/myuser/url-shortener
go: to add module requirements and sums:
   go mod tidy
📦 Adding dependencies...
🧹 Running go mod tidy...
✨ Formatting code...

✅ Project generated successfully!

📚 Next steps:
  cd url-shortener
  go run main.go run
```

This creates:

- `cmd/` - CLI commands (root.go, run.go)  
- `internal/core/` - Domain and services  
- `internal/adapters/` - Primary and secondary adapters  
- `migrations/` - Database migrations  
- `pkg/` - Reusable packages (logger, server, httpserver)  
- `main.go`, `Makefile`, `Dockerfile`, etc.  

> Make sure you move to the project folder to continue the next steps from there

```shell
cd url-shortener
```

---

## Step 2: Domain Layer + Ports

> **Note**: Running `hexago add domain entity` automatically generates BOTH the entity file AND the port interface file (in `internal/core/domain/urls/port.go`).

Create the domain entity and ports. This belongs in the core layer with no external dependencies.

```shell
$ hexago add domain entity URL \
  --fields "id:string,original_url:string,created_at:time.Time,click_count:int"

📦 Adding domain entity: URL
   Project: url-shortener

📝 Creating entity file: internal/core/domain/urls/urls.go
📝 Creating port file: internal/core/domain/urls/port.go
📝 Creating test file: internal/core/domain/urls/urls_test.go

✅ Domain entity added successfully!

📝 Next steps:
  1. Add business logic methods to the entity
  2. Add validation rules
  3. Write tests for domain logic
```

then update the code to add custom errors and implement the validation code

```go
// internal/core/domain/urls/urls.go
package domain

import (
   "errors"
   "time"
)

var (
   ErrInvalidURL   = errors.New("invalid URL format")
   ErrURLNotFound  = errors.New("URL not found")
   ErrCodeConflict = errors.New("short code already exists")
)

// URL represents a shortened URL entity.
// This is a domain entity with unique identity and business logic.
type URL struct {
   ID          string    // Short code, e.g., "abc123x"
   OriginalURL string    // Original long URL
   QR          string    // QR code
   CreatedAt   time.Time // Creation timestamp
   ClickCount  int       // Number of times accessed
}

type Config struct {
   OriginalURL string // Original long URL
}

// NewURL creates a new URL with validation
func NewURL(config *Config) (*URL, error) {
   entity := &URL{
      OriginalURL: config.OriginalURL,
   }

   if err := entity.Validate(); err != nil {
      return nil, err
   }

   return entity, nil
}

// Validate checks that the URL has valid data.
func (u *URL) Validate() error {
   if u.OriginalURL == "" {
      return ErrInvalidURL
   }
   // Basic URL validation
   if len(u.OriginalURL) < 5 {
      return ErrInvalidURL
   }
   return nil
}
```

update the generated port interfaces with custom methods

```go
// internal/core/domain/urls/port.go
package urls

import "context"

// URLStore is the port for URL persistence.
// Implemented by the SQLite secondary adapter.
type URLRepository interface {
   // Save creates a new short URL. Returns ErrCodeConflict if code already exists.
   Save(ctx context.Context, url *URL) error
   // GetByID retrieves a URL by its short code.
   GetByID(ctx context.Context, code string) (*URL, error)
   // SaveQR saves the qr.
   SaveQR(ctx context.Context, code, qr string) error
   // IncrementClick increments the click counter for a URL.
   IncrementClick(ctx context.Context, code string) error
   // List returns all URLs, limited by count.
   List(ctx context.Context, limit int) ([]*URL, error)
   // Delete removes a URL by its short code.
   Delete(ctx context.Context, code string) error
}
// QRGenerator is the port for QR code generation.
// Implemented by the external API secondary adapter.
type QRGenerator interface {
   // Generate creates a QR code for the given URL.
   // Returns base64-encoded PNG image.
   Generate(ctx context.Context, url string) (string, error)
}
```

---

## Step 3: Secondary Adapter - SQLite Repository

Implement the `URLRepository` port using SQLite.

First, create the database migration:

```shell
$ hexago add migration create_urls

📦 Adding migration: create_urls
   Project: url-shortener
   Type: sql

📝 Creating migration files:
   UP:   migrations/000001_create_urls.up.sql
   DOWN: migrations/000001_create_urls.down.sql
📝 Creating migration manager: internal/infrastructure/database/migrator.go

ℹ️  Add these commands to your Makefile:

migrate-up: ## Run database migrations
   @migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Rollback last migration
   @migrate -path migrations -database "$(DB_URL)" down 1

migrate-version: ## Show current migration version
   @migrate -path migrations -database "$(DB_URL)" version

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
   @migrate -path migrations -database "$(DB_URL)" force $(VERSION)

# Add DB_URL to your environment or Makefile:
# DB_URL=postgresql://user:password@localhost:5432/dbname?sslmode=disable

✅ Migration added successfully!

📝 Files created:
   - migrations/000001_create_urls.up.sql
   - migrations/000001_create_urls.down.sql

📝 Next steps:
  1. Edit the .up.sql file with your schema changes
  2. Edit the .down.sql file to reverse those changes
  3. Run migrations:
     make migrate-up
  4. To rollback:
     make migrate-down
```

This creates:

- `migrations/000001_create_urls.up.sql`  
- `migrations/000001_create_urls.down.sql`  

Update the migration SQL:

```sql
-- migrations/000001_create_urls.up.sql
CREATE TABLE IF NOT EXISTS urls (
    id TEXT PRIMARY KEY,
    original_url TEXT NOT NULL,
    qr TEXT NOT NULL DEFAULT "",
    created_at TEXT NOT NULL,
    click_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at DESC);
```

```sql
-- migrations/000001_create_urls.down.sql
DROP TABLE IF EXISTS urls;
```

Update the migrator to use SQLite
> By the time I write this the migrator generates code for Postgres

 ```go
// internal/infrastructure/database/migrator.go
package database

import (
   "database/sql"
   "fmt"

   "github.com/golang-migrate/migrate/v4"
   sqlitemig "github.com/golang-migrate/migrate/v4/database/sqlite"
   _ "github.com/golang-migrate/migrate/v4/source/file"
   "github.com/padiazg/url-shortener/pkg/logger"
   _ "modernc.org/sqlite"
)
...
// newMigration creates a migrate instance
func (m *Migrator) newMigration() (*migrate.Migrate, error) {
   driver, err := sqlitemig.WithInstance(m.db, &sqlitemig.Config{}) // change this
   if err != nil {
      return nil, err
   }

   // TODO: Update database name if not using postgres
   return migrate.NewWithDatabaseInstance(
      "file://migrations",
      "sqlite",               // change this
      driver,
   )
```

Create the SQLite repository:

```shell
$ hexago add adapter secondary database URLRepository \
  --entity URL

📦 Adding secondary adapter: URLRepository (database)
   Project: url-shortener
   Adapter dir: secondary

📝 Creating adapter file: internal/adapters/secondary/database/urls/urls.go
📝 Creating test file: internal/adapters/secondary/database/urls/urls_test.go

✅ Secondary adapter added successfully!

📝 Next steps:
  1. Implement the port interface methods
  2. Add database queries or external API calls
  3. Wire up dependencies in the DI container
```

Now implement the repository:

```go
// internal/adapters/secondary/database/urls/urls.go
package urls

import (
   "context"
   "database/sql"
   "fmt"
   "time"

   domain "github.com/padiazg/url-shortener/internal/core/domain/urls"
   _ "modernc.org/sqlite"
)

// URLRepository implements urlsDomain.URLRepository using PostgreSQL.
type URLRepository struct {
   db *sql.DB
}

// compile-time check that URLRepository satisfies the port.
var _ domain.URLRepository = (*URLRepository)(nil)

// New creates a new URLRepository.
func New(db *sql.DB) *URLRepository {
   return &URLRepository{db: db}
}

// Open opens a SQLite database at the given path.
func Open(path string) (*sql.DB, error) {
   if path == "" {
		return nil, fmt.Errorf("open database: must provide a path")
	}
   
   db, err := sql.Open("sqlite", path)
   if err != nil {
      return nil, fmt.Errorf("open database: %w", err)
   }
   db.SetMaxOpenConns(1) // SQLite single-writer
   return db, nil
}

// --- URLs ---

func (r *URLRepository) Save(ctx context.Context, url *domain.URL) error {
   const q = `INSERT INTO urls (id, original_url, created_at, click_count) VALUES (?, ?, ?, ?)`
   _, err := r.db.ExecContext(ctx, q,
      url.ID,
      url.OriginalURL,
      url.CreatedAt.Format(time.RFC3339),
      url.ClickCount,
   )
   if err != nil {
      return fmt.Errorf("repository save URL: %w", err)
   }
   return nil
}

func (r *URLRepository) GetByID(ctx context.Context, code string) (*domain.URL, error) {
   const q = `SELECT id, original_url, qr, created_at, click_count FROM urls WHERE id = ?`
   row := r.db.QueryRowContext(ctx, q, code)

   var (
      url          domain.URL
      createdAtStr string
   )

   err := row.Scan(&url.ID, &url.OriginalURL, &url.QR, &createdAtStr, &url.ClickCount)
   if err != nil {
      if err == sql.ErrNoRows {
         return nil, domain.ErrURLNotFound
      }
      return nil, fmt.Errorf("repository get URL: %w", err)
   }

   url.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

   return &url, nil
}

func (r *URLRepository) SaveQR(ctx context.Context, code, qr string) error {
   const q = `UPDATE urls SET qr = ? WHERE id = ?`
   res, err := r.db.ExecContext(ctx, q, qr, code)
   if err != nil {
      return fmt.Errorf("repository update qr: %w", err)
   }

   rows, err := res.RowsAffected()
   if err != nil {
      return fmt.Errorf("repository rows affected: %w", err)
   }
   if rows == 0 {
      return domain.ErrURLNotFound
   }

   return nil
}

func (r *URLRepository) IncrementClick(ctx context.Context, code string) error {
   const q = `UPDATE urls SET click_count = click_count + 1 WHERE id = ?`
   res, err := r.db.ExecContext(ctx, q, code)
   if err != nil {
      return fmt.Errorf("repository increment click: %w", err)
   }

   rows, err := res.RowsAffected()
   if err != nil {
      return fmt.Errorf("repository rows affected: %w", err)
   }
   if rows == 0 {
      return domain.ErrURLNotFound
   }

   return nil
}

func (r *URLRepository) List(ctx context.Context, limit int) ([]*domain.URL, error) {
   const q = `SELECT id, original_url, created_at, click_count FROM urls ORDER BY created_at DESC LIMIT ?`
   rows, err := r.db.QueryContext(ctx, q, limit)
   if err != nil {
      return nil, fmt.Errorf("repository list URLs: %w", err)
   }
   defer rows.Close()

   var urls []*domain.URL
   for rows.Next() {
      var (
         url          domain.URL
         createdAtStr string
      )

      if err := rows.Scan(&url.ID, &url.OriginalURL, &createdAtStr, &url.ClickCount); err != nil {
         return nil, fmt.Errorf("repository scan URL: %w", err)
      }

      url.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
      urls = append(urls, &url)
   }
   return urls, rows.Err()
}

func (r *URLRepository) Delete(ctx context.Context, code string) error {
   const q = `DELETE FROM urls WHERE id = ?`
   res, err := r.db.ExecContext(ctx, q, code)
   if err != nil {
      return fmt.Errorf("repository delete URL: %w", err)
   }

   rows, err := res.RowsAffected()
   if err != nil {
      return fmt.Errorf("repository delete rows: %w", err)
   }
   if rows == 0 {
      return domain.ErrURLNotFound
   }
   return nil
}
```

Now we need to wire up the migrator to a command

```go
// cmd/migrate.go
package cmd

import (
   "fmt"

   urlsRepository "github.com/padiazg/url-shortener/internal/adapters/secondary/database/urls"
   "github.com/padiazg/url-shortener/internal/infrastructure/database"
   "github.com/padiazg/url-shortener/pkg/logger"
   "github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
   Use:   "migrate",
   Short: "Database migration management",
   Long:  `Run, rollback or inspect database migrations.`,
}

var migrateUpCmd = &cobra.Command{
   Use:   "up",
   Short: "Apply all pending migrations",
   RunE: func(cmd *cobra.Command, args []string) error {
      return withMigrator(func(m *database.Migrator, _ logger.Logger) error {
         return m.Up()
      })
   },
}

var migrateDownCmd = &cobra.Command{
   Use:   "down",
   Short: "Roll back the last applied migration",
   RunE: func(cmd *cobra.Command, args []string) error {
      return withMigrator(func(m *database.Migrator, _ logger.Logger) error {
         return m.Down()
      })
   },
}

var migrateVersionCmd = &cobra.Command{
   Use:   "version",
   Short: "Show the current migration version",
   RunE: func(cmd *cobra.Command, args []string) error {
      return withMigrator(func(m *database.Migrator, log logger.Logger) error {
         version, dirty, err := m.Version()
         if err != nil {
            return err
         }
         dirtyFlag := ""
         if dirty {
            dirtyFlag = " (dirty)"
         }
         log.Info("Current migration version: %d%s\n", version, dirtyFlag)
         return nil
      })
   },
}

// withMigrator opens the DB, creates a Migrator and calls fn, then closes the DB.
func withMigrator(fn func(*database.Migrator, logger.Logger) error) error {
   cfg := GetConfig()
   log := logger.New(&logger.Config{
      Level:  cfg.LogLevel,
      Format: cfg.LogFormat,
   })

   db, err := urlsRepository.Open(cfg.DBPath)
   if err != nil {
      return fmt.Errorf("opening database: %w", err)
   }
   defer db.Close()

   if err := db.Ping(); err != nil {
      return fmt.Errorf("connecting to database: %w", err)
   }

   return fn(database.NewMigrator(&database.MigratorConfig{
      DB:     db,
      Logger: log,
   }), log)
}

func init() {
   migrateCmd.AddCommand(migrateUpCmd)
   migrateCmd.AddCommand(migrateDownCmd)
   migrateCmd.AddCommand(migrateVersionCmd)
   rootCmd.AddCommand(migrateCmd)
}

```

---

## Step 4: Secondary Adapter - QR Code API Client

Implement the `QRGenerator` port using an external API.

```shell
$ hexago add adapter secondary external QRClient

📦 Adding secondary adapter: QRClient (external)
   Project: url-shortener
   Adapter dir: secondary

📝 Creating adapter file: internal/adapters/secondary/external/q_r_client.go
📝 Creating test file: internal/adapters/secondary/external/q_r_client_test.go

✅ Secondary adapter added successfully!

📝 Next steps:
  1. Implement the port interface methods
  2. Add database queries or external API calls
  3. Wire up dependencies in the DI container
```

Now implement the QR client using qrcoder.co.uk (free, no API key needed):

```go
// internal/adapters/secondary/external/q_r_client.go
package external

import (
   "context"
   "encoding/base64"
   "fmt"
   "io"
   "net/http"
   "time"

   domain "github.com/padiazg/url-shortener/internal/core/domain/urls"
)

// Client implements ports.QRGenerator using qrcoder.co.uk API.
type Client struct {
   client *http.Client
}

var _ domain.QRGenerator = (*Client)(nil)

// New creates a new QR code client.
func New() *Client {
   return &Client{
      client: &http.Client{Timeout: 30 * time.Second},
   }
}

// Generate creates a QR code for the given URL.
// Returns base64-encoded PNG image.
func (c *Client) Generate(ctx context.Context, url string) (string, error) {
   // Using qrcoder.co.uk API - free, no API key needed
   // Format: https://qrtag.net/api/qr_4.png?url=<url>
   qrURL := fmt.Sprintf("https://qrtag.net/api/qr_4.png?url=%s", url)

   req, err := http.NewRequestWithContext(ctx, http.MethodGet, qrURL, nil)
   if err != nil {
      return "", fmt.Errorf("qr client: create request: %w", err)
   }

   resp, err := c.client.Do(req)
   if err != nil {
      return "", fmt.Errorf("qr client: execute request: %w", err)
   }
   defer resp.Body.Close()

   if resp.StatusCode != http.StatusOK {
      return "", fmt.Errorf("qr client: status %d", resp.StatusCode)
   }

   // Read the image data
   imgData, err := io.ReadAll(resp.Body)
   if err != nil {
      return "", fmt.Errorf("qr client: read body: %w", err)
   }

   // Encode to base64
   base64Str := base64.StdEncoding.EncodeToString(imgData)
   return base64Str, nil
}
```

---

## Step 5: Core Service

Create the core service that uses the ports defined in the domain layer  

```shell
$ hexago add service URLShortener \                      
  --entity URL \
  --description "URL shortening service"

📦 Adding service: URLShortener
   Project: url-shortener
   Module: github.com/padiazg/url-shortener
   Logic dir: services

📝 Creating service file: internal/core/services/urls/urls.go
📝 Creating test file: internal/core/services/urls/urls_test.go
📝 Updating services aggregator: internal/core/services/services.go

✅ Service added successfully!

📝 Next steps:
  1. Implement the business logic in the Execute method
  2. Add any required dependencies to the constructor
  3. Write tests in the generated test file
```

Now implement the service.

```go
// internal/core/services/urls/urls.go
package urls

import (
   "context"
   "crypto/rand"
   "encoding/base64"
   "fmt"
   "time"

   domain "github.com/padiazg/url-shortener/internal/core/domain/urls"
)

// URLWithQR is the response including QR code data.
type URLWithQR struct {
   OriginalURL  string `json:"original_url"`
   ShortURL     string `json:"short_url"`
   QRCodeBase64 string `json:"qr_code_base64"`
   ClickCount   int    `json:"click_count"`
   CreatedAt    string `json:"created_at"`
}

// Service handles URL shortening business logic.
type Service struct {
   repository domain.URLRepository
   qrClient   domain.QRGenerator
   baseURL    string // e.g., "http://localhost:8080"
}

// Config holds service configuration.
type Config struct {
   BaseURL    string
   Repository domain.URLRepository
   QRClient   domain.QRGenerator
}

// New creates a new URL shortener service.
func New(cfg *Config) *Service {
   return &Service{
      repository: cfg.Repository,
      qrClient:   cfg.QRClient,
      baseURL:    cfg.BaseURL,
   }
}

// Shorten creates a new short URL from the original URL.
func (s *Service) Shorten(ctx context.Context, originalURL string) (*domain.URL, error) {
   // Validate input
   if originalURL == "" {
      return nil, domain.ErrInvalidURL
   }

   // Generate unique short code
   code, err := generateShortCode()
   if err != nil {
      return nil, fmt.Errorf("generate short code: %w", err)
   }

   // Create domain entity
   url := &domain.URL{
      ID:          code,
      OriginalURL: originalURL,
      CreatedAt:   time.Now().UTC(),
      ClickCount:  0,
   }

   // Validate domain entity
   if err := url.Validate(); err != nil {
      return nil, fmt.Errorf("validate URL: %w", err)
   }

   // Persist
   if err := s.repository.Save(ctx, url); err != nil {
      return nil, fmt.Errorf("save URL: %w", err)
   }

   return url, nil
}

// GetOriginal retrieves the original URL for a given short code.
// Also increments the click counter.
func (s *Service) GetOriginal(ctx context.Context, shortCode string) (string, error) {
   url, err := s.repository.GetByID(ctx, shortCode)
   if err != nil {
      return "", err
   }

   // Increment click count
   if err := s.repository.IncrementClick(ctx, shortCode); err != nil {
      // Log error but don't fail the request
      // In production, use proper logging
      _ = err
   }

   return url.OriginalURL, nil
}

// GetWithQR returns the URL details including QR code.
func (s *Service) GetWithQR(ctx context.Context, shortCode string) (*URLWithQR, error) {
   url, err := s.repository.GetByID(ctx, shortCode)
   if err != nil {
      return nil, err
   }

   var qrBase64 string
   shortURL := s.baseURL + "/" + shortCode

   // check if QR was aleady generated
   if url.QR != "" {
      qrBase64 = url.QR
   } else {
      // Generate QR code for the short URL
      qrBase64, err = s.qrClient.Generate(ctx, shortURL)
      if err != nil {
         return nil, fmt.Errorf("generate QR: %w", err)
      }

      // Save generated QR to the record so next time it can be recovered
      if err = s.repository.SaveQR(ctx, shortCode, qrBase64); err != nil {
         return nil, fmt.Errorf("saving QR: %w", err)
      }
   }

   return &URLWithQR{
      OriginalURL:  url.OriginalURL,
      ShortURL:     shortURL,
      QRCodeBase64: qrBase64,
      ClickCount:   url.ClickCount,
      CreatedAt:    url.CreatedAt.Format(time.RFC3339),
   }, nil
}

// List returns all URLs.
func (s *Service) List(ctx context.Context, limit int) ([]*domain.URL, error) {
   if limit <= 0 {
      limit = 100
   }
   return s.repository.List(ctx, limit)
}

// Delete removes a URL by its short code.
func (s *Service) Delete(ctx context.Context, shortCode string) error {
   return s.repository.Delete(ctx, shortCode)
}

// generateShortCode generates a random 6-character alphanumeric code.
func generateShortCode() (string, error) {
   // 6 characters * 6 bits = 36 bits of entropy
   data := make([]byte, 6)
   if _, err := rand.Read(data); err != nil {
      return "", fmt.Errorf("read random: %w", err)
   }

   // Use base64 URL encoding (no padding)
   encoded := base64.URLEncoding.EncodeToString(data)
   // Take only first 6 characters
   return encoded[:6], nil
}
```

```go
// internal/core/services/services.go
package services

import (
   domain "github.com/padiazg/url-shortener/internal/core/domain/urls"
   svc "github.com/padiazg/url-shortener/internal/core/services/urls"
)

// Config holds the repository dependencies required to initialise entity-bound services.
type Config struct {
   Repository  domain.URLRepository
   QRGenerator domain.QRGenerator
}

// Services aggregates all domain services.
type Services struct {
   Urls *svc.Service
}

// New wires all services using the provided repository config.
func New(config *Config) *Services {
   return &Services{
      Urls: svc.New(&svc.Config{
         Repository: config.Repository,
         QRClient:   config.QRGenerator,
      }),
   }
}
```

---

## Step 6: Primary Adapter - HTTP Handler

Create the HTTP handler that exposes the service to the outside world.

```shell
hexago add adapter primary http URLShortenerHandler \
  --entity URL 
```

Now implement the handler:

```go
// internal/adapters/primary/http/urls/handlers.go
package urls

import (
   "encoding/json"
   "net/http"
   "strings"
   "time"

   "github.com/go-chi/chi/v5"
   domain "github.com/padiazg/url-shortener/internal/core/domain/urls"
)

type ErrorResponse struct {
   Error string `json:"error"`
}

// Request/Response types
type createRequest struct {
   URL string `json:"url"`
}

type createResponse struct {
   OriginalURL string `json:"original_url"`
   ShortURL    string `json:"short_url"`
   CreatedAt   string `json:"created_at"`
}

// Create handles POST /urls
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
   var req createRequest
   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      h.respondError(w, http.StatusBadRequest, "invalid JSON body")
      return
   }

   if req.URL == "" {
      h.respondError(w, http.StatusBadRequest, "url is required")
      return
   }

   url, err := h.manage.Shorten(r.Context(), req.URL)
   if err != nil {
      if err == domain.ErrInvalidURL {
         h.respondError(w, http.StatusBadRequest, err.Error())
         return
      }
      h.respondError(w, http.StatusInternalServerError, err.Error())
      return
   }

   h.respondJSON(w, http.StatusCreated, createResponse{
      OriginalURL: url.OriginalURL,
      ShortURL:    "/" + url.ID,
      CreatedAt:   url.CreatedAt.Format(time.RFC3339),
   })
}

// getByID handles GET /urls/{id}g
func (h *handler) getByID(w http.ResponseWriter, r *http.Request) {
   code := chi.URLParam(r, "id")

   result, err := h.manage.GetWithQR(r.Context(), code)
   if err != nil {
      if err == domain.ErrURLNotFound {
         h.respondError(w, http.StatusNotFound, "URL not found")
         return
      }
      h.respondError(w, http.StatusInternalServerError, err.Error())
      return
   }

   h.respondJSON(w, http.StatusOK, result)
}

// list handles GET /urls
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
   urls, err := h.manage.List(r.Context(), 100)
   if err != nil {
      h.respondError(w, http.StatusInternalServerError, err.Error())
      return
   }

   h.respondJSON(w, http.StatusOK, urls)
}

// delete DELETE /urls/{code} - Delete a URL
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
   code := chi.URLParam(r, "code")

   if err := h.manage.Delete(r.Context(), code); err != nil {
      if err == domain.ErrURLNotFound {
         h.respondError(w, http.StatusNotFound, "URL not found")
         return
      }
      h.respondError(w, http.StatusInternalServerError, err.Error())
      return
   }

   w.WriteHeader(http.StatusNoContent)
}

// redirect GET /{code} - Redirect to original URL
func (h *handler) redirect(w http.ResponseWriter, r *http.Request) {
   code := chi.URLParam(r, "code")

   originalURL, err := h.manage.GetOriginal(r.Context(), code)
   if err != nil {
      if err == domain.ErrURLNotFound {
         http.NotFound(w, r)
         return
      }
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }

   // Check if URL has protocol, add http:// if not
   if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
      originalURL = "http://" + originalURL
   }

   http.Redirect(w, r, originalURL, http.StatusFound)
}

// Helper methods
func (h *handler) respondJSON(w http.ResponseWriter, status int, data any) {
   w.Header().Set("Content-Type", "application/json")
   w.WriteHeader(status)
   json.NewEncoder(w).Encode(data)
}

func (h *handler) respondError(w http.ResponseWriter, status int, message string) {
   h.respondJSON(w, status, ErrorResponse{Error: message})
}
```

```go
// internal/adapters/primary/http/urls/u_r_l.go
package urls

import (
   "github.com/go-chi/chi/v5"
   "github.com/padiazg/url-shortener/internal/core/services"
   scv "github.com/padiazg/url-shortener/internal/core/services/urls"
   "github.com/padiazg/url-shortener/pkg/server"
)

// Config holds the dependencies for the URL HTTP handler.
type Config struct {
   Path     string
   Router   chi.Router
   Services *services.Services
}

type handler struct {
   *Config
   manage *scv.Service
}

// Compile-time interface check
var _ server.ServerHandler = (*handler)(nil)

// New creates a new URL HTTP handler and registers its routes.
func New(config *Config) *handler {
   return &handler{
      Config: config,
      manage: config.Services.Urls,
   }
}

// Configure registers the URL routes on the server.
func (h *handler) Configure(srv server.Server) {
   r := chi.NewRouter()

   r.Route("/", func(r chi.Router) {
      r.Get("/", h.list)
      r.Post("/", h.create)
      r.Get("/{id}", h.getByID)
      r.Delete("/{id}", h.delete)
   })

   h.Router.Mount(h.Path, r)

   // Redirect route (shorter path)
   h.Router.Get("/{code}", h.redirect)
}
```

Register handler

```go
// internal/adapters/primary/http/http.go
package http

import (
   "context"
...
   "github.com/padiazg/url-shortener/internal/adapters/primary/http/urls"
   httpsrv "github.com/padiazg/url-shortener/pkg/httpserver"
)

// New creates and wires the HTTP server with all registered route handlers.
// Framework-specific server mechanics live in pkg/httpserver.
func New(cfg *httpsrv.ServerConfig, services *services.Services) server.Server {
   srv := httpsrv.New(cfg).(*httpsrv.Server)

...

   srv.Router.Route("/api/v1", func(r chi.Router) {
      // Route-specific middlewares (apply only to /api/v1/* routes)
      // r.Use(chimiddleware.RequestID)           // inject X-Request-Id header
      // r.Use(chimiddleware.Logger)              // structured request logging
      // r.Use(chimiddleware.Recoverer)           // recover from panics
      // r.Use(yourauth.Middleware(services))     // JWT / session authorization

      // register the handlers
      srv.Use(urls.New(&urls.Config{
         Path:     "/urls",
         Router:   r,
         Services: services,
      }))
   })

   return srv

```

---

## Step 7: Config

For this example we will add only the database file path

```go
// internal/config/config.go
package config

import (
   "fmt"
   "time"

   "github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
   Server    ServerConfig
   LogLevel  string
   LogFormat string
   DBPath    string    // add this
}

...

func setDefaults() {
   // Server defaults
   viper.SetDefault("dbpath", "./url_shortener.db") // add this
...
}
```

---

## Step 8: Wire-Up in cmd/run.go

Now wire everything together in `cmd/run.go`:

```go
// cmd/run.go
/*
Copyright © 2026
*/
package cmd

import (
   "context"
   "fmt"
   "os"
   "os/signal"
   "syscall"

   "github.com/padiazg/url-shortener/internal/adapters/primary/http"
   urlsRepository "github.com/padiazg/url-shortener/internal/adapters/secondary/database/urls"
   qr "github.com/padiazg/url-shortener/internal/adapters/secondary/external"
   "github.com/padiazg/url-shortener/internal/core/services"
   httpsrv "github.com/padiazg/url-shortener/pkg/httpserver"
   "github.com/padiazg/url-shortener/pkg/logger"
   "github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
   Use:   "run",
   Short: "Start the url-shortener HTTP server",
   Long: `Start the url-shortener HTTP API server with graceful shutdown support.

The server will listen for SIGINT (Ctrl+C) and SIGTERM signals
and perform a graceful shutdown with a configurable timeout.`,
   RunE: func(cmd *cobra.Command, args []string) error {
      // ── Configuration ───────────────────────────────────────────────
      cfg := GetConfig()

      // Initialize logger from config
      log := logger.New(&logger.Config{
         Level:  cfg.LogLevel,
         Format: cfg.LogFormat,
      })

      log.Info("Starting url-shortener HTTP server...")

      // ── Database ──────────────────────────────────────────────────────
      db, err := urlsRepository.Open(cfg.DBPath)
      if err != nil {
         return fmt.Errorf("open database: %w", err)
      }
      defer db.Close()

      // ── Secondary Adapters ────────────────────────────────────────────
      urlsRepo := urlsRepository.New(db)
      qrClient := qr.New()

      // ── Services (core) ───────────────────────────────────────────────
      services := services.New(&services.Config{
         Repository:  urlsRepo,
         QRGenerator: qrClient,
      })

      // Create and configure the HTTP server.
      // Framework-specific server mechanics live in pkg/httpserver.
      // Route handler wiring (ping, health, metrics, ...) lives in
      // internal/adapters/primary/http/http.go.
      srv := http.New(&httpsrv.ServerConfig{
         Config: cfg,
         Logger: log,
      }, services)

      // Channel to capture OS signals
      sigChan := make(chan os.Signal, 1)
      signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

      // Channel for server errors
      errChan := make(chan error, 1)

      // Start server (non-blocking)
      srv.Run(errChan)

      // Wait for signal or fatal server error
      select {
      case sig := <-sigChan:
         log.Info("Received signal: %v, initiating graceful shutdown...", sig)

         // Create shutdown context with timeout
         shutdownCtx, shutdownCancel := context.WithTimeout(
            context.Background(),
            cfg.Server.ShutdownTimeout,
         )
         defer shutdownCancel()

         // Attempt graceful shutdown
         if err := srv.Stop(shutdownCtx); err != nil {
            log.Error("Server shutdown error: %v", err)
            return err
         }

         log.Info("Server stopped gracefully")
         return nil

      case err := <-errChan:
         log.Error("Server error: %v", err)
         return err
      }
   },
}

func init() {
   rootCmd.AddCommand(runCmd)
}
```

---

## Step 9: Validate Architecture

After all components are created, validate the hexagonal architecture:

```shell
$ hexago validate

🔍 Validating project: url-shortener
   Module: github.com/padiazg/url-shortener
   Adapter style: primary-secondary
   Core logic: services

📋 Validation Results:
✓ Domain directory exists
✓ Core logic directory exists
✓ Inbound adapters directory exists
✓ Outbound adapters directory exists
✓ Config directory exists
✓ Core domain has no external dependencies
✓ Services only depend on domain and ports
✓ Adapters follow dependency rules
✓ Using primary for inbound adapters
✓ Using secondary for outbound adapters
✓ Using services for business logic

📊 Summary:
   ✓ Passed: 11
   ⚠️  Warnings: 0
   ✗ Errors: 0

✅ Validation PASSED
```

This checks:

- ✓ Core domain has no external dependencies
- ✓ Services only depend on domain and ports
- ✓ Adapters don't import from other adapters
- ✓ Proper dependency direction (adapters → core)

---

## Step 10: Running the Application

### Build and Run

```shell
make build
```

### Run the migration

```shell
$ ./url-shortener migrate up

2026/04/08 19:43:52 [INFO] Running migrations...
2026/04/08 19:43:52 [INFO] Migrations completed successfully
```

### Run the server

```shell
$ ./url-shortener run

2026/04/08 18:52:06 [INFO] Starting url-shortener HTTP server...
2026/04/08 18:52:06 [INFO] Server listening on port 8080
```

### Create a Short URL

```shell
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/padiazg/hexago"}'
```

**Response:**

```json
{
  "original_url": "https://github.com/padiazg/hexago",
  "short_url": "/eOp1RA",
  "created_at": "2026-04-08T23:16:10Z"
}
```

### Get URL with QR Code

```shell
curl http://localhost:8080/api/v1/urls/eOp1RA
```

**Response:**

```json
{
  "original_url": "https://github.com/padiazg/hexago",
  "short_url": "/eOp1RA",
  "qr_code_base64": "iVBORw0KGgoAAAANSUhEUgAAAHwAAAB8AQMAAACR0Eb9AAAABlBMVEX///+AAAC098j+AAAACXBIWXMAAA7EAAAOxAGVKw4bAAAA9klEQVRIia2V0RHDMAhD2YD9t2QDVRJNev20Ep9bu++DAyTTqhdWA9DXzF4jwM1Vg+79FYDmbQShnQN+Bs9AoTH9AGijVNqv2lMAZbTr7voh+OrLFs2f4CdghVKDmulNBvoKNawwBBiY8Y7N8hywKvmFznXMBJQUZmksTF3KAFNSa6qwRwDaLepVHJUBltR2bdW+wwC0pSJjhg4aAMBekWl0RoACLZsaRU6A3TYO2T/7HQKNNCeGTewc1HZn1Kf7YR4CTwBorOES+xysZ/QCPZ0iYPN7FrDXnYPSvw112oGSAd8segZ2HjAngc6ADj3ldnIReGF9ANrUtaND7PMPAAAAAElFTkSuQmCC",
  "click_count": 0,
  "created_at": "2026-04-08T23:16:10Z"
}
```

### Redirect to Original URL

```shell
$ curl http://localhost:8080/api/v1/eOp1RA
<a href="https://github.com/padiazg/hexago">Found</a>.
# Returns 302 redirect to https://github.com/padiazg/hexago
```

### List All URLs

```shell
curl http://localhost:8080/api/v1/urls 
```

**Response:**

```json
[
  {
    "ID": "eOp1RA",
    "OriginalURL": "https://github.com/padiazg/hexago",
    "QR": "",
    "CreatedAt": "2026-04-08T23:16:10Z",
    "ClickCount": 2
  }
]
```

---

## Summary

This guide demonstrated how to build a complete URL shortener service using hexagonal architecture:

| Step | Layer | Component | HexaGo Command |
| --- | --- | --- | --- |
| 1 | - | Project initialization | `hexago init` |
| 2 | Domain + Ports | URL entity + port interfaces | `hexago add domain entity` |
| 3 | Secondary | SQLite Repository | `hexago add adapter secondary database` + `hexago add migration` |
| 4 | Secondary | QR API Client | `hexago add adapter secondary external` |
| 5 | Core | Service implementation | `hexago add service` |
| 6 | Primary | HTTP Handler | `hexago add adapter primary http` |
| 7 | - | Config | Manual in `internal/config/config.go` |
| 8 | - | Wire-up | Manual in `cmd/run.go` |
| 9 | - | Validation | `hexago validate` |

Key patterns demonstrated:

- **Dependency rule**: Adapters → Core (never the other way)  
- **Port interfaces**: Created by `hexago add domain entity`, implemented by adapters  
- **Compile-time checks**: `var _ ports.URLStore = (*URLRepository)(nil)`  
- **Context propagation**: All methods take `context.Context` as first argument  
- **Error wrapping**: Errors wrapped with operation context
