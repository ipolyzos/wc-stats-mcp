# Local development

## Prerequisites

- Go **1.22+**

## Commands

```bash
make build    # produces ./wc-stats-mcp
make test     # go test ./...
make fmt      # go fmt ./...
make lint     # go vet ./...
```

Or directly:

```bash
go build -o wc-stats-mcp ./cmd/wc-stats-mcp
go test ./...
```

## Run (stdio)

```bash
./wc-stats-mcp
```

Protocol assumptions (framing, methods, errors) are documented in [tool-contract.md](tool-contract.md).

## Version info

```bash
go run ./cmd/wc-stats-mcp --version
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md): PRs target `develop`.
