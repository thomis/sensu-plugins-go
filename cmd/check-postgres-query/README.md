# check-postgres-query

A flexible Sensu check plugin that runs an arbitrary query against a PostgreSQL
database and lets the **query itself decide the result**. The query returns two
values — a status and a message — so all control stays with the supplied SQL.

## How it works

The statement must return a single row with exactly two columns:

1. **status** — one of `ok`, `warn`, `warning` or `error` (case-insensitive)
2. **message** — a free-form string shown in the check output

The status is mapped to a Sensu exit code:

| Query status | Result | Exit code |
|--------------|--------|-----------|
| `ok` | OK | 0 |
| `warn` / `warning` | WARNING | 1 |
| `error` / `critical` | CRITICAL | 2 |

Because a PostgreSQL function returns a result set, a stored function fits the
same model — just `SELECT` from it (see examples).

## Features

- **Query-Driven Status**: Thresholds and logic live in SQL, not in the check
- **Inline or File**: Pass the query inline (`-q`) or from a file (`--query-file`)
- **Function Support**: Call a stored function via `SELECT ... FROM my_func()`
- **Custom Message**: The second column is shown verbatim in the check output
- **Configurable Timeout**: Set the connection/query timeout duration

## Usage

```bash
check-postgres-query [OPTIONS]
```

### Options

- `-h, --host` - Host (default: `localhost`)
- `-P, --port` - Port (default: `5432`)
- `-u, --user` - User
- `-p, --password` - Password
- `-d, --database` - Database (default: `test`)
- `-q, --query` - Inline query returning two values: status and message
- `--query-file` - File containing the query (alternative to `-q`)
- `-T, --timeout` - Connection/query timeout (default: 30s)

Provide exactly one of `-q` or `--query-file`.

## Examples

```bash
# Warn/critical based on active connections
check-postgres-query -u monitor -p secret -d appdb \
  -q "select case when count(*) > 100 then 'error'
                  when count(*) > 50  then 'warn'
                  else 'ok' end,
             'active connections: ' || count(*)
      from pg_stat_activity"

# Query from a file (handy for multi-line SQL)
check-postgres-query -h db.example.com -u monitor -p secret -d appdb --query-file /etc/sensu/queries/connections.sql

# Call a stored function that returns (status, message)
check-postgres-query -u monitor -p secret -d appdb \
  -q "select status, message from health_check()"

# Replication lag threshold
check-postgres-query -u monitor -p secret -d appdb \
  -q "select case when extract(epoch from (now() - pg_last_xact_replay_timestamp())) > 60
                  then 'error' else 'ok' end,
             'replication lag (s): ' || coalesce(extract(epoch from (now() - pg_last_xact_replay_timestamp()))::text, 'n/a')"
```

## Query File Format

When using `--query-file`, the file contains the full statement. A trailing `;`
is stripped automatically.

```sql
-- connections.sql
select case when count(*) > 100 then 'error'
            when count(*) > 50  then 'warn'
            else 'ok' end,
       'active connections: ' || count(*)
from pg_stat_activity;
```

## Exit Codes

- **0 (OK)**: query returned status `ok`
- **1 (WARNING)**: query returned status `warn`/`warning`
- **2 (CRITICAL)**: query returned status `error`/`critical`, or a database/
  connection/query error occurred (including timeout)
- **3 (ERROR)**: usage/contract error — no query source (or both `-q` and
  `--query-file`) provided, or the query returned an unrecognized status value

## Output Examples

```
check-postgres-query OK: active connections: 12
check-postgres-query WARNING: active connections: 73
check-postgres-query CRITICAL: active connections: 142
check-postgres-query CRITICAL: pq: relation "missing" does not exist
check-postgres-query ERROR: no query provided (use -q for an inline query or --query-file for a query file)
```

## Use Cases

- **Custom Thresholds**: Alert on any metric you can express in SQL (connections, lag, row counts, job status)
- **Business-Level Checks**: Validate application-specific invariants directly in the database
- **Stored Health Checks**: Reuse a function that already encapsulates health logic
- **Data Freshness**: Verify that a table was updated within an expected window

## Notes

- Uses the `lib/pq` PostgreSQL driver.
- Connects with `sslmode=disable`.
- Only the first row is evaluated; design the query to return a single summary row.
- Shares its query/status logic with `check-oracle-query` (package `pkg/dbquery`).
