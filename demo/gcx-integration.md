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
    └── products.go  → calls idx.Products()
```

gcx imports **only the core** (`pkg/grafanadocs`) — not the MCP or CLI adapter.
It writes its own command layer using gcx conventions: `output.Options`, styled
tables, agent annotations, and the standard opts pattern.

## Commands

```bash
# Search across all Grafana docs
gcx docs search "alerting rules"

# Filter to a product
gcx docs search "configuration" --product tempo

# Fetch a page (text mode shows raw markdown)
gcx docs get https://grafana.com/docs/loki/latest/query/

# Extract just one section
gcx docs get https://grafana.com/docs/loki/latest/query/ --section "Log queries"

# Page through a long document
gcx docs get https://grafana.com/docs/loki/latest/query/ --offset 80 --limit 80

# Show the heading structure
gcx docs outline https://grafana.com/docs/tempo/latest/

# List all documented products
gcx docs products
```

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

The index is loaded lazily — only `search` and `products` need it.
`get` and `outline` only call `FetchDoc`, so they work without loading
the full index (faster startup, works even if the index is temporarily
unreachable).

```go
// In gcx: lazy sync.Once on first subcommand that needs the index
loader := &indexLoader{}
searchCmd := searchCommand(loader)   // will trigger load
productsCmd := productsCommand(loader) // will trigger load
getCmd := getCommand()                // never loads index
outlineCmd := outlineCommand()        // never loads index
```

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
