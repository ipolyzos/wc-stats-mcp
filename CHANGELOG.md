# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-30

### Added

- **MCP stdio server** — newline-delimited JSON-RPC 2.0; responses on stdout, logs on stderr ([tool-contract.md](docs/tool-contract.md)).
- **Tools:** `server.ping`, `count_lines_words_bytes` (GNU `wc`), `top_terms` (Unicode token frequencies), `unique_values` and `field_frequency` (naive delimited fields; not CSV-safe).
- **Path sandbox** — relative paths, no `..`, symlink-safe root containment, regular files only, default **25 MiB** file cap; optional per-request **`max_bytes`** where implemented ([security.md](docs/security.md)).
- **Streaming limits** — capped lines and per-line bytes for text tools (`internal/textscan`).
- **CI** — Go test / gofmt / vet / govulncheck; markdown lint and link check ([.github/workflows](.github/workflows/)).
- **Release workflow** — tag `v*` builds Linux/macOS **amd64**/**arm64** archives and **SHA-256** checksums ([releases.md](docs/releases.md)).

[0.1.0]: https://github.com/ipolyzos/wc-stats-mcp/releases/tag/v0.1.0
