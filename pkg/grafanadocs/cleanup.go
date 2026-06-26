// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"fmt"
	"regexp"
	"strings"
)

// Cleanup removes presentation boilerplate from fetched markdown while
// preserving all documented content (invariant I8).
//
// Strips: YAML frontmatter, Hugo shortcodes, HTML comments.
// Preserves: prose, headings, code blocks (and their contents), tables, lists.
//
// Code blocks are extracted before stripping so that shortcodes and HTML
// comments inside fenced code blocks are not removed (invariant I14).
func Cleanup(raw []byte) []byte {
	s := string(raw)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// Trim the leading/trailing edge up front so frontmatter detection is
	// consistent across passes (the final TrimSpace would otherwise expose a
	// frontmatter delimiter only on a second pass — breaking idempotency).
	s = strings.TrimSpace(s)
	s = stripFrontmatter(s)
	s = strings.TrimSpace(s)

	s, blocks := extractCodeBlocks(s)
	for {
		prev := s
		s = strings.TrimSpace(s)
		s = stripFrontmatter(s)
		s = stripShortcodes(s)
		s = stripHTMLComments(s)
		if s == prev {
			break
		}
	}
	s = collapseBlankLines(s)
	s = restoreCodeBlocks(s, blocks)

	return []byte(strings.TrimSpace(s) + "\n")
}

// stripFrontmatter removes YAML (---) or TOML (+++) frontmatter at the start.
// The closing delimiter must appear at the start of a line AND be the complete
// line (followed by \n or end-of-string) to avoid matching a delimiter that
// appears mid-line inside a frontmatter value or on a line with trailing content.
func stripFrontmatter(s string) string {
	for _, delim := range []string{"---", "+++"} {
		if !strings.HasPrefix(s, delim) {
			continue
		}
		rest := s[len(delim):]
		needle := "\n" + delim
		pos := 0
		for {
			idx := strings.Index(rest[pos:], needle)
			if idx < 0 {
				return s
			}
			end := pos + idx
			after := rest[end+len(needle):]
			if len(after) == 0 || after[0] == '\n' {
				if len(after) > 0 {
					return after[1:]
				}
				return after
			}
			pos = end + 1
		}
	}
	return s
}

// shortcode matches Hugo shortcodes: {{< ... >}} and {{% ... %}}.
// Uses non-greedy .*? to avoid over-matching when shortcode arguments contain
// > or % characters.
var shortcode = regexp.MustCompile(`\{\{<.*?>}}` + `|` + `\{\{%.*?%}}`)

func stripShortcodes(s string) string {
	return shortcode.ReplaceAllString(s, "")
}

// htmlComment matches <!-- ... --> (non-greedy).
var htmlComment = regexp.MustCompile(`(?s)<!--.*?-->`)

func stripHTMLComments(s string) string {
	return htmlComment.ReplaceAllString(s, "")
}

// extractCodeBlocks replaces fenced code blocks with numbered placeholders
// and returns the blocks separately. This protects code block contents from
// the shortcode/comment strippers.
func extractCodeBlocks(s string) (string, []string) {
	var blocks []string
	lines := strings.Split(s, "\n")
	var out []string
	var openChar byte
	var openLen int
	var block []string

	emit := func() {
		placeholder := fmt.Sprintf("\x00CODEBLOCK_%d\x00", len(blocks))
		blocks = append(blocks, strings.Join(block, "\n"))
		out = append(out, placeholder)
		block = nil
	}

	for _, line := range lines {
		c, n, blankAfter := fenceInfo(line)
		if openLen == 0 {
			if n > 0 {
				openChar, openLen = c, n
				block = []string{line}
			} else {
				out = append(out, line)
			}
			continue
		}
		block = append(block, line)
		if n > 0 && c == openChar && n >= openLen && blankAfter {
			openChar, openLen = 0, 0
			emit()
		}
	}
	if openLen > 0 {
		// Unclosed fence (e.g. truncated page): emit the opening fence and
		// content in their original position rather than re-ordering them.
		emit()
	}
	return strings.Join(out, "\n"), blocks
}

// restoreCodeBlocks replaces numbered placeholders with the original code blocks.
func restoreCodeBlocks(s string, blocks []string) string {
	for i, block := range blocks {
		placeholder := fmt.Sprintf("\x00CODEBLOCK_%d\x00", i)
		s = strings.Replace(s, placeholder, block, 1)
	}
	return s
}

// collapseBlankLines reduces runs of 3+ blank lines to 2. Single-pass via
// strings.Builder to avoid O(n*m) repeated scans.
func collapseBlankLines(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	consecutive := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			consecutive++
		} else {
			consecutive = 0
		}
		if consecutive <= 3 {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
