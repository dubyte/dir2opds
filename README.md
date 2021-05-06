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

# change log
  - [Changelog](CHANGELOG.md)

# Installation
    go get -u github.com/dubyte/dir2opds

# Usage
    dir2opds -dir ./books -port 8080

# Tested on:
   - Moon+ reader

# More information
  - http://opds-spec.org

# binary release
  - https://github.com/dubyte/dir2opds/releases
