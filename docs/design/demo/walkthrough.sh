#!/usr/bin/env bash
# ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
# ┃  mcp-doc-server: Automated Video Walkthrough                 ┃
# ┃                                                                ┃
# ┃  A self-running demo for screen capture. Hit Enter to advance  ┃
# ┃  each step, or run with --auto for timed delays.               ┃
# ┃                                                                ┃
# ┃  Usage:                                                        ┃
# ┃    ./docs/design/demo/walkthrough.sh          # interactive     ┃
# ┃    ./docs/design/demo/walkthrough.sh --auto   # auto-paced      ┃
# ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
DOCS="$REPO_ROOT/bin/docs"
GCX="${GCX_BIN:-/tmp/gcx-demo}"
AUTO=false
DELAY=6

[[ "${1:-}" == "--auto" ]] && AUTO=true

# ── Colors ──────────────────────────────────────────────────────
BOLD='\033[1m'
DIM='\033[2m'
CYAN='\033[36m'
GREEN='\033[32m'
YELLOW='\033[33m'
MAGENTA='\033[35m'
WHITE='\033[97m'
RESET='\033[0m'

# ── Helpers ─────────────────────────────────────────────────────
pause() {
  if $AUTO; then
    sleep "$DELAY"
  else
    echo -e "${DIM}  ↵ press Enter to continue${RESET}"
    read -r
  fi
}

long_pause() {
  if $AUTO; then
    sleep $((DELAY + 2))
  else
    echo -e "${DIM}  ↵ press Enter to continue${RESET}"
    read -r
  fi
}

title_card() {
  clear 2>/dev/null || true
  echo ""
  echo ""
  echo -e "${BOLD}${CYAN}  ╔══════════════════════════════════════════════════════════════╗${RESET}"
  echo -e "${BOLD}${CYAN}  ║${RESET}${BOLD}${WHITE}  $1$(printf '%*s' $((56 - ${#1})) '')${RESET}${BOLD}${CYAN}║${RESET}"
  echo -e "${BOLD}${CYAN}  ╚══════════════════════════════════════════════════════════════╝${RESET}"
  echo ""
  echo -e "  ${DIM}$2${RESET}"
  echo ""
}

section() {
  echo ""
  echo -e "${BOLD}${YELLOW}  ▸ $1${RESET}"
  echo -e "  ${DIM}$2${RESET}"
  echo ""
}

narrate() {
  echo -e "  ${WHITE}$1${RESET}"
}

run_cmd() {
  echo -e "  ${GREEN}\$ $*${RESET}"
  echo ""
  "$@" 2>/dev/null | sed 's/^/  /' || true
  echo ""
}

# ── Build binaries if needed ────────────────────────────────────
if [[ ! -x "$DOCS" ]]; then
  echo -e "${DIM}  Building bin/docs...${RESET}"
  (cd "$REPO_ROOT" && go build -o bin/docs ./cmd/docs/)
fi
if [[ ! -x "$GCX" ]]; then
  echo -e "${DIM}  Building gcx (this takes a moment)...${RESET}"
  gcx_repo="$REPO_ROOT/../gcx"
  if [[ -d "$gcx_repo" ]]; then
    (cd "$gcx_repo" && go build -buildvcs=false -o "$GCX" ./cmd/gcx/) || true
  else
    echo -e "${YELLOW}  gcx repo not found at $gcx_repo: gcx sections will be skipped${RESET}"
    GCX=""
  fi
fi

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 1: What is mcp-doc-server?
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "mcp-doc-server" \
  "Docs retrieval for AI agents: current, cheap, precise"

narrate "Why not just prompt with docs? Three reasons:"
echo ""
narrate "  1. Always current: fetches live docs, not stale training data"
narrate "  2. Token-efficient: one section costs ~200 tokens, not 10,000"
narrate "  3. Precise: search → outline → extract, like a human would"
echo ""
narrate "4 tools, 2,000+ Grafana docs pages. Let's see it."

long_pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 2: The standalone CLI
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Part 1: Standalone CLI" \
  "The docs command: search, get, outline, products"

# -- list_products --
section "list_products" \
  "What documentation exists? The index groups pages by product."

run_cmd "$DOCS" products

narrate "An agent can discover the entire product landscape before drilling in."
pause

