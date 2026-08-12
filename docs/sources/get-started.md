---
title: Get started
menuTitle: Get started
description: Run the Grafana Docs MCP Server CLI to search, outline, and fetch Grafana documentation before installing the server or connecting an MCP client.
weight: 4
topicType: quickstart
versionDate: 2026-06-25
---

# Get started

Run the CLI to search, outline, and fetch Grafana documentation before you install the server or connect an MCP client.

## Before you begin

You need:

- Go 1.26 or later
- Git
- Network access to `grafana.com`

## Search for a page

Clone the repository and run the CLI directly from source:

```bash
git clone https://github.com/grafana/mcp-doc-server.git
cd mcp-doc-server
go run ./cmd/docs/ search "alerting rules" --limit 3
```

{{< admonition type="note" >}}
`go run` compiles the binary on every invocation. For repeated use, build once with `go build -o bin/docs ./cmd/docs/` and run `./bin/docs` instead.
{{< /admonition >}}

You should see a ranked table of matching pages:

```
TITLE                      PRODUCT  URL
Configure alerting         Grafana  https://grafana.com/docs/grafana/latest/alerting/
Alerting rules             Grafana  https://grafana.com/docs/grafana/latest/alerting/alerting-rules/
...
```

Ranked results confirm your setup works: Go builds the CLI and reaches the documentation index on `grafana.com`.

## Fetch a section

Pick a URL from the results and fetch a specific section instead of the whole page:

```bash
go run ./cmd/docs/ get https://grafana.com/docs/grafana/latest/alerting/ --section "Overview"
```

The output is cleaned Markdown with front matter, shortcodes, and HTML comments stripped.

## Run the full demo

The repository includes a demo script that exercises all four tools with live output:

```bash
./docs/design/demo/run-demo.sh
```

The script walks through `list_products`, `search_docs`, `get_doc_outline`, and `get_doc` (bounded and section-scoped), plus JSON output mode.

The `docs/design/demo/scenarios/` directory in the repository has full agent workflows: grounding answers, token-efficient retrieval, configuration lookups during coding, and more.

For a full worked example of how an agent uses these tools, refer to [Tools and CLI reference](../tools/#workflow).

## Next steps

- [Install and connect](../install/): Build the server and connect it to Cursor, Claude Desktop, or Claude Code
- [Tools reference](../tools/): Tool parameters, output formats, and CLI usage
