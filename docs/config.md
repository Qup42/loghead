# Config

loghead looks for a config file named `config.yaml`[^1] in `./`, `~/.loghead/` and `/etc/loghead/`.

## Listeners

The services (Client logs, SSH session recording, and Client metrics aggregation) can be exposed in one of two ways
- `type: plain` - publicly available directly over the network
- `type: tsnet` - only available over tailscale as a [tsnet](https://tailscale.com/blog/tsnet-virtual-private-services) service
When a service is exposed as a tsnet service an [AuthKey](https://tailscale.com/kb/1085/auth-keys) has to be provided.

## Default config

The default config is shown below.

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
    tsnet:
      authKey: ""
      controllURL: "https://controllplane.tailscale.com"
      hostname: "" # hostname of the tsnet service
      dir: "/tsnet-state/loghead" # where state is stored


ssh_recorder:
  dir: "./recordings"
  listener:
    type: "tsnet" # "plain" or "tsnet"
    addr: "0.0.0.0"
    port: "80" # `80` is used by both headscale and tailscale
    tsnet:
      authKey: ""
      controllURL: "https://controllplane.tailscale.com"
      hostname: "" # hostname of the tsnet service
      dir: "/tsnet-state/ssh_recorder" # where state is stored

node_metrics:
  listener:
    type: "plain" # "plain" or "tsnet"
    addr: "0.0.0.0"
    port: "5679"
    tsnet:
      authKey: ""
      controllURL: "https://controllplane.tailscale.com"
      hostname: "" # hostname of the tsnet service
      dir: "/tsnet-state/node_metrics" # where state is stored
```

[^1]: Only the name `config` is fixed. All formats supported by viper (JSON, YAML, TOML, HCL, envfile and Java properties config files) are supported. YAML is the recommended format.