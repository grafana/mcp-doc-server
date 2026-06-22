// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/grafana/mcp-doc-server/pkg/grafanadocs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type getOpts struct {
	out     output
	section string
	offset  int
	limit   int
	rawURL  string
}

func (o *getOpts) setup(flags *pflag.FlagSet) {
	o.out.text = getTextCodec{}
	o.out.bind(flags)
	flags.StringVar(&o.section, "section", "", "Heading text to extract (returns only that section)")
	flags.IntVar(&o.offset, "offset", 0, "Line offset for paging (0-indexed)")
	flags.IntVar(&o.limit, "limit", 0, "Maximum lines to return (0 = default)")
}

func (o *getOpts) Validate() error {
	if strings.TrimSpace(o.rawURL) == "" {
		return errors.New("url is required")
	}
	return nil
}

// getResult mirrors the MCP adapter's JSON keys for cross-surface consistency.
type getResult struct {
	Content       string `json:"content" yaml:"content"`
	URL           string `json:"url" yaml:"url"`
	TotalLines    int    `json:"total_lines" yaml:"total_lines"`
	ReturnedRange [2]int `json:"returned_range" yaml:"returned_range"`
}

func getCommand() *cobra.Command {
	opts := &getOpts{}
	cmd := &cobra.Command{
		Use:   "get <url>",
		Short: "Fetch a Grafana documentation page.",
		Long: "Fetch a documentation page as cleaned markdown. Supports section " +
			"extraction and offset/limit paging for bounded retrieval.",
		Example: `  # Fetch the first page of a doc
  gcx docs get https://grafana.com/docs/tempo/latest/

  # Extract a single section
  gcx docs get https://grafana.com/docs/tempo/latest/ --section "Configuration"

  # Page through a long doc
  gcx docs get https://grafana.com/docs/tempo/latest/ --offset 80 --limit 80`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.rawURL = args[0]
			if err := opts.Validate(); err != nil {
				return err
			}
			doc, err := grafanadocs.FetchDoc(cmd.Context(), opts.rawURL)
			if err != nil {
				return err
			}
			res := grafanadocs.Excerpt(doc, grafanadocs.ExcerptOpts{
				Section: opts.section,
				Offset:  opts.offset,
				Limit:   opts.limit,
			})
			if res.Content == "" && opts.section != "" {
				return fmt.Errorf("section %q not found; run 'gcx docs outline %s' to see available headings", opts.section, opts.rawURL)
			}
			return opts.out.encode(cmd.OutOrStdout(), getResult{
				Content:       res.Content,
				URL:           doc.URL,
				TotalLines:    res.Total,
				ReturnedRange: [2]int{res.Start, res.End},
			})
		},
	}
	opts.setup(cmd.Flags())
	return cmd
}

// getTextCodec renders the raw markdown content of a fetched page.
type getTextCodec struct{}

func (getTextCodec) Encode(w io.Writer, v any) error {
	r, ok := v.(getResult)
	if !ok {
		return fmt.Errorf("getTextCodec: expected getResult, got %T", v)
	}
	_, err := fmt.Fprintln(w, r.Content)
	return err
}
