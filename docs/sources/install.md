---
title: Install and configure
menuTitle: Install and configure
description: Build mcp-doc-server from source and connect it to Cursor, Claude Desktop, Claude Code, or any MCP client that supports stdio transport.
weight: 3
topicType: task
versionDate: 2026-06-25
---

# Install and configure

Build mcp-doc-server and connect it to your Model Context Protocol (MCP) client. The server works exclusively with Grafana Labs docs served from `grafana.com/docs/`.

## Before you begin

You need:

- Go 1.25 or later
- Network access to `grafana.com` (for the index and doc pages)
- An MCP-compatible client (Cursor, Claude Desktop, Claude Code, or any stdio client)

## Install with `go install`

Install both binaries directly:

```bash
go install github.com/grafana/mcp-doc-server/cmd/mcp-doc-server@latest
go install github.com/grafana/mcp-doc-server/cmd/docs@latest
```

The binaries are installed to `$(go env GOPATH)/bin/`. Make sure that directory is in your `PATH`.

## Build from source

Alternatively, clone the repository and build both the MCP server and the standalone CLI:

```bash
git clone git@github.com:grafana/mcp-doc-server.git
cd mcp-doc-server
go build -o bin/mcp-doc-server ./cmd/mcp-doc-server/
go build -o bin/docs ./cmd/docs/
```

## Run the server

Start it in stdio mode:

```bash
./bin/mcp-doc-server
```

On startup it logs two lines to `stderr`, then waits for JSON-RPC input on `stdin`:

```
mcp-doc-server 0.1.0: loading index from https://grafana.com/llms-full.txt
index loaded: 6645 entries, 24 products
```

There's no prompt after that. The server is ready and waiting for your MCP client to connect. The index is approximately 2 MB and typically loads in 1–3 seconds.

The server supports stdio transport only. It communicates over stdin/stdout using the MCP JSON-RPC protocol. HTTP and SSE transports are not supported.

## Connect your client

Add mcp-doc-server to your client's MCP configuration. The JSON structure is the same for every client. Only the file location differs:

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/absolute/path/to/bin/mcp-doc-server"
    }
  }
}
```

Use an absolute path to the binary. If you used `go install`, the path is typically `<GOPATH>/bin/mcp-doc-server`. Run `go env GOPATH` to find it.

Add this block to the configuration file for your client:

- **Cursor**: `.cursor/mcp.json` in your project root. Confirm the connection in Cursor's MCP settings panel; `grafana-docs` should appear with its four tools.
- **Claude Desktop**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows). Restart Claude Desktop after editing; the tools appear under the MCP tools menu.
- **Claude Code**: `.claude/mcp.json` in your project root. Run `/mcp` in Claude Code to confirm `grafana-docs` is listed.

### Run without a pre-built binary

To skip the build step, point the client at `go run`:

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "go",
      "args": ["run", "/absolute/path/to/mcp-doc-server/cmd/mcp-doc-server/"]
    }
  }
}
```

### Use a custom index

To load a different index, set `DOCS_INDEX_URL` (HTTPS only) via the `env` field:

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/absolute/path/to/bin/mcp-doc-server",
      "env": {
        "DOCS_INDEX_URL": "https://example.com/custom-index.txt"
      }
    }
  }
}
```

## Verify

Confirm the CLI works:

```bash
./bin/docs search "alerting rules" --limit 3
```

You should see a table of matching pages:

```
TITLE                      PRODUCT  URL
Configure alerting         Grafana  https://grafana.com/docs/grafana/latest/alerting/
Alerting rules             Grafana  https://grafana.com/docs/grafana/latest/alerting/alerting-rules/
...
```

Then confirm the client connection by asking your agent a question that needs a docs lookup, for example, "How does Loki retention work?" The agent should call `search_docs`, then `get_doc_outline` and `get_doc`, and cite a `grafana.com` URL.

For details on all tool parameters and output formats, refer to [Tools and CLI reference](../tools/).

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Build fails | Check your Go version is 1.25 or later (`go version`). |
| Server exits on startup with an index error | The server can't reach `grafana.com`. Check network access and any proxy settings. |
| `search` returns nothing | Usually a query mismatch. If every search is empty, the index failed to load; check the `stderr` startup logs. |
| `get_doc` fails mid-session | The index is cached in memory after startup, so `search_docs` and `list_products` continue working offline. However, `get_doc` and `get_doc_outline` fetch pages live and will return errors if the network drops. |
| Tools don't appear in the client | The `command` path must be absolute, and the client must be restarted after editing its configuration. |
| Permission denied running the binary | Run `chmod +x ./bin/mcp-doc-server` after building. |
| Need server logs | The server writes diagnostics to `stderr`; check your client's MCP log output. |

## Next steps

- [Tools reference](../tools/): tool parameters, CLI commands, and a full workflow example
- [Configuration](../configure/): environment variables and built-in limits
