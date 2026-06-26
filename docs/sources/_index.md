---
title: mcp-doc-server
menuTitle: mcp-doc-server
description: mcp-doc-server is an MCP server that gives AI agents live access to Grafana Labs product documentation instead of relying on stale training data.
weight: 1
topicType: introduction
versionDate: 2026-06-25
---

# mcp-doc-server

mcp-doc-server is a Model Context Protocol (MCP) server that gives AI agents access to Grafana Labs product documentation. MCP is a standard that lets AI assistants call external tools during a conversation.

Agents search, browse, and retrieve current documentation from `grafana.com/docs/` during a conversation instead of relying on potentially stale training data.

{{< admonition type="note" >}}
mcp-doc-server is an internal Grafana Labs project. The [source repository](https://github.com/grafana/mcp-doc-server) is private and requires Grafana GitHub organization membership.
{{< /admonition >}}

## Why use mcp-doc-server

AI models have a training-data cutoff. When an agent needs to answer a question about a Grafana product, it works from whatever documentation existed at that cutoff. 
That training data may be outdated, incomplete, or wrong for the version the user is running. 
mcp-doc-server solves this by fetching live documentation at query time.

Because the server retrieves individual sections rather than entire pages, it also keeps token usage low. 
An agent doesn't need to load a full configuration reference into its context window to answer a question about one setting.

## How it works

The server exposes four MCP tools:

| Tool | Purpose |
|------|---------|
| `search_docs` | Find relevant pages with ranked results |
| `get_doc_outline` | Get the heading structure of a page |
| `get_doc` | Fetch cleaned Markdown for a page or a specific section |
| `list_products` | List available product documentation groups |

A typical retrieval sequence uses three steps: search for a page, get its outline, then fetch only the needed section.

![Sequence diagram showing progressive retrieval: Agent searches, outlines, then fetches a section](progressive-retrieval.svg)

For example, given the question "How does Loki retention work?", an agent can search the docs, select the most relevant page, scan its headings, and fetch only the retention section — then cite the source URL in its response.

## Design

- TF-IDF search. No embeddings or server-side large language model (LLM). Search uses TF-IDF with title weighting and phrase bonuses.
- Section-level retrieval. `get_doc` returns a section or a paged slice of a page, not the whole document.
- URL allowlist. Only `https://grafana.com/docs/` URLs are fetched. The server doesn't retrieve arbitrary URLs.
- Rate limiting. A maximum of five concurrent fetches with a 200 ms minimum gap between requests.
- Cleaned output. Front matter, shortcodes, and HTML comments are stripped. Code blocks are preserved.

For the package structure and how downstream projects import the core library, refer to [Integrate the core library](integrate/).

## Next steps

- [Get started](get-started/) — Run the server and try a query
- [Install and configure](install/) — Build the server and connect an MCP client
- [Tools reference](tools/) — Tool parameters, CLI usage, and a full workflow example
- [Configuration](configure/) — Environment variables and limits
- [Integrate the core library](integrate/) — Embed the core in your own Go project
