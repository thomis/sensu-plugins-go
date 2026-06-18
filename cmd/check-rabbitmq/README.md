# check-rabbitmq

A Sensu check plugin that verifies a RabbitMQ virtual host is healthy using the
management API's aliveness test (`/api/aliveness-test/<vhost>`), which declares a
queue, publishes and consumes a message, then deletes the queue.

## Features

- **Aliveness Test**: Confirms a vhost can declare/publish/consume/delete
- **Virtual Host Scoped**: Check a specific vhost
- **Authenticated**: Uses HTTP basic auth against the management API
- **Configurable Timeout**: Set the HTTP request timeout

## Usage

```bash
check-rabbitmq [OPTIONS]
```

### Options

- `-h, --host` - Host (default: `localhost`)
- `-P, --port` - Management API port (default: `15672`)
- `-v, --vhost` - Virtual host, URL-encoded (default: `%2F`, i.e. `/`)
- `-u, --user` - User (default: `guest`)
- `-p, --password` - Password (default: `guest`)
- `-t, --timeout` - HTTP timeout in seconds (default: `10`)

## Examples

```bash
# Check the default vhost on a local node
check-rabbitmq

# Check a named vhost with credentials
check-rabbitmq -h mq.example.com -u monitor -p secret -v my-vhost

# The default vhost "/" must be URL-encoded as %2F
check-rabbitmq -h mq.example.com -v %2F
```

## Exit Codes

- **0 (OK)**: Aliveness test returned `ok`
- **1 (WARNING)**: Aliveness test did not return `ok` (e.g. vhost/object not found)
- **3 (ERROR)**: Request failed (host unreachable, timeout, auth error, etc.)

## Output Examples

```
CheckRabbitMQ OK: RabbitMQ server is alive
CheckRabbitMQ WARNING: Object Not Found
CheckRabbitMQ ERROR: Get "http://mq.example.com:15672/api/aliveness-test/%2F": dial tcp: ...
```

## Use Cases

- **Broker Health**: Confirm a vhost can actually pass messages end-to-end
- **Per-Vhost Monitoring**: Validate the specific vhost an application uses
- **Credential Validation**: Verify management API credentials work

## Notes

- Requires the RabbitMQ **management plugin** to be enabled.
- The vhost must be URL-encoded (the default vhost `/` becomes `%2F`).
- Connects over plain `http`; TLS is not currently supported.
