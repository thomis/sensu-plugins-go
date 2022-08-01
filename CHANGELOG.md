# Changelog

All notable changes will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [0.2.0] - 2022-07-29

### Added
- Use of GitHub actions for CI/CD
- Use of Dependabot to keep dependencies up to date
- Use of test and linter steps
- Support for 3 platforms (linux.amd64, darvin.arm64, darvin.amd64, but currently no oracle support for darwin.arm64)
- Use of Code Climate features
- Minor code changes based on go linter feedback

## [0.1.0] - 2021-10-02

With manual build pipeline.

### Added
- check-certificate
- check-http-json
- check-oracle-ping
- check-oracle-validity
- check-process
- check-uptime

## [0.0.1] - 2020-02-28

### Added

- Initial version based on fork from [nixwiz/sensu-plugins-go](https://github.com/nixwiz/sensu-plugins-go)
