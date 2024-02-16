# loghead

An open source, self-hosted implementation of the logging backend for Tailscale clients.

## Features

- :white_check_mark: Client logs
- :warning: WIP: SSH session recording
- :x: Network flow logs (Won't be supported - buy Tailscale instead)

## Running loghead

Simply execute `loghead`. **Additional configuration of the clients is required.**

### Client logs

Pass the client logs address to the tailscale daemon via the `TS_LOG_TARGET` environment variable.
It can be defined in `/etc/default/tailscaled` with `TS_LOG_TARGET=https://foo.bar` on my system.

### SSH session recording

The control plane must send the address of the recorder to the client. This requires patching of the control plane (e.g. [headscale](https://github.com/juanfont/headscale)).

## Documentation

See [`docs/`](docs/)
