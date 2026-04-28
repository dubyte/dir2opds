# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.9.0] - 2026-04-27

### Added

- **No Pagination option** — `-no-pagination` flag to disable pagination and show all entries in a single feed.

## [1.8.0] - 2026-03-28

### Added

- **Pagination** — `-page-size` flag to control entries per page (default: 50, max: 200). Feeds include `first`, `previous`, `next`, `last` navigation links.
- **ETag/Last-Modified caching** — `-enable-cache` flag for conditional requests (304 Not Modified) to reduce bandwidth.
- **Gzip compression** — `-gzip` flag to compress responses and reduce bandwidth.
- **Health endpoint** — `/health` endpoint returning `{"status":"ok"}` for monitoring and load balancers.
- **EPUB cover extraction** — `-extract-metadata` now extracts cover images from EPUB files and serves them via `/cover?file=` endpoint.
- **Recommended configuration** — README section with suggested flags for best experience.

### Changed

- `-extract-metadata` now extracts covers from EPUB files in addition to title/author.
- Removed separate `-extract-covers` flag (consolidated into `-extract-metadata`).

## [1.7.0] - 2026-03-26

### Added

- `-url` flag to specify a base URL for absolute links in the feed (useful for reverse proxies).
- `GEMINI.md` with project overview and development guidelines.
- Support for absolute URLs in OpenSearch and feed generation.
- Tests for Base URL functionality.
- Migrated to `log/slog` for structured logging.
- Added `-log-format` flag (options: `json`, `text`) to configure log output.
- Included `base_url` in default log attributes for better traceability.

### Changed

- Default log format is now JSON.
- Cleaned up logs to avoid exposing internal filesystem paths (`fPath`).
- Updated `errorHandler` to use structured logging.

## [1.6.0] - 2026-03-08

### Added

- `-sort` flag to sort entries by name, date, or size.
- `-show-covers` flag to automatically detect and show `cover.jpg` or `folder.jpg` as catalog covers.
- `-search` flag to enable a basic recursive filename search (OpenSearch compliant).
- `-mime-map` flag to allow custom MIME-type mapping for file extensions.
- `-extract-metadata` flag to extract Title and Author from EPUB and PDF files.

### Changed

- Catalog model expanded to include metadata for sorting, cover support and book metadata.
- Refactor `Handler` and `makeFeed` to use the new `Catalog` model for better separation of concerns.

## [1.5.0] - 2026-03-08

### Added

- `Catalog` and `CatalogEntry` models to represent the directory structure.
- `Scan` method in `OPDS` service to decouple filesystem discovery from feed generation.

### Changed

- Refactor `Handler` and `makeFeed` to use the new `Catalog` model for better separation of concerns.
- Update documentation with more installation options and building from source instructions.

## [1.4.0] - 2026-03-07

### Security

- Harden path traversal check: require trusted root + separator so paths like `/home/bookkeeping` are not allowed when root is `/home/books`.

### Fixed

- Handler no longer double-writes when path does not exist (return nil after sending 404).
- Avoid nil dereference in `getPathType` when `os.Stat` fails.
- Handle and log `ReadDir` error in `makeFeed` instead of ignoring it.
- Use `log` instead of `fmt.Println` in `verifyPath` for consistent logging.

### Changed

- Comment and error message typos (traversal, canonical, Navigation).
- Test: use `filepath.Join` for expected path and rename `Test_absoluteCanonicalPath`.

## [1.3.2] - 2025-11-01

### Dependencies

- Module go version changed to 1.25

## [1.3.1] - 2025-09-01

### Dependencies

- Bump gopkg.in/yaml.v3 from 3.0.0 to 3.0.1

## [1.3.0] - 2024-12-10

### Added

- pod container creation by @kulak

### Changed

- fail to start will log error on stderr in this case when book dir was not found.

## [1.2.0] - 2024-06-18

### Added

- no-cache argument can be passed to add Cache-Control and expires headers to let the client know we dont want to use cache.

### Changed

- make file allow to build for multiple goarch and goos.

## [1.1.0] - 2024-06-14

### Changed

- when calibre option is passed now also excludes .calnotes .caltrash

### Added

- hide-dot-files can be passed to ignere .files

## [1.0.6] - 2024-03-12

### Security

- Fix HTTP directory traversal vulnerability.
- Dir parameter will be used as trusted root.

### Changed

- Trusted root is obtained from dir parameter after get the absolut and canonical path.
- Module go version changed to 1.22

### Fixed

- Fix README.md error marked by linter

### Added

- Adding Make file

## [1.0.5] - 2024-01-30

### Changed

- Fix go releaser config
- name of binaries will result differently

## [1.0.4] - 2024-01-30

### Changed

- Update goreleaser deprecation fields

## [1.0.3] - 2024-01-30

### Changed

- Update some function calls to not deprecated versions.

## [1.0.2] - 2022-12-16

### Changed

- go releaser ignore windows arm64 build

## [1.0.1] - 2022-12-16

### Added

- pdf mime type

## [1.0.0] - 2022-04-30

### Changed

- logic that ignores favicon.ico was removed.

### Added

- when a request is recieved from a path that does not exists it logs and return 404
- calibre program argument was added to hide some files like metadata.opf and cover.jpg

### Fixed

- A panic when a path does not exists, not it logs an error instead.

## [0.1.2] - 2022-04-30

### Changed

- fix favicon.ico typo
- remove app arguments about author is not necesary and is not intuitive

## [0.1.1] - 2022-04-26

### Changed

- Returns 404 when the path contains favicon.ico

## [0.1.0] - 2021-06-10

### Changed

- time is calculated only once
- updated the mimetype returned for the navigation and acquisition xmls

## [0.0.11] - 2021-06-05

### Changed

- return to filepath as the best way to handle paths for different platforms.

## [0.0.10] - 2021-05-06

### Added

- Unittests

### Changed

- structure of the project changed but public endpoint remains backward compatible.

## [0.0.9] - 2021-05-06

### Changed

- using actions in github for building binaries only when tagging.

## [0.0.8] - 2021-05-06

### Changed

- using actions in github for building binaries only when tagging.

## [0.0.7] - 2021-05-06

### Changed

- using actions in github for testing

## [0.0.6] - 2021-05-06

### Changed

- using actions in github for testing

## [0.0.5] - 2020-05-06

### removed

- usage of filepath package

## [0.0.4] - 2020-05-15

### Added

- debug flag (idea from @clash04)
- comic support thanks to @clash04
- fix hrel that didnt allow the download. (found the issue @clash04)

## [0.0.3] - 2019-03-10

### Added

- A change log was added.
- Added a message when the server started in the stadin

### Changed

- fix rel and type for acquisition
- In the code change where the parameters are defined.
- Changed serveFeedauthor parameter for author.
- Adding host parameter.
- Start using go modules.
- Fixing typo in file extension gif
- The MIME-type for FB2 changed to text/fb2+xml

### Removed

- vendor folder.

## [0.0.2] - 2017-05-11

### Changed

- Using builders to generate the xml.
- Adding binaries in the release section.

## [0.0.1] - 2017-03-24

### Added

- first version of dir2opds was relased.
