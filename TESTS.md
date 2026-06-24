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

## Scenario: Cleanup preserves code block contents
**Setup:** A document with an HTML comment (`<!-- ... -->`) and a Hugo shortcode
(`{{< ... >}}`) both inside a fenced code block, and the same patterns outside.
**Action:** `Cleanup(doc)`.
**Assertion:** HTML comments and shortcodes outside code blocks are stripped; those
inside code blocks are preserved byte-for-byte (I8, I14).

## Scenario: Mismatched fence markers don't close the fence
**Setup:** A document where a backtick-opened fence (` ```yaml `) contains a tilde
line (`~~~`) followed by lines with `# comment` syntax.
**Action:** `Outline(doc)`.
**Assertion:** The tilde line does not close the backtick fence. Comment lines after
the tilde are still inside the fence and are not returned as headings. Only the matching
backtick close (` ``` `) ends the fence (I14, CommonMark spec).

## Scenario: Code-fence-aware outline (no false headings)
**Setup:** A document with YAML comments (`# comment`) inside ` ```yaml ` fenced code
blocks, and real headings outside.
**Action:** `Outline(doc)`.
**Assertion:** Only real markdown headings appear; YAML/shell/config comments inside
fenced code blocks are not returned as headings (I14). Tested with both backtick and
tilde fences.

## Scenario: Code-fence-aware section extraction
**Setup:** A document where a heading name (e.g. "Storage") appears both as a YAML
comment inside a code fence and as a real heading outside.
**Action:** `Excerpt(doc, ExcerptOpts{Section: "Storage"})`.
**Assertion:** Matches the real heading, not the code-fence comment. The section end is
also not confused by code-fence comments that look like same-level headings (I14).

## Scenario: Cleanup preserves blank lines inside code blocks
**Setup:** A document with 3+ consecutive blank lines inside a fenced code block.
**Action:** `Cleanup(doc)`.
**Assertion:** The consecutive blank lines inside the code block are preserved
byte-for-byte. Only blank lines outside code blocks are collapsed (I8, I14).

## Scenario: Frontmatter with dashes in YAML value
**Setup:** A document with YAML frontmatter where a value contains `---` mid-line
(e.g. `description: Use --- for separators`).
**Action:** `Cleanup(doc)`.
**Assertion:** The frontmatter is fully stripped. The `---` inside the YAML value does
not cause early termination; content after the real closing `---` delimiter is preserved.

## Scenario: Rate limiter enforces gap between acquires
**Setup:** A `rateLimiter` with a 50ms gap and concurrency of 2.
**Action:** Acquire, release, then immediately acquire again.
**Assertion:** The second acquire blocks for approximately the gap duration (I9). The gap
is measured from the previous acquire, not the previous release.

## Scenario: EnsureLines recomputes after Content mutation
**Setup:** A `Doc` with initial `Content` "line1\nline2".
**Action:** Call `EnsureLines()`, then change `Content` to "a\nb\nc", then call
`EnsureLines()` again.
**Assertion:** `Lines` reflects the updated `Content` ("a", "b", "c"), not the stale
original ("line1", "line2").

## Scenario: MCP search results use snake_case JSON keys
**Setup:** MCP `Server` built from a parsed index.
**Action:** Call `handleSearchDocs` with query "clustering".
**Assertion:** The JSON response contains `"title"`, `"url"`, `"description"`, `"product"`
keys (snake_case). PascalCase keys like `"Title"` or `"URL"` must NOT appear (I15).

## Scenario: MCP list_products uses snake_case JSON keys
**Setup:** MCP `Server` built from a parsed index.
**Action:** Call `handleListProducts`.
**Assertion:** The JSON response contains `"name"` and `"count"` keys. PascalCase keys
like `"Name"` or `"Count"` must NOT appear (I15).

## Scenario: URL .md suffix preserves fragments
**Setup:** A URL `https://grafana.com/docs/tempo/latest/configuration#auth`.
**Action:** `ensureMDSuffix(url)`.
**Assertion:** Returns `https://grafana.com/docs/tempo/latest/configuration.md#auth` —
the `.md` suffix is on the path, the fragment is preserved (I16).

## Scenario: URL .md suffix preserves query parameters
**Setup:** A URL `https://grafana.com/docs/tempo/latest/configuration?v=1`.
**Action:** `ensureMDSuffix(url)`.
**Assertion:** Returns `https://grafana.com/docs/tempo/latest/configuration.md?v=1` (I16).

## Scenario: URL .md suffix handles trailing slash
**Setup:** A URL `https://grafana.com/docs/tempo/latest/`.
**Action:** `ensureMDSuffix(url)`.
**Assertion:** Returns `https://grafana.com/docs/tempo/latest.md` — trailing slash is
stripped before appending `.md` (I16).

## Scenario: MCP rejects negative limit
**Setup:** MCP `Server` built from a parsed index.
**Action:** Call `handleSearchDocs` with `{"query": "clustering", "limit": -1}`.
**Assertion:** Returns an error result mentioning "negative" (I17).

## Scenario: MCP rejects negative offset
**Setup:** MCP `Server` built from a parsed index.
**Action:** Call `handleGetDoc` with `{"url": "...", "offset": -5}`.
**Assertion:** Returns an error result mentioning "negative" (I17).

## Scenario: safeInt rejects non-finite values
**Setup:** Direct unit test of `safeInt`.
**Action:** Call `safeInt` with positive, zero, negative, and decimal values.
**Assertion:** Positive and zero return the integer value; negative returns an error.
Decimal values are truncated (5.9 → 5).

## Scenario: CLI products listing
**Setup:** `cli.Command(idx)` built from a parsed index.
**Action:** Run `docs products` with `-o text` and `-o json`.
**Assertion:** `text` prints a `PRODUCT/COUNT` table excluding non-product sections (I12);
`json` wraps the list under a `products` key.

## Scenario: buildIDF handles empty index
**Setup:** An empty list of entries.
**Action:** `buildIDF(nil)`.
**Assertion:** Returns a non-nil, empty map. No panic or division by zero.

## Scenario: Product filter resolution precedence
**Setup:** Index with entries under "Grafana Agent".
**Action:** `Search(idx, "clustering", SearchOpts{Product: p})` for `p` in
{`"grafana agent"` (exact), `"grafana"` (prefix), `"agent"` (substring), `"nonexistent"`}.
**Assertion:** The exact, prefix, and substring filters all resolve to "Grafana Agent" and
return that product's entries; `"nonexistent"` returns no results. Resolution tries
exact → prefix → substring and stops at the first level that matches (I18).

## Scenario: Entries before any product header are dropped
**Setup:** An index with an entry before the first `## Product` header, and entries after.
**Action:** `LoadIndexFromReader(input)`.
**Assertion:** Only entries under a product header are indexed. The orphan entry is
silently dropped. Every entry has a non-empty `Product` (I20).

## Scenario: LoadIndex rejects non-https URLs
**Setup:** A `file:///etc/passwd` and `http://grafana.com/llms-full.txt` URL.
**Action:** `LoadIndex(ctx, url)`.
**Assertion:** Both are rejected with an error mentioning "rejected" before any network
or file read occurs (I19).

## Scenario: Unclosed fence preserved in place
**Setup:** A document with a `\`\`\`yaml` fence that never closes (simulating truncation).
**Action:** `Cleanup(doc)`.
**Assertion:** The unclosed fence and its content appear in the output in their original
position — not re-ordered to the end of the document (I23).

## Scenario: Shortcode regex handles > in arguments
**Setup:** A document with `{{< highlight go "linenos=true" >}}...{{< /highlight >}}`.
**Action:** `Cleanup(doc)`.
**Assertion:** Both shortcodes are fully stripped. The `>` inside the first shortcode's
arguments does not cause a partial match.

## Scenario: collapseBlankLines reduces correctly
**Setup:** A string with runs of 3, 4, and 5 consecutive blank lines.
**Action:** `collapseBlankLines(s)`.
**Assertion:** All runs are reduced to exactly 2 blank lines. Single and double blank
lines are preserved unchanged.

## Scenario: Cleanup is idempotent
**Setup:** A corpus of representative inputs including edge cases: nested shortcodes,
HTML comments adjacent to shortcodes, interleaved `{<!---->{<>}}`, empty strings,
unclosed fences, multiple blank lines.
**Action:** `Cleanup(Cleanup(input))`.
**Assertion:** `Cleanup(Cleanup(x)) == Cleanup(x)` for all inputs (I8).

## Scenario: Fuzz — Cleanup does not panic
**Setup:** Go fuzz corpus with seed inputs.
**Action:** `FuzzCleanup` with random bytes.
**Assertion:** No panics. Result ends with `\n`. Idempotent.

## Scenario: Fuzz — parseHeading bounds
**Setup:** Go fuzz corpus with heading-like strings.
**Action:** `FuzzParseHeading` with random strings.
**Assertion:** Level is always 0-6. Level 0 has empty text.

## Scenario: Fuzz — fenceBoundaryMarker returns valid values
**Setup:** Go fuzz corpus.
**Action:** `FuzzFenceBoundaryMarker` with random strings.
**Assertion:** Result is always one of `""`, `` "```" ``, `"~~~"`.

## Scenario: Fuzz — ensureMDSuffix does not panic
**Setup:** Go fuzz corpus with URL-like strings.
**Action:** `FuzzEnsureMDSuffix` with random strings.
**Assertion:** No panics. For hierarchical URLs, result contains `.md`.

## Scenario: Fuzz — LoadIndexFromReader does not panic
**Setup:** Go fuzz corpus with index-like text.
**Action:** `FuzzLoadIndexFromReader` with random bytes.
**Assertion:** No panics. All parsed entries have non-empty Title, URL, Product, and
URLs start with `https://grafana.com/`.

## Scenario: Rate limiter concurrent stress
**Setup:** A `rateLimiter` with concurrency=3 and gap=10ms.
**Action:** 20 goroutines each acquire/release 10 times.
**Assertion:** All goroutines complete within 30s. No panics, no data races (under
`-race`), no deadlocks.

## Scenario: Outline heading bounds invariant
**Setup:** A cleaned document from the sample fixture.
**Action:** `Outline(doc)`.
**Assertion:** Every heading has `1 <= Level <= 6`, `1 <= Line <= len(doc.Lines)`,
and non-empty `Text`.

## Scenario: Excerpt range consistency invariant
**Setup:** A cleaned document with various `ExcerptOpts` (offset, limit, section,
nonexistent section, offset past end).
**Action:** `Excerpt(doc, opts)`.
**Assertion:** For non-empty results: `Start >= 1`, `End <= Total`, `Start <= End`,
and `len(Split(Content, "\n")) == End - Start + 1`.

## Scenario: Search results have required fields
**Setup:** Live index or fixture.
**Action:** `Search(idx, "grafana", SearchOpts{Limit: 20})`.
**Assertion:** Every result has non-empty Title, URL, Product, and URL starts with
`https://grafana.com/`.

## Scenario: Live fetch and outline (network, skip in short mode)
**Setup:** Real grafana.com pages (small, large, various products).
**Action:** `FetchDoc` + `Outline` + `Excerpt` on each.
**Assertion:** Content is non-empty, headings have valid bounds, shortcodes are
stripped, excerpt range is consistent. Tests skip with `-short`.

## Scenario: Live index load (network, skip in short mode)
**Setup:** The real `grafana.com/llms-full.txt`.
**Action:** `LoadIndex` with the default URL.
**Assertion:** >1000 entries, >10 products, every entry has valid fields, search
returns results for "tempo configuration".

## Scenario: ATX heading closing sequence stripped
**Setup:** Heading lines with trailing `#`s: `## Storage ##`, `# Title #`,
`### Config #####`, plus `C#`, `foo#`, and `## ###`.
**Action:** `parseHeading(line)`.
**Assertion:** Trailing closing `#`s preceded by a space are stripped (`Storage`,
`Title`, `Config`). `C#` and `foo#` are preserved. `## ###` is not a heading (I24).

## Scenario: Indented code block is not a heading
**Setup:** Lines indented 0-5 spaces starting with `#`.
**Action:** `parseHeading(line)` and `Outline(doc)`.
**Assertion:** Lines indented 4+ spaces yield level 0 (indented code block); 0-3
spaces are headings. `Outline` excludes indented `# ...` code lines (I24).

## Scenario: Index tolerates UTF-8 BOM
**Setup:** An index whose first byte is a UTF-8 BOM (`\ufeff`) before `## Product`.
**Action:** `LoadIndexFromReader`.
**Assertion:** The first product header still matches; its entries are parsed with
the correct product name (I25).

## Scenario: Cleanup handles CRLF line endings
**Setup:** Markdown with `\r\n` endings, including 3+ blank lines, CRLF frontmatter,
and a CRLF code block.
**Action:** `Cleanup`.
**Assertion:** No `\r` remains, 3+ blank lines collapse to 2, frontmatter is stripped,
and code block content is preserved (I8).

## Scenario: Variable-length fences contain shorter fences
**Setup:** A 4-backtick fence containing a 3-backtick block, with a shortcode between
the inner pair and another shortcode outside the outer fence.
**Action:** `Cleanup` and `Outline`.
**Assertion:** The inner shortcode is preserved (the whole 4-backtick block is one
protected unit); the outer shortcode is stripped; a `#` line inside the nested block is
not treated as a heading (I14).

## Scenario: TOML frontmatter stripped
**Setup:** A page with `+++`-delimited TOML frontmatter.
**Action:** `Cleanup`.
**Assertion:** The TOML block and its keys are removed; body content is preserved (I8).
