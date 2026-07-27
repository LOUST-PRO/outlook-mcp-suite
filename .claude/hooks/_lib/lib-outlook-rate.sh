#!/usr/bin/env bash
# lib-outlook-rate.sh — Source-able helper.
#
# Per-account rate counter for outlook.apply.* tool calls. Uses a
# sliding 5-minute window persisted as JSON state in
# ~/.local/share/lzt-outlook/rate-counter/<account>.json.
#
# Contract:
#   outlook_rate_check <account> <max_per_5min>
#     → exits 0 if count < max, exits 1 if count >= max
#     → side effect: appends current timestamp to the window
#
#   outlook_rate_count <account>
#     → echoes the current count (does not modify state)
#
# State file format:
#   {"account": "<name>", "window_start_ts": <epoch>, "events": [<ts>, ...]}
#
# The sliding window is approximated by: if oldest event is older than
# 5 minutes, drop it. For low-rate workloads (<100 events/day) this is
# exact. For high-rate, O(n) per check is fine since n ≤ max.
#
# Usage:
#   . "${HOOK_DIR}/_lib/lib-outlook-rate.sh"
#   if ! outlook_rate_check "$LZT_OUTLOOK_ACCOUNT" 10; then
#     lzt_decision_emit_block "..." "..."
#   fi

RATE_DIR="${OUTLOOK_RATE_DIR:-$HOME/.local/share/lzt-outlook/rate-counter}"
RATE_WINDOW=300  # 5 minutes in seconds

_rate_state_path() {
  printf '%s/%s.json' "$RATE_DIR" "$(printf '%s' "$1" | tr '/ ' '__')"
}

_rate_now() {
  date +%s
}

_rate_load() {
  local path
  path="$(_rate_state_path "$1")"
  if [[ -f "$path" ]]; then
    cat "$path" 2>/dev/null
  else
    printf '{"account":"%s","window_start_ts":0,"events":[]}\n' "$1"
  fi
}

_rate_save() {
  local account="$1" json="$2"
  mkdir -p "$RATE_DIR" 2>/dev/null || return 1
  local path
  path="$(_rate_state_path "$account")"
  printf '%s\n' "$json" > "$path"
  chmod 0600 "$path" 2>/dev/null || true
}

outlook_rate_count() {
  local account="$1"
  local now
  now="$(_rate_now)"
  local state
  state="$(_rate_load "$account")"

  # Count events within the window
  printf '%s' "$state" | jq -r --argjson now "$now" --argjson win "$RATE_WINDOW" '
    [.events[] | select(. > ($now - $win))] | length
  ' 2>/dev/null || echo 0
}

outlook_rate_check() {
  local account="$1"
  local max="${2:-10}"
  local now
  now="$(_rate_now)"

  local state
  state="$(_rate_load "$account")"

  # Prune events older than window + append current + count
  local new_state
  new_state="$(printf '%s' "$state" | jq -c --arg acct "$account" --argjson now "$now" --argjson win "$RATE_WINDOW" '
    .account = $acct |
    .events = ([.events[] | select(. > ($now - $win))] + [$now]) |
    .window_start_ts = (if (.events | length) > 0 then .events[0] else $now end)
  ' 2>/dev/null)" || new_state=""

  [[ -n "$new_state" ]] || return 1

  _rate_save "$account" "$new_state"

  # Check count
  local count
  count="$(printf '%s' "$new_state" | jq -r '.events | length' 2>/dev/null || echo 0)"
  (( count <= max ))
}