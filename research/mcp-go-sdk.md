# Tooling: official MCP Go SDK

Notes on the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
(`github.com/modelcontextprotocol/go-sdk`) — the most likely answer to the open "MCP SDK
choice and pinned version" question in `SPECS.md`.

## What it is

The official Go SDK for the Model Context Protocol, maintained in collaboration with
Google. It aims to implement the full MCP spec. Latest release: `v1.6.1` (May 2026).

### Packages

- `.../mcp` — primary APIs for building MCP clients and servers.
- `.../jsonrpc` — for implementing custom transports.
- `.../auth` — OAuth primitives.
- `.../oauthex` — OAuth protocol extensions (e.g. ProtectedResourceMetadata).

### Spec/version compatibility

- `v1.4.0+` supports MCP spec `2025-11-25` (latest) plus `2025-06-18`, `2025-03-26`,
  `2024-11-05`.
- New SDK releases target only supported Go versions
  (see <https://go.dev/doc/devel/release#policy>).
- License: Apache 2.0 for new contributions (existing code MIT).

## Minimal server (stdio) — the shape mcp-doc-server takes

```go
package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type Output struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, Output, error) {
	return nil, Output{Greeting: "Hi " + input.Name}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
```

Key takeaways for our design:

- `mcp.NewServer` + `mcp.AddTool` + `server.Run(ctx, &mcp.StdioTransport{})` matches our
  v1 stdio decision in `SPECS.md`.
- Tools are plain typed Go funcs `(ctx, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)`.
  Input/output structs use `json` + `jsonschema` struct tags, so the SDK generates the tool
  schemas for us — this directly informs the "exact tool schemas" open question for
  `search_docs`, `get_doc`, `get_doc_outline`, `list_products`.
- A built-in `mcp.Client` + `CommandTransport` makes it straightforward to write Go
  integration tests that spawn our binary and call tools (pairs with `TESTS.md` and the
  MCP Inspector CLI for verification).

## Why this is the likely SDK choice

- Official and actively released (`v1.6.1`, 25 releases), tracks the latest spec.
- Pure Go, stdio transport out of the box — no extra runtime, fits a single static binary.
- Schema-from-struct-tags keeps tool contracts close to the Go types, which supports the
  SDD goal of specs and code moving together.
- Alternatives exist (`mcp-go` by Ed Zynda, `mcp-golang`, `go-mcp`) and remain viable, but
  the official SDK is the safest default for a new project.

## Open items this informs (from SPECS.md)

- **MCP SDK choice + pinned version** — propose `github.com/modelcontextprotocol/go-sdk`
  pinned at a current `v1.x` (e.g. `v1.6.1`), targeting MCP spec `2025-06-18`/`2025-11-25`.
- **Go version** — must be a currently supported Go release per the SDK's policy.
- **Tool schemas** — define via typed input/output structs with `jsonschema` tags.

## Source

- <https://github.com/modelcontextprotocol/go-sdk>
