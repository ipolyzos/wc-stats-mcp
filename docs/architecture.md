# Architecture

## Transport and control plane

The server speaks **newline-delimited JSON-RPC 2.0** on stdio (see [tool-contract.md](tool-contract.md)). The MCP layer (`internal/mcp`) parses requests, dispatches **tools** (`internal/tools`), and writes responses exclusively to **stdout**; logs go to **stderr**.

## Subprocess vs pure Go

| Capability | Implementation |
|------------|----------------|
| Line / word / byte counts | **`wc -l -w -c`** via `internal/exec` (validated path only). |
| Token frequencies (`top_terms`) | **Pure Go** streaming in `internal/textscan`. |
| Delimited field tools (`unique_values`, `field_frequency`) | **Pure Go** line scan + `strings.Split` (documented as not CSV-safe). |
| Health | **`server.ping`** — no filesystem access. |

Heavy scanning stays in **Go** so limits (`max_lines`, per-line bytes, `max_bytes` / default file cap) are enforced in one place with stable errors (`textscan.ErrLineTooLong`, `textscan.ErrTooManyLines`, `validate.ErrFileTooLarge`, etc.). The subprocess runner uses **`exec.CommandContext`**, explicit argv (no shell), timeouts, and capped stdout/stderr — see [security.md](security.md).

## Package map

| Package | Role |
|---------|------|
| `cmd/wc-stats-mcp` | `main`: registry wiring via `tools.RegisterProduction`, stdio server |
| `internal/mcp` | JSON-RPC loop, `tools/list`, `tools/call`, error envelope |
| `internal/tools` | Handlers, registry, MCP `inputSchema` metadata (`catalog.go`) |
| `internal/validate` | Path sandboxing, symlink-safe root checks, optional per-request `max_bytes` |
| `internal/textscan` | Capped line streaming, tokenization, normalization |
| `internal/exec` | Hardened `wc` (and future) subprocess runner |
| `internal/version` | Build metadata (`-ldflags` at release; default from source matches tag) |

## Validation and scanning

- **Open paths:** `validate.ResolveUnderRoot` → `os.Open`.
- **Stream lines:** `textscan.ScanLines` with `ScanLimits` (`MaxLines`, `MaxLineBytes`; defaults in [tool-contract.md](tool-contract.md)).
- **Tokens:** `textscan.ForEachNonLetterDigitToken` / `ForEachNonLetterDigitTokenNormalized` — see package doc for Unicode rules.
