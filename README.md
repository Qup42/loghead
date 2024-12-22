# loghead

An open source, self-hosted implementation of the logging backend for Tailscale clients.

## Features

- :white_check_mark: [Client logs](./docs/client_logs.md)
- :construction: [Client Metrics aggregation](./docs/client_metrics.md)
- :construction: [SSH session recording](./docs/ssh_recorder.md)
- :x: Network flow logs (Won't be supported - buy Tailscale instead)

## Running loghead

- [Configure](./docs/config.md) loghead
- Configure the components
    - See the [SSH session recording documentation](./docs/ssh_recorder.md) for further information. TL;DR: Patches to the control plane are required.
    - [Client logs](./docs/client_logs.md): configuration of the tailscale daemon is required. Set the environment variable `TS_LOG_TARGET` to the address of loghead. On systemd systems you might be able to set this in `/etc/default/tailscaled`.
    - See the [Client metrics documentation](./docs/client_metrics.md) for further information.
- Run `loghead` natively or in a [container](./docs/docker.md).

## Documentation

The documentation can be found in the [`docs/`](./docs/) folder.
