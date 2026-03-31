# Local development

## Prerequisites

- **Go** 1.23+ (`go version`; CI uses **1.23.12** via `go.mod` `toolchain`).

## Clone and branch

```bash
git clone https://github.com/ipolyzos/wc-stats-mcp.git
cd wc-stats-mcp
git checkout develop
git pull
git checkout -b feature/your-change
```

## Build

```bash
go build -o wc-stats-mcp ./cmd/wc-stats-mcp
```

When the `cmd/wc-stats-mcp` tree is not present yet, `go build ./...` validates packages only.

## Test

```bash
go test ./...
go vet ./...
```

On Linux, when touching concurrency:

```bash
go test -race ./...
```

## Run (stdio MCP)

After a successful build:

```bash
./wc-stats-mcp --version
./wc-stats-mcp
```

The server is intended to speak MCP over stdio; exact invocation will match the implementation in `README.md` as it lands.

Path validation and related security rules are documented in [security.md](security.md).

High-level design (stdio MCP, pure-Go scanning vs subprocesses) is in [architecture.md](architecture.md).

## Contributing workflow

See [CONTRIBUTING.md](../CONTRIBUTING.md): PRs target **`develop`**, releases come from **`main`** via PR.

## Troubleshooting

See [troubleshooting.md](troubleshooting.md) for common errors (`validate:`, `wc`, path rules, scan limits).
