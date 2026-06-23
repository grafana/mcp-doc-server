# SPECS — hack-doc-server

Source of truth for the behavior of `hack-doc-server`: a Go MCP server that gives
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
- **Fetch:** the `.md` trick — appending `.md` to a docs URL returns `text/markdown`.

### Architecture: reusable core + thin adapters (`NOTES.md` 3, 10, 11)
- A **public, dependency-light core package** (`pkg/grafanadocs`) holds the retrieval
  logic and plain Go types — the primary integration surface for all consumers. The core
  must NOT live under `internal/` (Go forbids cross-module `internal` imports) and must
  not depend on the MCP SDK or cobra.
- **Opt-in adapters wrap the core** for our standalone server:
  - an **MCP adapter** (`pkg/grafanadocs/mcp`) on `github.com/mark3labs/mcp-go` v0.46.0;
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
- `FetchDoc(ctx context.Context, url string) (*Doc, error)`
- `Cleanup(raw []byte) []byte`
- `Outline(doc *Doc) []Heading`
- `Excerpt(doc *Doc, opts ExcerptOpts) ExcerptResult`

### Entry data model
- `Entry{Title string, URL string, Description string, Product string}`

### Config & runtime
- **Go version:** 1.24+ (floor); mcp-grafana currently uses 1.26.3.
- **`mcp-go` version:** v0.55.0 (standalone server uses latest; core has no mcp-go dep).
- **Module path:** `github.com/grafana/hack-doc-server` (hackathon; eventual home TBD).
- **Transport:** stdio in v1.
- **License:** Apache 2.0 (matches mcp-grafana and gcx).
- **HTTP timeouts:** 30s for doc fetches, 60s for index load.

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
  only; documented content is unchanged.
- **I9 — Rate-limited outbound.** `FetchDoc` enforces a concurrency cap (5) and minimum
  gap (200ms) between requests to prevent abuse of grafana.com.
- **I10 — Body size caps.** Doc fetches are limited to 2 MiB; index fetches to 10 MiB.
  Prevents OOM from unexpected upstream responses.
- **I11 — Index entry validation.** URLs parsed from the index must start with
  `https://grafana.com/`; entries with other URLs are silently dropped at parse time.
- **I12 — Non-product sections excluded.** "Documentation home" and "Copyright notice"
  sections from the index are not exposed as products or searchable entries.
- **I13 — Actionable empty results.** When a tool returns no results, the response includes
  guidance on what to try next (e.g., "use list_products", "use get_doc_outline").

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
- **Index lifecycle in consumers:** singleton? per-request? background refresh?
- **gcx integration form:** subcommand vs. agent skill vs. both.

### CLI adapter surface (`pkg/grafanadocs/cli`)
A mountable cobra `docs` command group, exported as `Command(idx *grafanadocs.Index)
*cobra.Command`. The caller owns the index lifecycle; the adapter is stateless and
imports only `cobra`/`pflag` and the core (no gcx internals). Command grammar:
- `docs search <query> [--product=...] [--limit=5] [-o text|json|yaml|agents]`
- `docs get <url> [--section=...] [--offset=0] [--limit=0] [-o ...]`
- `docs outline <url> [-o ...]`
- `docs products [-o ...]`

Output formats: `text` (default, aligned table / raw markdown for `get`), `json`
(indented), `agents` (compact JSON), `yaml`. JSON/YAML keys mirror the MCP adapter so
both surfaces are consistent. Empty `search` results print guidance to stderr (I13);
stdout stays a clean, parseable result set. gcx mounts this by swapping the local
`output` helper for gcx's `output.Options` (package `internal/output`) and registering
its own agent annotations.

## Implemented packages

- **`pkg/grafanadocs`** (public core): index load/parse, search (TF-IDF + word boundary),
  product list, fetch (allowlisted, rate-limited), cleanup, outline, excerpt — plain Go
  types, no MCP/CLI deps.
- **`pkg/grafanadocs/mcp`** (MCP adapter): exposes tools on `mark3labs/mcp-go`, handles
  validation errors and empty-result guidance.
- **`pkg/grafanadocs/cli`** (cobra adapter): mountable `docs` command group for gcx /
  standalone CLI use; opts + custom text codecs + structured (json/yaml/agents) output.
- **`cmd/hack-doc-server`**: standalone MCP server (stdio, loads index at startup).
- **`cmd/docs`**: standalone CLI running the cobra adapter (loads index on demand;
  help/completion skip the fetch).

## Research & prior art

Supporting research is in the [`research/`](research/) folder:

- [`elastic-docs-mcp.md`](research/elastic-docs-mcp.md) — Elastic's hosted docs MCP
  server (`https://www.elastic.co/docs/_mcp/`): 6 tools including semantic search, AI
  summaries, and coherence/inconsistency checks. Comparison table against hack-doc-server
  invariants; ideas worth borrowing (`find_related_docs`, navigation context, hosted HTTP
  option) and deliberate differences (no server-side inference, bounded slices, version
  awareness).
- [`docs-mcp-server-use-cases.md`](research/docs-mcp-server-use-cases.md) — Recurring
  use-case patterns for docs MCP servers (grounding, version-aware lookups, citations,
  token-efficient retrieval, product discovery, deterministic retrieval in agentic
  workflows, onboarding/troubleshooting assistants, server-side index isolation) mapped to
  hack-doc-server invariants and tools.

## Reproducibility (the goal)

Specs are the durable artifact; code is disposable. As **Open questions** close, this file
becomes sufficient to delete the `.go` code and rebuild faithfully by implementing until
every contract/invariant holds and every `TESTS.md` scenario passes. We only write a
contract here once we have actually decided it.
