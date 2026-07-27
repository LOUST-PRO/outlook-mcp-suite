#!/usr/bin/env bats
# secret-scan.bats — Phase 0.5 tests for lib-outlook-secret.sh
#
# Run with:
#   bats tests/hooks/secret-scan.bats

HOOK_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../../.claude/hooks" && pwd)"
LIB="${HOOK_DIR}/_lib/lib-outlook-secret.sh"

setup() {
  # shellcheck disable=SC1090
  . "$LIB"
}

@test "empty input: clean (returns 0)" {
  run outlook_scan_secrets ""
  [ "$status" -eq 0 ]
}

@test "plain text: clean" {
  run outlook_scan_secrets "Hello, please find attached the report. Best, Lou"
  [ "$status" -eq 0 ]
}

@test "SSN detected" {
  run outlook_scan_secrets "My SSN is 123-45-6789 for verification."
  [ "$status" -eq 1 ]
  [ "$output" = "ssn" ]
}

@test "SSN at start of line detected" {
  run outlook_scan_secrets "987-65-4321 is my SSN"
  [ "$status" -eq 1 ]
  [ "$output" = "ssn" ]
}

@test "credit-card-shape detected (16 digits)" {
  run outlook_scan_secrets "Card: 4111 1111 1111 1111"
  [ "$status" -eq 1 ]
  [ "$output" = "credit-card-shape" ]
}

@test "AWS access key detected" {
  run outlook_scan_secrets "AKIAIOSFODNN7EXAMPLE"
  [ "$status" -eq 1 ]
  [ "$output" = "aws-access-key" ]
}

@test "GitHub PAT detected" {
  run outlook_scan_secrets "Token: ghp_aBcDeFgHiJkLmNoPqRsTuVwXyZ0123456789"
  [ "$status" -eq 1 ]
  [ "$output" = "github-pat" ]
}

@test "Bitwarden URI detected" {
  run outlook_scan_secrets "Open bw://vault/item/12345 to retrieve the secret"
  [ "$status" -eq 1 ]
  [ "$output" = "bitwarden-uri" ]
}

@test "private IPv4 (10.x) detected" {
  run outlook_scan_secrets "Server at 10.0.0.5 is down"
  [ "$status" -eq 1 ]
  [ "$output" = "private-ipv4" ]
}

@test "private IPv4 (192.168.x) detected" {
  run outlook_scan_secrets "Local gateway 192.168.1.1"
  [ "$status" -eq 1 ]
  [ "$output" = "private-ipv4" ]
}

@test "private IPv4 (172.16.x) detected" {
  run outlook_scan_secrets "Internal host 172.16.0.1"
  [ "$status" -eq 1 ]
  [ "$output" = "private-ipv4" ]
}

@test "public IPv4 NOT flagged" {
  run outlook_scan_secrets "Public server 8.8.8.8"
  [ "$status" -eq 0 ]
}

@test "api-key-shape detected" {
  run outlook_scan_secrets "api_key=abcdef0123456789ABCDEF_secret_value_here"
  [ "$status" -eq 1 ]
  [ "$output" = "api-key-shape" ]
}

@test "short api-key-shape (under 20 chars) NOT flagged" {
  run outlook_scan_secrets "token=short"
  [ "$status" -eq 0 ]
}

@test "100KB+ input: treated as clean (size cap, no DoS)" {
  # Build a 100001-char string of safe text
  big_text="$(printf 'a%.0s' $(seq 1 100001))"
  run outlook_scan_secrets "$big_text"
  [ "$status" -eq 0 ]
}