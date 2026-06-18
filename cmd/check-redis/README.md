# check-redis

A Sensu check plugin that monitors a Redis server by reading a field from its
`INFO` output and comparing it against an expected value. By default it checks
that the server `role` is `master`.

## Features

- **INFO Field Check**: Reads any field from the Redis `INFO` command output
- **Value Comparison**: Compares the field against an expected value
- **Replication Monitoring**: Defaults to verifying `role:master`
- **Connectivity Check**: A failure to connect or query is reported as an error

## Usage

```bash
check-redis [OPTIONS]
```

### Options

- `-h, --host` - Host (default: `localhost`)
- `-P, --port` - Port (default: `6379`)
- `-k, --key` - The `INFO` field to read (default: `role`)
- `-v, --value` - The expected value of the field (default: `master`)

## Examples

```bash
# Verify this instance is a replication master (default)
check-redis -h localhost

# Verify a replica
check-redis -h replica.example.com -k role -v slave

# Check a different INFO field
check-redis -h localhost -k loading -v 0
check-redis -h localhost -k rdb_last_bgsave_status -v ok
```

## Exit Codes

- **0 (OK)**: The field matches the expected value
- **1 (WARNING)**: The field does not match the expected value
- **3 (ERROR)**: Connection failed or the `INFO` command failed

## Output Examples

**Match:**
```
CheckRedis OK: Redis role is master
```

**Mismatch:**
```
CheckRedis WARNING: Redis role is slave
```

**Connection Failure:**
```
CheckRedis ERROR: dial tcp 127.0.0.1:6379: connect: connection refused
```

## Use Cases

- **Replication Role**: Ensure a node is the expected master or replica
- **Persistence Health**: Check fields like `rdb_last_bgsave_status` or `aof_last_write_status`
- **Startup State**: Verify `loading:0` (server finished loading the dataset)

## Notes

- Uses the `gomodule/redigo` Redis client.
- The field is matched against the `INFO` output as `key:value`; the value is
  taken verbatim (trailing carriage return trimmed).
- Authentication (`AUTH`) and TLS are not currently supported.
