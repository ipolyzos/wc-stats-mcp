# wc-stats-mcp (WIP)

**Work in progress:** a **read-only** [Model Context Protocol](https://modelcontextprotocol.io/) server over **stdio** for text statistics (wrapping allowlisted coreutils such as `wc`, `sort`, `uniq`, and optionally `cut`), with strict limits and deterministic locale (`LC_ALL=C` / `LANG=C`).

## Planned MCP tools

| Tool | Purpose (planned) |
|------|---------------------|
| `count_lines_words_bytes` | Line / word / byte / character counts via `wc`. |
| `top_terms` | Token frequency: tokenize in Go, then `sort` → `uniq -c` → `sort -nr`. |
| `unique_values` | Delimited field extraction and unique values (deterministic sort optional). |
| `field_frequency` | Per-field value counts via sort / uniq pipeline. |

Implementation, schemas, and client contract will be documented under `docs/` as they land.

## Local development

See [docs/development.md](docs/development.md) for build, test, and run commands.

## Contributing

Branches **`main`** and **`develop`**, PR-only workflow, branch prefixes, and **test-before-PR** are described in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache-2.0 — see [LICENSE](LICENSE).
