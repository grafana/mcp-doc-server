# Demo Scenarios

Real-world agent workflows showing mcp-doc-server tools in action.
Each scenario shows the complete tool call sequence an agent would use.

## Scenarios

| # | Scenario | Use case | Key insight |
|---|----------|----------|-------------|
| 01 | [Grounding answers](01-grounding-answers.md) | Agent answers from live docs instead of training data | No hallucination, always current |
| 02 | [Token-efficient retrieval](02-token-efficient-retrieval.md) | Outline → section fetch instead of full page dump | 10x token reduction |
| 03 | [Coding agent config lookup](03-coding-agent-config-lookup.md) | Mid-task syntax pull during code generation | Deterministic, < 100ms |
| 04 | [Troubleshooting assistant](04-troubleshooting-assistant.md) | Error message → diagnosis → fix from docs | Cross-product search |
| 05 | [Product discovery & onboarding](05-product-discovery-onboarding.md) | Map 30+ products, guide to right starting point | list_products as compass |
| 06 | [Multi-tool workflow](06-multi-tool-workflow.md) | Docs + mcp-grafana together for full-cycle tasks | The two-server complement |

## Common patterns

All scenarios follow the same retrieval strategy:

```
1. search_docs     → find the right page(s)
2. get_doc_outline → understand page structure
3. get_doc         → fetch only what's needed (section or offset/limit)
4. Cite the URL    → user can verify
```

The outline step is what makes this token-efficient — it's a cheap metadata
call that prevents wasteful full-page fetches.

## Running the scenarios

Each scenario includes CLI commands you can run directly:

```bash
# Build the CLI first
go build -o bin/docs ./cmd/docs/

# Then run any command from the scenarios
./bin/docs search "rate limiting"
./bin/docs outline https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/
./bin/docs get https://grafana.com/docs/tempo/latest/metrics-generator/span-metrics/ --section "Configuration"
```

Or run the full demo script:

```bash
./demo/run-demo.sh
```
