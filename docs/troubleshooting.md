# Troubleshooting

## `wc` not found (`count_lines_words_bytes`)

**Symptom:** Tool error mentioning `wc`, `executable file not found`, or `LookPath`.

**Fix:** Install **GNU coreutils** (or your OS equivalent) and ensure **`wc`** is on the server process **`PATH`**. In containers, add `coreutils` / `busybox` with `wc`. Other tools do not invoke `wc`.

**Verify:**

```bash
command -v wc && wc --version || wc -l /dev/null
```

## OS permissions and unreadable files

**Symptom:** `open ... permission denied`, or `validate: path does not exist` when the path looks correct.

**Fix:** The server runs as the **MCP host user**. The file must be **readable** by that user inside **`root`**. Symlinks are followed — target must also be readable and stay under the sandbox. Do not rely on tools to escalate privileges.

## Subprocess timeout or huge `wc` output

**Symptom:** `Timeout`, `TooLarge`, or generic `wc: ...` in **`reason`**.

**Fix:** Default **2 second** deadline (`exec.DefaultTimeout`) and **1 MiB** stdout/stderr cap per stream. Pathologically huge `wc` output should not occur for normal files; if customized runner limits are lower in your fork, raise them in `internal/exec` or reduce file use. For stuck NFS mounts, fix storage — the timeout is intentional.

## Path validation errors (`validate: …`)

**Symptoms:** `tool execution failed` with `reason` containing `validate:`.

| Message / error | Typical cause |
|-----------------|---------------|
| Absolute paths not allowed | `path` must be relative to `root`. |
| path contains .. | No `..` segments in the `path` string. |
| path escapes root boundary | Symlink or resolution left the sandbox. |
| path does not exist | Wrong path, wrong `root` / cwd, or typo. |
| not a regular file | Directory or special file. |
| file exceeds maximum size | File larger than server cap or your `max_bytes`. |
| symlink resolution failed | Broken symlink, loop, or permission on link chain. |

**Fix:** Use paths **relative** to the directory the host set as **cwd** (or pass **`root`** explicitly). Ensure the target is a **regular file** under that tree.

## `arguments:` / JSON decode errors

**Symptom:** `reason` starts with `arguments:`.

**Fix:** Valid JSON, no unknown keys (`DisallowUnknownFields`). Match **`inputSchema`** from `tools/list` / [tool-contract.md](tool-contract.md).

## No output on stdout / “hangs”

**Symptom:** Client sees nothing on stdout.

**Fix:** Responses are emitted only for requests with a **non-null `id`**. **Notifications** get **no** response line. Use **NDJSON** (one JSON object per line), **UTF-8**. Logs are **stderr**, not stdout.

## Line or scan limit errors (`textscan:`)

**Symptoms:** `textscan: line exceeds MaxLineBytes`, `line count exceeds MaxLines`.

**Fix:**

- **Per line:** max **4 MiB** (excluding newline) per `ScanLines` read. Split long lines upstream.
- **Per file:** max **1_000_000** lines per call (`max_lines` default/hard max).

## Encoding and invalid UTF-8

**Symptom:** Unexpected tokens or skipped characters in **`top_terms`**.

**Fix:** Input is read as **bytes**. Tokenization walks **UTF-8 runes**; **invalid UTF-8** bytes are consumed one at a time and act as **token boundaries** (see `internal/textscan` docs). Files should be **UTF-8** for predictable text tools. Delimited-field tools decode the **whole line as a Go string** for `Split` — invalid UTF-8 is preserved as bytes in subslices; prefer UTF-8 inputs.

## `unique_values` / `field_frequency` wrong or empty results

**Symptom:** Missing values, merged columns, or garbage fields.

**Fix:** Splitting is **`strings.Split`**, **not** CSV. Quoted delimiters, escapes, and multiline records are **unsupported**. **`field_index`** is **0-based**; short rows are **skipped**.

## MCP host cannot start the binary

**Symptom:** Spawn error from the host.

**Fix:** Absolute path to binary or ensure **`PATH`**. **`chmod +x`** after extracting release tarballs. Match **GOARCH** (amd64 vs arm64).

## Still stuck?

Open an issue with **tool name**, redacted **arguments**, and **`error.details.reason`**. See [security.md](security.md) and [examples.md](examples.md).
