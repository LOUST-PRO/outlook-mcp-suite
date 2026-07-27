#!/usr/bin/env bats
# html-sanitizer.bats — Phase 0.5 tests for lib-outlook-html.sh
#
# Run with:
#   bats tests/hooks/html-sanitizer.bats

HOOK_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../../.claude/hooks" && pwd)"
LIB="${HOOK_DIR}/_lib/lib-outlook-html.sh"

setup() {
  # shellcheck disable=SC1090
  . "$LIB"
}

@test "empty input: clean" {
  run outlook_scan_html ""
  [ "$status" -eq 0 ]
}

@test "plain HTML: clean" {
  run outlook_scan_html "<p>Hello world</p><a href='https://example.com'>link</a>"
  [ "$status" -eq 0 ]
}

@test "<script> tag detected" {
  run outlook_scan_html "<p>Hi</p><script>alert(1)</script>"
  [ "$status" -eq 1 ]
  [ "$output" = "script-tag" ]
}

@test "<script src=> detected" {
  run outlook_scan_html "<script src='evil.js'></script>"
  [ "$status" -eq 1 ]
  [ "$output" = "script-tag" ]
}

@test "<iframe src=javascript:> detected" {
  run outlook_scan_html "<iframe src='javascript:alert(1)'></iframe>"
  [ "$status" -eq 1 ]
  [ "$output" = "iframe-js-src" ]
}

@test "<iframe src=data:text/html> detected" {
  run outlook_scan_html "<iframe src=\"data:text/html,<script>alert(1)</script>\"></iframe>"
  [ "$status" -eq 1 ]
  [ "$output" = "iframe-js-src" ]
}

@test "<iframe src=normal URL> NOT flagged" {
  run outlook_scan_html "<iframe src='https://youtube.com/embed/abc'></iframe>"
  [ "$status" -eq 0 ]
}

@test "1x1 tracking pixel (<img width=1>) detected" {
  run outlook_scan_html "<img src='https://tracker.example/pixel' width='1' height='1' />"
  [ "$status" -eq 1 ]
  [ "$output" = "tracking-pixel" ]
}

@test "1x1 tracking pixel (height=1 only) detected" {
  run outlook_scan_html "<img src='https://tracker.example/pixel' height='1' />"
  [ "$status" -eq 1 ]
  [ "$output" = "tracking-pixel" ]
}

@test "normal-sized <img> NOT flagged" {
  run outlook_scan_html "<img src='logo.png' width='200' height='100' />"
  [ "$status" -eq 0 ]
}

@test "<object type=javascript> detected" {
  run outlook_scan_html "<object type='application/x-javascript' data='evil.js'></object>"
  [ "$status" -eq 1 ]
  [ "$output" = "object-embed-js" ]
}

@test "<embed src=javascript:> detected" {
  run outlook_scan_html "<embed src='javascript:alert(1)' />"
  [ "$status" -eq 1 ]
  [ "$output" = "object-embed-js" ]
}

@test "onclick= handler detected" {
  run outlook_scan_html "<button onclick='steal()'>Click me</button>"
  [ "$status" -eq 1 ]
  [ "$output" = "on-event-handler" ]
}

@test "onerror= handler detected" {
  run outlook_scan_html "<img src='x' onerror='alert(1)' />"
  [ "$status" -eq 1 ]
  [ "$output" = "on-event-handler" ]
}

@test "onmouseover= handler detected" {
  run outlook_scan_html "<div onmouseover='track()'>hover me</div>"
  [ "$status" -eq 1 ]
  [ "$output" = "on-event-handler" ]
}

@test "HTML comments stripped before scan (so pattern in comment doesn't trigger)" {
  # The sample below contains '<script>' inside a comment — should NOT trigger.
  run outlook_scan_html "<p>Hi</p><!-- <script>alert(1)</script> -->"
  [ "$status" -eq 0 ]
}

@test "CDATA stripped before scan" {
  # JS code in CDATA shouldn't trigger (it's a sample, not executable in mail)
  run outlook_scan_html "<![CDATA[<script>alert(1)</script>]]>"
  [ "$status" -eq 0 ]
}

@test "1MB+ input: treated as clean (size cap, no DoS)" {
  big_html="<p>$(printf 'a%.0s' $(seq 1 1000001))</p>"
  run outlook_scan_html "$big_html"
  [ "$status" -eq 0 ]
}