# Scenario: Product discovery and onboarding

**Use case:** A new user (or their AI assistant) needs to understand what Grafana
products exist, how they relate, and where to start. The `list_products` tool
provides the map; `search_docs` finds getting-started content.

**Problem:** Grafana Labs has 30+ documented products. A user asks "I need
observability for my Kubernetes cluster — what should I use?" The agent needs
to discover available products, then guide the user to the right starting point.

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
Grafana                        450
Grafana Alloy                  120
Grafana Loki                   95
Grafana Tempo                  78
Grafana Mimir                  85
Grafana Cloud                  200
Grafana Agent                  60
Grafana Beyla                  25
Grafana OnCall                 40
Grafana k6                     110
...
(30+ products)
```

### Step 2: Search for Kubernetes-specific onboarding

```bash
docs search "kubernetes observability getting started" --limit 5
```

```
TITLE                                  PRODUCT          URL
Set up Kubernetes Monitoring           Grafana Cloud    https://grafana.com/docs/grafana-cloud/monitor-infrastructure/kubernetes-monitoring/
Grafana Alloy on Kubernetes            Grafana Alloy    https://grafana.com/docs/alloy/latest/get-started/install/kubernetes/
Collect Kubernetes logs                Grafana Loki     https://grafana.com/docs/loki/latest/send-data/kubernetes/
Kubernetes with Grafana Beyla          Grafana Beyla    https://grafana.com/docs/beyla/latest/setup/kubernetes/
Monitor Kubernetes with Tempo          Grafana Tempo    https://grafana.com/docs/tempo/latest/getting-started/kubernetes/
```

### Step 3: Deep-dive on the recommended starting point

```bash
docs outline https://grafana.com/docs/grafana-cloud/monitor-infrastructure/kubernetes-monitoring/
```

```
LVL  HEADING                    LINE
1    Kubernetes Monitoring      1
2    Overview                   3
2    Before you begin           15
2    Install the Helm chart     25
2    What's collected           50
3    Metrics                    55
3    Logs                       62
3    Traces                     68
2    Next steps                 80
```

```bash
docs get https://grafana.com/docs/grafana-cloud/monitor-infrastructure/kubernetes-monitoring/ \
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

Start here: Set up Kubernetes Monitoring (one Helm chart installs everything)
→ https://grafana.com/docs/grafana-cloud/monitor-infrastructure/kubernetes-monitoring/

This installs Grafana Alloy pre-configured to collect metrics, logs, and traces
from your cluster and send them to Grafana Cloud.
```

## Why `list_products` matters

Without product discovery, an agent would:
- Miss products it doesn't know about (Beyla? Alloy? These are relatively new)
- Recommend outdated tools (Grafana Agent → now Grafana Alloy)
- Not know the breadth of the ecosystem

With it, the agent can map the user's needs to the right combination of products.

## Tools used

| Tool | Purpose |
|------|---------|
| `list_products` | Map the entire product landscape (30+ products) |
| `search_docs` | Find onboarding content across products |
| `get_doc_outline` | Understand page structure before fetching |
| `get_doc` (section) | Retrieve the overview/getting-started content |
