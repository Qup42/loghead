# Config
The default config is shown below. The config options are relatively self-explanatory.

```yaml
log:
  level: "info"
  format: "text"

loghead:
  listener:
    type: "plain" # "plain" or "tsnet"
    addr: "0.0.0.0"
    port: "5678"
#    tsnet:
#      authKey: ""
#      controllURL: "https://controllplane.tailscale.com"

processors:
  # write logs to a one file per node
  filelogger: true
  # expose the metrics contained in the logs in the prometheus format
  metrics: false
  # log the node's host info to the console
  hostinfo: false

filelogger:
  dir: "./logs"

ssh_recorder:
  dir: "./recordings"
  listener:
    type: "tsnet" # "plain" or "tsnet"
    addr: "0.0.0.0"
    port: "80"
#    tsnet:
#      authKey: ""
#      controllURL: "https://controllplane.tailscale.com"
```
