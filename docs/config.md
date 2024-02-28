# Config

loghead looks for a config file in `./`, `~/.loghead/` and finally in `/etc/loghead/`.

## Listeners

The services (client logs and SSH session recording) can be exposed in one of two ways
- publicly available directly over the network
- only available over tailscale as a [tsnet](https://tailscale.com/blog/tsnet-virtual-private-services) service
When a service is exposed as an tsnet service an [AuthKey](https://tailscale.com/kb/1085/auth-keys) has to be provided.

## Default config

The default config is shown below. The config options are relatively self-explanatory.

```yaml
log:
  level: "info"
  format: "text"

loghead:
  processors:
    # write logs to a one file per node
    filelogger:
      enabled: true
      dir: "./logs"
    # forward the logs to another logtail instance
    forward:
      enabled: false
      addr: "https://log.tailscale.io"
    # expose the metrics contained in the logs in the prometheus format
    metrics: true
    # log the node's host info to the console
    hostinfo: false
  listener:
    type: "plain" # "plain" or "tsnet"
    addr: "0.0.0.0"
    port: "5678"
#    tsnet:
#      authKey: ""
#      controllURL: "https://controllplane.tailscale.com"
#      hostname: "" # hostname of the tsnet service
#      dir: "" # where state is stored


ssh_recorder:
  dir: "./recordings"
  listener:
    type: "tsnet" # "plain" or "tsnet"
    addr: "0.0.0.0"
    port: "80" # `80` is used by both headscale and tailscale
#    tsnet:
#      authKey: ""
#      controllURL: "https://controllplane.tailscale.com"
#      hostname: "" # hostname of the tsnet service
#      dir: "" # where state is stored
```
