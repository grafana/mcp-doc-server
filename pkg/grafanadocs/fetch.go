// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const maxBodyBytes = 2 * 1024 * 1024 // 2 MiB — no doc page should exceed this

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if req.URL.Host != "grafana.com" {
			return fmt.Errorf("grafanadocs: redirect to %q blocked", req.URL.Host)
		}
		return nil
	},
}

// overrideAllowlistCheck is a test hook; never set outside of tests.
var overrideAllowlistCheck func(string) error

// fetchLimiter prevents excessive outbound requests. Allows a burst of 5
// concurrent fetches with a minimum 200ms gap between requests.
var fetchLimiter = newRateLimiter(5, 200*time.Millisecond)

type rateLimiter struct {
	mu       sync.Mutex
	lastCall time.Time
	sem      chan struct{}
	minGap   time.Duration
}

func newRateLimiter(maxConcurrent int, minGap time.Duration) *rateLimiter {
	return &rateLimiter{
		sem:    make(chan struct{}, maxConcurrent),
		minGap: minGap,
	}
}

func (rl *rateLimiter) acquire(ctx context.Context) error {
	select {
	case rl.sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	rl.mu.Lock()
	elapsed := time.Since(rl.lastCall)
	if wait := rl.minGap - elapsed; wait > 0 {
		rl.mu.Unlock()
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			<-rl.sem
			return ctx.Err()
		}
	} else {
		rl.mu.Unlock()
	}

	rl.mu.Lock()
	rl.lastCall = time.Now()
	rl.mu.Unlock()
	return nil
}

func (rl *rateLimiter) release() {
	<-rl.sem
}

// Doc holds fetched and cleaned documentation content.
type Doc struct {
	URL     string
	Content []byte
	Lines   []string // content split by newline, computed lazily

	linesContent string // snapshot of Content when Lines was computed
}

// EnsureLines splits Content into lines if not already done, or recomputes
// if Content has changed since the last call.
func (d *Doc) EnsureLines() {
	current := string(d.Content)
	if d.Lines == nil || d.linesContent != current {
		d.Lines = strings.Split(current, "\n")
		d.linesContent = current
	}
}

// FetchDoc retrieves a documentation page. The URL must pass the allowlist
// check (invariant I3) — only grafana.com /docs/ URLs are permitted.
// Outbound requests are rate-limited to prevent abuse.
func FetchDoc(ctx context.Context, rawURL string) (*Doc, error) {
	check := checkAllowlist
	if overrideAllowlistCheck != nil {
		check = overrideAllowlistCheck
	}
	if err := check(rawURL); err != nil {
		return nil, err
	}

	if err := fetchLimiter.acquire(ctx); err != nil {
		return nil, fmt.Errorf("grafanadocs: rate limit: %w", err)
	}
	defer fetchLimiter.release()

	fetchURL, err := ensureMDSuffix(rawURL)
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: invalid URL: %w", err)
	}
	if err := check(fetchURL); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: build request: %w", err)
	}
	req.Header.Set("Accept", "text/markdown")
	req.Header.Set("User-Agent", "hack-doc-server/0.1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: fetch doc: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grafanadocs: doc returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: read body: %w", err)
	}

	cleaned := Cleanup(body)
	return &Doc{URL: rawURL, Content: cleaned}, nil
}

// ensureMDSuffix appends .md to the URL path if it doesn't already end with it.
// Operates on the parsed path component so fragments and query params are preserved.
func ensureMDSuffix(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	p := strings.TrimRight(u.Path, "/")
	if p == "" {
		p = u.Path
	}
	if !strings.HasSuffix(p, ".md") {
		p += ".md"
	}
	u.Path = p
	return u.String(), nil
}

// checkAllowlist rejects URLs that are not canonical grafana.com docs pages.
// This is the SSRF guard (invariant I3).
func checkAllowlist(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("grafanadocs: invalid URL: %w", err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("grafanadocs: rejected scheme %q (only https allowed)", u.Scheme)
	}
	if u.Host != "grafana.com" {
		return fmt.Errorf("grafanadocs: rejected host %q (only grafana.com allowed)", u.Host)
	}
	if !strings.HasPrefix(u.Path, "/docs/") && u.Path != "/docs" {
		return fmt.Errorf("grafanadocs: rejected path %q (must be under /docs/)", u.Path)
	}

	return nil
}
