# check-oracle-validity

A Sensu check plugin for monitoring Oracle database object validity.

## Features

- **Object Validity Check**: Detect invalid database objects (procedures, functions, packages, views, etc.)
- **Single Schema Mode**: Check objects in a single database schema
- **Batch Mode**: Check multiple schemas from a configuration file
- **Object Type Filtering**: Exclude specific object types from validation
- **Parallel Checking**: Concurrent validity checks for multiple schemas
- **Detailed Reporting**: Lists all invalid objects with their types
- **Configurable Timeout**: Set query timeout duration

## Usage

```bash
check-oracle-validity [OPTIONS]
```

### Options

- `-u, --username` - Oracle username
- `-p, --password` - Oracle password
- `-d, --database` - Database name (TNS name or connection string)
- `-f, --file` - File with connection strings to check (line format: label,username/password@database)
- `-T, --timeout` - Query timeout (default: 30s)
- `-t, --exclude-types` - Exclude given object types from validity check (can be specified multiple times)

## Examples

```bash
# Check single schema
check-oracle-validity -u app_user -p app_pass -d PRODDB

# Check with TNS alias
check-oracle-validity -u scott -p tiger -d ORCL

# Exclude certain object types
check-oracle-validity -u user -p pass -d TESTDB -t INDEX -t SYNONYM

# Multiple exclusions
check-oracle-validity -u user -p pass -d DB -t INDEX -t SYNONYM -t TRIGGER

# Check with custom timeout
check-oracle-validity -u user -p pass -d SLOWDB -T 120s

# Check multiple schemas from file
check-oracle-validity -f /etc/sensu/oracle-schemas.txt

# Check multiple schemas excluding indexes
check-oracle-validity -f schemas.txt -t INDEX
```

## Connection File Format

When using the `-f` option, the file should contain one connection per line:

```
App Schema,app_user/app_pass@PRODDB
Reporting,report_user/report_pass@PRODDB
ETL Process,etl_user/etl_pass@//etl-host:1521/ETLDB
Archive,archive_user/archive_pass@ARCHDB
```

## Exit Codes

- **0 (OK)**: All objects are valid in all checked schemas
- **2 (CRITICAL)**: One or more invalid objects found or connection failed

## Output Examples

**All Objects Valid:**
```
check-oracle-validity OK: All objects are valid
```

**Invalid Objects Found:**
```
check-oracle-validity CRITICAL: invalid objects: 3
PACKAGE BODY                            PKG_CUSTOMER_MGMT
PROCEDURE                               PROC_DAILY_REPORT
VIEW                                    V_SALES_SUMMARY
```

**Multiple Schemas Success:**
```
check-oracle-validity OK: 4/4 connections are fine
```

**Multiple Schemas with Invalid Objects:**
```
check-oracle-validity CRITICAL: 2/4 connections are fine
- App Schema (app_user@PRODDB): invalid objects: 2
FUNCTION                                FN_CALCULATE_TAX
PACKAGE                                 PKG_UTILITIES

- ETL Process (etl_user@//etl-host:1521/ETLDB): invalid objects: 1
PROCEDURE                               PROC_LOAD_DATA
```

## Use Cases

- **Code Deployment Validation**: Verify deployments didn't break database objects
- **Schema Health Monitoring**: Detect compilation errors in stored procedures
- **Development Environment Checks**: Ensure dev/test schemas are valid
- **Pre-Production Validation**: Check object validity before production deployments
- **Dependency Management**: Detect broken dependencies between database objects

## Notes

- Queries the USER_OBJECTS view for invalid objects
- Only checks objects owned by the connected user
- Invalid objects can cause runtime errors in applications
- Common causes of invalid objects include missing dependencies or syntax errors
- Excluding INDEX type is common as indexes can become unusable but still function
- Objects become invalid when dependent objects are modified or dropped