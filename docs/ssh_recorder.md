# SSH session recording

[SSH session recording](https://tailscale.com/kb/1246/tailscale-ssh-session-recording) is a feature that allows you to stream logs of [Tailscale SSH](https://tailscale.com/kb/1193/tailscale-ssh) to another node. This can be used to collect recordings of the SSH sessions in a centralized place.

## Prerequisites

You need a control server that processes the `recorder` entries of SSH ACL rules.
[headscale](https://github.com/juanfont/headscale) currently does not support this. I would like to get support for this into headscale.
Support for this will (if at all) most likely only be added in the medium to long term[^1].
Once this seems likely there will be up-to-date forks/branches that contain this feature.
Until then, you have to patch headscale yourself. Have a look at [this PR](https://github.com/juanfont/headscale/pull/1820) for the required changes.


## Configuration

You have to [configure SSH session recording](https://tailscale.com/kb/1246/tailscale-ssh-session-recording#turn-on-session-recording-in-acls) in you control plane's ACL rules.

> [!NOTE]
> SSH session recording uses one long-lived HTTP connection per SSH session.
> The endpoints for SSH session recording therefore do not enforce any timeouts on the HTTP connections.
> This may be vulnerable to a DOS if this component is deployed publicly.

## Usage

The SSH session recorder writes the recorded sessions in the recording directory.
The recordings are in the [`asciinema`](https://asciinema.org/) format.
The recordings are saved as `<stablenodeid>/<RFC 3339 timestamp>.cast` under the configured recordings directory.
`<stablenodeid>` is the stable node id of the accessing node.

[^1]: Follow [juanfont/headscale#1793](https://github.com/juanfont/headscale/issues/1793) for updates on adding this feature. See [this comment](https://github.com/juanfont/headscale/pull/1820#issuecomment-2505640781) for timeframe and whether this has the possibility of being added.