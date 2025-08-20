# check-uptime

A Sensu check plugin for monitoring system uptime.

## Features

- **System Uptime Monitoring**: Reports the current system uptime
- **Human-Readable Format**: Displays uptime in days, hours, minutes, and seconds
- **Cross-Platform Support**: Works on Linux, macOS, and other Unix-like systems
- **Simple Output**: Provides clear, formatted uptime information
- **Always OK Status**: Returns system information without thresholds

## Usage

```bash
check-uptime
```

### Options

This check does not require any command-line options. It automatically retrieves the system uptime.

## Examples

```bash
# Check system uptime
check-uptime
```

## Exit Codes

- **0 (OK)**: Always returns OK with the current uptime
- **1 (WARNING)**: Not used by this check
- **2 (CRITICAL)**: Not used by this check
- **3 (ERROR)**: Only if unable to retrieve system uptime

## Output Examples

**Standard Output:**
```
CheckUptime OK: Uptime is 7 days, 14 hours, 32 minutes, 15 seconds
```

**After Recent Reboot:**
```
CheckUptime OK: Uptime is 2 hours, 15 minutes, 42 seconds
```

**Less Than One Hour:**
```
CheckUptime OK: Uptime is 45 minutes, 30 seconds
```

**Less Than One Minute:**
```
CheckUptime OK: Uptime is 28 seconds
```

## Use Cases

- **Server Monitoring**: Track server availability and detect unexpected reboots
- **Maintenance Windows**: Verify systems came back online after planned maintenance
- **Compliance**: Document system availability for audit purposes
- **Performance Analysis**: Correlate performance issues with system restarts

## Notes

- The check retrieves uptime using the system's boot time information
- Uptime is calculated from the last system boot, not service start time
- The output format automatically adjusts based on the uptime duration (omits zero values)
- This check is informational and does not trigger alerts based on uptime values