# Scenario: Troubleshooting assistant

**Use case:** A support copilot or chatbot answers setup, config, migration,
and error-message questions straight from the docs. The agent combines
product discovery with targeted retrieval to diagnose issues.

**Problem:** A user hits a rate-limit error they don't understand. The agent needs
to find the relevant troubleshooting page, locate the error section, and
provide the fix, all from authoritative docs.

## Agent workflow

```
User: Loki is rejecting my pushes with a rate_limited error.
      What's causing this and how do I fix it?
```

### Step 1: Search for the error

```bash
docs search "request validation rate limits" --product loki --limit 3
```

```
TITLE                                         PRODUCT        URL
Enforce rate limits and push request validation Grafana Loki  https://grafana.com/docs/loki/latest/operations/request-validation-rate-limits/
Monitor tenant limits using the Overrides Exporter Grafana Loki https://grafana.com/docs/loki/latest/operations/overrides-exporter/
```

### Step 2: Outline the rate-limits page

```bash
docs outline https://grafana.com/docs/loki/latest/operations/request-validation-rate-limits/
```

```
LVL  HEADING                                           LINE
1    Enforce rate limits and push request validation   3
2    Terminology                                       11
2    Rate-Limit Errors                                 17
3    rate_limited                                      23
3    per_stream_rate_limit                             44
2    Validation Errors                                 94
```

### Step 3: Fetch the specific error section

```bash
docs get https://grafana.com/docs/loki/latest/operations/request-validation-rate-limits/ \
  --section "Rate-Limit Errors"
```

### Step 4: Provide actionable answer

```
Agent: The rate_limited error means the tenant exceeded Loki's ingestion
rate limit. Here's how to fix it:

[config and remediation from the fetched section]

Source: https://grafana.com/docs/loki/latest/operations/request-validation-rate-limits/
```

## Multi-product troubleshooting

The agent can also cross-reference products. If the user's issue spans
Loki, Grafana, and Tempo:

```bash
docs search "connection refused" --limit 10
# Returns results across all products. The agent can correlate.
```

```bash
docs products
# Shows every product documentation group in the index (26 groups)
```

## Why this beats a web search

| Approach | Drawback |
|----------|----------|
| Google search | Returns HTML with nav chrome, ads, SEO filler |
| Scraping grafana.com | Unstructured, breaks on layout changes |
| Training data | May reference old config keys or removed features |
| **mcp-doc-server** | Clean markdown, section-targeted, always current |

## Tools used

| Tool | Purpose |
|------|---------|
| `search_docs` | Find pages matching the error |
| `get_doc_outline` | Navigate to the exact error + fix sections |
| `get_doc` (section) | Retrieve diagnosis + fix in bounded chunks |
| `list_products` | Discover what products are available for cross-referencing |
