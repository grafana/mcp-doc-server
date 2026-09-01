---
title: Install and connect
menuTitle: Install and connect
description: Build Grafana Docs MCP Server and connect it to Cursor, Claude Desktop, Claude Code, or any stdio MCP client.
weight: 5
topicType: task
versionDate: 2026-09-01
---

# Install and connect

Build Grafana Docs MCP Server and add it to your [Model Context Protocol (MCP)](https://modelcontextprotocol.io) client. The server reads Grafana Labs documentation from `grafana.com/docs/` — current published pages only.

## Before you begin

To follow this guide, you need:

- Go 1.26 or later
- Network access to `grafana.com` (for the index and doc pages)
- An MCP-compatible client (Cursor, Claude Desktop, Claude Code, or any stdio client)

## Install with `go install`

```bash
go install github.com/grafana/mcp-doc-server/cmd/mcp-doc-server@latest
go install github.com/grafana/mcp-doc-server/cmd/docs@latest
```

The binaries land in `$(go env GOPATH)/bin/`. Put that directory on your `PATH`.

## Build from source

```bash
git clone https://github.com/grafana/mcp-doc-server.git
cd mcp-doc-server
go build -o bin/mcp-doc-server ./cmd/mcp-doc-server/
go build -o bin/docs ./cmd/docs/
```

## Run the server

```bash
./bin/mcp-doc-server
```

On startup it logs two lines to `stderr`, then waits for JSON-RPC on `stdin`:

```
mcp-doc-server 0.1.0: loading index from https://grafana.com/llms-full.txt
index loaded: 7211 entries, 26 products
```

There's no prompt after that. The server is waiting for your client. The index is about 1.4 MB and usually loads in 1 to 3 seconds. Entry and product counts change as Grafana Labs publishes documentation.

Stdio only — `stdin`/`stdout`. HTTP and SSE aren't supported.

## Connect your client

The JSON is the same for every client. Only the file location changes:

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/absolute/path/to/bin/mcp-doc-server"
    }
  }
}
```

Use an absolute path. After `go install`, that's typically `<GOPATH>/bin/mcp-doc-server`. Run `go env GOPATH` to find it.

- **Cursor**: `.cursor/mcp.json` in your project root. Confirm the connection in **MCP settings**; `grafana-docs` should appear with its four tools.
- **Claude Desktop**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows). Restart Claude Desktop after editing; the tools appear under **MCP tools**.
- **Claude Code**: `.mcp.json` in your project root. Run `/mcp` to confirm `grafana-docs` is listed.

### Run without a pre-built binary

Point the client at `go run`:

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

Set `DOCS_INDEX_URL` (HTTPS only) in the `env` field:

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

Confirm the CLI:

```bash
./bin/docs search "traceql query" --product tempo --limit 3
```

You should see a ranked table of matching pages. For sample output, refer to [Get started](../get-started/#search-for-a-page).

Then ask your agent something that needs a docs lookup, for example "How does Loki retention work?" It should call `search_docs`, then `get_doc_outline` and `get_doc`, and cite a `grafana.com` URL.

## Update

Repeat the method you used:

- `go install`: re-run the install commands with `@latest`.
- From source:

```bash
git pull
go build -o bin/mcp-doc-server ./cmd/mcp-doc-server/
go build -o bin/docs ./cmd/docs/
```

Restart the server so the client loads the new binary. The running version is in the first startup log line.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Build fails | Check that your Go version is 1.26 or later (`go version`). |
| Server exits on startup with an index error | The server can't reach `grafana.com`. Check network access and any proxy settings. |
| `search` returns nothing | Usually a query or product-filter mismatch. Try different terms or drop `--product`. A running process already loaded the index; a failed load exits at startup and shows up in the `stderr` logs. |
| `get_doc` fails mid-session | The index stays in memory after startup, so `search_docs` and `list_products` still work offline. `get_doc` and `get_doc_outline` fetch pages live and return errors if the network drops. |
| Tools don't appear in the client | The `command` path must be absolute. Restart the client after editing its configuration. |
| Permission denied running the binary | Run `chmod +x ./bin/mcp-doc-server` after building. |
| Need server logs | Diagnostics go to `stderr`. Check your client's MCP log output. |

## Next steps

- [Tools reference](../tools/): parameters, CLI commands, and a full workflow
- [Configure the server](../configure/): environment variables and built-in limits
