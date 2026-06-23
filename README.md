# hack-doc-server

Hackathon project.

A Go module that gives AI agents version-aware access to Grafana Labs product
documentation (v1: `latest`), built on the official `grafana.com/llms-full.txt` index and
the `.md` docs endpoints.

**This module is designed to work with BOTH [gcx](https://github.com/grafana/gcx) and the
[Grafana MCP server (mcp-grafana)](https://github.com/grafana/mcp-grafana).** A
dependency-light public core holds the retrieval logic, with thin opt-in adapters: an MCP
adapter (in mcp-grafana's `mark3labs/mcp-go` shape) and a cobra adapter (a `docs` command
for gcx). It also runs standalone as its own MCP server. mcp-grafana and gcx today act on a
live Grafana instance but expose no product docs — this module fills that gap.

## Usage

The reusable asset is the dependency-light core, `pkg/grafanadocs`. Load the index once
(the caller owns its lifecycle), then hand it to whichever adapter you need.

```go
idx, err := grafanadocs.LoadIndex(ctx, grafanadocs.DefaultIndexURL)
```

### Standalone MCP server

```bash
go run ./cmd/hack-doc-server   # stdio; set DOCS_INDEX_URL to override the index
```

### Standalone CLI

Runs the cobra adapter on its own (same commands gcx would expose under `gcx docs`):

```bash
go run ./cmd/docs search "loki query" --limit 5      # text table (default)
go run ./cmd/docs search clustering -o json          # also: yaml, agents
go run ./cmd/docs outline https://grafana.com/docs/agent/latest/flow/concepts/clustering/
go run ./cmd/docs get https://grafana.com/docs/agent/latest/flow/concepts/clustering/ --section "Use cases"
go run ./cmd/docs products
```

### With gcx (cobra adapter)

`pkg/grafanadocs/cli` exposes a mountable `docs` command group. In gcx's root command:

```go
import "github.com/grafana/hack-doc-server/pkg/grafanadocs/cli"

rootCmd.AddCommand(cli.Command(idx)) // adds `gcx docs search|get|outline|products`
```

The adapter imports only `cobra`/`pflag` + the core (no gcx internals). To go fully
gcx-native, swap the local output helper for gcx's `output.Options` (package
`internal/output`) and register agent annotations from gcx's own registry — the command
shells already follow gcx's `opts`/`setup`/`Validate` pattern, so the change is
mechanical. See [NOTES.md](NOTES.md) entries 16–17.

### With the Grafana MCP server (mcp-grafana)

mcp-grafana imports the **core** and registers tools in its own `mark3labs/mcp-go` shape.
This repo's `pkg/grafanadocs/mcp` adapter shows the exact mapping and can be registered
directly:

```go
import mcpadapter "github.com/grafana/hack-doc-server/pkg/grafanadocs/mcp"

mcpadapter.New(idx).Register(srv) // adds search_docs, get_doc, get_doc_outline, list_products
```

JSON output keys match across both adapters, so `gcx docs get -o json` and the MCP
`get_doc` tool return the same shape.

## Spec-driven development

See [SPECS.md](SPECS.md), [NOTES.md](NOTES.md), [TESTS.md](TESTS.md),
[BENCHMARKS.md](BENCHMARKS.md), and [AGENTS.md](AGENTS.md).
