# CHANGELOG

## Unreleased

- fix: detection if ts is up
- style: bubble up and handle errors more consistently
- fix: endpoints return code 500 for fatal errors
- fix: bubble up errors that occour when starting a server
- feat: can set ts hostname and state dir via config
- feat: write ts logs to logs as level trace

## 0.0.3 (2024-02-25)

- refactor: generalize HTTP listener config loading
- refactor: generalize HTTP listener creation
- feat!: add client log collection as a tsnet listener
- docs: add BSD-3-Clause license
- feat: tsnet listener waits for ts to be ready before starting

## 0.0.2 (2024-02-22)

- Add a forwarder for the client logs
- Change default port for SSH session recording to 80
- Add SSH session recording as a tsnet listener

## 0.0.1 (2024-02-20)

- Initial release
