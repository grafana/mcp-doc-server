# NOTES — hack-doc-server

Append-only, dated design-decision log. Never delete or edit an existing entry; if a
decision is reversed, add an `*Addendum (YYYY-MM-DD):*` line to the original entry and
then add a new numbered entry.

## 1. Use `llms-full.txt` as the documentation index
*Added: 2026-06-22*
**Decision:** Build the searchable catalog by parsing `https://grafana.com/llms-full.txt`.
**Rationale:** It is an official, machine-readable index — 6,879 entries across 26
product groups, format `- [Title](url.md): description` under `## <Product>` headers,
all pointing at `latest`. No crawling or scraping needed.
**Consequence:** Search quality is bounded by the index's titles/descriptions; the
index is latest-only, which fixes v1 scope (see entry 2).

## 2. v1 serves `latest` only
*Added: 2026-06-22*
**Decision:** Ship latest-only first; defer versioned and tip-of-main docs.
**Rationale:** Most users want latest; `llms-full.txt` is latest-only; this removes the
hard per-version discovery problem from v1.
**Consequence:** A `version` parameter is reserved but inert (invariant I6). Versions
arrive later via URL/ref transforms without reworking the core.

## 3. Front end is an MCP server with a reusable core
*Added: 2026-06-22*
**Decision:** Expose functionality as a standalone MCP server; keep retrieval logic in
`internal/*` packages.
**Rationale:** Matches the agent ecosystem (drops into the same clients as mcp-grafana),
demos immediately, and the core can later be contributed into mcp-grafana or reused by gcx.
**Consequence:** Transport is stdio in v1; an HTTP front end is a later add over the same core.
*Addendum (2026-06-22):* The `internal/*` layout for the core is superseded by entry 10 —
the core is now a public package (`pkg/grafanadocs`) so that both `grafana/mcp-grafana` and
`grafana/gcx` can import it cross-module.

## 4. No server-side LLM or embedding calls in v1
*Added: 2026-06-22*
**Decision:** Retrieval is purely deterministic (token ranking + fetch).
**Rationale:** Avoid adding inference cost for customers; keep behavior reproducible.
**Consequence:** No semantic/RAG search in v1 (invariant I2). Semantic search is a later option.

## 5. Token cost is a first-class constraint
*Added: 2026-06-22*
**Decision:** `get_doc` is sectioned and paged; add `get_doc_outline`; clean markdown;
keep search output lean.
**Rationale:** A single page can be ~35K tokens (e.g. `tempo/latest/configuration.md` ~139 KB).
Returning targeted slices keeps customer token cost low.
**Consequence:** Invariants I1, I4, I7, I8. Net cheaper than browsing whole pages.

## 6. Fetch allowlist (SSRF guard)
*Added: 2026-06-22*
**Decision:** Only fetch `grafana.com` URLs under `/docs/`.
**Rationale:** Tools take a URL argument from an agent; an unconstrained fetcher is an SSRF risk.
**Consequence:** Invariant I3. Versioned/GitHub sources will extend the allowlist explicitly.

## 7. Adopt spec-driven development
*Added: 2026-06-22*
**Decision:** Follow the SDD convention from mattdurham/bob — `SPECS.md`, `NOTES.md`,
`TESTS.md`, `BENCHMARKS.md`, and the `// NOTE:` marker on spec-driven `.go` files.
**Rationale:** Keep contracts and decisions authoritative and cross-referenced with code.
**Consequence:** Each `internal/*` module gets its own spec suite as it is built; specs
are updated whenever the corresponding code changes.

## 8. Versioning model (deferred design)
*Added: 2026-06-22*
**Decision:** When versions land, support two sources: published grafana.com versioned
URLs (`/docs/<product>/<version>/...md`) and GitHub raw source docs at a ref (tip-of-main).
**Rationale:** Confirmed both work; published covers releases, GitHub covers unreleased.
**Consequence:** Needs a product→repo→docs-path map and per-source allowlist entries;
search remains latest-backed until a per-version discovery mechanism exists.

## 9. Specs are the regeneration source of truth
*Added: 2026-06-22*
**Decision:** Treat the spec suite as the durable artifact and the code as disposable.
`SPECS.md` is kept regeneration-grade (data shapes, index grammar, tool schemas, ranking,
outline/section and cleanup rules, config) so the code can be wiped and rebuilt faithfully
from specs + tests alone.
**Rationale:** The user wants to replicate the result consistently and retain what we
learned independent of any specific implementation.
**Consequence:** Behavior-changing code edits must update `SPECS.md`/`TESTS.md` in the same
change, or regeneration drifts. Specs stay behavior-level (pin contract + acceptance), not
line-by-line code, so any correct implementation passes.
*Addendum (2026-06-22):* Regeneration-grade is the **goal**, reached incrementally — not a
license to write contracts we have not decided. `SPECS.md` is a living document: it states
only what is decided and tracks the rest under **Open questions**, which close as we go.

