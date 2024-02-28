# Client Logs

## Configuration

Assume that your loghead instance is available at `https://loghead.foo.bar`.
You have to set the environment variable `TS_LOG_TARGET="https://loghead.foo.bar"` for all tailscale daemons (client devices that should send their logs to loghead.
On systemd systems you might be able to set this in `/etc/default/tailscaled`.

## Processors

Three processors are available to process the client logs:
- `filelogger`
- `metrics`
- `forward`
- `hostinfo`

### `filelogger`

The log lines (json objects) of a node are written to a file. The logs are written one json object per line. The logs of a node are written to a separate file per node. The file's name is the node's private id.

### `metrics`

The logs messages also contain some client metrics. This processor parses these metrics and exposes them in the prometheus format. The metrics are available at `/metrics`.

### `forward`

The logs are forwarded to another host. The tailscale daemon only sends the logs to one location. With this processor you can use `loghead` but still have the logs available in the Tailscale management interface. To forward the logs to Tailscale set the forward addr to `http://log.tailscale.io`.

### `hostinfo`

Some info about the host (os, arch, ...) is sent as part of the client logs. This processor logs this information to the console.
