// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/hack-doc-server/pkg/grafanadocs"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func testIndex(t *testing.T) *grafanadocs.Index {
	t.Helper()
	input := `## Grafana Tempo documentation
- [Configure Tempo](https://grafana.com/docs/tempo/latest/configuration.md): Configure all aspects of Tempo
- [Tempo clustering](https://grafana.com/docs/tempo/latest/operations/clustering.md): Learn about Tempo clustering
- [TraceQL](https://grafana.com/docs/tempo/latest/traceql.md): Query traces using TraceQL syntax
## Grafana Loki documentation
- [LogQL](https://grafana.com/docs/loki/latest/logql.md): Query logs using LogQL
- [Loki configuration](https://grafana.com/docs/loki/latest/configuration.md): Configure Loki
`
	idx, err := grafanadocs.LoadIndexFromReader(strings.NewReader(input))
	require.NoError(t, err)
	return idx
}

func callTool(t *testing.T, s *Server, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := handler(context.Background(), req)
	require.NoError(t, err)
	return result
}

func TestHandleSearchDocs(t *testing.T) {
	s := New(testIndex(t))

	t.Run("basic search returns results", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query": "clustering",
		})
		require.False(t, result.IsError)

		var entries []grafanadocs.Entry
		require.NoError(t, json.Unmarshal([]byte(textContent(t, result)), &entries))
		require.Greater(t, len(entries), 0)
		require.Contains(t, entries[0].Title, "lustering")
	})

	t.Run("missing query returns error", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{})
		require.True(t, result.IsError)
	})

	t.Run("empty query string returns error", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query": "",
		})
		require.True(t, result.IsError)
	})

	t.Run("product filter works", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query":   "configuration",
			"product": "Loki",
		})
		require.False(t, result.IsError)

		var entries []grafanadocs.Entry
		require.NoError(t, json.Unmarshal([]byte(textContent(t, result)), &entries))
		for _, e := range entries {
			require.Contains(t, e.Product, "Loki")
		}
	})

	t.Run("no results returns guidance", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query":   "xyznonexistent",
			"product": "Tempo",
		})
		require.False(t, result.IsError)
		text := textContent(t, result)
		require.Contains(t, text, "No results found")
		require.Contains(t, text, "list_products")
	})

	t.Run("limit is respected", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query": "configuration",
			"limit": float64(1),
		})
		require.False(t, result.IsError)

		var entries []grafanadocs.Entry
		require.NoError(t, json.Unmarshal([]byte(textContent(t, result)), &entries))
		require.Equal(t, 1, len(entries))
	})
}

func TestHandleGetDoc(t *testing.T) {
	s := New(testIndex(t))

	t.Run("missing url returns error", func(t *testing.T) {
		result := callTool(t, s, s.handleGetDoc, map[string]any{})
		require.True(t, result.IsError)
	})

	t.Run("empty url returns error", func(t *testing.T) {
		result := callTool(t, s, s.handleGetDoc, map[string]any{
			"url": "",
		})
		require.True(t, result.IsError)
	})

	t.Run("non-grafana url rejected", func(t *testing.T) {
		result := callTool(t, s, s.handleGetDoc, map[string]any{
			"url": "https://evil.com/docs/page.md",
		})
		require.True(t, result.IsError)
		require.Contains(t, textContent(t, result), "rejected")
	})
}

func TestHandleGetDocOutline(t *testing.T) {
	s := New(testIndex(t))

	t.Run("missing url returns error", func(t *testing.T) {
		result := callTool(t, s, s.handleGetDocOutline, map[string]any{})
		require.True(t, result.IsError)
	})

	t.Run("non-grafana url rejected", func(t *testing.T) {
		result := callTool(t, s, s.handleGetDocOutline, map[string]any{
			"url": "http://169.254.169.254/latest",
		})
		require.True(t, result.IsError)
	})
}

func TestHandleListProducts(t *testing.T) {
	s := New(testIndex(t))

	result := callTool(t, s, s.handleListProducts, map[string]any{})
	require.False(t, result.IsError)

	var resp struct {
		Products []grafanadocs.Product `json:"products"`
	}
	require.NoError(t, json.Unmarshal([]byte(textContent(t, result)), &resp))
	require.Equal(t, 2, len(resp.Products))

	names := make([]string, len(resp.Products))
	for i, p := range resp.Products {
		names[i] = p.Name
		require.Greater(t, p.Count, 0)
	}
	require.Contains(t, names, "Grafana Tempo")
	require.Contains(t, names, "Grafana Loki")
}

// textContent extracts the text from the first content block.
func textContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result)
	require.Greater(t, len(result.Content), 0)
	tc, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", result.Content[0])
	return tc.Text
}
