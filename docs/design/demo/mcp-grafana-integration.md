# mcp-grafana Integration

**mcp-grafana** (Grafana MCP server) imports the `pkg/grafanadocs` core and
registers two documentation tools alongside its other tool categories
(dashboards, alerting, datasources, incidents, etc.).

The Docs MCP server in this repo still exposes four tools
(`search_docs`, `get_doc`, `get_doc_outline`, `list_products`). Grafana MCP
folds those into two wrappers so the host adds fewer schema tokens per session.
The workflow is the same: search, then outline, then fetch a section.

## How it works

```
mcp-grafana (github.com/grafana/mcp-grafana)
└── tools/docs.go
    ├── SearchDocsTool → mcpgrafana.MustTool("search_docs", searchDocs)
    └── GetDocTool     → mcpgrafana.MustTool("get_doc", getDoc)
```

`AddDocsTools(srv)` registers both tools. The `docs` category is on by default;
`--disable-docs` turns them off.

| Grafana MCP tool | Docs MCP server equivalent |
|------------------|----------------------------|
| `search_docs` | `search_docs`. Omit `query` to list products (`list_products`). |
| `get_doc` | `get_doc`. Set `outline_only=true` for headings (`get_doc_outline`). |

## Integration pattern

mcp-grafana uses its own `MustTool` helper which generates JSON schemas from Go
struct tags: different from the raw `mcp-go` API that the standalone server uses:

```go
type SearchDocsParams struct {
    Query   string `json:"query,omitempty" jsonschema:"description=Search query for Grafana documentation. Omit to list all available product groups."`
    Product string `json:"product,omitempty" jsonschema:"description=Filter results to a specific product"`
    Limit   int    `json:"limit,omitempty" jsonschema:"description=Maximum results to return (default 5)"`
}

func searchDocs(ctx context.Context, args SearchDocsParams) (*SearchDocsResult, error) {
    idx, err := loadDocsIndex(ctx)
    if err != nil {
        return nil, fmt.Errorf("load docs index: %w", err)
    }
    if args.Query == "" {
        // list products
    }
    entries := grafanadocs.Search(idx, args.Query, grafanadocs.SearchOpts{
        Product: args.Product,
        Limit:   args.Limit,
    })
    // ... wrap results
}

var SearchDocsTool = mcpgrafana.MustTool(
    "search_docs",
    "Search Grafana documentation.",
    searchDocs,
    mcp.WithReadOnlyHintAnnotation(true),
    mcp.WithDestructiveHintAnnotation(false),
    mcp.WithOpenWorldHintAnnotation(true),
)
```

`get_doc` takes `outline_only`, `section`, `offset`, and `limit`. An outline-only
response returns headings and omits `returned_range`.

## Index lifecycle

The index is loaded lazily on first **successful** `search_docs` call (including
an empty query that lists products). `get_doc` only calls `FetchDoc`, so it never
triggers the load.

Failures are not cached: a cancelled or timed-out first request must not poison
later ones. Concurrent first loads share one fetch (`singleflight`). The fetch is
detached from the caller's cancellation (`context.WithoutCancel`) and bounded by
its own 60s timeout. A one-shot `sync.Once` is deliberately not used: it would
store the first error forever.

The Grafana MCP server starts fast. It doesn't block on index loading at
startup.

## What users see

When using Grafana MCP server (for example through Cursor, Claude Desktop, or
any MCP client), the docs tools appear alongside the other Grafana tools:

```
Available tools (mcp-grafana):
  search_dashboards    - Search for dashboards
  get_dashboard        - Get a dashboard by UID
  search_docs          - Search Grafana docs (omit query to list products)
  get_doc              - Fetch a page (outline_only for headings)
  list_datasources     - List datasources
  query_prometheus     - Run a PromQL query
  ...
```

## Example agent session (via mcp-grafana)

```
User: How do I set up Loki with S3 storage?

Agent: [calls search_docs(query="loki s3 storage configuration")]
       → finds the storage config page

Agent: [calls get_doc(url="https://grafana.com/docs/loki/latest/configure/storage/", outline_only=true)]
       → sees sections: "S3", "GCS", "Azure", "Filesystem"

Agent: [calls get_doc(url=..., section="S3")]
       → gets the exact S3 configuration block

Agent: Here's how to configure Loki with S3 storage:
       ```yaml
       storage_config:
         aws:
           s3: s3://region/bucket-name
           ...
       ```
       Source: https://grafana.com/docs/loki/latest/configure/storage/
```

## Why mcp-grafana doesn't use the MCP adapter

The standalone server's MCP adapter (`pkg/grafanadocs/mcp`) uses raw `mcp-go`
tool registration:

```go
srv.AddTool(s.searchDocsTool(), s.handleSearchDocs)
```

mcp-grafana has its own `MustTool` pattern that handles schema generation,
type-safe params, and error wrapping consistently across all of its tool
categories. Importing the raw adapter would introduce a second registration
style, and would register four tools instead of the two this host chose.

Instead, mcp-grafana imports only `pkg/grafanadocs` (the core) and writes
thin `MustTool` wrappers that follow its own conventions.

## Code location

```
github.com/grafana/mcp-grafana/
├── tools/
│   ├── docs.go              ← Tool implementations + registration
│   ├── docs_unit_test.go    ← Unit tests
│   └── ...                  ← other tool files
└── go.mod                   ← depends on github.com/grafana/mcp-doc-server
```