## 10. Reusable core + thin adapters (usable by gcx AND mcp-grafana)
*Added: 2026-06-22*
**Decision:** Ship the retrieval logic as a **public, dependency-light Go package** (the
core) with thin, opt-in adapters: an **MCP adapter** (targets `github.com/mark3labs/mcp-go`,
the SDK mcp-grafana uses — observed v0.46.0) and a **cobra adapter** (a `docs` command,
optionally a gcx agent skill). Our standalone `hack-doc-server` and `grafana/mcp-grafana`
consume the MCP adapter; `grafana/gcx` consumes the cobra adapter.
**Rationale:** Confirmed mcp-grafana is built on mark3labs/mcp-go with `{Tool, Handler}`
tools, while gcx is a cobra CLI with no MCP dependency. A layered core lets both reuse the
same logic without one inheriting the other's dependencies.
**Consequence:**
- The core must NOT live under `internal/` (Go forbids cross-module `internal` imports);
  it goes in a public package (e.g. `pkg/grafanadocs`). Supersedes the all-`internal/`
  layout in earlier planning.
- The MCP SDK is confined to the MCP adapter so gcx does not transitively pull it.
- The MCP adapter's mcp-go version should track mcp-grafana to avoid conflicts on import.
- Open: eventual module path/ownership (ideally `github.com/grafana/...`) and the gcx
  integration form (subcommand vs. skill vs. both). Tracked in `SPECS.md` Open questions.
*Addendum (2026-06-23):* The version-tracking concern above is moot. Entry 11 superseded
the "consumers import the MCP adapter" model — both `grafana/mcp-grafana` and `grafana/gcx`
import the **core only** (`pkg/grafanadocs`), which has no mcp-go dependency. The adapter's
mcp-go version (v0.55.0) is therefore free to diverge from mcp-grafana's (v0.46.0) without
any import conflict, as confirmed by the working mcp-grafana integration (entry 19). The
adapter's version only matters to the standalone `hack-doc-server` binary.

## 11. Consumer integration model: consumers write their own wrappers
*Added: 2026-06-22*
**Decision:** `grafana/mcp-grafana` and `grafana/gcx` import the **core** (`pkg/grafanadocs`)
directly and write their own idiomatic wrappers. They do NOT import our MCP or cobra
adapters. The adapters serve the standalone server and generic consumers only.
**Rationale:** Verified against both codebases:
- mcp-grafana defines all MCP tools inline in `tools/*.go` using `mcpgrafana.MustTool`.
  External modules are imported as client libraries (like `incident-go`), never as
  pre-built tool registrations. Docs tools would follow the same pattern: `tools/docs.go`
  importing our core.
- gcx's output system (`internal/output`, `cmdio.Options`) and agent annotation system
  (`internal/agent`) are not importable by external modules. The established pattern
  (see `instrumentation/check` wrapping `github.com/grafana/otel-checker`) is: external
  core logic + gcx-owned CLI layer with `cmdio` and agent annotations.
**Consequence:**
- The core's API surface is the primary compatibility contract — clean exported functions
  with plain Go types, no framework dependencies.
- Our adapters can use whatever patterns are most natural (mark3labs/mcp-go directly for
  the MCP adapter, cobra for the CLI) without constraining the consumers.
- No circular import risk: consumers depend on our core, never the reverse.
- Supersedes the "consumed by mcp-grafana" language in entry 10 regarding the MCP adapter.

## 12. Search ranking: word boundaries + TF-IDF + phrase bonus
*Added: 2026-06-22*
**Decision:** Replace naive substring scoring with:
1. Word-boundary matching (tokenize both query and entry fields, match whole words only).
2. TF-IDF weighting (IDF computed once at index load; rare terms score higher).
3. Title 3× weight, description 1×.
4. All-tokens multiplier (1.5× when every query token matches).
5. Exact phrase bonus (2× when full query appears in title).
**Rationale:** Substring matching produced false positives (e.g., "rate" matching "migrate").
TF-IDF is the simplest effective relevance signal — no external deps, no network calls,
computed once at startup. The combined approach keeps invariants I2 (zero inference cost)
and I7 (lean search) while dramatically improving precision.
**Consequence:** `buildIDF` runs at index load time (O(n) over entries). The `idf` map is
stored on the `Index` struct. Score values are now float-derived integers (×100 for
granularity), so absolute score values changed — but ordering is what matters.

