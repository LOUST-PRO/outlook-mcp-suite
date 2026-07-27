#!/usr/bin/env bash
# pre-tool-outlook-rate-limit.sh — Phase 0.5 hook #2.
#
# Gates every mcp__outlook__apply_* tool call. Blocks when the per-account
# apply count exceeds the rolling 5-minute window limit (default: 10).
#
# Why this exists:
#   Without a rate limit, a runaway agent loop could send 1000s of emails
#   in seconds. Even with shadow mode, propose_* tools could overwhelm
#   Graph throttling. This hook bounds apply rate at the agent-loop level.
#
# Exit codes:
#   0  allow (or not-an-apply-tool, no-op)
#   2  block

set -uo pipefail

HOOK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "${HOOK_DIR}/_lib/lib-outlook-input.sh"
. "${HOOK_DIR}/_lib/lib-outlook-rate.sh"
. "${HOOK_DIR}/_lib/lib-outlook-decision.sh"

lzt_outlook_exit_if_not_apply

# Default rate limit (overridable via env)
MAX_PER_5MIN="${OUTLOOK_RATE_LIMIT_PER_5MIN:-10}"

# Need account for rate limiting
if [[ -z "${LZT_OUTLOOK_ACCOUNT:-}" ]]; then
  # Let the require-approval hook handle the missing-account case
  lzt_outlook_decision_allow
fi

if ! outlook_rate_check "$LZT_OUTLOOK_ACCOUNT" "$MAX_PER_5MIN"; then
  local_count="$(outlook_rate_count "$LZT_OUTLOOK_ACCOUNT" 2>/dev/null || echo "11+")"
  lzt_outlook_decision_block \
    "rate limit exceeded for account '$LZT_OUTLOOK_ACCOUNT'" \
    "Apply count in last 5 min: $local_count (limit: $MAX_PER_5MIN). Wait 5 min, or set OUTLOOK_RATE_LIMIT_PER_5MIN env var in the MCP server config to raise the cap."
fi

lzt_outlook_decision_allow