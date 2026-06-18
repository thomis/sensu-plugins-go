# check-nginx

A Sensu check plugin that verifies the NGINX process is running (via its PID
file) and can optionally read the active connection count from the NGINX status
page (`stub_status`).

## Features

- **Process Check**: Confirms the NGINX process referenced by the PID file is alive
- **Optional Status Check**: Reads the active connection count from the status page
- **Perfdata Output**: Emits `nginx_connections` when the status check is enabled
- **Configurable Timeout**: Set the status page request timeout

## Usage

```bash
check-nginx [OPTIONS]
```

### Options

- `-u, --url` - NGINX status page URL (default: `http://localhost/nginx-status`)
- `-p, --pidFile` - NGINX PID file (default: `/var/run/nginx.pid`)
- `-t, --timeout` - Status page check timeout in seconds (default: `15`)
- `-c, --checkStatus` - Also query the NGINX status page (default: `false`)

## Examples

```bash
# Check that the NGINX process is running
check-nginx

# Custom PID file location
check-nginx -p /run/nginx.pid

# Also read the active connection count from the status page
check-nginx -c -u http://localhost/nginx-status
```

## Exit Codes

- **0 (OK)**: Process is running (and, with `-c`, the status page was read)
- **2 (CRITICAL)**: Process is not running, or the status page check failed

## Output Examples

```
CheckNGINX OK: OK
CheckNGINX OK: connections = 43 | nginx_connections=43
CheckNGINX CRITICAL: failed to read PID file /var/run/nginx.pid, error: open ...: no such file or directory
```

## Use Cases

- **Process Liveness**: Alert when the NGINX master process dies
- **Connection Load**: Track active connections via perfdata (with `-c`)

## Notes

- The status check requires NGINX's `stub_status` module to be enabled and the
  status location reachable at the configured URL.
- The process check reads the PID from the PID file and sends signal `0` to
  verify the process exists; the check must run with sufficient privileges.
