# `.claude/hooks/` — per-call gates for `outlook.apply.*`

This directory contains **shell scripts** that Claude Code invokes
**before** any `mcp__outlook__apply_*` tool call. Each hook may
**block** the call, **allow** it through, or **ask** the user.

> **⚠️ Phase 0 SPEC.** As of Phase 0, no scripts exist yet. Phase 0.5
> implements the 4 hooks below. Until Phase 0.5 ships, the hooks are
> **not installed** and `outlook.apply.*` calls go through without
> gating. Do not deploy this MCP to a real account until Phase 0.5
> is complete.

## Hook inventory (Phase 0.5 target)

| Script | Gates | Blocks when |
|---|---|---|
| `pre-tool-outlook-apply-require-approval.sh` | all `outlook.apply.*` | (a) missing literal phrase `"outlook apply: <reason>"` in recent user message, OR (b) `--account` not in `config.toml` `[accounts].allowed` |
| `pre-tool-outlook-rate-limit.sh` | all `outlook.apply.*` | more than 10 applies per account in rolling 5 minutes |
| `pre-tool-outlook-secret-scan.sh` | `apply_send*` / `*_reply` / `*_forward` with body/subject/attachment | body/subject contains PII pattern (SSN, credit card, API key, bitwarden URI) |
| `pre-tool-outlook-html-sanitizer.sh` | `apply_send*` / `*_reply` / `*_forward` with `--html` | HTML contains `<script>`, `<iframe src=javascript:>`, tracking pixel 1×1, or `<object>`/`<embed>` pointing to JS |

## Hook order

Hooks fire in **this order** (Claude Code invokes them sequentially
via the user's `~/.claude/settings.json` `PreToolUse` hooks array):

```
1. require-approval        ← cheapest; catches missing-phrase + bad-account
2. rate-limit              ← cheap; catches 11th apply in 5min
3. secret-scan             ← moderate; runs regex over body/subject
4. html-sanitizer          ← moderate; parses HTML if --html
```

If any hook blocks, the chain stops. Subsequent hooks don't run.

## Helper libs (`.claude/hooks/_lib/`)

Common helpers extracted into `_lib/`:

| File | Purpose |
|---|---|
| `_lib/hook-log.sh` | JSONL logging to `~/.local/share/lzt-outlook/hooks.log` (per-hook + per-decision) |
| `_lib/phrase-detect.sh` | Look-back regex over recent user messages for the literal phrase |
| `_lib/account-allowlist.sh` | Read `config.toml` `[accounts].allowed` (cached for 60s) |
| `_lib/rate-counter.sh` | Apply counter per account in rolling 5min window (file-based, JSON state) |
| `_lib/secret-patterns.sh` | PII regex patterns (SSN, credit card, AWS key, GitHub PAT, bitwarden URI, etc.) |
| `_lib/html-patterns.sh` | Blocked HTML tag/attribute patterns |
| `_lib/json-envelope.sh` | Build the standard `{success, data, error, metadata}` envelope used by all hooks |

## Hook contract

Each hook receives the standard Claude Code hook environment:
`$CLAUDE_TOOL_CALL_JSON`, `$CLAUDE_HOOK_NAME`, `$CLAUDE_USER_MESSAGE`,
`$CLAUDE_CONFIG_DIR`. The hook emits:

```json
{
  "decision": "allow" | "block" | "ask",
  "additionalContext": "string",
  "tool_use_block_message": "string (only when decision=block)"
}
```

For `block`, the `tool_use_block_message` is the literal error text
the agent will see. Make it actionable: tell the agent exactly what
to do to unblock (e.g., "Add phrase 'outlook apply: confirm' to
your next message, OR remove 'foo@x.com' from --to (not in allowlist)").

For `ask`, the agent stops and surfaces the question to the user.

For `allow`, the tool fires normally.

## Hook log format

`~/.local/share/lzt-outlook/hooks.log` (one JSON line per hook invocation):

```json
{
  "ts": "2026-07-27T15:42:11.123Z",
  "hook": "pre-tool-outlook-apply-require-approval.sh",
  "tool": "outlook.apply.send",
  "account": "personal",
  "decision": "block",
  "reason": "missing literal phrase",
  "args_hash": "sha256:..."
}
```

SHA-256 chained (same pattern as `lzt-cloud-cost-gate-mcp`).

## Hook installation

After Phase 0.5:

```bash
# Dry-run: show what install would do
make install-hooks-dry-run

# Real install (after operator approval)
make install-hooks
```

`make install-hooks` symlinks the 4 scripts into `~/.claude/hooks/`
and updates `~/.claude/settings.json` `hooks.PreToolUse` to register
them. Stage-local-first applies: the install command requires
explicit operator approval per the project's CONTRIBUTING.md §
AGENT-AUTOMATED.

## Tests

`tests/hooks/` (Phase 0.5) will contain `bats` tests:

```
tests/hooks/
├── require-approval.bats
├── rate-limit.bats
├── secret-scan.bats
└── html-sanitizer.bats
```

Run with:

```bash
make phase0.5-test
```

## Hook security notes

These hooks are the **only** security gate between the AI agent and
`outlook.apply.*` operations. They must:

- Be POSIX-sh compatible (not bash-isms beyond `[[`, `printf -v`,
  `${var,,}`, `${var^^}`)
- Never `exec` anything that reads the agent's prompt context
- Never write outside `~/.local/share/lzt-outlook/`
- Never include `curl`, `wget`, `eval`, or arbitrary file paths
- Have file permissions `0755` (rwxr-xr-x); owned by the user, not root
- Be re-runnable idempotently (no state outside `_lib/` JSONL logs)
- Have a 2-second timeout (set via Claude Code hook config) to
  prevent hangs blocking the agent loop

## License

Apache-2.0, same as the repo root. See [`/LICENSE`](../../LICENSE).