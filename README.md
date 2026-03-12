# dir2opds - serve books from a directory

dir2opds inspects the given folder and acts as an OPDS 1.1–compliant server.

## Overview

There are good options for serving books using OPDS. Calibre is a popular
choice, but if you have a headless server, installing Calibre might not be
the best option.

That's where dir2opds comes in. If you have a large number of
books and don't want to create a Calibre library, dir2opds can help you set up an OPDS server from a directory tree with one simple recommendation for optimal client compatibility:

- **Organize by levels:** A folder should ideally contain either only subfolders (for navigation) or only book files (for acquisition).

## Installation & deployment

Choose one of the following ways to run dir2opds.

### Go install

```bash
go install github.com/dubyte/dir2opds@latest
```

### Pre-built binaries

- [Releases](https://github.com/dubyte/dir2opds/releases) — Linux, macOS, Windows, and other platforms.

### Docker

Pull the image:

```bash
docker pull ghcr.io/dubyte/dir2opds:v1.6.0
```

Run the container (mount your books directory; the OPDS catalog is served on port 8080):

```bash
docker run \
  -d \
  -m 256MB \
  --restart always \
  -p 8080:8080 \
  -v ./books:/books \
  --name dir2opds \
  ghcr.io/dubyte/dir2opds:v1.6.0
```

Thanks to [rockavoldy](https://hub.docker.com/u/rockavoldy) for the command.

### Podman

You can use the same image as Docker, or build locally.

**Option 1 — use the pre-built image:**

```bash
podman pull ghcr.io/dubyte/dir2opds:v1.6.0
```

**Option 2 — build from source:**

```bash
podman build -t localhost/dir2opds .
```

Run the container (mount your books directory; OPDS catalog on port 8080):

```bash
podman run \
  -d \
  -m 256MB \
  --restart always \
  -p 8080:8080 \
  -v ./books:/books \
  --name dir2opds \
  ghcr.io/dubyte/dir2opds:v1.6.0
```

For a **rootless** setup (e.g. non-root user, SELinux), use a bind mount with the `Z` option and keep your user namespace:

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
  ghcr.io/dubyte/dir2opds:v1.6.0
```

Add `-debug` to the image command for request logging, e.g. `... ghcr.io/dubyte/dir2opds:v1.6.0 /dir2opds -debug`.

### Raspberry Pi (binary + systemd)

```bash
cd && mkdir dir2opds && cd dir2opds

# get the binary (replace v1.6.0 with the release that matches your system)
wget https://github.com/dubyte/dir2opds/releases/download/v1.6.0/dir2opds_1.6.0_linux_armv7.tar.gz

tar xvf dir2opds_1.6.0_linux_armv7.tar.gz

sudo touch /etc/systemd/system/dir2opds.service

# Paste the content below but remember to pass the full path of your books in -dir
sudo nano /etc/systemd/system/dir2opds.service

sudo systemctl enable dir2opds.service

sudo systemctl start dir2opds.service
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

ExecStart=/home/pi/dir2opds/dir2opds -dir <FULL PATH OF BOOKS FOLDER> -port 8080

[Install]
WantedBy=multi-user.target
```

### Other platforms

Installation scripts and configuration files for FreeBSD, Illumos, and Linux are in the [files/](files/) directory.

### Build from source

Build for multiple platforms using the provided `Makefile`:

```bash
make build-all
```

Binaries are generated in the `bin/` directory.

## Usage

Run the server (default: serve `./books` on `http://0.0.0.0:8080`):

```bash
dir2opds -dir /path/to/books -port 8080
```

All flags:

```bash
Usage of dir2opds:
  -calibre
        Hide files stored by calibre.
  -debug
        If it is set it will log the requests.
  -dir string
        A directory with books. (default "./books")
  -extract-metadata
        Extract metadata (title, author) from EPUB and PDF files.
  -hide-dot-files
        Hide files that start with a dot.
  -host string
        The server will listen on this host. (default "0.0.0.0")
  -mime-map string
        Custom mime types (e.g., '.mobi:application/x-mobipocket-ebook,.azw3:application/vnd.amazon.ebook')
  -no-cache
        Add response headers to prevent the client from caching.
  -port string
        The server will listen on this port. (default "8080")
  -search
        Enable basic filename search.
  -show-covers
        Show cover.jpg or folder.jpg as catalog cover.
  -sort string
        Sort entries by: name, date, size. (default "name")
```

## Compatible clients

The following OPDS clients have been tested with dir2opds:

### Moon+ Reader

Tested on Android.

### Cantook

Tested on iPhone with the [Cantook app](https://apps.apple.com/us/app/cantook-by-aldiko/id1476410111).

### KYBook 3

It works with [KyBook 3 Ebook Reader](https://apps.apple.com/us/app/kybook-3-ebook-reader/id1348198785) if access to Local Network is enabled in settings.  

To enable access, go to Settings -> Apps -> KyBook 3 -> Local Network (checked).

It seems that KyBook is so old that it does not trigger the access prompt on iOS, so it has to be configured manually.

## Documentation

- [Changelog](CHANGELOG.md)
- [OPDS specification](http://opds-spec.org)
- [Contributing](CONTRIBUTING.md)

## Special thanks

- @clach04: for testing and reporting missing content type for comics.
- @masked-owl: for reporting the security issue about HTTP traversal.
