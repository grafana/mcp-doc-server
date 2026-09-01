#!/usr/bin/env bash
set -euf -o pipefail

function usage {
	cat <<EOF
Run an interactive or auto-paced walkthrough of mcp-doc-server.

Usage:
  ./docs/design/demo/walkthrough.bash --interactive
  ./docs/design/demo/walkthrough.bash --auto
  ./docs/design/demo/walkthrough.bash --help

Examples:
  ./docs/design/demo/walkthrough.bash --interactive
  ./docs/design/demo/walkthrough.bash --auto
EOF
}

if [[ $# -ne 1 ]]; then
	usage
	exit 2
fi

AUTO=false
DELAY=6

case "${1}" in
-h | --help)
	usage
	exit 0
	;;
--interactive)
	AUTO=false
	;;
--auto)
	AUTO=true
	;;
*)
	echo "Unknown option: ${1}" >&2
	usage
	exit 2
	;;
esac

REPO_ROOT="$(cd "$(dirname "${0}")/../../.." && pwd)"
DOCS="${REPO_ROOT}/bin/docs"
GCX="${GCX_BIN:-/tmp/gcx-demo}"

BOLD='\033[1m'
DIM='\033[2m'
CYAN='\033[36m'
GREEN='\033[32m'
YELLOW='\033[33m'
WHITE='\033[97m'
RESET='\033[0m'

pause() {
	if [[ "${AUTO}" == true ]]; then
		sleep "${DELAY}"
	else
		echo -e "${DIM}  ↵ press Enter to continue${RESET}"
		read -r
	fi
}

long_pause() {
	if [[ "${AUTO}" == true ]]; then
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
	echo -e "${BOLD}${CYAN}  ║${RESET}${BOLD}${WHITE}  ${1}$(printf '%*s' $((56 - ${#1})) '')${RESET}${BOLD}${CYAN}    ║${RESET}"
	echo -e "${BOLD}${CYAN}  ╚══════════════════════════════════════════════════════════════╝${RESET}"
	echo ""
	echo -e "  ${DIM}${2}${RESET}"
	echo ""
}

section() {
	echo ""
	echo -e "${BOLD}${YELLOW}  ▸ ${1}${RESET}"
	echo -e "  ${DIM}${2}${RESET}"
	echo ""
}

narrate() {
	echo -e "  ${WHITE}${1}${RESET}"
}

run_cmd() {
	local rendered=""
	printf -v rendered ' %q' "$@"
	echo -e "  ${GREEN}\$${rendered}${RESET}"
	echo ""
	"$@" 2>/dev/null | sed 's/^/  /' || true
	echo ""
}

if [[ ! -x "${DOCS}" ]]; then
	echo -e "${DIM}  Building bin/docs...${RESET}"
	(cd "${REPO_ROOT}" && go build -o bin/docs ./cmd/docs/)
fi

gcx_repo="${REPO_ROOT}/../gcx"
gcx_has_docs=false
if [[ -x "${GCX}" ]] && "${GCX}" docs --help >/dev/null 2>&1; then
	gcx_has_docs=true
fi
if [[ ! -x "${GCX}" || "${gcx_has_docs}" != true ]]; then
	echo -e "${DIM}  Building gcx with docs commands (this takes a moment)...${RESET}"
	if [[ -d "${gcx_repo}" ]]; then
		(cd "${gcx_repo}" && go build -buildvcs=false -o "${GCX}" ./cmd/gcx/) || true
	else
		echo -e "${YELLOW}  gcx repo not found at ${gcx_repo} — gcx sections will be skipped${RESET}"
		GCX=""
	fi
fi

title_card "mcp-doc-server" \
	"Docs retrieval for AI agents which is current, cheap, and precise"

narrate "Why not just prompt with docs? Three reasons to prefer mcp-doc-server:"
echo ""
narrate "  1. It's always current. It fetches live docs instead of using stale training data."
narrate "  2. It's token-efficient. One section costs only hundreds of tokens and not thousands."
narrate "  3. It's precise and matches human behavior from search, to outline, and then extract the information."
echo ""
narrate "4 tools, 2,000+ Grafana docs pages. Let's see it."

long_pause

title_card "Part 1: Standalone CLI" \
	"Available tool calls"

