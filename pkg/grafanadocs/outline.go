// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import "strings"

// Heading is a markdown heading with its position in the document.
type Heading struct {
	Level int
	Text  string
	Line  int // 1-indexed line number
}

// Outline extracts the heading structure from a document.
// Lines inside fenced code blocks (``` ... ```) are skipped so that
// comment lines (e.g. YAML `# comment`) are not misidentified as headings.
func Outline(doc *Doc) []Heading {
	doc.EnsureLines()
	var headings []Heading
	var fence fenceTracker

	for i, line := range doc.Lines {
		if fence.skip(line) {
			continue
		}
		level, text := parseHeading(line)
		if level > 0 {
			headings = append(headings, Heading{
				Level: level,
				Text:  text,
				Line:  i + 1,
			})
		}
	}
	return headings
}

// fenceInfo inspects a line as a potential fenced code block boundary. It
// returns the fence character ('`' or '~'), the run length, and whether the
// line has only whitespace after the run (required for a closing fence — an
// info string is allowed only on the opening fence). A non-fence line returns
// (0, 0, false). Per CommonMark, the run must be 3+ chars and indented ≤3.
func fenceInfo(line string) (byte, int, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || len(trimmed) < 3 {
		return 0, 0, false
	}
	c := trimmed[0]
	if c != '`' && c != '~' {
		return 0, 0, false
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == c {
		n++
	}
	if n < 3 {
		return 0, 0, false
	}
	blankAfter := strings.TrimRight(trimmed[n:], " \t\r") == ""
	return c, n, blankAfter
}

// fenceTracker tracks fenced code block state across lines, handling
// variable-length fences per CommonMark: a closing fence must use the same
// character as the opening fence and be at least as long, with no info string.
type fenceTracker struct {
	char   byte
	length int
}

// skip reports whether the line is part of a code block (a fence boundary or a
// line inside an open fence) and therefore must not be parsed as a heading.
// It updates the tracker state.
func (f *fenceTracker) skip(line string) bool {
	c, n, blankAfter := fenceInfo(line)
	if f.length == 0 {
		if n > 0 {
			f.char, f.length = c, n
			return true
		}
		return false
	}
	if n > 0 && c == f.char && n >= f.length && blankAfter {
		f.char, f.length = 0, 0
	}
	return true
}

// parseHeading returns (level, text) for ATX headings (# ... ######).
// Returns (0, "") for non-heading lines.
//
// Per CommonMark: lines indented 4+ spaces are code blocks, not headings; and
// an optional trailing run of #s (preceded by a space) is a closing sequence
// and is stripped from the heading text.
func parseHeading(line string) (int, string) {
	if len(line)-len(strings.TrimLeft(line, " ")) >= 4 {
		return 0, ""
	}

	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return 0, ""
	}

	level := 0
	for _, c := range trimmed {
		if c == '#' {
			level++
		} else {
			break
		}
	}
	if level > 6 || level >= len(trimmed) {
		return 0, ""
	}
	// Must have a space after the #s.
	if trimmed[level] != ' ' {
		return 0, ""
	}

	text := stripClosingHashes(strings.TrimSpace(trimmed[level+1:]))
	if text == "" {
		return 0, ""
	}
	return level, text
}

// stripClosingHashes removes an optional ATX closing sequence: a trailing run
// of #s preceded by whitespace (e.g. "Storage ##" -> "Storage"). A # that is
// part of a word ("C#") or not preceded by a space ("foo#") is preserved.
func stripClosingHashes(text string) string {
	trimmed := strings.TrimRight(text, " \t")
	i := len(trimmed)
	for i > 0 && trimmed[i-1] == '#' {
		i--
	}
	if i == len(trimmed) {
		return text // no trailing #
	}
	if i == 0 {
		return "" // entire text is #s
	}
	if trimmed[i-1] == ' ' || trimmed[i-1] == '\t' {
		return strings.TrimRight(trimmed[:i], " \t")
	}
	return text // not a valid closing sequence
}
