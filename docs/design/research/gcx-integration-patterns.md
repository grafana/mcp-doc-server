# gcx Integration Patterns

Research on how `grafana/gcx` structures commands, handles output, and wraps external
libraries — gathered to inform the `pkg/grafanadocs/cli` cobra adapter design.

> **Update (2026-06-23):** Re-verified against a fresh `gcx` pull. The output options
> type is package `internal/output` as `output.Options` (older notes called it
> `cmdio.Options`). The `DefaultFormat`/`RegisterCustomCodec`/`BindFlags`/`Validate`
> methods are unchanged. New additive flags `--jq` and `--json` (field selection) are
> wired in by `BindFlags`, so a ported `gcx docs` would gain them for free. cobra is
> v1.10.2 (matches ours); pflag is v1.0.10 (we bumped to match). No breaking changes
> to the patterns below.

## Command Mounting

Commands are mounted in `cmd/gcx/root/command.go` via `rootCmd.AddCommand(...)`.
After the full tree is assembled, `agent.ApplyAnnotations(rootCmd)` runs once to fill
in agent metadata.

## The opts Pattern (canonical form)

Every non-trivial command uses:

```go
type myOpts struct {
    IO      output.Options
    // command-specific flags
}

func (o *myOpts) setup(flags *pflag.FlagSet) {
    o.IO.DefaultFormat("text")
    o.IO.RegisterCustomCodec("text", &myTableCodec{})
    o.IO.RegisterCustomCodec("wide", &myWideCodec{})
    o.IO.BindFlags(flags)
    // bind command-specific flags
}

func (o *myOpts) Validate() error {
    return o.IO.Validate()
}
```

Rules:
- `setup(flags)` called at construction time; register codecs before `BindFlags`
- `Validate()` is the first call in `RunE`; no I/O before it
- `RunE` fetches all data regardless of `-o`; encodes via `opts.IO.Encode()`

## Output System (`internal/output`)

`output.Options` provides:
- Built-in codecs: `json`, `yaml`, `agents` (free via `BindFlags()`)
- Custom codecs: `RegisterCustomCodec(name, format.Codec)`
- Default format: `DefaultFormat("text")` — overridden to `agents` in agent mode
- Output flags: `-o/--output`, `--json` (field selection/discovery)
- Pipe awareness: `IsPiped`, `NoTruncate`
- Diagnostics: `EmitHint`, `EmitWarn`, `EmitNote` (stderr; JSONL in agent mode)

The codec interface (`internal/format`):
```go
type Codec interface {
    Encoder
    Decoder
    Format() Format
}
```

Key design rules:
- Format-agnostic fetching — never gate API calls on `OutputFormat`
- Codec registration in `setup`, not `RunE`
- Data-display commands default to `text`/`table`, not `json`
- `agents` codec: compact JSON below 100 KiB; spills large payloads to temp file

## External Library Wrapping Precedent

`cmd/gcx/instrumentation/check/` wrapping `github.com/grafana/otel-checker`:

| Layer | File | Responsibility |
|-------|------|----------------|
| Cobra + flags | `check/command.go` | opts, flag binding, library input mapping (`toCommands()`), error translation |
| Execution | `check/run.go` | Thin wrapper with injectable `checker` func for tests |
| Output | `check/codec.go` | gcx-owned table rendering of library result types |
| Mount | `instrumentation/command.go` | `check.Command()` added to group |

Pattern:
1. Map gcx flags → library input struct
2. Validate library input in `Validate()`, translate sentinel errors
3. Call library via testable seam (`runWith` + func type)
4. Normalize output (nil slices → empty for JSON)
5. Register custom codecs for library result types
6. Do not use the library's own UI

## Agent Annotations (`internal/agent`)

Annotation keys:
- `agent.token_cost` — `"small"`, `"medium"`, `"large"` (required on all non-hidden leaves)
- `agent.llm_hint` — scoping hint for agents (required when cost is medium/large)
- `agent.required_scope`, `agent.required_role`, `agent.required_action` — auth metadata
- `agent.skill` — comma-joined related Agent Skill names

Annotations live in a centralized registry (`internal/agent/command_annotations.go`)
keyed by full command path (e.g., `"gcx instrumentation check"`).
`ApplyAnnotations()` fills missing keys without overwriting inline values.

Consistency enforced by `cmd/gcx/root/consistency_test.go`.

## CLI Grammar for `gcx docs`

gcx uses `$AREA $NOUN $VERB` or `$AREA $VERB` (when no meaningful noun).

For docs (read-only lookup, not CRUD on Grafana resources):
```
gcx docs search <query>     -- $AREA $VERB pattern
gcx docs get <url>          -- $AREA $VERB pattern
gcx docs outline <url>      -- $AREA $VERB (extension)
gcx docs products           -- leaf (always lists)
```

Precedents: `gcx dashboards search`, `gcx logs query`, `gcx datasources list`.

## Safety

All four commands are read-only:
- No confirmation prompts, `--force`, or `--dry-run`
- No mutation summaries
- Truncation/limit warnings go to stderr

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Fetch/parse failures, unknown doc |
| 2 | Bad flags, missing/invalid path |
| 5 | Cancelled (Ctrl+C) |

## Sources

- `cmd/gcx/root/command.go` — command tree assembly
- `cmd/gcx/instrumentation/check/` — external library wrapping pattern
- `cmd/gcx/version/command.go` — minimal inline command
- `internal/agent/annotations.go` — annotation keys
- `internal/agent/command_annotations.go` — centralized registry
- `docs/architecture/cli-layer.md` — canonical opts pattern
- `docs/design/output.md` — output model rules
- `docs/design/agent-mode.md` — agent mode behavior
- `DESIGN.md` — CLI UX grammar, exit codes, safety
