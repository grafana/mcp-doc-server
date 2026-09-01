# Scenario: Product discovery and onboarding

**Use case:** A new user (or their AI assistant) needs to understand what Grafana
products exist, how they relate, and where to start. The `list_products` tool
provides the map; `search_docs` finds getting-started content.

**Problem:** The index groups documentation into dozens of products. A user asks
"I need observability for my Kubernetes cluster. What should I use?" The agent
needs to discover available products, then guide the user to the right starting
point.

## Agent workflow

```
User: I want to set up full-stack observability for my Kubernetes cluster using
      Grafana Cloud. What products do I need and how do I get started?
```

### Step 1: Discover the product landscape

```bash
docs products
```

```
NAME                           COUNT
Grafana Agent                  311
Grafana Alloy                  339
Grafana Beyla                   46
Grafana Cloud                 2158
Grafana                        745
Grafana Loki                   ...
...
(26 product groups)
```

### Step 2: Search for Kubernetes-specific onboarding

```bash
docs search "Kubernetes Monitoring" --product "Grafana Cloud" --limit 3
```

```
TITLE                       PRODUCT        URL
Kubernetes Monitoring       Grafana Cloud  https://grafana.com/docs/grafana-cloud/observe-and-act/monitor-infrastructure/kubernetes-monitoring/
Configure Kubernetes Monitoring Grafana Cloud https://grafana.com/docs/grafana-cloud/observe-and-act/monitor-infrastructure/kubernetes-monitoring/configuration/
```

### Step 3: Deep-dive on the recommended starting point

```bash
docs outline https://grafana.com/docs/grafana-cloud/observe-and-act/monitor-infrastructure/kubernetes-monitoring/
```

```
LVL  HEADING              LINE
1    Kubernetes Monitoring 3
2    Overview              9
2    Guidance and help     16
2    Explore               42
```

```bash
docs get https://grafana.com/docs/grafana-cloud/observe-and-act/monitor-infrastructure/kubernetes-monitoring/ \
  --section "Overview"
```

### Step 4: Provide a structured recommendation

```
Agent: For full-stack Kubernetes observability on Grafana Cloud, here's your stack:

| Signal  | Product        | Purpose                          |
|---------|----------------|----------------------------------|
| Metrics | Grafana Mimir  | Long-term Prometheus storage     |
| Logs    | Grafana Loki   | Log aggregation (LogQL)          |
| Traces  | Grafana Tempo  | Distributed tracing              |
| Collect | Grafana Alloy  | Unified telemetry collector      |

Start here: Kubernetes Monitoring
→ https://grafana.com/docs/grafana-cloud/observe-and-act/monitor-infrastructure/kubernetes-monitoring/
```

## Why `list_products` matters

Without product discovery, an agent would:

- Miss products it doesn't know about (Beyla? Alloy? These are relatively new)
- Recommend outdated tools (Grafana Agent is now Grafana Alloy)
- Not know the breadth of the ecosystem

With it, the agent can map the user's needs to the right combination of products.

## Tools used

| Tool | Purpose |
|------|---------|
| `list_products` | Map the product groups in the index |
| `search_docs` | Find onboarding content across products |
| `get_doc_outline` | Understand page structure before fetching |
| `get_doc` (section) | Retrieve the overview content |
