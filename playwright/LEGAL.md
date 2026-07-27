# LEGAL — Path C (playwright/) Terms of Service implications

> **If you are reading this before considering using or contributing
> to the `playwright/` subtree, STOP and read this entire file.**
>
> Path A (`graph/`) is the supported path. Path C is a community
> fallback for users with genuine technical blockers. It carries
> real legal risk.

## What Path C does

Path C drives a headless Chromium browser (via Playwright) against
`https://outlook.live.com` or `https://outlook.office.com`. The
browser:

1. Loads the page like any other browser would
2. Submits the login form interactively during one-time setup
3. Persists session cookies in an encrypted file
4. Reuses the cookies for subsequent MCP tool calls
5. Issues DOM queries (CSS selectors, accessibility tree reads) to
   extract mail, calendar, and folder structure

This is **not** a documented API. It is browser automation against
a public web application.

## What Microsoft's terms say about this

From the **Microsoft Services Agreement** (effective August 2024,
Section 2.a — "Code of Conduct"):

> Don't do anything that could harm Microsoft, our customers, or
> our services. This includes using the services in any way that
> could disable, overburden, impair, or otherwise interfere with
> any of our services.

And from **Section 3.b — "Automated Access"**:

> You may not use automated means (robots, bots, scripts, etc.) to
> access the Services or extract data from the Services, except as
> permitted by Microsoft (for example, through our published APIs).

Browser automation against `outlook.live.com` is **not permitted
under Section 3.b**, because it is not access through a published API.

## What that means concretely

| Action | Likely outcome | Severity |
|---|---|---|
| One-time login with cookies | Session works | None initially |
| Low-rate polling (≥60s between calls) | Works for weeks/months | None visible |
| High-rate polling (<60s) | CAPTCHA appears, may require re-login | Mild disruption |
| Aggressive scraping (every few seconds) | Account flagged for review | Moderate |
| Sustained automation | Account force-invalidated, may require reset | High |
| Sustained automation against M365 corporate tenant | IT may block your account | High |
| Continued use after warning | Account suspended | Severe |

Microsoft also publishes a separate **Acceptable Use Policy** that
covers anti-bot detection. Their enforcement is automated and
sometimes inconsistent — a user following every reasonable practice
can still be flagged.

## What this repo's maintainers do and don't promise

**Do promise:**

- The legal separation of Path A (maintained, ToS-compliant) from
  Path C (community, ToS-questionable) is real and visible
- The ToS implications are surfaced here, in `LEGAL-NOTICE.md`,
  in `/README.md`, in `/COMPARISON.md`, in `playwright/README.md`,
  and in CONTRIBUTING.md
- No maintainer encouragement to use Path C for production or for
  accounts you cannot afford to lose
- Bug reports about Path C are acknowledged but may not be fixed
  promptly (Path C is community-maintained)

**Do NOT promise:**

- That Path C will continue to work tomorrow or next week
- That Microsoft will not flag, suspend, or terminate your account
- That using Path C will not violate the MSA or other agreements
- That any specific workaround or mitigation is permanent

## What you should do before using Path C

1. **Read and understand the MSA yourself.** This file is not legal
   advice; it is a summary of the maintainer's understanding.
2. **Use a sacrificial account.** If you have a spare MSA account
   that holds no important mail, use it. Do NOT use Path C against
   your primary account.
3. **Set conservative polling intervals.** The wrapper's default is
   60s. Don't override it to <30s.
4. **Don't use Path C from an IP your corporate IT watches.** Many
   enterprises have egress monitoring that flags automation patterns.
5. **Have a plan if your account gets flagged.** Microsoft sometimes
   requires identity verification (phone number, ID) to recover a
   flagged account. Make sure you have access to whatever
   verification method your account uses.

## Why the maintainers include Path C at all

Three reasons:

1. **Visibility.** Hiding that browser automation exists doesn't make
   it stop; it just pushes users to worse-quality tools that don't
   warn them about the risks.
2. **Research.** Anti-bot researchers, browser automation engineers,
   and academic researchers need a starting point. Path C with
   prominent ToS warnings is a more responsible starting point than
   hidden alternatives.
3. **Sacrificial workflows.** Users with low-stakes accounts (alerts,
   newsletters, secondary mailbox) genuinely benefit from a path that
   doesn't require Azure registration.

If you disagree with Path C existing in this repo, open an issue
against `LEGAL-NOTICE.md` with your reasoning. Maintainers will
consider deprecation but, per precedent in similar OSS projects
(yt-dlp, gallery-dl), are likely to keep it with stronger warnings.

## License

This file is part of `outlook-mcp-suite`, Apache-2.0. See
[`/LICENSE`](../LICENSE).