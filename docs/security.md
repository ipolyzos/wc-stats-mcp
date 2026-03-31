# Security model (subprocesses)

This document captures policy for running external programs from `wc-stats-mcp`. The implementation lives in `internal/exec`.

## No shell

- Processes are started with **`exec.CommandContext`** (or equivalent) using a **resolved executable path** and an **explicit `[]string` argv**.
- **Never** invoke `/bin/sh`, `cmd.exe`, or `sh -c` to run user-influenced strings.
- Arguments are fixed by tool code or validated before being passed through; there is no string concatenation into a shell command line.

## Timeouts

- Every run uses a **bounded context** with a configurable **default deadline** (see `internal/exec.DefaultTimeout`).
- Exceeding the deadline is classified as **`Timeout`** (`internal/exec.Kind`) so callers can distinguish hangs from binary failures.

## Output limits

- **Stdout** and **stderr** are captured through **capped writers** with configurable maxima (defaults: **1 MiB each**, `internal/exec.DefaultMaxStreamBytes`).
- When a stream exceeds its cap, the runner returns **`TooLarge`** and retains only the prefix up to the cap (truncation). This mitigates memory exhaustion from chatty subprocesses.

## Error taxonomy

Failures are mapped to stable categories (`InvalidArgument`, `NotFound`, `TooLarge`, `Timeout`, `ExecFailed`) via `internal/exec.Error` so upper layers can handle them without parsing strings.

## PATH and environment

- Resolution uses `LookPath` (typically `exec.LookPath`) against the process environment.
- Tooling that needs stricter guarantees (fixed allowlists, sanitized `PATH`, `LC_ALL=C`) should be layered **above** this runner in dedicated tool packages; this document will be extended as those controls land.

## Read-only posture

- The server must not write arbitrary files or open outbound network connections for tool execution; subprocess I/O is stdin/stdout/stderr only from the runner’s perspective.
