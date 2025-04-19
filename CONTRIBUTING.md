# Contributing

## Updating dependencies

```bash
go get -u ./...
go mod tidy
```

> [!CAUTION] 
> The version of the `gvisor.dev/gvisor` must be from the `go` branch. An update will most likely break this. If in
> doubt just copy the package's version from tailscale's `go.mod`.