#!/usr/bin/env bash
# mcp-doc-server demo — runs the standalone CLI to showcase all 4 tools.
# Requires: go build -o bin/docs ./cmd/docs/
set -euo pipefail

BOLD='\033[1m'
DIM='\033[2m'
CYAN='\033[36m'
GREEN='\033[32m'
RESET='\033[0m'

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "${SCRIPT_DIR}/../../.." && pwd)"
DOCS="${DOCS_BIN:-${REPO_ROOT}/bin/docs}"

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
if [[ ! -x "${DOCS}" ]]; then
  echo -e "${DIM}Building ${DOCS}...${RESET}"
  (
    cd -- "${REPO_ROOT}"
    go build -o ./bin/docs ./cmd/docs
  )
  echo ""
fi

# ╔══════════════════════════════════════════════════════════════════╗
# ║  1. DISCOVER — What products have docs?                         ║
# ╚══════════════════════════════════════════════════════════════════╝
header "1. list_products" \
  "Discover all Grafana product documentation groups and their page counts."

run "$DOCS" products

# ╔══════════════════════════════════════════════════════════════════╗
# ║  2. SEARCH — Find relevant pages                                ║
# ╚══════════════════════════════════════════════════════════════════╝
header "2. search_docs" \
  "TF-IDF ranked search across all products — rare terms score higher."

run "$DOCS" search "rate limiting"

echo -e "${DIM}Filter to a single product:${RESET}"
run "$DOCS" search "configuration" --product tempo --limit 3

# ╔══════════════════════════════════════════════════════════════════╗
# ║  3. OUTLINE — Cheap heading scan before fetching                ║
# ╚══════════════════════════════════════════════════════════════════╝
header "3. get_doc_outline" \
  "Get the structure of a page so you know which section to fetch."

run "$DOCS" outline https://grafana.com/docs/grafana/latest/alerting/

# ╔══════════════════════════════════════════════════════════════════╗
# ║  4. GET — Bounded, section-aware retrieval                      ║
# ╚══════════════════════════════════════════════════════════════════╝
header "4. get_doc (full page, bounded)" \
  "Fetches cleaned markdown — bounded to first 30 lines to show paging."

run "$DOCS" get https://grafana.com/docs/grafana/latest/alerting/ --limit 30

echo -e "${DIM}Extract a specific section by heading name:${RESET}"
run "$DOCS" get https://grafana.com/docs/grafana/latest/alerting/ --section "Overview"

# ╔══════════════════════════════════════════════════════════════════╗
# ║  5. JSON output — for programmatic / agent consumption          ║
# ╚══════════════════════════════════════════════════════════════════╝
header "5. JSON output mode" \
  "All commands support -o json for agent consumption."

run "$DOCS" search "tracing instrumentation" -o json --limit 2

echo ""
echo -e "${BOLD}${GREEN}Demo complete.${RESET} See docs/design/demo/scenarios/ for full agent workflows."
