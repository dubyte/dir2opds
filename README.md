# dir2opds - serve books from a directory

 dir2opds inspects the given folder and acts as an OPDS 1.1–compliant server.

## Overview

There are good options for serving books using OPDS. Calibre is a popular
choice, but if you have a headless server, installing Calibre might not be
the best option.

That's where calibre2opds comes in. However, if you have a large number of
books and don't want to create a Calibre library, dir2opds can help you set up an OPDS server from a directory tree with one simple recommendation for optimal client compatibility:

- **Organize by levels:** A folder should ideally contain either only subfolders (for navigation) or only book files (for acquisition).

## Changelog

- [Changelog](CHANGELOG.md)

## Installation

```bash
go install github.com/dubyte/dir2opds@latest
```

## Usage

```bash
Usage of dir2opds:
  -calibre
        Hide files stored by calibre.
  -debug
        If it is set it will log the requests.
  -dir string
        A directory with books. (default "./books")
  -hide-dot-files
        Hide files that start with a dot.
  -host string
        The server will listen on this host. (default "0.0.0.0")
  -no-cache
        Add response headers to prevent the client from caching.
  -port string
        The server will listen on this port. (default "8080")
```

## Tested on

### Moon+ Reader

### Cantook

Tested on iPhone with the Cantook app.

<https://apps.apple.com/us/app/cantook-by-aldiko/id1476410111>

### KYBook 3

It works with [KyBook 3 Ebook Reader](https://apps.apple.com/us/app/kybook-3-ebook-reader/id1348198785) if access to Local Network is enabled in settings.  

To enable access, go to Settings -> Apps -> KyBook 3 -> Local Network (checked).

It seems that KyBook is so old that it does not trigger the access prompt on iOS, so it has to be configured manually.

## More information

- <http://opds-spec.org>

## Binary release

- <https://github.com/dubyte/dir2opds/releases>

### Raspberry Pi deployment using binary release

```bash
cd && mkdir dir2opds && cd dir2opds

# get the binary (replace v1.4.0 with the release that matches your system)
wget https://github.com/dubyte/dir2opds/releases/download/v1.4.0/dir2opds_1.4.0_linux_armv7.tar.gz

tar xvf dir2opds_1.4.0_linux_armv7.tar.gz

sudo touch /etc/systemd/system/dir2opds.service

# Paste the content below but remember to pass the full path of your books in -dir
sudo nano /etc/systemd/system/dir2opds.service

sudo systemctl enable dir2opds.service

sudo systemctl start dir2opds.service
```

/etc/systemd/system/dir2opds.service

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

## Other Installation methods

There are installation scripts and configuration files for FreeBSD, Illumos, and Linux in the [files/](files/) directory.

## Building from source

You can build the project for multiple platforms using the provided `Makefile`:

```bash
make build-all
```

This will generate binaries in the `bin/` directory for various operating systems and architectures.

## Rootless Container with podman

```sh
# build image
podman build -t localhost/dir2opds .
# prepare Books directory
mkdir /data/Books
chown -R $USER:$USER /data/Books
# run built image
podman run --name dir2opds --rm --userns=keep-id --mount type=bind,src=/data/Books,dst=/books,Z --publish 8008:8080 -i -t localhost/dir2opds /dir2opds -debug
```

Where:

- `/data/Books` is the path to the directory containing your books.

Test from host with

```sh
curl http://localhost:8008
```

## How to contribute

- [Contributing](CONTRIBUTING.md)

## Special thanks

- @clach04: for testing and reporting missing content type for comics.
- @masked-owl: for reporting the security issue about HTTP traversal.
