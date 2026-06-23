// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExcerpt(t *testing.T) {
	raw, err := os.ReadFile("testdata/pages/sample.md")
	require.NoError(t, err)

	doc := &Doc{URL: "https://grafana.com/docs/test.md", Content: Cleanup(raw)}

	t.Run("section by heading", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Section: "Authentication"})
		require.NotEmpty(t, result.Content)
		require.Contains(t, result.Content, "## Authentication")
		require.Contains(t, result.Content, "bearer tokens")
		// Should not contain sibling sections.
		require.NotContains(t, result.Content, "## Storage")
		require.Greater(t, result.Start, 0)
		require.Greater(t, result.Total, 0)
	})

	t.Run("section case insensitive", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Section: "authentication"})
		require.NotEmpty(t, result.Content)
	})

	t.Run("subsection extraction", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Section: "Local storage"})
		require.NotEmpty(t, result.Content)
		require.Contains(t, result.Content, "Local storage")
		require.NotContains(t, result.Content, "Remote storage")
	})

	t.Run("missing section returns empty", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Section: "nonexistent"})
		require.Empty(t, result.Content)
		require.Greater(t, result.Total, 0)
	})

	t.Run("offset paging", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Offset: 0, Limit: 5})
		lines := strings.Split(result.Content, "\n")
		require.LessOrEqual(t, len(lines), 5)
		require.Equal(t, 1, result.Start)
		require.Equal(t, 5, result.End)
	})

	t.Run("default limit bounds output", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Offset: 0})
		lines := strings.Split(result.Content, "\n")
		require.LessOrEqual(t, len(lines), defaultExcerptLines)
	})

	t.Run("offset past end returns empty", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Offset: 99999})
		require.Empty(t, result.Content)
	})
}
