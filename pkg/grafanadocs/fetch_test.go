// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package grafanadocs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
