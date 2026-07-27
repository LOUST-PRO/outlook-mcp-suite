#!/usr/bin/env bash
# pre-tool-outlook-secret-scan.sh — Phase 0.5 hook #3.
#
# Gates apply_send / apply_reply / apply_forward tool calls. Scans
# subject + body + attachment path for PII patterns (SSN, credit card,
# AWS key, GitHub PAT, Bitwarden URI, private IPv4, generic api-key shape).
# Blocks if any pattern matches.
#
# Why this exists:
#   PII in emails is the #1 cause of accidental data leaks. The agent
#   might compose a reply that quotes a user message containing an SSN,
#   credit card, or API key. The hook catches this before the apply.
#
# Exit codes:
#   0  allow (or not-a-send-family tool, no-op)
#   2  block

set -uo pipefail

HOOK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "${HOOK_DIR}/_lib/lib-outlook-input.sh"
. "${HOOK_DIR}/_lib/lib-outlook-secret.sh"
. "${HOOK_DIR}/_lib/lib-outlook-decision.sh"

# Only gate send family
lzt_outlook_is_send_family || lzt_outlook_decision_allow

# Concatenate fields to scan
scan_text="${LZT_OUTLOOK_SUBJECT:-} ${LZT_OUTLOOK_BODY:-} ${LZT_OUTLOOK_HTML:-} ${LZT_OUTLOOK_TO:-} ${LZT_OUTLOOK_ATTACHMENT:-}"

if match="$(outlook_scan_secrets "$scan_text")"; then
  lzt_outlook_decision_allow
else
  pattern="${match:-unknown}"
  lzt_outlook_decision_block \
    "PII pattern detected: $pattern" \
    "Outlook-hook secret-scan matched pattern '$pattern' in subject/body/html/to/attachment. Review the message and redact the pattern manually before re-trying. Allowed patterns are in .claude/hooks/_lib/lib-outlook-secret.sh."
fi