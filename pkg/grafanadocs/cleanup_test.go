// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanup(t *testing.T) {
	raw, err := os.ReadFile("testdata/pages/sample.md")
	require.NoError(t, err)

	cleaned := Cleanup(raw)
	result := string(cleaned)

	t.Run("strips frontmatter", func(t *testing.T) {
		require.NotContains(t, result, "---")
		require.NotContains(t, result, "title: Sample")
	})

	t.Run("strips shortcodes", func(t *testing.T) {
		require.NotContains(t, result, "{{<")
		require.NotContains(t, result, "docs/shared")
	})

	t.Run("strips HTML comments", func(t *testing.T) {
		require.NotContains(t, result, "<!--")
		require.NotContains(t, result, "Navigation placeholder")
	})

	t.Run("preserves headings", func(t *testing.T) {
		require.Contains(t, result, "# Sample Configuration")
		require.Contains(t, result, "## Authentication")
		require.Contains(t, result, "## Storage")
	})

	t.Run("preserves code blocks", func(t *testing.T) {
		require.Contains(t, result, "```yaml")
		require.Contains(t, result, "auth:")
		require.Contains(t, result, "```")
	})

	t.Run("preserves prose", func(t *testing.T) {
		require.Contains(t, result, "Use bearer tokens for secure access.")
	})
}
