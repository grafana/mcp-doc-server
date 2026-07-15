# Scenario: Troubleshooting assistant

**Use case:** A support copilot or chatbot answers setup, config, migration,
and error-message questions straight from the docs. The agent combines
product discovery with targeted retrieval to diagnose issues.

**Problem:** A user hits an error they don't understand. The agent needs to
find the relevant troubleshooting page, locate the error section, and
provide the fix — all from authoritative docs.

## Agent workflow

```
User: I'm getting "too many outstanding requests" errors in my Loki deployment.
      What's causing this and how do I fix it?
```

### Step 1: Search for the error message

```bash
docs search "too many outstanding requests" --product loki
```

```
TITLE                          PRODUCT        URL
Rate-limit errors              Grafana Loki   https://grafana.com/docs/loki/latest/operations/request-validation-and-rate-limits/
Configuration                  Grafana Loki   https://grafana.com/docs/loki/latest/configure/
Scaling out                    Grafana Loki   https://grafana.com/docs/loki/latest/operations/scaling/
```

### Step 2: Outline the rate-limits page

```bash
docs outline https://grafana.com/docs/loki/latest/operations/request-validation-and-rate-limits/
```

```
LVL  HEADING                         LINE
1    Request validation and limits   1
2    Per-tenant rate limits          5
2    Global rate limits              22
2    Error messages                  45
3    too many outstanding requests   48
3    rate limit exceeded             62
2    Fixing rate limit errors        75
3    Increase limits                 78
3    Reduce cardinality              90
```

### Step 3: Fetch the specific error section and the fix

```bash
docs get https://grafana.com/docs/loki/latest/operations/request-validation-and-rate-limits/ \
  --section "too many outstanding requests"
```

```bash
docs get https://grafana.com/docs/loki/latest/operations/request-validation-and-rate-limits/ \
  --section "Fixing rate limit errors"
```

### Step 4: Provide actionable answer

```
Agent: The "too many outstanding requests" error means your Loki ingester has hit
its concurrent request limit. Here's how to fix it:

1. Increase `max_outstanding_per_tenant` in your Loki config:
   [config snippet from docs]

2. Or reduce ingestion volume by:
   [steps from docs]

Source: https://grafana.com/docs/loki/latest/operations/request-validation-and-rate-limits/
```

## Multi-product troubleshooting

The agent can also cross-reference products. If the user's issue spans
Loki + Grafana + Tempo:

```bash
docs search "connection refused" --limit 10
# Returns results across all products — agent can correlate
```

```bash
docs products
# Shows all 30+ product documentation groups available
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
| `search_docs` | Find pages matching the error message |
| `get_doc_outline` | Navigate to the exact error + fix sections |
| `get_doc` (section) | Retrieve diagnosis + fix in bounded chunks |
| `list_products` | Discover what products are available for cross-referencing |
