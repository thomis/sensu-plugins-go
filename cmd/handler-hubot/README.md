# handler-hubot

A Sensu event handler that forwards check results to a Hubot chat bot endpoint.

## Features

- **Chat Notifications**: POSTs the event to a Hubot `/sensu` endpoint
- **Room Targeting**: Sends to a configurable room
- **Compact Payload**: Includes client, check, output, status and occurrences

## How it works

The handler reads the Sensu event JSON on **stdin** and its configuration from a
JSON file, then POSTs a JSON payload to Hubot:

```json
{
  "client": "web01",
  "check": "check-http",
  "output": "...",
  "status": 2,
  "occurrences": 3
}
```

## Configuration

Default config path: `/etc/sensu/conf.d/handler-hubot.json`

```json
{
  "hubot": {
    "host": "localhost",
    "port": 8080,
    "room": 1
  }
}
```

The request is sent to `http://<host>:<port>/sensu?room=<room>`.

## Usage

Configure it as a Sensu handler; Sensu pipes the event to the handler's stdin:

```bash
echo "$EVENT_JSON" | handler-hubot
```

## Notes

- Requires a Hubot instance exposing a `/sensu` route that accepts the payload.
- Connects over plain `http`.
