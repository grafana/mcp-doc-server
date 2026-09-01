---
title: Integrate the core library
menuTitle: Integrate
description: Import pkg/grafanadocs to search and fetch Grafana documentation from Go, without MCP or CLI dependencies.
weight: 8
topicType: reference
versionDate: 2026-09-01
---

# Integrate the core library

If you want the MCP server as-is, refer to [Install and connect](../install/). This page is for embedding `pkg/grafanadocs` in your own Go project.

The core uses the Go standard library only. It doesn't depend on MCP or Cobra. Timeouts, rate limits, and body-size caps are the same values documented in [Configure the server](../configure/).

[mcp-grafana](https://github.com/grafana/mcp-grafana) and [gcx](https://github.com/grafana/gcx) import this package and write their own wrappers. They don't import the MCP or CLI adapters.

## Before you begin

- A Go module (Go 1.26+)
- Network access to `grafana.com`

## Add the dependency

```bash
go get github.com/grafana/mcp-doc-server/pkg/grafanadocs
```

## Complete example

This program loads the index, searches once, fetches the top result, and prints the first 20 lines:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/grafana/mcp-doc-server/pkg/grafanadocs"
)

func main() {
	ctx := context.Background()

	idx, err := grafanadocs.LoadIndex(ctx, grafanadocs.DefaultIndexURL)
	if err != nil {
		log.Fatal(err)
	}

	results := grafanadocs.Search(idx, "traceql query", grafanadocs.SearchOpts{Limit: 1})
	if len(results) == 0 {
		log.Fatal("no matching pages")
	}

	doc, err := grafanadocs.FetchDoc(ctx, results[0].URL)
	if err != nil {
		log.Fatal(err)
	}

	excerpt := grafanadocs.Excerpt(doc, grafanadocs.ExcerptOpts{Limit: 20})
	fmt.Printf("%s (lines %d-%d of %d)\n\n", results[0].URL, excerpt.Start, excerpt.End, excerpt.Total)
	fmt.Println(excerpt.Content)
}
```

The rest of this page unpacks each call.

## Load the index

Load the catalog once and reuse it:

```go
import "github.com/grafana/mcp-doc-server/pkg/grafanadocs"

idx, err := grafanadocs.LoadIndex(ctx, grafanadocs.DefaultIndexURL)
if err != nil {
    return err
}
```

For tests or fixtures, load from a reader:

```go
idx, err := grafanadocs.LoadIndexFromReader(reader)
```

You own the lifecycle. Lazy `sync.Once` on first use is a common pattern.

## Search the index

```go
results := grafanadocs.Search(idx, "traceql query", grafanadocs.SearchOpts{
    Product: "tempo",
    Limit:   5,
})

for _, entry := range results {
    fmt.Printf("%s: %s\n", entry.Title, entry.URL)
}
```

`SearchOpts.Product` filters to one product (empty means all). `Limit` caps results (`0` defaults to 5). `Search` makes no network calls and never returns an error. An empty slice means no matches.

## Fetch a page

```go
doc, err := grafanadocs.FetchDoc(ctx, "https://grafana.com/docs/loki/latest/configure/")
if err != nil {
    return err
}
```

`FetchDoc` returns cleaned Markdown. It enforces the `grafana.com/docs/` allowlist and rate limiting. The URL works with or without a trailing `.md`, so you can pass a `Search` result straight through.

## Get the heading outline

```go
headings := grafanadocs.Outline(doc)
for _, h := range headings {
    fmt.Printf("%s%s (line %d)\n", strings.Repeat("  ", h.Level-1), h.Text, h.Line)
}
```

## Extract a section or page slice

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

`Excerpt` never errors. `Limit` `0` defaults to 80 lines: a short page comes back in full, a long one comes back as a slice unless you raise the limit. If a named `Section` isn't found, `result.Content` is empty.

## List products

```go
products := idx.Products()
for _, p := range products {
    fmt.Printf("%s (%d pages)\n", p.Name, p.Count)
}
```

## Mount the MCP adapter

```go
import mcpadapter "github.com/grafana/mcp-doc-server/pkg/grafanadocs/mcp"

mcpadapter.New(idx).Register(srv)
```

That registers all four tools.

## Mount the CLI adapter

```go
import "github.com/grafana/mcp-doc-server/pkg/grafanadocs/cli"

rootCmd.AddCommand(cli.Command(idx))
```

That adds `docs search`, `docs get`, `docs outline`, and `docs products`.

## Error handling

| Function | Returns errors? | Common causes | Safe to retry? |
|----------|-----------------|---------------|------------|
| `LoadIndex` | Yes | Non-HTTPS URL, network failure, non-200 status | Network and 5xx: yes. Bad URL or scheme: no. |
| `FetchDoc` | Yes | Allowlist rejection (wrong scheme, host, or path), network failure, non-200 status, blocked redirect | Network and 5xx: yes. Allowlist rejection: no. |
| `Search` | No | None (empty slice on no match) | N/A |
| `Outline` | No | None | N/A |
| `Excerpt` | No | None (empty `Content` when a section isn't found) | N/A |
| `LoadIndexFromReader` | Yes | Reader I/O errors, including lines longer than the 1 MiB scanner buffer | No |

Errors use a `grafanadocs:` prefix. There are no sentinel types, so match on the prefix with string inspection. `errors.Is` works only when the error wraps a standard library error, for example `context.Canceled`.

## Concurrency

After construction, `*Index` is safe for concurrent reads. `Search` and `Products` can run from many goroutines. `FetchDoc`'s rate limiter is process-global: every caller shares the five-concurrent / 200 ms-gap budget.

## Index lifecycle patterns

| Pattern | When to use |
|---------|-------------|
| Load at startup | Standalone servers that always need the index |
| Lazy `sync.Once` | CLI tools or services where some commands don't need it |
| Load from reader | Testing with fixture data |

`get` and `outline` only call `FetchDoc`. They don't need the index.

## Related resources

- [Configure the server](../configure/)
- [Tools and CLI reference](../tools/)
