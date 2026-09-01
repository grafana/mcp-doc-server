# Scenario: Token-efficient bounded retrieval

**Use case:** Instead of dumping a whole page into the agent's context window,
the agent uses outline then a targeted fetch to retrieve only what it needs.
This is the core efficiency play.

**Problem:** A Grafana Loki page is hundreds of lines. Fetching the whole thing
wastes tokens and pushes other context out of the window. The agent only needs
one specific section.

## Agent workflow

```
User: What LogQL query syntax does Loki support for metric queries?
```

### Step 1: Search

```bash
docs search "LogQL metric queries" --product loki --limit 3
```

```
TITLE            PRODUCT                  URL
Metric queries   Grafana Enterprise Logs  https://grafana.com/docs/enterprise-logs/latest/query/metric_queries/
Metric queries   Grafana Loki             https://grafana.com/docs/loki/latest/query/metric_queries/
```

### Step 2: Outline the target page

```bash
docs outline https://grafana.com/docs/loki/latest/query/metric_queries/
```

```
LVL  HEADING                        LINE
1    Metric queries                 3
2    Range Vector aggregation       11
3    Log range aggregations         19
3    Unwrapped range aggregations   61
2    Built-in aggregation operators 103
```

**Agent decides:** "Range Vector aggregation" is the section it needs.

### Step 3: Fetch bounded slice

```bash
docs get https://grafana.com/docs/loki/latest/query/metric_queries/ \
  --section "Range Vector aggregation"
```

**Result:** one section of focused content, not the entire page.

### The alternative (without this server)

Without bounded retrieval, an agent would either:

1. Fetch the entire page and waste thousands of tokens of context
2. Rely on training data and risk outdated syntax examples
3. Use a generic web search and get HTML with navigation chrome

## Token savings breakdown

| Approach | Tokens consumed | Accuracy |
|----------|----------------|----------|
| Full page in context | ~3,000 | Current |
| Training data only | 0 | May be stale |
| Outline + section fetch | ~300 | Current |

**10x token reduction** for the same quality answer.

## Tools used

| Tool | Purpose |
|------|---------|
| `search_docs` | Find the right page |
| `get_doc_outline` | Map the page structure (cheap, no full content fetch) |
| `get_doc` (section) | Retrieve only the needed section |
