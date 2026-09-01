# MCP Client Configuration

Drop-in configurations to connect mcp-doc-server to your AI client.

## Cursor IDE

Add to your Cursor MCP settings (`.cursor/mcp.json` in your project or global config):

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/path/to/mcp-doc-server/bin/mcp-doc-server",
      "args": [],
      "env": {}
    }
  }
}
```

## Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)
or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/path/to/mcp-doc-server/bin/mcp-doc-server",
      "args": []
    }
  }
}
```

## Claude Code (CLI)

```bash
claude mcp add grafana-docs /path/to/mcp-doc-server/bin/mcp-doc-server
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DOCS_INDEX_URL` | `https://grafana.com/llms-full.txt` | Override the docs index URL (useful for testing with a local file) |

## Verify it works

Once configured, ask your AI client:

> "Search the Grafana docs for how to set up Loki"

The agent should call `search_docs` and return results with titles, URLs, and products.

## Example session

```
User: How do I construct a TraceQL query?

Agent: [calls search_docs(query="traceql query", product="tempo")]
       → finds Construct a TraceQL query

Agent: [calls get_doc_outline(url="https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/")]
       → sees headings including "Comparison operators"

Agent: [calls get_doc(url, section="Comparison operators")]
       → retrieves just that section (bounded, clean markdown)

Agent: Here's how comparison operators work in TraceQL:
       [answers with citation to source URL]
```
