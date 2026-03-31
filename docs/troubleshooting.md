# Troubleshooting

## `wc` not found (`count_lines_words_bytes`)

**Symptom:** Tool error mentioning `wc` or executable not found.

**Fix:** Install coreutils and ensure `wc` is on the server process `PATH` (e.g. `/usr/bin/wc` on Linux). Other tools do not require `wc`.

## Path validation errors (`validate: …`)

**Symptoms:** `tool execution failed` with `reason` containing `validate:`.

| Message / error | Typical cause |
|-----------------|---------------|
| Absolute paths not allowed | `path` must be relative to `root`. |
| path contains .. | No `..` segments in the `path` string. |
| path escapes root boundary | Symlink or resolution left the sandbox. |
| path does not exist | Wrong path or wrong `root` / cwd. |
| not a regular file | Directory or special file. |
| file exceeds maximum size | File larger than server cap or your `max_bytes`. |

**Fix:** Use a path relative to the directory the host set as cwd (or pass `root` explicitly). Ensure the target is a normal file under that tree.

## `arguments:` / JSON decode errors

**Symptom:** `reason` starts with `arguments:`.

**Fix:** Send valid JSON for `tools/call` `arguments`; unknown fields are rejected (`DisallowUnknownFields`). Check required fields (`path`, `field_index`, etc.) in [tool-contract.md](tool-contract.md).

## No output on stdout / “hangs”

**Symptom:** Client sees nothing on stdout.

**Fix:** The server writes **one JSON-RPC response per input line** only for requests with a non-null `id`. **Notifications** (no `id`) get no response. Ensure stdin uses **one JSON object per line** (NDJSON) and UTF-8. Do not expect logs on stdout — they go to **stderr**.

## Line or scan limit errors (`textscan:`)

**Symptoms:** Errors mentioning line length or line count.

**Fix:** A single line may not exceed **4 MiB** (`DefaultMaxLineBytes`). At most **1_000_000** lines are read per tool call (`max_lines` default/hard max). Split or truncate huge lines/files, or reduce input size.

## `unique_values` / `field_frequency` wrong or empty results

**Symptom:** Missing values or wrong columns.

**Fix:** Splitting is **naive** (`strings.Split`), not RFC 4180 CSV — quoted commas break columns. Use simple delimited logs/TSV-style data, or preprocess with a real CSV parser outside the server. **`field_index`** is 0-based; short rows are skipped.

## MCP host cannot start the binary

**Symptom:** Host reports spawn error.

**Fix:** Use an absolute path to the binary or ensure the command is on `PATH`. On macOS/Linux, `chmod +x` release binaries after `tar xzf`. Confirm architecture matches (e.g. `arm64` vs `amd64`).

## Still stuck?

Open an issue with the **tool name**, redacted **arguments**, and the JSON-RPC **`error.details.reason`** string (no secrets). See [security.md](security.md) for what the server is allowed to do.
