# Installation

This document provides detailed instructions for installing `dir2opds` on various platforms.

## Pre-built binaries

Download binaries for Linux, macOS, Windows, and other platforms from the [Releases](https://github.com/dubyte/dir2opds/releases) page.

## Docker

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

## Podman

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

## Raspberry Pi (binary + systemd)

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

## Other platforms

Installation scripts and configs for FreeBSD, Illumos, and Linux are in the [files/](files/) directory.

## Build from source

Requires [Go 1.21+](https://go.dev/doc/install). Build for multiple platforms:

```bash
make build-all
```

Binaries are written to `bin/`.
