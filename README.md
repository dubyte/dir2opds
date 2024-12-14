# dir2opds - serve books from a directory

 dir2opds inspects the given folder and serves an OPDS 1.1 compliant server.

## Overview

There are good options for serving books using OPDS. Calibre is a popular
choice, but if you have a headless server, installing Calibre might not be
the best option.

That's where calibre2opds comes in. However, if you have a large number of
books and don't want to create a Calibre library, dir2opds can help you
set up an OPDS server from a directory with one condition:

- A folder should contain either only folders or only files.

## Change log

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
        Hide files that starts with dot.
  -host string
        The server will listen in this host. (default "0.0.0.0")
  -no-cache
        adds reponse headers to avoid client from caching.
  -port string
        The server will listen in this port. (default "8080")
```

## Tested on

### Moon+ reader

### Cantook

It works on Cantook reader for me on iPhone.

<https://apps.apple.com/us/app/cantook-by-aldiko/id1476410111>

It would be nice if this was stated in README.

### KYBook 3

It works with [KyBook 3 Ebook Reader](https://apps.apple.com/us/app/kybook-3-ebook-reader/id1348198785) if access to Local Network is enabled in settings.  

To enable access go to Settings -> Apps -> KyBook 3 -> Local Network (checked).

It seems that KyBook is so old, that it does not trigger access prompt from iOS, so it has to be configured manually.

## More information

- <http://opds-spec.org>

## Binary release

- <https://github.com/dubyte/dir2opds/releases>

### Raspberry pi deployment using binary release

```bash
cd && mkdir dir2opds && cd dir2opds

# get the binary
wget https://github.com/dubyte/dir2opds/releases/download/v1.1.0/dir2opds_1.1.0_linux_armv7.tar.gz

tar xvf dir2opds_1.1.0_linux_armv7.tar.gz

sudo touch /etc/systemd/system/dir2opds.service

# Paste the content below but rember to pass the fullpath of your books in -dir
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

where

- `/data/Books` is a path to directory containing books.

Test from host with

```sh
curl http://localhost:8008
```

## How to contribute

- [Contributing](CONTRIBUTING.md)

## Special thanks

- @clach04: for testing and report missing content type for comics.
- @masked-owl: for reporting security issue about http transversal.
