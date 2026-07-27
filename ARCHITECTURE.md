# Architecture

This document describes the **technical** architecture: how the three
components (Path A wrapper, Path C wrapper, unified CLI) talk to each
other and how the MCP server is structured. For the **why** (rationale
for two paths), see `COMPARISON.md`.

## Component diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                      Claude Code Agent                                │
│  + invokes mcp__outlook__* tools                                       │
│  + emits hooks: pre-tool-{apply-require-approval, rate-limit,         │
│                     secret-scan, html-sanitizer}                       │
└─────────────────────────────┬────────────────────────────────────────┘
                              │ stdio (JSON-RPC, MCP)
                              ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    lzt-outlook unified CLI                            │
│  + reads ~/.config/lzt-outlook/config.toml                             │
│  + dispatches based on `server.path` = "graph" | "playwright"          │
│  + same JSON contract for both paths                                   │
└────────────┬────────────────────────────────────┬────────────────────┘
             │                                    │
   server.path=graph                  server.path=playwright
             │                                    │
             ▼                                    ▼
┌───────────────────────────────┐  ┌───────────────────────────────────┐
│   graph/outlook-mcp-go         │  │  playwright/outlook-scraper-go    │
│                                 │  │                                   │
│  ToolAnnotations:               │  │  ToolAnnotations:                 │
│    ReadOnlyHint:    true        │  │    ReadOnlyHint:    true          │
│    DestructiveHint: false       │  │    DestructiveHint: false         │
│    IdempotentHint:  true        │  │    IdempotentHint:  true          │
│    OpenWorldHint:   false       │  │    OpenWorldHint:   false         │
│                                 │  │                                   │
│  tools registered (Phase 1+2):  │  │  tools registered (Phase 3+):     │
│    11 read-only + 3 mutative    │  │    subset of Path A tools         │
│    via propose_X / apply_X      │  │    (move/categorize = NO          │
│    with shadow mode             │  │     for path C — scraping is      │
│    (12 + 6 = 18 tools total)    │  │     read-mostly by nature)        │
│                                 │  │                                   │
│  storage:                        │  │  storage:                         │
│    - refresh token: Bitwarden   │  │    - browser cookies: encrypted   │
│      via env var OUTLOOK_       │  │      file with passphrase from    │
│      REFRESH_TOKEN              │  │      env var OUTLOOK_KEY          │
│    - apply audit ledger:        │  │    - no apply ledger (no shadow   │
│      JSONL SHA-256 chained      │  │      mode for path C)             │
│      in ~/.local/share/         │  │                                   │
│      lzt-outlook/ledger/        │  │    see: playwright/LEGAL.md       │
└────────────┬────────────────────┘  └────────────┬──────────────────────┘
             │                                    │
             ▼                                    ▼
┌───────────────────────────────┐  ┌───────────────────────────────────┐
│   Microsoft Graph              │  │   web.outlook.live.com /          │
│   https://graph.microsoft.com  │  │   web.outlook.office.com          │
│                                 │  │   via headless Chromium           │
│   auth: OAuth 2.0 device-code  │  │                                   │
│   profile: app registered in    │  │   auth: cookie session after      │
│   user Azure portal, MSA       │  │   one-time interactive login      │
│   personal                      │  │   (no Azure, no app, no OAuth)    │
└─────────────────────────────────┘  └───────────────────────────────────┘
```

## MCP server shape

Both paths use the same tool surface (Path C is a subset):

| Tool | Path | ReadOnlyHint | Phase |
|---|---|---|---|
| `outlook.folders` | A, C | true | 1 |
| `outlook.list` | A, C | true | 1 |
| `outlook.read` | A, C | true | 1 |
| `outlook.search` | A, C | true | 1 |
| `outlook.conversation` | A, C | true | 1 |
| `outlook.calendar.list` | A, C | true | 1 |
| `outlook.calendar.get` | A, C | true | 1 |
| `outlook.calendar.free_busy` | A, C | true | 1 |
| `outlook.attachments.list` | A, C | true | 1 |
| `outlook.extract_signature` | A, C | true | 1 |
| `outlook.propose_move` | A only | true (propose) | 2 |
| `outlook.apply_move` | A only | false (mutative) | 2 |
| `outlook.propose_categorize` | A only | true (propose) | 2 |
| `outlook.apply_categorize` | A only | false (mutative) | 2 |
| `outlook.propose_mark_read` | A only | true (propose) | 2 |
| `outlook.apply_mark_read` | A only | false (mutative) | 2 |
| `outlook.propose_send` | A only | true (propose) | 2 |
| `outlook.apply_send` | A only | false (mutative) | 2 |
| `outlook.propose_reply` | A only | true (propose) | 2 |
| `outlook.apply_reply` | A only | false (mutative) | 2 |
| `outlook.propose_forward` | A only | true (propose) | 2 |
| `outlook.apply_forward` | A only | false (mutative) | 2 |

= 11 read-only tools (shared) + 12 mutate tools (Path A only, half propose / half apply).

Path C **does not** implement propose/apply mutates. Scraping is read-mostly
by nature (composing a new email via DOM is fragile and not in scope).

## Hook layer (per-call gates)

The 4 hooks in `.claude/hooks/` fire before any `mcp__outlook__apply_*`
tool call:

```
Claude Code turn starts
    │
    ▼
