# check-postfix

A Sensu check plugin that reports the number of messages in the Postfix mail
queue (via `mailq`) and alerts when the queue exceeds the configured thresholds.

## Features

- **Queue Size Monitoring**: Parses `mailq` output for the queued message count
- **Warning/Critical Thresholds**: Alert when the queue grows too large
- **Configurable Binary**: Point at a specific `mailq` path

## Usage

```bash
check-postfix [OPTIONS]
```

### Options

- `-p, --path` - Path to the `mailq` binary (default: `/usr/bin/mailq`)
- `-w, --warn` - Warning threshold (default: `5`)
- `-c, --crit` - Critical threshold (default: `10`)

## Examples

```bash
# Default thresholds (warn > 5, critical > 10)
check-postfix

# Custom thresholds
check-postfix -w 50 -c 100

# Custom mailq path
check-postfix -p /opt/postfix/sbin/mailq -w 20 -c 50
```

## Exit Codes

- **0 (OK)**: Queue size `<= warn`
- **1 (WARNING)**: Queue size `> warn`
- **2 (CRITICAL)**: Queue size `> crit`
- **3 (ERROR)**: Failed to run/parse `mailq`

## Output Examples

```
CheckPostfix OK: 0 messages in the postfix mail queue
CheckPostfix WARNING: 7 messages in the postfix mail queue
CheckPostfix CRITICAL: 142 messages in the postfix mail queue
```

## Use Cases

- **Mail Delivery Backlog**: Detect when outbound mail is piling up
- **Relay Health**: Catch delivery failures that grow the deferred queue

## Notes

- Invokes `mailq` through a shell and parses its summary line; the empty-queue
  message (`Mail queue is empty`) is treated as `0`.
- Requires the `mailq` binary and a shell (`bash`) to be available on the host.
