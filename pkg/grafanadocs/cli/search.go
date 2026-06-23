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

type searchOpts struct {
	out     output
	product string
	limit   int
}

func (o *searchOpts) setup(flags *pflag.FlagSet) {
	o.out.text = searchTableCodec{}
	o.out.bind(flags)
	flags.StringVar(&o.product, "product", "", "Filter results to a specific product")
	flags.IntVar(&o.limit, "limit", 5, "Maximum number of results")
}

func (o *searchOpts) Validate(query string) error {
	if strings.TrimSpace(query) == "" {
		return errors.New("query is required")
	}
	return nil
}

func searchCommand(idx *grafanadocs.Index) *cobra.Command {
	opts := &searchOpts{}
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Grafana documentation.",
		Long:  "Search the documentation index by keyword. Returns matching pages ranked by relevance.",
		Example: `  # Search across all products
  gcx docs search "rate limiting"

  # Scope the search to one product
  gcx docs search "metrics generator" --product tempo

  # Return more results as JSON
  gcx docs search dashboards --limit 10 -o json`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			if err := opts.Validate(query); err != nil {
				return err
			}
			results := grafanadocs.Search(idx, query, grafanadocs.SearchOpts{
				Product: opts.product,
				Limit:   opts.limit,
			})
			// Guidance goes to stderr so stdout stays a clean, parseable result
			// set (invariant I13: actionable empty results).
			if len(results) == 0 {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), emptySearchHint(opts.product))
			}
			return opts.out.encode(cmd.OutOrStdout(), results)
		},
	}
	opts.setup(cmd.Flags())
	return cmd
}

func emptySearchHint(product string) string {
	if product != "" {
		return "No results found. Try broadening the product filter or run 'gcx docs products' to see available products."
	}
	return "No results found. Try different search terms."
}

// searchTableCodec renders search hits as an aligned TITLE/PRODUCT/URL table.
type searchTableCodec struct{}

func (searchTableCodec) Encode(w io.Writer, v any) error {
	entries, ok := v.([]grafanadocs.Entry)
	if !ok {
		return fmt.Errorf("searchTableCodec: expected []grafanadocs.Entry, got %T", v)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	// Writes to tabwriter are buffered; the only meaningful error surfaces at Flush.
	_, _ = fmt.Fprintln(tw, "TITLE\tPRODUCT\tURL")
	for _, e := range entries {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Title, e.Product, e.URL)
	}
	return tw.Flush()
}
