#!/usr/bin/env bash
# pre-tool-outlook-apply-require-approval.sh — Phase 0.5 hook #1.
#
# Gates every mcp__outlook__apply_* tool call. Blocks when:
#   1. --account is not in config.toml [accounts].allowed (default-deny)
#   2. tool_name does not match outlook.apply.* (no-op)
#
# Why this exists:
#   The agent could call outlook.apply.send with arbitrary account names.
#   Without a hook, the wrapper has no second line of defense beyond its
#   own validation. This hook adds an operator-controlled allowlist.
#
# Exit codes (Claude Code PreToolUse convention):
#   0  allow (or not-an-apply-tool, no-op)
#   2  block (stderr JSON envelope with reason + additionalContext)

set -uo pipefail

# Resolve hook dir relative to this script
HOOK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "${HOOK_DIR}/_lib/lib-outlook-input.sh"
. "${HOOK_DIR}/_lib/lib-outlook-allowlist.sh"
. "${HOOK_DIR}/_lib/lib-outlook-decision.sh"

# No-op if not an apply tool
lzt_outlook_exit_if_not_apply

# Default-deny: if account is empty, block with explicit reason
if [[ -z "${LZT_OUTLOOK_ACCOUNT:-}" ]]; then
  lzt_outlook_decision_block \
    "apply tool called without --account" \
    "Every outlook.apply.* call MUST include --account <name>. The name must be in ~/.config/lzt-outlook/config.toml [accounts].allowed."
fi

# Check allowlist (default-deny if config missing or empty)
if ! outlook_account_is_allowed "$LZT_OUTLOOK_ACCOUNT"; then
  lzt_outlook_decision_block \
    "account '$LZT_OUTLOOK_ACCOUNT' is not in the allowlist" \
    "Add '$LZT_OUTLOOK_ACCOUNT' to ~/.config/lzt-outlook/config.toml under [accounts].allowed, then restart the MCP server. Allowed accounts: $(outlook_account_list 2>/dev/null || echo '(config missing — default-deny active)')"
fi

lzt_outlook_decision_allow