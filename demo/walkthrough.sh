#!/usr/bin/env bash
# ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
# ┃  hack-doc-server — Automated Video Walkthrough                 ┃
# ┃                                                                ┃
# ┃  A self-running demo for screen capture. Hit Enter to advance  ┃
# ┃  each step, or run with --auto for timed delays.               ┃
# ┃                                                                ┃
# ┃  Usage:                                                        ┃
# ┃    ./demo/walkthrough.sh          # interactive (Enter key)    ┃
# ┃    ./demo/walkthrough.sh --auto   # auto-paced (3s delays)    ┃
# ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DOCS="$REPO_ROOT/bin/docs"
GCX="${GCX_BIN:-/tmp/gcx-demo}"
AUTO=false
DELAY=3

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
    echo -e "${YELLOW}  gcx repo not found at $gcx_repo — gcx sections will be skipped${RESET}"
    GCX=""
  fi
fi

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 1: What is hack-doc-server?
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "hack-doc-server" \
  "A docs-retrieval MCP server for Grafana Labs product documentation"

narrate "AI agents need accurate, current documentation — not stale training data."
echo ""
narrate "hack-doc-server gives agents 4 tools to search, browse, and retrieve"
narrate "from 2,000+ official Grafana docs pages. No embeddings. No server-side LLM."
narrate "Just fast, deterministic, token-efficient retrieval."
echo ""
narrate "Let's see it in action."

long_pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 2: The standalone CLI
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Part 1: Standalone CLI" \
  "The docs command — search, get, outline, products"

# -- list_products --
section "list_products" \
  "What documentation exists? 30+ Grafana products, each with hundreds of pages."

run_cmd "$DOCS" products

narrate "An agent can discover the entire product landscape before drilling in."
pause

# -- search_docs --
section "search_docs" \
  "TF-IDF ranked search — rare terms score higher, title matches get 3x weight."

narrate "Scenario: A user asks about rate limiting in Tempo."
echo ""
run_cmd "$DOCS" search "rate limiting" --product tempo --limit 5

narrate "Found the relevant pages. Now let's look inside one."
pause

# -- get_doc_outline --
section "get_doc_outline" \
  "Cheap heading scan — map a page's structure before fetching content."

run_cmd "$DOCS" outline https://grafana.com/docs/tempo/latest/getting-started/

narrate "Now the agent knows exactly which section to fetch."
pause

# -- get_doc (section) --
section "get_doc — section extraction" \
  "Fetch just one section instead of the whole page. 10x fewer tokens."

run_cmd "$DOCS" get https://grafana.com/docs/tempo/latest/getting-started/ \
  --section "Tracing pipeline components"

narrate "Clean markdown. No HTML chrome. No shortcodes. Code blocks preserved."
pause

# -- get_doc (paging) --
section "get_doc — offset/limit paging" \
  "For long sections, page through with offset and limit."

run_cmd "$DOCS" get https://grafana.com/docs/tempo/latest/getting-started/ --offset 0 --limit 20

narrate "The agent sees total_lines and returned_range — it knows there's more."
pause

# -- JSON output --
section "JSON output" \
  "All commands support -o json for programmatic / agent consumption."

run_cmd "$DOCS" search "alerting notification policy" -o json --limit 2

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

  run_cmd "$GCX" docs search "loki storage configuration" --limit 3

  pause

  section "gcx docs outline" \
    "Heading outline with styled table output."

  run_cmd "$GCX" docs outline https://grafana.com/docs/loki/latest/

  pause

  section "gcx docs get — section extraction" \
    "Fetch a specific section, rendered as clean markdown."

  run_cmd "$GCX" docs get https://grafana.com/docs/loki/latest/ --section "Overview" --limit 20

  pause

  section "gcx docs products" \
    "All products, same core data, gcx styling."

  run_cmd "$GCX" docs products

  pause

  narrate "gcx imports only pkg/grafanadocs (the core) — not the MCP or CLI adapter."
  narrate "It writes its own command layer with output.Options, styled tables, and agent mode."
  echo ""

else
  narrate "(gcx binary not available — skipping gcx demo)"
  echo ""
fi

pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 4: mcp-grafana integration
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Part 3: mcp-grafana Integration" \
  "Doc tools registered alongside 30+ existing Grafana MCP tools"

