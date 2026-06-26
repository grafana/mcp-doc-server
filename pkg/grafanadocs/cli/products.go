// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/grafana/mcp-doc-server/pkg/grafanadocs"
	"github.com/spf13/cobra"
)

// productsResult mirrors the MCP adapter's JSON keys for cross-surface consistency.
type productsResult struct {
	Products []cliProduct `json:"products" yaml:"products"`
}

func productsCommand(idx *grafanadocs.Index) *cobra.Command {
	out := output{text: productsTableCodec{}}
	cmd := &cobra.Command{
		Use:     "products",
		Short:   "List Grafana documentation products.",
		Long:    "List all product documentation groups in the index with their entry counts.",
		Example: `  gcx docs products`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			raw := idx.Products()
			products := make([]cliProduct, len(raw))
			for i, p := range raw {
				products[i] = cliProduct{Name: p.Name, Count: p.Count}
			}
			return out.encode(cmd.OutOrStdout(), productsResult{Products: products})
		},
	}
	out.bind(cmd.Flags())
	return cmd
}

// productsTableCodec renders products as an aligned PRODUCT/COUNT table.
type productsTableCodec struct{}

func (productsTableCodec) Encode(w io.Writer, v any) error {
	res, ok := v.(productsResult)
	if !ok {
		return fmt.Errorf("productsTableCodec: expected productsResult, got %T", v)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PRODUCT\tCOUNT")
	for _, p := range res.Products {
		_, _ = fmt.Fprintf(tw, "%s\t%d\n", p.Name, p.Count)
	}
	return tw.Flush()
}
