# Research

Background research and prior-art notes informing the design of `mcp-doc-server` — a Go
MCP server that gives AI agents version-aware access to Grafana Labs product documentation.
See the repo `SPECS.md`, `NOTES.md`, `TESTS.md`, and `BENCHMARKS.md` for the committed
specs these notes feed into.

## Contents

- **[docs-mcp-server-use-cases.md](docs-mcp-server-use-cases.md)** — the common use cases
  for a docs MCP server (grounding/RAG, version-aware lookups, citations, token-efficient
  retrieval, discovery, deterministic agentic retrieval) and how `mcp-doc-server`'s
  invariants map onto them.
- **[docs-mcp-ecosystem.md](docs-mcp-ecosystem.md)** — prior art: a survey of the most
  popular and best-crafted docs MCP servers on GitHub (Context7, MarkItDown, Microsoft
  Learn, Elastic, GitMCP, mcpdoc, Grounded Docs, Rust Docs), the recurring patterns
  (remote HTTP, `llms.txt`, resolve→fetch, tool-count discipline), and the lessons,
  borrowings, and risks for `mcp-doc-server` — plus where it sits in the landscape.
- **[mcp-inspector.md](mcp-inspector.md)** — tooling: the MCP Inspector for testing and
  debugging our server (UI vs CLI mode, how to run it against our stdio binary, security
  considerations).
- **[mcp-go-sdk.md](mcp-go-sdk.md)** — tooling: the official MCP Go SDK as the likely SDK
  choice (packages, spec compatibility, minimal stdio server, schema-from-struct-tags), and
  the `SPECS.md` open questions it resolves.
- **[gcx-integration-patterns.md](gcx-integration-patterns.md)** — how `grafana/gcx`
  structures commands, output (`output.Options`/`format.Codec`), agent annotations, and
  external-library wrapping; the basis for the `pkg/grafanadocs/cli` adapter design.
- **[gcx-prototype-kickoff.md](gcx-prototype-kickoff.md)** — a paste-ready kickoff prompt
  for a fresh chat that mounts a `gcx docs` command on mcp-doc-server's core in a gcx
  branch (dependency wiring, gcx-native swaps, index-lifecycle decision, verification).

## Status

These are research notes, not contracts. Decisions that graduate from here belong in
`SPECS.md` (invariants/contracts) and `NOTES.md` (dated decision log) per the SDD workflow
in `AGENTS.md`.
