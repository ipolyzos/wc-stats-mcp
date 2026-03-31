# Releases

## Overview

- **Integration** happens on **`develop`** (feature PRs merge here).
- **Releases** are cut from **`main`**: changes reach `main` only through a PR from `develop`, then a **SemVer git tag** triggers CI to build binaries and attach them to a **GitHub Release**.

See also [CHANGELOG.md](../CHANGELOG.md) for user-facing notes per version.

## First-time `main` branch

If the repository has **`develop`** but no **`main`** yet, create `main` once (for example after release prep is merged to `develop`):

```bash
git fetch origin develop
git push origin origin/develop:main
```

Then use PRs **`develop` → `main`** for subsequent releases. Set GitHub’s **default branch** to `main` if you want the UI and badge defaults to track the release line.

## 1. Merge `develop` → `main`

1. Ensure **`develop`** is green (CI, local `make test` / `make lint` as needed).
2. Update **[CHANGELOG.md](../CHANGELOG.md)** on `develop`: move items from `[Unreleased]` into a new section `## [X.Y.Z] - YYYY-MM-DD` for this release.
3. Open a **PR `develop` → `main`** (release PR). Get review and merge when ready.

Do not push directly to `main`; use the PR workflow described in [CONTRIBUTING.md](../CONTRIBUTING.md).

## 2. Tag `vX.Y.Z`

After the release PR is merged:

```bash
git checkout main
git pull origin main
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

Use **Semantic Versioning** (`MAJOR.MINOR.PATCH`), tag name **`v` + version** (examples: `v0.1.0`, `v1.2.3`).

Pushing the tag runs [`.github/workflows/release.yml`](../.github/workflows/release.yml), which:

- cross-compiles **`wc-stats-mcp`** for **Linux** and **macOS**, **amd64** and **arm64**;
- packages each binary as a **`.tar.gz`**;
- publishes **`checksums.txt`** (SHA-256) alongside the archives on the GitHub Release.

## 3. Verify the GitHub Release

Open the release on GitHub and confirm:

- attached **`.tar.gz`** files for all four platform/arch combinations;
- **`checksums.txt`** present;
- release notes (auto-generated and/or edited as needed).

## Local release-style build

To approximate a release binary locally (example: Linux amd64):

```bash
VERSION=v0.0.0-test
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -X wc-stats-mcp/internal/version.Version=${VERSION} -X wc-stats-mcp/internal/version.Commit=${COMMIT} -X wc-stats-mcp/internal/version.Date=${DATE}" \
  -o wc-stats-mcp ./cmd/wc-stats-mcp
```
