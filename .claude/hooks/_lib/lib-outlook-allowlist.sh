#!/usr/bin/env bash
# lib-outlook-allowlist.sh — Source-able helper.
#
# Reads ~/.config/lzt-outlook/config.toml [accounts].allowed and exposes
# a function to check whether a given account name is in the allowlist.
# Cached for 60s to avoid re-parsing on every tool call.
#
# Contract:
#   outlook_account_is_allowed <account_name>
#     → exits 0 if allowed, exits 1 if not (or not configured)
#
# Default-deny: if config.toml doesn't exist or [accounts].allowed is
# missing/empty, ALL accounts are blocked. This is intentional — it's
# safer to block than to silently allow.
#
# Usage:
#   . "${HOOK_DIR}/_lib/lib-outlook-allowlist.sh"
#   if outlook_account_is_allowed "$LZT_OUTLOOK_ACCOUNT"; then
#     echo "allowed"
#   else
#     lzt_decision_emit_block "..." "..."
#   fi

CONFIG_PATH="${OUTLOOK_CONFIG:-$HOME/.config/lzt-outlook/config.toml}"
CACHE_DIR="${TMPDIR:-/tmp}/lzt-outlook-allowlist"
CACHE_TTL=60  # seconds

_outlook_allowlist_cache_path() {
  printf '%s/%s' "$CACHE_DIR" "$(printf '%s' "$CONFIG_PATH" | tr '/ ' '__')"
}

_outlook_allowlist_cache_fresh() {
  local cache_file
  cache_file="$(_outlook_allowlist_cache_path)"
  [[ -f "$cache_file" ]] || return 1
  local cache_mtime now
  cache_mtime="$(stat -c %Y "$cache_file" 2>/dev/null || stat -f %m "$cache_file" 2>/dev/null || echo 0)"
  now="$(date +%s)"
  (( now - cache_mtime < CACHE_TTL )) || return 1
  return 0
}

_outlook_allowlist_load() {
  mkdir -p "$CACHE_DIR" 2>/dev/null || return 0
  local cache_file
  cache_file="$(_outlook_allowlist_cache_path)"

  # If config doesn't exist → cache empty allowlist (default-deny)
  if [[ ! -f "$CONFIG_PATH" ]]; then
    : > "$cache_file"
    return 0
  fi

  # Extract [accounts] section, then read allowed = [...] line(s).
  # Supports single-line arrays only ("a", "b", "c"). Multi-line arrays
  # are rejected with empty cache.
  awk '
    BEGIN { in_accounts = 0 }
    /^\[accounts\]/ { in_accounts = 1; next }
    /^\[/ { in_accounts = 0 }
    in_accounts && /allowed[[:space:]]*=/ {
      # Extract content between first [ and last ]
      line = $0
      sub(/.*\[/, "", line)
      sub(/\].*/, "", line)
      gsub(/[" ]/, "", line)
      n = split(line, parts, ",")
      for (i = 1; i <= n; i++) {
        if (parts[i] != "") print parts[i]
      }
      exit
    }
  ' "$CONFIG_PATH" > "$cache_file" 2>/dev/null || : > "$cache_file"
}

outlook_account_is_allowed() {
  local account="$1"
  [[ -n "$account" ]] || return 1

  _outlook_allowlist_cache_fresh || _outlook_allowlist_load

  local cache_file
  cache_file="$(_outlook_allowlist_cache_path)"
  [[ -s "$cache_file" ]] || return 1  # default-deny if empty

  # grep -Fxq: whole-line exact match, silent
  grep -Fxq -- "$account" "$cache_file" 2>/dev/null
}