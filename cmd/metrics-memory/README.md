# metrics-memory

A Sensu metrics plugin that emits memory usage as a percentage in Graphite
plaintext format.

## Features

- **Memory Usage**: Computes used memory as a percentage of total
- **Graphite Output**: Prints a single `memory.usage` metric

## Usage

```bash
metrics-memory
```

This plugin takes no options.

## Output

Graphite plaintext: `<hostname>.<scheme> <value> <unix-timestamp>`

```
myhost.memory.usage 47.81 1718700000
```

`value` is `100 - (100 * free / total)` as reported by `free`.

## Examples

```bash
metrics-memory
```

## Use Cases

- **Memory Trending**: Track memory utilisation over time
- **Graphite/Carbon Pipelines**: Feed memory usage into time-series storage

## Notes

- Parses the output of the `free` command (Linux), so it requires `free` to be
  available on the host.
