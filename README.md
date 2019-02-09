# dir2opds - serve books from a directory
 dir2opds inspect the given folder and serve an opds 1.1 compliant server

# Overview
 There are good options to serve books using opds. Calibre is good for
 that but if your server is headless sometimes install calibre
 just for serve the books don't seems to be the best option.

 That is why calibre2opds exists, but if you have too many books and
 you don't want to create a calibre library.

 dir2opds could help you to have an opds server from a directory if you
 follow one rule:

 - A folder should have only folders or only files (aka books).

 It is ok to have sub-folders.

# Installation
    go get github.com/dubyte/dir2opds

# Usage
    dir2opds -dir ./books -port 8080

# Tested in:
   - Moon+ reader

# More information
  - http://opds-spec.org

# binary release
  - https://github.com/dubyte/dir2opds/releases
