# SPECS — mcp-doc-server

Source of truth for the behavior of `mcp-doc-server`: a Go MCP server that gives
AI agents version-aware access to Grafana Labs product documentation.

This is a **living document**. It captures what we have actually decided so far and
explicitly marks what is still open. It grows as we make decisions — we do not invent
contracts we have not agreed on. The goal is that, over time, `SPECS.md` + `NOTES.md` +
`TESTS.md` + `BENCHMARKS.md` become complete enough to wipe the code and rebuild from
specs alone. Until then, **Open questions** below tracks what is undecided.

Invariants are precise claims we have committed to; never remove or weaken one (add or
annotate instead). See `AGENTS.md` for the SDD convention and `NOTES.md` for decision history.

## Decided

### Purpose & scope
- A docs-retrieval layer for AI agents that complements
  [mcp-grafana](https://github.com/grafana/mcp-grafana) and
  [gcx](https://github.com/grafana/gcx) (which act on a live instance but expose no docs).
- **v1 serves `latest` docs only.** Versioned and tip-of-main docs are deferred
  (`NOTES.md` 2, 8).

### Sources
- **Index:** `https://grafana.com/llms-full.txt` — an official machine-readable index.
  Format: `## <Product> [Dd]ocumentation` headers grouping bullet entries of the form
  `- [Title](https://grafana.com/....md): description`. Parse rules:
  - Product name = header text with trailing " documentation"/" Documentation" stripped.
  - Entry regex: `^- \[([^\]]+)\]\(([^)]+)\)(?::\s*(.*))?$`
  - Entries with non-`https://grafana.com/` URLs are silently dropped (I11).
  - "Documentation home" and "Copyright notice" sections are excluded (I12).
- **Fetch:** the `.md` trick — appending `.md` to a docs URL path returns
  `text/markdown`. The suffix is applied to the parsed path component (not raw string)
  so fragments and query parameters are preserved (I16).

### Architecture: reusable core + thin adapters (`NOTES.md` 3, 10, 11)
- A **public, dependency-light core package** (`pkg/grafanadocs`) holds the retrieval
  logic and plain Go types — the primary integration surface for all consumers. The core
  must NOT live under `internal/` (Go forbids cross-module `internal` imports) and must
  not depend on the MCP SDK or cobra.
- **Opt-in adapters wrap the core** for our standalone server:
  - an **MCP adapter** (`pkg/grafanadocs/mcp`) on `github.com/mark3labs/mcp-go` v0.55.0;
  - a **cobra adapter** (`pkg/grafanadocs/cli`) for standalone CLI use.
- **Consumer integration model (`NOTES.md` 11):** `grafana/mcp-grafana` and `grafana/gcx`
  import the **core only** and write their own idiomatic wrappers — mcp-grafana writes a
  `tools/docs.go` using `MustTool`; gcx writes `cmd/gcx/docs/` using `output.Options`.
  The adapters serve our standalone server and generic consumers, not the two primary
  integration targets.
- Our standalone **MCP server** (stdio) is one consumer of the MCP adapter.

### Tools (snake_case, matching mcp-grafana convention)
- `search_docs` — find relevant docs (returns title, url, description, product).
- `get_doc` — fetch a doc's markdown; supports section-by-heading and offset/limit paging.
- `get_doc_outline` — cheap heading outline so an agent can target a section.
- `list_products` — the product groups from the index.

### Tool schemas (initial — may evolve during implementation)
- **`search_docs`:** Input `{query: string, product?: string, limit?: int (default 5)}`.
  Output: list of `{title, url, description, product}`.
- **`get_doc`:** Input `{url: string, section?: string, offset?: int, limit?: int}`.
  Output: `{content: string, url: string, total_lines: int, returned_range: [start, end]}`.
- **`get_doc_outline`:** Input `{url: string}`.
  Output: `{url: string, headings: [{level: int, text: string, line: int}]}`.
- **`list_products`:** Input `{}`.
  Output: `{products: [{name: string, count: int}]}`.

### Core API surface (`pkg/grafanadocs`)
Exported functions — plain Go, zero framework deps (stdlib + `net/http`):
- `LoadIndex(ctx context.Context, url string) (*Index, error)`
- `LoadIndexFromReader(r io.Reader) (*Index, error)`
- `Search(idx *Index, query string, opts SearchOpts) []Entry`
- `(*Index).Products() []Product`
- `(*Index).EntryCount() int`
- `FetchDoc(ctx context.Context, url string) (*Doc, error)`
- `Cleanup(raw []byte) []byte`
- `Outline(doc *Doc) []Heading`
- `Excerpt(doc *Doc, opts ExcerptOpts) ExcerptResult`

### Core data types
Plain Go structs with no `json` tags — the core is serialization-agnostic by design.
Consumers that need JSON output (gcx, MCP adapter) write thin wrapper types with their
own tags. This keeps the core free of framework opinions.

- `Entry{Title string, URL string, Description string, Product string}` — a single
  documentation page from the index.
- `Product{Name string, Count int}` — a documentation group with its entry count.
- `Index{Entries []Entry, products []Product, idf map[string]float64}` — the parsed
  documentation catalog; safe for concurrent read access after construction. `products`
  and `idf` are unexported; accessed via `Products()` and used internally by `Search`.
- `Doc{URL string, Content []byte, Lines []string}` — a fetched and cleaned documentation
  page. `Lines` is computed lazily via `EnsureLines()` and automatically recomputed if
  `Content` has changed since the last call.
- `Heading{Level int, Text string, Line int}` — a markdown heading with its 1-indexed
  line position.
- `SearchOpts{Product string, Limit int}` — controls search behavior. `Product` filters
  to a specific product (empty = all); `Limit` caps results (0 = default 5).
- `ExcerptOpts{Section string, Offset int, Limit int}` — controls bounded retrieval.
  `Section` extracts by heading text; when empty, `Offset`/`Limit` do line-based paging.
- `ExcerptResult{Content string, Start int, End int, Total int}` — the excerpted content
  with 1-indexed position metadata.

### Config & runtime
- **Go version:** 1.26.5 (matches `go.mod`).
- **`mcp-go` version:** v0.55.0 (standalone server uses latest; core has no mcp-go dep).
- **Module path:** `github.com/grafana/mcp-doc-server`.
- **Transport:** stdio in v1.
- **License:** Apache 2.0 (matches mcp-grafana and gcx).
- **HTTP timeouts:** 30s for doc fetches, 60s for index load.
- **Scanner buffer:** Index parser uses a 1 MiB max line buffer (vs. `bufio`'s 64 KiB
  default) to handle entries with long descriptions without silent truncation.

## Invariants (committed)

- **I1 — Index stays server-side.** The parsed index is never returned to the model; only
  `search_docs` result entries are.
- **I2 — Zero added inference cost (v1).** No server-side LLM or embedding calls;
  retrieval is deterministic.
- **I3 — Fetch allowlist.** `get_doc`/`get_doc_outline` fetch only canonical `grafana.com`
  docs URLs; other URLs are rejected before any network call. Predicate:
  `scheme == "https" && host == "grafana.com" && path starts with "/docs/"`.
  Redirects to non-grafana hosts are also blocked.
- **I4 — Bounded responses.** `get_doc` does not return an oversized full page by default;
  it returns a bounded slice and reports the total size so the caller can read more.
- **I5 — Citations.** Every response containing doc content includes its canonical source URL.
- **I6 — Latest only (v1).** All served URLs resolve to the `latest` channel; a `version`
  parameter is reserved but inert.
- **I7 — Lean search.** `search_docs` defaults to a small result count and concise output.
- **I8 — Cleanup preserves meaning.** Any markdown cleanup removes presentation/boilerplate
  only; documented content is unchanged. Blank-line collapsing runs *before* code blocks are
  restored, so intentional whitespace inside code blocks is preserved. Frontmatter, shortcode,
  and HTML comment stripping run together in a convergence loop so that removing one cannot
  reveal another (e.g. `{<!---->{<>}}` → `{{<>}}` → empty; or an HTML comment hiding a
  frontmatter block). Input line endings are normalized (CRLF/CR → LF) and the leading/trailing
  edge is trimmed up front so detection is pass-stable. `Cleanup` is idempotent:
  `Cleanup(Cleanup(x)) == Cleanup(x)` for all inputs.
- **I9 — Rate-limited outbound.** `FetchDoc` enforces a concurrency cap (5) and minimum
  gap (200ms) between requests to prevent abuse of grafana.com. Each `acquire` reserves a
  unique slot under the lock (advancing a `nextAllowed` cursor by the gap) and waits for it
  outside the lock, so N concurrent acquirers are spaced by the gap rather than all reading
  the same timestamp and proceeding together.
- **I10 — Body size caps.** Doc fetches are limited to 2 MiB; index fetches to 10 MiB.
  Prevents OOM from unexpected upstream responses.
- **I11 — Index entry validation.** URLs parsed from the index must start with
  `https://grafana.com/`; entries with other URLs are silently dropped at parse time.
- **I12 — Non-product sections excluded.** "Documentation home" and "Copyright notice"
  sections from the index are not exposed as products or searchable entries.
- **I13 — Actionable empty results.** When a tool returns no results, the response includes
  guidance on what to try next (e.g., "use list_products", "use get_doc_outline").
- **I14 — Code-fence-aware processing.** `Outline`, `excerptBySection`, and `Cleanup`
  all respect fenced code block boundaries via a shared `fenceInfo`/`fenceTracker`. Fences
  are matched per CommonMark: the marker is a run of 3+ backticks or tildes (indented ≤3
  spaces); a closing fence must use the same character, be at least as long as the opening
  fence, and carry no info string. This means a longer fence (e.g. ` ```` `) correctly
  contains shorter fences (` ``` `) as literal content. Content inside code blocks is never
  modified: comment lines (e.g. `# comment`) must not be misidentified as headings, and HTML
  comments / Hugo shortcodes inside code blocks must not be stripped.
- **I15 — Snake_case JSON keys.** All JSON output from the MCP adapter uses `snake_case`
  keys (e.g. `"title"`, `"url"`, `"level"`, `"name"`). The core types are intentionally
  tag-free; the MCP adapter owns the wire format via wrapper types with explicit `json` tags.
- **I16 — URL-safe `.md` suffix.** The `.md` suffix logic operates on the parsed URL path,
  not the raw URL string. Fragments (`#section`), query parameters (`?v=1`), and trailing
  slashes are handled correctly — the suffix is appended to the path only.
- **I17 — MCP numeric input validation.** MCP handlers reject `NaN`, `Inf`, negative, and
  out-of-range values for numeric parameters (`limit`, `offset`) with a descriptive error
  before passing them to the core. The `safeInt` helper converts `float64` → `int` only
  within `[0, 2^31]`; values above the cap are rejected so an out-of-range conversion (which
  is implementation-dependent in Go and can wrap to a negative int) cannot bypass validation.
- **I18 — Product filter resolution precedence.** The `product` parameter on `search_docs`
  resolves to canonical product names by trying match levels in order and stopping at the
  first that yields any match: exact (case-insensitive) → prefix → substring. A precise name
  selects exactly one product; a loose term still resolves (`"agent"` → `"Grafana Agent"`,
  `"loki"` → `"Grafana Loki"`). When an exact name also exists (`"grafana"` matching a bare
  `"Grafana"` product), the exact level wins and broader prefix/substring matches are not
  added. A filter that matches no product yields zero results.
  *Addendum (2026-06-23): supersedes the earlier exact-only rule (`strings.EqualFold`); see
  NOTE 28.*
- **I19 — Index scheme validation.** `LoadIndex` only accepts `https` URLs. `file://`,
  `http://`, and other schemes are rejected before any network call.
- **I20 — No orphan entries.** Entries appearing in the index before any `## Product`
  header are silently dropped. Every indexed entry has a non-empty `Product`.
- **I21 — Double allowlist check.** `FetchDoc` checks the allowlist on both the original
  URL and the URL after `.md` suffix transformation, preventing path manipulation through
  the transform.
- **I22 — User-Agent header.** All outbound HTTP requests include a `User-Agent:
  mcp-doc-server/0.1` header to identify the client to CDNs and WAFs.
- **I23 — Unclosed fence safety.** If a fenced code block is never closed (e.g. from
  body-size truncation), `extractCodeBlocks` treats the partial block as a protected code
  block in its original position rather than re-ordering it to the end.
- **I24 — ATX heading semantics.** `parseHeading` follows CommonMark for ATX headings:
  (a) a line indented 4+ spaces is an indented code block, not a heading; (b) an optional
  trailing run of `#`s preceded by whitespace is a closing sequence and is stripped from the
  text (`## Storage ##` → `Storage`), while a `#` inside a word or not preceded by a space is
  preserved (`C#`, `foo#`); (c) a heading whose text is empty after stripping is not a heading.
  Setext headings (underline `===`/`---`) are intentionally not recognized.
- **I25 — BOM tolerance.** A leading UTF-8 byte-order mark (`\ufeff`) on the index is stripped
  before parsing so the first `## Product` header still matches.
- **I26 — Index-need gating.** `cli.NeedsIndex(args)` reports whether an invocation requires a
  loaded index. Only index-reading subcommands (`search`, `products`) return true; `get` and
  `outline` (FetchDoc-only), help, completion, and bare invocations return false, so they work
  offline. The allow-list is owned by the `cli` package alongside the command definitions and
  must stay in sync with the wired subcommands (enforced by a drift test).

### Search ranking (decided)
- **Word-boundary matching:** tokens match against whole words (not substrings). "rate"
  matches an entry with "rate" in title but not "migrate".
- **TF-IDF weighting:** IDF computed once at index load time; rare terms score higher than
  ubiquitous ones (e.g., "clustering" > "grafana").
- **Title 3x weight:** title word matches contribute 3× their IDF weight vs. 1× for
  description matches.
- **All-tokens bonus (1.5×):** entries matching every query token get a 1.5× multiplier.
- **Exact phrase bonus (2×):** if the full query appears verbatim in the title (case-insensitive),
  score is doubled.
- **Deterministic:** no randomness, no network calls. Same index + same query = same results.

## Open questions (to decide as we go)

- **Caching:** what is cached and for how long (index TTL, fetched-page TTL).

### Closed questions

- **CLI consumer integration form** *(closed 2026-06-23, `NOTES.md` 18):* subcommand
  group (`<host> docs search/get/outline/products`). The host CLI imports the core only
  and writes its own output/annotation layer, as predicted by `NOTES.md` entry 11.
- **Index lifecycle in consumers** *(closed 2026-06-23, `NOTES.md` 18):* lazy `sync.Once`
  on first subcommand that needs the index (`search`, `products`). Commands that only
  need `FetchDoc` (`get`, `outline`) never trigger the load. The standalone CLI
  (`cmd/docs`) loads on demand at startup. Background refresh is deferred to the caching
  question.
- **MCP consumer integration form** *(closed 2026-06-23, `NOTES.md` 19):* the host MCP
  server registers four tools (`search_docs`, `get_doc`, `get_doc_outline`,
  `list_products`) using its own tool-registration helpers, and imports the core only
  (`pkg/grafanadocs`); the MCP adapter is not used. Index loaded lazily via `sync.Once`
  on first `search_docs`/`list_products` call. `get_doc` and `get_doc_outline` only call
  `FetchDoc`, so they never trigger the load. `DOCS_INDEX_URL` env var override mirrors
  the standalone server. Tool handlers should not use the default `slog` logger; let
  errors propagate and leave logging to the host.

### CLI adapter surface (`pkg/grafanadocs/cli`)
A mountable cobra `docs` command group, exported as `Command(idx *grafanadocs.Index)
*cobra.Command`. The caller owns the index lifecycle; the adapter is stateless and
imports only `cobra`/`pflag` and the core. Command grammar:
- `docs search <query> [--product=...] [--limit=5] [-o text|json|yaml|agents]`
- `docs get <url> [--section=...] [--offset=0] [--limit=0] [-o ...]`
- `docs outline <url> [-o ...]`
- `docs products [-o ...]`

Output formats: `text` (default, aligned table / raw markdown for `get`), `json`
(indented), `agents` (compact JSON), `yaml`. JSON/YAML keys mirror the MCP adapter so
both surfaces are consistent. Empty `search` results print guidance to stderr (I13);
stdout stays a clean, parseable result set. A host CLI mounts this by swapping the
local `output` helper for its own output system and registering its own agent
annotations.

## Implemented packages

- **`pkg/grafanadocs`** (public core): index load/parse, search (TF-IDF + word boundary),
  product list, fetch (allowlisted, rate-limited), cleanup, outline, excerpt — plain Go
  types, no MCP/CLI deps.
- **`pkg/grafanadocs/mcp`** (MCP adapter): exposes tools on `mark3labs/mcp-go`, handles
  validation errors and empty-result guidance.
- **`pkg/grafanadocs/cli`** (cobra adapter): mountable `docs` command group for
  standalone or embedded CLI use; opts + custom text codecs + structured
  (json/yaml/agents) output.
- **`cmd/mcp-doc-server`**: standalone MCP server (stdio, loads index at startup).
- **`cmd/docs`**: standalone CLI running the cobra adapter (loads index on demand;
  help/completion skip the fetch).

## Research & prior art

Supporting research is in the [`docs/design/research/`](docs/design/research/) folder:

- [`docs-mcp-server-use-cases.md`](docs/design/research/docs-mcp-server-use-cases.md) — Recurring
  use-case patterns for docs MCP servers (grounding, version-aware lookups, citations,
  token-efficient retrieval, product discovery, deterministic retrieval in agentic
  workflows, onboarding/troubleshooting assistants, server-side index isolation) mapped to
  mcp-doc-server invariants and tools.

## Reproducibility (the goal)

Specs are the durable artifact; code is disposable. As **Open questions** close, this file
becomes sufficient to delete the `.go` code and rebuild faithfully by implementing until
every contract/invariant holds and every `TESTS.md` scenario passes. We only write a
contract here once we have actually decided it.