## 13. Security hardening: defense-in-depth
*Added: 2026-06-22*
**Decision:** Add multiple defense layers beyond the existing SSRF allowlist:
1. Rate limiter (5 concurrent, 200ms gap) on outbound doc fetches.
2. Body size caps (2 MiB docs, 10 MiB index) via `io.LimitReader`.
3. HTTP client timeouts (30s docs, 60s index) — no more `http.DefaultClient`.
4. Redirect guard (`CheckRedirect`) blocks any redirect leaving `grafana.com`.
5. Index entry URL validation — non-`https://grafana.com/` entries dropped at parse.
6. Non-product section exclusion ("Documentation home", "Copyright notice").
**Rationale:** OWASP Secure Product Design review identified gaps in defense-in-depth
and zero-trust-upstream. Each fix is minimal code with no behavioral impact on normal
operation — they only trigger under abnormal/adversarial conditions.
**Consequence:** New invariants I9–I13 added to SPECS.md.

## 14. License: Apache 2.0
*Added: 2026-06-22*
**Decision:** Use Apache License 2.0, matching `grafana/mcp-grafana` and `grafana/gcx`.
**Rationale:** Consistency with the two primary integration targets. All transitive
dependencies (MIT, BSD-3-Clause, ISC) are compatible with Apache 2.0.
**Consequence:** `LICENSE` file added to repo root.

