# check-certificate

A Sensu check plugin for monitoring TLS/SSL certificate validity and expiration.

## Features

- **Certificate Validation**: Verifies that certificates are valid and properly configured
- **Expiration Monitoring**: Warns when certificates are approaching expiration
- **Hostname Verification**: Ensures certificates match the requested hostname
- **Detailed Output**: Provides comprehensive certificate information including issuer, validity dates, and DNS names
- **Configurable Thresholds**: Customize warning periods for certificate expiration
- **Timeout Support**: Configurable connection timeout for network operations

## Usage

```bash
check-certificate [OPTIONS]
```

### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--host` | `-h` | `localhost` | Hostname or IP address to check |
| `--port` | `-P` | `443` | Port number for TLS connection |
| `--timeout` | `-t` | `5` | Connection timeout in seconds |
| `--expiry` | `-e` | `30` | Days before expiration to trigger warning |

## Examples

```bash
# Check a certificate on default HTTPS port
check-certificate --host example.com

# Check certificate on mail server
check-certificate --host mail.example.com --port 993

# Warn if certificate expires within 60 days
check-certificate --host example.com --expiry 60

# Check internal service with custom timeout
check-certificate --host internal-api.company.local --port 8443 --timeout 10

# Check multiple domains in a script
for domain in example.com api.example.com mail.example.com; do
    check-certificate --host "$domain" --expiry 45
done
```

## Exit Codes

- **0 (OK)**: Certificate is valid and not expiring soon
- **1 (WARNING)**: Certificate is expiring within the specified threshold
- **2 (CRITICAL)**: Not used by this check
- **3 (ERROR)**: Certificate validation failed or connection error

## Output Examples

**Success:**
```
CheckCertificate OK: example.com:443

Issuer Name: CN=R3,O=Let's Encrypt,C=US
Not Before : 2024-10-15 08:45:00 UTC
Not After  : 2025-01-13 08:44:59 UTC (94.2 days left)
Common Name: R3
DNS Names  : example.com, www.example.com
```

**Possible Errors:**
- `ERROR: certificate not after: 2024-08-01 23:59:59 UTC` - Certificate expired
- `ERROR: certificate not before: 2025-01-01 00:00:00 UTC` - Certificate not yet valid
- `ERROR: x509: certificate is valid for www.example.com, not example.org` - Hostname mismatch
- `ERROR: dial tcp 192.168.1.100:443: i/o timeout` - Connection timeout
- `WARNING: Certificate about to expire in less than 30 days` - Expiring soon