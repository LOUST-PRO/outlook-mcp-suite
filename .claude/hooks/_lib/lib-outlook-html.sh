#!/usr/bin/env bash
# lib-outlook-html.sh — Source-able helper.
#
# Scans HTML for blocked tags/patterns that should not be sent via email:
#   - <script> tags (any content)
#   - <iframe> with javascript: or data:text/html src
#   - 1×1 tracking pixels (<img width="1" height="1" ...>)
#   - <object> / <embed> with javascript: src or type
#   - on* event handlers on any tag
#
# Contract:
#   outlook_scan_html <html>
#     → exits 0 if clean, exits 1 if any pattern matched
#     → on match, echoes the matched pattern name to stdout
#
# Patterns use POSIX ERE via grep -E. They are deliberately conservative:
# a false positive (legitimate mail blocked) is better than a false
# negative (tracking pixel sent).
#
# Usage:
#   . "${HOOK_DIR}/_lib/lib-outlook-html.sh"
#   if ! outlook_scan_html "$LZT_OUTLOOK_HTML"; then
#     lzt_decision_emit_block "..." "..."
#   fi

outlook_scan_html() {
  local html="${1:-}"
  [[ -n "$html" ]] && [[ "${#html}" -lt 1000000 ]] || return 0  # cap 1MB

  # Strip CDATA + comments to avoid pattern collisions in JS samples
  local stripped
  stripped="$(printf '%s' "$html" \
    | sed -e ':a;N;$!ba;s/<!--.*-->//g' \
          -e 's/<!\[CDATA\[.*\]\]>//g' 2>/dev/null)"

  # <script> tags
  if printf '%s' "$stripped" | grep -qiE '<script[[:space:]]'; then
    echo "script-tag"
    return 1
  fi

  # <iframe> with javascript: or data:text/html src
  if printf '%s' "$stripped" | grep -qiE '<iframe[^>]+src[[:space:]]*=[[:space:]]*["'"'"']?(javascript|data:text/html)'; then
    echo "iframe-js-src"
    return 1
  fi

  # 1×1 tracking pixel: <img ...> with width="1" or height="1"
  if printf '%s' "$stripped" | grep -qiE '<img[^>]*(width[[:space:]]*=[[:space:]]*["'"'"']?1["'"'"']?|height[[:space:]]*=[[:space:]]*["'"'"']?1["'"'"']?)'; then
    echo "tracking-pixel"
    return 1
  fi

  # <object> / <embed> with JS or script src/type
  if printf '%s' "$stripped" | grep -qiE '<(object|embed)[[:space:]][^>]*((src|type)[[:space:]]*=[[:space:]]*["'"'"']?(javascript|application/x-javascript|data:text/html)|script)'; then
    echo "object-embed-js"
    return 1
  fi

  # on* event handlers (onclick, onerror, onload, etc.)
  if printf '%s' "$stripped" | grep -qiE 'on(load|click|error|mouseover|focus|blur|submit)[[:space:]]*='; then
    echo "on-event-handler"
    return 1
  fi

  return 0
}