agent emits `mcp__outlook__apply_send(...)`
    │
    ├─→ pre-tool-outlook-apply-require-approval.sh
    │       ↳ blocks if missing literal phrase OR --account not in allowlist
    │
    ├─→ pre-tool-outlook-rate-limit.sh
    │       ↳ blocks if >10 applies per account in rolling 5min
    │
    ├─→ pre-tool-outlook-secret-scan.sh
    │       ↳ blocks if body/subject contains PII pattern
    │
    ├─→ pre-tool-outlook-html-sanitizer.sh
    │       ↳ blocks if --html contains <script>, iframe[src=javascript:],
    │                  or tracking pixel 1×1
    │
    ▼
all hooks passed → tool fires
    │
    ▼
agent sees tool_use_result → next iteration
```

Each hook emits `additionalContext` with explicit instructions on how
to unblock itself. See `.claude/hooks/README.md` for full spec.

## Apply ledger (Path A mutates only)

```
~/.local/share/lzt-outlook/ledger/YYYY-MM-DD.jsonl

  {"seq": 42, "ts": "...", "account": "personal",
   "tool": "outlook.apply.send", "args_hash": "sha256:...",
   "subject_hash": "sha256:...", "to_hashes": ["sha256:..."],
   "cc_hashes": [], "attachment_hashes": [],
   "status": "executed", "duration_ms": 312,
   "prev_hash": "sha256:<previous entry hash>",
   "self_hash": "sha256:..." // computed over all of the above
  }
```

SHA-256 chained ensures **tamper detection** on the audit trail; a
modified entry breaks the chain and is detectable on replay. Same pattern
as `lzt-cloud-cost-gate-mcp` and `lzt-hub-sync B1.13`.

## Account pinning + allowlist

`config.toml`:

```toml
[accounts]
allowed = ["personal", "sacrificial"]   # hook-validated
default = "personal"
```

The agent cannot call `outlook.apply.* --account foo` if `foo` is not
in `allowed`. To add a new account: edit `config.toml` manually, restart
the MCP server.

This is **defense-in-depth**: it stops the LLM from accidentally
inventing account names that don't exist or names that exist but
shouldn't be touched from this Claude Code install.

## Unified CLI dispatch

```bash
lzt-outlook [global-flags] <command> [command-flags]

# global flags
--config PATH         # default ~/.config/lzt-outlook/config.toml
--account NAME        # default from config.toml [accounts].default
--json                # output as JSON envelope (default for --ready; AI-friendly)
--human               # output human-readable (for terminal use)

# commands
auth login --client-id <id>    # device-code flow
auth logout
mail folders
mail list [--folder X] [--unread] [--limit N] [--full]
mail read <entry-id>
mail search --query "..." [--from X] [--after X] [--before X]
mail conversation <entry-id>
mail attach-list <entry-id>
mail extract-signature <entry-id> --output sig.html
calendar list [--start X] [--end X] [--limit N] [--full]
calendar get <entry-id>
calendar free-busy --email <addr>
mail propose-move <entry-id> --to-folder X
mail apply-move <entry-id> --to-folder X        # requires literal phrase
mail propose-categorize <entry-id> --set X
mail apply-categorize <entry-id> --set X
mail propose-mark-read <entry-id>
mail apply-mark-read <entry-id>
mail propose-send --to X --subject Y --body Z [--html] [--attachment F]
mail apply-send --to X --subject Y --body Z [--html] [--attachment F]  # requires literal phrase
mail propose-reply <entry-id> --body Y [--reply-all] [--draft]
mail apply-reply <entry-id> --body Y [--reply-all]   # requires literal phrase
mail propose-forward <entry-id> --to X [--body Y]
mail apply-forward <entry-id> --to X [--body Y]      # requires literal phrase

# JSON contract envelope (default AI-friendly):
{
  "success": true | false,
  "command": "<full command line>",
  "data": <T> | null,
  "error": { "code": "STRING", "message": "STRING" } | null,
  "metadata": { "count": N, "duration_ms": N }
}
```

## Phase order

```
Phase 0       ← you are here (scaffold only)
Phase 0.5     4 hooks (real shell scripts)
Phase 1       Path A read-only (11 tools)
Phase 2       Path A mutates (12 tools with propose/apply)
Phase 3       Path C structure + LEGAL.md
Phase 4       Unified CLI dispatcher
Phase 5       Onboarding docs
```

See `docs/` (placeholder until Phase 5).
