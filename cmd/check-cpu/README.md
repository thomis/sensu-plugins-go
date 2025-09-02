# check-cpu

A Sensu check plugin for monitoring CPU usage.

## Features

- **CPU Usage Monitoring**: Tracks CPU utilization across user, system, iowait, and idle states
- **Configurable Thresholds**: Set warning and critical levels based on idle CPU percentage
- **Sampling Period**: Adjustable sleep time for CPU usage sampling
- **Performance Data**: Outputs performance metrics for graphing and trending
- **Cross-Platform Support**: Works on Linux, macOS, and other Unix-like systems

## Usage

```bash
check-cpu [OPTIONS]
```

### Options

- `-w, --warn` - Warning threshold for idle CPU (default: 80%)
- `-c, --crit` - Critical threshold for idle CPU (default: 90%)
- `-s, --sleep` - Sleep time in seconds for CPU sampling (default: 1)

## Examples

```bash
# Check CPU with default thresholds (warn at 80% idle, critical at 90% idle)
check-cpu

# Set custom thresholds
check-cpu -w 70 -c 85

# Use longer sampling period for more accurate measurements
check-cpu -s 5
```

## Exit Codes

- **0 (OK)**: CPU idle percentage is above the warning threshold
- **1 (WARNING)**: CPU idle percentage is between warning and critical thresholds
- **2 (CRITICAL)**: CPU idle percentage is below the critical threshold
- **3 (ERROR)**: Unable to retrieve CPU statistics

## Output Examples

**Normal Load:**
```
CheckCPU OK: user=15.32% system=8.45% iowait=0.12% other=0.00% idle=76.11% | cpu_user=15.32%;80;90 cpu_system=8.45%;80;90 cpu_iowait=0.12%;80;90 cpu_other=0.00%;80;90 cpu_idle=76.11%
```

**High Load (Warning):**
```
CheckCPU WARNING: user=45.67% system=22.33% iowait=2.00% other=0.00% idle=30.00% | cpu_user=45.67%;80;90 cpu_system=22.33%;80;90 cpu_iowait=2.00%;80;90 cpu_other=0.00%;80;90 cpu_idle=30.00%
```

**Critical Load:**
```
CheckCPU CRITICAL: user=70.25% system=25.50% iowait=1.25% other=0.00% idle=3.00% | cpu_user=70.25%;80;90 cpu_system=25.50%;80;90 cpu_iowait=1.25%;80;90 cpu_other=0.00%;80;90 cpu_idle=3.00%
```

## Use Cases

- **Server Monitoring**: Track CPU utilization to identify performance bottlenecks
- **Capacity Planning**: Monitor CPU trends to plan for hardware upgrades
- **Performance Troubleshooting**: Identify processes consuming excessive CPU resources
- **Alert Automation**: Trigger alerts when CPU usage exceeds acceptable thresholds

## Notes

- The check uses idle CPU percentage for thresholds (lower idle means higher usage)
- CPU usage is calculated by sampling /proc/stat over the specified sleep period
- The user percentage includes both user and nice CPU time
- Performance data is included in the output for integration with monitoring systems