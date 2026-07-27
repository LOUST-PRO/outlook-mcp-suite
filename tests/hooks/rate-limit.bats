#!/usr/bin/env bats
# rate-limit.bats — Phase 0.5 tests for lib-outlook-rate.sh
#
# Run with:
#   bats tests/hooks/rate-limit.bats

HOOK_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../../.claude/hooks" && pwd)"
LIB="${HOOK_DIR}/_lib/lib-outlook-rate.sh"

setup() {
  export HOME="$BATS_TEST_TMPDIR/home"
  mkdir -p "$HOME/.local/share/lzt-outlook/rate-counter"
  export OUTLOOK_RATE_DIR="$HOME/.local/share/lzt-outlook/rate-counter"
  # Clear any existing state
  rm -f "$OUTLOOK_RATE_DIR"/*.json 2>/dev/null || true
}

@test "first call: allowed" {
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_rate_check "personal" 10
  [ "$status" -eq 0 ]
}

@test "within limit: 10/10 calls allowed" {
  # shellcheck disable=SC1090
  . "$LIB"
  for i in $(seq 1 10); do
    run outlook_rate_check "personal" 10
    [ "$status" -eq 0 ]
  done
}

@test "11th call: blocked" {
  # shellcheck disable=SC1090
  . "$LIB"
  for i in $(seq 1 10); do
    outlook_rate_check "personal" 10
  done
  run outlook_rate_check "personal" 10
  [ "$status" -eq 1 ]
}

@test "separate accounts: independent limits" {
  # shellcheck disable=SC1090
  . "$LIB"
  # Fill personal's window
  for i in $(seq 1 10); do
    outlook_rate_check "personal" 10
  done
  # sacrificial still has full budget
  run outlook_rate_check "sacrificial" 10
  [ "$status" -eq 0 ]
}

@test "outlook_rate_count: reflects current state" {
  # shellcheck disable=SC1090
  . "$LIB"
  outlook_rate_check "personal" 10 >/dev/null
  outlook_rate_check "personal" 10 >/dev/null
  outlook_rate_check "personal" 10 >/dev/null
  count="$(outlook_rate_count "personal")"
  [ "$count" -eq 3 ]
}

@test "outlook_rate_count: returns 0 for unknown account" {
  # shellcheck disable=SC1090
  . "$LIB"
  count="$(outlook_rate_count "nonexistent")"
  [ "$count" -eq 0 ]
}

@test "custom max: 3/3 calls allowed, 4th blocked" {
  # shellcheck disable=SC1090
  . "$LIB"
  outlook_rate_check "personal" 3 >/dev/null
  outlook_rate_check "personal" 3 >/dev/null
  run outlook_rate_check "personal" 3
  [ "$status" -eq 0 ]
  run outlook_rate_check "personal" 3
  [ "$status" -eq 1 ]
}