// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"math"
	"strings"
	"unicode"
)

// SearchOpts controls search behavior.
type SearchOpts struct {
	Product string // filter to a specific product (empty = all)
	Limit   int    // max results (0 = default 5)
}

// Search finds entries matching the query using word-boundary matching,
// TF-IDF weighting, and phrase bonuses. No network calls — operates entirely
// on the in-memory index.
func Search(idx *Index, query string, opts SearchOpts) []Entry {
	if opts.Limit <= 0 {
		opts.Limit = 5
	}

	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	queryLower := strings.ToLower(query)

	var results []scored
	for _, e := range idx.Entries {
		if opts.Product != "" && !containsFold(e.Product, opts.Product) {
			continue
		}
		s := score(e, tokens, queryLower, idx.idf)
		if s > 0 {
			results = append(results, scored{entry: e, score: s})
		}
	}

	sortScored(results)

	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	out := make([]Entry, len(results))
	for i, r := range results {
		out[i] = r.entry
	}
	return out
}

// score computes a relevance score using:
//   - Word boundary matching (not substring — "rate" won't match "migrate")
//   - TF-IDF weighting (rare terms score higher)
//   - Title matches weighted 3x over description
//   - Bonus for matching all query tokens
//   - Bonus for exact phrase in title
func score(e Entry, tokens []string, queryLower string, idf map[string]float64) int {
	titleWords := wordSet(e.Title)
	descWords := wordSet(e.Description)

	var total float64
	matched := 0

	for _, tok := range tokens {
		weight := idf[tok]
		if weight == 0 {
			weight = 1
		}

		if titleWords[tok] {
			total += 3 * weight
			matched++
		} else if descWords[tok] {
			total += 1 * weight
			matched++
		}
	}

	if total == 0 {
		return 0
	}

	// Bonus: all tokens matched.
	if matched == len(tokens) && len(tokens) > 1 {
		total *= 1.5
	}

	// Bonus: exact phrase appears in title.
	if len(tokens) > 1 && strings.Contains(strings.ToLower(e.Title), queryLower) {
		total *= 2
	}

	return int(total * 100)
}

// wordSet tokenizes a string and returns the unique words as a set.
func wordSet(s string) map[string]bool {
	words := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}

// tokenize splits a query into lowercase tokens, dropping short words.
func tokenize(s string) []string {
	words := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	var tokens []string
	for _, w := range words {
		if len(w) >= 2 {
			tokens = append(tokens, w)
		}
	}
	return tokens
}

type scored struct {
	entry Entry
	score int
}

// containsFold reports whether s contains substr, case-insensitive.
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// sortScored sorts by score descending. Uses insertion sort — result sets are
// small (thousands of entries, capped by limit).
func sortScored(s []scored) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].score > s[j-1].score; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// buildIDF computes inverse document frequency for all words across entries.
// Called once at index load time.
func buildIDF(entries []Entry) map[string]float64 {
	docCount := float64(len(entries))
	df := make(map[string]int) // document frequency per term

	for _, e := range entries {
		seen := make(map[string]bool)
		for _, w := range tokenize(e.Title + " " + e.Description) {
			if !seen[w] {
				df[w]++
				seen[w] = true
			}
		}
	}

	idf := make(map[string]float64, len(df))
	for term, count := range df {
		idf[term] = math.Log(docCount / float64(count))
	}
	return idf
}
