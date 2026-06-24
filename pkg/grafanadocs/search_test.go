// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func loadTestIndex(t *testing.T) *Index {
	t.Helper()
	f, err := os.Open("testdata/llms-full-sample.txt")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	idx, err := LoadIndexFromReader(f)
	require.NoError(t, err)
	return idx
}

func TestSearch(t *testing.T) {
	idx := loadTestIndex(t)

	tests := []struct {
		name    string
		query   string
		opts    SearchOpts
		wantMin int // minimum expected results
		checkFn func(t *testing.T, results []Entry)
	}{
		{
			name:    "basic query returns results",
			query:   "clustering",
			wantMin: 1,
		},
		{
			name:    "default limit is 5",
			query:   "grafana agent",
			wantMin: 1,
			checkFn: func(t *testing.T, results []Entry) {
				require.LessOrEqual(t, len(results), 5)
			},
		},
		{
			name:  "custom limit",
			query: "grafana",
			opts:  SearchOpts{Limit: 3},
			checkFn: func(t *testing.T, results []Entry) {
				require.LessOrEqual(t, len(results), 3)
			},
		},
		{
			name:  "product filter",
			query: "flow",
			opts:  SearchOpts{Product: "Grafana Agent"},
			checkFn: func(t *testing.T, results []Entry) {
				for _, e := range results {
					require.Equal(t, "Grafana Agent", e.Product)
				}
			},
		},
		{
			name:    "empty query returns nil",
			query:   "",
			wantMin: 0,
			checkFn: func(t *testing.T, results []Entry) {
				require.Nil(t, results)
			},
		},
		{
			name:    "title matches rank higher",
			query:   "clustering",
			wantMin: 1,
			checkFn: func(t *testing.T, results []Entry) {
				require.Contains(t, results[0].Title, "lustering")
			},
		},
		{
			name:    "word boundary — no substring false positives",
			query:   "file",
			wantMin: 1,
			checkFn: func(t *testing.T, results []Entry) {
				// "file" should match entries with "file" as a word,
				// not entries where it's part of "profile" or "files".
				for _, e := range results {
					words := wordSet(e.Title + " " + e.Description)
					require.True(t, words["file"],
						"result %q should contain 'file' as a whole word", e.Title)
				}
			},
		},
		{
			name:    "multi-token all-match bonus",
			query:   "grafana agent flow",
			wantMin: 1,
			checkFn: func(t *testing.T, results []Entry) {
				// Entries matching all three tokens should rank highest.
				words := wordSet(results[0].Title + " " + results[0].Description)
				require.True(t, words["grafana"] && words["agent"] && words["flow"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Search(idx, tt.query, tt.opts)
			if tt.wantMin > 0 {
				require.GreaterOrEqual(t, len(results), tt.wantMin)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, results)
			}
		})
	}
}

func TestWordSet(t *testing.T) {
	ws := wordSet("Grafana Agent Flow")
	require.True(t, ws["grafana"])
	require.True(t, ws["agent"])
	require.True(t, ws["flow"])
	require.False(t, ws["graf"]) // no substrings
}

func TestBuildIDF(t *testing.T) {
	idx := loadTestIndex(t)
	require.NotNil(t, idx.idf)
	require.Greater(t, len(idx.idf), 0)

	// "grafana" appears in nearly every entry, so its IDF should be low.
	// "clustering" appears rarely, so its IDF should be higher.
	require.Greater(t, idx.idf["clustering"], idx.idf["grafana"])
}

func TestBuildIDF_EmptyIndex(t *testing.T) {
	idf := buildIDF(nil)
	require.NotNil(t, idf)
	require.Empty(t, idf)
}

func TestSearch_ProductResolution(t *testing.T) {
	idx := loadTestIndex(t)

	t.Run("exact name matches case-insensitively", func(t *testing.T) {
		results := Search(idx, "clustering", SearchOpts{Product: "grafana agent"})
		require.NotEmpty(t, results)
		for _, e := range results {
			require.Equal(t, "Grafana Agent", e.Product)
		}
	})

	t.Run("prefix resolves to the product", func(t *testing.T) {
		results := Search(idx, "clustering", SearchOpts{Product: "grafana"})
		require.NotEmpty(t, results)
		for _, e := range results {
			require.Equal(t, "Grafana Agent", e.Product)
		}
	})

	t.Run("substring resolves to the product", func(t *testing.T) {
		results := Search(idx, "clustering", SearchOpts{Product: "agent"})
		require.NotEmpty(t, results)
		for _, e := range results {
			require.Equal(t, "Grafana Agent", e.Product)
		}
	})

	t.Run("unknown product returns nothing", func(t *testing.T) {
		results := Search(idx, "clustering", SearchOpts{Product: "nonexistent"})
		require.Empty(t, results)
	})
}

func TestSearch_TitleMatchBeatsDescriptionOnly(t *testing.T) {
	idx := loadTestIndex(t)

	results := Search(idx, "clustering", SearchOpts{Limit: 10})
	require.NotEmpty(t, results)

	// The top result should have "clustering" in its title.
	topWords := wordSet(results[0].Title)
	require.True(t, topWords["clustering"],
		"top result %q should have 'clustering' in title", results[0].Title)
}

func TestSearch_ResultsHaveRequiredFields(t *testing.T) {
	idx := loadTestIndex(t)

	results := Search(idx, "grafana", SearchOpts{Limit: 20})
	for _, e := range results {
		require.NotEmpty(t, e.Title)
		require.NotEmpty(t, e.URL)
		require.NotEmpty(t, e.Product)
		require.True(t, strings.HasPrefix(e.URL, "https://grafana.com/"),
			"result URL %q not from grafana.com", e.URL)
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"configure tempo", []string{"configure", "tempo"}},
		{"PromQL query", []string{"promql", "query"}},
		{"a b", nil}, // single chars dropped
		{"k6-browser", []string{"k6", "browser"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, tokenize(tt.input))
		})
	}
}
