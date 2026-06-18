# metrics-cpu

A Sensu metrics plugin that emits overall CPU usage as a percentage in Graphite
plaintext format.

## Features

- **CPU Usage Sampling**: Measures non-idle CPU time over a sampling interval
- **Graphite Output**: Prints a single `cpu.usage` metric
- **Configurable Sampling**: Adjust the sampling window

## Usage

```bash
metrics-cpu [OPTIONS]
```

### Options

- `-s, --sleep` - Sampling interval in seconds (default: `1`)

## Output

Graphite plaintext: `<hostname>.<scheme> <value> <unix-timestamp>`

```
myhost.cpu.usage 12.345678 1718700000
```

`value` is the percentage of CPU time spent non-idle during the sampling window.

## Examples

```bash
# Sample over 1 second (default)
metrics-cpu

# Sample over 5 seconds
metrics-cpu -s 5
```

## Use Cases

- **Graphite/Carbon Pipelines**: Feed CPU utilisation into time-series storage
- **Trending**: Track CPU usage over time alongside other metrics

## Notes

- Reads CPU counters from the system (`/proc/stat` via the shared `common` package).
- On error nothing is printed (no metric line is emitted).
