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

func TestEntriesBeforeProductHeaderDropped(t *testing.T) {
	input := `- [Orphan](https://grafana.com/docs/orphan.md): no product header yet
## Grafana Tempo documentation
- [Real](https://grafana.com/docs/tempo.md): has a product
`
	idx, err := LoadIndexFromReader(strings.NewReader(input))
	require.NoError(t, err)
	require.Equal(t, 1, idx.EntryCount(), "entries before any product header must be dropped")
	require.Equal(t, "Grafana Tempo", idx.Entries[0].Product)
}

func TestLoadIndex_RejectsNonHTTPS(t *testing.T) {
	_, err := LoadIndex(t.Context(), "file:///etc/passwd")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rejected")

	_, err = LoadIndex(t.Context(), "http://grafana.com/llms-full.txt")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rejected")
}

func FuzzLoadIndexFromReader(f *testing.F) {
	f.Add("## Product A documentation\n- [Title](https://grafana.com/docs/a.md): desc\n")
	f.Add("## Bad\n- [X](https://evil.com/x.md): evil\n")
	f.Add("")
	f.Add("just some random text\n\n\n")
	f.Add("## \n- [](https://grafana.com/docs/empty.md): \n")

	f.Fuzz(func(t *testing.T, input string) {
		idx, err := LoadIndexFromReader(strings.NewReader(input))
		if err != nil {
			return
		}
		for _, e := range idx.Entries {
			if e.Title == "" {
				t.Error("entry has empty Title")
			}
			if e.URL == "" {
				t.Error("entry has empty URL")
			}
			if e.Product == "" {
				t.Error("entry has empty Product")
			}
			if !strings.HasPrefix(e.URL, "https://grafana.com/") {
				t.Errorf("entry URL %q not from grafana.com", e.URL)
			}
		}
	})
}

func TestLoadIndex_StripsBOM(t *testing.T) {
	// A UTF-8 BOM prefix must not prevent the first product header from matching.
	input := "\ufeff## Grafana Tempo documentation\n- [Configure](https://grafana.com/docs/tempo/latest/configuration.md): Configure Tempo\n"
	idx, err := LoadIndexFromReader(strings.NewReader(input))
	require.NoError(t, err)
	require.Equal(t, 1, idx.EntryCount(), "BOM-prefixed first product must still be parsed")
	require.Equal(t, "Grafana Tempo", idx.Entries[0].Product)
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
	s = strings.TrimSuffix(s, " documentation")
	s = strings.TrimSuffix(s, " Documentation")
	return s
}
