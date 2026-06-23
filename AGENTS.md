# AGENTS — spec-driven development for hack-doc-server

This repo follows **spec-driven development (SDD)**, adapted from
[mattdurham/bob](https://github.com/mattdurham/bob). Specs are the source of truth;
code is kept consistent with them.

## The spec suite

A **spec-driven module** is a directory that carries these files. The repo root is one
such module; each `internal/*` package becomes one as it is built.

- **`SPECS.md`** — interface contracts and behavioral **invariants**. Invariants are
  precise, machine-checkable claims. Never remove or weaken an existing invariant; add
  or annotate instead.
- **`NOTES.md`** — **append-only**, dated design-decision log. Entry format:
  ```markdown
  ## N. Title
  *Added: YYYY-MM-DD*
  **Decision:** ...
  **Rationale:** ...
  **Consequence:** ...
  ```
  Reverse a decision with an `*Addendum (YYYY-MM-DD):*` line on the original entry plus a
  new entry — never by editing or deleting history.
- **`TESTS.md`** — scenarios as **Setup / Action / Assertion**.
- **`BENCHMARKS.md`** — a **Metric Targets** table; don't lower targets (regression).

## Code marker

Every `.go` file in a spec-driven module starts with:

```go
// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.
```

## Working rules

1. **Read specs first.** Before changing a module, read its `SPECS.md` and `NOTES.md`.
2. **Specs and code move together.** A code change that alters a contract, invariant,
   test scenario, or perf target must update the matching spec file in the same change.
3. **New public API ⇒ new `SPECS.md` contract** (and usually a `TESTS.md` scenario).
4. **New significant decision ⇒ new `NOTES.md` entry.**
5. **Audit = read-only.** Verifying code-vs-spec drift reports issues; it does not invent
   new functionality.
