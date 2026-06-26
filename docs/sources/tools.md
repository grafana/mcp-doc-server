---
title: Tools and CLI reference
menuTitle: Tools
description: Reference for the four mcp-doc-server tools — search_docs, get_doc_outline, get_doc, and list_products — with parameters, output examples, and CLI usage.
weight: 4
topicType: reference
versionDate: 2026-06-25
---

# Tools and CLI reference

mcp-doc-server exposes four tools, available both as Model Context Protocol (MCP) tools and as `docs` CLI commands. They share the same data shapes, so JSON output is identical across both surfaces.

| Tool | CLI command | Purpose |
|------|-------------|---------|
| `search_docs` | `docs search` | Find relevant pages by keyword |
| `get_doc_outline` | `docs outline` | Get a page's heading structure |
| `get_doc` | `docs get` | Fetch cleaned Markdown, by section or line range |
| `list_products` | `docs products` | List product documentation groups |

## Workflow

Agents narrow in progressively—search, outline, then fetch only the section they need.

![Sequence diagram showing the full tool workflow: the agent searches for pages, requests a page outline, and fetches one section while the server retrieves documentation on demand](tools-workflow.svg)

### Worked example

**1. Search** for pages:

```json
{"query": "metrics generator", "product": "tempo", "limit": 3}
```

```json
[
  {
    "title": "Span metrics",
    "url": "https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/",
    "description": "Generate metrics from spans.",
    "product": "Grafana Tempo"
  },
  {
    "title": "Metrics-generator",
    "url": "https://grafana.com/docs/tempo/latest/metrics-generator/",
    "description": "Configure the metrics generator.",
    "product": "Grafana Tempo"
  }
]
```

**2. Outline** the top result:

```json
{"url": "https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/"}
```

```json
{
  "url": "https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/",
  "headings": [
    {"level": 1, "text": "Span metrics", "line": 1},
    {"level": 2, "text": "Configuration", "line": 15},
    {"level": 2, "text": "Example", "line": 56}
  ]
}
```

**3. Fetch** just the section you need:

```json
{
  "url": "https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/",
  "section": "Configuration"
}
```

```json
{
  "content": "## Configuration\n\nConfigure the span metrics processor...",
  "url": "https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/",
  "total_lines": 72,
  "returned_range": [15, 55]
}
```

---

## `search_docs`

Search Grafana documentation by keyword. Returns matching pages ranked by relevance.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | — | Search query. Matches whole words in titles and descriptions. |
| `product` | string | No | — | Filter to a specific product. Resolves by exact match, then prefix, then partial match. |
| `limit` | integer | No | 5 | Maximum results to return. |

### Output

A JSON array of entries, each with `title`, `url`, `description`, and `product` (all strings).

```json
[
  {
    "title": "Configure alerting",
    "url": "https://grafana.com/docs/grafana/latest/alerting/configure-alerting/",
    "description": "Configure your alerting environment and notification policies.",
    "product": "Grafana"
  }
]
```

### Ranking

Results use a deterministic scoring algorithm:

- **Word-boundary matching**—tokens match whole words. "rate" matches "rate" but not "migrate".
- **TF-IDF weighting**—rare terms score higher than common ones.
- **Title 3x weight**—title matches count three times as much as description matches.
- **All-tokens bonus (1.5x)**—when every query token matches and the query has more than one token.
- **Exact phrase bonus (2x)**—when a multi-token query appears verbatim in the title.

### Product filter

The `product` value resolves to canonical names in precedence order, stopping at the first level that matches anything:

1. **Exact** (case-insensitive)—`Grafana Loki`
2. **Prefix**—`Grafana` matches "Grafana Loki", "Grafana Mimir", etc.
3. **Partial**—`loki` matches "Grafana Loki"

Use `list_products` to discover valid names.

### Empty results

When nothing matches, the tool returns a plain-text message (not a JSON array) with guidance—suggesting broader terms, or, if a product filter was set, suggesting you broaden it or call `list_products`.

### Examples

```json
{"query": "rate limiting"}
{"query": "retention", "product": "Loki"}
{"query": "alerting rules", "limit": 10}
```

---

## `get_doc_outline`

Get the heading outline of a page. Use it to find section names before calling `get_doc` with a `section`.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `url` | string | Yes | — | The `grafana.com/docs/` URL to inspect. |

### Output

A JSON object with the page `url` and an array of `headings`:

```json
{
  "url": "https://grafana.com/docs/loki/latest/configure/",
  "headings": [
    {"level": 1, "text": "Configure Loki", "line": 1},
    {"level": 2, "text": "Limits", "line": 42},
    {"level": 3, "text": "Per-tenant limits", "line": 58}
  ]
}
```

Each heading has `level` (1 = `#`, 2 = `##`, …), `text` (formatting stripped), and `line` (1-indexed).

### Parsing rules

The outline parses only headings that start with `#`. Additionally:

- Headings inside fenced code blocks are excluded.
- Lines indented four or more spaces are treated as code, not headings.
- Trailing `#` sequences are stripped (`## Storage ##` becomes `Storage`).
- Headings that are empty after stripping are excluded.

