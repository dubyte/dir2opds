# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