section "list_products" \
	"What product documentation exists?"

run_cmd "${DOCS}" products

narrate "An agent can discover the entire product landscape before digging in."
pause

section "search_docs" \
	"Ranked results relating to the search term."

narrate "Scenario: A user asks how to construct a TraceQL query."
echo ""
run_cmd "${DOCS}" search "traceql query" --product tempo --limit 5

narrate "Found the relevant pages. Now let's look inside one."
pause

section "get_doc_outline" \
	"List a page's headings to understand the structure before fetching the content."

run_cmd "${DOCS}" outline https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/

narrate "Now the agent knows exactly which section to fetch."
pause

section "get_doc: retrieve a section" \
	"Fetch just one section instead of the whole page for far fewer tokens."

run_cmd "${DOCS}" get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/ \
	--section "Comparison operators"

narrate "Returns our Markdown representation of the documentation."
pause

section "get_doc: offsets and limits" \
	"For long pages, retrieve results in slices with offset and limit."

run_cmd "${DOCS}" get https://grafana.com/docs/tempo/latest/configuration/ --offset 80 --limit 80

narrate "The agent sees total_lines and returned_range and knows if there's more."
pause

section "JSON output" \
	"All commands support -o json for programmatic consumption."

run_cmd "${DOCS}" search "alerting rules" -o json --limit 2

pause

title_card "Part 2: gcx Integration" \
	"gcx imports the core and mounts it as 'gcx docs'"

if [[ -n "${GCX}" && -x "${GCX}" ]]; then
	narrate "gcx is a unified CLI for managing Grafana resources."
	narrate "It now has docs commands alongside dashboards, alerting, datasources, and the rest."
	echo ""

	section "gcx docs search" \
		"Same search but with gcx output styling"

	run_cmd "${GCX}" docs search "traceql query" --product tempo --limit 3

	pause

	section "gcx docs outline" \
		"Heading outline with styled table output."

	run_cmd "${GCX}" docs outline https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/

	pause

	section "gcx docs get: retrieve a section" \
		"Fetch a specific section and return the Markdown."

	run_cmd "${GCX}" docs get https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/ --section "Comparison operators" --limit 20

	pause

	section "gcx docs list-products" \
		"All product groups with gcx styling"

	run_cmd "${GCX}" docs list-products

	pause

	section "gcx docs list-links" \
		"Curated canonical URLs from the local registry"

	run_cmd "${GCX}" docs list-links

	echo ""
else
	narrate "(gcx binary not available — skipping gcx demo)"
	echo ""
fi

pause

title_card "Part 3: mcp-grafana Integration" \
	"Doc tools registered along the other existing Grafana MCP tools"

narrate "mcp-grafana is the official Grafana MCP server."
narrate "It folds the four standalone docs tools into two wrappers:"
echo ""
echo -e "  ${CYAN}search_docs${RESET}       Search docs, or list products when query is omitted"
echo -e "  ${CYAN}get_doc${RESET}           Fetch a page, or return an outline with outline_only=true"
echo ""
narrate "These appear alongside dashboards, alerting, datasources, incidents, and more."
narrate "The docs category is enabled by default. --disable-docs turns it off."

pause

section "It's a complementary interface" \
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

narrate "Together: awareness of what's running and knowledge of how things work."

pause

title_card "Part 4: Real-World Agent Workflow" \
	"A coding agent looks up Grafana Alerting using docs and a live instance"

narrate "User: 'How do I get started with Grafana Alerting?'"
echo ""
narrate "The agent needs:"
narrate "  1. current guidance from docs"
narrate "  2. current state from the live instance."

pause

section "Step 1: Search docs for the alerting reference" ""

run_cmd "${DOCS}" search "alerting rules" --limit 3

pause

section "Step 2: Outline the page to find the right section" ""

run_cmd "${DOCS}" outline https://grafana.com/docs/grafana/latest/alerting/

pause

section "Step 3: Fetch the Overview section" ""

run_cmd "${DOCS}" get https://grafana.com/docs/grafana/latest/alerting/ --section "Overview"

narrate "The agent now has the current overview and combines this with the live state"
narrate "from mcp-grafana to answer from docs and not from training cutoff."

pause

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
