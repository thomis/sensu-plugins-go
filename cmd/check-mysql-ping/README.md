# check-mysql-ping

A Sensu check plugin for monitoring MySQL database connectivity. It connects to a
MySQL server and reports the server version, confirming the database is reachable
and accepting queries.

## Features

- **Connectivity Check**: Verify a MySQL server is reachable and accepting connections
- **Version Reporting**: Reports the MySQL server version on success
- **Credential Validation**: Confirms the supplied user/password can authenticate
- **Environment Defaults**: User and password default to `MYSQL_USER` / `MYSQL_PASSWORD`

## Usage

```bash
check-mysql-ping [OPTIONS]
```

### Options

- `-h, --host` - MySQL host to connect to (default: `localhost`)
- `-P, --port` - MySQL TCP port (default: `3306`)
- `-u, --user` - MySQL user (default: `$MYSQL_USER`)
- `-p, --password` - MySQL user password (default: `$MYSQL_PASSWORD`)
- `-d, --database` - MySQL database (default: `mysql`)

## Examples

```bash
# Check a local MySQL instance
check-mysql-ping -u monitor -p secret

# Check a remote server on a custom port
check-mysql-ping -h db.example.com -P 3307 -u monitor -p secret -d appdb

# Use credentials from the environment
export MYSQL_USER=monitor MYSQL_PASSWORD=secret
check-mysql-ping -h db.example.com
```

## Exit Codes

- **0 (OK)**: Connection succeeded; server version reported
- **3 (ERROR)**: Connection failed, authentication failed, the query failed, or
  the version could not be parsed

## Output Examples

**Success:**
```
CheckMySQLPing OK: Server version 8.0.36
```

**Connection/Authentication Failure:**
```
CheckMySQLPing ERROR: Error 1045: Access denied for user 'monitor'@'10.0.0.5'
```

## Use Cases

- **Database Availability**: Monitor that a MySQL server is up and accepting connections
- **Credential Validation**: Verify a monitoring/service account can authenticate
- **Connectivity Checks**: Confirm network reachability between Sensu agents and the database

## Notes

- Uses the `go-sql-driver/mysql` driver.
- A successful check runs `select version()`, so it validates both connectivity
  and the ability to execute a query.
