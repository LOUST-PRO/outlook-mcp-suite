# LEGAL NOTICE — read before contributing, using, or filing issues

This repository hosts **two implementation paths** for the same problem: an
automated bridge between AI agents and a user's Outlook/Microsoft 365
account. The two paths are NOT equally maintained and NOT equally safe.

## Path A: `graph/` — Microsoft Graph (PRIMARY, maintained)

`graph/` uses **Microsoft Graph API** with **OAuth 2.0 device-code flow**,
registered as a personal app in the user's own Azure portal. This is the
official Microsoft-blessed way to read mail, calendar, and contacts.

- **Maintained by**: project maintainers (responds to issues within ~7 days)
- **Reliability**: high — Graph is a documented, supported API; MS may deprecate endpoints with 24-month notice
- **Legal exposure**: standard — Microsoft permits personal app registration
  and OAuth flow under default MSA terms
- **Operational exposure**: standard — rate limits apply (per-user throttling
  per endpoint); see `graph/outlook-mcp-go/docs/rate-limits.md` (Fase 1)
- **Verification**: smoke-tested against a sacrificial MSA account before
  any tagged release

This path is suitable for **production use** by individuals on their own
accounts.

## Path C: `playwright/` — web.outlook.com scraping (COMMUNITY, may break)

`playwright/` drives a headless browser against `https://outlook.live.com`
or `https://outlook.office.com` using **Playwright**. The user logs in with
their normal password + 2FA in a one-time setup; the wrapper reuses the
browser session for subsequent calls.

- **Maintained by**: community contributors; **NOT a maintained path**
- **Reliability**: low — Microsoft may change the web UI at any time without
  notice; CAPTCHA may be triggered; session may be force-invalidated
- **Legal exposure**: **higher** — Microsoft's Terms of Service prohibit
  automated access to the services except via official APIs (Graph,
  EWS-with-app-password). Using this path may violate those terms. The
  maintainers of this repository **do not encourage** this path and
  accept **no liability** for accounts flagged, suspended, or terminated
  as a result of using it.
- **Operational exposure**: polling intervals of ≥60 seconds are STRONGLY
  RECOMMENDED; faster intervals will trigger anti-bot defenses quickly
- **Verification**: only tested against mock HTML; no live testing is
  performed in CI because such testing would constitute the very ToS
  violation `playwright/LEGAL.md` warns against

**Use this path ONLY if** you understand and accept the ToS risk, and you
cannot use Path A for genuine technical reasons (e.g. locked-out corporate
account with MFA you can't unbind from Graph).

## What this means for contributors

| If you want to... | Use path | And read |
|---|---|---|
| Open an issue about Path A | `graph/.github/ISSUE_TEMPLATE/bug.md` | N/A |
| Contribute to Path A | open PR against `graph/` subtree | `CONTRIBUTING.md` § AGENT-AUTOMATED, `graph/CONTRIBUTING.md` |
| Open an issue about Path C | `playwright/.github/ISSUE_TEMPLATE/community.md` | N/A |
| Contribute to Path C | open PR against `playwright/` subtree | `playwright/LEGAL.md`, `playwright/CONTRIBUTING.md` |
| Send a legal notice (DMCA, ToS complaint, etc.) | email maintainer | This file |

## What this means for users

**Default to Path A.** Path C exists because Microsoft removed Basic Auth
for IMAP (October 2023), which historically was the path-of-last-resort for
no-Azure setups. Path C is the next-best-of-bad-options, not a blessed
alternative.

**Do not use Path C in production or against accounts you cannot afford to
lose.** Anti-bot heuristics + ToS enforcement are real, not theoretical.

## Disclaimer of warranty

THE SOFTWARE IN BOTH PATHS IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY
KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
SEE `LICENSE` (Apache-2.0) FOR FULL TEXT.

IN PARTICULAR, THE MAINTAINERS MAKE NO REPRESENTATION THAT USE OF EITHER
PATH WILL NOT RESULT IN ACCOUNT SUSPENSION, TERMINATION, OR OTHER ACTION
BY MICROSOFT. THE USER ASSUMES ALL RISK.
