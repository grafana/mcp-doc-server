// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimiter_GapEnforced(t *testing.T) {
	gap := 50 * time.Millisecond
	tolerance := 5 * time.Millisecond
	rl := newRateLimiter(2, gap)
	ctx := context.Background()

	require.NoError(t, rl.acquire(ctx))
	rl.release()

	start := time.Now()
	require.NoError(t, rl.acquire(ctx))
	elapsed := time.Since(start)
	rl.release()

	require.GreaterOrEqual(t, elapsed, gap-tolerance, "second acquire should wait approximately the gap duration")
}

func TestEnsureLines_RecomputesAfterContentChange(t *testing.T) {
	doc := &Doc{URL: "test", Content: []byte("line1\nline2")}
	doc.EnsureLines()
	require.Equal(t, []string{"line1", "line2"}, doc.Lines)

	doc.Content = []byte("a\nb\nc")
	doc.EnsureLines()
	require.Equal(t, []string{"a", "b", "c"}, doc.Lines, "Lines must reflect updated Content")
}

func TestEnsureMDSuffix(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bare path", "https://grafana.com/docs/tempo/latest/configuration", "https://grafana.com/docs/tempo/latest/configuration.md"},
		{"already .md", "https://grafana.com/docs/tempo/latest/configuration.md", "https://grafana.com/docs/tempo/latest/configuration.md"},
		{"trailing slash", "https://grafana.com/docs/tempo/latest/", "https://grafana.com/docs/tempo/latest.md"},
		{"fragment preserved", "https://grafana.com/docs/tempo/latest/configuration#auth", "https://grafana.com/docs/tempo/latest/configuration.md#auth"},
		{"query params preserved", "https://grafana.com/docs/tempo/latest/configuration?v=1", "https://grafana.com/docs/tempo/latest/configuration.md?v=1"},
		{"fragment and query", "https://grafana.com/docs/tempo/latest/configuration?v=1#auth", "https://grafana.com/docs/tempo/latest/configuration.md?v=1#auth"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ensureMDSuffix(tt.in)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLiveFetch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live fetch tests (use -short=false to enable)")
	}

	ctx := context.Background()

	pages := []struct {
		name string
		url  string
	}{
		{"small page", "https://grafana.com/docs/grafana/latest/getting-started/"},
		{"large page", "https://grafana.com/docs/tempo/latest/configuration/"},
		{"alloy page", "https://grafana.com/docs/alloy/latest/get-started/run/linux/"},
		{"loki page", "https://grafana.com/docs/loki/latest/get-started/"},
	}

	for _, p := range pages {
		t.Run(p.name, func(t *testing.T) {
			doc, err := FetchDoc(ctx, p.url)
			require.NoError(t, err)
			require.NotEmpty(t, doc.Content)
			require.Equal(t, p.url, doc.URL, "URL should be the original, not the .md-suffixed version")

			doc.EnsureLines()
			require.Greater(t, len(doc.Lines), 1, "should have multiple lines")

			headings := Outline(doc)
			require.NotEmpty(t, headings, "should have at least one heading")

			for _, h := range headings {
				require.GreaterOrEqual(t, h.Level, 1)
				require.LessOrEqual(t, h.Level, 6)
				require.GreaterOrEqual(t, h.Line, 1)
				require.LessOrEqual(t, h.Line, len(doc.Lines))
				require.NotEmpty(t, h.Text)
			}

			result := Excerpt(doc, ExcerptOpts{Offset: 0, Limit: 20})
			require.NotEmpty(t, result.Content)
			require.Equal(t, 1, result.Start)
			require.LessOrEqual(t, result.End, 20)
			require.GreaterOrEqual(t, result.Total, result.End)

			content := string(doc.Content)
			require.NotContains(t, content, "{{<", "shortcodes should be stripped")
			require.NotContains(t, content, "{{%", "shortcodes should be stripped")
		})
	}
}

func TestLiveIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live index test (use -short=false to enable)")
	}

	ctx := context.Background()
	idx, err := LoadIndex(ctx, DefaultIndexURL)
	require.NoError(t, err)
	require.Greater(t, idx.EntryCount(), 1000, "live index should have >1000 entries")
	require.Greater(t, len(idx.Products()), 10, "live index should have >10 products")

	for _, e := range idx.Entries {
		require.NotEmpty(t, e.Title)
		require.NotEmpty(t, e.URL)
		require.NotEmpty(t, e.Product)
		require.True(t, strings.HasPrefix(e.URL, "https://grafana.com/"),
			"entry URL %q not from grafana.com", e.URL)
	}

	results := Search(idx, "tempo configuration", SearchOpts{Limit: 5})
	require.NotEmpty(t, results, "search should find tempo configuration docs")
}

func TestRateLimiter_ConcurrentStress(t *testing.T) {
	rl := newRateLimiter(3, 10*time.Millisecond)
	ctx := context.Background()

	const goroutines = 20
	const iterations = 10
	const total = goroutines * iterations
	const gap = 10 * time.Millisecond

	var wg sync.WaitGroup
	wg.Add(goroutines)
	errs := make(chan error, total)

	start := time.Now()
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if err := rl.acquire(ctx); err != nil {
					errs <- err
					return
				}
				rl.release()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("stress test timed out — possible deadlock")
	}
	elapsed := time.Since(start)

	close(errs)
	for err := range errs {
		t.Errorf("acquire error: %v", err)
	}

	// Spacing must hold under concurrency: N acquires spaced by gap take at
	// least (N-1)*gap. Without per-slot reservation, all concurrent acquirers
	// would proceed together and finish near-instantly.
	minExpected := time.Duration(total-1) * gap
	tolerance := 5 * time.Millisecond
	require.GreaterOrEqual(t, elapsed, minExpected-tolerance,
		"concurrent acquires must be spaced by minGap")
}

