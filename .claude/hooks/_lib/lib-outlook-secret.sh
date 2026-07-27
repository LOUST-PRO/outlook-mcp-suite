#!/usr/bin/env bash
# lib-outlook-secret.sh — Source-able helper.
#
# Scans text for PII patterns that should never be sent via email:
# SSN, credit card numbers, AWS keys, GitHub PATs, Bitwarden URIs,
# private IPv4 addresses, generic high-entropy API key shapes.
#
# Contract:
#   outlook_scan_secrets <text>
#     → exits 0 if clean, exits 1 if any pattern matched
#     → on match, echoes the matched pattern name to stdout
#
# Patterns are deliberately conservative (high precision, lower recall).
# False negatives are worse than false positives here — better to block
# a legitimate email than to leak an SSN.
#
# Usage:
#   . "${HOOK_DIR}/_lib/lib-outlook-secret.sh"
#   if ! outlook_scan_secrets "$LZT_OUTLOOK_BODY"; then
#     lzt_decision_emit_block "..." "..."
#   fi

# Patterns are evaluated with grep -E (POSIX ERE). Each pattern has a
# short label that becomes the diagnostic if matched.

outlook_scan_secrets() {
  local text="${1:-}"
  [[ -n "$text" ]] && [[ "${#text}" -lt 100000 ]] || return 0

  # SSN: NNN-NN-NNNN (with word boundaries, not part of longer digit run)
  if printf '%s' "$text" | grep -qE '(^|[^0-9])[0-9]{3}-[0-9]{2}-[0-9]{4}([^0-9]|$)'; then
    echo "ssn"
    return 1
  fi

  # Credit card: 13-19 digits with optional spaces/dashes (Luhn-not-checked;
  # conservative pattern catches obvious numbers, may have false positives
  # like phone numbers. Users can whitelist in config.toml if needed.)
  if printf '%s' "$text" | grep -qE '(^|[^0-9])[0-9]{4}[ -]?[0-9]{4}[ -]?[0-9]{4}[ -]?[0-9]{4}([^0-9]|$)'; then
    echo "credit-card-shape"
    return 1
  fi

  # AWS access key: AKIA / ASIA prefix + 16 chars
  if printf '%s' "$text" | grep -qE '(AKIA|ASIA)[0-9A-Z]{16}'; then
    echo "aws-access-key"
    return 1
  fi

  # GitHub PAT: ghp_ / gho_ / ghu_ / ghs_ / ghr_ + 36 chars
  if printf '%s' "$text" | grep -qE 'gh[pousr]_[A-Za-z0-9]{36}'; then
    echo "github-pat"
    return 1
  fi

  # Bitwarden URI scheme: bw:// or https://vault.bitwarden.com/...
  if printf '%s' "$text" | grep -qE 'bw://|https?://vault\.bitwarden\.com/'; then
    echo "bitwarden-uri"
    return 1
  fi

  # Private IPv4: 10.x, 192.168.x, 172.16-31.x (informational; not strictly
  # a secret, but accidentally emailing internal IPs is a leak)
  if printf '%s' "$text" | grep -qE '(^|[^0-9.])(10\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}|192\.168\.[0-9]{1,3}\.[0-9]{1,3}|172\.(1[6-9]|2[0-9]|3[01])\.[0-9]{1,3}\.[0-9]{1,3})([^0-9.]|$)'; then
    echo "private-ipv4"
    return 1
  fi

  # Generic API key shape: "api_key=..." or "apikey:..." (high-entropy)
  # Conservative: requires the literal "api_key" or "apikey" prefix.
  if printf '%s' "$text" | grep -qiE '(api[_-]?key|secret|token)[[:space:]]*[:=][[:space:]]*[A-Za-z0-9_/+=-]{20,}'; then
    echo "api-key-shape"
    return 1
  fi

  return 0
}