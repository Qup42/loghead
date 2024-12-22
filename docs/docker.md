# docker

It is also possible to deploy loghead in docker. See an example deployment below that uses docker compose and traefik.

> [!NOTE]
> This deployment does not use the default configuration.

> [!TIP]
> By persisting the state directory of any tsnet listeners you use, an authkey is only required for the first start or if the login expires.

```yaml
version: "3.9"

services:
  loghead:
    image: ghcr.io/qup42/loghead/loghead
    volumes:
      - ./config.yaml:/etc/loghead/config.yaml
        # Persist the state of the tsnet listener for SSH session recording
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
