# internal/service - OPDS Service Layer

Core business logic for OPDS catalog generation and HTTP handling.

## WHERE TO LOOK

| Task | File | Key Functions |
|------|------|----------------|
| Add new HTTP endpoint | `service.go` | `Handler()`, `SearchHandler()`, `OpenSearchHandler()` |
| Modify catalog generation | `service.go` | `Scan()`, `makeFeed()` |
| Add metadata extraction | `service.go` | `extractMetadata()`, `extractEpubMetadata()`, `extractPdfMetadata()` |
| Add middleware | `gzip.go` | `GzipMiddleware()` |
| Security/path validation | `service.go` | `verifyPath()`, `inTrustedRoot()` |
| Pagination logic | `service.go` | `parsePage()`, `pageSize()` |
| Caching/ETag | `service.go` | `etag()` |

## KEY STRUCTS

```go
type OPDS struct {
    TrustedRoot      string
    HideCalibreFiles bool
    HideDotFiles     bool
    NoCache          bool
    EnableCache      bool      // ETag/Last-Modified
    SortBy           string    // name, date, size
    ShowCovers       bool
    MimeMap          map[string]string
    EnableSearch     bool
    ExtractMetadata  bool
    BaseURL          string
    PageSize         int
}

type Catalog struct {
    ID       string
    Title    string
    Type     int           // pathTypeFile, pathTypeDirOfDirs, pathTypeDirOfFiles
    Entries  []CatalogEntry
    Cover    string
    Total    int           // Before pagination
    Page     int           // Current page (1-indexed)
    PageSize int
    ModTime  time.Time    // For ETag/Last-Modified
}
```

## CONVENTIONS

### Handler Pattern
- Handlers return `error` (don't write to ResponseWriter directly)
- `errorHandler()` wrapper converts errors to HTTP 500
- File requests use `http.ServeFile()`, directories return OPDS XML

### Path Security
- ALWAYS call `verifyPath()` before filesystem access
- `TrustedRoot` is canonicalized at startup (prevents symlink escapes)
- Path traversal tests in `internal_test.go`

### Pagination
- Default: 50 entries/page, max: 200
- Query param: `?page=N`
- OPDS links: `first`, `previous`, `next`, `last`

### Caching
- `-enable-cache` flag enables ETag + Last-Modified
- ETag: SHA-256 hash of (path + mtime + page)
- 304 Not Modified on match

## ANTI-PATTERNS

- DO NOT skip `verifyPath()` before filesystem operations
- DO NOT use `as any` or `@ts-ignore` equivalents
- DO NOT suppress errors with `_ =` unless intentional (e.g., `mime.AddExtensionType`)
- DO NOT write to ResponseWriter in handlers—return errors instead