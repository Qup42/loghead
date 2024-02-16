## Client Logs

Three processors are available to process the client logs:
- `filelogger`
- `metrics`
- `hostinfo`

## `filelogger`

The log lines (json objects) of a node are written to a file. The logs are written one json object per line. The logs of a node are written to a separate file per node. The file's name is the node's private id.

## `metrics`

The logs messages also contain some client metrics. This processor parses these metrics and exposes them in the prometheus format. The metrics are available at `/metrics`.

## `hostinfo`

Some info about the host (os, arch, ...) is sent as part of the client logs. This processor logs this information to the console.
