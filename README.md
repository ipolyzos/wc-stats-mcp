# wc-stats-mcp

[![CI](https://github.com/ipolyzos/wc-stats-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/ipolyzos/wc-stats-mcp/actions/workflows/ci.yml)
[![Docs](https://github.com/ipolyzos/wc-stats-mcp/actions/workflows/docs.yml/badge.svg)](https://github.com/ipolyzos/wc-stats-mcp/actions/workflows/docs.yml)

**v0.1.0** — a **read-only** [Model Context Protocol](https://modelcontextprotocol.io/) server over **stdio** for text statistics. It validates all file paths under a configurable root, applies strict size and line limits for streaming tools, and uses coreutils **`wc`** only for line/word/byte counts.

## Requirements

- **Go 1.23+** if you build from source.
- **`wc`** on `PATH` for the `count_lines_words_bytes` tool (typical on Linux/macOS).

## Install

### From source

```bash
git clone https://github.com/ipolyzos/wc-stats-mcp.git
cd wc-stats-mcp
go build -o wc-stats-mcp ./cmd/wc-stats-mcp
```

Install into `GOBIN` or `GOPATH/bin` if you prefer:

```bash
go install ./cmd/wc-stats-mcp
```

### From GitHub Releases

Prebuilt **Linux** and **macOS** binaries (**amd64** / **arm64**) are attached to each **`v*`** tag as `.tar.gz` archives plus `checksums.txt`. See [docs/releases.md](docs/releases.md).

## Run

The process speaks **newline-delimited JSON-RPC 2.0** on stdio: one JSON object per line on stdin; responses on stdout; logs on stderr.

```bash
./wc-stats-mcp --version
./wc-stats-mcp
```

Configure your MCP host to launch this command with a working directory (or explicit `root` in tool arguments) that contains the files you want tools to read. Every file path passed to a tool must be **relative** to that root (no `..` segments). See [docs/tool-contract.md](docs/tool-contract.md) and [docs/security.md](docs/security.md).

## Tools (summary)

| Tool | Purpose |
|------|---------|
| `server.ping` | Liveness check. |
| `count_lines_words_bytes` | `wc -l -w -c` on a validated file. |
| `top_terms` | Token frequencies (Unicode letters/digits), top N by count. |
| `unique_values` | Distinct values of one delimited field per line (naive split; not CSV-safe). |
| `field_frequency` | Per-field value counts, top N by frequency. |

---

### `count_lines_words_bytes`

Runs **`wc -l -w -c`** after **path validation** (`internal/validate`): path must be **relative** to **`root`** (default: server **cwd**), no `..`, symlink-safe containment, **regular file** only, default **25 MiB** max size (tightenable with **`max_bytes`**).

| Argument | Required | Description |
|----------|----------|-------------|
| `path` | yes | Relative path under `root`. |
| `root` | no | Sandbox root; defaults to process cwd. |
| `max_bytes` | no | Optional stricter max file size in bytes (cannot exceed server cap). |

**Example NDJSON** (one line on stdin):

```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"count_lines_words_bytes","arguments":{"path":"data/sample.txt","root":"/srv/project"}}}
```

---

### `top_terms`

Streaming tokenizer over a validated file. Tokens = Unicode **letters and numbers** only; shorter than `min_len` runes are skipped. Sorted **count descending**, then **term ascending** (UTF-8 byte order).

| Argument | Default | Limits |
|----------|---------|--------|
| `path` | (required) | Relative under `root`. |
| `root` | process `cwd` | — |
| `max_bytes` | (omit) | Optional stricter file size cap. |
| `top_n` | `20` | max `100` |
| `min_len` | `3` | max `30` (runes) |
| `case_sensitive` | `false` | If `false`, `bytes.ToLower` per line before tokenizing. |
| `max_lines` | `1_000_000` | Hard max `1_000_000`. |

```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"top_terms","arguments":{"path":"notes/sample.txt","root":"/data/repo","top_n":5,"min_len":3,"case_sensitive":false}}}
```

---

### `unique_values`

Splits each line on **`delimiter`** (`strings.Split`; **not CSV-safe**), reads **`field_index`** (0-based), optional ASCII trim, returns **sorted distinct** values. Short rows skipped.

| Argument | Default | Limits |
|----------|---------|--------|
| `path` | (required) | Relative under `root`. |
| `field_index` | (required) | `0` … `1000`. |
| `root` | process `cwd` | — |
| `max_bytes` | (omit) | Optional stricter file size cap. |
| `delimiter` | `,` | Max **3** bytes. |
| `trim_space` | `true` | — |
| `max_unique` | `1000` | Hard max `10_000` (scan may stop early). |
| `max_lines` | `1_000_000` | Hard max `1_000_000`. |

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"unique_values","arguments":{"path":"data/rows.csv","root":"/project","field_index":1,"delimiter":","}}}
```

---

### `field_frequency`

Same delimiter/field rules as `unique_values`. Counts values; returns **top `top_n`** with **count desc**, then **value asc**. Empty fields skipped unless **`include_empty`** is `true`.

| Argument | Default | Limits |
|----------|---------|--------|
| `path` | (required) | Relative under `root`. |
| `field_index` | (required) | `0` … `1000`. |
| `root` | process `cwd` | — |
| `max_bytes` | (omit) | Optional stricter file size cap. |
| `delimiter` | `,` | Max **3** bytes. |
| `trim_space` | `true` | — |
| `max_lines` | `1_000_000` | Hard max `1_000_000`. |
| `top_n` | `20` | Max `100`. |
| `include_empty` | `false` | — |

```json
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"field_frequency","arguments":{"path":"data/rows.csv","root":"/project","field_index":0,"top_n":10,"include_empty":false}}}
```

---

### `server.ping`

Health check → `{"ok": true}` in `structuredContent`.

```json
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"server.ping","arguments":{}}}
```

## Documentation

| Doc | Topic |
|-----|--------|
| [docs/tool-contract.md](docs/tool-contract.md) | JSON-RPC, errors, limits, per-tool contract |
| [docs/security.md](docs/security.md) | Path policy, subprocess rules, read-only posture |
| [docs/architecture.md](docs/architecture.md) | Packages, stdio MCP, why pure-Go scanning |
| [docs/troubleshooting.md](docs/troubleshooting.md) | Common failures and fixes |
| [docs/releases.md](docs/releases.md) | Tagging and GitHub Releases |
| [docs/development.md](docs/development.md) | Local build and test |

## Contributing

Branches **`main`** and **`develop`**, PR-only workflow, and conventions: [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache-2.0 — see [LICENSE](LICENSE).
