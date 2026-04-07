# Configuration

How to configure your HexaGo-generated application.

---

## Config File

The generated project uses a YAML config file named after the project:

| Project Name | Config File |
|--------------|-------------|
| `my-app` | `.my-app.yaml` |
| `trading-bot` | `.trading-bot.yaml` |

### Default Config Structure

```yaml
server:
  port: 8080
  readtimeout: 15s
  writetimeout: 15s
  shutdowntimeout: 30s

loglevel: info
logformat: json
```

### Extended Config (with features)

```yaml
server:
  port: 8080
  readtimeout: 15s
  writetimeout: 15s
  shutdowntimeout: 30s

database:
  url: postgres://localhost:5432/mydb
  maxopen: 25
  maxidle: 5

loglevel: info
logformat: json

workers:
  enabled: true
  concurrency: 5

observability:
  metrics:
    enabled: true
    port: 9090
```

---

## Environment Variables

All config values can be overridden with environment variables using the `PROJECT_PREFIX_` format:

| Config Path | Environment Variable |
|-------------|---------------------|
| `server.port` | `MY_APP_SERVER_PORT` |
| `server.readtimeout` | `MY_APP_SERVER_READTIMEOUT` |
| `loglevel` | `MY_APP_LOGLEVEL` |
| `database.url` | `MY_APP_DATABASE_URL` |

### Example

```bash
# Override port
export MY_APP_SERVER_PORT=9000

# Override log level
export MY_APP_LOGLEVEL=debug

# Override database
export MY_APP_DATABASE_URL=postgres://prod-server/mydb
```

---

## Config File Locations

Config is loaded in priority order (highest first):

1. **Environment variable**: `TRADING_BOT_CONFIG=/path/to/config.yaml`
2. **Current directory**: `./.trading-bot.yaml`
3. **Home directory**: `~/.trading-bot.yaml`
4. **Defaults**: Built-in values

The first found file is used.

---

## Config Structure

### Server Configuration

```yaml
server:
  port: 8080              # HTTP server port
  readtimeout: 15s        # Max time to read request
  writetimeout: 15s      # Max time to write response
  shutdowntimeout: 30s   # Graceful shutdown timeout
```

### Logging Configuration

```yaml
loglevel: debug          # debug, info, warn, error
logformat: json          # json, text
```

### Database Configuration

```yaml
database:
  url: postgres://user:pass@localhost:5432/db
  maxopen: 25            # Max open connections
  maxidle: 5             # Max idle connections
  maxlifetime: 5m        # Max connection lifetime
```

### Custom Configuration

Add your own config sections:

```yaml
myapp:
  apikey: ${MY_APP_APIKEY}  # Reference env var
  features:
    feature_a: true
    feature_b: false
```

Access in code:

```go
type Config struct {
    MyApp MyAppConfig `mapstructure:"myapp"`
}

type MyAppConfig struct {
    APIKey  string        `mapstructure:"apikey"`
    Features FeatureConfig `mapstructure:"features"`
}
```

---

## Environment-Specific Config

### Development

```yaml
# .my-app.yaml (in .gitignore)
server:
  port: 8080
loglevel: debug
logformat: text

database:
  url: postgres://localhost:5432/dev_db
```

### Production

```bash
# Set environment variables in production
export MY_APP_SERVER_PORT=80
export MY_APP_LOGLEVEL=warn
export MY_APP_DATABASE_URL=postgres://prod-host/prod_db
```

Or use a dedicated config file:

```bash
export MY_APP_CONFIG=/etc/myapp/config.yaml
```

---

## Configuration in Code

Access configuration in your services:

```go
type Service struct {
    cfg *config.Config
    store Store
}

func NewService(cfg *config.Config, store Store) *Service {
    return &Service{
        cfg: cfg,
        store: store,
    }
}

func (s *Service) DoSomething(ctx context.Context) error {
    timeout := s.cfg.Server.ReadTimeout
    // Use config values
}
```

---

## Validation

Config should be validated at startup:

```go
func loadConfig() (*Config, error) {
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    // Validate required fields
    if cfg.Server.Port == 0 {
        return nil, errors.New("server.port is required")
    }

    return &cfg, nil
}
```

---

## Best Practices

1. **Never commit secrets** — Use env vars for API keys, passwords
2. **Validate early** — Fail fast on invalid config
3. **Document defaults** — Add comments in config template
4. **Use env vars for CI** — Different configs for dev/staging/prod