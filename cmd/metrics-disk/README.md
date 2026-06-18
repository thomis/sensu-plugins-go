# metrics-disk

A Sensu metrics plugin that emits overall disk usage as a percentage across local
filesystems, in Graphite plaintext format.

## Features

- **Aggregate Disk Usage**: Sums used and available space across local filesystems
- **Graphite Output**: Prints a single `disk.usage` metric

## Usage

```bash
metrics-disk
```

This plugin takes no options.

## Output

Graphite plaintext: `<hostname>.<scheme> <value> <unix-timestamp>`

```
myhost.disk.usage 63.42 1718700000
```

`value` is `100 * used / (used + available)` aggregated over all local
filesystems.

## Examples

```bash
metrics-disk
```

## Use Cases

- **Capacity Trending**: Track aggregate disk utilisation over time
- **Graphite/Carbon Pipelines**: Feed disk usage into time-series storage

## Notes

- Uses `df -lP` and aggregates local (`-l`) filesystems only.
- Requires the `df` binary to be available on the host.