# -- search_docs --
section "search_docs" \
  "TF-IDF ranked search: rare terms score higher, title matches get 3x weight."

narrate "Scenario: A user asks how to construct a TraceQL query."
echo ""
run_cmd "$DOCS" search "traceql query" --product tempo --limit 5

narrate "First hit is Construct a TraceQL query. Now let's look inside it."
pause

# -- get_doc_outline --
section "get_doc_outline" \
  "Cheap heading scan: map a page's structure before fetching content."

run_cmd "$DOCS" outline https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/

narrate "Now the agent knows exactly which section to fetch."
pause

# -- get_doc (section) --
section "get_doc: section extraction" \
  "Fetch just one section instead of the whole page. 10x fewer tokens."

run_cmd "$DOCS" get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/ \
  --section "Comparison operators"

narrate "Clean markdown. No HTML chrome. No shortcodes. Code blocks preserved."
pause

# -- get_doc (paging) --
section "get_doc: offset/limit paging" \
  "For a long page (Tempo Configure is ~3000 lines), page with offset and limit."

run_cmd "$DOCS" get https://grafana.com/docs/tempo/latest/configuration/ --offset 80 --limit 80

narrate "The agent sees total_lines and returned_range: it knows there's more."
pause

# -- JSON output --
section "JSON output" \
  "All commands support -o json for programmatic / agent consumption."

run_cmd "$DOCS" search "alerting rules" -o json --limit 2

pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 3: gcx integration
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Part 2: gcx Integration" \
  "gcx imports the core and mounts it as 'gcx docs' with styled output"

if [[ -n "$GCX" && -x "$GCX" ]]; then

  narrate "gcx is a unified CLI for managing Grafana resources."
  narrate "It now has docs commands alongside dashboards, alerting, datasources, etc."
  echo ""

  section "gcx docs search" \
    "Same search, styled with Grafana Neon Dark theme."

  run_cmd "$GCX" docs search "traceql query" --product tempo --limit 3

  pause

  section "gcx docs outline" \
    "Heading outline with styled table output."

  run_cmd "$GCX" docs outline https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/

  pause

  section "gcx docs get: section extraction" \
    "Fetch a specific section, rendered as clean markdown."

  run_cmd "$GCX" docs get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/ \
    --section "Comparison operators" --limit 20

  pause

  section "gcx docs list-products" \
    "Indexed product groups with entry counts. Catalog leaves use list-<subject>."

  run_cmd "$GCX" docs list-products

  pause

  section "gcx docs list-links" \
    "Curated canonical URLs: offline, no index fetch."

  run_cmd "$GCX" docs list-links

  pause

  narrate "gcx imports only pkg/grafanadocs (the core): not the MCP or CLI adapter."
  narrate "It writes its own command layer with output.Options, styled tables, and agent mode."
  echo ""

else
  narrate "(gcx binary not available: skipping gcx demo)"
  echo ""
fi

pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 4: mcp-grafana integration
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Part 3: mcp-grafana Integration" \
  "Two doc tools registered alongside the other Grafana MCP tools"

narrate "mcp-grafana is the official Grafana MCP server."
narrate "It folds the four Docs MCP tools into two wrappers (fewer schema tokens):"
echo ""
echo -e "  ${CYAN}search_docs${RESET}       Search docs; omit query to list products"
echo -e "  ${CYAN}get_doc${RESET}           Fetch a page; outline_only=true for headings"
echo ""
narrate "These appear alongside dashboards, alerting, datasources, incidents, etc."
narrate "The docs category is on by default; --disable-docs turns them off."

pause

section "Integration architecture" \
  "Three consumers, one core: each writes its own idiomatic wrapper."

echo -e "  ${BOLD}${WHITE}pkg/grafanadocs/${RESET}              ${DIM}← Public core (zero framework deps)${RESET}"
echo ""
echo -e "  ${MAGENTA}Consumer 1: mcp-doc-server${RESET}   ${DIM}← MCP adapter on mark3labs/mcp-go${RESET}"
echo -e "    ${DIM}pkg/grafanadocs/mcp/server.go: raw AddTool() registration${RESET}"
echo ""
echo -e "  ${MAGENTA}Consumer 2: gcx${RESET}               ${DIM}← own command layer (not the CLI adapter)${RESET}"
echo -e "    ${DIM}cmd/gcx/docs/: output.Options, agent annotations, Neon theme${RESET}"
echo ""
echo -e "  ${MAGENTA}Consumer 3: mcp-grafana${RESET}        ${DIM}← MustTool wrappers with jsonschema${RESET}"
echo -e "    ${DIM}tools/docs.go: search_docs + get_doc (outline_only / omit query)${RESET}"
echo ""

