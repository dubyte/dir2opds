# GEMINI.md - dir2opds

## Project Overview

`dir2opds` is a lightweight, database-free OPDS 1.1-compliant book catalog server written in Go. It turns a local directory structure into an OPDS feed that can be consumed by e-book readers and apps (e.g., Moon+ Reader, Cantook).

### Key Features:
- **Zero Database**: Directly scans the filesystem to generate catalogs.
- **Metadata Extraction**: Optionally extracts title and author from EPUB and PDF files.
- **Flexible Sorting**: Supports sorting by name, modification date, or file size.
- **Search**: Built-in OpenSearch support for filename-based search.
- **Customizable**: Support for custom MIME types and hiding specific files (e.g., Calibre-generated files, dotfiles).
- **Secure**: Implements path traversal protection.

### Architecture:
- `main.go`: Handles CLI flags, configures the service, and starts the HTTP server.
- `internal/service/`: Contains the core logic for scanning directories, extracting metadata, and handling HTTP requests for feeds and files.
- `opds/`: Provides a set of fluent, immutable builders for generating Atom/OPDS XML feeds.
- `files/`: Contains platform-specific installation scripts and configuration files (FreeBSD, Illumos, Linux).

---

## Building and Running

### Prerequisites:
- Go 1.21 or later (as per `README.md`, though `go.mod` indicates a newer version).

### Key Commands:
- **Build**: `make build` (builds the binary for the current platform).
- **Build All Platforms**: `make build-all` (builds for Darwin, FreeBSD, Linux, Windows, etc.).
- **Run**: `./dir2opds -dir /path/to/books -port 8080`
- **Test**: `go test ./...`
- **Lint/Format**: `make fmt` and `make vet`.
- **Clean**: `make clean`.

---

## Development Conventions

### Coding Style:
- **Fluent Builders**: The project uses `github.com/lann/builder` to create immutable builders for OPDS/Atom XML elements (see `opds/` directory).
- **Service Pattern**: Core logic is encapsulated in the `service.OPDS` struct, which implements the HTTP handlers.
- **Standard Formatting**: Adhere to `go fmt` and `go vet` standards.

### Testing:
- Tests are located alongside source files (e.g., `main_test.go`, `internal/service/service_test.go`).
- Always run `go test ./...` before submitting changes.
- Test data is stored in `internal/service/testdata/`.

### Contributions:
- The project is licensed under **GPL v3**.
- Ensure that any new file types or MIME mappings are added to the `init()` function in `internal/service/service.go`.
- Follow the guidelines in `CONTRIBUTING.md` (if present and detailed).

---

## Key Files:
- `main.go`: Application entry point and CLI flag definitions.
- `internal/service/service.go`: Core directory scanning and feed generation logic.
- `opds/feed_builder.go`: Fluent API for building OPDS feeds.
- `Makefile`: Build automation for multiple platforms.
