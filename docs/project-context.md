# Project context (local)

This file is used as a pointer file for AI and agents, especiallly used for documentation skills.

## Identity

- **Product name (short):** mcp-doc-server
- **Product name (first mention in prose):** mcp-doc-server
- **GitHub org/repo:** grafana/mcp-doc-server

## Branches and releases

- **Default development branch:** `main`
- **Release branch pattern:** N/A — no formal release cadence yet. Feature branches merge to `main`.
- **Docs version mapping:** N/A — this repo is not a docs-site source. It *serves* Grafana docs via MCP tools.

## Documentation paths

- **Documentation root (filesystem):** `docs/` (project-level context only; no published doc site)
- **Generated pages (do not hand-edit):** none
- **Configuration reference index:** N/A — configuration is via environment variables only (`DOCS_INDEX_URL`)
- **Changelog:** none yet
- **Architecture / "start here" page:** `README.md` (overview), `SPECS.md` (contracts and invariants), `AGENTS.md` (SDD conventions)

## Code ↔ documentation mapping

| Code area | Documentation area |
| --------- | ------------------- |
| `pkg/grafanadocs/` | `SPECS.md` §Core API surface, §Core data types, §Invariants |
| `pkg/grafanadocs/mcp/` | `SPECS.md` §Tools, §Tool schemas |
| `pkg/grafanadocs/cli/` | `SPECS.md` §CLI adapter surface |
| `cmd/mcp-doc-server/` | `SPECS.md` §Config & runtime |
| `cmd/docs/` | `SPECS.md` §CLI adapter surface |
| `docs/design/demo/` | `docs/design/demo/README.md` (standalone demo docs) |
| `docs/design/research/` | `docs/design/research/README.md` (prior art and analysis) |

## Code validation paths

Paths the agent should check when validating documentation claims against code.

| What to validate | Where to look |
|-----------------|---------------|
| Tool names and schemas | `pkg/grafanadocs/mcp/server.go` (MCP tool definitions) |
| Core API surface (exported functions) | `pkg/grafanadocs/*.go` (non-test files) |
| Core data types (Entry, Product, etc.) | `pkg/grafanadocs/entry.go`, `pkg/grafanadocs/index.go`, `pkg/grafanadocs/fetch.go`, `pkg/grafanadocs/outline.go`, `pkg/grafanadocs/excerpt.go` |
| Security invariants (SSRF, rate limiting) | `pkg/grafanadocs/fetch.go` (allowlist, rate limiter, redirects) |
| Index parsing rules | `pkg/grafanadocs/index.go` |
| Search ranking algorithm | `pkg/grafanadocs/search.go` |
| Cleanup / markdown processing | `pkg/grafanadocs/cleanup.go` |
| CLI command wiring and flags | `pkg/grafanadocs/cli/command.go`, `cli/search.go`, `cli/get.go`, `cli/outline.go`, `cli/products.go` |
| Test scenarios vs TESTS.md | `pkg/grafanadocs/*_test.go`, `pkg/grafanadocs/mcp/server_test.go`, `pkg/grafanadocs/cli/cli_test.go` |

## Frontmatter and site conventions

- N/A — this repo has no published documentation site. Markdown files are spec docs (`SPECS.md`, `NOTES.md`, `TESTS.md`, `BENCHMARKS.md`) and research notes, not site-rendered pages.

## Conventions for agents

- **Spec-driven development:** Read `SPECS.md` and `NOTES.md` before modifying code. Code changes that alter contracts, invariants, or test scenarios must update the matching spec file in the same change.
- **Code marker:** Every `.go` file in the repo starts with `// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.`
- **NOTES.md is append-only and local to this repo:** Never edit or delete existing entries. Reverse a decision with an addendum plus a new numbered entry. Do not record how mcp-grafana or gcx currently cache or wrap the core — those live in `docs/design/demo/` and will go stale in NOTES.
- **Invariant numbering:** Invariants in `SPECS.md` are numbered I1–I26. Never remove or weaken one; add or annotate instead.
- **JSON output:** All MCP tool JSON uses `snake_case` keys (invariant I15). Core types carry no `json` tags by design — adapters own the wire format.
- **Security:** SSRF allowlist (I3), rate limiting (I9), body-size caps (I10), index scheme validation (I19). See `SPECS.md` §Invariants.
- **Vale or linter config location:** `golangci-lint` via pre-commit hook (runs automatically on `git commit`). No Vale config.

## Subsystem knowledge

- [`AGENTS.md`](../AGENTS.md) — SDD conventions and working rules for the repo root module.
- [`SPECS.md`](../SPECS.md) — complete behavioral contract (invariants I1–I26, tool schemas, core API, search ranking, cleanup rules).
- [`NOTES.md`](../NOTES.md) — append-only decision log for this repository (not a live snapshot of consumers).
- [`TESTS.md`](../TESTS.md) — test scenarios mapped to invariants.
- [`BENCHMARKS.md`](../BENCHMARKS.md) — performance and token-cost targets.

## Shared features across sub-products

The `pkg/grafanadocs` core is imported by two downstream consumers:

| Consumer | Repo | Integration file(s) |
|----------|------|---------------------|
| **mcp-grafana** | `grafana/mcp-grafana` | `tools/docs.go`, `tools/docs_unit_test.go` |
| **gcx** | `grafana/gcx` | `cmd/gcx/docs/` (command.go, search.go, get.go, outline.go, products.go) |

Changes to the core API surface (`pkg/grafanadocs`) require coordinated updates in both consumers. Changes to the MCP or CLI adapters (`pkg/grafanadocs/mcp/`, `pkg/grafanadocs/cli/`) do not affect consumers — they write their own wrappers.
