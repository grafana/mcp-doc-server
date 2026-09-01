# gcx Integration

**gcx** imports the `pkg/grafanadocs` core directly and mounts it as `gcx docs`
subcommands with full output format support and terminal styling.

## How it works

```
gcx (github.com/grafana/gcx)
└── cmd/gcx/docs/
    ├── search.go    → calls grafanadocs.Search()
    ├── get.go       → calls grafanadocs.FetchDoc() + grafanadocs.Excerpt()
    ├── outline.go   → calls grafanadocs.FetchDoc() + grafanadocs.Outline()
    ├── products.go  → calls idx.Products()          (command: list-products)
    └── links.go     → gcx-local internal/docs registry (command: list-links)
```

gcx imports **only the core** (`pkg/grafanadocs`), not the MCP or CLI adapter.
It writes its own command layer using gcx conventions: `output.Options`, styled
tables, agent annotations, and the standard opts pattern.

`get` and `outline` take an injected `docFetcher` (`grafanadocs.FetchDoc` in
production; `CommandWithFetcher` replaces it in tests). That keeps the fetch
off the package scope (`gochecknoglobals`) and mirrors `CommandWithIndex` for
the index-backed commands.

## Commands

```bash
# Search across all Grafana docs
gcx docs search "alerting rules"

# Filter to a product (case-insensitive: exact, then prefix, then substring)
gcx docs search "traceql query" --product tempo

# Fetch a page (text mode shows raw markdown)
gcx docs get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/

# Extract just one section
gcx docs get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/ --section "Comparison operators"

# Show the heading structure
gcx docs outline https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/

# List indexed doc products with entry counts
gcx docs list-products

# List curated canonical URLs (offline; no index)
gcx docs list-links

# Advanced: page a long doc (Tempo Configure is ~3000 lines; title is Configure Tempo)
gcx docs get https://grafana.com/docs/tempo/latest/configuration/ --section "Server"
gcx docs get https://grafana.com/docs/tempo/latest/configuration/ --offset 80 --limit 80
```

Catalog leaves use `list-<subject>` so they match the gcx naming guide for
ID-less catalog facets. `list-links` is gcx-local (`internal/docs`); it is not
a `pkg/grafanadocs` call.

## Search output

JSON is an envelope, not a bare array. A capped page carries `list_meta`;
absence of `list_meta` means the page is complete. gcx requests `limit+1` and
uses `TruncatePagedList` so truncation is proven by a spare row: `Search()`
returns no total. `--limit` 0 or negative uses the default of 5 (the library
treats `<=0` as 5, so `BindListLimit`'s "0 means all" would be a lie).

```json
{
  "results": [ { "title": "...", "url": "...", "description": "...", "product": "..." } ],
  "list_meta": { "truncated": true, "returned": 5, "continue": "gcx docs search … --limit 10" }
}
```

Empty results serialize as `"results": []`, not `null`. Text mode stays a
TITLE / PRODUCT / URL table; the truncation hint is also written to stderr
via `EmitListTruncationHint`.

## Output formats

All commands support `-o` for output format selection:

```bash
# Styled table (default for human use)
gcx docs search "dashboards"

# JSON (for scripting / piping)
gcx docs search "dashboards" -o json

# YAML
gcx docs search "dashboards" -o yaml

# Compact JSON (optimized for agent consumption)
gcx docs search "dashboards" -o agents
```

## Architecture pattern

gcx follows the same pattern for docs as for all its providers:

```go
// opts struct with validation
type searchOpts struct {
    IO      cmdio.Options
    query   string
    product string
    limit   int
}

// Custom text codec for styled terminal output
type searchTextCodec struct{}
func (c *searchTextCodec) Encode(w io.Writer, v any) error {
    t := style.NewTable("TITLE", "PRODUCT", "URL")
    // ...
}
```

This means docs commands get the same UX as every other gcx command:
consistent flags, consistent output handling, agent-mode awareness.

## Index lifecycle

The index is loaded lazily: only `search` and `list-products` need it.
`get` and `outline` only call `FetchDoc`, so they work without loading
the full index (faster startup, works even if the index is temporarily
unreachable). `list-links` is offline and never touches the network.

```go
// In gcx: lazy sync.Once on first subcommand that needs the index
loader := &indexLoader{}
searchCmd := searchCommand(loader)         // will trigger load
listProductsCmd := productsCommand(loader) // will trigger load
getCmd := getCommand(fetch)                // never loads index
outlineCmd := outlineCommand(fetch)        // never loads index
linksCmd := linksCommand()                 // never loads index
```

Unlike mcp-grafana (success-only cache; see
[mcp-grafana-integration.md](mcp-grafana-integration.md)), gcx still uses
`sync.Once`. A failed first load is retained for the process lifetime. That
is the current gcx implementation, not a recommendation for new consumers.

`DOCS_INDEX_URL` is an unadvertised override (must be `https`; enforced by
`LoadIndex`).

## Why gcx doesn't use the adapters

The MCP adapter wraps tools for the mcp-go SDK. The CLI adapter is a generic
cobra command. gcx has its own conventions that are more opinionated:

| Feature | CLI adapter | gcx |
|---------|-------------|-----|
| Output | Custom text codecs | `output.Options` with format registry |
| Styling | Plain text tables | Grafana Neon Dark theme + lipgloss |
| Agent mode | Not aware | Auto-detects, changes defaults |
| Error handling | cobra default | Structured `DetailedError` with docs links |

By importing only the core, gcx gets the retrieval logic without any
framework opinions.
