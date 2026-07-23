# Contributing

Thanks for your interest in mcp-doc-server.

## Development

Requirements:

- Go 1.26 or later (matches `go.mod`).
- `golangci-lint` (version pinned in [`.golangci-lint-version`](.golangci-lint-version)).

Build and test:

```bash
go build ./...
go test -race -count=1 ./...
golangci-lint run
```

## Spec-driven development

This repository follows spec-driven development.
Specs are the source of truth; code is kept consistent with them.

- [`SPECS.md`](SPECS.md) — interface contracts and behavioral invariants.
- [`NOTES.md`](NOTES.md) — append-only, dated design-decision log.
- [`TESTS.md`](TESTS.md) — test scenarios as Setup / Action / Assertion.
- [`BENCHMARKS.md`](BENCHMARKS.md) — performance targets.
- [`AGENTS.md`](AGENTS.md) — the SDD convention.

A code change that alters a contract, invariant, test scenario, or performance target must update the matching spec file in the same change.

## Design and architecture

See [`docs/design/`](docs/design/) for architecture, research notes, and demo scenarios.

## Security

Do not open a public issue for security vulnerabilities.
See [SECURITY.md](SECURITY.md) for the reporting process.