narrate "All three import pkg/grafanadocs directly. None import each other's adapters."

pause

section "The two-server complement" \
  "mcp-grafana acts on a live instance. mcp-doc-server provides the docs."

echo ""
echo -e "  ${BOLD}${WHITE}┌─────────────────────┬───────────────────────────────────┐${RESET}"
echo -e "  ${BOLD}${WHITE}│  mcp-grafana        │  mcp-doc-server                  │${RESET}"
echo -e "  ${BOLD}${WHITE}│  (live instance)     │  (documentation)                  │${RESET}"
echo -e "  ${BOLD}${WHITE}├─────────────────────┼───────────────────────────────────┤${RESET}"
echo -e "  ${WHITE}│  What exists now?   │  What's the correct syntax?       │${RESET}"
echo -e "  ${WHITE}│  Apply this change  │  What are the config options?     │${RESET}"
echo -e "  ${WHITE}│  Read this metric   │  What does this error mean?       │${RESET}"
echo -e "  ${BOLD}${WHITE}└─────────────────────┴───────────────────────────────────┘${RESET}"
echo ""

narrate "Together: awareness (what's running) + knowledge (how things work)."

pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 5: Real-world agent workflow
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Part 4: Real-World Agent Workflow" \
  "A coding agent looks up Grafana Alerting using docs + live instance"

narrate "User: 'How do I get started with Grafana Alerting?'"
echo ""
narrate "The agent needs: (1) the current overview (docs) + (2) what is already configured (live)."

pause

section "Step 1: Search docs for the alerting reference" ""

run_cmd "$DOCS" search "alerting rules" --limit 3

pause

section "Step 2: Outline the Grafana Alerting page" ""

run_cmd "$DOCS" outline https://grafana.com/docs/grafana/latest/alerting/

pause

section "Step 3: Fetch the Overview section" ""

run_cmd "$DOCS" get https://grafana.com/docs/grafana/latest/alerting/ --section "Overview"

narrate "The agent now has the current overview. It combines this with the live state"
narrate "from mcp-grafana to answer from docs, not from training cutoff."

pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 6: Wrap up
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Summary" \
  "mcp-doc-server: docs retrieval for AI agents"

echo -e "  ${BOLD}${WHITE}What we built${RESET}"
echo -e "  ${WHITE}• 4 standalone MCP tools: search, get, outline, products${RESET}"
echo -e "  ${WHITE}• Reusable Go core with zero framework dependencies${RESET}"
echo -e "  ${WHITE}• TF-IDF search, bounded retrieval, section extraction${RESET}"
echo -e "  ${WHITE}• Integrated into gcx and mcp-grafana${RESET}"
echo ""
echo -e "  ${BOLD}${WHITE}Key design choices${RESET}"
echo -e "  ${WHITE}• No embeddings, no server-side LLM: deterministic and cheap${RESET}"
echo -e "  ${WHITE}• Bounded by default: agents get slices, not whole pages${RESET}"
echo -e "  ${WHITE}• Allowlisted fetches: only grafana.com/docs/ URLs${RESET}"
echo -e "  ${WHITE}• Clean markdown: frontmatter, shortcodes, HTML stripped${RESET}"
echo ""
echo -e "  ${BOLD}${WHITE}Three surfaces, one core${RESET}"
echo -e "  ${WHITE}• mcp-doc-server (4 MCP tools, stdio)${RESET}"
echo -e "  ${WHITE}• gcx docs (search, get, outline, list-products, list-links)${RESET}"
echo -e "  ${WHITE}• mcp-grafana (search_docs + get_doc alongside other Grafana tools)${RESET}"
echo ""
echo -e "  ${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "  ${BOLD}${WHITE}  github.com/grafana/mcp-doc-server${RESET}"
echo -e "  ${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
