// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

// Command docs is a standalone CLI front-end for the grafanadocs cobra adapter.
// It runs the same `docs` command group (search, get, outline, products) that
// gcx would mount, so the adapter is exercisable on its own.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/grafana/hack-doc-server/pkg/grafanadocs"
	"github.com/grafana/hack-doc-server/pkg/grafanadocs/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Skip the network fetch for help/completion so they work offline.
	idx := &grafanadocs.Index{}
	if cli.NeedsIndex(os.Args[1:]) {
		loaded, err := grafanadocs.LoadIndex(ctx, indexURL())
		if err != nil {
			fmt.Fprintf(os.Stderr, "docs: failed to load index: %v\n", err)
			os.Exit(1)
		}
		idx = loaded
	}

	root := cli.Command(idx) // Use is already "docs"; serves as our root command.
	root.SilenceUsage = true
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1) // cobra already printed the error.
	}
}

// indexURL returns the index location, allowing a DOCS_INDEX_URL override.
func indexURL() string {
	if u := os.Getenv("DOCS_INDEX_URL"); u != "" {
		return u
	}
	return grafanadocs.DefaultIndexURL
}
