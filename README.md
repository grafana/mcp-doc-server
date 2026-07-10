# mcp-doc-server

A docs-retrieval MCP server that gives AI agents and LLMs live, section-level access to Grafana Labs product documentation.

It exposes four read-only tools (search, outline, fetch, list products) over the 2,000+ pages served from `grafana.com/docs`, backed by a deterministic TF-IDF index with no embeddings and no server-side LLM.

Run it as a standalone MCP server in a client such as Cursor, Claude Desktop, or Claude Code, through the `gcx docs` CLI, or alongside the other tools in `mcp-grafana`.

**When to use it:** Use these tools to give an AI agent or LLM direct access to official Grafana documentation during a conversation, so it reads the actual content instead of guessing. Compared to a general web search, the agent gets ranked pages from `grafana.com/docs` and fetches just the section it needs as clean, citable Markdown, at a lower token cost than loading full pages.

> **Internal project.** This repository is internal and requires Grafana GitHub organization membership. Refer to the [documentation](https://grafana.com/docs/mcp-doc-server/latest/) for full usage instructions.

## Quick start

Install the server and CLI (requires `GOPRIVATE=github.com/grafana/*`):

```bash
go install github.com/grafana/mcp-doc-server/cmd/mcp-doc-server@latest
go install github.com/grafana/mcp-doc-server/cmd/docs@latest
```

Connect it to your MCP client (Cursor, Claude Desktop, Claude Code):

```json
{
  "mcpServers": {
    "grafana-docs": {
      "command": "/absolute/path/to/bin/mcp-doc-server"
    }
  }
}
```

Run `go env GOPATH` to find the binary path. For full setup instructions, refer to [Install and configure](https://grafana.com/docs/mcp-doc-server/latest/install/).

## Why an MCP server instead of prompting?

| Problem with prompting | What this solves |
|------------------------|------------------|
| Training data goes stale the day it's cut | Fetches live docs at query time: always up-to-date |
| Loading "just in case" context burns tokens | Surgical retrieval: fetch one section, not the whole corpus |
| The model guesses syntax from memory | The model *reads* the actual reference page, then acts |

Four tools let an agent navigate 2,000+ Grafana docs pages the way a human would:
search (`search_docs`), scan headings (`get_doc_outline`), extract the relevant section (`get_doc`), and
list products (`list_products`). No embeddings, no server-side LLM, deterministic and cheap.

## Architecture

A dependency-light Go core (`pkg/grafanadocs`) with thin adapters for three surfaces:

- **mcp-doc-server**: standalone MCP server (stdio) for accessing Grafana docs
- **[gcx](https://github.com/grafana/gcx)**: CLI subcommands (`gcx docs search|get|outline|products`)
- **[mcp-grafana](https://github.com/grafana/mcp-grafana)**: doc tools alongside 30+ existing Grafana MCP tools

All three import the core directly. None import each other's adapters.

## Usage

The reusable asset is the dependency-light core, `pkg/grafanadocs`. Load the index once
(the caller owns its lifecycle), then hand it to whichever adapter you need.

```go
idx, err := grafanadocs.LoadIndex(ctx, grafanadocs.DefaultIndexURL)
```

### Standalone MCP server

```bash
mcp-doc-server   # stdio; set DOCS_INDEX_URL to override the index
```

Or from source: `go run ./cmd/mcp-doc-server/`

### Standalone CLI

Runs the cobra adapter on its own (same commands gcx would expose under `gcx docs`):

```bash
docs search "loki query" --limit 5      # text table (default)
docs search clustering -o json          # also: yaml, agents
docs outline https://grafana.com/docs/agent/latest/flow/concepts/clustering/
docs get https://grafana.com/docs/agent/latest/flow/concepts/clustering/ --section "Use cases"
docs products
```

Or from source: `go run ./cmd/docs/ <command>`

### With gcx (cobra adapter)

`pkg/grafanadocs/cli` exposes a mountable `docs` command group. In gcx's root command:

```go
import "github.com/grafana/mcp-doc-server/pkg/grafanadocs/cli"

rootCmd.AddCommand(cli.Command(idx)) // adds `gcx docs search|get|outline|products`
```

The adapter imports only `cobra`/`pflag` + the core (no gcx internals). To go fully
gcx-native, swap the local output helper for gcx's `output.Options` (package
`internal/output`) and register agent annotations from gcx's own registry. The command
shells already follow gcx's `opts`/`setup`/`Validate` pattern, so the change is
mechanical. See [NOTES.md](NOTES.md) entries 16–17.

### With the Grafana MCP server (mcp-grafana)

mcp-grafana imports the **core** and registers tools in its own `mark3labs/mcp-go` shape.
This repo's `pkg/grafanadocs/mcp` adapter shows the exact mapping and can be registered
directly:

```go
import mcpadapter "github.com/grafana/mcp-doc-server/pkg/grafanadocs/mcp"

mcpadapter.New(idx).Register(srv) // adds search_docs, get_doc, get_doc_outline, list_products
```

JSON output keys match across both adapters, so `gcx docs get -o json` and the MCP
`get_doc` tool return the same shape.

## Spec-driven development

See [SPECS.md](SPECS.md), [NOTES.md](NOTES.md), [TESTS.md](TESTS.md),
[BENCHMARKS.md](BENCHMARKS.md), and [AGENTS.md](AGENTS.md).
