# wc-stats-mcp (WIP)

**Work in progress:** a **read-only** [Model Context Protocol](https://modelcontextprotocol.io/) server over **stdio** for text statistics (wrapping coreutils such as `wc`, `sort`, `uniq`, and optionally `cut`), with strict limits and deterministic locale where subprocesses are used.

## Implemented MCP tools

### `count_lines_words_bytes`

Runs **`wc -l -w -c`** on a file after **path validation** (`internal/validate`): path must be **relative** to a **root** directory (default: server process **current working directory**), with no `..` segments, symlink-safe containment, **regular file** only, and default **25 MiB** max file size.

| Argument | Required | Description |
|----------|----------|-------------|
| `path` | yes | Relative path under `root`. |
| `root` | no | Directory that bounds resolution; defaults to `cwd`. |

**Structured result:**

```json
{
  "path": "/absolute/resolved/path.txt",
  "lines": 10,
  "words": 42,
  "bytes": 256
}
```

**Example NDJSON request** (one line on stdin):

```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"count_lines_words_bytes","arguments":{"path":"data/sample.txt","root":"/srv/project"}}}
```

**Example success line** (stdout; `result` abbreviated):

```json
{"jsonrpc":"2.0","id":1,"result":{"structuredContent":{"path":"/srv/project/data/sample.txt","lines":2,"words":4,"bytes":28}}}
```

### `unique_values`

Streams a validated text file line-by-line (`internal/textscan`), splits each line with **`strings.Split(line, delimiter)`** (default delimiter **`,`**, max **3 bytes**). **Not CSV-aware:** no quoted fields, no escape rules—delimiters inside “fields” will split incorrectly.

| Argument | Default | Notes |
|----------|---------|--------|
| `path` | required | Relative under `root`. |
| `root` | `cwd` | — |
| `delimiter` | `,` | Max length 3 (bytes). |
| `field_index` | required | **0-based** column; max **1000**. |
| `trim_space` | `true` | `strings.TrimSpace` on the field. |
| `max_unique` | `1000` | Hard max **10_000**; scanning stops once this many **distinct** values were seen (in file order). |
| `max_lines` | `1_000_000` | Hard max **1_000_000**. |

Rows with **too few fields** (missing `field_index`) are **skipped**. Returned `values` are sorted by **`sort.Strings`**: lexicographic order of UTF-8 byte sequences (Go’s `<` on strings).

**Structured result:**

```json
{
  "path": "/resolved/file.csv",
  "field_index": 1,
  "values": ["admin", "guest", "user"]
}
```

**Example `tools/call`:**

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"unique_values","arguments":{"path":"data/rows.txt","root":"/srv/app","field_index":2,"delimiter":"|","max_unique":500}}}
```

### `server.ping`

Health check; returns `{"ok": true}` inside `structuredContent`.

## Planned MCP tools

| Tool | Purpose (planned) |
|------|-------------------|
| `top_terms` | Token frequency: tokenize in Go, then `sort` → `uniq -c` → `sort -nr`. |
| `field_frequency` | Per-field value counts via sort / uniq pipeline. |

## Protocol & security

- Framing and JSON-RPC assumptions: [docs/tool-contract.md](docs/tool-contract.md)
- Subprocess and path policy: [docs/security.md](docs/security.md)

## Local development

See [docs/development.md](docs/development.md) for build, test, and run commands (`make test`, etc.).

## Contributing

Branches **`main`** and **`develop`**, PR-only workflow, branch prefixes, and **test-before-PR** are described in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache-2.0 — see [LICENSE](LICENSE).
