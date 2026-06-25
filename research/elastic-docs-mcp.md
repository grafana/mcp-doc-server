# Prior art: Elastic's docs MCP server

Notes on Elastic's documentation MCP server, shipped as part of
[`docs-builder`](https://github.com/elastic/docs-builder/blob/main/docs/mcp/index.md),
and how it compares to `mcp-doc-server`.

## What it is

- Ships as part of `docs-builder`.
- Deployed as a **stateless HTTP service**, exposing all tools through a single
  Streamable HTTP endpoint at `https://www.elastic.co/docs/_mcp/`.
- **Hosted by Elastic** — users do not run a local binary; they point their MCP client at
  the URL.

### Distribution model

Install is just a URL in the MCP client config (Cursor, Claude Code, VS Code, IntelliJ).
Example `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "elastic-docs": { "url": "https://www.elastic.co/docs/_mcp/" }
  }
}
```

### Tools (6, in 3 groups)

- **Search**
  - `search_docs` — semantic ("by meaning") search; returns docs with AI summaries,
    relevance scores, and navigation context; filterable by product and section.
  - `find_related_docs` — discover related pages around a topic.
- **Document**
  - `get_document_by_url` — fetch a page by URL or path; returns title, AI summaries,
    headings, navigation context, and optionally the full body.
  - `analyze_document_structure` — heading count, link count, parent pages, AI enrichment
    status.
- **Coherence** (the distinctive ones)
  - `check_docs_coherence` — how coherently a topic is covered across all docs.
  - `find_docs_inconsistencies` — finds overlapping/contradictory content within a product
    area.

### Testing

Elastic suggests the MCP Inspector:

```bash
npx @modelcontextprotocol/inspector --url https://www.elastic.co/docs/_mcp/
```

## Comparison to mcp-doc-server

| Dimension | Elastic docs MCP | mcp-doc-server (per SPECS.md) |
|---|---|---|
| Transport | Hosted Streamable HTTP | Local stdio binary |
| Search | Semantic (embeddings/AI) | Deterministic, no LLM/embeddings (I2) |
| Doc enrichment | AI summaries, relevance scores | Raw markdown via the `.md` trick; cleanup preserves meaning (I8) |
| Fetch | `get_document_by_url` (optional full body) | `get_doc` bounded slice + `get_doc_outline` (I4, I7) |
| Structure tool | `analyze_document_structure` | `get_doc_outline` |
| Discovery | `find_related_docs` | `list_products` |
| Distinctive | Coherence/consistency tools | Version-aware channel resolution (I6), citations (I5) |

## Things worth borrowing

- **`find_related_docs`** maps to a gap: `mcp-doc-server` has `list_products` for
  top-level discovery but nothing for "related to this topic."
- **`analyze_document_structure`** is close to the planned `get_doc_outline`; including
  parent pages / navigation context is a useful idea for citations.
- **Hosted HTTP option** — v1 is stdio, but Elastic shows demand for a zero-install hosted
  endpoint.

## Things mcp-doc-server does differently (deliberately)

- **No server-side inference** (I2) vs. Elastic leaning on AI summaries and semantic search.
  Cheaper and deterministic, but won't match semantic recall without a ranking strategy
  (still an open question in `SPECS.md`).
- **Bounded slices by default** (I4) vs. their optional-full-body model — more token-
  disciplined.
- **Version awareness** (I6) — Elastic's tools do not expose a version/channel concept the
  way Grafana's multi-version docs require.

## Builder's perspective

Fabrizio Ferri Benedetti (Elastic) wrote about building a docs MCP server
([post](https://passo.uno/mcp-server-docs-tooling/), Oct 2025). Key takeaways that
reinforce mcp-doc-server decisions:

- **Deterministic tools + LLM reasoning is a "killer combo."** His experience with
  vale-mcp-server confirms that giving agents precise, structured context and letting the
  model do the reasoning beats embedding everything server-side — the same bet behind I2.
- **Distribution UX is the main friction.** Setting up stdio servers means "editing obscure
  JSON config files"; hosted HTTP (URL-as-config) is the path to broader adoption. Aligns
  with the v2 hosted-endpoint plan flagged in the ecosystem survey.
- **Self-describing error recovery matters.** His server detects when Vale isn't installed
  and suggests how to fix it — the same pattern as I13 (actionable empty results). Real
  practitioner validation that guidance-on-failure improves agent autonomy.

## Sources

- <https://github.com/elastic/docs-builder/blob/main/docs/mcp/index.md>
- <https://passo.uno/mcp-server-docs-tooling/>
