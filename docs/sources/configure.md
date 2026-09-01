---
title: Configure the server
menuTitle: Configure
description: Set DOCS_INDEX_URL and review the compiled-in timeouts, rate limits, body-size caps, and URL allowlist.
weight: 7
topicType: reference
versionDate: 2026-09-01
---

# Configure the server

One environment variable is configurable. Everything else is compiled into the binary.

## `DOCS_INDEX_URL`

Where the server loads the documentation index.

| Property | Value |
|----------|-------|
| **Default** | `https://grafana.com/llms-full.txt` |
| **Constraint** | HTTPS only. `file://`, `http://`, and other schemes are rejected before any network call. |

```bash
DOCS_INDEX_URL=https://example.com/custom-index.txt ./bin/mcp-doc-server
```

In an MCP client, pass it through the `env` field. Refer to [Install and connect](../install/).

## Built-in limits

These aren't flags. Changing them means changing the code.

### HTTP timeouts

| Operation | Timeout |
|-----------|---------|
| Document fetch | 30 seconds |
| Index load | 60 seconds |

### Rate limiting

Page fetches (`FetchDoc`) are rate-limited. Index loads use a separate HTTP client and aren't.

| Setting | Value |
|---------|-------|
| Maximum concurrent fetches | 5 |
| Minimum gap between requests | 200 ms |

Each request gets its own spaced time slot, so concurrent callers don't all fire at once.

### Body size limits

| Resource | Maximum size |
|----------|-------------|
| Documentation page | 2 MiB |
| Documentation index | 10 MiB |

Responses over these limits are truncated at read time.

### Scanner buffer

The index parser uses a 1 MiB line buffer (Go's default is 64 KiB) so long index entries aren't silently truncated.

## URL allowlist

`get_doc` and `get_doc_outline` fetch only URLs that match:

- **Scheme:** `https`
- **Host:** `grafana.com`
- **Path:** under `/docs/`

The check runs twice: on the original URL and again after the `.md` suffix is added. Redirects to any host other than `grafana.com` are blocked.

If a limit or allowlist rejection shows up at runtime, refer to the [Install and connect](../install/) troubleshooting table.

## Related resources

- [Install and connect](../install/)
- [Integrate the core library](../integrate/)
