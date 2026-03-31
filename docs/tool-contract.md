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

## File path validation (read-only tools)

Tools that open paths on behalf of the user must treat paths as **untrusted** and validate them against a **server-configured root directory** using `internal/validate`:

- **Relative paths only** — absolute `userPath` values are rejected.
- **No `..` elements** in the user-provided path string.
- **Root containment** after `Join`/`Clean`, then again after **symlink resolution** (`filepath.EvalSymlinks`). Targets that resolve outside the real root are rejected.
- **Regular files only** (not directories or special files).
- **Default max file size: 25 MiB** per resolved file unless overridden via `validate.Limits.MaxFileBytes`.

Stable errors: `ErrInvalidInput`, `ErrAbsolutePath`, `ErrDotDot`, `ErrTraversal`, `ErrNotExist`, `ErrNotRegular`, `ErrFileTooLarge`, `ErrSymlink`. See [security.md](security.md) for the full policy table.

## Limits

- Scanner buffer per line: **4 MiB** (oversized lines error at read time).

Full product limits (text size, subprocess caps) will be documented alongside future tools.
