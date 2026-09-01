# Scenario: Coding agent pulls config syntax mid-task

**Use case:** A coding or ops agent mid-task ("set up alerting", "write a dashboard
config") pulls exact syntax or config keys on demand. Deterministic retrieval with
no server-side LLM makes it cheap and predictable inside agentic loops.

**Problem:** An agent is writing a Terraform config for Grafana alerting rules.
It needs the exact structure for notification policies, not a guess.

## Agent workflow

```
User: Create a Terraform resource for a Grafana notification policy that routes
      severity=critical alerts to the PagerDuty integration and everything else
      to Slack.
```

The agent is mid-code-generation and needs exact config reference:

### Step 1: Quick search for the config reference

```bash
docs search "notification policy configuration" --product grafana --limit 3
```

```
TITLE                                         PRODUCT   URL
Configure notification alerts using Terraform Grafana   https://grafana.com/docs/grafana/latest/as-code/infrastructure-as-code/terraform/terraform-knowledge-graph/notification-alerts/
Configure notification policies               Grafana   https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/
Manage notification templates                 Grafana   https://grafana.com/docs/grafana/latest/alerting/configure-notifications/template-notifications/manage-notification-templates/
```

The second hit is the notification-policy page. The agent outlines that URL.

### Step 2: Get the outline to find the example section

```bash
docs outline https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/
```

```
LVL  HEADING                              LINE
1    Configure notification policies      3
2    Edit the default notification policy 23
2    Add a child policy                   38
2    Example                              94
```

### Step 3: Fetch the Example section

```bash
docs get https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/ \
  --section "Example"
```

### Step 4: Agent generates correct code

The agent now has a working example (including a critical-severity PagerDuty route).
It writes the Terraform resource with confidence:

```hcl
resource "grafana_notification_policy" "main" {
  contact_point = "slack-default"

  policy {
    matcher {
      label = "severity"
      match = "="
      value = "critical"
    }
    contact_point = "pagerduty-oncall"
  }
}
```

## Why deterministic retrieval matters here

- The agent may call `get_doc` 3-5 times during a single code generation task
- Each call is **< 100ms** (no embedding lookup, no LLM summarization)
- Same query always returns same content. No flaky behavior in CI/test loops
- Rate-limited to prevent abuse even under heavy agentic use (5 concurrent, 200ms gap)

## Pairing with mcp-grafana

This scenario shows docs and live instance working together:

```
1. mcp-doc-server → "Here's the correct config syntax" (docs)
2. mcp-grafana    → "Here are your existing notification policies" (live data)
3. Agent          → generates code that matches both the spec and current state
```

## Tools used

| Tool | Purpose |
|------|---------|
| `search_docs` | Find relevant config reference |
| `get_doc_outline` | Locate the example section |
| `get_doc` (section) | Pull exact syntax needed for code generation |
