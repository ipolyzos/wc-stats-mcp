# Architecture

## MCP stdio flow

```text
  Client (MCP host)                    wc-stats-mcp process
  ─────────────────                    ─────────────────────
        │                                       │
        │  NDJSON line(s) on stdin             │
        ├──────────────────────────────────────►│  bufio.Scanner (≤ 4 MiB/line)
        │                                       │  json.Unmarshal → Request
        │                                       │
        │                                       ├─ initialize → cap handshake JSON
        │                                       ├─ tools/list → ToolMeta(registry names)
        │                                       └─ tools/call → Registry.Call(name, args)
        │                                              │
        │                                              ├─ validate.ResolveUnderRoot (file tools)
        │                                              ├─ textscan.ScanLines / tokenize (pure Go)
        │                                              └─ exec.Runner (wc only)
        │                                       │
        │  One JSON-RPC response line / request   │
        │◄──────────────────────────────────────┤  json.Encoder → stdout only
        │                                       │
        │  (optional) logs                     │  log.Logger → stderr
        │◄ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┤
```

**Rules:** Responses and results must never be written to **stderr**; clients parse **stdout** only as JSON-RPC. Blank input lines are ignored.

## Registry and handlers

- **`internal/tools.Registry`**: thread-safe map `name → Handler`. **`RegisterProduction`** installs the five production tools in a fixed order for parity with `cmd/wc-stats-mcp`.
- **`Handler`**: `(ctx, json.RawMessage arguments) → (json.RawMessage result, error)`. Each handler uses **`json.Decoder` with `DisallowUnknownFields()`** on `arguments`.
- **`ToolMeta(name)`** (`catalog.go`): returns MCP **`description`** + **`inputSchema`** JSON-object shape (`type`, `required`, `properties`) for `tools/list`. Schemas are defined in code, not separate `.json` files.

## Validation layer

- **`validate.ResolveUnderRoot(root, userPath, limits)`** runs before any file open: relative path, no `..`, join+clean under real root, **EvalSymlinks** on root and target, regular file, size check (default **25 MiB**, optional **`max_bytes`** tightening).
- File tools return **`validate` errors as-is** (`ResolveUnderRoot` already uses the `validate: …` prefix); avoid double-wrapping.

## Exec runner

- **`internal/exec.Runner`**: **`exec.CommandContext`** only — **no shell**. **`wc`** argv is `[]string{"-l", "-w", "-c", absPath}` after validation.
- **Defaults:** **2s** timeout per run (**`DefaultTimeout`**), **1 MiB** cap on captured stdout and stderr each (**`DefaultMaxStreamBytes`**). Truncation yields **`TooLarge`**; deadline exceeded yields **`Timeout`**.

## Error model (JSON-RPC)

| Layer | Responsibility |
|-------|----------------|
| **`internal/mcp`** | Parse line → `Request`; map failures to **JSON-RPC `error`** (`CodeParseError`, `CodeInvalidRequest`, `CodeMethodNotFound`, `CodeInvalidParams`, `CodeInternal`). |
| **`tools/call`** | Unknown tool or handler `error` → **`CodeToolError` (-32000)**, **`message`**: `tool execution failed`, **`details`**: `{ "tool", "reason" }` via **`ToolExecutionFailed`**. |
| **Handlers** | Return Go `error`; no JSON-RPC codes inside tools. |

See [tool-contract.md](tool-contract.md) for the full code table and examples.

## Testing strategy

| Area | Typical tests |
|------|----------------|
| **MCP** | `internal/mcp` — stdio harness, `tools/call` ping, unknown tool envelope. |
| **Tools** | Golden / integration tests per tool (`testdata/`), registry routing. |
| **Validate** | Symlink traversal, size cap, `WithRequestMaxBytes`. |
| **Textscan** | Line-too-long, too-many-lines, tokenizer normalization. |
| **Exec** | Runner timeout and stream caps (where not requiring real `wc`). |
| **Consistency** | `ProductionToolNames` ↔ `RegisterProduction`; `ToolMeta` JSON round-trip. |

Run **`make test`** and **`go test -race ./...`** when touching concurrency.

## Subprocess vs pure Go

| Capability | Implementation |
|------------|----------------|
| Line / word / byte counts | **`wc -l -w -c`** via `internal/exec`. |
| `top_terms` | **Pure Go** streaming in `internal/textscan`. |
| `unique_values`, `field_frequency` | **Pure Go** `ScanLines` + **`strings.Split`** (not CSV-safe). |
| `server.ping` | No filesystem. |

## Package map

| Package | Role |
|---------|------|
| `cmd/wc-stats-mcp` | `main`: `RegisterProduction`, stdio server |
| `internal/mcp` | JSON-RPC loop, `tools/list`, `tools/call`, error envelope |
| `internal/tools` | Handlers, registry, `ToolMeta` / `catalog.go` |
| `internal/validate` | Path sandbox, `ResolveUnderRoot`, `Limits` |
| `internal/textscan` | Capped line streaming, tokenization |
| `internal/exec` | Hardened `wc` runner |
| `internal/version` | Build metadata (`-ldflags` at release) |

## Validation and scanning (detail)

- **Open paths:** `ResolveUnderRoot` → **`os.Open`** (absolute path passed to `wc` is the resolved path).
- **Stream lines:** `textscan.ScanLines` + **`ScanLimits`** (`MaxLines`, `MaxLineBytes`).
- **Tokens:** `ForEachNonLetterDigitToken` / `ForEachNonLetterDigitTokenNormalized` — see package docs for Unicode rules.

Related: [security.md](security.md), [tool-contract.md](tool-contract.md), [examples.md](examples.md).
