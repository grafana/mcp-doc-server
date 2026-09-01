# Scenario: Grounding an answer in authoritative docs

**Use case:** An agent answers a user question by retrieving the relevant doc page
instead of relying on stale training data. Especially valuable for fast-moving
products where the model's knowledge cutoff lags reality.

**Problem:** A user asks "How do I configure Tempo's metrics generator?" The LLM's
training data is 6 months old. The config format changed in the latest release.

## Agent workflow

```
User: How do I configure Tempo's metrics generator to emit span metrics?

Agent thinks: I should check the current docs rather than rely on my training data.
```

### Step 1: Search for relevant pages

```bash
docs search "metrics generator" --product tempo --limit 3
```

```
TITLE                PRODUCT        URL
Metrics-generator    Grafana Tempo  https://grafana.com/docs/tempo/latest/metrics-from-traces/metrics-generator/
Metrics-generator    Grafana Tempo  https://grafana.com/docs/tempo/latest/reference-tempo-architecture/components/metrics-generator/
```

### Step 2: Get the outline to find the right section

```bash
docs outline https://grafana.com/docs/tempo/latest/metrics-from-traces/metrics-generator/
```

```
LVL  HEADING              LINE
1    Metrics-generator    3
2    Architecture         7
3    Service graphs       25
3    Span metrics         33
2    Native histograms    53
```

### Step 3: Fetch just the Native histograms section

```bash
docs get https://grafana.com/docs/tempo/latest/metrics-from-traces/metrics-generator/ \
  --section "Native histograms"
```

The agent gets the current section, not the entire page, and not hallucinated
syntax from 6 months ago.

### Step 4: Answer with citation

```
Agent: Here's how Tempo's metrics-generator emits native histograms:

[provides answer based on fetched content]

Source: https://grafana.com/docs/tempo/latest/metrics-from-traces/metrics-generator/
```

## Why this matters

- **No hallucination risk:** the answer comes from the live docs, not training data
- **Token efficient:** only the relevant section was fetched
- **Traceable:** the source URL is included so the user can verify
- **Always current:** fetches the `latest` channel, reflecting the newest release

## Tools used

| Tool | Purpose |
|------|---------|
| `search_docs` | Find the right page in the index |
| `get_doc_outline` | Identify the exact section needed |
| `get_doc` (section) | Fetch only what's needed for the answer |
