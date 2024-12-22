# Client Logs

Tailscale agents send their client logs to a central log server[^1].
This loghead component enables you to collect and process these logs on your own servers.

## Configuration

To use the Client Logs component some configuration is required.
Assume that the Client Logs component is available at `https://loghead.foo.bar`.
You have to set the environment variable `TS_LOG_TARGET="https://loghead.foo.bar"` for all tailscale agents for which the logs should be collected.

> [!TIP]
> If tailscale is running as a systemd service `TS_LOG_TARGET` can be set in `/etc/default/tailscaled`.

## Processors

The Client Logs component by default only receives the logs but do nothing with them. Four processors are available to process the logs:
- [`filelogger`](#filelogger)
- [`metrics`](#metrics)
- [`forward`](#forward)
- [`hostinfo`](#hostinfo)

### `filelogger`

The received logs (which are json objects) are written to a file. The logs are written one json object per line. The logs are written to a separate files for each instance. The file's name is the instances' [private id](https://github.com/tailscale/tailscale/blob/main/logtail/api.md#instances).

### `metrics`

The log messages sometimes also contain client metrics. This processor parses the metrics send in log messages and exposes them in the prometheus format. The metrics are available at the same endpoint as the Client Logs under the path `/metrics`. (So `https://loghead.foo.bar/metrics` in the example.)

### `forward`

The logs are forwarded to another host. The tailscale agents only send the logs to one location. You can use this processor to process the logs with `loghead` but still have the logs available in the Tailscale management interface. To do this forward the logs to `http://log.tailscale.io`.

### `hostinfo`

Some info about the host (os, arch, ...) is sent as part of the client logs. This processor logs this information to the console.

[^1]: [Tailscale KB: Logging Overview](https://tailscale.com/kb/1011/log-mesh-traffic)