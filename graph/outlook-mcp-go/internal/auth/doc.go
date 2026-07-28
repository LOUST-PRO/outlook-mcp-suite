// Package auth implements OAuth 2.0 device-code flow against Microsoft
// Identity Platform and age-encrypted token persistence on disk.
//
// Three sub-packages live here:
//
//   config.go        — TOML loader for ~/.config/lzt-outlook/config.toml
//   token_store.go   — age-encrypted blob store at ~/.local/share/lzt-outlook/tokens/
//   device_code.go   — MSAL Go device-code flow with terminal UX
//   manager.go       — high-level "give me a bearer token for account X"
//
// The hook layer (.claude/hooks/_lib/lib-outlook-allowlist.sh) reads the
// same config.toml [accounts].allowed list. The MCP binary and the
// shell hooks share the config but not the runtime state (the hooks
// run before the binary even loads).
package auth
