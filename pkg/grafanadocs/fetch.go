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

// overrideAllowlistCheck is a test hook. When non-nil, it replaces checkAllowlist
// in FetchDoc. Never set outside of tests.
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
	return nil
}

func (rl *rateLimiter) release() {
	rl.mu.Lock()
	rl.lastCall = time.Now()
	rl.mu.Unlock()
	<-rl.sem
}

// Doc holds fetched and cleaned documentation content.
type Doc struct {
	URL     string
	Content []byte
	Lines   []string // content split by newline, computed lazily
}

// EnsureLines splits Content into lines if not already done.
func (d *Doc) EnsureLines() {
	if d.Lines == nil {
		d.Lines = strings.Split(string(d.Content), "\n")
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

	fetchURL := rawURL
	if !strings.HasSuffix(fetchURL, ".md") {
		fetchURL += ".md"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("grafanadocs: build request: %w", err)
	}
	req.Header.Set("Accept", "text/markdown")

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
