# handler-elasticsearch

A Sensu event handler that indexes a check's metric output into Elasticsearch,
one document per metric line.

## Features

- **Metric Indexing**: Stores each metric line as an Elasticsearch document
- **Daily Indices**: Writes to a date-suffixed index (`<index>-YYYY.MM.DD`)
- **Graphite Line Format**: Parses `key value timestamp` metric lines

## How it works

The handler reads the Sensu event JSON on **stdin**, splits the check output into
lines, and POSTs each line as a document. Each metric line is expected in
Graphite plaintext form:

```
<key> <value> <unix-timestamp>
```

and is indexed as:

```json
{ "key": "<key>", "value": <value>, "@timestamp": "<RFC3339>" }
```

## Configuration

Default config path: `/etc/sensu/conf.d/handler-elasticsearch.json`

```json
{
  "elasticsearch": {
    "host": "localhost",
    "port": 9200,
    "index": "sensu-metrics"
  }
}
```

Documents are written to `http://<host>:<port>/<index>-YYYY.MM.DD/<check-name>/<id>`.

## Usage

Configure it as a Sensu handler; Sensu pipes the event to the handler's stdin:

```bash
echo "$EVENT_JSON" | handler-elasticsearch
```

## Notes

- Best paired with metrics checks that emit Graphite plaintext lines.
- Lines that don't parse as `key value timestamp` are skipped.
- Connects over plain `http`; no authentication or TLS is currently supported.
