# mcp-grafana Integration

**mcp-grafana** imports the `pkg/grafanadocs` core and registers four documentation
tools alongside its 30+ existing tool categories (dashboards, alerting, datasources,
incidents, etc.).

## How it works

```
mcp-grafana (github.com/grafana/mcp-grafana)
└── tools/docs.go
    ├── SearchDocsTool   → mcpgrafana.MustTool("search_docs", searchDocs)
    ├── GetDocTool       → mcpgrafana.MustTool("get_doc", getDoc)
    ├── GetDocOutlineTool → mcpgrafana.MustTool("get_doc_outline", getDocOutline)
    └── ListProductsTool → mcpgrafana.MustTool("list_products", listProducts)
```

A single `AddDocsTools(srv)` call registers all four tools on the MCP server.

## Integration pattern

mcp-grafana uses its own `MustTool` helper which generates JSON schemas from Go
struct tags — different from the raw `mcp-go` API that the standalone server uses:

```go
type SearchDocsParams struct {
    Query   string `json:"query" jsonschema:"required,description=Search query for Grafana documentation"`
    Product string `json:"product,omitempty" jsonschema:"description=Filter results to a specific product"`
    Limit   int    `json:"limit,omitempty" jsonschema:"description=Maximum results to return (default 5)"`
}

func searchDocs(ctx context.Context, args SearchDocsParams) (*SearchDocsResult, error) {
    idx, err := loadDocsIndex(ctx)
    if err != nil {
        return nil, fmt.Errorf("load docs index: %w", err)
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
)
```

## Index lifecycle

The index is loaded lazily on first use via `sync.Once`:

- `search_docs` → triggers index load
- `list_products` → triggers index load
- `get_doc` → only calls `FetchDoc` (no index needed)
- `get_doc_outline` → only calls `FetchDoc` (no index needed)

This means the mcp-grafana server starts fast — it doesn't block on index
loading at startup. The index fetches on first search/products call.

## What users see

When using mcp-grafana (e.g. through Cursor, Claude Desktop, or any MCP client),
the docs tools appear alongside all other Grafana tools:

```
Available tools (mcp-grafana):
  search_dashboards    - Search for dashboards
  get_dashboard        - Get a dashboard by UID
  search_docs          - Search Grafana documentation    ← NEW
  get_doc              - Fetch a documentation page      ← NEW
  get_doc_outline      - Get heading outline             ← NEW
  list_products        - List product doc groups         ← NEW
  list_datasources     - List datasources
  query_prometheus     - Run a PromQL query
  ...
```

## Example agent session (via mcp-grafana)

```
User: How do I set up Loki with S3 storage?

Agent: [calls search_docs(query="loki s3 storage configuration")]
       → finds the storage config page

Agent: [calls get_doc_outline(url="https://grafana.com/docs/loki/latest/configure/storage/")]
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
type-safe params, and error wrapping consistently across all 30+ tool categories.
Importing the raw adapter would introduce a second registration style.

Instead, mcp-grafana imports only `pkg/grafanadocs` (the core) and writes
~245 lines of tool wrappers that follow its own conventions perfectly.

## Code location

```
github.com/grafana/mcp-grafana/
├── tools/
│   ├── docs.go              ← Tool implementations + registration
│   ├── docs_unit_test.go    ← Unit tests
│   └── ...                  ← 30+ other tool files
└── go.mod                   ← depends on github.com/grafana/mcp-doc-server
```
