# metrics-snmp

A Sensu metrics plugin that collects per-interface network traffic from one or
more devices via SNMP and emits it in Graphite plaintext format.

## Features

- **SNMP Interface Counters**: Reads `ifInOctets` / `ifOutOctets` per interface
- **Multiple Hosts**: Polls a comma-separated list of devices concurrently
- **Per-Interface Metrics**: Emits RX/TX bytes per interface index
- **Configurable Sampling**: Adjust the sampling window

## Usage

```bash
metrics-snmp [OPTIONS]
```

### Options

- `-h, --hosts` - Comma-separated list of hosts to poll (default: `127.0.0.1`)
- `-c, --community` - SNMP community string (default: `public`)
- `-s, --sleep` - Sampling interval in seconds (default: `1`)

## Output

Graphite plaintext: `<host>.<scheme> <value> <unix-timestamp>`, one pair of lines
per interface (the metric hostname is the polled device):

```
192.0.2.1.snmp.rx_bytes.1 1048576 1718700000
192.0.2.1.snmp.tx_bytes.1 524288 1718700000
192.0.2.1.snmp.rx_bytes.2 0 1718700000
192.0.2.1.snmp.tx_bytes.2 0 1718700000
```

`value` is the number of bytes transferred on that interface during the sampling
window; the trailing number is the interface index.

## Examples

```bash
# Poll a single device
metrics-snmp -h 192.0.2.1 -c public

# Poll several devices over a 10 second window
metrics-snmp -h 192.0.2.1,192.0.2.2 -c private -s 10
```

## Use Cases

- **Network Device Monitoring**: Trend per-port throughput on switches/routers
- **Multi-Device Collection**: Gather metrics from many devices in one run

## Notes

- Uses SNMP v2c via the `snmpwalk` command, which must be installed and on `PATH`.
- OIDs polled: `1.3.6.1.2.1.2.2.1.10` (ifInOctets) and `1.3.6.1.2.1.2.2.1.16` (ifOutOctets).
- A host that cannot be polled is skipped silently.
