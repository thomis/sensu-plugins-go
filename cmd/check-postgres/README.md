# check-postgres

A Sensu check plugin for monitoring PostgreSQL database connectivity. It connects
to a PostgreSQL server and reports the server version, confirming the database is
reachable and accepting queries.

## Features

- **Connectivity Check**: Verify a PostgreSQL server is reachable and accepting connections
- **Version Reporting**: Reports the PostgreSQL server version on success
- **Credential Validation**: Confirms the supplied user/password can authenticate
- **Configurable Connection**: Host, port, user, password and database are all configurable

## Usage

```bash
check-postgres [OPTIONS]
```

### Options

- `-h, --host` - Host (default: `localhost`)
- `-P, --port` - Port (default: `5432`)
- `-u, --user` - User
- `-p, --password` - Password
- `-d, --database` - Database (default: `test`)

## Examples

```bash
# Check a local PostgreSQL instance
check-postgres -u postgres -p secret -d postgres

# Check a remote server on a custom port
check-postgres -h db.example.com -P 5433 -u monitor -p monitorpass -d appdb

# Use defaults (localhost:5432, database "test")
check-postgres -u postgres -p secret
```

## Exit Codes

- **0 (OK)**: Connection succeeded; server version reported
- **3 (ERROR)**: Connection failed, authentication failed, the query failed, or
  the version could not be parsed

## Output Examples

**Success:**
```
CheckPostgres OK: Server version 15.3
```

**Connection/Authentication Failure:**
```
CheckPostgres ERROR: pq: password authentication failed for user "monitor"
```

**Server Unreachable:**
```
CheckPostgres ERROR: dial tcp 10.0.0.5:5432: connect: connection refused
```

## Use Cases

- **Database Availability**: Monitor that a PostgreSQL server is up and accepting connections
- **Credential Validation**: Verify a monitoring/service account can authenticate
- **Connectivity Checks**: Confirm network reachability between Sensu agents and the database

## Notes

- Uses the `lib/pq` PostgreSQL driver.
- Connects with `sslmode=disable`.
- A successful check runs `select version()` and reports the parsed version
  number, so it validates both connectivity and the ability to execute a query.
