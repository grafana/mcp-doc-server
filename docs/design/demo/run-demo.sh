#!/usr/bin/env bash
# mcp-doc-server demo: runs the standalone CLI to showcase all 4 tools.
# Requires: go build -o bin/docs ./cmd/docs/
set -euo pipefail

BOLD='\033[1m'
DIM='\033[2m'
CYAN='\033[36m'
GREEN='\033[32m'
RESET='\033[0m'

DOCS="${DOCS_BIN:-./bin/docs}"

header() {
  echo ""
  echo -e "${BOLD}${CYAN}━━━ $1 ━━━${RESET}"
  echo -e "${DIM}$2${RESET}"
  echo ""
}

run() {
  echo -e "${GREEN}\$ $*${RESET}"
  "$@"
  echo ""
}

# --- Build if needed ---
if [[ ! -x "$DOCS" ]]; then
  echo -e "${DIM}Building bin/docs...${RESET}"
  go build -o bin/docs ./cmd/docs/
  echo ""
fi

# ╔══════════════════════════════════════════════════════════════════╗
# ║  1. DISCOVER: What products have docs?                         ║
# ╚══════════════════════════════════════════════════════════════════╝
header "1. list_products" \
  "Discover all Grafana product documentation groups and their page counts."

run "$DOCS" products

# ╔══════════════════════════════════════════════════════════════════╗
# ║  2. SEARCH: Find relevant pages                                ║
# ╚══════════════════════════════════════════════════════════════════╝
header "2. search_docs" \
  "TF-IDF ranked search across all products: rare terms score higher."

run "$DOCS" search "alerting rules"

echo -e "${DIM}Filter to a single product:${RESET}"
run "$DOCS" search "traceql query" --product tempo --limit 3

# ╔══════════════════════════════════════════════════════════════════╗
# ║  3. OUTLINE: Cheap heading scan before fetching                ║
# ╚══════════════════════════════════════════════════════════════════╝
header "3. get_doc_outline" \
  "Get the structure of a page so you know which section to fetch."

run "$DOCS" outline https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/

# ╔══════════════════════════════════════════════════════════════════╗
# ║  4. GET: Bounded, section-aware retrieval                      ║
# ╚══════════════════════════════════════════════════════════════════╝
header "4. get_doc (full page, bounded)" \
  "Fetches cleaned markdown: Tempo Configure is ~3000 lines; page with offset/limit."

run "$DOCS" get https://grafana.com/docs/tempo/latest/configuration/ --offset 80 --limit 80

echo -e "${DIM}Extract a specific section by heading name:${RESET}"
run "$DOCS" get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/ --section "Comparison operators"

# ╔══════════════════════════════════════════════════════════════════╗
# ║  5. JSON output: for programmatic / agent consumption          ║
# ╚══════════════════════════════════════════════════════════════════╝
header "5. JSON output mode" \
  "All commands support -o json for agent consumption."

run "$DOCS" search "alerting rules" -o json --limit 2

echo ""
echo -e "${BOLD}${GREEN}Demo complete.${RESET} See demo/scenarios/ for full agent workflows."