func FuzzEnsureMDSuffix(f *testing.F) {
	f.Add("https://grafana.com/docs/tempo/latest/configuration")
	f.Add("https://grafana.com/docs/tempo/latest/configuration.md")
	f.Add("https://grafana.com/docs/tempo/latest/")
	f.Add("https://grafana.com/docs/tempo/latest/configuration#auth")
	f.Add("https://grafana.com/docs/tempo/latest/configuration?v=1")
	f.Add("")
	f.Add("not-a-url")

	f.Fuzz(func(t *testing.T, input string) {
		result, err := ensureMDSuffix(input)
		if err != nil {
			return
		}
		// Only assert .md for URLs that actually have a hierarchical path to
		// suffix. Opaque URIs (e.g. "a:0", "A:://") have no path component, so
		// ensureMDSuffix correctly leaves them unchanged.
		u, perr := url.Parse(input)
		if perr != nil || u.Opaque != "" || u.Path == "" {
			return
		}
		if !strings.Contains(result, ".md") {
			t.Errorf("result should contain .md: %q (from %q)", result, input)
		}
	})
}

func TestCheckAllowlist(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid
		{"grafana docs page", "https://grafana.com/docs/tempo/latest/configuration.md", false},
		{"docs root", "https://grafana.com/docs/", false},
		{"nested path", "https://grafana.com/docs/grafana/latest/dashboards/build-dashboards.md", false},

		// SSRF attempts
		{"metadata endpoint", "http://169.254.169.254/latest/meta-data/", true},
		{"private IP", "http://10.0.0.1/internal", true},
		{"wrong host", "https://evil.example/docs/page.md", true},
		{"non-grafana host", "https://attacker.grafana.com.evil.net/docs/x.md", true},
		{"http scheme", "http://grafana.com/docs/page.md", true},
		{"ftp scheme", "ftp://grafana.com/docs/page.md", true},
		{"wrong path", "https://grafana.com/blog/something.md", true},
		{"root path", "https://grafana.com/", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkAllowlist(tt.url)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFetchDoc(t *testing.T) {
	// Bypass the allowlist for httptest URLs (they use 127.0.0.1).
	origOverride := overrideAllowlistCheck
	overrideAllowlistCheck = func(string) error { return nil }
	t.Cleanup(func() { overrideAllowlistCheck = origOverride })

	t.Run("fetches and cleans markdown", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/markdown")
			_, _ = w.Write([]byte("---\ntitle: Test\n---\n\n{{< some-shortcode >}}\n\n# Hello World\n\nContent here.\n"))
		}))
		defer ts.Close()

		doc, err := FetchDoc(context.Background(), ts.URL+"/docs/test")
		require.NoError(t, err)
		require.Equal(t, ts.URL+"/docs/test", doc.URL)

		content := string(doc.Content)
		require.NotContains(t, content, "---")
		require.NotContains(t, content, "{{<")
		require.Contains(t, content, "# Hello World")
		require.Contains(t, content, "Content here.")
	})

	t.Run("appends .md suffix when missing", func(t *testing.T) {
		var requestedPath string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestedPath = r.URL.Path
			_, _ = w.Write([]byte("# Doc\n"))
		}))
		defer ts.Close()

		_, err := FetchDoc(context.Background(), ts.URL+"/docs/page")
		require.NoError(t, err)
		require.Equal(t, "/docs/page.md", requestedPath)
	})

	t.Run("does not double .md suffix", func(t *testing.T) {
		var requestedPath string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestedPath = r.URL.Path
			_, _ = w.Write([]byte("# Doc\n"))
		}))
		defer ts.Close()

		_, err := FetchDoc(context.Background(), ts.URL+"/docs/page.md")
		require.NoError(t, err)
		require.Equal(t, "/docs/page.md", requestedPath)
	})

	t.Run("body size cap enforced", func(t *testing.T) {
		oversized := strings.Repeat("x", maxBodyBytes+1000)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(oversized))
		}))
		defer ts.Close()

		doc, err := FetchDoc(context.Background(), ts.URL+"/docs/big.md")
		require.NoError(t, err)
		// LimitReader caps at maxBodyBytes; Cleanup may add a trailing \n.
		// The key invariant: content is significantly smaller than what was served.
		require.Less(t, len(doc.Content), len(oversized))
	})

	t.Run("non-200 status returns error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		_, err := FetchDoc(context.Background(), ts.URL+"/docs/missing.md")
		require.Error(t, err)
		require.Contains(t, err.Error(), "404")
	})

	t.Run("redirect to non-grafana host blocked", func(t *testing.T) {
		evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("request reached evil server")
		}))
		defer evil.Close()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, evil.URL+"/steal", http.StatusFound)
		}))
		defer ts.Close()

		_, err := FetchDoc(context.Background(), ts.URL+"/docs/redir.md")
		require.Error(t, err)
		require.Contains(t, err.Error(), "blocked")
	})

	t.Run("context cancellation respected", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Hang until context is cancelled.
			<-r.Context().Done()
		}))
		defer ts.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		_, err := FetchDoc(ctx, ts.URL+"/docs/slow.md")
		require.Error(t, err)
	})
}
