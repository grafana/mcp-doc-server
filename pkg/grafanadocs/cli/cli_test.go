// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/grafana/mcp-doc-server/pkg/grafanadocs"
	"github.com/stretchr/testify/require"
)

func loadIndex(t *testing.T) *grafanadocs.Index {
	t.Helper()
	f, err := os.Open("../testdata/llms-full-sample.txt")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	idx, err := grafanadocs.LoadIndexFromReader(f)
	require.NoError(t, err)
	return idx
}

// run executes the docs command group with the given args, capturing stdout and
// stderr separately so tests can assert on machine output and guidance apart.
func run(t *testing.T, idx *grafanadocs.Index, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := Command(idx)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestSearchCommand(t *testing.T) {
	idx := loadIndex(t)

	tests := []struct {
		name        string
		args        []string
		wantErr     string
		wantStdout  []string
		wantStderr  string
		checkStdout func(t *testing.T, stdout string)
	}{
		{
			name:       "text output has header and a hit",
			args:       []string{"search", "clustering"},
			wantStdout: []string{"TITLE", "PRODUCT", "URL", "Clustering"},
		},
		{
			name:       "product filter is case-insensitive exact match",
			args:       []string{"search", "clustering", "--product", "grafana agent"},
			wantStdout: []string{"Clustering"},
		},
		{
			name:    "empty query is rejected",
			args:    []string{"search", ""},
			wantErr: "query is required",
		},
		{
			name:       "no matches still emits a clean table and guidance",
			args:       []string{"search", "zzzznotathing"},
			wantStdout: []string{"TITLE"},
			wantStderr: "No results found",
		},
		{
			name: "json output is valid and structured",
			args: []string{"search", "clustering", "-o", "json"},
			checkStdout: func(t *testing.T, stdout string) {
				var entries []grafanadocs.Entry
				require.NoError(t, json.Unmarshal([]byte(stdout), &entries))
				require.NotEmpty(t, entries)
				require.True(t, strings.Count(stdout, "\n") > 1, "json should be indented")
			},
		},
		{
			name: "agents output is compact json",
			args: []string{"search", "clustering", "-o", "agents"},
			checkStdout: func(t *testing.T, stdout string) {
				var entries []grafanadocs.Entry
				require.NoError(t, json.Unmarshal([]byte(stdout), &entries))
				require.Equal(t, 1, strings.Count(stdout, "\n"), "agents output should be a single compact line")
			},
		},
		{
			name:    "unknown format is rejected",
			args:    []string{"search", "clustering", "-o", "xml"},
			wantErr: "unknown output format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := run(t, idx, tt.args...)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			for _, want := range tt.wantStdout {
				require.Contains(t, stdout, want)
			}
			if tt.wantStderr != "" {
				require.Contains(t, stderr, tt.wantStderr)
			}
			if tt.checkStdout != nil {
				tt.checkStdout(t, stdout)
			}
		})
	}
}

func TestProductsCommand(t *testing.T) {
	idx := loadIndex(t)

	t.Run("text lists products and counts", func(t *testing.T) {
		stdout, _, err := run(t, idx, "products")
		require.NoError(t, err)
		require.Contains(t, stdout, "PRODUCT")
		require.Contains(t, stdout, "COUNT")
		require.Contains(t, stdout, "Grafana Agent")
		// Non-product sections must be excluded (invariant I12).
		require.NotContains(t, stdout, "Documentation home")
		require.NotContains(t, stdout, "Copyright notice")
	})

	t.Run("json wraps products", func(t *testing.T) {
		stdout, _, err := run(t, idx, "products", "-o", "json")
		require.NoError(t, err)
		var got productsResult
		require.NoError(t, json.Unmarshal([]byte(stdout), &got))
		require.NotEmpty(t, got.Products)
	})
}

// get and outline reach the network; tests exercise wiring up to the allowlist
// guard (invariant I3) without making real requests.
func TestGetCommandGuards(t *testing.T) {
	idx := loadIndex(t)

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "empty url", args: []string{"get", ""}, wantErr: "url is required"},
		{name: "non-grafana host", args: []string{"get", "https://evil.com/docs/x.md"}, wantErr: "rejected host"},
		{name: "outline empty url", args: []string{"outline", ""}, wantErr: "url is required"},
		{name: "outline non-grafana host", args: []string{"outline", "https://evil.com/docs/x"}, wantErr: "rejected host"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := run(t, idx, tt.args...)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCodecs(t *testing.T) {
	t.Run("search table renders rows", func(t *testing.T) {
		var buf bytes.Buffer
		err := searchTableCodec{}.Encode(&buf, []grafanadocs.Entry{
			{Title: "Clustering", Product: "Agent", URL: "https://grafana.com/docs/agent/x.md"},
		})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "Clustering")
		require.Contains(t, buf.String(), "Agent")
	})

	t.Run("get text renders raw content", func(t *testing.T) {
		var buf bytes.Buffer
		err := getTextCodec{}.Encode(&buf, getResult{Content: "# Title\n\nbody"})
		require.NoError(t, err)
		require.Equal(t, "# Title\n\nbody\n", buf.String())
	})

	t.Run("outline table renders headings", func(t *testing.T) {
		var buf bytes.Buffer
		err := outlineTableCodec{}.Encode(&buf, outlineResult{
			Headings: []grafanadocs.Heading{{Level: 2, Text: "Setup", Line: 5}},
		})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "Setup")
		require.Contains(t, buf.String(), "LVL")
	})

	t.Run("products table renders rows", func(t *testing.T) {
		var buf bytes.Buffer
		err := productsTableCodec{}.Encode(&buf, productsResult{
			Products: []grafanadocs.Product{{Name: "Tempo", Count: 12}},
		})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "Tempo")
		require.Contains(t, buf.String(), "12")
	})

	t.Run("codecs reject wrong types", func(t *testing.T) {
		var buf bytes.Buffer
		require.Error(t, searchTableCodec{}.Encode(&buf, "nope"))
		require.Error(t, getTextCodec{}.Encode(&buf, "nope"))
		require.Error(t, outlineTableCodec{}.Encode(&buf, "nope"))
		require.Error(t, productsTableCodec{}.Encode(&buf, "nope"))
	})
}

func TestNeedsIndex(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"search reads index", []string{"search", "clustering"}, true},
		{"products reads index", []string{"products"}, true},
		{"get does not", []string{"get", "https://grafana.com/docs/x"}, false},
		{"outline does not", []string{"outline", "https://grafana.com/docs/x"}, false},
		{"bare invocation", nil, false},
		{"help flag", []string{"--help"}, false},
		{"completion", []string{"completion", "bash"}, false},
		{"unknown subcommand", []string{"frobnicate"}, false},
		{"leading flag skipped", []string{"-v", "search"}, true},
		{"leading flag before non-index cmd", []string{"-v", "get", "x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, NeedsIndex(tt.args))
		})
	}
}

// TestNeedsIndexInSyncWithCommands guards against drift: every command named in
// indexReadingCommands must be a real subcommand of the docs group.
func TestNeedsIndexInSyncWithCommands(t *testing.T) {
	real := map[string]bool{}
	for _, c := range Command(&grafanadocs.Index{}).Commands() {
		real[c.Name()] = true
	}
	for name := range indexReadingCommands {
		require.True(t, real[name], "indexReadingCommands lists %q, which is not a real subcommand", name)
	}
}
