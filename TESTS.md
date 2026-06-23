# TESTS — hack-doc-server

Test scenarios for the product contract in `SPECS.md`. Each scenario is
**Setup / Action / Assertion**. Add scenarios for new functionality; do not remove a
scenario unless its feature is removed.

## Testing approach

Principles: no network in tests, fixtures over mocks, test invariants directly.

- **Fixtures:** A captured `testdata/llms-full.txt` and a few representative page
  snapshots (`testdata/pages/`). Tests never hit `grafana.com`.
- **Golden files:** Markdown cleanup uses input/expected pairs in `testdata/cleanup/`.
  Regressions show as diffs.
- **Invariant tests:** Each invariant (I1–I8) has at least one test that fails if
  violated — not just scenarios that happen to exercise them.
- **Security (I3):** The allowlist is tested with an exhaustive table: valid URLs,
  SSRF attempts (metadata endpoints, private IPs, non-grafana hosts, scheme tricks),
  path traversal. This is the primary security surface — keep it tight.
- **Bounds (I4, I7):** Slice and search tests assert token/size budgets against
  representative pages (e.g. Tempo config ~139 KB).
- **Table-driven:** One `tests` slice, one loop — no per-case functions.
- **HTTP layer:** `httptest.Server` returning fixture content for fetch tests.
- **MCP integration:** Spawn the binary, call tools via `mcp.Client`. Runs last.

## Scenario: Configure Tempo (retrieval)
**Setup:** Server running with a parsed index.
**Action:** `search_docs("configure tempo")` then `get_doc` on the top result, then
`get_doc(url, section="...")` for a relevant section.
**Assertion:** A result resolves to `.../tempo/latest/configuration.md`; `get_doc`
returns cleaned markdown with the canonical source URL (I5); a sectioned call returns
only that section, not the full ~139 KB page (I4).

## Scenario: Ground dashboard creation
**Setup:** Server running.
**Action:** `search_docs("create dashboard provisioning schema")` then `get_doc`.
**Assertion:** Results include dashboard/provisioning docs; content is sufficient for an
agent to then call mcp-grafana/gcx to create a dashboard.

## Scenario: Generate a query (PromQL/LogQL/TraceQL)
**Setup:** Server running.
**Action:** `search_docs("traceql query syntax")` (and LogQL/PromQL variants), then `get_doc`.
**Assertion:** Results point at the correct query-language docs for each language.

## Scenario: Index parsing
**Setup:** A captured `llms-full.txt` fixture.
**Action:** Parse it.
**Assertion:** Entry count and product-group count match the fixture; every entry has a
non-empty Title, a `grafana.com` `.md` URL, and a Product; the raw index is not exposed
via any tool (I1).

## Scenario: Fetch allowlist (SSRF guard)
**Setup:** Server running.
**Action:** `get_doc("http://169.254.169.254/...")` and `get_doc("https://evil.example/x.md")`.
**Assertion:** Both are rejected without a network call; only `grafana.com` `/docs/` URLs
are fetched (I3).

## Scenario: Outline and section slicing
**Setup:** A large doc page (e.g. Tempo configuration).
**Action:** `get_doc_outline(url)`, then `get_doc(url, section=<a heading>)`, then a paged
`get_doc(url, offset, limit)`.
**Assertion:** Outline is a small heading list; sectioned content matches that heading;
paging returns contiguous, bounded slices and reports total size (I4).

## Scenario: Markdown cleanup
**Setup:** A page containing frontmatter, Hugo shortcodes, and nav boilerplate.
**Action:** `get_doc(url)`.
**Assertion:** Frontmatter/shortcodes/nav removed; documented prose, code blocks, and
tables preserved unchanged (I8).

## Scenario: Product listing
**Setup:** Server running with a parsed index.
**Action:** `list_products()`.
**Assertion:** Returns all product groups from the index; each product has a non-empty
name and a positive entry count; product names match the `## <Product> documentation`
headers in `llms-full.txt`. "Documentation home" and "Copyright notice" are excluded (I12).

## Scenario: Index entry URL validation
**Setup:** An index with mixed URLs (grafana.com and non-grafana.com).
**Action:** `LoadIndexFromReader(input)`.
**Assertion:** Only entries with `https://grafana.com/` URLs are stored; non-grafana
entries are silently dropped; product counts reflect only valid entries (I11).

## Scenario: Word-boundary search (no false positives)
**Setup:** Server running with live index.
**Action:** `search_docs(query="rate", limit=10)`.
**Assertion:** No results contain "migrate", "generate", or other words that merely
contain "rate" as a substring. Only entries with "rate" as a whole word match.

## Scenario: TF-IDF ranking (rare terms score higher)
**Setup:** Server running with live index.
**Action:** Compare IDF of "clustering" vs "grafana" in the built index.
**Assertion:** "clustering" has a higher IDF weight than "grafana" (rarer = more specific).

## Scenario: Empty search results return guidance
**Setup:** Server running.
**Action:** `search_docs(query="rate", product="tempo")` (produces no results).
**Assertion:** Response is a human-readable message mentioning `list_products`, not bare `[]`.

## Scenario: Missing section returns guidance
**Setup:** Server running.
**Action:** `get_doc(url="...", section="Nonexistent")`.
**Assertion:** Response is an error mentioning `get_doc_outline`, not empty content.

## Scenario: Redirect guard
**Setup:** Server running (or unit test with mock).
**Action:** A request to grafana.com that redirects to `evil.com`.
**Assertion:** The redirect is blocked; no request reaches the non-grafana host.

## Scenario: CLI search output formats
**Setup:** `cli.Command(idx)` built from a parsed index.
**Action:** Run `docs search clustering` with `-o text`, `-o json`, and `-o agents`.
**Assertion:** `text` prints an aligned `TITLE/PRODUCT/URL` table; `json` is indented and
unmarshals to a list of entries; `agents` is a single compact JSON line. An unknown format
returns an error mentioning the valid formats.

## Scenario: CLI empty search guidance
**Setup:** `cli.Command(idx)` built from a parsed index.
**Action:** Run `docs search zzzznotathing`.
**Assertion:** stdout stays a clean (header-only) parseable result set; guidance
("No results found...") is written to stderr, not stdout (I13).

## Scenario: CLI fetch guards (no network)
**Setup:** `cli.Command(idx)` built from a parsed index.
**Action:** Run `docs get ""`, `docs get https://evil.com/docs/x.md`, and the equivalent
`docs outline` invocations.
**Assertion:** Empty URLs return "url is required"; non-grafana hosts are rejected by the
allowlist (I3) before any network call, surfaced as a "rejected host" error.

## Scenario: CLI products listing
**Setup:** `cli.Command(idx)` built from a parsed index.
**Action:** Run `docs products` with `-o text` and `-o json`.
**Assertion:** `text` prints a `PRODUCT/COUNT` table excluding non-product sections (I12);
`json` wraps the list under a `products` key.
