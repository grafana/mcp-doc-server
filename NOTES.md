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
