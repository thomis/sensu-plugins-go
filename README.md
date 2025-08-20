[![01 - Test and Build](https://github.com/thomis/sensu-plugins-go/actions/workflows/01_test_and_build.yml/badge.svg)](https://github.com/thomis/sensu-plugins-go/actions/workflows/01_test_and_build.yml)
[![02 - Test, Build and Release](https://github.com/thomis/sensu-plugins-go/actions/workflows/02_test_build_and_release.yml/badge.svg)](https://github.com/thomis/sensu-plugins-go/actions/workflows/02_test_build_and_release.yml)

# sensu-plugins-go

A comprehensive collection of Sensu monitoring plugins and handlers written in Go. This repository is part of a fork chain ([hico-horiuchi](https://github.com/hico-horiuchi/sensu-plugins-go) → ... → [thomis](https://github.com/thomis/sensu-plugins-go)) based on Sensu plugins & libraries by Golang, extended with additional functionality for modern infrastructure.

## Overview

This project provides production-ready monitoring checks and handlers for Sensu, covering a wide range of services and system metrics. All plugins are compiled into a single compressedbinary for easy deployment.

## Components

| Category | Component | Description | Documentation |
|----------|-----------|-------------|---------------|
| **System Checks** | check-cpu | Monitor CPU usage and alert on high utilization | |
| | check-disk | Check disk space usage and available capacity | |
| | check-memory | Monitor memory usage and swap utilization | |
| | check-process | Verify processes are running with configurable thresholds | |
| | check-uptime | Monitor system uptime | |
| **Network & Connectivity** | check-ping | ICMP ping check with packet loss and latency monitoring | |
| | check-http | HTTP/HTTPS endpoint monitoring with response validation | |
| | check-http-json | JSON API monitoring with response parsing and validation | |
| | check-certificate | SSL/TLS certificate expiration and validation | [README](cmd/check-certificate/README.md) |
| **Database Monitoring** | check-postgres | PostgreSQL connectivity and query performance | |
| | check-mysql-ping | MySQL connectivity check | |
| | check-mysql-processes | Monitor MySQL process list and connections | |
| | check-oracle-ping | Oracle database connectivity | |
| | check-oracle-validity | Oracle database object validity checks | |
| | check-redis | Redis server monitoring | |
| **Application Services** | check-elasticsearch | Elasticsearch cluster health monitoring | |
| | check-nginx | Nginx status and performance metrics | |
| | check-postfix | Postfix mail server monitoring | |
| | check-postfix-queue | Monitor Postfix queue size | |
| | check-rabbitmq | RabbitMQ queue and cluster monitoring | |
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

Download the latest release from the [Releases](https://github.com/thomis/sensu-plugins-go/releases) page. The binary includes all checks and handlers in a single executable.

```bash
# Extract the archive
tar -xzf sensu-checks-go.linux.amd64.tar.gz

# Make it executable (if needed)
chmod +x sensu-checks-go

# Run a specific check
./sensu-checks-go check-disk -w 80 -c 90
```

## Development

### Build Process

The typical development workflow is:

1. **Write code and tests** - Implement features with corresponding test coverage
2. **Run `make`** - Default command that formats, lints, tests, and builds for your local platform
3. **Run `make cover`** - Review test coverage in your browser
4. **Create a GitHub release** - The "02 - Test, Build and Release" workflow automatically builds for all platforms

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
- Go 1.25.0 or later
- Oracle Instant Client SDK (for Oracle checks)
- Make

## System Requirements

### Supported Platforms
- **Linux**: AMD64, ARM64
- **macOS**: ARM64 (Apple Silicon)

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
