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

func TestOutline_SkipsCodeFenceComments(t *testing.T) {
	content := []byte("# Real Heading\n\n```yaml\n# YAML comment that looks like heading\n# Another comment\nkey: value\n```\n\n## Second Heading\n\n~~~bash\n# Shell comment\n~~~\n\n### Third Heading\n")
	doc := &Doc{URL: "https://grafana.com/docs/test.md", Content: content}
	headings := Outline(doc)

	texts := make([]string, len(headings))
	for i, h := range headings {
		texts[i] = h.Text
	}

	require.Equal(t, 3, len(headings), "should find exactly 3 real headings, not code fence comments")
	require.Equal(t, "Real Heading", headings[0].Text)
	require.Equal(t, 1, headings[0].Level)
	require.Equal(t, "Second Heading", headings[1].Text)
	require.Equal(t, 2, headings[1].Level)
	require.Equal(t, "Third Heading", headings[2].Text)
	require.Equal(t, 3, headings[2].Level)

	require.NotContains(t, texts, "YAML comment that looks like heading")
	require.NotContains(t, texts, "Another comment")
	require.NotContains(t, texts, "Shell comment")
}

func TestOutline_MismatchedFenceMarkers(t *testing.T) {
	// A backtick-opened fence should only close with backticks, not tildes.
	content := []byte("# Real\n\n```yaml\n# Inside backtick fence\nvalue: x\n~~~\n# Still inside — tilde shouldn't close backtick fence\n```\n\n## After\n")
	doc := &Doc{URL: "test", Content: content}
	headings := Outline(doc)

	texts := make([]string, len(headings))
	for i, h := range headings {
		texts[i] = h.Text
	}

	require.Equal(t, 2, len(headings), "should find exactly 2 real headings")
	require.Equal(t, "Real", headings[0].Text)
	require.Equal(t, "After", headings[1].Text)
	require.NotContains(t, texts, "Inside backtick fence")
	require.NotContains(t, texts, "Still inside — tilde shouldn't close backtick fence")
}

func FuzzParseHeading(f *testing.F) {
	f.Add("# Hello")
	f.Add("## Sub heading")
	f.Add("###### Deep")
	f.Add("####### Too deep")
	f.Add("#NoSpace")
	f.Add("not a heading")
	f.Add("")
	f.Add("### ")

	f.Fuzz(func(t *testing.T, line string) {
		level, text := parseHeading(line)
		if level < 0 || level > 6 {
			t.Errorf("level %d out of range [0,6]", level)
		}
		if level == 0 && text != "" {
			t.Errorf("level 0 should have empty text, got %q", text)
		}
		_ = text
	})
}

func FuzzFenceInfo(f *testing.F) {
	f.Add("```")
	f.Add("~~~")
	f.Add("```yaml")
	f.Add("````")
	f.Add("  ```")
	f.Add("    ```")
	f.Add("not a fence")
	f.Add("")

	f.Fuzz(func(t *testing.T, line string) {
		c, n, _ := fenceInfo(line)
		if n == 0 {
			if c != 0 {
				t.Errorf("non-fence line should return char 0, got %q", c)
			}
			return
		}
		if c != '`' && c != '~' {
			t.Errorf("fence char must be backtick or tilde, got %q", c)
		}
		if n < 3 {
			t.Errorf("fence length must be >= 3, got %d", n)
		}
	})
}

func TestParseHeading_StripsTrailingHashes(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantLevel int
		wantText  string
	}{
		{"closed atx", "## Storage ##", 2, "Storage"},
		{"single trailing hash", "# Title #", 1, "Title"},
		{"many trailing hashes", "### Config #####", 3, "Config"},
		{"hash in word preserved", "## C#", 2, "C#"},
		{"hash without preceding space preserved", "## foo#", 2, "foo#"},
		{"only hashes is not a heading", "## ###", 0, ""},
		{"plain heading unchanged", "## Authentication", 2, "Authentication"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, text := parseHeading(tt.line)
			require.Equal(t, tt.wantLevel, level)
			require.Equal(t, tt.wantText, text)
		})
	}
}

func TestParseHeading_IgnoresIndentedCodeBlocks(t *testing.T) {
	// A line indented 4+ spaces is an indented code block per CommonMark,
	// not a heading — even though it starts with '#'.
	tests := []struct {
		name      string
		line      string
		wantLevel int
	}{
		{"4 spaces is code", "    # not a heading", 0},
		{"5 spaces is code", "     # also code", 0},
		{"3 spaces is heading", "   # still a heading", 1},
		{"no indent is heading", "# heading", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, _ := parseHeading(tt.line)
			require.Equal(t, tt.wantLevel, level)
		})
	}
}

func TestOutline_IgnoresIndentedCodeHash(t *testing.T) {
	content := []byte("# Real Heading\n\nSome prose.\n\n    # This is indented code, not a heading\n\n## Second\n")
	doc := &Doc{URL: "test", Content: content}
	headings := Outline(doc)

	texts := make([]string, len(headings))
	for i, h := range headings {
		texts[i] = h.Text
	}
	require.Equal(t, []string{"Real Heading", "Second"}, texts)
}

func TestOutline_VariableLengthFences(t *testing.T) {
	// A 4-backtick fence containing a 3-backtick line must not be closed early
	// by the inner 3-backtick line. The "# Inner heading" inside is code.
	content := []byte("# Real\n\n````markdown\n```\n# Inner heading (code)\n```\n````\n\n## After\n")
	doc := &Doc{URL: "test", Content: content}
	headings := Outline(doc)

	texts := make([]string, len(headings))
	for i, h := range headings {
		texts[i] = h.Text
	}
	require.Equal(t, []string{"Real", "After"}, texts)
	require.NotContains(t, texts, "Inner heading (code)")
}

func TestOutline_BoundsInvariants(t *testing.T) {
	raw, err := os.ReadFile("testdata/pages/sample.md")
	require.NoError(t, err)

	doc := &Doc{URL: "test", Content: Cleanup(raw)}
	doc.EnsureLines()
	headings := Outline(doc)

	for _, h := range headings {
		require.GreaterOrEqual(t, h.Level, 1, "level must be >= 1")
		require.LessOrEqual(t, h.Level, 6, "level must be <= 6")
		require.GreaterOrEqual(t, h.Line, 1, "line must be >= 1")
		require.LessOrEqual(t, h.Line, len(doc.Lines), "line must be <= total lines")
		require.NotEmpty(t, h.Text, "heading text must not be empty")
	}
}

func TestOutline_SampleFixtureNoCodeFenceHeadings(t *testing.T) {
	raw, err := os.ReadFile("testdata/pages/sample.md")
	require.NoError(t, err)

	doc := &Doc{URL: "https://grafana.com/docs/test.md", Content: Cleanup(raw)}
	headings := Outline(doc)

	texts := make([]string, len(headings))
	for i, h := range headings {
		texts[i] = h.Text
	}

	require.NotContains(t, texts, "Enable authentication for all API endpoints.")
	require.NotContains(t, texts, "Supported types: bearer, basic, api-key.")
}
