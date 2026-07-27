# `docs/` — onboarding, shadow-mode, ledger-format, performance-cost

This directory will hold user-facing documentation for `outlook-mcp-suite`
once Phase 5 lands.

## Planned documents

| Document | Purpose | Phase |
|---|---|---|
| `onboarding-real-account.md` | Step-by-step Azure portal walkthrough + device-code flow | 5 |
| `shadow-mode.md` | How propose/apply work; what the preview contains; when to call apply | 5 |
| `ledger-format.md` | SHA-256-chained JSONL spec for the apply audit ledger | 5 |
| `performance-cost.md` | Graph API rate limits, anti-bot detection thresholds, recommended polling intervals | 5 |

## Status

Phase 0: this `README.md` exists as a placeholder. The four
documents above will land in Phase 5 alongside the first tagged
release.

## Why these four

- **onboarding-real-account** — the single biggest barrier to
  adoption is the Azure portal walkthrough. A clear 5-minute
  walkthrough doubles completion rate vs. a forum-thread "here's
  what I did".
- **shadow-mode** — agents that don't understand propose/apply will
  call apply with garbage. This doc is the formal contract.
- **ledger-format** — operators want to know what they can audit.
  Specifying the JSONL chain lets them write their own auditors.
- **performance-cost** — rate limits are non-obvious and depend
  on the user's tenant tier. Documenting them prevents wasted
  debugging time.

## License

Apache-2.0, same as the repo root. See [`/LICENSE`](../LICENSE).