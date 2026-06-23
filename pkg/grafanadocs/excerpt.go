// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"strings"
)

const defaultExcerptLines = 80 // ~2000 tokens at ~25 tokens/line

// ExcerptOpts controls how a document is excerpted for bounded retrieval.
type ExcerptOpts struct {
	Section string // heading text to extract (empty = use offset/limit)
	Offset  int    // 0-indexed line offset (used when Section is empty)
	Limit   int    // max lines to return (0 = default)
}

// ExcerptResult contains the excerpted content and position metadata.
type ExcerptResult struct {
	Content string
	Start   int // 1-indexed start line
	End     int // 1-indexed end line (inclusive)
	Total   int // total lines in the document
}

// Excerpt extracts a bounded portion of a document. If Section is set, it
// returns the content under that heading. Otherwise it uses Offset/Limit for
// paging. This enforces invariant I4 (bounded responses).
func Excerpt(doc *Doc, opts ExcerptOpts) ExcerptResult {
	doc.EnsureLines()
	total := len(doc.Lines)

	if opts.Section != "" {
		return excerptBySection(doc, opts.Section, total)
	}
	return excerptByOffset(doc, opts.Offset, opts.Limit, total)
}

func excerptBySection(doc *Doc, section string, total int) ExcerptResult {
	sectionLower := strings.ToLower(section)
	startIdx := -1
	startLevel := 0

	for i, line := range doc.Lines {
		level, text := parseHeading(line)
		if level > 0 && strings.ToLower(text) == sectionLower {
			startIdx = i
			startLevel = level
			break
		}
	}

	if startIdx < 0 {
		return ExcerptResult{Total: total}
	}

	// Find the end: next heading at same or higher level.
	endIdx := total
	for i := startIdx + 1; i < total; i++ {
		level, _ := parseHeading(doc.Lines[i])
		if level > 0 && level <= startLevel {
			endIdx = i
			break
		}
	}

	content := strings.Join(doc.Lines[startIdx:endIdx], "\n")
	return ExcerptResult{
		Content: content,
		Start:   startIdx + 1,
		End:     endIdx,
		Total:   total,
	}
}

func excerptByOffset(doc *Doc, offset, limit, total int) ExcerptResult {
	if limit <= 0 {
		limit = defaultExcerptLines
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return ExcerptResult{Total: total}
	}

	end := offset + limit
	if end > total {
		end = total
	}

	content := strings.Join(doc.Lines[offset:end], "\n")
	return ExcerptResult{
		Content: content,
		Start:   offset + 1,
		End:     end,
		Total:   total,
	}
}
