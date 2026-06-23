// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"strings"
)

// Heading is a markdown heading with its position in the document.
type Heading struct {
	Level int
	Text  string
	Line  int // 1-indexed line number
}

// Outline extracts the heading structure from a document.
func Outline(doc *Doc) []Heading {
	doc.EnsureLines()
	var headings []Heading

	for i, line := range doc.Lines {
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

// parseHeading returns (level, text) for ATX headings (# ... ######).
// Returns (0, "") for non-heading lines.
func parseHeading(line string) (int, string) {
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

	text := strings.TrimSpace(trimmed[level+1:])
	return level, text
}
