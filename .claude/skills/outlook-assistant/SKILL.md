---
name: outlook-assistant
description: |
  Use the outlook-mcp-suite MCP server (when available) to read and (with
  explicit user approval) act on the user's Outlook / Microsoft 365
  account. Default to read-only tools (folders, list, read, search,
  conversation, calendar.*, attachments.*). For mutates (move,
  categorize, mark-read, send, reply, forward), always call the matching
  `propose_*` tool first, surface the preview to the user, and only
  call `apply_*` after the user explicitly approves. Honour the
  4-hook gate (require-approval, rate-limit, secret-scan,
  html-sanitizer); read the error message and adjust before retrying.
---

# outlook-assistant

This skill wraps the `outlook-mcp-suite` MCP server for use inside
Claude Code. It does not invent tools — it enforces the protocol
around the existing tools and helps you avoid the common mistakes
(fabricating recipients, paraphrasing replies without preview, hitting
the rate limit, leaking PII in HTML sends).

## When to invoke this skill

- The user mentions Outlook, email, mail, calendar, schedule, inbox,
  attachments, signature — and the `outlook-mcp-suite` MCP server is
  available (check `mcp__outlook__*` tool names).
- The user asks "what's in my inbox", "when is my next meeting", "find
  the email from X about Y", "draft a reply", "send this for me",
  "categorize these as invoices", "extract my signature", etc.

## When NOT to invoke this skill

- The user asks about email from a non-Outlook source (Gmail, FastMail,
  ProtonMail, etc.) — out of scope, decline politely.
- The user wants shared mailbox, group mailbox, public folders,
  compliance search, or admin-level M365 operations — out of scope,
  point them to Outlook desktop or M365 admin center.
- The `outlook-mcp-suite` MCP server is not installed — tell the user
  to install it; do not try to fake the tools with Bash + Graph SDK
  directly (that bypasses the propose/apply gate).
- The user wants you to mark mail as read in bulk without preview —
  refuse, even if they ask. Surface the risk; let them decide.

## Workflow

### Read path (most common)

1. `mcp__outlook__folders` — discover folder structure if not known
2. `mcp__outlook__list --folder <folder> --limit 20 --unread` — narrow
   down to candidates
3. `mcp__outlook__search --query "..." --from ... --after ... --before ...`
   — for free-text searches
4. `mcp__outlook__read <entry_id>` — fetch full body (capped 1MB)
5. `mcp__outlook__attachments__list <entry_id>` — see attachments
6. `mcp__outlook__conversation <entry_id>` — see thread context

For calendar:

1. `mcp__outlook__calendar__list --start today --end "+7d"` — upcoming
2. `mcp__outlook__calendar__get <event_id>` — single event details
3. `mcp__outlook__calendar__free_busy --email <addr>` — for meeting
   scheduling

For signatures:

1. `mcp__outlook__extract_signature <entry_id> --output /tmp/sig.html`
2. The user can paste /tmp/sig.html into their mail client's signature
   settings.

### Mutate path (always with approval)

1. **Find candidates via read tools first.** Don't move mail you
   haven't read; don't reply to mail you haven't read.
2. **Call `propose_*` first.** Read the preview carefully.
3. **Surface the preview to the user.** Quote the affected
   subject/from/to/diff in your reply.
4. **Wait for explicit approval.** Do not call `apply_*` until the
   user says "yes, apply that" or equivalent. The hook `require-approval`
   enforces a literal phrase; surface that constraint to the user.
5. **If hook blocks, read the error.** Common causes:
   - missing literal phrase (`outlook apply: <reason>`)
   - `--account` not in allowlist (add to config.toml)
   - rate-limit hit (wait 5 minutes)
   - PII detected in body/subject (redact)
   - HTML blocked tags (drop `--html`, use plain text body)

### Send / reply / forward especially

These are the most consequential tools. Always:

- Confirm recipient addresses against an earlier `outlook.read` or
  `outlook.search` output — don't invent addresses.
- For reply: surface the proposed reply text verbatim and wait for
  "ship it" before calling `apply_reply`.
- For forward: surface the proposed body and the recipient list
  before calling `apply_forward`.
- For send: surface subject + body + recipients + attachments. Wait
  for "ship it" before calling `apply_send`.
- For HTML sends: validate that the HTML doesn't contain `<script>`,
  `<iframe src=javascript:>`, tracking pixels 1×1, or `<object>`
  pointing to JS. If it does, rewrite or use plain text body.

### Calendar mutates

Calendar propose/apply tools (create/update/delete/respond) land in
Phase 2. Same protocol: propose → preview → user approval → apply.

## Reference docs in this repo

| Before doing this... | Read |
|---|---|
| Any `apply_*` | `/docs/shadow-mode.md` (Phase 5) |
| Configure a new account | `/docs/onboarding-real-account.md` (Phase 5) |
| Debug missing apply entries | `/docs/ledger-format.md` (Phase 5) |
| Hit 429 / rate limit | `/docs/performance-cost.md` (Phase 5) |

## Hard "don't" list

- Don't fabricate recipient addresses. Don't paraphrase the user's
  mail into a reply without preview. Don't put `<script>` in HTML.
- Don't invoke `outlook.apply.*` from cron, scripts, or CI — only
  from interactive sessions with the user in the loop.
- Don't rely on the hook being a hard guarantee. Your default is
  "ask the user, don't assume".
- Don't send mail at 3am even if hooks pass. Surface your intent,
  not the action.
- Don't mark mail as read/unread in bulk without previewing the list.

If a request pushes against any of these, decline politely and
explain which rule applies.