# AGENTS.md - dir2opds

This file provides guidelines for AI coding agents working on the dir2opds codebase.

## Build, Test, and Lint Commands

### Building
```bash
# Build for current platform (default target)
make build

# Build for all platforms (Darwin, FreeBSD, Linux, Windows, etc.)
make build-all

# Build with Go directly
go build .
```

### Testing
```bash
# Run all tests
go test ./...
go test -v ./...

# Run a single test
go test -v -run TestHandler ./internal/service/
go test -v -run TestScan ./internal/service/

# Run tests with coverage
go test -cover ./...
```

### Linting and Formatting
```bash
# Format code (uses go fmt)
make fmt

# Run go vet
make vet

# Full build pipeline (fmt -> vet -> build)
make
```

### Cleaning
```bash
make clean
```

## Code Style Guidelines

### General Go Conventions
- Follow standard Go formatting (`go fmt`)
- Pass `go vet` without warnings
- Use Go 1.25.3+ (see go.mod)
- Maximum line length: aim for readability, no strict limit

### Imports
- Group imports: stdlib first, then external packages, then internal
- Use blank imports only when necessary
- Example:
```go
import (
    "archive/zip"
    "encoding/xml"
    "log/slog"
    
    "github.com/lann/builder"
    "golang.org/x/tools/blog/atom"
    
    "github.com/dubyte/dir2opds/opds"
)
```

### Naming Conventions
- **Packages**: lowercase, single word (e.g., `service`, `opds`)
- **Exported**: PascalCase (e.g., `Handler`, `Scan`)
- **Unexported**: camelCase (e.g., `extractMetadata`, `sortEntries`)
- **Constants**: camelCase for unexported, PascalCase for exported
- **Test files**: `*_test.go` suffix, package name suffixed with `_test` for external tests
- **Interfaces**: Noun ending in "-er" (e.g., `IsDirer`)

### Types and Structs
- Use struct tags for XML/JSON marshaling
- Use `iota` for related constants
- Document public types with comments starting with the type name

### Error Handling
- Always check errors and handle appropriately
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Log errors using `slog.Error()` with structured attributes
- Return errors rather than swallowing them
- Example:
```go
if err != nil {
    slog.Error("operation failed", "error", err)
    return fmt.Errorf("doing operation: %w", err)
}
```

### Logging
- Use `log/slog` for all logging (structured logging)
- Use key-value pairs: `slog.Info("message", "key", value)`
- Prefer `slog.Debug` for verbose output
- Use JSON handler by default, text handler optional

### Builder Pattern
The project uses `github.com/lann/builder` for immutable builders:
- Define builder types (e.g., `type feedBuilder builder.Builder`)
- Each setter returns the builder for chaining
- End with `Build()` to get the final struct
- Export a singleton instance (e.g., `var FeedBuilder`)

### Testing Patterns
- Use table-driven tests with `map[string]struct{...}`
- Use `t.Run(name, func(t *testing.T){...})` for subtests
- Prefer `testify/assert` and `testify/require`
- Use `httptest` for HTTP handler tests
- Store test data in `testdata/` directories
- Mock time using injectable functions (e.g., `TimeNow`)

### HTTP Handlers
- Return errors from handlers instead of writing directly to ResponseWriter
- Use `errorHandler` wrapper for consistent error responses
- Verify paths to prevent directory traversal attacks
- Set appropriate Content-Type headers

### Security
- Always use `verifyPath()` to prevent path traversal
- Use `filepath.Clean()` and `filepath.EvalSymlinks()` for path sanitization
- Check that paths are within `TrustedRoot` before accessing filesystem

### Comments and Documentation
- Package comments should describe the package purpose
- Start with `// Package name ...`
- Public functions/methods must have documentation comments
- Comments should start with the name being documented

### Commit Messages
Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat: add new feature`
- `fix: resolve bug`
- `docs: update documentation`
- `test: add tests`
- `refactor: restructure code`

## Project Structure

```
.
├── main.go                    # Entry point, CLI flags
├── main_test.go               # Main package tests
├── internal/
│   └── service/
│       ├── service.go         # Core business logic
│       ├── service_test.go    # External tests
│       └── internal_test.go   # Internal tests
├── opds/                      # OPDS/Atom XML builders
│   ├── feed_builder.go
│   ├── entry_builder.go
│   ├── link_builder.go
│   ├── author_builder.go
│   └── text_builder.go
├── files/                     # Platform-specific files
├── Makefile                   # Build automation
└── go.mod                     # Go module definition
```

## Key Dependencies
- `github.com/lann/builder` - Immutable builders
- `github.com/stretchr/testify` - Testing utilities
- `golang.org/x/tools/blog/atom` - Atom feed structs
