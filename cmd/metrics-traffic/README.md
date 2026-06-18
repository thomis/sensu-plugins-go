# metrics-traffic

A Sensu metrics plugin that emits network traffic (received and transmitted
bytes) over a sampling interval, in Graphite plaintext format.

## Features

- **RX/TX Bytes**: Reports bytes received and transmitted during the interval
- **Graphite Output**: Prints `traffic.rx_bytes` and `traffic.tx_bytes` metrics
- **Configurable Sampling**: Adjust the sampling window

## Usage

```bash
metrics-traffic [OPTIONS]
```

### Options

- `-s, --sleep` - Sampling interval in seconds (default: `1`)

## Output

Graphite plaintext: `<hostname>.<scheme> <value> <unix-timestamp>`

```
myhost.traffic.rx_bytes 1048576 1718700000
myhost.traffic.tx_bytes 524288 1718700000
```

`value` is the number of bytes transferred during the sampling window.

## Examples

```bash
# Sample over 1 second (default)
metrics-traffic

# Sample over 10 seconds
metrics-traffic -s 10
```

## Use Cases

- **Bandwidth Trending**: Track interface throughput over time
- **Graphite/Carbon Pipelines**: Feed traffic counters into time-series storage

## Notes

- Reads counters from `/sys/class/net/*/statistics/{rx,tx}_bytes` (Linux); the
  loopback interface (`lo`) is excluded.
- Counters from the last non-loopback interface seen are reported.
