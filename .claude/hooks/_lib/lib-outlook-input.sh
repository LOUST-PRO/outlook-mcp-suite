#!/usr/bin/env bash
# lib-outlook-input.sh — Source-able helper for pre-tool-outlook-*.sh hooks.
#
# Parses the Claude Code PreToolUse JSON envelope piped to stdin and
# extracts outlook-specific fields. Centralizes the jq boilerplate.
#
# Contract (variables set on source):
#   LZT_TOOL_NAME          — string, e.g. "mcp__outlook__apply_send"
#   LZT_OUTLOOK_ACCOUNT    — string, --account arg (may be empty)
#   LZT_OUTLOOK_SUBJECT    — string, --subject arg (may be empty)
#   LZT_OUTLOOK_BODY       — string, --body arg (may be empty)
#   LZT_OUTLOOK_HTML       — string, --html arg (may be empty)
#   LZT_OUTLOOK_TO         — string, --to arg (may be empty)
#   LZT_OUTLOOK_ATTACHMENT — string, --attachment arg (may be empty)
#   LZT_OUTLOOK_FOLDER     — string, --to-folder arg for move (may be empty)
#   LZT_OUTLOOK_CATEGORY   — string, --set arg for categorize (may be empty)
#   LZT_OUTLOOK_ENTRY_ID   — string, first positional arg (may be empty)
#   LZT_RAW                — the full JSON for callers that need more
#
# Contract (helper functions):
#   lzt_outlook_is_apply_tool    — true if LZT_TOOL_NAME matches outlook.apply.*
#   lzt_outlook_is_send_family   — true if apply_send / apply_reply / apply_forward
#   lzt_outlook_is_mutate_family — true if apply_move / apply_categorize / apply_mark_read / send family
#   lzt_outlook_exit_if_not_apply — exits 0 if tool_name doesn't match
#
# Usage:
#   #!/usr/bin/env bash
#   set -euo pipefail
#   . "${HOOK_DIR}/_lib/lib-outlook-input.sh"
#   lzt_outlook_exit_if_not_apply
#   # now use $LZT_OUTLOOK_ACCOUNT etc freely

# ---- source-time initialization ----
# These vars are intentionally exported for callers (hooks) to consume.
# shellcheck disable=SC2034

LZT_RAW="$(cat 2>/dev/null || true)"

_lzt_outlook_extract() {
  local field="$1"
  printf '%s' "${LZT_RAW:-}" | jq -r --arg field "$field" \
    "(.tool_input[\$field] // .input[\$field] // empty)" 2>/dev/null || true
}

LZT_TOOL_NAME="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_name // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_ACCOUNT="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.account // .input.account // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_SUBJECT="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.subject // .input.subject // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_BODY="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.body // .input.body // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_HTML="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.html // .input.html // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_TO="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.to // .input.to // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_ATTACHMENT="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.attachment // .input.attachment // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_FOLDER="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.to_folder // .input.to_folder // empty)' 2>/dev/null || true)"

LZT_OUTLOOK_CATEGORY="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.set // .input.set // empty)' 2>/dev/null || true)"

# First positional arg (entry-id for read/move/categorize/mark-read tools)
LZT_OUTLOOK_ENTRY_ID="$(printf '%s' "${LZT_RAW:-}" | jq -r \
  '(.tool_input.entry_id // .input.entry_id // .tool_input.id // .input.id // empty)' 2>/dev/null || true)"

# ---- helpers ----
lzt_outlook_is_apply_tool() {
  [[ "${LZT_TOOL_NAME:-}" =~ ^mcp__outlook__apply_ ]]
}

lzt_outlook_is_send_family() {
  [[ "${LZT_TOOL_NAME:-}" =~ ^mcp__outlook__apply_(send|reply|forward)$ ]]
}

lzt_outlook_is_mutate_family() {
  [[ "${LZT_TOOL_NAME:-}" =~ ^mcp__outlook__apply_(move|categorize|mark_read|send|reply|forward)$ ]]
}

lzt_outlook_exit_if_not_apply() {
  lzt_outlook_is_apply_tool || exit 0
}