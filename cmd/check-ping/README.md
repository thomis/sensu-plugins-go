# check-ping

A Sensu check plugin for TCP connectivity testing.

## Features

- **TCP Connectivity Check**: Verifies TCP connection to a host and port
- **Configurable Timeout**: Set connection timeout in seconds
- **Port Flexibility**: Test any TCP port (SSH, HTTP, database ports, etc.)
- **Simple Output**: Returns connection status with host:port information
- **Cross-Platform Support**: Works on Linux, macOS, Windows, and other platforms

## Usage

```bash
check-ping [OPTIONS]
```

### Options

- `-h, --host` - Host to connect to (default: "localhost")
- `-P, --port` - Port to connect to (default: 22)
- `-t, --timeout` - Connection timeout in seconds (default: 5)

## Examples

```bash
# Check SSH connectivity (default port 22)
check-ping -h server.example.com

# Check HTTP connectivity
check-ping -h www.example.com -P 80

# Check HTTPS connectivity
check-ping -h www.example.com -P 443

# Check MySQL connectivity
check-ping -h db.example.com -P 3306

# Check with custom timeout
check-ping -h slow-server.example.com -P 8080 -t 10

# Check local service
check-ping -h localhost -P 9200
```

## Exit Codes

- **0 (OK)**: Successfully connected to host:port
- **3 (ERROR)**: Connection failed (timeout, connection refused, host unreachable, etc.)

## Output Examples

**Successful Connection:**
```
CheckPing OK: server.example.com:22
```

**Connection Failed:**
```
CheckPing ERROR: dial tcp 192.168.1.100:80: connect: connection refused
```

**Timeout:**
```
CheckPing ERROR: dial tcp 10.0.0.1:443: i/o timeout
```

**Host Not Found:**
```
CheckPing ERROR: dial tcp: lookup nonexistent.example.com: no such host
```

## Use Cases

- **Service Availability**: Verify that services are listening on expected ports
- **Network Connectivity**: Test network paths between systems
- **Firewall Testing**: Confirm firewall rules allow required connections
- **Load Balancer Health**: Check backend server availability
- **Database Connectivity**: Verify database servers are accepting connections

## Notes

- This is a TCP connectivity check, not ICMP ping
- Only tests if a TCP connection can be established, not application-level health
- Does not send or validate any data after connection
- Useful for quick port availability checks
- The connection is immediately closed after successful establishment