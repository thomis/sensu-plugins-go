# check-elasticsearch

A Sensu check plugin that monitors Elasticsearch cluster health via the
`_cluster/health` API and maps the cluster status colour to a Sensu result.

## Features

- **Cluster Health Check**: Queries the `_cluster/health` endpoint
- **Status Mapping**: `green` → OK, `yellow` → WARNING, `red` → CRITICAL
- **Configurable Timeout**: Set the HTTP request timeout

## Usage

```bash
check-elasticsearch [OPTIONS]
```

### Options

- `-h, --host` - Host (default: `localhost`)
- `-P, --port` - Port (default: `9200`)
- `-t, --timeout` - HTTP timeout in seconds (default: `30`)

## Examples

```bash
# Check a local Elasticsearch node
check-elasticsearch

# Check a remote cluster with a custom timeout
check-elasticsearch -h es.example.com -P 9200 -t 10
```

## Exit Codes

- **0 (OK)**: Cluster status is `green`
- **1 (WARNING)**: Cluster status is `yellow`
- **2 (CRITICAL)**: Cluster status is `red`
- **3 (ERROR)**: Request failed (host unreachable, timeout, etc.)

## Output Examples

```
CheckElasticsearch OK: Cluster is green
CheckElasticsearch WARNING: Cluster is yellow
CheckElasticsearch CRITICAL: Cluster is red
CheckElasticsearch ERROR: Get "http://es.example.com:9200/_cluster/health": dial tcp: ...
```

## Use Cases

- **Cluster Availability**: Detect red clusters (unassigned primary shards)
- **Degraded State**: Catch yellow clusters (unassigned replica shards) early
- **Reachability**: Confirm the Elasticsearch HTTP API is responding

## Notes

- Uses the HTTP REST API; no authentication or TLS is currently supported.
- Connects over plain `http`.
