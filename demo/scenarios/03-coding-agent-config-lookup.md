# Scenario: Coding agent pulls config syntax mid-task

**Use case:** A coding or ops agent mid-task ("set up alerting", "write a dashboard
config") pulls exact syntax or config keys on demand. Deterministic retrieval with
no server-side LLM makes it cheap and predictable inside agentic loops.

**Problem:** An agent is writing a Terraform config for Grafana alerting rules.
It needs the exact YAML structure for notification policies — not a guess.

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
TITLE                              PRODUCT   URL
Configure notification policies    Grafana   https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/
Notification policies              Grafana   https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/notification-policies/
Manage notification policies       Grafana   https://grafana.com/docs/grafana/latest/alerting/manage-notifications/create-notification-policy/
```

### Step 2: Get the outline to find the config/example section

```bash
docs outline https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/
```

```
LVL  HEADING                    LINE
1    Configure notification...  1
2    Before you begin           5
2    Routing                    12
2    Matchers                   30
3    Label matching             35
2    Timing options             52
2    Example                    68
2    API reference              85
```

### Step 3: Fetch just the example and matchers sections

```bash
docs get https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/ \
  --section "Matchers"
```

```bash
docs get https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/ \
  --section "Example"
```

### Step 4: Agent generates correct code

The agent now has the exact label matcher syntax and a working example.
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
- Same query always returns same content — no flaky behavior in CI/test loops
- Rate-limited to prevent abuse even under heavy agentic use (5 concurrent, 200ms gap)

## Pairing with mcp-grafana

This scenario shows docs + live instance working together:

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
