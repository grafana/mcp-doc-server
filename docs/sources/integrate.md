---
title: Integrate the core library
menuTitle: Integrate
description: Import the pkg/grafanadocs core library into a Go project to add Grafana documentation search and retrieval without MCP or CLI dependencies.
weight: 6
topicType: task
versionDate: 2026-06-25
---

# Integrate the core library

This page is for Go developers embedding the core library in their own projects. To run the server as-is, refer to [Install and configure](../install/).

Import `pkg/grafanadocs` into your Go project to add Grafana documentation retrieval without depending on Model Context Protocol (MCP) or CLI frameworks. The core is dependency-light — just the Go standard library.

## Architecture

The core package (`pkg/grafanadocs`) has no dependencies on MCP or CLI frameworks. The MCP server and CLI adapters live in subpackages (`pkg/grafanadocs/mcp` and `pkg/grafanadocs/cli`) that import the core and wire it to `mcp-go` and Cobra respectively.

Two other projects import the core library directly:

- [mcp-grafana](https://github.com/grafana/mcp-grafana) — adds doc tools to its MCP server
- [gcx](https://github.com/grafana/gcx) — exposes `gcx docs` commands

Both consumers import `pkg/grafanadocs` directly and write their own idiomatic wrappers — they don't import the MCP or CLI adapters.

For the timeout, rate-limit, and body-size values these packages enforce, refer to [Configuration](../configure/).

## Before you begin

You need:

- A Go module (Go 1.25+)
- Access to the `github.com/grafana` GitHub organization (the repository is internal)
- Network access to `grafana.com`

## Add the dependency

The module is in an internal repository. Set `GOPRIVATE` before fetching:

```bash
go env -w GOPRIVATE=github.com/grafana/*
go get github.com/grafana/mcp-doc-server/pkg/grafanadocs
```

## Load the index

Load the catalog once and reuse it for all searches:

```go
import "github.com/grafana/mcp-doc-server/pkg/grafanadocs"

idx, err := grafanadocs.LoadIndex(ctx, grafanadocs.DefaultIndexURL)
if err != nil {
    return err
}
```

For testing or custom sources, load from a reader:

```go
idx, err := grafanadocs.LoadIndexFromReader(reader)
```

The caller owns the index lifecycle. A common pattern is lazy initialization with `sync.Once` on first use.

## Search the index

```go
results := grafanadocs.Search(idx, "alerting rules", grafanadocs.SearchOpts{
    Product: "Grafana",
    Limit:   5,
})

for _, entry := range results {
    fmt.Printf("%s — %s\n", entry.Title, entry.URL)
}
```

`SearchOpts.Product` filters to a specific product (empty means all). `Limit` caps results (0 defaults to 5). `Search` makes no network calls and never returns an error—an empty slice means no matches.

## Fetch a page

```go
doc, err := grafanadocs.FetchDoc(ctx, "https://grafana.com/docs/loki/latest/configure/")
if err != nil {
    return err
}
```

`FetchDoc` returns cleaned Markdown. It enforces the URL allowlist (only `grafana.com/docs/`) and rate limiting automatically.

## Get the heading outline

```go
headings := grafanadocs.Outline(doc)
for _, h := range headings {
    fmt.Printf("%s%s (line %d)\n", strings.Repeat("  ", h.Level-1), h.Text, h.Line)
}
```

## Extract a section or page slice

Use `Excerpt` for bounded retrieval:

```go
// Extract by section heading
result := grafanadocs.Excerpt(doc, grafanadocs.ExcerptOpts{
    Section: "Limits",
})

// Or page by line range
result := grafanadocs.Excerpt(doc, grafanadocs.ExcerptOpts{
    Offset: 100,
    Limit:  50,
})

fmt.Printf("Lines %d-%d of %d\n", result.Start, result.End, result.Total)
fmt.Println(result.Content)
```

`Excerpt` never errors. When `Limit` is 0 it defaults to 80 lines, so it won't return the whole page unless you raise the limit. If a named `Section` isn't found, `result.Content` is empty—check for that.

## List products

```go
products := idx.Products()
for _, p := range products {
    fmt.Printf("%s (%d pages)\n", p.Name, p.Count)
}
```

## Mount the MCP adapter

To build an MCP server, register all four tools:

```go
import mcpadapter "github.com/grafana/mcp-doc-server/pkg/grafanadocs/mcp"

mcpadapter.New(idx).Register(srv)
```

## Mount the CLI adapter

To build a cobra-based CLI, mount the docs command group:

```go
import "github.com/grafana/mcp-doc-server/pkg/grafanadocs/cli"

rootCmd.AddCommand(cli.Command(idx))
```

This adds `docs search`, `docs get`, `docs outline`, and `docs products`.

## Error handling

| Function | Returns errors? | Common causes | Safe to retry? |
|----------|-----------------|---------------|------------|
| `LoadIndex` | Yes | Non-HTTPS URL, network failure, non-200 status, oversized index | Network and 5xx: yes. Bad URL or scheme: no. |
| `FetchDoc` | Yes | Allowlist rejection (wrong scheme, host, or path), network failure, non-200 status, blocked redirect | Network and 5xx: yes. Allowlist rejection: no. |
| `Search` | No | — (returns empty slice on no match) | — |
| `Outline` | No | — | — |
| `Excerpt` | No | — (empty `Content` when a section isn't found) | — |
| `LoadIndexFromReader` | Yes | Malformed index stream | No |

All errors are wrapped with a `grafanadocs:` prefix, so you can match on them with `errors.Is`/`errors.As` or string inspection.

## Concurrency

After construction, `*Index` is safe for concurrent reads—`Search` and `Products` can run from many goroutines. `FetchDoc`'s rate limiter is process-global, so all callers in the process share the same five-concurrent / 200ms-gap budget.

## Index lifecycle patterns

| Pattern | When to use |
|---------|-------------|
| Load at startup | Standalone servers that always need the index |
| Lazy `sync.Once` | CLI tools or services where some commands don't need it |
| Load from reader | Testing with fixture data |

Commands that only call `FetchDoc` (like `get` and `outline`) never need the index loaded.

## Related resources

- [Configuration](../configure/)—the timeout and rate limit values these functions enforce
- [Tools and CLI reference](../tools/)—the input/output contracts exposed by the adapters
