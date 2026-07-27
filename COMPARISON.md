# Comparison — path A (Graph) vs path C (Playwright)

This is the **decision document** for users who have to choose between
the two paths. It does NOT recommend Path C for production use; see
`LEGAL-NOTICE.md` first.

## Decision table

| Concern | Path A (Graph) | Path C (Playwright) | Winner |
|---|---|---|---|
| **Legal / ToS** | Personal app registration permitted by Microsoft under default MSA terms | Browser automation is explicitly prohibited by Microsoft's ToS for automated access (the section they carved out is "official APIs only") | Path A |
| **Reliability** | High — Graph is documented, supported, deprecation gives 24mo notice | Low — UI changes weekly, MS may force session invalidation with no notice | Path A |
| **Performance** | Direct HTTPS — typical list call <500ms, single mail <1s | Headless browser launch + DOM parse — typical list call 5-15s, single mail 3-8s | Path A |
| **Latency** | Same as above. Webhooks supported for push (~immediate) | Polling only — minimum 60s recommended to avoid anti-bot; typical 5-15min | Path A |
| **Multi-account** | Native — each account = one device-code flow, isolated token store | Native per browser profile, but anti-bot heuristics may flag cross-account | Path A |
| **Setup effort** | 5 minutes manual Azure portal (one-time, ~forever) | 1 minute interactive login (one-time, repeats every ~30d when MS forces re-auth) | Path C (borderline) |
| **Compliance** | Auditable via Azure portal + DevTools + Graph logs | Only visible to the user; no operator-side audit trail | Path A |
| **Maintenance** | Maintainers track MS Graph changes (~few/year) | Community tracks UI changes (high churn) | Path A |
| **Send/Reply/Forward** | Native propose/apply with shadow mode + audit ledger | NOT IMPLEMENTED — scraping DOM for new mail is fragile, not in scope | Path A |
| **Calendar** | Full support (list/create/update/delete/respond) | Limited (list/get only) | Path A |
| **Attachments** | Full support (list + extract) | Limited (read-only) | Path A |
| **Machine-readable output** | Always JSON envelope (per ARCHITECTURE.md § JSON contract) | Same | tie |
| **Cost to maintainers** | ~100 LOC/wk | ~500 LOC/wk if actively scraped against MS UI churn | Path A |
| **Anti-bot detection** | N/A (official API) | Real risk; CAPTCHA may appear; sessions may be force-invalidated | Path A |

## When to use each

### Choose Path A unless you have a specific blocker

Path A is the right default for everyone with a personal @outlook.com /
@hotmail.com account.

### Path C is appropriate ONLY when:

1. **You have a corporate M365 account** and your IT department refuses
   to register an Azure app for you, AND you accept the ToS risk.
2. **You want to test the project** but don't want to spend 5 minutes
   doing the Azure portal setup, AND you have a sacrificial MSA account
   you're willing to lose.
3. **You're researching** anti-bot heuristics / browser automation / DOM
   scraping techniques, AND you read and understood `playwright/LEGAL.md`.

If none of these apply, **use Path A**.

## Why include Path C at all?

Because the alternative — saying "use only Path A and register an Azure
app" — excludes users in restricted environments (corporate IT, locked
tenants, etc.) who might benefit from a scraper for research or
sacrificial workflows.

By including Path C in the same repo with prominent legal disclaimers,
we:

1. **Make the ToS risk visible** rather than hidden in third-party tools.
2. **Allow community improvements** to a path that users genuinely need
   for research / fallback / sacrificial workflows.
3. **Maintain legal separation** between the maintained path and the
   community-contributed path.

This is the same reason `yt-dlp` ships 200+ extractors for sites that
explicitly prohibit scraping in their ToS — because users need them,
and shipping them with prominent disclaimers is more responsible than
pretending the need doesn't exist.

If you're reading this and disagree with Path C existing, file an issue
against `LEGAL-NOTICE.md` with your reasoning. Maintainers will
consider deprecation but, per precedent, will likely keep it with
stronger warnings.

## What's NOT in either path

Both paths are scoped **per-user, per-account**, single-process. They
do NOT support:

- Multi-tenant administration (use Microsoft 365 admin center for that)
- Shared mailbox delegation (use Outlook desktop client)
- Public folder access (Path A could in theory; not implemented; Path C
  cannot)
- Compliance / eDiscovery search (admin-only)
- Cross-account automation as a service (neither path is designed to
  be exposed as a network service)

If you need any of these, this project is not the right tool.

## Maintenance status (read this before filing issues)

| Status | Path A | Path C |
|---|---|---|
| Maintained | yes — issues are triaged within ~7d | no — issues may sit indefinitely |
| Releases | follow semver, ~1 per quarter | sporadic, when community drives them |
| Security patches | prioritized immediately | best-effort |
| Backports | last 2 minor versions | last 1 release |
| Deprecation policy | 24mo deprecation notice (mirrors MS Graph) | no formal policy — may stop working without notice |
