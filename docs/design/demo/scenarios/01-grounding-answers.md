# Scenario: Grounding an answer in authoritative docs

**Use case:** An agent answers a user question by retrieving the relevant doc page
instead of relying on stale training data. Especially valuable for fast-moving
products where the model's knowledge cutoff lags reality.

**Problem:** A user asks "How do I configure Tempo's metrics generator?" The LLM's
training data is 6 months old — the config format changed in the latest release.

## Agent workflow

```
User: How do I configure Tempo's metrics generator to emit span metrics?

Agent thinks: I should check the current docs rather than rely on my training data.
```

### Step 1: Search for relevant pages

```bash
docs search "metrics generator span metrics" --product tempo
```

```
TITLE                              PRODUCT        URL
Span metrics                       Grafana Tempo  https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/
Metrics-generator                  Grafana Tempo  https://grafana.com/docs/tempo/latest/metrics-generator/
Metrics-generator configuration    Grafana Tempo  https://grafana.com/docs/tempo/latest/configuration/#metrics-generator
```

### Step 2: Get the outline to find the right section

```bash
docs outline https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/
```

```
LVL  HEADING                    LINE
1    Span metrics               1
2    Overview                   3
2    Configuration              15
3    Filtering spans            28
3    Histogram buckets          42
2    Example                    56
```

### Step 3: Fetch just the Configuration section

```bash
docs get https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/ \
  --section "Configuration"
```

The agent gets ~30 lines of current, accurate config documentation — not the
entire 200-line page, and not hallucinated syntax from 6 months ago.

### Step 4: Answer with citation

```
Agent: Here's how to configure Tempo's span metrics generator:

[provides answer based on fetched content]

Source: https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/
```

## Why this matters

- **No hallucination risk** — the answer comes from the live docs, not training data
- **Token efficient** — only the relevant section was fetched (~30 lines vs 200+)
- **Traceable** — the source URL is included so the user can verify
- **Always current** — fetches the `latest` channel, reflecting the newest release

## Tools used

| Tool | Purpose |
|------|---------|
| `search_docs` | Find the right page in 2,000+ doc entries |
| `get_doc_outline` | Identify the exact section needed |
| `get_doc` (section) | Fetch only what's needed for the answer |
