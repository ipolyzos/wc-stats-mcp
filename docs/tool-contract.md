# Tool / stdio protocol contract

This document describes the MCP-style stdio server in this repository: transport, JSON-RPC profile, **shared tool input conventions**, **error shape**, deterministic **sorting**, and per-tool behavior.

## Transport

- **Framing**: one JSON value per line (newline-delimited JSON), UTF-8.
- **stdin**: inbound requests only.
- **stdout**: outbound JSON-RPC **responses only** (one JSON object per line). No log or diagnostic output on stdout.
- **stderr**: unstructured logs (timestamps, diagnostics). Clients must not parse stderr as RPC.

## JSON-RPC profile

- **Version**: each request must include `"jsonrpc": "2.0"`.
- **Batching**: not supported; one object per line.
- **Notifications**: if `id` is omitted or JSON `null`, the server does not write a response line for that message (MCP notifications).

## Methods

| Method | Purpose |
|--------|---------|
| `initialize` | Handshake stub returning `protocolVersion`, `capabilities`, `serverInfo`. |
| `tools/list` | Lists registered tools (`name`, `description`, `inputSchema`). |
| `tools/call` | Dispatches by tool `name` with JSON `arguments`. |

### `tools/call` parameters

```json
{
  "name": "server.ping",
  "arguments": {}
}
```

`arguments` may be omitted; it defaults to `{}`.

### `tools/call` success result

The `result` object includes MCP-oriented wrapping:

```json
{
  "structuredContent": { "ok": true }
}
```

(`structuredContent` is the JSON value returned by the tool handler.)

## Error envelope

Errors use a stable object shape on the JSON-RPC `error` field:

| Field | Type | Meaning |
|-------|------|---------|
| `code` | number | Machine-readable code (see table below). |
| `message` | string | Short human-readable summary. |
| `details` | object | Optional structured context (JSON). Omitted when empty. |

### Error codes

| Code | When |
|------|------|
| `-32700` | Parse error — inbound line is not valid JSON for a request object. `details.reason` may contain the decode error. |
| `-32600` | Invalid request — e.g. wrong `jsonrpc` version or missing `method`. |
| `-32601` | Method not found — unknown `method`. `details.method` is set. |
| `-32602` | Invalid params — e.g. `tools/call` params cannot be decoded. `details.reason` may be set; missing tool name uses `details.field` = `"name"`. |
| `-32603` | Internal error — e.g. failure marshaling a tool result. |
| `-32000` | Tool execution failed — handler returned an error or unknown tool name. See below. |

### Tool execution failures (`-32000`)

For any failure from `tools/call` after dispatch (unknown tool, invalid arguments reported by the handler, validation failure, I/O, etc.), the server uses:

- **`message`**: always `tool execution failed`.
- **`details`**: JSON object with:
  - **`tool`** (string): the `name` from `tools/call` params.
  - **`reason`** (string): error text from the handler or registry (often prefixed, e.g. `validate: …`, `arguments: …`).

