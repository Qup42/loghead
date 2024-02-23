## SSH session recording

The SSH session recorder writes the recorded sessions in the the recording directory. The recordings are in the [`asciinema`](https://asciinema.org/) format. The files have the format `ssh-session-{ns timestamp}-{random id}.cast`.
The recorder can listen directly or over tailscale as a [tsnet](https://tailscale.com/blog/tsnet-virtual-private-services) service.
When the recorder is used over tailscale an [AuthKey](https://tailscale.com/kb/1085/auth-keys) has to be provided.
