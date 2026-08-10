# BENCHMARKS: mcp-doc-server

Performance and token-cost targets. **The values below are provisional placeholders**,
not decided commitments; they exist to capture intent and will be confirmed (or replaced)
as the corresponding **Open questions** in `SPECS.md` close. Once a target is ratified,
treat it as authoritative: do not lower it (that is a regression); add new rows for new
benchmarks.

## Metric Targets (provisional)

| Benchmark / Metric            | Target            | Notes                                                        |
| ----------------------------- | ----------------- | ----------------------------------------------------------- |
| Index parse (`llms-full.txt`) | < 250 ms          | One-time on startup / refresh; ~1.3 MB, ~6,879 entries.     |
| `search_docs` latency         | < 10 ms           | In-memory ranking over the parsed index, warm.              |
| `get_doc` cached fetch        | < 5 ms            | Served from the TTL cache (no network).                     |
| `search_docs` response size   | <= ~600 tokens    | Default `limit=5`, one-line descriptions (I7).              |
| `get_doc_outline` size        | <= ~400 tokens    | Heading outline only.                                        |
| `get_doc` default slice       | <= ~2,000 tokens  | Size-guarded default; full page only on explicit paging (I4).|
| Added inference cost          | 0                 | No server-side LLM/embedding calls in v1 (I2).             |

## Notes

- Token figures are budgets that enforce the cost invariants in `SPECS.md`, not raw
  Go benchmarks; verify with representative pages (e.g. `tempo/latest/configuration.md`).
- Latency targets assume a warm in-memory index and cache; cold startup is dominated by
  the one-time index fetch + parse.
