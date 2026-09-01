---
title: Get started
menuTitle: Get started
description: Search, outline, and fetch Grafana documentation from the CLI before you install the MCP server or connect a client.
weight: 4
topicType: tutorial
versionDate: 2026-09-01
---

# Get started

Using this walkthrough, you can search Grafana documentation and fetch a section from the CLI before you install the server or connect an MCP client.

The CLI serves current published documentation only. Version-specific retrieval isn't available.

## Before you begin

To follow this guide, you need:

- Go 1.26 or later
- Git
- Network access to `grafana.com`

## Search for a page

1. Clone the repository:

   ```bash
   git clone https://github.com/grafana/mcp-doc-server.git
   cd mcp-doc-server
   ```

2. Run a search:

   ```bash
   go run ./cmd/docs/ search "traceql query" --product tempo --limit 3
   ```

   `go run` compiles on every invocation. For repeated use, build once with `go build -o bin/docs ./cmd/docs/` and run `./bin/docs` instead.

You should see a ranked table. Search results keep the `.md` suffix from the documentation index:

```
TITLE                           PRODUCT        URL
Construct a TraceQL query       Grafana Tempo  https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries.md
Tune TraceQL query performance  Grafana Tempo  https://grafana.com/docs/tempo/latest/traceql/tune-traceql-queries.md
TraceQL metrics sampling        Grafana Tempo  https://grafana.com/docs/tempo/latest/metrics-from-traces/metrics-queries/sampling-guide.md
```

If that table appears, Go built the CLI and reached the index on `grafana.com`.

## Fetch a section

Pick a URL from the results and fetch one heading instead of the whole page:

```bash
go run ./cmd/docs/ get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries.md --section "Comparison operators"
```

The output is cleaned Markdown: front matter, shortcodes, and HTML comments are gone.

## Run the full demo

The repository includes a script that hits all four tools with live output:

```bash
./docs/design/demo/run-demo.sh
```

It walks through `list_products`, `search_docs`, `get_doc_outline`, and `get_doc` (bounded and section-scoped), plus JSON output. Full agent workflows — grounding answers, token-efficient retrieval, configuration lookups during coding — live in `docs/design/demo/scenarios/`.

For a worked example of how an agent uses these tools, refer to [Tools and CLI reference](../tools/#workflow).

## Next steps

You've searched the index and fetched a section. Next:

- [Install and connect](../install/): build the server and connect Cursor, Claude Desktop, or Claude Code
- [Tools reference](../tools/): parameters, output formats, and CLI flags
