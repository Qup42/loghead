# loghead

An open source, self-hosted implementation of the logging backend for Tailscale clients.

## Features

- :white_check_mark: Client logs
- :white_check_mark: WIP: SSH session recording
- :x: Network flow logs (Won't be supported - buy Tailscale instead)

## Running loghead

- [Configure](./docs/config.md) loghead
- Configure the clients
    - SSH session recordings require patches to the control plane. It is planned to get these into [headscale](https://github.com/juanfont/headscale) in the future. For now you can use [this fork](https://github.com/Qup42/headscale/tree/feat/sshSessionRecording).
    - Client logs require configuration of the tailscale daemon. Set the environment variable `TS_LOG_TARGET` to the address of loghead. On systemd systems you might be able to set this in `/etc/default/tailscaled`.
- Run `loghead`

## Documentation

See [`docs/`](docs/)
