# Scenario: Token-efficient bounded retrieval

**Use case:** Instead of dumping a whole page into the agent's context window,
the agent uses outline → targeted fetch to retrieve only what it needs.
This is the core efficiency play.

**Problem:** A Grafana Loki page is 400 lines. Fetching the whole thing wastes
tokens and pushes other context out of the window. The agent only needs one
specific section.

## Agent workflow

```
User: What LogQL query syntax does Loki support for metric queries?
```

### Step 1: Search

```bash
docs search "LogQL metric queries" --product loki --limit 3
```

```
TITLE                     PRODUCT        URL
Metric queries            Grafana Loki   https://grafana.com/docs/loki/latest/query/metric_queries/
LogQL query examples      Grafana Loki   https://grafana.com/docs/loki/latest/query/query-examples/
Log queries               Grafana Loki   https://grafana.com/docs/loki/latest/query/log_queries/
```

### Step 2: Outline the target page

```bash
docs outline https://grafana.com/docs/loki/latest/query/metric_queries/
```

```
LVL  HEADING                        LINE
1    Metric queries                  1
2    Range vector aggregation        8
2    Built-in aggregation operators  45
2    Unwrap expression               89
3    Supported functions             102
2    Examples                        150
```

**Agent decides:** "Range vector aggregation" (lines 8-44) is what I need — just 37 lines.

### Step 3: Fetch bounded slice

```bash
docs get https://grafana.com/docs/loki/latest/query/metric_queries/ \
  --section "Range vector aggregation"
```

**Result:** 37 lines of focused content, not 400 lines of the entire page.

### The alternative (without this server)

Without bounded retrieval, an agent would either:
1. Fetch the entire 400-line page → wastes ~3,000 tokens of context
2. Rely on training data → risks outdated syntax examples
3. Use a generic web search → gets HTML with navigation chrome, ads, etc.

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
| `get_doc_outline` | Map the page structure (cheap — no full content fetch) |
| `get_doc` (section) | Retrieve only the 37 lines needed |
