# Architecture

## Transport and control plane

The server speaks **newline-delimited JSON-RPC 2.0** on stdio (see [tool-contract.md](tool-contract.md)). The MCP layer (`internal/mcp`) parses requests, dispatches **tools** (`internal/tools`), and writes responses exclusively to **stdout**; logs go to **stderr**.

## Why advanced text tools stay pure Go

Several planned tools (token frequency, field extraction, etc.) need to scan potentially large user text or files. Those code paths are implemented as **pure Go** streaming logic in **`internal/textscan`** (and friends), not as extra shell pipelines, for:

1. **Smaller attack surface** — Fewer subprocess invocations mean fewer opportunities for argument injection, `PATH` confusion, or unexpected binary behavior. Where we do call coreutils (`wc`, `sort`, …), we keep a **small, reviewed set** and use **`exec.CommandContext`** with explicit argv (`internal/exec`).
2. **Deterministic resource limits** — Go code can enforce **line counts**, **per-line byte caps** (via `bufio.Scanner` buffer limits), and **file size** checks (`internal/validate`) with stable, testable errors (`textscan.ErrLineTooLong`, `textscan.ErrTooManyLines`).
3. **Predictable semantics** — Tokenization and normalization rules are defined in code and documented next to the implementation (e.g. `ForEachNonLetterDigitToken`), instead of depending on locale-specific shell tools for early parsing stages.

Subprocesses are reserved for phases where a vetted coreutils behavior is desired (e.g. `sort` collation under `LC_ALL=C`) after input has already been bounded and validated.

## Package map (concise)

| Package | Role |
|---------|------|
| `internal/mcp` | stdio JSON-RPC loop, tool dispatch |
| `internal/tools` | Tool handlers and registry |
| `internal/validate` | Path sandboxing, symlink-safe root checks, file size cap before open |
| `internal/textscan` | Safe open + capped line streaming, tokenization, normalization |
| `internal/exec` | Hardened subprocess runner (timeouts, stdout/stderr caps) |
| `internal/version` | Build metadata |

## Validation and scanning

- **Open paths** for file-based tools: `validate.ResolveUnderRoot` then `os.Open` (`textscan.OpenFile`).
- **Stream lines**: `textscan.ScanLines` / `textscan.ScanLinesFile` with `ScanLimits` (`DefaultMaxLines`, `DefaultMaxLineBytes`).
- **Tokens**: `textscan.ForEachNonLetterDigitToken` (see doc comment for the exact Unicode rule).
