// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package mcp

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/grafana/mcp-doc-server/pkg/grafanadocs"
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

		var entries []searchEntry
		require.NoError(t, json.Unmarshal([]byte(textContent(t, result)), &entries))
		require.Greater(t, len(entries), 0)
		require.Contains(t, entries[0].Title, "lustering")
	})

	t.Run("json keys are snake_case", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query": "clustering",
		})
		text := textContent(t, result)
		require.Contains(t, text, `"title"`)
		require.Contains(t, text, `"url"`)
		require.Contains(t, text, `"description"`)
		require.Contains(t, text, `"product"`)
		require.NotContains(t, text, `"Title"`)
		require.NotContains(t, text, `"URL"`)
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
			"product": "Grafana Loki",
		})
		require.False(t, result.IsError)

		var entries []searchEntry
		require.NoError(t, json.Unmarshal([]byte(textContent(t, result)), &entries))
		for _, e := range entries {
			require.Equal(t, "Grafana Loki", e.Product)
		}
	})

	t.Run("no results returns guidance", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query":   "xyznonexistent",
			"product": "Grafana Tempo",
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

		var entries []searchEntry
		require.NoError(t, json.Unmarshal([]byte(textContent(t, result)), &entries))
		require.Equal(t, 1, len(entries))
	})

	t.Run("negative limit rejected", func(t *testing.T) {
		result := callTool(t, s, s.handleSearchDocs, map[string]any{
			"query": "clustering",
			"limit": float64(-1),
		})
		require.True(t, result.IsError)
		require.Contains(t, textContent(t, result), "negative")
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

	t.Run("negative offset rejected", func(t *testing.T) {
		result := callTool(t, s, s.handleGetDoc, map[string]any{
			"url":    "https://grafana.com/docs/tempo/latest/configuration.md",
			"offset": float64(-5),
		})
		require.True(t, result.IsError)
		require.Contains(t, textContent(t, result), "negative")
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
		Products []productEntry `json:"products"`
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

	t.Run("json keys are snake_case", func(t *testing.T) {
		text := textContent(t, result)
		require.Contains(t, text, `"name"`)
		require.Contains(t, text, `"count"`)
		require.NotContains(t, text, `"Name"`)
		require.NotContains(t, text, `"Count"`)
	})
}

func TestSafeInt(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		want    int
		wantErr bool
	}{
		{"positive", 5.0, 5, false},
		{"zero", 0.0, 0, false},
		{"truncates decimal", 5.9, 5, false},
		{"negative", -1.0, 0, true},
		{"large", 1000.0, 1000, false},
		{"NaN", math.NaN(), 0, true},
		{"+Inf", math.Inf(1), 0, true},
		{"-Inf", math.Inf(-1), 0, true},
		{"overflow does not wrap to negative", 1e300, 0, true},
		{"above cap", float64(maxSafeInt) + 1, 0, true},
		{"at cap", float64(maxSafeInt), maxSafeInt, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeInt(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
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
