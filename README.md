# Go Gin Starter

A batteries-included starter kit for building server-rendered web applications with [Gin](https://github.com/gin-gonic/gin). It pairs HTML templates, session helpers, and structured logging with a small sample feature so you can focus on shipping your own functionality. Gin was chosen because it's docs are extensive and excellent; nearly anything you want to do has a concrete implementation example provided for you.

## Features
- Gin router with example controllers, HTML templates, and integration tests.
- Cookie-backed sessions via `gin-contrib/sessions` plus flash-message helpers.
- Config loader with env interpolation, `.env` support, and a `generateconfig` CLI.
- Structured JSON logging powered by `log/slog`.
- DataSourceOrchestration (DSO) pattern for dependency injection without globals.
- Make targets and Dockerfile for reproducible builds, tests, and packaging.

## Prerequisites
- Go 1.24+ (module-aware toolchain)
- Make (optional but recommended for the provided targets)
- Docker (optional, required only for container builds)

## Quick Start
1. Copy or clone this starter into your project directory and `cd` into it.
2. Update the module path in `go.mod` if you plan to publish under your own import path (for example `module github.com/my-org/my-app`).
3. Build the project:

```bash
make build
# or: go build -o ./bin/go-gin-starter .
```

4. Generate a configuration file with fresh secure-cookie keys. The generator refuses to overwrite an existing file:

```bash
./bin/go-gin-starter generateconfig
# or: go run . generateconfig
```

5. Review and edit `config.toml` as needed. The application searches for the file in:
   - The directory that contains the executable
   - Your current working directory
   - `.` (so running `make run` from the project root just works)

6. Start the server:

```bash
make run
# or: go run .
```

7. Visit <http://localhost:8080> to see the sample pages (`/`, `/books`, `/books/new`, `/books/:id`).

`config.toml` ships with sensible defaults for development: template caching is off so edits reload automatically, SSL is disabled, and the generated cookie keys are ready for local use. For production, turn on `cache_templates`, disable `ssl_disabled`, and supply secure keys via environment variables instead of committing them to source control.

## Make Targets

`make help` lists every target. Notable commands:

- `make build` – compile the binary to `./bin/go-gin-starter`
- `make run` – rebuild and execute the binary
- `make test` – run `go test ./...` (tests cover routes and configuration helpers)
- `make fmt` / `make vet` – format sources and run `go vet`
- `make tidy` – sync module dependencies
- `make deps` – update dependencies and verify the module graph
- `make clean` – remove the `bin` directory
- `make docker-build` / `make docker-run` – build and run the scratch-based container image
- `make cross-compile` – produce Linux and Darwin binaries for common architectures
- `make print-go-version` – echo the Go version declared in `go.mod`

## Configuration

- `config.toml` powers runtime settings; comments in the generated file explain every option.
- `${VAR_NAME}` syntax interpolates environment variables. `.env` values are loaded automatically thanks to `godotenv/autoload`.
- `regenerate_secure_keys = true` prints new signing/encryption keys and exits so you can copy them into your secrets store.
- Logging uses `log/slog`. Set `log_level` (1-5) and optionally `log_file` to persist logs to disk.
- `cache_templates` controls whether templates are read from disk (great for development) or served from the embedded assets (recommended for production).
- `ssl_*` settings enable TLS via `gin.Engine.RunTLS`.
- `secure_cookie_max_age` governs the session lifetime. Cookies are marked secure when TLS is enabled.

## Sessions and Flash Helpers

`session.go` configures a cookie-backed store (`gin-contrib/sessions`) using the secure keys from your config. Helpers include:

- `getUser(session)` – returns the stored `SessionUser`
- `getFlashes(session)` – retrieves one-time flash messages
- `addFlash(message, session)` – queues a message for the next request
- `SessionUser` exposes convenience methods such as `IsAdmin()` and `SessionIsValid()` for role and expiry checks.

## Code Conventions & Patterns

This starter follows a few battle-tested conventions to keep projects consistent and easy to navigate.

### 1. Route Naming Convention

Route handlers follow `route_<Resource>_<Action>()` so related endpoints group together in file listings. Examples:

- GET handlers: `route_Books_Index()` lists all books; `route_Books_Show()` renders a single book.
- POST handlers append `_POST` to distinguish them (for example, `route_Books_Create_POST()`).
- Nested resources use additional segments (for example, `route_Users_Profile_Update_POST()`).

### 2. Controller File Naming

HTTP handlers live in files prefixed with `ctr_` (for example `ctr_books.go`, `ctr_root.go`). Supporting code—config, logging, utilities—stays in plainly named files, making controllers easy to spot.

### 3. Error Wrapping Style

Wrap errors with the calling function name: `fmt.Errorf("pkgName.functionName(): %w", err)`. The prefix pinpoints where the error originated while `%w` preserves the chain for `errors.Is()` and `errors.Unwrap()`.

### 4. DataSourceOrchestration (DSO) Pattern

Shared dependencies (config, logger, forthcoming database connections) live in the `DataSourceOrchestration` struct. Middleware injects the DSO into each request (`c.MustGet("dso")`) so handlers can grab what they need without relying on globals or ever-growing function signatures.

### 5. HTTP Handler Error Flow

Handlers follow a four-step pattern when things go wrong:
1. Log the detailed error with structured fields (`logger.Error("book not found", "id", id)`).
2. Add a user-friendly flash message (`addFlash("Book not found", session)`).
3. Redirect to a safe location (`c.Redirect(http.StatusSeeOther, "/books")`).
4. `return` immediately so execution stops.

This keeps logs actionable while surfacing friendly feedback to users.

### 6. Structured Logging with slog

`SetupLogger` configures `log/slog` with JSON output. Access it through the DSO (`logger := dso.Logger`) and always log key/value pairs. `log_level` in `config.toml` controls verbosity; optional `log_file` redirects output to disk. Refer to https://go.dev/blog/slog for use and best practices.

### 7. Session Helpers

`session.go` registers `SessionUser` with the cookie store and offers helpers for flashes, role checks, and session validation. Store and retrieve your authenticated user via `gin.AuthUserKey`.

## Adding Routes

1. Create or duplicate a template folder under `templates/`. Layouts live in `templates/layouts`.
2. Add a handler that returns a `gin.HandlerFunc`. Pattern:

```go
func route_Books_Index() gin.HandlerFunc {
	return func(c *gin.Context) {
		dso := c.MustGet("dso").(*DataSourceOrchestration)
		logger := dso.Logger
		session := sessions.Default(c)

		user := getUser(session)
		flashes := getFlashes(session)

		books := []Book{ /* fetch from your store */ }

		logger.Debug("serving books index", "count", len(books))

		c.HTML(http.StatusOK, "books/index", struct {
			AppConfig   *AppConfig
			SessionUser *SessionUser
			Flash       []string
			Books       []Book
		}{
			dso.AppConfig,
			&user,
			flashes,
			books,
		})
	}
}
```

3. Register the handler in `routes.go`, for example `r.GET("/books", route_Books_Index())`.
4. Add any forms or JSON handlers as needed (use `c.JSON` for APIs).
5. Add tests in `ctr_<resource>_test.go`. The existing books tests demonstrate how to spin up a Gin engine with the DSO middleware, templates, and sessions.

## Templates

- Templates are stored under `templates/<resource>/` and rendered with layout wrappers defined in `templates/layouts`.
- Custom helpers are defined in `server.go` (`customTmplFuncMap`) and include formatting helpers, Markdown rendering, title casing, and more.
- When `cache_templates = false`, Gin reloads templates from disk on each request using the directory that contains your `config.toml`. When enabled, the starter serves the embedded templates bundled with the binary.

## Docker

The included `Dockerfile` builds a static binary and produces a minimal scratch-based image:

```bash
make docker-build
make docker-run
```

The build stage passes the Go version declared in `go.mod` (`make docker-build` handles this for you). Runtime images copy the binary, config, templates, and static assets into `/app` and run as an unprivileged user on port 8080.

## Testing

Run the tests with:

```bash
make test
# or: go test ./...
```

`ctr_books_test.go` covers the example handlers (including validation and redirects), while `config_test.go` verifies environment-variable expansion logic. Add similar tests as you extend the application.

## Next Steps

- Replace the mock book data with your own persistence layer (add a database handle to the DSO).
- Expand the templates and routes to match your product requirements.
- Integrate authentication by populating `SessionUser` through your identity provider.
- Configure CI to run `make test` and, optionally, `make docker-build` for release pipelines.

