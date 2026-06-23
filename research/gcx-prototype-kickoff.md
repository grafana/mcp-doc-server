# Kickoff: Prototype a `gcx docs` command backed by hack-doc-server's core

Paste the prompt below into a fresh chat opened on the **gcx** repo. It is self-contained.

See also [`gcx-integration-patterns.md`](gcx-integration-patterns.md) and `NOTES.md`
entries 16–17 for the full integration mapping and the latest gcx compatibility check.

---

## Task: Prototype a `gcx docs` command backed by hack-doc-server's core

Create a **branch in gcx** that mounts a new `docs` command group powered by the
`hack-doc-server` documentation core. This is a proof-of-concept integration, not a
production merge — keep it on a branch.

### Repos

- Working repo: `/Users/kim-nylander/Repositories/gcx` (module `github.com/grafana/gcx`)
- Dependency: `/Users/kim-nylander/Repositories/hack-doc-server`
  (module `github.com/grafana/hack-doc-server`) — DO NOT modify this repo.

### Wiring the dependency (it's unpublished)

Add a local replace in gcx's `go.mod`, then `go mod tidy`:

    require github.com/grafana/hack-doc-server v0.0.0
    replace github.com/grafana/hack-doc-server => /Users/kim-nylander/Repositories/hack-doc-server

Import ONLY the core: `github.com/grafana/hack-doc-server/pkg/grafanadocs`.
Do NOT import `pkg/grafanadocs/cli` — re-skin it with gcx's own output system instead.

### What the core gives you (plain Go, no framework deps)

- `grafanadocs.LoadIndex(ctx, grafanadocs.DefaultIndexURL) (*Index, error)`
- `grafanadocs.Search(idx *Index, query string, grafanadocs.SearchOpts{Product, Limit}) []Entry`
- `(*Index).Products() []Product`
- `grafanadocs.FetchDoc(ctx, url) (*Doc, error)` — allowlisted to grafana.com /docs/, rate-limited
- `grafanadocs.Excerpt(doc, grafanadocs.ExcerptOpts{Section, Offset, Limit}) ExcerptResult`
- `grafanadocs.Outline(doc) []Heading`
- Types: `Entry{Title,URL,Description,Product}`, `Product{Name,Count}`, `Heading{Level,Text,Line}`

### Reference implementation to port

`hack-doc-server/pkg/grafanadocs/cli/{command,search,get,outline,products}.go` already
follows gcx's `opts` + `setup(flags)` + `Validate()` pattern. Use it as the blueprint.

Command grammar to reproduce:

    gcx docs search <query> [--product=...] [--limit=5] [-o ...]
    gcx docs get <url> [--section=...] [--offset=0] [--limit=0] [-o ...]
    gcx docs outline <url> [-o ...]
    gcx docs products [-o ...]

### gcx-native swaps (the only real work)

1. Replace the reference adapter's local `output` helper with gcx's
   `output.Options` (package `internal/output`): `DefaultFormat("text")`,
   `RegisterCustomCodec("text", &codec{})`, `BindFlags(flags)`, `Validate()`.
   This gives you json/yaml/agents + `--json` field selection + `--jq` for free.
2. Each text codec must implement gcx's `internal/format.Codec`
   (`Encode(w,v) error` + `Format() Format` + `Decode()` returning an "unsupported"
   error). Mirror `cmd/gcx/version/command.go`'s `versionTextCodec`. The reference
   codecs' `Encode` bodies (tabwriter tables; raw markdown for `get`) carry over as-is.
3. Place commands under `cmd/gcx/docs/`. Follow the canonical options pattern:
   `setup` registers codecs BEFORE `BindFlags`; `Validate()` is the first call in `RunE`;
   `RunE` fetches data regardless of `-o` and encodes via the options.

### Index lifecycle — DECIDE THIS FIRST

The reference adapter takes a pre-loaded `*grafanadocs.Index`. gcx must decide where that
load happens. Options: lazy-load on first `docs` subcommand run (recommended — avoids a
network fetch on unrelated commands / `--help`), vs. eager at root. Pick one, document the
rationale, and keep `get`/`outline` working without the index (they only need `FetchDoc`).
Note `DOCS_INDEX_URL` as an override, matching hack-doc-server.

### Mounting + agent metadata

- Mount in `cmd/gcx/root/command.go` via `rootCmd.AddCommand(docs.Command(...))`.
- Add agent annotations in `internal/agent/command_annotations.go`, mirroring existing
  entries' shape and cost taxonomy. Suggested: `docs search` small, `docs get` medium,
  `docs outline` small, `docs products` small, with short llm hints.

### Compliance + verification (gcx rules)

- Check work against the compliance hierarchy: CONSTITUTION (invariants), VISION
  (does a docs command belong in gcx? flag if unsure), DESIGN (output model, exit codes,
  safety), ARCHITECTURE (package placement). Read gcx's `AGENTS.md` first.
- Before finishing: `GCX_AGENT_MODE=false mise run all` (lint + tests + build + docs).
  Also `GCX_AGENT_MODE=false mise run reference` if commands/flags changed, and the
  doc-maintenance gate (`docs/reference/doc-maintenance.md`).
- Add table-driven tests for the new commands capturing stdout (see the reference repo's
  `pkg/grafanadocs/cli/cli_test.go` for the approach: build the command, set args, capture
  out/err, assert; test `get`/`outline` wiring via the allowlist rejection so no network).

### Deliverable

A gcx branch with `cmd/gcx/docs/`, the `go.mod` replace, agent annotations, tests, and a
short note in the PR/commit describing the index-lifecycle decision and that the core is
consumed via a local replace (would point to a tagged module before any real merge).

### Background context (don't re-derive)

hack-doc-server compatibility with gcx was verified on 2026-06-23: cobra v1.10.2 matches;
pflag aligned to v1.0.10; gcx's options type was renamed `cmdio.Options` → `output.Options`
(pkg `internal/output`); `--jq`/`--json` are additive. See hack-doc-server's
`research/gcx-integration-patterns.md` and `NOTES.md` entries 16–17 for the full mapping.

One consideration: the core's `Entry`/`Heading`/`Product` types have no JSON tags (Go field
names serialize capitalized). gcx's `--json` field selection works regardless, but if gcx
prefers lowercase keys, wrap them in view structs in `cmd/gcx/docs/`.
