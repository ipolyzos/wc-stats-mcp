# Contributing

## Branches

- **`main`**: release line; protected; merges via PR only.
- **`develop`**: integration branch; default target for feature PRs; merges via PR only.

Do not push directly to `main` or `develop`.

## Pull requests

- **PR-only merges** into `develop` (and `develop` → `main` for releases).
- Open a PR **before** merging any work.
- Run **`go test ./...`** (and `go vet ./...`) **locally before opening or updating a PR**. Fix failures before requesting review.

## Branch naming

Use a prefix that matches the change type:

| Prefix | Use |
|--------|-----|
| `feature/*` | New capability |
| `fix/*` | Bug fixes |
| `chore/*` | Tooling, CI, repo hygiene |
| `docs/*` | Documentation only |

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `docs:`, `chore:`, etc.

Prefer **small, focused commits** (one logical change per commit).

## Docs and behavior changes

When behavior, interfaces, limits, or usage change, update `README.md` and/or `docs/**` **in the same PR** (ideally in commits before or with the code change).

## Before merge

1. **CI must be green** when `.github/workflows` exists.
2. On Linux, use `go test -race ./...` when concurrency is involved.

## Engineering constraints

- **No shell pipelines**: use `exec.CommandContext` with explicit argument slices only (no `sh -c`).
- **Read-only server**: no filesystem writes and no outbound network I/O from the MCP process for tool execution.