Example:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32000,
    "message": "tool execution failed",
    "details": { "tool": "unknown.tool", "reason": "unknown tool: \"unknown.tool\"" }
  }
}
```

## Shared tool conventions

Tools that read files share the following **argument** names and semantics (see `tools/list` → `inputSchema`).

### `path` and `root`

- **`path`** (required for file tools): **relative** path under **`root`**. Absolute paths are rejected. The path string must not contain a `..` path element. Resolution uses `internal/validate.ResolveUnderRoot` (symlink-safe containment, regular file only).
- **`root`**: optional directory used as the sandbox for resolving `path`. If omitted or whitespace-only, the server uses the **process current working directory**.

### `max_bytes` (optional)

- **`max_bytes`**: optional **integer**. When greater than zero, the effective maximum size of the **resolved** file for that request is the **minimum** of the server’s configured cap and `max_bytes`. Clients **cannot** raise the limit above the server’s effective maximum (default **25 MiB** unless the binary sets a different `validate.Limits`).
- Omitted or zero means “use the server default cap only.”

### `max_lines` (line-streaming tools)

- **`max_lines`**: maximum number of **lines** read from the file (after opening). Default **1_000_000**; hard maximum **1_000_000**.
- Per-line memory cap for the scanner is **4 MiB** (`textscan.DefaultMaxLineBytes`), not configurable via tool arguments.

### `top_n` (ranked frequency tools)

- **`top_n`**: maximum number of ranked rows returned. Default **20**; maximum **100** (tool-specific defaults may apply).

## Deterministic ordering

Sort keys are compared as **UTF-8 byte sequences** (Go `string` comparison order), **not** Unicode collation tables or locale-aware sorting. Same numeric **`count`** ties break by lexicographic key order.

| Tool | Ordering rule |
|------|----------------|
| `top_terms` | **Count descending**, then **term ascending** (UTF-8 byte order). |
| `unique_values` | Distinct values **ascending** (UTF-8 byte order). |
| `field_frequency` | **Count descending**, then **value ascending** (UTF-8 byte order); truncated to `top_n`. |
| `server.ping` | N/A |
| `count_lines_words_bytes` | N/A (scalar counts) |

## Tool: `server.ping`

| Result |
|--------|
| `{"ok": true}` inside `structuredContent` |

No arguments.

## Tool: `count_lines_words_bytes`

Runs **`wc -l -w -c`** on a validated file.

**Arguments:** `path` (required), `root`, `max_bytes` (optional; see shared conventions).

**`structuredContent`:** `{ "path", "lines", "words", "bytes" }`.

## Tool: `top_terms`

Counts token frequencies in a UTF-8 text file using **pure Go** streaming (`internal/textscan`).

**Arguments:**

| Field | Type | Default | Notes |
|-------|------|---------|--------|
| `path` | string | required | Relative to `root`. |
| `root` | string | server `cwd` | Sandbox directory. |
| `max_bytes` | integer | (omit) | Optional stricter file size cap; see shared conventions. |
| `top_n` | integer | `20` | Maximum `100`. |
| `offset` | integer | `0` | Pagination start index in the sorted results. |
| `min_len` | integer | `3` | Minimum token length in **Unicode runes**; maximum `30`. |
| `case_sensitive` | boolean | `false` | When `false`, each line is normalized with `bytes.ToLower` before tokenization. |
| `max_lines` | integer | `1_000_000` | Hard-capped at `1_000_000`; see shared conventions for line/byte caps. |

**Token rule:** maximal substrings of Unicode letters or numbers (`unicode.IsLetter` / `unicode.IsNumber`); other runes delimit tokens.

**`structuredContent` shape:**

```json
{
  "path": "/resolved/path.txt",
  "top_n": 5,
  "offset": 0,
  "total_unique": 42,
  "has_more": true,
  "next_offset": 5,
  "terms": [{ "term": "example", "count": 12 }]
}
```

**Example `tools/call`:**

```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"top_terms","arguments":{"path":"data.txt","top_n":10,"offset":0,"min_len":3}}}
```

## Tool: `unique_values`

Streams delimited lines from a UTF-8 file (`internal/textscan` + `internal/validate`). Field extraction uses **naive** `strings.Split` on `delimiter` — **not** RFC 4180 CSV (no quoted fields, no escapes). Rows with fewer than `field_index+1` fields are skipped.

**Arguments:**

| Field | Type | Default | Notes |
|-------|------|---------|--------|
| `path` | string | required | Relative to `root`. |
| `field_index` | integer | required | 0-based; maximum `1000`. |
| `root` | string | server `cwd` | Sandbox directory. |
| `max_bytes` | integer | (omit) | Optional stricter file size cap; see shared conventions. |
| `delimiter` | string | `,` | Maximum **3** bytes. |
| `trim_space` | boolean | `true` | Trim ASCII spaces around the extracted field. |
| `max_unique` | integer | `1000` | Stop after this many distinct values; hard max `10_000`. |
| `max_lines` | integer | `1_000_000` | Hard-capped at `1_000_000`; per-line cap **4 MiB** (`textscan.DefaultMaxLineBytes`). |

**`structuredContent` shape:**

```json
{
  "path": "/resolved/path.txt",
  "field_index": 1,
  "values": ["admin", "guest", "user"]
}
```

Sorting: distinct `values` are **ascending** lexicographic UTF-8 byte order.

## Tool: `field_frequency`

Same streaming, path validation, delimiter, and field-index semantics as `unique_values`. Counts occurrences of each field value (after optional trim). Empty strings are **omitted** from counts unless `include_empty` is `true`.

**Arguments:** same as `unique_values` **except** no `max_unique`, **plus**:

| Field | Type | Default | Notes |
|-------|------|---------|--------|
| `top_n` | integer | `20` | Maximum `100`; number of `{value,count}` rows returned. |
| `include_empty` | boolean | `false` | When `true`, count rows where the field is empty after trim. |

(`path`, `root`, `max_bytes`, `delimiter`, `field_index`, `trim_space`, and `max_lines` behave as for `unique_values`.)

**`structuredContent` shape:**

```json
{
  "path": "/resolved/path.txt",
  "field_index": 1,
  "top_n": 20,
  "frequencies": [{ "value": "x", "count": 12 }]
}
```

Sorting: **count descending**, then **value ascending** (UTF-8 byte order), then truncated to `top_n` entries.

## File path validation (read-only tools)

Tools that open paths use `internal/validate`:

- **Relative paths only** — absolute `userPath` values are rejected.
- **No `..` elements** in the user-provided path string.
- **Root containment** after `Join`/`Clean`, then again after **symlink resolution** (`filepath.EvalSymlinks`).
- **Regular files only**.
- **Default max file size: 25 MiB** per resolved file unless overridden server-side; **`max_bytes`** may only tighten this per request.

Stable errors: `ErrInvalidInput`, `ErrAbsolutePath`, `ErrDotDot`, `ErrTraversal`, `ErrNotExist`, `ErrNotRegular`, `ErrFileTooLarge`, `ErrSymlink`. See [security.md](security.md).

## Limits

- JSON-RPC line buffer: **4 MiB** per inbound message.
- `top_terms` line buffer: **4 MiB** per file line (`textscan`); line count capped by `max_lines` (default / hard max **1_000_000**).
- `unique_values` and `field_frequency`: same per-line and `max_lines` caps as above; `field_frequency` holds all distinct values in memory for counting (no `max_unique` cap).

The `cmd/wc-stats-mcp` binary calls `tools.RegisterProduction`, which registers **`count_lines_words_bytes`**, **`field_frequency`**, **`server.ping`**, **`top_terms`**, and **`unique_values`** (`internal/tools.ProductionToolNames`). Tests assert `tools/list` schemas for those names marshal to JSON and round-trip.

## `tools/list` inputSchema (authoritative shapes)

These mirror `internal/tools/catalog.go` (`ToolMeta`). Property types are JSON Schema–like: **`integer`**, **`string`**, **`boolean`**. Omitted keys use handler defaults.

### `server.ping`

```json
{ "type": "object", "properties": {} }
```

### `count_lines_words_bytes`

```json
{
  "type": "object",
  "required": ["path"],
  "properties": {
    "path": { "type": "string" },
    "root": { "type": "string" },
    "max_bytes": { "type": "integer" }
  }
}
```

### `top_terms`

```json
{
  "type": "object",
  "required": ["path"],
  "properties": {
    "path": { "type": "string" },
    "root": { "type": "string" },
    "max_bytes": { "type": "integer" },
    "top_n": { "type": "integer" },
    "offset": { "type": "integer" },
    "min_len": { "type": "integer" },
    "case_sensitive": { "type": "boolean" },
    "max_lines": { "type": "integer" }
  }
}
```

### `unique_values`

```json
{
  "type": "object",
  "required": ["path", "field_index"],
  "properties": {
    "path": { "type": "string" },
    "root": { "type": "string" },
    "max_bytes": { "type": "integer" },
    "delimiter": { "type": "string" },
    "field_index": { "type": "integer" },
    "trim_space": { "type": "boolean" },
    "max_unique": { "type": "integer" },
    "max_lines": { "type": "integer" }
  }
}
```

### `field_frequency`

```json
{
  "type": "object",
  "required": ["path", "field_index"],
  "properties": {
    "path": { "type": "string" },
    "root": { "type": "string" },
    "max_bytes": { "type": "integer" },
    "delimiter": { "type": "string" },
    "field_index": { "type": "integer" },
    "trim_space": { "type": "boolean" },
    "max_lines": { "type": "integer" },
    "top_n": { "type": "integer" },
    "include_empty": { "type": "boolean" }
  }
}
```

`tools/list` also returns a string **`description`** per tool (not shown above).

## End-to-end examples (request → success result)

Illustrative **`result`** objects omit MCP wrapper noise; real stdout lines are single JSON-RPC responses whose **`result`** contains **`structuredContent`** (see above).

### Ping

**Request:**

```json
{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"server.ping","arguments":{}}}
```

**`structuredContent`:**

```json
{ "ok": true }
```

### Count lines/words/bytes

**Request:**

```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"count_lines_words_bytes","arguments":{"path":"sample.txt","root":"/project"}}}
```

**`structuredContent`:**

```json
{ "path": "/project/sample.txt", "lines": 2, "words": 4, "bytes": 28 }
```

### Top terms

**Request:**

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"top_terms","arguments":{"path":"notes.txt","root":"/project","top_n":3,"offset":0,"min_len":2,"case_sensitive":false}}}
```

