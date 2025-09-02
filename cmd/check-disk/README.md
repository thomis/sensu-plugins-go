# check-disk

A Sensu check plugin for monitoring disk space usage.

## Features

- **Disk Usage Monitoring**: Checks available disk space across all mounted filesystems
- **Configurable Thresholds**: Set warning and critical levels for disk usage percentage
- **Magic Factor Adjustment**: Automatically adjusts thresholds based on filesystem size
- **Filesystem Filtering**: Exclude specific filesystem types or mount points
- **Path-Specific Checks**: Monitor specific paths or all filesystems
- **Performance Data**: Outputs metrics for graphing and trending

## Usage

```bash
check-disk [OPTIONS]
```

### Options

- `-w, --warn` - Warning percentage threshold (default: 80%)
- `-c, --crit` - Critical percentage threshold (default: 100%)
- `-m, --magic` - Magic factor to adjust thresholds based on filesystem size (default: 1.0)
- `-n, --normal` - "Normal" size in GB for threshold baseline (default: 20 GB)
- `-l, --minimum` - Minimum size in GB before applying magic adjustment (default: 100 GB)
- `-x, --exclude` - Comma-separated list of filesystem types to exclude
- `-i, --ignore` - Comma-separated list of mount points to ignore
- `-p, --path` - Limit check to specified path

## Examples

```bash
# Check all filesystems with default thresholds
check-disk

# Set custom warning and critical thresholds
check-disk -w 70 -c 90

# Exclude temporary and special filesystems
check-disk -x tmpfs,devtmpfs

# Ignore specific mount points
check-disk -i /mnt/backup,/tmp

# Check only a specific path
check-disk -p /var

# Use magic factor for dynamic thresholds on large filesystems
check-disk -m 0.9 -n 50
```

## Exit Codes

- **0 (OK)**: All filesystems are below warning threshold
- **1 (WARNING)**: One or more filesystems exceed warning threshold
- **2 (CRITICAL)**: One or more filesystems exceed critical threshold
- **3 (ERROR)**: Unable to retrieve disk usage information

## Output Examples

**All filesystems OK:**
```
CheckDisk OK: OK | /=45%;80.00;100.00 /boot=32%;80.00;100.00 /home=67%;80.00;100.00
```

**Warning threshold exceeded:**
```
CheckDisk WARNING: /var 82% | /=45%;80.00;100.00 /boot=32%;80.00;100.00 /var=82%;80.00;100.00
```

**Critical threshold exceeded:**
```
CheckDisk CRITICAL: /var 95%, /tmp 92% | /=45%;80.00;100.00 /var=95%;80.00;100.00 /tmp=92%;80.00;100.00
```

## Magic Factor Adjustment

The magic factor feature automatically adjusts thresholds based on filesystem size:

- Filesystems smaller than the "normal" size get stricter thresholds
- Filesystems larger than the "normal" size get more lenient thresholds
- Only applies to filesystems larger than the minimum size
- A magic factor < 1.0 makes the adjustment more aggressive
- A magic factor > 1.0 makes the adjustment less aggressive

Example: With `-m 0.9 -n 20`, a 1TB filesystem might have its 80% warning threshold adjusted to 85%, while a 10GB filesystem might have it adjusted to 75%.

## Use Cases

- **Server Monitoring**: Prevent disk space issues before they cause outages
- **Database Servers**: Monitor data partitions to prevent database failures
- **Application Servers**: Track log partition usage to prevent application issues
- **Backup Systems**: Ensure adequate space for backup operations
- **Container Hosts**: Monitor Docker/container storage volumes

## Notes

- Uses the `df` command to retrieve disk usage information
- Automatically excludes network filesystems (only checks local filesystems)
- Percentage calculation is based on used space vs. total space
- Performance data includes warning and critical thresholds for each filesystem
- The check can monitor all filesystems or be limited to specific paths