narrate "mcp-grafana is the official Grafana MCP server."
narrate "It now includes 4 documentation tools from hack-doc-server:"
echo ""
echo -e "  ${CYAN}search_docs${RESET}       Search Grafana documentation"
echo -e "  ${CYAN}get_doc${RESET}           Fetch a documentation page"
echo -e "  ${CYAN}get_doc_outline${RESET}   Get heading outline of a page"
echo -e "  ${CYAN}list_products${RESET}     List product documentation groups"
echo ""
narrate "These appear alongside dashboards, alerting, datasources, incidents, etc."

pause

section "Integration architecture" \
  "Three consumers, one core — each writes its own idiomatic wrapper."

echo -e "  ${BOLD}${WHITE}pkg/grafanadocs/${RESET}              ${DIM}← Public core (zero framework deps)${RESET}"
echo ""
echo -e "  ${MAGENTA}Consumer 1: hack-doc-server${RESET}   ${DIM}← MCP adapter on mark3labs/mcp-go${RESET}"
echo -e "    ${DIM}pkg/grafanadocs/mcp/server.go — raw AddTool() registration${RESET}"
echo ""
echo -e "  ${MAGENTA}Consumer 2: gcx${RESET}               ${DIM}← Cobra adapter with styled output${RESET}"
echo -e "    ${DIM}cmd/gcx/docs/ — output.Options, agent annotations, Neon theme${RESET}"
echo ""
echo -e "  ${MAGENTA}Consumer 3: mcp-grafana${RESET}        ${DIM}← MustTool wrappers with jsonschema${RESET}"
echo -e "    ${DIM}tools/docs.go — typed params, consistent with 30+ other tools${RESET}"
echo ""

narrate "All three import pkg/grafanadocs directly. None import each other's adapters."

pause

section "The two-server complement" \
  "mcp-grafana acts on a live instance. hack-doc-server provides the docs."

echo ""
echo -e "  ${BOLD}${WHITE}┌─────────────────────┬───────────────────────────────────┐${RESET}"
echo -e "  ${BOLD}${WHITE}│  mcp-grafana        │  hack-doc-server                  │${RESET}"
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
  "A coding agent configures Grafana alerting using docs + live instance"

narrate "User: 'Set up a notification policy that routes critical alerts to PagerDuty.'"
echo ""
narrate "The agent needs: (1) correct config syntax (docs) + (2) current state (live)."

pause

section "Step 1 — Search docs for the config reference" ""

run_cmd "$DOCS" search "notification policy configuration" --product grafana --limit 3

pause

section "Step 2 — Outline the page to find the right section" ""

run_cmd "$DOCS" outline https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/

pause

section "Step 3 — Fetch the config section" ""

run_cmd "$DOCS" get https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/ \
  --section "Example"

narrate "The agent now has the exact syntax. It combines this with the live state"
narrate "from mcp-grafana to generate correct, non-duplicating configuration."

pause

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# PART 6: Wrap up
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

title_card "Summary" \
  "hack-doc-server — docs retrieval for AI agents"

echo -e "  ${BOLD}${WHITE}What we built${RESET}"
echo -e "  ${WHITE}• 4 MCP tools: search, get, outline, products${RESET}"
echo -e "  ${WHITE}• Reusable Go core with zero framework dependencies${RESET}"
echo -e "  ${WHITE}• TF-IDF search, bounded retrieval, section extraction${RESET}"
echo -e "  ${WHITE}• Integrated into gcx and mcp-grafana${RESET}"
echo ""
echo -e "  ${BOLD}${WHITE}Key design choices${RESET}"
echo -e "  ${WHITE}• No embeddings, no server-side LLM — deterministic and cheap${RESET}"
echo -e "  ${WHITE}• Bounded by default — agents get slices, not whole pages${RESET}"
echo -e "  ${WHITE}• Allowlisted fetches — only grafana.com/docs/ URLs${RESET}"
echo -e "  ${WHITE}• Clean markdown — frontmatter, shortcodes, HTML stripped${RESET}"
echo ""
echo -e "  ${BOLD}${WHITE}Three surfaces, one core${RESET}"
echo -e "  ${WHITE}• hack-doc-server (standalone MCP server, stdio)${RESET}"
echo -e "  ${WHITE}• gcx docs (CLI subcommands with styled output)${RESET}"
echo -e "  ${WHITE}• mcp-grafana (doc tools alongside 30+ Grafana tools)${RESET}"
echo ""
echo -e "  ${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "  ${BOLD}${WHITE}  github.com/grafana/hack-doc-server${RESET}"
echo -e "  ${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
