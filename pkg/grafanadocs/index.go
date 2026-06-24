// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const maxIndexBytes = 10 * 1024 * 1024 // 10 MiB cap for the index file

var indexClient = &http.Client{Timeout: 60 * time.Second}

const DefaultIndexURL = "https://grafana.com/llms-full.txt"

// productHeader matches lines like "## Grafana Tempo documentation"
var productHeader = regexp.MustCompile(`^## (.+)$`)

// skipProducts are index sections that aren't real products.
var skipProducts = map[string]bool{
	"Documentation home": true,
	"Copyright notice":   true,
}

// entryLine matches lines like "- [Title](https://grafana.com/docs/...md): description"
// The description after the colon is optional.
var entryLine = regexp.MustCompile(`^- \[([^\]]+)\]\(([^)]+)\)(?::\s*(.*))?$`)

// LoadIndex fetches and parses the documentation index from the given URL.
// Only https URLs are accepted to prevent local file reads via file://.
func LoadIndex(ctx context.Context, rawURL string) (*Index, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: invalid index URL: %w", err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("grafanadocs: rejected index URL scheme %q (only https allowed)", u.Scheme)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: build request: %w", err)
	}
	req.Header.Set("User-Agent", "hack-doc-server/0.1")
	resp, err := indexClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: fetch index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grafanadocs: index returned %d", resp.StatusCode)
	}
	return LoadIndexFromReader(io.LimitReader(resp.Body, maxIndexBytes))
}

// LoadIndexFromReader parses the documentation index from an io.Reader.
// Use this with a local file or embedded fixture for testing.
func LoadIndexFromReader(r io.Reader) (*Index, error) {
	var (
		entries  []Entry
		products []Product
		current  string
		count    int
	)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1 MiB max line
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			line = strings.TrimPrefix(line, "\ufeff") // strip UTF-8 BOM
			first = false
		}

		if m := productHeader.FindStringSubmatch(line); m != nil {
			if current != "" && !skipProducts[current] {
				products = append(products, Product{Name: current, Count: count})
			}
			current = strings.TrimSuffix(m[1], " documentation")
			current = strings.TrimSuffix(current, " Documentation")
			count = 0
			continue
		}

		if m := entryLine.FindStringSubmatch(line); m != nil {
			if current == "" || skipProducts[current] {
				continue
			}
			entryURL := m[2]
			if !strings.HasPrefix(entryURL, "https://grafana.com/") {
				continue
			}
			entries = append(entries, Entry{
				Title:       m[1],
				URL:         entryURL,
				Description: m[3],
				Product:     current,
			})
			count++
		}
	}
	if current != "" && !skipProducts[current] {
		products = append(products, Product{Name: current, Count: count})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("grafanadocs: scan index: %w", err)
	}

	idx := &Index{Entries: entries, products: products}
	idx.idf = buildIDF(entries)
	return idx, nil
}
