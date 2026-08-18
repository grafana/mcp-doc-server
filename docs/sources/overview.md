---
title: Overview
menuTitle: Overview
description: What Grafana Docs MCP Server does, when to use it, who it's for, how retrieval works, and its design constraints.
weight: 2
topicType: introduction
versionDate: 2026-06-25
---

# Overview

Grafana Docs MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that gives AI agents access to Grafana Labs product documentation. MCP is a standard that lets AI assistants call external tools during a conversation.

Agents search, browse, and retrieve current documentation from `grafana.com/docs/` during a conversation instead of relying on potentially stale training data.

{{< admonition type="note" >}}
Grafana Docs MCP Server only retrieves the current published documentation. Future plans include being able to retrieve specific versions. 
{{< /admonition >}}

## When to use it

Use Grafana Docs MCP Server to give an AI agent or LLM direct access to official Grafana documentation during a conversation, so it reads the actual content instead of guessing.
Compared to a general web search, the agent gets ranked pages from `grafana.com/docs` and fetches just the section it needs as clean, citable Markdown, at a lower token cost than loading full pages.

## Who it's for

Grafana Docs MCP Server is for engineers and technical writers who work with Grafana products and drive an MCP client such as Cursor, Claude Desktop, or Claude Code. Go developers can also embed the core library directly in their own projects. Refer to [Integrate the core library](../integrate/).

## Which surface to use

The same doc tools ship in three places. Pick the one that matches your setup:

| Surface | Use it when |
|---------|-------------|
| **Grafana Docs MCP Server** (this project) | You want a standalone MCP server that exposes only the docs tools. |
| **[gcx](https://github.com/grafana/gcx)** | You already use the gcx CLI. The `gcx docs` commands run the same retrieval. |
| **[mcp-grafana](https://github.com/grafana/mcp-grafana)** | You already run mcp-grafana and want the docs tools alongside its other Grafana MCP tools. |

## Why use Grafana Docs MCP Server

AI models have a training-data cutoff. When an agent needs to answer a question about a Grafana product, it works from whatever documentation existed at that cutoff.
That training data may be outdated, incomplete, or wrong for the version the user is running.
Grafana Docs MCP Server solves this by fetching live documentation at query time.

Because the server retrieves individual sections rather than entire pages, it also keeps token usage low.
An agent doesn't need to load a full configuration reference into its context window to answer a question about one setting.
For how this translates into what you pay, refer to [Cost](../cost/).

## How it works

The server exposes four MCP tools:

| Tool | Purpose |
|------|---------|
| `search_docs` | Find relevant pages with ranked results |
| `get_doc_outline` | Get the heading structure of a page |
| `get_doc` | Fetch cleaned Markdown for a page or a specific section |
| `list_products` | List available product documentation groups |

A typical retrieval sequence uses three steps: search for a page, get its outline, then fetch only the needed section.

![Sequence diagram showing progressive retrieval: Agent searches, outlines, then fetches a section](../progressive-retrieval.svg)

For example, given the question "How does Loki retention work?", an agent can search the docs, select the most relevant page, scan its headings, and fetch only the retention section, then cite the source URL in its response.

You don't call these tools yourself. From an MCP client, ask the agent a question in plain language and it calls the tools and cites the source. For example:

> How does Loki retention work?

The agent runs `search_docs`, then `get_doc_outline` and `get_doc`, and answers with a `grafana.com/docs/` citation.

## Design

- TF-IDF search. No embeddings or server-side large language model (LLM). Search uses TF-IDF with title weighting and phrase bonuses.
- Section-level retrieval. `get_doc` returns a section or a paged slice of a page, not the whole document.
- URL allowlist. Only `https://grafana.com/docs/` URLs are fetched. The server doesn't retrieve arbitrary URLs.
- Rate limiting. A maximum of five concurrent fetches with a 200 ms minimum gap between requests.
- Cleaned output. Front matter, shortcodes, and HTML comments are stripped. Code blocks are preserved.

For the package structure and how downstream projects import the core library, refer to [Integrate the core library](../integrate/).

## Next steps

- [Cost](../cost/): What the server costs to run and what you pay per query
- [Get started](../get-started/): Run the server and try a query
- [Install and connect](../install/): Build the server and connect an MCP client
- [Tools reference](../tools/): Tool parameters, CLI usage, and a full workflow example
- [Configure the server](../configure/): Environment variables and limits
- [Integrate the core library](../integrate/): Embed the core in your own Go project
