# Contributing

## Workflow (every change)

Follow this sequence for **every** step of work:

1. **Branch**: create a branch using one of `feature/*`, `fix/*`, `chore/*`, or `docs/*` (never commit directly on `main` or `develop`).
2. **Docs first**: if behavior, interfaces, limits, or usage change, update `README.md` and/or `docs/**` **before** you commit that change.
3. **Commits**: keep commits **small and focused** — one logical change per commit.
4. **Test**: run `go test ./...` (and `go vet ./...`) locally **before pushing**.
5. **Push**: push the branch **after** each commit (or after each logical commit you intend to review separately).
6. **Pull request**: open a PR into **`develop`**. **Merges are PR-only**; do not push-merge to `develop` or `main`.
7. **Releases**: **`main` only**. Land work on `develop`, then merge `develop` → `main` via PR; tag and release from `main`.

## Branch naming

- `feature/*` — new capability
- `fix/*` — bug fixes
- `chore/*` — tooling, CI, repo hygiene
- `docs/*` — documentation-only

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `chore:`, etc.).

## Before merge

1. **CI must be green** (see `.github/workflows/ci.yml` when present).
2. On Linux, use `go test -race ./...` when touching concurrency.

## Engineering constraints

- **No shell pipelines**: spawn subprocesses with `exec.CommandContext` and explicit argument slices only.
- **Read-only server**: no file writes and no outbound network calls from the MCP process.
