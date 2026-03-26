# dir2opds

[![Go Reference](https://pkg.go.dev/badge/github.com/dubyte/dir2opds.svg)](https://pkg.go.dev/github.com/dubyte/dir2opds)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Releases](https://img.shields.io/github/v/release/dubyte/dir2opds?include_prereleases&label=release)](https://github.com/dubyte/dir2opds/releases)

**Serve an OPDS 1.1–compliant book catalog from a directory.** No database, no Calibre—just point dir2opds at a folder and use any OPDS client to browse and download your books.

---

## Table of contents

- [What is OPDS?](#what-is-opds)
- [Features](#features)
- [Quick start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Compatible clients](#compatible-clients)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## What is OPDS?

[OPDS](http://opds-spec.org) (Open Publication Distribution System) is a standard for cataloging and distributing digital publications. OPDS clients (ebook readers, apps) can discover, browse, and download books from an OPDS server. dir2opds turns a plain directory tree into such a server.

## Features

- **OPDS 1.1 compliant** — Works with standard ebook readers and OPDS clients
- **No database** — Reads directly from your filesystem; no Calibre or extra setup
- **Flexible layout** — Organize by folders; optional metadata from EPUB/PDF
- **Search** — Optional filename search (OpenSearch)
- **Covers** — Optional `cover.jpg` / `folder.jpg` as catalog covers
- **Structured Logging** — Uses `log/slog` for JSON (default) or text logging
- **Multiple formats** — EPUB, PDF, MOBI, AZW3, and more via configurable MIME types
- **Lightweight** — Single binary; suitable for headless servers and containers

## Quick start

Using Docker (replace `v1.7.0` with the [latest release](https://github.com/dubyte/dir2opds/releases) if desired):

```bash
docker run -d -p 8080:8080 -v ./books:/books --name dir2opds ghcr.io/dubyte/dir2opds:v1.7.0
```

Then open `http://localhost:8080` in an OPDS client or browser.

**Using Go:**

```bash
go install github.com/dubyte/dir2opds@latest
dir2opds -dir /path/to/books -port 8080
```

**Tip:** For best client compatibility, use folders that contain either only subfolders (navigation) or only book files (acquisition), not mixed.

---

## Installation

### Go install

```bash
go install github.com/dubyte/dir2opds@latest
```

For other installation methods (Docker, Podman, pre-built binaries, etc.), see [INSTALLATION.md](INSTALLATION.md).

---

## Usage

Default: serve `./books` on `http://0.0.0.0:8080`.

```bash
dir2opds -dir /path/to/books -port 8080
```

### Options

| Flag | Description |
|------|-------------|
| `-calibre` | Hide files stored by Calibre |
| `-debug` | Log requests |
| `-dir` | Directory with books (default: `./books`) |
| `-extract-metadata` | Extract title/author from EPUB and PDF |
| `-hide-dot-files` | Hide files whose names start with a dot |
| `-host` | Listen address (default: `0.0.0.0`) |
| `-log-format` | Log format: `json` (default), `text` |
| `-mime-map` | Custom MIME types, e.g. `.mobi:application/x-mobipocket-ebook,.azw3:application/vnd.amazon.ebook` |
| `-no-cache` | Add response headers to disable client caching |
| `-port` | Listen port (default: `8080`) |
| `-search` | Enable basic filename search |
| `-show-covers` | Use `cover.jpg` or `folder.jpg` as catalog covers |
| `-sort` | Sort entries: `name`, `date`, or `size` (default: `name`) |
| `-url` | The base URL used for absolute links in the feed (e.g., `https://opds.example.com`) |

---

## Compatible clients

These OPDS clients have been tested with dir2opds:

| Client | Platform | Notes |
|--------|----------|--------|
| [Moon+ Reader](https://play.google.com/store/apps/details?id=com.flyersoft.moonreader) | Android | Tested |
| [Cantook](https://apps.apple.com/us/app/cantook-by-aldiko/id1476410111) | iPhone | Tested |
| [KYBook 3](https://apps.apple.com/us/app/kybook-3-ebook-reader/id1348198785) | iOS | Enable **Settings → Apps → KyBook 3 → Local Network**. Older app may not show the prompt; enable manually. |

---

## Documentation

- [Installation](INSTALLATION.md)
- [Changelog](CHANGELOG.md)
- [OPDS specification](http://opds-spec.org)
- [Contributing](CONTRIBUTING.md)

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) for license agreements, development setup, and pull request process.

---

## License

This project is licensed under the **GNU General Public License v3.0**. See [LICENSE](LICENSE) for the full text.

---

## Acknowledgments

- **@clach04** — Testing and reporting missing content type for comics.
- **@masked-owl** — Reporting the HTTP path traversal security issue.
- **@mufeedali** — Update to push image to ghcr.io.
- **@kulak** — Add podman support.
- **@thenktor** - init files and Makefile improvements.
- **@rockavoldy** — For the docker command example.