The same `https://grafana.com/docs/` allowlist as `get_doc` applies.

---

## `get_doc`

Fetch a page as cleaned Markdown. Supports section extraction and offset/limit paging for bounded retrieval.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `url` | string | Yes | — | The `grafana.com/docs/` URL to fetch. |
| `section` | string | No | — | Heading text to extract. Returns only that section. |
| `offset` | integer | No | 0 | Line offset for paging (0-indexed). Used when `section` isn't set. |
| `limit` | integer | No | 80 | Maximum lines to return. Used when `section` isn't set. |

### Output

```json
{
  "content": "# Configure alerting\n\nConfigure your alerting...",
  "url": "https://grafana.com/docs/grafana/latest/alerting/configure-alerting/",
  "total_lines": 245,
  "returned_range": [1, 80]
}
```

`content` is the cleaned Markdown, `url` is the canonical source (for citation), `total_lines` is the full document length, and `returned_range` is the 1-indexed `[start, end]` of what came back.

### Section extraction

When `section` is set, `get_doc` returns the content from that heading to the next heading of equal or higher level. Matching is case-insensitive against the heading text. Use `get_doc_outline` first to find exact names.

### Paging

When `section` isn't set, use `offset` and `limit`. The default `limit` is 80 lines—`get_doc` never returns the whole page unless you raise the limit. Use `total_lines` and `returned_range` to page through a long document.

### Content cleanup

The returned Markdown is stripped of front matter, Hugo shortcodes, and HTML comments; blank lines are collapsed. Code blocks are preserved intact.

### Errors

If `section` is set but no matching heading exists, the tool returns an error message rather than content:

```
section "Retention" not found. Use get_doc_outline to see available headings.
```

### Examples

```json
{"url": "https://grafana.com/docs/loki/latest/configure/", "limit": 50}
{"url": "https://grafana.com/docs/grafana/latest/alerting/configure-alerting/", "section": "Notification policies"}
{"url": "https://grafana.com/docs/mimir/latest/configure/configuration-parameters/", "offset": 100, "limit": 50}
```

---

## `list_products`

List all product documentation groups with their page counts. Use it to discover products for filtering `search_docs`. Takes no parameters.

### Output

```json
{
  "products": [
    {"name": "Grafana", "count": 892},
    {"name": "Grafana Loki", "count": 245},
    {"name": "Grafana Tempo", "count": 178}
  ]
}
```

Each product has a `name` and a `count` of indexed pages.

### What counts as a product

Products come from the `## ... documentation` headers in the index. "Documentation home" and "Copyright notice" are excluded, and a trailing "documentation" is stripped from each name.

---

## CLI usage

The `docs` CLI mirrors the tools in a terminal-friendly form—useful for testing, scripting, and direct access without an MCP client.

### Output formats

Every command accepts `-o` / `--output`:

| Format | Description |
|--------|-------------|
| `text` | Human-readable aligned tables or raw Markdown (default) |
| `json` | Indented JSON |
| `yaml` | YAML |
| `agents` | Compact single-line JSON (for machine consumption) |

JSON and YAML keys match the MCP tool output exactly.

### `docs search`

```
docs search <query> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--product` | — | Filter results to a specific product |
| `--limit` | 5 | Maximum results to return |
| `-o`, `--output` | `text` | Output format |

```bash
docs search "rate limiting"
docs search "alerting rules" --limit 10
docs search "retention" --product "Grafana Loki"
docs search clustering -o json
```

`search` loads the index on first use.

### `docs get`

```
docs get <url> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--section` | — | Heading text to extract |
| `--offset` | 0 | Line offset for paging (0-indexed) |
| `--limit` | 0 (= 80 lines) | Maximum lines to return |
| `-o`, `--output` | `text` | Output format |

```bash
docs get https://grafana.com/docs/grafana/latest/alerting/configure-alerting/
docs get https://grafana.com/docs/loki/latest/configure/ --section "Limits"
docs get https://grafana.com/docs/mimir/latest/configure/configuration-parameters/ --offset 100 --limit 50
```

`get` fetches directly from `grafana.com` and doesn't need the index.

### `docs outline`

```
docs outline <url> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-o`, `--output` | `text` | Output format |

```bash
docs outline https://grafana.com/docs/grafana/latest/alerting/configure-alerting/
docs outline https://grafana.com/docs/loki/latest/configure/ -o json
```

Like `get`, `outline` doesn't need the index.

### `docs products`

```
docs products [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-o`, `--output` | `text` | Output format |

```bash
docs products
docs products -o json
```

`products` loads the index on first use.

### Index loading

The CLI loads the index lazily—only `search` and `products` need it. `get` and `outline` fetch pages directly and run with no startup delay. Set `DOCS_INDEX_URL` to override the index location.

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (bad arguments, index load failure, fetch failure, or section not found) |

### Piping

The `agents` format pairs well with `jq` for scripting:

```bash
docs search "alerting" -o agents | jq -r '.[].url'
```

## Related resources

- [Install and configure](../install/)—build the binaries and connect a client
- [Configuration](../configure/)—environment variables and limits
- [Integrate the core library](../integrate/)—the Go API behind these tools
