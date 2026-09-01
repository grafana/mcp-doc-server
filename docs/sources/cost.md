---
title: Cost
menuTitle: Cost
description: Grafana Docs MCP Server adds no inference cost. You pay the agent's tokens for the documentation it retrieves.
weight: 3
topicType: reference
versionDate: 2026-09-01
---

# Cost

The server doesn't run an LLM and doesn't call an embedding API, so it adds no inference cost. What you pay is the agent's input tokens for the documentation it reads.

A typical lookup is three calls:

| Step | Tokens the agent reads |
|------|------------------------|
| `search_docs` (5 results) | ~600 |
| `get_doc_outline` | ~400 |
| `get_doc` (one section) | ~2,000 |
| **Workflow total** | **~3,000** |

Those numbers are the response budgets in the project's benchmarks. `get_doc` defaults to 80 lines, so a short page comes back in full and a long one comes back as a slice.

At $3.00 per million input tokens, that workflow is about $0.009. Pulling a whole configuration reference instead (around 30,000 tokens) costs roughly 10 times as much for the same answer and uses more of the context window.

{{< admonition type="note" >}}
These figures cover documentation retrieval only, not the rest of the agent's task. Model rates change, so treat the dollar amounts as examples.
{{< /admonition >}}

For server footprint, other retrieval approaches, and the methodology, refer to the [cost-analysis document](https://github.com/grafana/mcp-doc-server/blob/main/docs/design/research/cost-analysis.md) in the repository. That file isn't a published documentation page.

## Related resources

- [Tools and CLI reference](../tools/)
- [Configure the server](../configure/)
