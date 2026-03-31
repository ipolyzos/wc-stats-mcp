# Tool / stdio protocol contract (skeleton)

This document describes the assumptions implemented by the minimal MCP-style stdio server in this repository.

## Transport

- **Framing**: one JSON value per line (newline-delimited JSON), UTF-8.
- **stdin**: inbound requests only.
- **stdout**: outbound JSON-RPC **responses only** (one JSON object per line). No log or diagnostic output on stdout.
- **stderr**: unstructured logs (timestamps, diagnostics). Clients must not parse stderr as RPC.

## JSON-RPC profile

- **Version**: each request must include `"jsonrpc": "2.0"`.
- **Batching**: not supported; one object per line.
- **Notifications**: if `id` is omitted or JSON `null`, the server does not write a response line for that message (MCP notifications).

## Methods (skeleton)

| Method | Purpose |
|--------|---------|
| `initialize` | Handshake stub returning `protocolVersion`, `capabilities`, `serverInfo`. |
| `tools/list` | Lists registered tools (`name`, placeholder `description`, minimal `inputSchema`). |
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
| `code` | number | Machine-readable code (JSON-RPC range + app-specific). |
| `message` | string | Short human-readable summary. |
| `details` | object | Optional structured context (JSON). Omitted when empty. |

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

## Placeholder tool

| Name | Result |
|------|--------|
| `server.ping` | `{"ok": true}` (inside `structuredContent`) |

## Tool: `top_terms`

Counts token frequencies in a UTF-8 text file using **pure Go** streaming (`internal/textscan`).

**Arguments:**

| Field | Type | Default | Notes |
|-------|------|---------|--------|
| `path` | string | required | Relative to `root`. |
| `root` | string | server `cwd` | Sandbox directory (`validate.ResolveUnderRoot`). |
| `top_n` | integer | `20` | Maximum `100`. |
| `min_len` | integer | `3` | Minimum token length in **Unicode runes**; maximum `30`. |
| `case_sensitive` | boolean | `false` | When `false`, each line is lowercased (`bytes.ToLower`) before tokenization. |
| `max_lines` | integer | `1_000_000` | Hard-capped at `1_000_000`; per-line cap remains **4 MiB** (`textscan.DefaultMaxLineBytes`). |

**Token rule:** maximal substrings of Unicode letters or numbers (`unicode.IsLetter` / `unicode.IsNumber`); all other runes delimit tokens.

**`structuredContent` shape:**

```json
{
  "path": "/resolved/path.txt",
  "top_n": 5,
  "terms": [{ "term": "example", "count": 12 }]
}
```

Sorting: **count descending**, then **term ascending** (lexicographic UTF-8 byte order).

**Example `tools/call`:**

```json
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"top_terms","arguments":{"path":"data.txt","top_n":10,"min_len":3}}}
```

## Tool: `unique_values`

Streams delimited lines from a UTF-8 file (`internal/textscan` + `internal/validate`). Field extraction uses **naive** `strings.Split` on `delimiter` — **not** RFC 4180 CSV (no quoted fields, no escapes). Rows with fewer than `field_index+1` fields are skipped.

**Arguments:**

| Field | Type | Default | Notes |
|-------|------|---------|--------|
| `path` | string | required | Relative to `root`. |
| `field_index` | integer | required | 0-based; maximum `1000`. |
| `root` | string | server `cwd` | Sandbox directory. |
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

**Arguments:** all of `unique_values` **except** `max_unique`, **plus**:

| Field | Type | Default | Notes |
|-------|------|---------|--------|
| `top_n` | integer | `20` | Maximum `100`; number of `{value,count}` rows returned. |
| `include_empty` | boolean | `false` | When `true`, count rows where the field is empty after trim. |

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

Tools that open paths on behalf of the user must treat paths as **untrusted** and validate them against a **server-configured root directory** using `internal/validate`:

- **Relative paths only** — absolute `userPath` values are rejected.
- **No `..` elements** in the user-provided path string.
- **Root containment** after `Join`/`Clean`, then again after **symlink resolution** (`filepath.EvalSymlinks`). Targets that resolve outside the real root are rejected.
- **Regular files only** (not directories or special files).
- **Default max file size: 25 MiB** per resolved file unless overridden via `validate.Limits.MaxFileBytes`.

Stable errors: `ErrInvalidInput`, `ErrAbsolutePath`, `ErrDotDot`, `ErrTraversal`, `ErrNotExist`, `ErrNotRegular`, `ErrFileTooLarge`, `ErrSymlink`. See [security.md](security.md) for the full policy table.

## Limits

- JSON-RPC line buffer: **4 MiB** per inbound message.
- `top_terms` line buffer: **4 MiB** per file line (`textscan`); line count capped by `max_lines` (default / hard max **1_000_000**).
- `unique_values` and `field_frequency`: same per-line and `max_lines` caps as above; `field_frequency` holds all distinct values in memory for counting (no `max_unique` cap).

Full product limits (text size, subprocess caps) will be documented alongside future tools.
