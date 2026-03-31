# Security model

This document summarizes how `wc-stats-mcp` limits risk: **no shell**, **bounded subprocess I/O**, **sandboxed file access**, and **read-only** tool behavior.

## Subprocess execution (`internal/exec`)

Used today for **`wc`** (`count_lines_words_bytes`).

### No shell

- Processes use **`exec.CommandContext`** with a **resolved executable** and an explicit **`[]string` argv**.
- **Never** `sh -c`, `/bin/sh`, or string-built command lines for user-influenced data.

### Timeouts

- Runs use a **bounded context** and **default deadline** (`internal/exec.DefaultTimeout`). Overruns surface as **`Timeout`**.

### Output limits

- **Stdout** and **stderr** use **capped buffers** (defaults: **1 MiB** each, `DefaultMaxStreamBytes`). Overflow → **`TooLarge`** with truncated capture.

### PATH

- **`wc`** is resolved with `exec.LookPath`. Hardening (fixed paths, sanitized `PATH`) can be added in the runner or deployment if needed.

## File paths (`internal/validate`)

All tools that read files resolve paths with **`ResolveUnderRoot(root, userPath, limits)`**:

| Rule | Detail |
|------|--------|
| Relative only | Absolute `userPath` rejected (`ErrAbsolutePath`). |
| No `..` | `..` path elements rejected (`ErrDotDot`). |
| Root containment | After `Join`/`Clean` and after **`filepath.EvalSymlinks`** (`ErrTraversal` if escape). |
| Regular files | Directories and specials rejected (`ErrNotRegular`). |
| Size cap | Default **25 MiB** per file (`DefaultMaxFileBytes`). Optional tool arg **`max_bytes`** tightens the cap per request; it **cannot** exceed the server’s effective maximum (`Limits.WithRequestMaxBytes`). |

Stable errors (for logs and `reason` strings): `ErrInvalidInput`, `ErrAbsolutePath`, `ErrDotDot`, `ErrTraversal`, `ErrNotExist`, `ErrNotRegular`, `ErrFileTooLarge`, `ErrSymlink`.

## Read-only posture

- Tool execution must not **write arbitrary files** or open **outbound network** I/O. Subprocess use is limited to **`wc`** with a validated absolute file path.

## Delimited-field tools

`unique_values` and `field_frequency` use **naive splitting** — not a standards-compliant CSV parser. Do not treat them as safe for untrusted quoted CSV; use for simple delimited logs under the same path rules above.

## Reporting issues

See [SECURITY.md](../SECURITY.md) for vulnerability reporting.