## 15. Actionable empty-result messages
*Added: 2026-06-22*
**Decision:** When a tool returns no results, respond with a human-readable message
that guides the agent toward a corrective action (e.g., "use list_products to see
available products") rather than returning bare `[]` or empty content.
**Rationale:** AI agents seeing `[]` have no signal on why or what to do next. Guidance
reduces wasted tool calls and improves agent workflow efficiency.
**Consequence:** `search_docs` returns guidance text (not JSON) on empty results.
`get_doc` returns an error with suggestion when a section is not found.

## 16. Cobra adapter: standalone, gcx-compatible
*Added: 2026-06-23*
**Decision:** Build `pkg/grafanadocs/cli` as a mountable cobra `docs` command group
(`search`, `get`, `outline`, `products`), following gcx's opts/setup/Validate pattern and
custom-codec output model. Per entry 11, gcx is expected to write its own `cmd/gcx/docs/`
against the core, so this adapter is a **proof-of-concept and reference**, not the import
path gcx will take. It therefore:
- depends only on `cobra`/`pflag` + our core — **no gcx internal imports** (gcx's
  `internal/output`/`cmdio` and `internal/agent` are not importable cross-module);
- uses a small local `output` helper that mirrors `cmdio.Options` (text codec + json /
  yaml / agents) so a gcx port is mechanical;
- accepts the `*grafanadocs.Index` from the caller — the adapter is stateless and does no
  index loading, avoiding duplicate fetches;
- documents the agent annotations (cost/hint) gcx should apply from its own registry.
**Rationale:** Validates that the core API works cleanly with gcx conventions, and gives
the standalone repo a usable CLI, while honoring the "consumers own their CLI layer"
decision from entry 11.
**Consequence:** New direct deps `github.com/spf13/cobra` and `github.com/spf13/pflag`;
`gopkg.in/yaml.v3` promoted from indirect to direct. JSON/YAML output keys mirror the MCP
adapter for cross-surface consistency. SPECS.md "CLI adapter surface" section added; the
cobra adapter moved from Planned to Implemented packages.

## 17. gcx compatibility re-verified after upstream pull
*Added: 2026-06-23*
**Decision:** Re-checked the adapter against a fresh `grafana/gcx` pull. Findings and the
small alignment we made:
- **`cmdio` → `output`:** gcx's options type is now `output.Options` in package
  `internal/output` (entries 11/16 call it `cmdio.Options` — kept as historical record;
  this entry is the current truth). Methods `DefaultFormat`/`RegisterCustomCodec`/
  `BindFlags`/`Validate` are unchanged.
- **Custom-codec interface:** gcx's `format.Codec` is `Encode + Decode + Format()`. Our
  text codecs already match `Encode(w, v) error` with a type assertion (same pattern as
  gcx's `versionTextCodec`); a port adds a trivial `Format()` and an unsupported
  `Decode()` — still mechanical.
- **Deps:** cobra v1.10.2 (identical to gcx). Bumped our pflag v1.0.9 → v1.0.10 to match
  gcx. Our `go 1.25.5` directive is below gcx's `go 1.26.2`, which is fine — the consumer
  module sets the toolchain when importing our lower-versioned module.
- **New gcx flags:** `--jq` and `--json` (field selection) are additive and wired by
  `BindFlags`, so a gcx-native `gcx docs` would inherit them automatically. No breakage.
**Rationale:** A doc-only drift (the rename) would mislead whoever ports the adapter into
gcx; aligning pflag avoids a needless version bump surprise.
**Consequence:** Updated `cmdio.Options` → `output.Options` references in SPECS.md,
README.md, `research/gcx-integration-patterns.md`, and the `cli/command.go` doc comment.
Our code does not import gcx, so there were no compile-level compatibility changes.

## 18. gcx integration validated: core-only import, own CLI layer
*Added: 2026-06-23*
**Decision:** Built a working `gcx docs` prototype (`feat/docs-command` branch in
`grafana/gcx`) that imports only `pkg/grafanadocs` and writes its own CLI layer, validating
the consumer integration model from entry 11. Key findings:
- **Core types are serialization-agnostic by design.** `Entry`, `Product`, `Heading`,
  `ExcerptResult`, and `SearchOpts` carry no `json` tags. gcx creates thin wrapper types
  (`searchEntry`, `getResult`, `outlineHeading`, etc.) with `snake_case` `json` tags to
  support `--json` field selection and structured output. This is intentional — the core
  stays free of serialization opinions, and each consumer applies its own conventions.
- **`Validate()` takes no parameters.** gcx's canonical opts pattern is `Validate() error`
  with positional args stored on the opts struct in `RunE` before validation. The initial
  gcx prototype had `Validate(query string)` / `Validate(rawURL string)` which diverged;
  this was caught and fixed during review.
- **Agent-mode diagnostics use `output.EmitHint`, not raw `fmt.Fprintln`.** gcx's
  dual-purpose output contract (CONSTITUTION FR-104) requires stderr diagnostics to be
  typed JSONL in agent mode. The empty-search guidance (invariant I13) must go through
  `cmdio.EmitHint` in gcx, not raw stderr writes. Our cobra adapter's plain-text stderr
  approach is fine for the standalone CLI but would violate gcx's contract if copied
  verbatim.
- **Index lifecycle: lazy `sync.Once`.** gcx's prototype loads the index on first
  subcommand that needs it (`search`, `products`). `get` and `outline` bypass the index
  entirely — they only call `FetchDoc`. This avoids a ~1 MB network fetch on unrelated
  commands or `--help`.
- **Agent annotations added.** All four leaf commands received `token_cost` and `llm_hint`
  annotations, passing gcx's `TestConsistency_AllLeafCommandsHaveTokenCost` suite.
- **`--jq` and `--json` inherited for free.** Because gcx wires these via `BindFlags`,
  the docs commands got field selection and jq transformation with zero extra code.
**Rationale:** A real integration attempt surfaces friction points that spec review alone
cannot. The findings above close two open questions in SPECS.md and refine the guidance
for future consumers.
**Consequence:** SPECS.md open questions for "gcx integration form" and "Index lifecycle in
consumers" moved to Closed. Core data types documented in SPECS.md to make the
serialization-agnostic design explicit. No changes to hack-doc-server code — all fixes
were in the gcx branch.

## 19. mcp-grafana integration validated: core-only import, MustTool wrappers
*Added: 2026-06-23*
**Decision:** Built a working docs integration in `grafana/mcp-grafana` (`tools/docs.go`)
that imports only `pkg/grafanadocs` and registers four tools — `search_docs`, `get_doc`,
`get_doc_outline`, `list_products` — using mcp-grafana's own `mcpgrafana.MustTool`,
`AddDocsTools(*server.MCPServer)`, and a `toolEntries()` row, exactly matching the pattern
shared by all 30+ existing tool categories. This validates entry 11's consumer integration
model against the second primary consumer. Key findings:
- **Core-only import; adapter not used.** mcp-grafana wrote its own param/result types with
  `json` + `jsonschema` struct tags (`SearchDocsParams`, `GetDocResult`, etc.) wrapping the
  core's plain Go types. The MCP adapter (`pkg/grafanadocs/mcp`) was never imported, as
  predicted. The core's zero-framework-dependency design meant only stdlib-shaped types
  crossed the boundary.
- **mcp-go version mismatch is a non-issue.** mcp-grafana uses mcp-go v0.46.0; our adapter
  uses v0.55.0. Because the consumer imports the core (no mcp-go dep) and not the adapter,
  there is no conflict — confirming the addendum on entry 10.
- **Consumer forbids the default logger.** mcp-grafana's `tools/` package enforces
  `golangci-lint`'s `sloglint` with `no-global: "all"` (only `cmd/` is excluded). The first
  implementation logged index load/errors via `slog.Info()`/`slog.Error()` and failed lint;
  the fix was to remove logging from the tool layer and let errors propagate. Future consumer
  guidance: keep tool handlers free of package-global logging.
- **Index lifecycle: lazy `sync.Once`.** The index loads on first `search_docs`/`list_products`
  call. `get_doc`/`get_doc_outline` only call `FetchDoc`, so they never trigger the load —
  same split as the gcx prototype (entry 18). `DOCS_INDEX_URL` env var overrides the default
  index URL, mirroring the standalone server.
- **`MustTool` auto-generates JSON Schema.** The `jsonschema` struct tags on the consumer's
  param types drive tool schema generation; commas inside tag descriptions must be escaped
  (`\\,`) per mcp-grafana's custom jsonschema linter.
- **go.mod uses a local `replace`.** For now the dependency is wired via
  `replace github.com/grafana/hack-doc-server => <local path>` pending a published module.
**Rationale:** A second real integration confirms the core API is the durable contract and
that consumers can adopt their own framework conventions without the core imposing any. The
sloglint finding is a concrete, transferable constraint for anyone writing mcp-grafana tools.
**Consequence:** SPECS.md gains a "mcp-grafana integration form" closed question; entry 10
gets an addendum about version independence. No changes to hack-doc-server code — all work
was in mcp-grafana (`tools/docs.go`, `tools/docs_unit_test.go`, `cmd/mcp-grafana/main.go`,
`go.mod`).

## 20. Code-fence-aware processing (headings, cleanup)
*Added: 2026-06-23*
**Decision:** `Outline`, `excerptBySection`, and `Cleanup` now respect fenced code block
boundaries (` ``` ` and `~~~`, per CommonMark). Content inside code blocks is never
modified or misidentified.
**Rationale:** Three bugs were found via the Tempo configuration page — the longest page
in Tempo's docs (~3,000 lines with dozens of YAML code blocks):
1. `Outline` treated YAML comments (`# comment`) inside ` ```yaml ` blocks as markdown
   headings. The Tempo page produced 1,191 "headings" instead of 50 real ones.
2. `excerptBySection` matched heading names inside code fences before the real heading.
3. `Cleanup`'s `stripShortcodes` and `stripHTMLComments` regex replacements ran across
   the entire document, stripping HTML comments and Hugo shortcodes from inside code blocks
   (e.g. template examples showing `<!-- ... -->` or `{{< shortcode >}}`). This violates
   I8 (cleanup preserves meaning).
**Consequence:** New invariant I14. Fixes:
- `Outline`/`excerptBySection`: `isFenceBoundary` helper tracks fence state; lines inside
  fences are skipped when scanning for headings. Tempo outline: 1,191 → 50.
- `Cleanup`: `extractCodeBlocks`/`restoreCodeBlocks` replace code blocks with NUL-byte
  placeholders before running strip operations, then restore them. Code block content is
  byte-identical after cleanup.
Additionally, `isFenceBoundary` was refactored to `fenceBoundaryMarker` — it now returns
the marker type (`"``\`"` or `"~~~"`) so callers can enforce CommonMark's rule that a
closing fence must use the same character as the opening fence. Without this, a `~~~`
line inside a backtick-opened fence would incorrectly close it, re-exposing code block
content to heading detection.
Test coverage added for all four bugs (backtick fences, tilde fences, mismatched fence
markers, HTML comments and shortcodes inside code blocks, sample fixture regression).

## 21. Defense-in-depth hardening: cleanup, rate limiter, Doc.Lines
*Added: 2026-06-23*
**Decision:** Four additional robustness fixes discovered during systematic audit of
markdown processing and supporting infrastructure:
1. `collapseBlankLines` ran *after* `restoreCodeBlocks`, collapsing intentional consecutive
   blank lines inside code blocks. Fix: reorder to run before restoration — blank line
   collapsing now only affects content outside code blocks (strengthens I8/I14).
2. `stripFrontmatter` matched `---` anywhere in the string, including mid-line in a YAML
   value (e.g. `description: Use --- for separators`). Fix: require the closing `---` to
   appear at the start of a line (`\n---`).
3. Rate limiter gap was enforced between *releases*, not *acquires*. Under burst concurrency,
   multiple goroutines could fire requests simultaneously without the 200ms gap. Fix: update
   `lastCall` at acquire time after the wait completes, so the next acquirer sees the correct
   timestamp. `release()` now only returns the semaphore slot (I9).
4. `Doc.EnsureLines()` cached `Lines` on first call and never recomputed if `Content` was
   mutated afterwards. Since `Doc` fields are public, external consumers could hit stale
   `Lines`. Fix: `EnsureLines` stores a snapshot of `Content` and recomputes `Lines` when
   `Content` has changed.
**Rationale:** Each fix addresses a correctness issue found by systematic code review. All
four are low-risk changes with targeted test coverage.
**Consequence:** I8 and I14 strengthened (cleanup ordering). I9 now accurately enforces
the documented gap. `Doc.Lines` is always consistent with `Content`.

## 22. MCP adapter hardening: JSON casing, URL parsing, input validation
*Added: 2026-06-23*
**Decision:** Three bugs fixed in the MCP adapter and fetch layer:
1. **JSON key casing (I15).** Core types (`Entry`, `Heading`, `Product`) carry no `json`
   tags — intentionally serialization-agnostic (entry 18). The MCP adapter was marshaling
   them directly, producing PascalCase JSON keys (`"Title"`, `"Level"`, `"Name"`) instead
   of the snake_case keys documented in SPECS.md tool schemas. Fix: wrapper types
   (`searchEntry`, `outlineHeading`, `productEntry`) with explicit `json:"snake_case"` tags.
   The `handleGetDoc` response already used `map[string]any` with correct keys; the other
   three handlers now use wrappers.
2. **URL `.md` suffix (I16).** The `.md` suffix was appended to the raw URL string. URLs
   with fragments (`#section`), query params (`?v=1`), or trailing slashes (`/`) produced
   broken fetch URLs (`...#section.md`, `...?v=1.md`, `.../.md`). Fix: `ensureMDSuffix`
   parses the URL, strips trailing slash, appends `.md` to the path, and reconstructs —
   fragments and query params are preserved.
3. **Numeric input validation (I17).** MCP handlers cast `float64` args directly to `int`
   without checking for `NaN`, `Inf`, or negative values. Fix: `safeInt` helper rejects
   non-finite and negative values with a descriptive error before they reach the core.
**Rationale:** Bugs 1 and 2 were found by running the server against real URLs and
inspecting the JSON output. Bug 3 was found by systematic audit of the `float64 → int`
conversions. All three are correctness issues in the adapter layer — the core API was
unaffected.
**Consequence:** New invariants I15, I16, I17 in SPECS.md. Wrapper types added to the MCP
adapter (but NOT to the core — the serialization-agnostic design from entry 18 is preserved).
`net/url` now used in `fetch.go` for URL manipulation.

## 23. Systematic hardening pass: search, index, cleanup, fetch
*Added: 2026-06-23*
**Decision:** Eleven edge-case bugs fixed across the codebase, discovered by systematic
code-path audit:
1. **`buildIDF` empty-index guard.** Early return for empty corpus avoids `math.Log(0/0)`.
   Not a crash (the loop body never runs), but now explicitly handled.
2. **Scanner buffer.** `bufio.Scanner` default max token is 64 KiB. Set to 1 MiB for the
   index parser to handle entries with very long descriptions without silent truncation.
3. **`needsIndex` rewrite.** Old logic checked all args against a deny list (`"-h"`,
   `"help"`, `"completion"`), so `docs get --section help` would skip loading the index
   because `"help"` matched as a bare arg. New logic: skip flags, then check the first
   positional arg against an allow list (`"search"`, `"products"` — the only commands
   that read the index). `get` and `outline` only call `FetchDoc`.
4. **Double allowlist check.** `FetchDoc` now checks the allowlist on both the original
   URL and the URL after `ensureMDSuffix` transforms it. Prevents path manipulation
   through the URL transform (I21).
5. **Unclosed fence in `extractCodeBlocks`.** If a code fence opens but never closes
   (e.g. body-size truncation), the old code appended the raw fence lines at the end of
   `out`, re-ordering content. Fix: treat the unclosed block as a protected code block
   via a placeholder in its original position (I23).
6. **Exact product filter.** `Search` was using `containsFold` (substring match), so
   `product="Lo"` matched `"Grafana Loki"`. Changed to `strings.EqualFold` for exact
   case-insensitive matching (I18).
   *Addendum (2026-06-23): replaced by the hybrid precedence resolver in NOTE 28 —
   exact-only matching was too strict for agents/CLI users who pass loose terms.*
7. **Shortcode regex.** The old regex `[^}]*[>%]\}\}` could over-match when shortcode
   arguments contained `>`. Replaced with two non-greedy alternatives:
   `\{\{<.*?>}}` and `\{\{%.*?%}}`.
8. **User-Agent header.** All outbound HTTP requests (`httpClient`, `indexClient`) now
   include `User-Agent: hack-doc-server/0.1` to prevent CDN/WAF throttling (I22).
9. **`collapseBlankLines` optimization.** Replaced the `for Contains + ReplaceAll` loop
   (O(n*m) worst case with hundreds of consecutive blank lines) with a single-pass
   `strings.Builder` approach.
10. **Orphan entries.** Entries before any `## Product` header had `Product: ""` and were
    searchable without product affiliation. Now silently dropped (I20).
11. **Index URL scheme validation.** `LoadIndex` now rejects non-`https` URLs (e.g.
    `file:///etc/passwd`, `http://`) before making any network call (I19).
**Rationale:** Each fix addresses a concrete edge case found by reviewing every code path
in the four core files. All are low-risk, targeted changes with test coverage.
**Consequence:** New invariants I18–I23 in SPECS.md. `needsIndex` rewritten with an
allow-list approach. `collapseBlankLines` is now O(n). Product filter semantics changed
from substring to exact match — callers must pass the full product name.

## 24. Fuzz-discovered bug: Cleanup not idempotent
*Added: 2026-06-23*
**Decision:** `Cleanup` now runs shortcode and HTML comment stripping in a convergence
loop (`for { prev := s; strip; strip; if s == prev { break } }`).
**Rationale:** Fuzz testing (`FuzzCleanup`, 5s run) discovered that removing an HTML
comment can reveal a shortcode: `{<!---->{<>}}` → strip HTML comment → `{{<>}}` →
strip shortcode → empty. A second `Cleanup` pass would strip the revealed shortcode,
breaking idempotency (`Cleanup(Cleanup(x)) != Cleanup(x)`). The convergence loop
ensures both strippers run until no more changes occur, guaranteeing idempotency.
**Consequence:** I8 strengthened — `Cleanup` is now provably idempotent (tested by
fuzz and a targeted `TestCleanup_Idempotent` with 8 representative inputs including
the exact fuzz failure). Negligible perf impact — the loop runs at most 2 iterations
for any realistic input.

## 25. Behavioral-depth fixes: ATX heading semantics and BOM tolerance
*Added: 2026-06-23*
**Decision:** Three CommonMark-conformance bugs found during behavioral-depth review
are fixed in `parseHeading` ([outline.go]) and `LoadIndexFromReader` ([index.go]):
1. **Trailing `#` closing sequence** is now stripped: `## Storage ##` yields text
   `Storage`, not `Storage ##`. A `#` inside a word (`C#`) or not preceded by a space
   (`foo#`) is preserved. A heading that is only `#`s is treated as a non-heading.
2. **Indented code blocks** (4+ leading spaces) starting with `#` are no longer
   misidentified as headings. `Outline` only protects *fenced* blocks; indented code
   blocks previously slipped through to `parseHeading`.
3. **UTF-8 BOM** on the index is stripped from the first line, so a BOM-prefixed
   `## Product` header still matches (previously the whole first product and its
   entries were dropped).
**Rationale:** (1) broke `get_doc --section "Storage"` matching against pages using
closed ATX headings; (2) produced phantom outline entries from code samples; (3) is a
realistic failure for index files served with a BOM.
**Consequence:** New invariants I24 (ATX heading semantics) and I25 (BOM tolerance).
Tests added in `outline_test.go` (`TestParseHeading_StripsTrailingHashes`,
`TestParseHeading_IgnoresIndentedCodeBlocks`, `TestOutline_IgnoresIndentedCodeHash`)
and `index_test.go` (`TestLoadIndex_StripsBOM`). Setext headings remain unsupported by
design — the index/llms format uses ATX exclusively.

## 26. Behavioral-depth fixes: CRLF, variable-length fences, TOML frontmatter
*Added: 2026-06-23*
**Decision:** Three more cleanup/parsing bugs found during behavioral-depth review are fixed:
1. **CRLF line endings**: `Cleanup` now normalizes `\r\n` and `\r` to `\n` up front.
   Previously `collapseBlankLines` counted `\n` only, so a `\r` in `\r\n\r\n` reset the run
   counter and blank-line collapsing silently no-opped on CRLF content.
2. **Variable-length fences**: fence detection is unified into a shared `fenceInfo` +
   `fenceTracker` (in [outline.go]) used by `Outline`, `excerptBySection`, and
   `extractCodeBlocks`. Fences now track the marker character *and* run length, so a 4-backtick
   fence containing a 3-backtick line is no longer closed early (which previously leaked the
   inner block's content to the shortcode/comment strippers and to heading detection). A closing
   fence must match the char, be ≥ the opening length, and have no info string (per CommonMark).
   The old fixed-3-char `fenceBoundaryMarker` is removed.
3. **TOML frontmatter**: `stripFrontmatter` now handles `+++`-delimited TOML in addition to
   `---` YAML (Grafana docs are built with Hugo, which supports both).
**Idempotency follow-on:** adding TOML stripping surfaced a latent idempotency bug (also
present for `---`): an HTML comment or leading whitespace could hide a frontmatter delimiter so
it was only stripped on the second pass. Fixed by (a) trimming the leading/trailing edge before
frontmatter detection, and (b) moving `stripFrontmatter` into the existing shortcode/comment
convergence loop. `FuzzCleanup` confirms idempotency holds.
**Rationale:** CRLF and TOML are realistic for HTTP-fetched Hugo content; nested fences appear
in docs that document markdown itself.
**Consequence:** I8 extended (CRLF normalization, edge trimming, frontmatter in the convergence
loop) and I14 extended (variable-length, CommonMark-conformant fence matching via shared helper).
Tests: `TestCleanup_CRLF`, `TestCleanup_VariableLengthFences`, `TestCleanup_TOMLFrontmatter`,
`TestOutline_VariableLengthFences`, and `FuzzFenceInfo` (replacing `FuzzFenceBoundaryMarker`).

## 27. Move index-need gating into the cli package
*Added: 2026-06-23*
**Decision:** The `needsIndex` helper that lived in `cmd/docs/main.go` (untested, `package
main`) is promoted to `cli.NeedsIndex` alongside the command definitions, backed by an
`indexReadingCommands` allow-list. `cmd/docs` now calls `cli.NeedsIndex`; `indexURL` stays in
`cmd/docs` because `DOCS_INDEX_URL` is that binary's concern.
**Rationale:** Keep `main` thin and put branching logic in an importable, testable package
(standard Go practice). This function had a prior bug (fragile deny-list → allow-list) with no
regression test; it is also reusable by other front-ends (e.g. gcx) that gate index loading.
**Consequence:** New invariant I26. Table test `TestNeedsIndex` plus a drift guard
`TestNeedsIndexInSyncWithCommands` that asserts every allow-listed name is a real subcommand.
`cmd/docs/main.go` is now just wiring (no testable logic beyond the env-var lookup).

## 28. Hybrid product filter resolution (exact → prefix → substring)
*Added: 2026-06-23*
**Decision:** Replace the exact-only product filter (NOTE 23, item 6) with a tiered
resolver, `resolveProductFilter`. It maps the user's `product` string to the set of
canonical product names by trying match levels in precedence order and stopping at the
first non-empty level: exact (case-insensitive), then prefix, then substring. A precise
name selects exactly one product; a loose term still resolves (`"agent"` →
`"Grafana Agent"`); an exact hit (`"grafana"` = bare `"Grafana"`) wins over broader
prefix/substring candidates; an unmatched filter yields zero results.
**Rationale:** Exact-only matching (introduced to make results deterministic) was hostile to
the primary callers — agents and CLI users routinely pass short product hints (`loki`,
`agent`) rather than the full catalog name. The tiered resolver keeps determinism (same
input → same output, no relevance dependence) while restoring friendly partial matching as
a *fallback*, not the default. Resolution is intentionally in the core so all three surfaces
(cmd/docs CLI, gcx, MCP `search_docs`) get identical behavior with no consumer code changes.
**Consequence:** I18 rewritten (precedence rule + addendum). `Search` resolves the filter to
a product-name set before scoring; `strings.EqualFold` filtering removed. Test
`TestSearch_ExactProductMatch` became `TestSearch_ProductResolution` covering exact, prefix,
substring, and unknown cases. Downstream: gcx's `--product agent` test now passes via the
substring fallback after it re-vendors; mcp-grafana (exact names already) is unaffected.

## 29. PR #3 review fixes: rate-limiter spacing and safeInt overflow
*Added: 2026-06-23*
**Decision:** Address two Copilot review findings on PR #3.
1. **Rate-limiter gap not enforced under concurrency.** With `maxConcurrent=5`, multiple
   goroutines could pass the semaphore, all read the same `lastCall`, compute the same wait,
   wake together, and proceed near-simultaneously — collapsing the 200ms gap. Replaced
   `lastCall` with a `nextAllowed` cursor: `acquire` reserves `slot = max(now, nextAllowed)`
   and advances `nextAllowed = slot + minGap` while holding the lock, then waits for `slot`
   outside the lock. Each concurrent acquirer now gets a distinct, spaced slot.
2. **`safeInt` overflow.** `int(v)` for a finite float64 beyond int range is
   implementation-dependent in Go (yields a negative number on amd64), bypassing the
   negativity check. Added an explicit upper bound (`maxSafeInt = 1<<31`); larger values are
   rejected with a descriptive error.
**Rationale:** Both are correctness gaps in code this PR was already hardening. The
rate-limiter fix makes the documented I9 spacing invariant actually hold under concurrency;
the safeInt fix closes a validation bypass in the MCP input sanitizer (I17).
**Consequence:** I9 and I17 reworded. `TestRateLimiter_ConcurrentStress` strengthened to
assert total elapsed ≥ (N−1)·gap (true spacing, not just no-deadlock). `TestSafeInt` extended
with NaN/Inf/overflow/cap cases. No public API change; both consumers unaffected.
