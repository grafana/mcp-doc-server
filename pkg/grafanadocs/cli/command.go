// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

// Package cli provides a cobra adapter for grafanadocs. It exposes the core
// retrieval functions as a mountable "docs" command group (search, get,
// outline, products), following gcx's command conventions.
//
// The adapter depends only on cobra/pflag and the grafanadocs core — it does
// NOT import gcx internals (which are not importable from external modules).
// Output is rendered by a small local helper (output) that mirrors the shape
// of gcx's output.Options (package internal/output) closely enough that a
// gcx-native port is mechanical: swap output for output.Options, and have each
// text codec implement gcx's format.Codec (the Encode method carries over as-is;
// add a trivial Format() and an unsupported Decode()).
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

	"github.com/grafana/hack-doc-server/pkg/grafanadocs"
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
