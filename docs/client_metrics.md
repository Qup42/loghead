# Client metrics aggregation

[Client metrics](https://tailscale.com/kb/1482/client-metrics) is a feature for exposing metrics about the individual agents at the agents.
This component can aggregate the metrics from agents and exposes them under a single endpoint in the prometheus format.

## Configuration

- Enable the [web interface](https://tailscale.com/kb/1325/device-web-interface) on all agents to be monitored with `tailscale set --webclient`.
- Specify the addresses of all agents from which the metrics should be aggregated in the `targets` field.
- The tailscale daemon running on the node must have access to port `5252` of the agents from which to aggregate metrics.
  In the future it will be sufficient that the `tsnet` service of the listener has access to port `5252` of the agents from which to aggregate metrics.