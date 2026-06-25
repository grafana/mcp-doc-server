# mcp-doc-server Demo

A docs-retrieval MCP server that gives AI agents version-aware access to Grafana Labs
product documentation — no embeddings, no server-side LLM, just fast deterministic
search and bounded retrieval from the official index.

## What it does

| Tool | Purpose |
|------|---------|
| `search_docs` | Find relevant docs with TF-IDF ranked results |
| `get_doc` | Fetch cleaned markdown with section extraction and paging |
| `get_doc_outline` | Cheap heading outline for targeted retrieval |
| `list_products` | Discover all product documentation groups |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  AI Agent (Cursor, Claude Desktop, custom)                  │
└────────────┬────────────────────────────────────────────────┘
             │ MCP (stdio)
┌────────────▼────────────────────────────────────────────────┐
│  mcp-doc-server                                            │
│  ┌──────────────────┐  ┌─────────────────────────────────┐ │
│  │ search_docs      │  │ get_doc (bounded, section-aware) │ │
│  │ list_products    │  │ get_doc_outline                  │ │
│  └────────┬─────────┘  └────────────────┬────────────────┘ │
│           │                              │                   │
│  ┌────────▼─────────┐  ┌────────────────▼────────────────┐ │
│  │ llms-full.txt    │  │ grafana.com/docs/*.md            │ │
│  │ (official index) │  │ (live fetch, allowlisted)        │ │
│  └──────────────────┘  └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Quick start

### Build

```bash
cd /path/to/mcp-doc-server
go build -o bin/mcp-doc-server ./cmd/mcp-doc-server/
go build -o bin/docs ./cmd/docs/
```

### Run the MCP server (stdio)

```bash
./bin/mcp-doc-server
# Loads the index, then accepts MCP JSON-RPC on stdin/stdout
```

### Run the standalone CLI

```bash
./bin/docs search "rate limiting"
./bin/docs get https://grafana.com/docs/tempo/latest/
./bin/docs outline https://grafana.com/docs/loki/latest/
./bin/docs products
```

## Demo contents

| File | What it shows |
|------|---------------|
| [run-demo.sh](run-demo.sh) | Executable script running all 4 tools with live output |
| [mcp-config/](mcp-config/) | Drop-in configs for Cursor and Claude Desktop |
| [scenarios/](scenarios/) | Real-world agent workflows showing the tools in action |
| [gcx-integration.md](gcx-integration.md) | How gcx imports the core for `gcx docs` commands |
| [mcp-grafana-integration.md](mcp-grafana-integration.md) | How mcp-grafana adds doc tools to its server |

## Integration model

The project is structured as a **reusable core + thin adapters**:

```
pkg/grafanadocs/          ← Public core (zero MCP/CLI deps)
├── mcp/                  ← MCP adapter (mark3labs/mcp-go)
└── cli/                  ← Cobra adapter (standalone CLI)

Consumers import the core only:
├── grafana/mcp-grafana   → tools/docs.go (MustTool wrappers)
└── grafana/gcx           → cmd/gcx/docs/ (output.Options + tables)
```

Both mcp-grafana and gcx import `pkg/grafanadocs` directly and write their own
idiomatic wrappers — they never import the MCP or CLI adapters.

## Key design choices

- **No embeddings, no server-side LLM** — TF-IDF with word-boundary matching,
  title weighting, and phrase bonuses. Cheap, fast, deterministic.
- **Bounded by default** — `get_doc` returns a slice, not the whole page.
  Agents use `get_doc_outline` first to target exactly what they need.
- **Allowlisted fetches** — only `https://grafana.com/docs/` URLs pass the
  allowlist. No arbitrary URL fetching.
- **Rate-limited** — 5 concurrent fetches max, 200ms minimum gap.
- **Clean markdown** — frontmatter, shortcodes, HTML comments stripped;
  code blocks preserved intact.
