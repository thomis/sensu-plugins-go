# check-oracle-ping

A Sensu check plugin for monitoring Oracle database connectivity.

## Features

- **Database Connectivity Check**: Verify Oracle database connections
- **Single Connection Mode**: Test individual database connections
- **Batch Mode**: Test multiple connections from a configuration file
- **Parallel Testing**: Concurrent connection testing for multiple databases
- **Configurable Timeout**: Set connection timeout duration
- **Error Details**: Provides specific Oracle error information
- **UTF-8 Support**: Handles international character sets

## Usage

```bash
check-oracle-ping [OPTIONS]
```

### Options

- `-u, --username` - Oracle username
- `-p, --password` - Oracle password
- `-d, --database` - Database name (TNS name or connection string)
- `-f, --file` - File with connection strings to check (line format: label,username/password@database)
- `-T, --timeout` - Connection timeout (default: 30s)

## Examples

```bash
# Test single database connection
check-oracle-ping -u scott -p tiger -d ORCL

# Test connection with TNS alias
check-oracle-ping -u appuser -p apppass -d PRODDB

# Test connection with full connection string
check-oracle-ping -u user -p pass -d "//localhost:1521/XEPDB1"

# Test with custom timeout
check-oracle-ping -u user -p pass -d TESTDB -T 60s

# Test multiple connections from file
check-oracle-ping -f /etc/sensu/oracle-connections.txt
```

## Connection File Format

When using the `-f` option, the file should contain one connection per line:

```
Production,produser/prodpass@PRODDB
Staging,stageuser/stagepass@//staging-host:1521/STAGEDB
Development,devuser/devpass@DEVDB
Analytics,analyst/pass123@//analytics:1521/DWH
```

## Exit Codes

- **0 (OK)**: All connections are successful
- **2 (CRITICAL)**: One or more connections failed or timeout occurred

## Output Examples

**Single Connection Success:**
```
check-oracle-ping OK: Connection is pingable
```

**Single Connection Failure:**
```
check-oracle-ping CRITICAL: ORA-01017: invalid username/password; logon denied
```

**Multiple Connections Success:**
```
check-oracle-ping OK: 4/4 connections are pingable
```

**Multiple Connections with Failures:**
```
check-oracle-ping CRITICAL: 2/4 connections are pingable
- Production (produser@PRODDB): ORA-12154: TNS:could not resolve the connect identifier specified
- Analytics (analyst@//analytics:1521/DWH): timeout reached
```

**Timeout Error:**
```
check-oracle-ping CRITICAL: timeout reached
```

## Common Oracle Error Codes

| Error Code | Description |
|------------|-------------|
| ORA-01017 | Invalid username/password |
| ORA-12154 | TNS: could not resolve the connect identifier |
| ORA-12514 | TNS: listener does not know of service |
| ORA-12541 | TNS: no listener |
| ORA-01034 | Oracle not available |
| ORA-27101 | Shared memory realm does not exist |

## Use Cases

- **Database Availability**: Monitor database uptime and accessibility
- **Connection Pool Health**: Verify application connection credentials
- **Disaster Recovery**: Test standby database connections
- **Multi-Database Monitoring**: Check multiple database instances simultaneously
- **Service Account Validation**: Verify service account credentials are working

## Notes

- Requires Oracle client libraries or Oracle Instant Client
- Uses godror driver for Oracle connectivity
- Connection string can be TNS alias or full connection descriptor
- Parallel testing speeds up checks for multiple databases
- Timeout applies to each individual connection attempt
- All connections are tested even if some fail (in batch mode)