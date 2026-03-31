# Security model

## Threat model (scope)

### In scope

- **Caller-controlled input:** JSON-RPC payloads (`path`, `root`, numeric limits, tool names) and file contents read through validated paths.
- **Process boundary:** MCP host spawns `wc-stats-mcp`; compromise of the **host** or **OS user** is outside this document.

### Out of scope

- Secrecy of file contents once legitimately opened (server reads files as the host user).
- Network transport (stdio only; no TLS here).

### Design goals

1. **No shell injection** — never pass user strings to `/bin/sh -c`.
2. **Filesystem sandbox** — only read files proven to live under **`root`** after symlink resolution.
3. **Resource bounds** — file size, line count, line length, subprocess wall-clock time, subprocess output size.
4. **Read-only tools** — no intentional writes or outbound network from tool execution.

## File sandboxing

All file-backed tools use **`internal/validate.ResolveUnderRoot`**:

| Control | Purpose |
|---------|---------|
| Relative `path` only | Blocks `userPath` that is absolute. |
| No `..` segments | Blocks dot-dot tricks in the **string** before join. |
| Clean + join under `root` | Normalized candidate path. |
| **`EvalSymlinks`** on root and target | Final path must still lie under the real root (`ErrTraversal`). |
| **`Stat` regular file** | No directories or device nodes as read targets. |
| **Size limit** | Default **25 MiB** (`DefaultMaxFileBytes`); **`max_bytes`** in tool args can only **lower** the effective cap for that call (`Limits.WithRequestMaxBytes`). |

The **`root`** default is the server process **cwd** — the MCP host chooses cwd when spawning the binary; treat that as the trust boundary for “which tree is readable.”

## Limits (summary)

| Limit | Value | Where enforced |
|-------|--------|----------------|
| RPC line | **4 MiB** | `internal/mcp` scanner buffer |
| File size | **25 MiB** default (tightenable) | `validate.ResolveUnderRoot` |
| Lines per streaming read | **1_000_000** max | `ScanLines` / `max_lines` |
| Bytes per line | **4 MiB** | `textscan.DefaultMaxLineBytes` |
| `wc` wall time | **2 s** default | `exec.DefaultTimeout` |
| `wc` stdout/stderr capture | **1 MiB** each default | `exec.DefaultMaxStreamBytes` |
| `top_n` | max **100** | tool handlers |
| `field_index` | max **1000** | delimited-field tools |
| Delimiter length | max **3** bytes | `unique_values` / `field_frequency` |

`field_frequency` accumulates **all distinct values in memory** — very wide keys can increase RSS; there is no `max_unique` cap for that tool.

## Subprocess policy

- **Allowed:** **`wc`** with fixed flags **`-l -w -c`** and a **single absolute path** produced by validation.
- **Mechanism:** **`exec.CommandContext`** + explicit argv slice; **stdin** not wired to user data.
- **Timeout:** bounded context (**default 2 s**).
- **Output:** stdout/stderr copied to capped buffers; overflow → categorized error, not unbounded growth.

Future subprocess tools (if any) should follow the same pattern and be listed here.

## Path to disclosure

Report security vulnerabilities via **[SECURITY.md](../SECURITY.md)** (GitHub Security Advisories / private contact). Do not file public issues for undisclosed exploits.

## Operational notes

- **`PATH`:** `wc` is resolved with **`exec.LookPath`**. Locked-down deploys may set `PATH` explicitly or extend the runner with a fixed `wc` path.
- **Permissions:** If the worker user cannot read a path, `open` fails — surfaced as a tool error, not a server crash.
- **Delimited-field tools** are **not** CSV-safe; do not assume quoting or escapes for untrusted CSV.

Related: [tool-contract.md](tool-contract.md), [architecture.md](architecture.md), [troubleshooting.md](troubleshooting.md).
