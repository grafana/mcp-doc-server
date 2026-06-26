# Changelog

All notable changes to mcp-doc-server are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-25

Initial release of mcp-doc-server.

### Added

- **Core library** (`pkg/grafanadocs`): dependency-light Go package for Grafana documentation search and retrieval.
  - `LoadIndex` / `LoadIndexFromReader` — parse the Grafana docs index.
  - `Search` — TF-IDF keyword search with title weighting, phrase bonuses, and product filtering.
  - `FetchDoc` — fetch and clean a documentation page (URL allowlist enforced).
  - `Outline` — extract heading structure from a fetched document.
  - `Excerpt` — bounded retrieval by section name or line range.
  - Rate limiting (5 concurrent fetches, 200ms minimum gap).
  - URL allowlist restricted to `https://grafana.com/docs/`.

- **MCP server** (`cmd/mcp-doc-server`): stdio-based MCP server exposing four tools:
  - `search_docs` — search documentation by keyword.
  - `get_doc_outline` — get heading structure of a page.
  - `get_doc` — fetch cleaned Markdown by section or line range.
  - `list_products` — list available product documentation groups.

- **CLI** (`cmd/docs`): standalone command-line interface with the same four operations (`search`, `get`, `outline`, `products`). Supports text, JSON, YAML, and agents output formats.

- **MCP adapter** (`pkg/grafanadocs/mcp`): registers documentation tools on a `mark3labs/mcp-go` server. Used by mcp-grafana.

- **CLI adapter** (`pkg/grafanadocs/cli`): mountable Cobra command group. Used by gcx (`gcx docs`).

- **Documentation**: full docs suite covering get-started, install, tools reference, configuration, and library integration.

- **Demo**: demo script and agent workflow scenarios.

### Security

- SSRF protection: URL allowlist rejects non-`grafana.com` hosts and non-HTTPS schemes.
- Redirect blocking: redirects to non-`grafana.com` hosts are blocked.
- Body size limits: 2 MiB per page, 10 MiB for the index.
- Index URL restricted to HTTPS scheme only.

[0.1.0]: https://github.com/grafana/mcp-doc-server/releases/tag/v0.1.0
