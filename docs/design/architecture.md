# Architecture

A dependency-light Go core (`pkg/grafanadocs`) with two thin adapters for the binaries
this repo ships, plus two downstream hosts that import the core only:

- **mcp-doc-server**: standalone MCP server (stdio) — uses the MCP adapter.
- **docs**: standalone CLI — uses the cobra adapter.
- **[gcx](https://github.com/grafana/gcx)**: `gcx docs` subcommands — imports the core, writes its own command layer.
- **[mcp-grafana](https://github.com/grafana/mcp-grafana)**: doc tools alongside the other Grafana MCP tools — imports the core, writes its own `MustTool` wrappers.

mcp-grafana and gcx do not import the MCP or CLI adapters. Those adapters exist so this
repo can ship `cmd/mcp-doc-server` and `cmd/docs`; they are not unused.

## Why an MCP server instead of prompting?

| Problem with prompting | What this solves |
|------------------------|------------------|
| Training data goes stale the day it's cut | Fetches live docs at query time. |
| Loading "just in case" context burns tokens | Surgical retrieval: fetch one section, not the whole corpus. |
| The model guesses syntax from memory | The model reads the actual reference page, then acts. |

Four tools let an agent navigate the docs the way a human would: search
(`search_docs`), scan headings (`get_doc_outline`), extract the relevant section
(`get_doc`), and list products (`list_products`).
No embeddings, no server-side LLM.

## Usage

The reusable asset is the dependency-light core, `pkg/grafanadocs`.
Load the index once (the caller owns its lifecycle), then hand it to whichever adapter you need.

```go
idx, err := grafanadocs.LoadIndex(ctx, grafanadocs.DefaultIndexURL)
```

### Standalone MCP server

```bash
mcp-doc-server   # stdio; set DOCS_INDEX_URL to override the index
```

Or from source: `go run ./cmd/mcp-doc-server/`.

### Standalone CLI

```bash
docs search "loki query" --limit 5      # text table (default)
docs search clustering -o json          # also: yaml, agents
docs outline https://grafana.com/docs/agent/latest/flow/concepts/clustering/
docs get https://grafana.com/docs/agent/latest/flow/concepts/clustering/ --section "Use cases"
docs products
```

Or from source: `go run ./cmd/docs/ <command>`.

### Mount as a subcommand in another cobra CLI

`pkg/grafanadocs/cli` exposes a mountable `docs` command group:

```go
import "github.com/grafana/mcp-doc-server/pkg/grafanadocs/cli"

rootCmd.AddCommand(cli.Command(idx)) // adds `<host> docs search|get|outline|products`
```

The adapter imports only `cobra`/`pflag` and the core.
To integrate with a host CLI's own output system and agent annotations, swap the local output helper and register the host's annotations.
The command shells follow a plain `opts`/`setup`/`Validate` pattern, so the change is mechanical.

### Register on another MCP server

```go
import mcpadapter "github.com/grafana/mcp-doc-server/pkg/grafanadocs/mcp"

mcpadapter.New(idx).Register(srv) // adds search_docs, get_doc, get_doc_outline, list_products
```

JSON output keys match across both adapters, so `docs get -o json` and the MCP `get_doc` tool return the same shape.

## Further reading

- [Research notes](research/) — prior art, ecosystem survey, cost analysis.
- [Demo scenarios](demo/) — agent workflows exercising the tools.
