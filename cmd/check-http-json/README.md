# check-http-json

A Sensu check plugin for monitoring JSON API endpoints with advanced validation.

## Features

- **JSON API Monitoring**: Check REST API endpoints returning JSON responses
- **Response Validation**: Validate response codes and body content with regex patterns
- **Multiple HTTP Methods**: Support for GET, POST, PUT, DELETE, PATCH, etc.
- **Request Body Support**: Send JSON payloads with requests
- **Pattern Matching**: Use regular expressions to validate response content
- **Proxy Support**: Configure HTTP proxy settings
- **Authentication**: Basic authentication support
- **Performance Metrics**: Reports response time in milliseconds
- **SSL/TLS Support**: HTTPS with optional certificate validation

## Usage

```bash
check-http-json [OPTIONS]
```

### Options

- `-u, --url` - URL to check (default: "http://localhost/")
- `-t, --timeout` - Request timeout (default: 15s)
- `--username` - Username for basic authentication
- `--password` - Password for basic authentication
- `-k, --insecure` - Skip SSL certificate verification (default: false)
- `-m, --method` - HTTP method (GET, POST, PUT, DELETE, PATCH, etc.) (default: "GET")
- `-b, --body` - JSON body string to send with request
- `-p, --pattern` - Regular expression pattern to match against response body
- `--proxy-url` - Proxy URL (can include port)
- `--no-proxy` - Disable proxy usage (including environment variables)
- `-c, --code` - Expected response code (default: 200)

## Examples

```bash
# Simple GET request
check-http-json -u https://api.example.com/health

# Check for specific response code
check-http-json -u https://api.example.com/v1/status -c 201

# POST request with JSON body
check-http-json -u https://api.example.com/data \
  -m POST \
  -b '{"name":"test","value":123}'

# Validate response contains specific pattern
check-http-json -u https://api.example.com/status \
  -p '"status":\s*"healthy"'

# Check with authentication
check-http-json -u https://api.example.com/secure \
  --username apiuser \
  --password apikey

# Use proxy
check-http-json -u https://external-api.com/data \
  --proxy-url http://proxy.company.com:8080

# PUT request with body and pattern validation
check-http-json -u https://api.example.com/users/123 \
  -m PUT \
  -b '{"email":"user@example.com"}' \
  -p '"updated":\s*true'

# DELETE request expecting 204 No Content
check-http-json -u https://api.example.com/items/456 \
  -m DELETE \
  -c 204

# Check with custom timeout
check-http-json -u https://slow-api.example.com/report \
  -t 30s
```

## Exit Codes

- **0 (OK)**: Response code matches expected and pattern matches (if specified)
- **2 (CRITICAL)**: Response code mismatch or pattern doesn't match
- **3 (ERROR)**: Connection error, timeout, or invalid configuration

## Output Examples

**Successful Response:**
```
check-http-json OK: Status code [200], took [125.3 ms]
```

**Response Code Mismatch:**
```
check-http-json CRITICAL: Status code [404], body [{"error":"Not found"}]
```

**Pattern Mismatch:**
```
check-http-json CRITICAL: Status code [200], pattern ["status":\s*"healthy"] doesn't match with [{"status":"degraded","services":["db"]}]
```

**Connection Error:**
```
check-http-json ERROR: Post "https://api.example.com/data": dial tcp: i/o timeout
```

## Pattern Matching Examples

| Pattern | Matches |
|---------|---------|
| `"status":\s*"ok"` | JSON with status field equal to "ok" |
| `"count":\s*[1-9][0-9]*` | JSON with count field > 0 |
| `"error":\s*null` | JSON with null error field |
| `\{"success":\s*true` | JSON starting with success: true |
| `"items":\s*\[.+\]` | JSON with non-empty items array |
| `"timestamp":\s*"2024-` | JSON with timestamp starting with 2024 |

## Use Cases

- **REST API Monitoring**: Comprehensive API endpoint health checks
- **Microservices Health**: Monitor service mesh health endpoints
- **Data Validation**: Verify API responses contain expected data
- **Integration Testing**: Continuous monitoring of API integrations
- **Performance Monitoring**: Track API response times
- **CRUD Operations**: Test all HTTP methods (GET, POST, PUT, DELETE)

## Notes

- Content-Type header is automatically set to "application/json"
- Response time is measured and reported in milliseconds
- Pattern matching uses RE2 syntax (Go regular expressions)
- Body parameter should be valid JSON when provided
- Proxy settings can be overridden by environment variables unless --no-proxy is used
- Empty response bodies are handled gracefully