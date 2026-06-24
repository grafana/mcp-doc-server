# MCP Client Configuration

Drop-in configurations to connect hack-doc-server to your AI client.

## Cursor IDE

Add to your Cursor MCP settings (`.cursor/mcp.json` in your project or global config):

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/path/to/hack-doc-server/bin/hack-doc-server",
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
      "command": "/path/to/hack-doc-server/bin/hack-doc-server",
      "args": []
    }
  }
}
```

## Claude Code (CLI)

```bash
claude mcp add grafana-docs /path/to/hack-doc-server/bin/hack-doc-server
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
User: How do I configure rate limiting in Tempo?

Agent: [calls search_docs("rate limiting", product="tempo")]
       → finds "Global rate limiting" page

Agent: [calls get_doc_outline(url)]
       → sees headings: "Override strategy", "Configuration", "Examples"

Agent: [calls get_doc(url, section="Configuration")]
       → retrieves just the config section (bounded, clean markdown)

Agent: Here's how to configure rate limiting in Tempo:
       [answers with citation to source URL]
```
