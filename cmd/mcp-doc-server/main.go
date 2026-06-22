// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/grafana/mcp-doc-server/pkg/grafanadocs"
	mcpadapter "github.com/grafana/mcp-doc-server/pkg/grafanadocs/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const version = "0.1.0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	indexURL := grafanadocs.DefaultIndexURL
	if u := os.Getenv("DOCS_INDEX_URL"); u != "" {
		indexURL = u
	}

	fmt.Fprintf(os.Stderr, "mcp-doc-server %s: loading index from %s\n", version, indexURL)

	idx, err := grafanadocs.LoadIndex(ctx, indexURL)
	if err != nil {
		log.Fatalf("failed to load index: %v", err)
	}
	fmt.Fprintf(os.Stderr, "index loaded: %d entries, %d products\n", idx.EntryCount(), len(idx.Products()))

	srv := mcpadapter.NewMCPServer(idx, version)
	if err := server.ServeStdio(srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
