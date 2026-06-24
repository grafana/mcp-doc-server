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

func TestExcerpt_RangeInvariants(t *testing.T) {
	raw, err := os.ReadFile("testdata/pages/sample.md")
	require.NoError(t, err)

	doc := &Doc{URL: "test", Content: Cleanup(raw)}
	doc.EnsureLines()

	cases := []ExcerptOpts{
		{},
		{Offset: 0, Limit: 10},
		{Offset: 5, Limit: 20},
		{Offset: 99999},
		{Section: "Authentication"},
		{Section: "nonexistent"},
	}

	for _, opts := range cases {
		result := Excerpt(doc, opts)

		require.GreaterOrEqual(t, result.Total, 0, "Total must be non-negative")

		if result.Content == "" {
			continue
		}

		require.GreaterOrEqual(t, result.Start, 1, "Start must be >= 1")
		require.LessOrEqual(t, result.End, result.Total, "End must be <= Total")
		require.LessOrEqual(t, result.Start, result.End, "Start must be <= End")

		lines := strings.Split(result.Content, "\n")
		expectedLines := result.End - result.Start + 1
		require.Equal(t, expectedLines, len(lines),
			"line count must match range [%d, %d]", result.Start, result.End)
	}
}

func TestExcerptBySection_SkipsCodeFenceHeadings(t *testing.T) {
	content := []byte("## Config\n\nSome intro.\n\n```yaml\n# Storage\nbackend: s3\n```\n\n## Storage\n\nReal storage section.\n\n## Other\n")
	doc := &Doc{URL: "https://grafana.com/docs/test.md", Content: content}

	t.Run("matches real heading not code fence comment", func(t *testing.T) {
		result := Excerpt(doc, ExcerptOpts{Section: "Storage"})
		require.NotEmpty(t, result.Content)
		require.Contains(t, result.Content, "## Storage")
		require.Contains(t, result.Content, "Real storage section.")
		require.NotContains(t, result.Content, "backend: s3")
	})

	t.Run("section end not confused by code fence comment", func(t *testing.T) {
		content := []byte("## Main\n\nIntro.\n\n```yaml\n# Main\nkey: val\n```\n\nMore content under Main.\n\n## Next\n")
		doc := &Doc{URL: "https://grafana.com/docs/test.md", Content: content}
		result := Excerpt(doc, ExcerptOpts{Section: "Main"})
		require.Contains(t, result.Content, "More content under Main.")
	})
}
