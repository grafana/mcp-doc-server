# Scenario: Multi-tool agentic workflow (mcp-grafana + docs)

**Use case:** A coding or ops agent combines live instance data (via mcp-grafana)
with documentation (via mcp-doc-server) to complete a task that requires both
current state awareness and correct syntax.

**Problem:** An agent needs to add a new recording rule to an existing Grafana Mimir
deployment. It needs to know: (1) what rules already exist (live data), and (2) the
correct YAML format for recording rules (docs).

## Agent workflow

```
User: Add a recording rule to my Mimir ruler that pre-computes the p99 latency
      for the payments service, grouped by region.
```

### Step 1: Check existing rules (mcp-grafana — live instance)

```
Agent calls mcp-grafana: list_prometheus_rules()
→ Returns current rule groups, namespaces, existing rules
```

The agent now knows the namespace structure and existing rules.

### Step 2: Look up recording rule syntax (mcp-doc-server — docs)

```bash
docs search "recording rules configuration" --product mimir --limit 3
```

```
TITLE                           PRODUCT         URL
Configure recording rules       Grafana Mimir   https://grafana.com/docs/mimir/latest/references/architecture/components/ruler/recording-rules/
Ruler configuration             Grafana Mimir   https://grafana.com/docs/mimir/latest/configure/configuration-parameters/#ruler
Mimirtool rules                 Grafana Mimir   https://grafana.com/docs/mimir/latest/manage/tools/mimirtool/#rules
```

```bash
docs get https://grafana.com/docs/mimir/latest/references/architecture/components/ruler/recording-rules/ \
  --section "Configuration"
```

### Step 3: Look up PromQL histogram syntax (docs)

```bash
docs search "histogram_quantile recording rule" --limit 3
```

The agent pulls the exact `histogram_quantile` function documentation to ensure
correct syntax.

### Step 4: Generate the rule (combining both sources)

```yaml
groups:
  - name: payments_latency
    rules:
      - record: payments:request_duration_seconds:p99_by_region
        expr: |
          histogram_quantile(0.99,
            sum(rate(request_duration_seconds_bucket{service="payments"}[5m])) by (le, region)
          )
        labels:
          severity: info
```

### Step 5: Apply the rule (mcp-grafana — live instance)

```
Agent calls mcp-grafana: push_prometheus_rules(namespace, group, rules)
→ Rule applied to the live Mimir ruler
```

## The two-server dance

```
┌───────────────────────────────────────────────────────────────┐
│  AI Agent                                                      │
├───────────────────────┬───────────────────────────────────────┤
│  mcp-grafana          │  mcp-doc-server                      │
│  (live instance)      │  (documentation)                      │
├───────────────────────┼───────────────────────────────────────┤
│  "What exists now?"   │  "What's the correct syntax?"         │
│  "Apply this change"  │  "What are the config options?"       │
│  "Read this metric"   │  "What does this error mean?"         │
└───────────────────────┴───────────────────────────────────────┘
```

Together, they give an agent both **awareness** (what's running) and
**knowledge** (how things should be configured).

## Why you need both

| Without docs server | Without live instance server |
|--------------------|-----------------------------|
| Agent guesses syntax → may generate invalid config | Agent writes correct config → doesn't know current state |
| Hallucinated field names that don't exist in latest | Duplicates existing rules it couldn't see |
| No link to source for user verification | Can't apply the change |

## Tools used (mcp-doc-server)

| Tool | Purpose |
|------|---------|
| `search_docs` | Find recording rule and PromQL docs |
| `get_doc` (section) | Pull exact syntax reference |

## Tools used (mcp-grafana)

| Tool | Purpose |
|------|---------|
| `list_prometheus_rules` | See existing rules |
| `push_prometheus_rules` | Apply the new rule |
