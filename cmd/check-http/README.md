# check-http

A Sensu check plugin for monitoring HTTP/HTTPS endpoints.

## Features

- **HTTP/HTTPS Monitoring**: Check web service availability and response codes
- **Basic Authentication**: Support for username/password authentication
- **SSL/TLS Support**: Verify HTTPS endpoints with optional certificate validation
- **Response Code Validation**: Alert based on HTTP status codes
- **Configurable Timeout**: Set request timeout for slow endpoints
- **Redirect Detection**: Returns warning status for 3xx redirect responses

## Usage

```bash
check-http [OPTIONS]
```

### Options

- `-u, --url` - URL to check (default: "http://localhost/")
- `-t, --timeout` - Request timeout in seconds (default: 15)
- `--username` - Username for basic authentication
- `--password` - Password for basic authentication
- `-k, --insecure` - Skip SSL certificate verification (default: false)

## Examples

```bash
# Check HTTP endpoint
check-http -u http://example.com

# Check HTTPS endpoint
check-http -u https://example.com

# Check with basic authentication
check-http -u http://api.example.com --username user --password secret

# Check with longer timeout
check-http -u http://slow-api.example.com -t 30

# Check HTTPS with self-signed certificate
check-http -u https://internal.example.com -k

# Check specific path
check-http -u http://example.com/health

# Check API endpoint
check-http -u https://api.example.com/v1/status
```

## Exit Codes

- **0 (OK)**: Response code 200-299 (success)
- **1 (WARNING)**: Response code 300-399 (redirect)
- **2 (CRITICAL)**: Response code 400+ (client/server error)
- **3 (ERROR)**: Connection error, timeout, or other failure

## Output Examples

**Successful Response:**
```
CheckHTTP OK: 200
```

**Redirect (Warning):**
```
CheckHTTP WARNING: 301
```

**Client Error:**
```
CheckHTTP CRITICAL: 404
```

**Server Error:**
```
CheckHTTP CRITICAL: 503
```

**Connection Error:**
```
CheckHTTP ERROR: Get "http://example.com": dial tcp: i/o timeout
```

## Response Code Interpretation

| Code Range | Status | Description |
|------------|--------|-------------|
| 200-299 | OK | Success responses |
| 300-399 | WARNING | Redirection messages |
| 400-499 | CRITICAL | Client error responses |
| 500-599 | CRITICAL | Server error responses |

## Use Cases

- **Website Monitoring**: Ensure websites are accessible and responding
- **API Health Checks**: Monitor REST API endpoints
- **Service Availability**: Verify microservices are running
- **SSL Certificate Validation**: Detect certificate issues
- **Load Balancer Health**: Check application endpoints behind load balancers

## Notes

- Does not follow redirects (returns the redirect status code)
- Only performs GET requests
- Does not validate response body content
- Timeout applies to the entire request/response cycle
- Basic authentication credentials are sent in the Authorization header