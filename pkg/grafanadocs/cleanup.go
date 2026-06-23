// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"regexp"
	"strings"
)

// Cleanup removes presentation boilerplate from fetched markdown while
// preserving all documented content (invariant I8).
//
// Strips: YAML frontmatter, Hugo shortcodes, HTML comments.
// Preserves: prose, headings, code blocks, tables, lists.
func Cleanup(raw []byte) []byte {
	s := string(raw)
	s = stripFrontmatter(s)
	s = stripShortcodes(s)
	s = stripHTMLComments(s)
	s = collapseBlankLines(s)
	return []byte(strings.TrimSpace(s) + "\n")
}

// stripFrontmatter removes YAML frontmatter delimited by --- at the start.
func stripFrontmatter(s string) string {
	if !strings.HasPrefix(s, "---") {
		return s
	}
	end := strings.Index(s[3:], "---")
	if end < 0 {
		return s
	}
	return s[end+6:] // skip past closing ---
}

// shortcode matches Hugo shortcodes: {{< ... >}} and {{% ... %}}.
var shortcode = regexp.MustCompile(`\{\{[<%]\s*/?[^}]*[>%]\}\}`)

func stripShortcodes(s string) string {
	return shortcode.ReplaceAllString(s, "")
}

// htmlComment matches <!-- ... --> (non-greedy).
var htmlComment = regexp.MustCompile(`(?s)<!--.*?-->`)

func stripHTMLComments(s string) string {
	return htmlComment.ReplaceAllString(s, "")
}

// collapseBlankLines reduces runs of 3+ blank lines to 2.
func collapseBlankLines(s string) string {
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return s
}
