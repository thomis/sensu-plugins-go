# check-postfix-queue

A Sensu check plugin that counts the messages in a specific Postfix spool queue
directory (e.g. `deferred`) and alerts when the count exceeds the configured
thresholds. Unlike `check-postfix`, it inspects the spool directory directly
rather than calling `mailq`.

## Features

- **Per-Queue Monitoring**: Inspect a specific Postfix queue (`deferred`, `active`, `incoming`, …)
- **Warning/Critical Thresholds**: Alert when a queue grows too large
- **No External Binary**: Counts spool files directly on disk

## Usage

```bash
check-postfix-queue [OPTIONS]
```

### Options

- `-q, --queue` - Postfix queue to check (default: `deferred`)
- `-w, --warn` - Warning threshold (default: `5`)
- `-c, --crit` - Critical threshold (default: `10`)

The queue directory inspected is `/var/spool/postfix/<queue>`.

## Examples

```bash
# Check the deferred queue with default thresholds
check-postfix-queue

# Check the active queue with custom thresholds
check-postfix-queue -q active -w 50 -c 100

# Check the incoming queue
check-postfix-queue -q incoming -w 20 -c 50
```

## Exit Codes

- **0 (OK)**: Queue size `<= warn`
- **1 (WARNING)**: Queue size `> warn`
- **2 (CRITICAL)**: Queue size `> crit`
- **3 (ERROR)**: The queue directory cannot be accessed (missing or no permission)

## Output Examples

```
CheckPostfixQueue OK: 0 messages in the postfix mail queue
CheckPostfixQueue WARNING: 7 messages in the postfix mail queue
CheckPostfixQueue CRITICAL: 142 messages in the postfix mail queue
CheckPostfixQueue ERROR: cannot access queue directory /var/spool/postfix/deferred: ...
```

## Use Cases

- **Deferred Backlog**: Watch the `deferred` queue for delivery problems
- **Queue-Specific Alerts**: Set different thresholds per queue
- **Agent-Local Checks**: Run on the mail host where the spool is readable

## Notes

- Counts files recursively under `/var/spool/postfix/<queue>`.
- The Postfix spool is typically root-owned; the check usually needs to run with
  sufficient privileges to read it.
