#!/usr/bin/env bats
# require-approval.bats — Phase 0.5 tests for lib-outlook-allowlist.sh
#
# Run with:
#   bats tests/hooks/require-approval.bats
#
# Requires: bats-core, jq, bash 4+

HOOK_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../../.claude/hooks" && pwd)"
LIB="${HOOK_DIR}/_lib/lib-outlook-allowlist.sh"

setup() {
  # Use a sandboxed HOME so we don't pollute the real config
  export HOME="$BATS_TEST_TMPDIR/home"
  mkdir -p "$HOME/.config/lzt-outlook"
  export OUTLOOK_CONFIG="$HOME/.config/lzt-outlook/config.toml"
  # Clear cache
  rm -rf "${TMPDIR:-/tmp}/lzt-outlook-allowlist"
}

teardown() {
  rm -rf "${TMPDIR:-/tmp}/lzt-outlook-allowlist"
}

@test "missing config.toml: default-deny" {
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_account_is_allowed "personal"
  [ "$status" -eq 1 ]
}

@test "empty [accounts].allowed: default-deny" {
  cat > "$OUTLOOK_CONFIG" <<'EOF'
[server]
path = "graph"

[accounts]
allowed = []
default = "personal"
EOF
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_account_is_allowed "personal"
  [ "$status" -eq 1 ]
}

@test "explicit allow: account in list is allowed" {
  cat > "$OUTLOOK_CONFIG" <<'EOF'
[accounts]
allowed = ["personal", "sacrificial"]
EOF
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_account_is_allowed "personal"
  [ "$status" -eq 0 ]
  run outlook_account_is_allowed "sacrificial"
  [ "$status" -eq 0 ]
}

@test "account not in list: blocked" {
  cat > "$OUTLOOK_CONFIG" <<'EOF'
[accounts]
allowed = ["personal"]
EOF
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_account_is_allowed "other_account"
  [ "$status" -eq 1 ]
}

@test "exact-match: prefix-only is NOT a hit" {
  cat > "$OUTLOOK_CONFIG" <<'EOF'
[accounts]
allowed = ["personal"]
EOF
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_account_is_allowed "personally"
  [ "$status" -eq 1 ]
}

@test "exact-match: substring is NOT a hit" {
  cat > "$OUTLOOK_CONFIG" <<'EOF'
[accounts]
allowed = ["personal"]
EOF
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_account_is_allowed "my-personal"
  [ "$status" -eq 1 ]
}

@test "empty account name: blocked" {
  # shellcheck disable=SC1090
  . "$LIB"
  run outlook_account_is_allowed ""
  [ "$status" -eq 1 ]
}