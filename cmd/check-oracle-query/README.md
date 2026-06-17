# check-oracle-query

A flexible Sensu check plugin that runs an arbitrary query against an Oracle
database and lets the **query itself decide the result**. The query returns two
values — a status and a message — so all control stays with the supplied SQL or
PL/SQL.

## Features

- **Query-Driven Status**: The query returns the status, so thresholds and logic live in SQL/PL-SQL
- **Inline or File**: Pass the query inline (`-q`) or from a file (`--query-file`) for multi-line statements
- **SQL and PL/SQL**: Run a `SELECT` or a PL/SQL block — automatically detected
- **Procedure & Function Support**: Call stored procedures or functions via `:status`/`:message` OUT binds
- **Batch Mode**: Run the same query against many connections from a file (`-f`), in parallel
- **Custom Message**: The second return value is shown verbatim in the check output
- **Configurable Timeout**: Set the connection/query timeout duration
- **Error Details**: Provides specific Oracle error information

## How it works

The statement must produce exactly two values:

1. **status** — one of `ok`, `warn`, `warning` or `error` (case-insensitive)
2. **message** — a free-form string shown in the check output

The status is mapped to a Sensu exit code:

| Query status | Result | Exit code |
|--------------|--------|-----------|
| `ok` | OK | 0 |
| `warn` / `warning` | WARNING | 1 |
| `error` / `critical` | CRITICAL | 2 |

Two statement kinds are supported and detected automatically:

- **SQL** (default): a `SELECT` returning a single row with two columns
  `(status, message)`. Only the first row is used.
- **PL/SQL block**: a statement starting with `begin` or `declare`. It is
  executed with two OUT bind variables, `:status` and `:message`, which lets you
  call a **procedure** or **function**.

In **batch mode** (`-f`) the same query runs against every connection in the
file (in parallel). The overall result is the worst single result
(**worst-status-wins**: critical > warning > ok). A connection/query failure or
an unrecognized status counts as critical for that connection.

## Usage

```bash
check-oracle-query [OPTIONS]
```

### Options

- `-u, --username` - Oracle username
- `-p, --password` - Oracle password
- `-d, --database` - Database name (TNS name or connection string)
- `-q, --query` - Inline query returning two values: status and message
- `--query-file` - File containing the query (alternative to `-q`)
- `-f, --file` - File with connection strings for batch mode (line format: `label,username/password@database`)
- `-T, --timeout` - Connection/query timeout (default: 30s)

Provide exactly one query source (`-q` or `--query-file`). Provide either a
single connection (`-u`/`-p`/`-d`) or a batch connection file (`-f`).

## Examples

```bash
# Inline SQL against a single database: warn/critical based on active sessions
check-oracle-query -u scott -p tiger -d ORCL \
  -q "select case when count(*) > 100 then 'error'
                  when count(*) > 50  then 'warn'
                  else 'ok' end,
             'active sessions: ' || count(*)
      from v\$session where status = 'ACTIVE'"

# Query from a file (handy for multi-line SQL / PL-SQL)
check-oracle-query -u appuser -p apppass -d PRODDB --query-file /etc/sensu/queries/sessions.sql

# Call a stored procedure with two OUT parameters
check-oracle-query -u user -p pass -d TESTDB \
  -q "begin my_pkg.health_check(:status, :message); end;"

# Use a function's return values
check-oracle-query -u user -p pass -d TESTDB \
  -q "begin :status := my_pkg.health_status; :message := my_pkg.health_message; end;"

# Batch: run the same query against all connections in a file
check-oracle-query -f /etc/sensu/oracle-connections.txt -q "select 'ok', 'fine' from dual"

# Batch with the query in a file and a custom timeout
check-oracle-query -f connections.txt --query-file sessions.sql -T 120s
```

## Query File Format

When using `--query-file`, the file contains the full statement. A trailing `;`
(SQL) or `/` terminator (PL/SQL) is stripped automatically, so files exported
from SQL\*Plus / SQLcl work as-is.

```sql
-- sessions.sql
select case when count(*) > 100 then 'error'
            when count(*) > 50  then 'warn'
            else 'ok' end,
       'active sessions: ' || count(*)
from v$session
where status = 'ACTIVE'
```

```sql
-- health.sql (PL/SQL calling a procedure)
begin
  my_pkg.health_check(:status, :message);
end;
/
```

## Connection File Format

When using `-f` (batch mode), the file contains one connection per line:

```
Production,produser/prodpass@PRODDB
Staging,stageuser/stagepass@//staging-host:1521/STAGEDB
Development,devuser/devpass@DEVDB
```

The same query is executed against each connection.

## Exit Codes

- **0 (OK)**: query returned status `ok` (all connections in batch mode)
- **1 (WARNING)**: query returned status `warn`/`warning` (worst result in batch mode)
- **2 (CRITICAL)**: query returned status `error`/`critical`, or a database/
  connection/query error occurred (including timeout)
- **3 (ERROR)**: usage/contract error — no query source (or both `-q` and
  `--query-file`) provided, or the query returned an unrecognized status value
  (single mode)

## Output Examples

**Single connection — status OK:**
```
check-oracle-query OK: active sessions: 12
```

**Single connection — status CRITICAL (from query):**
```
check-oracle-query CRITICAL: active sessions: 142
```

**Single connection — database/query error:**
```
check-oracle-query CRITICAL: ORA-00942: table or view does not exist
```

**Batch — all OK:**
```
check-oracle-query OK: 0 critical, 0 warning, 3 ok (of 3)
```

**Batch — with failures:**
```
check-oracle-query CRITICAL: 1 critical, 1 warning, 1 ok (of 3)
- Production (produser@PRODDB): CRITICAL active sessions: 142
- Staging (stageuser@//staging-host:1521/STAGEDB): WARNING active sessions: 73
```

**Usage error:**
```
check-oracle-query ERROR: no query provided (use -q for an inline query or --query-file for a query file)
```

## Common Oracle Error Codes

| Error Code | Description |
|------------|-------------|
| ORA-00942 | Table or view does not exist |
| ORA-01017 | Invalid username/password |
| ORA-06550 | PL/SQL compilation error (e.g. wrong procedure/argument) |
| ORA-12154 | TNS: could not resolve the connect identifier |
| ORA-12541 | TNS: no listener |
| ORA-01034 | Oracle not available |

## Use Cases

- **Custom Thresholds**: Alert on any metric you can express in SQL (sessions, lag, row counts, job status)
- **Business-Level Checks**: Validate application-specific invariants directly in the database
- **Stored Health Checks**: Reuse an existing procedure/function that already encapsulates health logic
- **Data Freshness**: Verify that a table was updated within an expected window
- **Multi-Database Monitoring**: Run one health query across many databases at once (batch mode)

## Notes

- Requires Oracle client libraries or Oracle Instant Client.
- Uses the godror driver for Oracle connectivity.
- The connection string can be a TNS alias or a full connection descriptor.
- For PL/SQL, the bind variable names are fixed: `:status` and `:message`.
- Only the first row of a `SELECT` is evaluated; design the query to return a
  single summary row.
- In batch mode the same query runs against every connection; the timeout
  applies per connection and to the overall batch.
