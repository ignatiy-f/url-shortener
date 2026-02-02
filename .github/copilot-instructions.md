# AI Coding Agent Instructions for URL Shortener

## Architecture Overview

This is a **Go HTTP service** that shortens URLs and stores them in SQLite. The architecture follows clean separation:

- **Entry point**: [cmd/url-shortener/main.go](cmd/url-shortener/main.go) - bootstraps config, logger, storage, and router
- **HTTP layer**: [internal/http-server/handlers/](internal/http-server/handlers/) - chi router with handlers and middleware
- **Storage layer**: [internal/storage/sqlite/sqlite.go](internal/storage/sqlite/sqlite.go) - SQLite database abstraction
- **Shared utilities**: [internal/lib/](internal/lib/) - logger helpers, response formatting, random generation

**Critical flow**: Request → chi router → save handler → storage interface → SQLite

## Key Architectural Patterns

### 1. Interface-Based Storage
The save handler depends on an `URLSaver` interface, NOT concrete storage:
```go
type URLSaver interface {
    SaveUrl(urlToSave string, alias string) (int64, error)
}
```
This enables **mock-based unit testing**. See [internal/http-server/handlers/url/save/save_test.go](internal/http-server/handlers/url/save/save_test.go) which uses generated mocks via `mockery`.

### 2. Structured Logging with slog
All logging uses Go's `log/slog` with contextual attributes:
- Use `slog.String()`, `slog.Int()` for structured fields
- Always log operation context: `const op = "package.function"`
- Error logging uses helper: `sl.Err(err)` from [internal/lib/logger/sl/sl.go](internal/lib/logger/sl/sl.go)
- Environment-specific handlers: pretty output for "local", JSON for "dev"/"prod"

### 3. Request/Response Pattern
Standard JSON response envelope via [internal/lib/api/response/response.go](internal/lib/api/response/response.go):
```go
type Response struct {
    Status string `json:"status"` // "ok" or "error"
    Error  string `json:"error,omitempty"`
}
```
Validation errors are converted to user-friendly messages, NOT raw validator errors.

### 4. Configuration Management
Uses `cleanenv` + YAML file. Config is loaded once at startup from `CONFIG_PATH` env var.
- File location: [config/local.yaml](config/local.yaml)
- Struct tags: `yaml:"name" env:"NAME" env-default:"value"`
- **Important**: All required fields must be explicitly set or have defaults; missing files cause fatal exit.

## Testing Conventions

1. **Mock generation**: Run `go generate` in [internal/http-server/handlers/url/save/](internal/http-server/handlers/url/save/):
   ```bash
   go generate ./...
   ```
   This uses `mockery` directive in code comments to auto-generate mocks.

2. **Table-driven tests**: [save_test.go](internal/http-server/handlers/url/save/save_test.go) uses test case tables with `name`, `alias`, `url`, `respError`, `mockError` fields.

3. **Mock assertion**: Use `mock.AssertExpectations(t)` to verify mock calls.

## Build & Run

- **Set config**: `$env:CONFIG_PATH = "config/local.yaml"` (PowerShell)
- **Run server**: `go run ./cmd/url-shortener`
- **Run tests**: `go test ./...`
- **Generate mocks**: `go generate ./...`

## Common Tasks

**Adding a handler**:
1. Create handler package under [internal/http-server/handlers/](internal/http-server/handlers/)
2. Define a `New()` function that returns `http.HandlerFunc`
3. Accept logger and dependency interfaces (not concrete types)
4. Use `render.DecodeJSON()` to parse requests, `render.JSON()` for responses
5. Register route in [main.go](cmd/url-shortener/main.go) under router setup

**Adding storage method**:
1. Define interface in handler package (e.g., `URLSaver`)
2. Implement in [internal/storage/sqlite/sqlite.go](internal/storage/sqlite/sqlite.go)
3. Add `//go:generate mockery` directive above interface
4. Run `go generate` to update mocks

**Logging errors**: Always use operation name + error context:
```go
const op = "storage.sqlite.SaveUrl"
return 0, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
```
