# CLI adapter patterns

Notes on the conventions that shape `pkg/grafanadocs/cli`, so a downstream CLI (such as
`grafana/gcx`) can mount the `docs` command group with minimal glue.

The adapter uses cobra and pflag, and imports only the core (`pkg/grafanadocs`). Output
formatting is intentionally kept behind a small local helper so callers can swap it for
their own output system without touching the command shells.

## opts pattern

Every command uses a small `opts` struct and follows the same three-step shape:

```go
type myOpts struct {
    IO output.Options
    // command-specific flags
}

func (o *myOpts) setup(flags *pflag.FlagSet) {
    o.IO.DefaultFormat("text")
    o.IO.RegisterCustomCodec("text", &myTableCodec{})
    o.IO.BindFlags(flags)
    // bind command-specific flags
}

func (o *myOpts) Validate() error {
    return o.IO.Validate()
}
```

Rules:

- `setup(flags)` runs at construction time; register codecs before `BindFlags`.
- `Validate()` is the first call in `RunE`; no I/O before it.
- `RunE` fetches data regardless of `-o` and encodes via the output helper.

## Command grammar

```
docs search <query>     find matching pages
docs get <url>          fetch a page or section
docs outline <url>      list headings
docs products           list product groups
```

## Output formats

- Built in: `json`, `yaml`, `agents` (compact JSON for machine consumption).
- Custom: a per-command text codec, registered before `BindFlags`.
- Default: `text` for humans, overridden to `agents` in agent-mode environments.

## Safety

All four commands are read-only. There are no confirmation prompts, no `--force`, no
`--dry-run`. Truncation and rate-limit warnings go to stderr; stdout stays parseable.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Fetch or parse failure, unknown doc |
| 2 | Bad flags, missing or invalid path |
| 5 | Cancelled (Ctrl+C) |
