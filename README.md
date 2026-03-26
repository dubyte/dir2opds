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

Choose one of the following.

### Go install

```bash
go install github.com/dubyte/dir2opds@latest
```

### Pre-built binaries

Download binaries for Linux, macOS, Windows, and other platforms from the [Releases](https://github.com/dubyte/dir2opds/releases) page.

### Docker

```bash
docker pull ghcr.io/dubyte/dir2opds:v1.7.0
```

```bash
docker run \
  -d \
  -m 256MB \
  --restart always \
  -p 8080:8080 \
  -v ./books:/books \
  --name dir2opds \
  ghcr.io/dubyte/dir2opds:v1.7.0
```

Use a specific [release](https://github.com/dubyte/dir2opds/releases) tag instead of `v1.7.0` if needed. Thanks to [rockavoldy](https://hub.docker.com/u/rockavoldy) for the run example.

### Podman

**Pre-built image:**

```bash
podman pull ghcr.io/dubyte/dir2opds:v1.7.0
```

**Build from source:**

```bash
podman build -t localhost/dir2opds .
```

**Run** (OPDS on port 8080):

```bash
podman run \
  -d \
  -m 256MB \
  --restart always \
  -p 8080:8080 \
  -v ./books:/books \
  --name dir2opds \
  ghcr.io/dubyte/dir2opds:v1.7.0
```

**Rootless** (e.g. non-root user, SELinux): use a bind mount with the `Z` option and keep the user namespace:

```bash
mkdir -p /data/Books
podman run \
  -d \
  -m 256MB \
  --restart always \
  --userns=keep-id \
  -p 8080:8080 \
  -v /data/Books:/books:Z \
  --name dir2opds \
  ghcr.io/dubyte/dir2opds:v1.7.0
```

Add `-debug` for request logging, e.g. `... ghcr.io/dubyte/dir2opds:v1.7.0 /dir2opds -debug`.

### Raspberry Pi (binary + systemd)

```bash
cd && mkdir dir2opds && cd dir2opds
# Replace v1.7.0 and the asset name with the release that matches your system
wget https://github.com/dubyte/dir2opds/releases/download/v1.7.0/dir2opds_1.7.0_linux_armv7.tar.gz
tar xvf dir2opds_1.7.0_linux_armv7.tar.gz

sudo nano /etc/systemd/system/dir2opds.service
# Paste the unit below, then set the full path of your books folder in -dir

sudo systemctl enable --now dir2opds.service
```

`/etc/systemd/system/dir2opds.service`:

```ini
[Unit]
Description=dir2opds
Documentation=https://github.com/dubyte/dir2opds
After=network-online.target

[Service]
User=pi
Restart=on-failure
ExecStart=/home/pi/dir2opds/dir2opds -dir <FULL_PATH_TO_BOOKS> -port 8080

[Install]
WantedBy=multi-user.target
```

### Other platforms

Installation scripts and configs for FreeBSD, Illumos, and Linux are in the [files/](files/) directory.

### Build from source

Requires [Go 1.21+](https://go.dev/doc/install). Build for multiple platforms:

```bash
make build-all
```

Binaries are written to `bin/`.

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