**`structuredContent`:**

```json
{
  "path": "/project/notes.txt",
  "top_n": 3,
  "offset": 0,
  "total_unique": 2,
  "has_more": false,
  "terms": [
    { "term": "alpha", "count": 4 },
    { "term": "beta", "count": 2 }
  ]
}
```

### Unique values

**Request:**

```json
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"unique_values","arguments":{"path":"rows.csv","root":"/project","field_index":1,"delimiter":","}}}
```

**`structuredContent`:**

```json
{
  "path": "/project/rows.csv",
  "field_index": 1,
  "values": ["east", "north", "south"]
}
```

### Field frequency

**Request:**

```json
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"field_frequency","arguments":{"path":"rows.csv","root":"/project","field_index":0,"top_n":5,"include_empty":false}}}
```

**`structuredContent`:**

```json
{
  "path": "/project/rows.csv",
  "field_index": 0,
  "top_n": 5,
  "frequencies": [
    { "value": "user-a", "count": 10 },
    { "value": "user-b", "count": 3 }
  ]
}
```

### Tool error (validation)

**Request** (absolute `path`):

```json
{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"top_terms","arguments":{"path":"/etc/passwd"}}}
```

**Response `error`:**

```json
{
  "code": -32000,
  "message": "tool execution failed",
  "details": {
    "tool": "top_terms",
    "reason": "validate: absolute paths are not allowed"
  }
}
```
