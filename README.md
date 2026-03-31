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

### `top_terms`

Pure-Go streaming tokenizer over a validated file (`internal/validate` + `internal/textscan`). Tokens are Unicode **letters and numbers** only (see `textscan.ForEachNonLetterDigitToken`). Shorter tokens than `min_len` runes are skipped. Results are sorted by **count descending**, then **term ascending** (UTF-8 byte order).

| Argument | Default | Limits |
|----------|---------|--------|
| `path` | (required) | Relative under `root`. |
| `root` | process `cwd` | — |
| `top_n` | `20` | max `100` |
| `min_len` | `3` | max `30` (Unicode runes per token) |
| `case_sensitive` | `false` | If `false`, fold with `bytes.ToLower` on each line before tokenizing. |
| `max_lines` | `1_000_000` | Hard max `1_000_000` lines scanned. |

**Structured result:**

```json
{
  "path": "/absolute/resolved/file.txt",
  "top_n": 3,
  "terms": [
    { "term": "apple", "count": 3 },
    { "term": "banana", "count": 2 }
  ]
}
```

**Example `tools/call` (one NDJSON line):**

```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"top_terms","arguments":{"path":"notes/sample.txt","root":"/data/repo","top_n":5,"min_len":3,"case_sensitive":false}}}
```

### `server.ping`

Health check; returns `{"ok": true}` inside `structuredContent`.

## Planned MCP tools

| Tool | Purpose (planned) |
|------|-------------------|
| `unique_values` | Delimited field extraction and unique values (deterministic sort optional). |
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
