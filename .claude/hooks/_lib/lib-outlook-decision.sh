#!/usr/bin/env bash
# lib-outlook-decision.sh — Source-able helper.
#
# Thin wrapper over lzt-harness's lib-lzt-decision-emit.sh pattern, but
# self-contained so this repo's hooks don't depend on the lzt-harness
# plugin being installed in the consumer's environment.
#
# Contract:
#   lzt_outlook_decision_block <reason> [<additional_context>]
#     → emits {decision: block, reason, hookSpecificOutput: {additionalContext}}
#       to stderr in Claude Code 2.1+ format; exits 2
#
#   lzt_outlook_decision_allow
#     → exits 0 silently
#
# Usage:
#   . "${HOOK_DIR}/_lib/lib-outlook-decision.sh"
#   if bad_pattern; then
#     lzt_outlook_decision_block "..." "..."
#   fi
#   lzt_outlook_decision_allow

_lzt_outlook_decision_emit() {
  local decision="$1"
  local reason="$2"
  local context="${3:-}"
  local hook_event="${4:-PreToolUse}"

  jq -nc \
    --arg decision "${decision}" \
    --arg reason "${reason}" \
    --arg context "${context}" \
    --arg hook_event "${hook_event}" \
    '{
      decision: $decision,
      reason: $reason,
      hookSpecificOutput: {
        hookEventName: $hook_event,
        additionalContext: $context
      }
    }' >&2
}

lzt_outlook_decision_block() {
  local reason="${1:-blocked by outlook-mcp-suite hook}"
  local context="${2:-}"
  _lzt_outlook_decision_emit "block" "${reason}" "${context}" "PreToolUse" || true
  exit 2
}

lzt_outlook_decision_allow() {
  exit 0
}

# Helper for the allowlist hook: list allowed accounts for diagnostics
outlook_account_list() {
  local cache_file
  cache_file="${TMPDIR:-/tmp}/lzt-outlook-allowlist/$(printf '%s' "${OUTLOOK_CONFIG:-$HOME/.config/lzt-outlook/config.toml}" | tr '/ ' '__')"
  [[ -s "$cache_file" ]] && tr '\n' ' ' < "$cache_file" | sed 's/ $//'
  return 0
}