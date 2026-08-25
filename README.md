# mcp-doc-server

A docs-retrieval MCP server that gives AI agents and LLMs live, section-level access to Grafana Labs product documentation.

It exposes four read-only tools (`search_docs`, `get_doc_outline`, `get_doc`, `list_products`) over the pages served from `grafana.com/docs`, backed by a deterministic TF-IDF index with no embeddings and no server-side LLM.

Run it as a standalone MCP server in a client such as Cursor, Claude Desktop, or Claude Code, through the [`gcx docs`](https://github.com/grafana/gcx) CLI, or alongside the other tools in [`mcp-grafana`](https://github.com/grafana/mcp-grafana).

## Quick start

Install the server and CLI:

```bash
go install github.com/grafana/mcp-doc-server/cmd/mcp-doc-server@latest
go install github.com/grafana/mcp-doc-server/cmd/docs@latest
```

Connect it to your MCP client:

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/absolute/path/to/bin/mcp-doc-server"
    }
  }
}
```

Run `go env GOPATH` to find the binary path.

## Future plans 

We are working on integrations with GCX and Grafana MCP server directly. Those PRs are a work in progress. 

We are also exploring a hosted Grafana Docs MCP server (no timeframe yet). 

## Documentation

- [Architecture and design notes](docs/design/): how the core and adapters fit together, prior art, and cost analysis.
- [Contributing](CONTRIBUTING.md): spec-driven development workflow.
- [Security](SECURITY.md): vulnerability reporting.

## License

Apache License 2.0. See [LICENSE](LICENSE).
