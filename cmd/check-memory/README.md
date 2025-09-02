# check-memory

A Sensu check plugin for monitoring memory usage.

## Features

- **Memory Usage Monitoring**: Tracks total memory usage and available memory
- **Configurable Thresholds**: Set warning and critical levels for memory usage percentage
- **MemAvailable Support**: Uses MemAvailable metric when available for accurate free memory calculation
- **Fallback Calculation**: Falls back to Buffers+Cached on older systems without MemAvailable
- **Performance Data**: Outputs memory metrics for graphing and trending
- **Linux Support**: Reads memory information from /proc/meminfo

## Usage

```bash
check-memory [OPTIONS]
```

### Options

- `-w, --warn` - Warning threshold for memory usage percentage (default: 80%)
- `-c, --crit` - Critical threshold for memory usage percentage (default: 90%)

## Examples

```bash
# Check memory with default thresholds (warn at 80%, critical at 90%)
check-memory

# Set custom thresholds
check-memory -w 70 -c 85

# More conservative thresholds for production systems
check-memory -w 60 -c 80
```

## Exit Codes

- **0 (OK)**: Memory usage is below the warning threshold
- **1 (WARNING)**: Memory usage is between warning and critical thresholds
- **2 (CRITICAL)**: Memory usage exceeds the critical threshold
- **3 (ERROR)**: Unable to retrieve memory statistics

## Output Examples

**Normal Memory Usage:**
```
CheckMemory OK: 45.23% MemTotal:16384.00MB MemAvailable:8977.28MB | mem_usage=45.23%;80;90 mem_available=8977.28MB
```

**High Memory Usage (Warning):**
```
CheckMemory WARNING: 82.50% MemTotal:8192.00MB MemAvailable:1433.60MB | mem_usage=82.50%;80;90 mem_available=1433.60MB
```

**Critical Memory Usage:**
```
CheckMemory CRITICAL: 95.75% MemTotal:4096.00MB MemAvailable:174.08MB | mem_usage=95.75%;80;90 mem_available=174.08MB
```

## Memory Calculation

The check calculates memory usage as:
```
Usage % = 100 - (100 * MemAvailable / MemTotal)
```

Where:
- **MemTotal**: Total physical memory
- **MemAvailable**: Estimate of available memory for starting new applications
- **Fallback**: On systems without MemAvailable, uses Buffers + Cached as an approximation

## Use Cases

- **Server Monitoring**: Prevent out-of-memory conditions before they occur
- **Application Performance**: Identify memory leaks or excessive memory consumption
- **Capacity Planning**: Track memory usage trends for upgrade planning
- **Container Monitoring**: Monitor memory usage in containerized environments
- **Database Servers**: Ensure adequate memory for database operations

## Notes

- Requires access to /proc/meminfo (Linux systems)
- MemAvailable provides a better estimate than free memory alone
- The check automatically handles systems without MemAvailable metric
- Memory values are reported in megabytes (MB)
- Performance data includes usage percentage and available memory