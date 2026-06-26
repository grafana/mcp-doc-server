---
title: Configuration
menuTitle: Configure
description: Environment variables, built-in limits, and the URL allowlist that control mcp-doc-server behavior.
weight: 5
topicType: reference
versionDate: 2026-06-25
---

# Configuration

mcp-doc-server has one configurable setting; everything else is a built-in limit compiled into the binary.

## Configurable settings

One environment variable controls where the server loads its documentation index from.

### `DOCS_INDEX_URL`

The URL of the documentation index to load.

| Property | Value |
|----------|-------|
| **Default** | `https://grafana.com/llms-full.txt` |
| **Constraint** | Must use HTTPS. `file://`, `http://`, and other schemes are rejected before any network call. |

Set it as an environment variable:

```bash
DOCS_INDEX_URL=https://example.com/custom-index.txt ./bin/mcp-doc-server
```

Or, for a Model Context Protocol (MCP) client, pass it through the `env` field of your server configuration. Refer to [Install and configure](../install/).

## Built-in limits

These values are compiled into the binary. Changing them requires a code change—they aren't environment variables or flags. They're documented here so you know the server's behavior.

### HTTP timeouts

| Operation | Timeout |
|-----------|---------|
| Document fetch | 30 seconds |
| Index load | 60 seconds |

### Rate limiting

All outbound requests to `grafana.com` are rate-limited:

| Setting | Value |
|---------|-------|
| Maximum concurrent fetches | 5 |
| Minimum gap between requests | 200ms |

The limiter reserves a unique, spaced time slot for each request, so concurrent callers don't all fire at once.

### Body size limits

| Resource | Maximum size |
|----------|-------------|
| Documentation page | 2 MiB |
| Documentation index | 10 MiB |

Responses over these limits are truncated at read time to prevent out-of-memory conditions.

### Scanner buffer

The index parser uses a 1 MiB line buffer (versus Go's 64 KiB default) so long index entries aren't silently truncated.

## URL allowlist

`get_doc` and `get_doc_outline` only fetch URLs that match:

- **Scheme:** `https`
- **Host:** `grafana.com`
- **Path:** under `/docs/`

The check runs twice—on the original URL and again after the `.md` suffix is added—to prevent path manipulation. Redirects to any non-`grafana.com` host are blocked.

## Related resources

- [Install and configure](../install/)—passing `DOCS_INDEX_URL` through client configuration, plus troubleshooting startup and index errors
- [Integrate the core library](../integrate/)—how these limits apply when you import the core
