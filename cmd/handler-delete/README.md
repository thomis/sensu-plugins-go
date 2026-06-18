# handler-delete

A Sensu event handler that deletes a client from the Sensu API when its
`keepalive` check reaches a configured status — useful for automatically
reaping clients that have stopped sending keepalives (e.g. terminated instances).

## Features

- **Automatic Client Cleanup**: Removes stale clients via the Sensu API
- **Scoped by Subscription**: Only acts on clients with matching subscriptions
- **Status-Gated**: Only acts at a configured keepalive status

## How it works

The handler reads the Sensu event JSON on **stdin** and its configuration from a
JSON file. It deletes the client only when **all** of the following hold:

- the check name is `keepalive`,
- the check status equals the configured `status`, and
- the client has at least one of the configured `subscriptions`.

When matched, it calls the Sensu API to delete the client.

## Configuration

Default config path: `/etc/sensu/conf.d/handler-delete.json`

```json
{
  "delete": {
    "status": 2,
    "subscriptions": ["ephemeral", "autoscale"],
    "host": "localhost",
    "port": 4567,
    "user": "admin",
    "password": "secret"
  }
}
```

## Usage

Configure it as a Sensu handler; Sensu pipes the event to the handler's stdin:

```bash
echo "$EVENT_JSON" | handler-delete
```

## Notes

- Targets the legacy Sensu Core API (via `ohgibone/sensu`).
- Only `keepalive` events are considered; other checks are ignored.
- Use a dedicated subscription for clients that should be auto-reaped to avoid
  deleting clients unintentionally.
