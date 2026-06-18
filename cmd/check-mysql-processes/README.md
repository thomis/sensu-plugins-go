# check-mysql-processes

A Sensu check plugin that monitors the number of MySQL processes (connections)
from `information_schema.PROCESSLIST` and alerts when the count falls outside the
configured thresholds.

## Features

- **Process Count Monitoring**: Counts rows in `information_schema.PROCESSLIST`
- **Min/Max Thresholds**: Warning and critical thresholds support both a lower
  bound (too few processes) and an optional upper bound (too many)
- **Perfdata Output**: Emits `mysql_processes` performance data
- **Environment Defaults**: User and password default to `MYSQL_USER` / `MYSQL_PASSWORD`

## Usage

```bash
check-mysql-processes -w <min[:max]> -c <min[:max]> [OPTIONS]
```

### Options

- `-h, --host` - MySQL host to connect to (default: `localhost`)
- `-P, --port` - MySQL TCP port (default: `3306`)
- `-u, --user` - MySQL user (default: `$MYSQL_USER`)
- `-p, --password` - MySQL user password (default: `$MYSQL_PASSWORD`)
- `-w, --warning` - Warning threshold as `min:max` (max is optional)
- `-c, --critical` - Critical threshold as `min:max` (max is optional)

## Thresholds

Thresholds are given as `min:max`, where `max` is optional:

- The check is **critical** if the process count is `>= critical max` (when a max
  is set) or `<= critical min`.
- Otherwise it is **warning** if the count is `>= warning max` (when a max is set)
  or `<= warning min`.
- Otherwise it is **OK**.

A `min` of `0` effectively disables the lower bound; omitting `max` disables the
upper bound. `min` must not be greater than `max`.

## Examples

```bash
# Warn at >= 200 connections, critical at >= 300
check-mysql-processes -u monitor -p secret -w 0:200 -c 0:300

# Also alert if there are too few connections (e.g. app not connecting)
check-mysql-processes -u monitor -p secret -w 5:200 -c 1:300

# Remote server
check-mysql-processes -h db.example.com -u monitor -p secret -w 0:200 -c 0:300
```

## Exit Codes

- **0 (OK)**: Process count within thresholds
- **1 (WARNING)**: Process count crossed a warning bound
- **2 (CRITICAL)**: Process count crossed a critical bound
- **3 (ERROR)**: Invalid threshold argument, connection failed, or the query failed

## Output Examples

```
CheckMySQLProceses OK: MySQL process Count 42 | mysql_processes=42;0:200;0:300;0
CheckMySQLProceses WARNING: 210 MySQL processes exceed threshold of 200 | mysql_processes=210;0:200;0:300;0
CheckMySQLProceses CRITICAL: 305 MySQL processes exceed threshold of 300 | mysql_processes=305;0:200;0:300;0
```

## Use Cases

- **Connection Saturation**: Alert before reaching `max_connections`
- **Application Health**: Detect when an application stops connecting (too few processes)
- **Capacity Planning**: Track connection counts over time via perfdata

## Notes

- Uses the `go-sql-driver/mysql` driver.
- Queries `information_schema.PROCESSLIST`; the connecting user needs the
  `PROCESS` privilege to see processes owned by other users.
