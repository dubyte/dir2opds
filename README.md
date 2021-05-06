# dir2opds - serve books from a directory
 dir2opds inspects the given folder and serve an OPDS 1.1 compliant server.

# Overview
 There are good options to serve books using OPDS. Calibre is good for
 that but if your server is headless, install Calibre doesn't seem to
 be the best option.

 That is why calibre2opds exists, but if you have too many books and
 you don't want to create a Calibre library dir2opds could help you to
 have an OPDS server from a directory with one condition:

 - A folder should have only folders or only files.

# Change log
  - [Changelog](CHANGELOG.md)

# Installation
    go get -u github.com/dubyte/dir2opds

# Usage
    dir2opds -dir ./books -port 8080

# Tested on:
   - Moon+ reader

# More information
  - http://opds-spec.org

# Binary release
  - https://github.com/dubyte/dir2opds/releases


## Raspberry pi deployment using binary release
```bash
cd && mkdir dir2opds && cd dir2opds

# get the binary
wget https://github.com/dubyte/dir2opds/releases/download/v0.0.10/dir2opds_0.0.10_Linux_ARMv7.tar.gz

tar xvf dir2opds_0.0.9_Linux_ARMv7.tar.gz

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
