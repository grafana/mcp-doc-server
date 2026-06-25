// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

// Package cli provides a cobra adapter for grafanadocs. It exposes the core
// as a mountable "docs" command group (search, get, outline, products).
//
// Depends only on cobra/pflag and the grafanadocs core — no gcx internals.
// A gcx-native port swaps the local output helper for output.Options.
//
// Agent annotations (applied by gcx from its own registry, not here):
//
//	gcx docs search    cost=small   hint="search Grafana docs by keyword"
//	gcx docs get       cost=medium  hint="fetch a doc page (bounded markdown)"
//	gcx docs outline   cost=small   hint="list headings of a doc page"
//	gcx docs products  cost=small   hint="list available doc products"
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/grafana/mcp-doc-server/pkg/grafanadocs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Command returns the "docs" command group wired to the given index. Mount it
// on a parent command (e.g. gcx's root) with AddCommand. The caller owns the
// index lifecycle; the adapter is stateless.
func Command(idx *grafanadocs.Index) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Search and read Grafana documentation.",
		Long: "Search, fetch, and outline Grafana Labs product documentation, " +
			"backed by an in-memory index of grafana.com docs.",
	}
	cmd.AddCommand(
		searchCommand(idx),
		getCommand(),
		outlineCommand(),
		productsCommand(idx),
	)
	return cmd
}

// indexReadingCommands are the docs subcommands that query the in-memory index.
// `get` and `outline` only call FetchDoc, so they never need it. Keep this in
// sync with the subcommands wired in Command.
var indexReadingCommands = map[string]bool{
	"search":   true,
	"products": true,
}

// NeedsIndex reports whether a docs invocation requires a loaded index, given
// the args after the program name. Only the commands in indexReadingCommands
// read the index; help, completion, and bare invocations need nothing. Leading
// flags (e.g. -v) are skipped to find the subcommand token.
func NeedsIndex(args []string) bool {
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		return indexReadingCommands[a]
	}
	return false
}

// textCodec renders a value as human-readable text for the default format.
type textCodec interface {
	Encode(w io.Writer, v any) error
}

// output controls how command results are rendered. It is bound to the
// -o/--output flag and dispatches to the registered text codec or a built-in
// structured encoder.
type output struct {
	format string
	text   textCodec
}

func (o *output) bind(flags *pflag.FlagSet) {
	flags.StringVarP(&o.format, "output", "o", "text",
		"Output format: text, json, yaml, agents")
}

// encode writes v in the selected format. text uses the registered codec;
// json is indented, agents is compact JSON, yaml is YAML.
func (o *output) encode(w io.Writer, v any) error {
	switch o.format {
	case "", "text":
		return o.text.Encode(w, v)
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case "agents":
		return json.NewEncoder(w).Encode(v)
	case "yaml":
		enc := yaml.NewEncoder(w)
		if err := enc.Encode(v); err != nil {
			return err
		}
		return enc.Close()
	default:
		return fmt.Errorf("unknown output format %q (use text, json, yaml, or agents)", o.format)
	}
}
