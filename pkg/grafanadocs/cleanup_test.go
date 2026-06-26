// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"bytes"
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

	t.Run("strips shortcodes outside code blocks", func(t *testing.T) {
		require.NotContains(t, result, "agent-deprecation.md")
	})

	t.Run("strips HTML comments outside code blocks", func(t *testing.T) {
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

	t.Run("preserves HTML comments inside code blocks", func(t *testing.T) {
		require.Contains(t, result, "<!-- This comment should be preserved inside the code block -->")
	})

	t.Run("preserves shortcodes inside code blocks", func(t *testing.T) {
		require.Contains(t, result, `{{< docs/shared source="tempo" lookup="config.md" >}}`)
	})
}

func TestCleanup_FrontmatterWithDashesInValue(t *testing.T) {
	input := "---\ntitle: My Page\ndescription: Use --- for separators\n---\n\n# Content\n"
	result := string(Cleanup([]byte(input)))

	require.NotContains(t, result, "title:")
	require.NotContains(t, result, "description:")
	require.Contains(t, result, "# Content")
}

func TestCleanup_PreservesBlankLinesInCodeBlocks(t *testing.T) {
	input := "# Title\n\n```\nline1\n\n\n\nline2\n```\n"
	result := string(Cleanup([]byte(input)))

	require.Contains(t, result, "line1\n\n\n\nline2", "consecutive blank lines inside code blocks must be preserved")
}

func TestCleanup_UnclosedFencePreserved(t *testing.T) {
	input := "# Title\n\n```yaml\nkey: value\n# not a heading\n"
	result := string(Cleanup([]byte(input)))

	require.Contains(t, result, "# Title")
	require.Contains(t, result, "```yaml")
	require.Contains(t, result, "key: value")
	require.Contains(t, result, "# not a heading")
}

func TestCleanup_ShortcodeWithAngleBracketInArgs(t *testing.T) {
	input := `# Title

{{< highlight go "linenos=true" >}}
fmt.Println("hello")
{{< /highlight >}}

Keep this.
`
	result := string(Cleanup([]byte(input)))

	require.NotContains(t, result, "{{<")
	require.Contains(t, result, "Keep this.")
}

func TestCleanup_CollapseBlankLinesPreservesDouble(t *testing.T) {
	result := collapseBlankLines("a\n\nb\n\n\n\nc\n\n\n\n\nd")
	require.Equal(t, "a\n\nb\n\n\nc\n\n\nd", result)
}

func TestCleanup_CodeBlockProtection(t *testing.T) {
	input := "# Title\n\n```html\n<!-- keep this -->\n{{< shortcode >}}\n```\n\n<!-- strip this -->\n{{< strip_this >}}\n"
	result := string(Cleanup([]byte(input)))

	t.Run("strips outside code blocks", func(t *testing.T) {
		require.NotContains(t, result, "strip this")
		require.NotContains(t, result, "strip_this")
	})

	t.Run("preserves inside code blocks", func(t *testing.T) {
		require.Contains(t, result, "<!-- keep this -->")
		require.Contains(t, result, "{{< shortcode >}}")
	})
}

func TestCleanup_CRLF(t *testing.T) {
	t.Run("collapses blank lines with CRLF endings", func(t *testing.T) {
		input := "# Title\r\n\r\n\r\n\r\n\r\nMore text.\r\n"
		result := string(Cleanup([]byte(input)))

		require.NotContains(t, result, "\r", "carriage returns should be normalized away")
		require.NotContains(t, result, "\n\n\n\n", "3+ blank lines must be collapsed to 2 even with CRLF input")
		require.Contains(t, result, "# Title")
		require.Contains(t, result, "More text.")
	})

	t.Run("strips CRLF frontmatter", func(t *testing.T) {
		input := "---\r\ntitle: Test\r\n---\r\n\r\n# Content\r\n"
		result := string(Cleanup([]byte(input)))

		require.NotContains(t, result, "title:")
		require.NotContains(t, result, "\r")
		require.Contains(t, result, "# Content")
	})

	t.Run("preserves code block content with CRLF", func(t *testing.T) {
		input := "# T\r\n\r\n```yaml\r\nkey: value\r\n```\r\n"
		result := string(Cleanup([]byte(input)))
		require.Contains(t, result, "key: value")
	})
}

func TestCleanup_VariableLengthFences(t *testing.T) {
	t.Run("4-backtick fence protects inner 3-backtick block content", func(t *testing.T) {
		// The shortcode lives between the inner ``` pair, which is itself inside
		// the outer ```` fence. If the outer fence is closed early by the inner
		// ```, the shortcode is treated as prose and wrongly stripped.
		input := "# Title\n\n````markdown\n```\n{{< keep-me >}}\n```\n````\n\n{{< strip-me >}}\n"
		result := string(Cleanup([]byte(input)))

		require.Contains(t, result, "{{< keep-me >}}", "shortcode inside nested fence must be preserved")
		require.NotContains(t, result, "strip-me", "shortcode outside the fence is still stripped")
	})
}

func TestCleanup_TOMLFrontmatter(t *testing.T) {
	input := "+++\ntitle = \"Test\"\ndate = 2024-01-01\n+++\n\n# Content\n"
	result := string(Cleanup([]byte(input)))

	require.NotContains(t, result, "title =")
	require.NotContains(t, result, "+++")
	require.Contains(t, result, "# Content")
}

func TestCleanup_Idempotent(t *testing.T) {
	inputs := []string{
		"# Title\n\n```yaml\nkey: value\n```\n",
		"---\ntitle: Test\n---\n\n# Content\n",
		"{{< shortcode >}}\n<!-- comment -->\n# Heading\n",
		"",
		"```\nunclosed\n",
		"# A\n\n\n\n\n\n# B\n",
		"<!-- c1 --><!-- c2 -->\n{{< s >}}text\n",
		"{<!---->{<>}}",
	}
	for i, input := range inputs {
		first := Cleanup([]byte(input))
		second := Cleanup(first)
		require.Equal(t, string(first), string(second), "input %d not idempotent", i)
	}
}

func FuzzCleanup(f *testing.F) {
	f.Add([]byte("# Hello\n\n```yaml\nkey: value\n```\n"))
	f.Add([]byte("---\ntitle: Test\n---\n\n# Content\n"))
	f.Add([]byte("{{< shortcode >}}\n<!-- comment -->\n"))
	f.Add([]byte(""))
	f.Add([]byte("```\nunclosed fence\n"))

	f.Fuzz(func(t *testing.T, input []byte) {
		result := Cleanup(input)

		// Must end with newline.
		if len(result) > 0 && result[len(result)-1] != '\n' {
			t.Errorf("result does not end with newline")
		}

		// Idempotent: Cleanup(Cleanup(x)) == Cleanup(x).
		second := Cleanup(result)
		if !bytes.Equal(result, second) {
			t.Errorf("Cleanup is not idempotent:\n  first:  %q\n  second: %q", result, second)
		}
	})
}
