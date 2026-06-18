[![01 - Test and Build](https://github.com/thomis/sensu-plugins-go/actions/workflows/01_test_and_build.yml/badge.svg)](https://github.com/thomis/sensu-plugins-go/actions/workflows/01_test_and_build.yml)
[![02 - Test, Build and Release](https://github.com/thomis/sensu-plugins-go/actions/workflows/02_test_build_and_release.yml/badge.svg)](https://github.com/thomis/sensu-plugins-go/actions/workflows/02_test_build_and_release.yml)
[![Latest Release](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.github.com%2Frepos%2Fthomis%2Fsensu-plugins-go%2Freleases%2Flatest&query=%24.tag_name&label=release)](https://github.com/thomis/sensu-plugins-go/releases/latest)

# sensu-plugins-go

A comprehensive collection of Sensu monitoring plugins and handlers written in Go. This repository has evolved through community contributions and continues as an independent project with extended functionality for modern infrastructure.

## Overview

This project provides production-ready monitoring checks and handlers for Sensu, covering a wide range of services and system metrics. All plugins are compiled as individual executables and distributed in a compressed tar archive for easy deployment.

### Fork History

This repository has evolved through several community forks:
- [hico-horiuchi/sensu-plugins-go](https://github.com/hico-horiuchi/sensu-plugins-go) (original)
- [portertech/sensu-plugins-go](https://github.com/portertech/sensu-plugins-go)
- [nixwiz/sensu-plugins-go](https://github.com/nixwiz/sensu-plugins-go)
- [thomis/sensu-plugins-go](https://github.com/thomis/sensu-plugins-go) (decoupled)

The thomis/sensu-plugins-go repository has been decoupled from its upstream and continues as an independent project with extended functionality for modern infrastructure. The project has been restructured to follow Go best practices, with each binary organized in its own subfolder under `cmd/` along with dedicated tests and documentation (work in progress).

## Components

| Category | Component | Description | Documentation |
|----------|-----------|-------------|---------------|
| **System Checks** | check-cpu | Monitor CPU usage and alert on high utilization | [README](cmd/check-cpu/README.md) |
| | check-disk | Check disk space usage and available capacity | [README](cmd/check-disk/README.md) |
| | check-memory | Monitor memory usage and swap utilization | [README](cmd/check-memory/README.md) |
| | check-process | Verify processes are running with configurable thresholds | [README](cmd/check-process/README.md) |
| | check-uptime | Monitor system uptime | [README](cmd/check-uptime/README.md) |
| **Network & Connectivity** | check-ping | ICMP ping check with packet loss and latency monitoring | [README](cmd/check-ping/README.md) |
| | check-http | HTTP/HTTPS endpoint monitoring with response validation | [README](cmd/check-http/README.md) |
| | check-http-json | JSON API monitoring with response parsing and validation | [README](cmd/check-http-json/README.md) |
| | check-certificate | SSL/TLS certificate expiration and validation | [README](cmd/check-certificate/README.md) |
| **Database Monitoring** | check-postgres | PostgreSQL connectivity and version check | [README](cmd/check-postgres/README.md) |
| | check-postgres-query | Run a custom PostgreSQL query/function that returns status and message | [README](cmd/check-postgres-query/README.md) |
| | check-mysql-ping | MySQL connectivity check | [README](cmd/check-mysql-ping/README.md) |
| | check-mysql-processes | Monitor MySQL process list and connections | [README](cmd/check-mysql-processes/README.md) |
| | check-oracle-ping | Oracle database connectivity | [README](cmd/check-oracle-ping/README.md) |
| | check-oracle-validity | Oracle database object validity checks | [README](cmd/check-oracle-validity/README.md) |
| | check-oracle-query | Run a custom Oracle query/procedure that returns status and message | [README](cmd/check-oracle-query/README.md) |
| | check-redis | Redis server monitoring | [README](cmd/check-redis/README.md) |
| **Application Services** | check-elasticsearch | Elasticsearch cluster health monitoring | [README](cmd/check-elasticsearch/README.md) |
| | check-nginx | Nginx status and performance metrics | [README](cmd/check-nginx/README.md) |
| | check-postfix | Postfix mail server monitoring | [README](cmd/check-postfix/README.md) |
| | check-postfix-queue | Monitor Postfix queue size | [README](cmd/check-postfix-queue/README.md) |
| | check-rabbitmq | RabbitMQ queue and cluster monitoring | [README](cmd/check-rabbitmq/README.md) |
| **Metrics Collection** | metrics-cpu | Collect CPU metrics in Graphite format | |
| | metrics-disk | Disk usage metrics collection | |
| | metrics-memory | Memory usage metrics | |
| | metrics-traffic | Network traffic metrics | |
| | metrics-snmp | SNMP metrics collection | |
| **Event Handlers** | handler-slack | Send alerts to Slack channels | |
| | handler-elasticsearch | Index events in Elasticsearch | |
| | handler-hubot | Send notifications to Hubot | |
| | handler-delete | Clean up stale check results | |

## Installation

Download the latest release from the [Releases](https://github.com/thomis/sensu-plugins-go/releases) page. The archive contains all checks and handlers as separate executables in a `bin/` directory.

```bash
# Extract the archive (creates a bin/ directory with all checks)
tar -xzf sensu-checks-go.linux.amd64.tar.gz

# Make executables accessible (if needed)
chmod +x bin/*

# Run a specific check
./bin/check-disk -w 80 -c 90

# Or add to PATH for easier access
export PATH=$PATH:$(pwd)/bin
check-disk -w 80 -c 90
```

## Development

### Build Process

The typical development workflow is:

1. **Write code and tests** - Implement features with corresponding test coverage
2. **Run `make`** - Default command that formats, lints, tests, and builds for your local platform
3. **Create pull request** - Submit changes for review
4. **Create release** - Create a release via the GitHub UI, triggering the "02 - Test, Build and Release" workflow

### Building from Source

```bash
# Default build - format, lint, test, and build for local platform
make

# Run tests only with coverage report
make test

# Build for specific platforms
make build_linux_amd64
make build_darwin_arm64

# See all available targets
make help
```

### Requirements
- Go 1.26.0 or later
- Oracle Instant Client SDK (for Oracle checks)
- Make

## System Requirements

### Supported Platforms
- **Linux**: AMD64, ARM64
- **macOS**: ARM64 (Apple Silicon)

### Linux Compatibility (glibc)

Since **release 2.59**, Linux release binaries are built for broad compatibility across distributions, including older glibc runtimes such as Rocky Linux 8 / RHEL 8 (glibc 2.28) and container platforms like AWS Fargate. (Earlier releases were built against a newer glibc and may fail on such runtimes.)

- **Most plugins are statically linked** (built with `CGO_ENABLED=0`) and have **no glibc dependency at all** — they run on any Linux distribution.
- **The Oracle plugins** (`check-oracle-ping`, `check-oracle-validity`) require CGO (`godror`) and are therefore dynamically linked. To keep them runnable on older runtimes, the Linux release is **built inside a Rocky Linux 8 (glibc 2.28) container**, so these binaries run on **glibc 2.28 and newer** (RHEL/Rocky/Alma 8+, Ubuntu 18.04+, Debian 10+, …).

Because glibc is backward compatible, a binary linked against an older glibc also runs on newer ones — so a single set of artifacts covers both old and current distributions.

> If you see an error like `version 'GLIBC_2.34' not found`, you are running a binary that was linked against a newer glibc than the host provides. Use the official release artifacts (built as described above) or rebuild on a matching/older glibc.

### Oracle Components
The Oracle-based checks (`check-oracle-ping`, `check-oracle-validity`) require Oracle Instant Client libraries to be installed on the system. These components may have platform restrictions based on Oracle's library availability.

### Dependencies
Most checks are self-contained, but some require:
- Network access for remote service checks (e.g., `check-http`, `check-ping`, `check-elasticsearch`)
- System permissions for metrics collection (e.g., `metrics-disk`, `metrics-memory` may need access to `/proc` or `/sys`)
- Database client libraries for database checks (e.g., Oracle Instant Client for `check-oracle-*`, PostgreSQL client for `check-postgres`)

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/thomis/sensu-plugins-go. This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Contributor Covenant](https://www.contributor-covenant.org/) code of conduct.

1. Fork it ( https://github.com/thomis/sensu-plugins-go/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## License

This project is derived from upstream repositories that did not include explicit licenses. Please contact the repository maintainer for licensing questions.
