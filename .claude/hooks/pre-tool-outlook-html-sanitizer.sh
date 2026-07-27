#!/usr/bin/env bash
# pre-tool-outlook-html-sanitizer.sh — Phase 0.5 hook #4.
#
# Gates apply_send / apply_reply / apply_forward when --html is provided.
# Scans the HTML for blocked tags: <script>, <iframe src=javascript:>,
# tracking pixels (1×1 <img>), <object>/<embed> pointing to JS,
# on* event handlers.
#
# Why this exists:
#   HTML email is rendered by recipients' mail clients, which may
#   execute scripts, fetch tracking pixels (privacy leak), and load
#   external iframes (clickjacking / fingerprinting). Even legitimate
#   corporate newsletters include tracking pixels. The agent should
#   send plain text by default; if HTML is used, it must be clean.
#
# Exit codes:
#   0  allow (or not a send-family tool, or no --html)
#   2  block

set -uo pipefail

HOOK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "${HOOK_DIR}/_lib/lib-outlook-input.sh"
. "${HOOK_DIR}/_lib/lib-outlook-html.sh"
. "${HOOK_DIR}/_lib/lib-outlook-decision.sh"

# Only gate send family
lzt_outlook_is_send_family || lzt_outlook_decision_allow

# No --html → nothing to sanitize
[[ -n "${LZT_OUTLOOK_HTML:-}" ]] || lzt_outlook_decision_allow

if match="$(outlook_scan_html "$LZT_OUTLOOK_HTML")"; then
  lzt_outlook_decision_allow
else
  pattern="${match:-unknown}"
  lzt_outlook_decision_block \
    "blocked HTML pattern: $pattern" \
    "Outlook-hook html-sanitizer matched pattern '$pattern'. Drop --html (use --body plain text), or rewrite the HTML to remove the blocked tag. Patterns are in .claude/hooks/_lib/lib-outlook-html.sh."
fi