# Tooling: MCP Inspector

Notes on the [MCP Inspector](https://github.com/modelcontextprotocol/inspector) — the
official visual testing and debugging tool for MCP servers — and how it fits into
developing and verifying `hack-doc-server`.

## What it is

A developer tool for testing and debugging MCP servers. It has two components:

- **MCP Inspector Client (MCPI)** — a React web UI for interactive testing/debugging
  (default port `6274`).
- **MCP Proxy (MCPP)** — a Node.js protocol bridge that connects the web UI to an MCP
  server over stdio, SSE, or streamable-http (default port `6277`). It is *not* a network
  intercepting proxy; it acts as an MCP client + HTTP server so the browser can talk to
  servers on any transport.

Requires Node.js `^22.7.5`. Current release: `0.22.0` (Jun 2026).

## Two modes

- **UI mode** — visual, form-based tool invocation, resource browsing, request history,
  and error visualization. Best for interactive development.
- **CLI mode** (`--cli`) — programmatic/scriptable, JSON output. Ideal for automation,
  CI/CD, and tight feedback loops with coding assistants.

## How we use it with hack-doc-server

`hack-doc-server` is a Go **stdio** server, so the relevant pattern is launching the
inspector against the built binary:

```bash
# UI mode against the built binary
npx @modelcontextprotocol/inspector ./hack-doc-server

# Pass env vars / args (everything after the binary goes to our server)
npx @modelcontextprotocol/inspector -e LOG_LEVEL=debug -- ./hack-doc-server --some-flag
```

CLI mode is the most useful for scripted verification of our four tools
(`search_docs`, `get_doc`, `get_doc_outline`, `list_products`):

```bash
# List the tools our server exposes
npx @modelcontextprotocol/inspector --cli ./hack-doc-server --method tools/list

# Exercise search_docs
npx @modelcontextprotocol/inspector --cli ./hack-doc-server \
  --method tools/call --tool-name search_docs --tool-arg query="alerting"

# Fetch a bounded doc slice
npx @modelcontextprotocol/inspector --cli ./hack-doc-server \
  --method tools/call --tool-name get_doc \
  --tool-arg 'options={"url":"https://grafana.com/docs/...","section":"..."}'
```

This gives a fast, deterministic feedback loop and is a natural fit for the SDD workflow:
CLI invocations map directly onto `TESTS.md` scenarios (Setup / Action / Assertion) and
can run in CI.

## Security considerations (important)

- The proxy can **spawn local processes** and connect to arbitrary MCP servers, so it must
  not be exposed to untrusted networks.
- Auth is **on by default**: a random session token is printed at startup and required as a
  Bearer token. The browser is auto-opened with the token pre-filled.
- `DANGEROUSLY_OMIT_AUTH=true` disables auth and is strongly discouraged — it enabled
  CVE-2025-49596 (RCE reachable even via a malicious web page). Do not use it.
- Both client and proxy bind to `localhost` only by default; `HOST=0.0.0.0` overrides this
  (trusted networks only). `Origin` header validation guards against DNS rebinding;
  additional origins via `ALLOWED_ORIGINS`.

## Relevance to this project

- **Primary verification tool** for the MCP layer (`internal/mcp`, `cmd/hack-doc-server`)
  before wiring the server into a real client like Cursor.
- **CLI mode in CI** can assert tool schemas and outputs, supporting the reproducibility
  goal in `SPECS.md` (tool contracts hold against a live server, not just on paper).
- Elastic's docs MCP page also points at the Inspector for testing
  (`npx @modelcontextprotocol/inspector --url ...`), confirming it as the de facto way to
  validate a docs MCP server regardless of transport.

## Source

- <https://github.com/modelcontextprotocol/inspector>
