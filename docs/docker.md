# docker

It is also possible to deploy loghead in docker. See an example deployment below that uses docker compose.

Note that this deployment does not use the default configuration.
The directory for tailscale's state has been set.

```yaml
version: "3.9"

services:
  loghead:
    image: ghcr.io/qup42/loghead/loghead
    volumes:
      - ./config.yaml:/etc/loghead/config.yaml
        # persist tsnet's state. Without this a valid authkey is required on every start.
        # When the state is persistet an authkey is only required when the node is expired.
      - ./state:/ssh-state
      - ./client-logs:/logs
      - ./ssh-sessions:/recordings
    networks:
      - traefik
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.loghead.rule=Host(`loghead.foo.bar`) && PathPrefix(`/c/`)"
      - "traefik.http.routers.loghead.entrypoints=web"
      - "traefik.http.routers.loghead.tls=true"
      - "traefik.http.routers.loghead.tls.certresolver=le"
      - "traefik.http.services.loghead.loadbalancer.server.port=5678"
    restart: always

networks:
  traefik:
    external: true
    name: network_traefik
```
