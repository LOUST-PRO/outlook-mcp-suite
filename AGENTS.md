# AGENTS.md — Manifiesto para AI Agents

> This file is for AI agents (Claude Code, GPT, Gemini, Cursor, Cline,
> OpenCode, etc.) **using** this repo's MCP server. It tells you what
> tools exist, how to call them, what to never do, and where to read
> more.
>
> If you are an AI agent **contributing** to this repo's code, see
> `CONTRIBUTING.md` instead (and ask the human operator — `AGENTS.md`
> is for tool users, `CONTRIBUTING.md` is for code authors).

## What this repo gives you

`outlook-mcp-suite` is an MCP (Model Context Protocol) server that lets
you read and, with explicit human approval, act on a user's Outlook /
Microsoft 365 account.

Two implementation paths exist (`graph/` and `playwright/`); see
`LEGAL-NOTICE.md`. **Default to Path A** unless your user has explicit
reason to use Path C.

## What tools exist

### Read (always safe to call)

```
outlook.folders               list folder structure
outlook.list                  list mail in a folder
outlook.read                  read single mail (body capped 1MB)
outlook.search                full-text search
outlook.conversation          thread view
outlook.calendar.list         list calendar events
outlook.calendar.get          single event detail
outlook.calendar.free_busy    free/busy data
outlook.attachments.list      list attachments on a mail
outlook.extract_signature     extract + save signature HTML
```

### Mutate (require user approval via hooks)

```
outlook.propose_move          shadow: shows diff, no action
outlook.apply_move            real: moves mail to folder
outlook.propose_categorize    shadow: shows new categories, no action
outlook.apply_categorize      real: replaces/clears/adds categories
outlook.propose_mark_read     shadow: shows entry ID, no action
outlook.apply_mark_read       real: marks entry as read
outlook.propose_send          shadow: shows compose preview, no action
outlook.apply_send            real: sends mail NOW
outlook.propose_reply         shadow: shows reply preview, no action
outlook.apply_reply           real: replies NOW
outlook.propose_forward       shadow: shows forward preview, no action
outlook.apply_forward         real: forwards NOW
```

## Tool argument conventions

- All entry IDs are opaque strings (~40 chars). Do NOT parse them.
- All date arguments are ISO 8601 (`2026-07-27` or `2026-07-27T15:00:00Z`)
  or natural (`today`, `+7d`, `+1h`). The wrapper normalizes to ISO 8601.
- `--to` accepts comma-separated emails. Display names go in
  `"Name <addr>"` form. The wrapper validates with `MailAddress.TryCreate`
  before sending — invalid addresses are rejected client-side.
- `--body` is plain text. `--html` is HTML; the HTML sanitizer hook
  blocks `<script>`, `<iframe src=javascript:>`, tracking pixels 1×1,
  and `<object>`/`<embed>` pointing to JS.
- `--attachment` accepts file paths; size-capped per Graph limits
  (25MB total per send). Use absolute paths in well-known dirs; the
  hook does path-traversal prevention.

## Hook layer (per-call gates)

The 4 hooks listed in `.claude/hooks/` may BLOCK your tool call. When
blocked, the tool error message tells you exactly what to do to
unblock. Re-read the error; do not retry with the same arguments.

| Hook | What it gates | Unblock strategy |
|---|---|---|
| `require-approval` | any `apply_*` tool | Ask the user to add phrase `"outlook apply: <reason>"` to their next message, OR ensure `--account` is in `config.toml` allowlist |
| `rate-limit` | any `apply_*` after 10/5min/account | Wait 5 minutes, OR ask the user to bump `rate_limit_per_5min` in config |
| `secret-scan` | `apply_send*` / `*_reply` / `*_forward` with body/subject/attachment | Remove PII patterns (SSN, credit card, API keys, bitwarden URIs) from your content |
| `html-sanitizer` | `apply_send*` / `*_reply` / `*_forward` with `--html` | Strip blocked tags, OR drop `--html` (use plain text body) |

The hooks are read-then-act: they inspect the proposed call, and only
allow it if all four gates pass. None of them mutate state — they
either block the call or let it through.

## Account selection (allowlist enforcement)

The `require-approval` hook also verifies that `--account` is in
`config.toml` `[accounts].allowed`. If you try to apply with an
unknown account, the hook blocks. **Do not** invent account names;
only use values explicitly configured by the user.

## Shadow mode contract

When you call `outlook.propose_X`, the wrapper:

1. Returns a structured preview (`{"preview": {...}, "diff": {...}}`)
2. Does NOT mutate anything
3. Does NOT log to the apply ledger
4. Does NOT trigger any hook

You can call `propose` as many times as needed to compose the right
content. You only call `apply` when the user has explicitly approved
(or in this repo's case, when the hook gates have passed).

If the user hasn't yet approved after you proposed:

> Show them the propose preview and ask if it's good. Don't apply.

## What you must NEVER do

- **Do not** fabricate recipient addresses. If you don't have a valid
  email from `outlook.read` output, don't put it in `--to`.
- **Do not** paraphrase the user's email body into a reply and ship it
  without showing the user the proposed reply text first.
- **Do not** include `<script>`, tracking pixels, or external iframes
  in HTML sends. Even if the hook lets it through some day, the
  recipient's mail client may execute it.
- **Do not** rely on the hook being a hard guarantee. The hook may
  have a bug. Your default behavior is "ask the user, don't assume".
- **Do not** mark mail as read/unread in bulk without previewing the
  list first. Category operations are similarly permanent.
- **Do not** invoke `outlook.apply.*` from a non-interactive
  environment (cron, scripts, CI). This MCP is designed for
  interactive use with the user in the loop. Programmatic use must
  go through the official Graph SDK or EWS-with-app-password.

## Reference docs to read before operating

| Before doing this... | Read |
|---|---|
| Any `apply_*` | `docs/shadow-mode.md` |
| Configure a new account | `docs/onboarding-real-account.md` |
| Debug missing apply entries | `docs/ledger-format.md` |
| Hit 429 / rate limit | `docs/performance-cost.md` |
| File a bug | the issue template for your path |
| Understand the trust model | `LEGAL-NOTICE.md`, `COMPARISON.md` |

## What this manifest does NOT cover

- Behavior outside this repo's tools (other MCPs, Bash, Read, etc.)
- Trust boundaries between this MCP and other systems
- Reading non-Outlook data (calendar invitations from Google,
  contacts from Apple, etc.) — out of scope
- Group / shared mailbox scenarios — use Outlook desktop for those

If a request requires any of the above, decline politely with a reason
that points to the right tool.