# wc-stats-mcp

A **read-only** [Model Context Protocol](https://modelcontextprotocol.io/) server over **stdio** that exposes text statistics tools backed by coreutils-style behavior (`wc`, `sort`, `uniq`, and optionally `cut`), with strict limits and deterministic locale settings.

## Guarantees

- **Read-only**: the server does not write to the filesystem or perform outbound network I/O for tool execution.
- **No shell**: subprocesses are started with explicit argv only (`exec.CommandContext`); no `sh -c`.
- **Determinism**: child processes run with `LC_ALL=C` and `LANG=C` for stable collation and tool output where applicable.
- **Allowlisted binaries**: only `wc`, `sort`, `uniq`, and `cut` may be executed (resolved via `PATH`).

## Supported platforms

- **Linux** and **macOS**: primary targets.
- **Windows**: best-effort (path resolution and coreutils availability vary).

## Build

```bash
go build -o wc-stats-mcp ./cmd/wc-stats-mcp
```

With version metadata (optional):

```bash
go build -o wc-stats-mcp \
  -ldflags "-X wc-stats-mcp/internal/version.Version=0.1.0 -X wc-stats-mcp/internal/version.Commit=$(git rev-parse --short HEAD) -X wc-stats-mcp/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  ./cmd/wc-stats-mcp
```

## Run

The server speaks **newline-delimited JSON-RPC 2.0** on stdio (see [docs/architecture.md](docs/architecture.md)).

```bash
./wc-stats-mcp
```

```bash
./wc-stats-mcp --version
```

## Tool catalog

| Tool | Description |
|------|-------------|
| `count_lines_words_bytes` | `wc` over stdin (`lines`, `words`, `bytes`, `chars`, `all`). |
| `top_terms` | Tokenize in Go, then `sort` → `uniq -c` → `sort -nr`. |
| `unique_values` | Extract a delimited field in Go; unique values with optional lex sort. |
| `field_frequency` | Extract a field in Go, then `sort` → `uniq -c` → `sort -nr`. |

Shared limits and environment knobs are documented in [docs/tools/README.md](docs/tools/README.md) and [docs/security.md](docs/security.md).

## Configuration (environment)

| Variable | Purpose |
|----------|---------|
| `WC_STATS_MCP_TIMEOUT` | Default subprocess timeout (e.g. `2s`). |
| `WC_STATS_MCP_MAX_STDIN_BYTES` | Max stdin bytes per subprocess. |
| `WC_STATS_MCP_MAX_STDOUT_BYTES` | Max captured stdout bytes. |
| `WC_STATS_MCP_MAX_STDERR_BYTES` | Max captured stderr bytes. |
| `WC_STATS_MCP_MAX_TEXT_BYTES` | Max `text` argument size for tools. |
| `WC_STATS_MCP_MAX_LINES` | Max logical lines processed for line-oriented tools. |
| `WC_STATS_MCP_MAX_TOKENS` | Max tokens emitted by tokenizers. |
| `WC_STATS_MCP_MAX_TOP_N` | Upper bound for `top_n`-style parameters. |

## Documentation

- [Architecture](docs/architecture.md)
- [Security](docs/security.md)
- [Client contract](docs/client-contract.md)
- [Tools](docs/tools/README.md)
- [Releasing](docs/releasing.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for branch naming, **docs-before-commit**, test-before-push, PRs to `develop`, and releases from `main` only.

## License

Apache-2.0 — see [LICENSE](LICENSE).
