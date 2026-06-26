---
title: Cost
menuTitle: Cost
description: How mcp-doc-server costs break down—the server adds no inference cost, so you pay only for the agent's tokens for retrieved documentation.
weight: 7
topicType: reference
versionDate: 2026-06-26
---

# Cost

mcp-doc-server has two cost layers: the server runtime and the agent's large language model (LLM) tokens. The server adds no inference cost of its own—it runs no LLM and performs no embedding calls—so the agent's token usage for retrieved documentation is the only metered cost.

## Token cost

The agent reads the documentation the server returns as input tokens. A typical retrieval workflow has three steps:

| Step | Tokens the agent reads |
|------|------------------------|
| `search_docs` (5 results) | ~600 |
| `get_doc_outline` | ~400 |
| `get_doc` (one section) | ~2,000 |
| **Workflow total** | **~3,000** |

These figures are the response budgets defined in the project's benchmarks. `get_doc` returns a bounded slice—80 lines by default—rather than a whole page.

At an example rate of $3.00 per million input tokens, a ~3,000-token workflow costs about $0.009. Retrieving a whole page instead, for example a large configuration reference can be around 30,000 tokens, costs roughly 10 times as much for the same answer and uses more of the context window.

{{< admonition type="note" >}}
These figures cover documentation retrieval only, not the agent's total token usage for a task. Model rates change over time, so treat the dollar amounts as illustrative.
{{< /admonition >}}

## Related resources

- For a full cost breakdown—server footprint, comparison with other retrieval approaches, scale projections, and methodology—refer to `research/cost-analysis.md` in the repository.
- [Tools and CLI reference](../tools/)—the search, outline, and fetch workflow
- [Configuration](../configure/)—built-in limits, including outbound rate limiting
