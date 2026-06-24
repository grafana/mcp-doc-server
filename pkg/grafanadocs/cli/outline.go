// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/grafana/hack-doc-server/pkg/grafanadocs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type outlineOpts struct {
	out output
}

func (o *outlineOpts) setup(flags *pflag.FlagSet) {
	o.out.text = outlineTableCodec{}
	o.out.bind(flags)
}

func (o *outlineOpts) Validate(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("url is required")
	}
	return nil
}

// outlineResult mirrors the MCP adapter's JSON keys for cross-surface consistency.
type outlineResult struct {
	URL      string                `json:"url" yaml:"url"`
	Headings []grafanadocs.Heading `json:"headings" yaml:"headings"`
}

func outlineCommand() *cobra.Command {
	opts := &outlineOpts{}
	cmd := &cobra.Command{
		Use:   "outline <url>",
		Short: "Show the heading outline of a documentation page.",
		Long: "List the headings of a documentation page so you can target a " +
			"section with 'gcx docs get --section'.",
		Example: `  gcx docs outline https://grafana.com/docs/tempo/latest/`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawURL := args[0]
			if err := opts.Validate(rawURL); err != nil {
				return err
			}
			doc, err := grafanadocs.FetchDoc(cmd.Context(), rawURL)
			if err != nil {
				return err
			}
			return opts.out.encode(cmd.OutOrStdout(), outlineResult{
				URL:      doc.URL,
				Headings: grafanadocs.Outline(doc),
			})
		},
	}
	opts.setup(cmd.Flags())
	return cmd
}

// outlineTableCodec renders headings as an aligned LVL/HEADING/LINE table.
type outlineTableCodec struct{}

func (outlineTableCodec) Encode(w io.Writer, v any) error {
	res, ok := v.(outlineResult)
	if !ok {
		return fmt.Errorf("outlineTableCodec: expected outlineResult, got %T", v)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	// Writes to tabwriter are buffered; the only meaningful error surfaces at Flush.
	_, _ = fmt.Fprintln(tw, "LVL\tHEADING\tLINE")
	for _, h := range res.Headings {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%d\n", h.Level, h.Text, h.Line)
	}
	return tw.Flush()
}
