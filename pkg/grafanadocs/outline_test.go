// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutline(t *testing.T) {
	raw, err := os.ReadFile("testdata/pages/sample.md")
	require.NoError(t, err)

	doc := &Doc{URL: "https://grafana.com/docs/test.md", Content: Cleanup(raw)}
	headings := Outline(doc)

	require.Greater(t, len(headings), 0)

	t.Run("correct levels", func(t *testing.T) {
		// Should find h1, h2, and h3 headings.
		levels := map[int]bool{}
		for _, h := range headings {
			levels[h.Level] = true
		}
		require.True(t, levels[1], "should have h1")
		require.True(t, levels[2], "should have h2")
		require.True(t, levels[3], "should have h3")
	})

	t.Run("line numbers are positive and ascending", func(t *testing.T) {
		prev := 0
		for _, h := range headings {
			require.Greater(t, h.Line, prev)
			prev = h.Line
		}
	})

	t.Run("expected headings present", func(t *testing.T) {
		texts := make([]string, len(headings))
		for i, h := range headings {
			texts[i] = h.Text
		}
		require.Contains(t, texts, "Sample Configuration")
		require.Contains(t, texts, "Authentication")
		require.Contains(t, texts, "Local storage")
	})
}
