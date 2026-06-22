// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

// Package grafanadocs provides version-aware access to Grafana Labs product
// documentation. It parses the official llms-full.txt index, offers
// deterministic search, and fetches individual doc pages with cleanup and
// sectioning — all without server-side inference.
//
// This package is the public core intended for import by grafana/mcp-grafana,
// grafana/gcx, and the standalone mcp-doc-server. It has zero framework
// dependencies (no MCP SDK, no cobra).
package grafanadocs
