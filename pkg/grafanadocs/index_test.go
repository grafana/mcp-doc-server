// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadIndexFromReader(t *testing.T) {
	f, err := os.Open("testdata/llms-full-sample.txt")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	idx, err := LoadIndexFromReader(f)
	require.NoError(t, err)

	t.Run("entries parsed", func(t *testing.T) {
		require.Greater(t, idx.EntryCount(), 0)
	})

	t.Run("products extracted", func(t *testing.T) {
		products := idx.Products()
		require.Greater(t, len(products), 0)
		for _, p := range products {
			require.NotEmpty(t, p.Name)
			require.Greater(t, p.Count, 0)
		}
	})

	t.Run("every entry has required fields", func(t *testing.T) {
		for _, e := range idx.Entries {
			require.NotEmpty(t, e.Title, "entry missing title")
			require.NotEmpty(t, e.URL, "entry missing URL")
			require.NotEmpty(t, e.Product, "entry missing product")
			require.Contains(t, e.URL, "grafana.com", "URL not from grafana.com: %s", e.URL)
		}
	})

	t.Run("product counts sum to entry count", func(t *testing.T) {
		total := 0
		for _, p := range idx.Products() {
			total += p.Count
		}
		require.Equal(t, idx.EntryCount(), total)
	})

	t.Run("non-product sections excluded", func(t *testing.T) {
		for _, p := range idx.Products() {
			require.NotEqual(t, "Documentation home", p.Name)
			require.NotEqual(t, "Copyright notice", p.Name)
		}
		for _, e := range idx.Entries {
			require.NotEqual(t, "Documentation home", e.Product)
			require.NotEqual(t, "Copyright notice", e.Product)
		}
	})
}

func TestIndexRejectsNonGrafanaURLs(t *testing.T) {
	input := `## Test Product
- [Good](https://grafana.com/docs/test.md): valid entry
- [Bad](https://evil.com/docs/steal.md): should be skipped
- [Also Bad](http://grafana.com/docs/http.md): wrong scheme
- [Good Two](https://grafana.com/docs/other.md): also valid
`
	idx, err := LoadIndexFromReader(strings.NewReader(input))
	require.NoError(t, err)
	require.Equal(t, 2, idx.EntryCount(), "only grafana.com URLs should be indexed")
	for _, e := range idx.Entries {
		require.True(t, strings.HasPrefix(e.URL, "https://grafana.com/"),
			"entry URL %q should start with https://grafana.com/", e.URL)
	}
}

func TestProductNameCleaning(t *testing.T) {
	tests := []struct {
		header   string
		expected string
	}{
		{"## Grafana Tempo documentation", "Grafana Tempo"},
		{"## Grafana Agent Documentation", "Grafana Agent"},
		{"## k6 Studio", "k6 Studio"},
		{"## OpenTelemetry at Grafana Labs", "OpenTelemetry at Grafana Labs"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			m := productHeader.FindStringSubmatch(tt.header)
			require.NotNil(t, m)
			name := m[1]
			name = trimDocSuffix(name)
			require.Equal(t, tt.expected, name)
		})
	}
}

// trimDocSuffix mirrors the logic in LoadIndexFromReader.
func trimDocSuffix(s string) string {
	s = trimSuffix(s, " documentation")
	s = trimSuffix(s, " Documentation")
	return s
}

func trimSuffix(s, suffix string) string {
	if len(s) > len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
