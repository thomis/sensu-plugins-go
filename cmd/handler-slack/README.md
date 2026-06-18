# handler-slack

A Sensu event handler that posts check results to a Slack channel via an incoming
webhook, colour-coded by the check status.

## Features

- **Slack Notifications**: Posts a formatted attachment to a Slack webhook
- **Status Colours**: Green (OK), amber (WARNING), red (CRITICAL), grey (unknown)
- **Rich Context**: Includes client, address, subscriptions, check name and output

## How it works

The handler reads the Sensu event JSON on **stdin** and its configuration from a
JSON file, then POSTs a message to the configured Slack incoming webhook.

## Configuration

Default config path: `/etc/sensu/conf.d/handler-slack.json`

```json
{
  "slack": {
    "webhook_url": "https://hooks.slack.com/services/XXX/YYY/ZZZ"
  }
}
```

## Usage

Configure it as a Sensu handler; Sensu pipes the event to the handler's stdin:

```bash
echo "$EVENT_JSON" | handler-slack
```

## Status Colours

| Status | Colour |
|--------|--------|
| 0 (OK) | green (`#43ac6a`) |
| 1 (WARNING) | amber (`#f9ba46`) |
| 2 (CRITICAL) | red (`#ea5443`) |
| other | grey (`#9c9990`) |

## Notes

- Uses a Slack **incoming webhook** URL (no token/OAuth).
- The message text uses Slack markdown and includes the check output in a code block.
