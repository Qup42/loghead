# SSH session recording

## Prerequisites

You need a control server that processes the `recorder` entries of SSH ACL rules.
[headscale](https://github.com/juanfont/headscale) does not support this yet. It is planed to merge this into headscale.
Until then you can use [this fork](https://github.com/Qup42/headscale/tree/feat/sshSessionRecording) that is patched.

## Configuration

You have to [configure SSH session recording](https://tailscale.com/kb/1246/tailscale-ssh-session-recording#turn-on-session-recording-in-acls) in you control plane's ACL rules.

## Usage

The SSH session recorder writes the recorded sessions in the the recording directory.
The recordings are in the [`asciinema`](https://asciinema.org/) format.
The recordings are saved as `<stablenodeid>/<RFC 3339 timestamp>.cast` under the configured recordings directory.
`<stablenodeid>` is the stable node id of the accessing node.
