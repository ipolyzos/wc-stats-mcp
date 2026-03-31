# Security notes

## File path validation (`internal/validate`)

Read-only tools that accept a filesystem path must use **`validate.ResolveUnderRoot`** (or equivalent logic) so untrusted strings cannot escape a configured sandbox directory.

### Rules

| Rule | Behavior |
|------|-----------|
| **Root** | The server resolves the root with `filepath.Abs` + `filepath.EvalSymlinks` once per call chain. The root directory must exist and be resolvable. |
| **User path** | Must be **relative** (`ErrAbsolutePath` if absolute). Leading/trailing ASCII space is trimmed. Empty root or path → `ErrInvalidInput`. |
| **`..` segments** | Any path element equal to `..` is rejected (`ErrDotDot`) before joining with root. This blocks obvious traversal strings. |
| **Join + boundary** | After `filepath.Join(root, userPath)` and `Clean`, the candidate must still lie under the real root (`filepath.Rel` check). Otherwise `ErrTraversal`. |
| **Symlinks** | **Follow and re-check.** `filepath.EvalSymlinks` is applied to the candidate path. The fully resolved path must again lie under the real root; otherwise `ErrTraversal`. Symlink cycles or resolution failures → `ErrSymlink` (wrapped with the OS error). |
| **Type** | Final target must be a **regular file** (`os.Stat` after resolution). Directories, devices, sockets, etc. → `ErrNotRegular`. |
| **Size** | File size must not exceed **`Limits.MaxFileBytes`**. When zero, **25 MiB** (`validate.DefaultMaxFileBytes`) is used. Oversize → `ErrFileTooLarge`. |

### Rationale for following symlinks

Following symlinks and then re-applying the root containment check prevents **“symlink out”** escapes: a link inside the root that points outside is rejected after resolution, while links that stay inside the root (including to another file under root) remain valid.

## Subprocess execution

When the server runs external binaries, policy is:

- **No shell**: use `exec.CommandContext` with explicit argv only (no `sh -c`).
- **Timeouts and stream caps**: bound wall time and captured stdout/stderr size (see `internal/exec` when present in the tree).

These layers complement path validation: path checks gate *what file* is read; the runner gates *how* helper binaries are invoked.
