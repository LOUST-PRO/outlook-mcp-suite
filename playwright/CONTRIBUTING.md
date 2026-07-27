# Contributing to `playwright/` (Path C)

> **Read `playwright/LEGAL.md` first.** Every PR to this subtree
> must acknowledge the ToS implications.

Path C is community-maintained. This document covers the rules
specific to this path; the rules in [`/CONTRIBUTING.md`](../CONTRIBUTING.md)
apply on top.

## What maintainers will and won't review

| Kind of contribution | Reviewable? |
|---|---|
| Bug fix that adapts to a recent MS UI change | yes — high priority |
| Refactor that doesn't change behavior | yes |
| Performance improvement | yes |
| Anti-bot-detection circumvention | **rejected** |
| Captcha-solver integration | **rejected** |
| Session-cookie theft or session-fixation | **rejected** |
| Anything that bypasses MS account security | **rejected** |
| Multi-account parallel scraping | rejected (raises anti-bot risk for everyone) |

## Required in every PR description

Every PR description must include a section `## ToS acknowledgment`
that states, in the contributor's own words, that they understand
Path C violates the Microsoft Services Agreement §3.b and that they
are contributing to a community-maintained path that the maintainers
do not encourage for production use.

A template line:

```markdown
## ToS acknowledgment
I have read [`playwright/LEGAL.md`](./LEGAL.md) and understand that
Path C uses browser automation against `outlook.live.com` /
`outlook.office.com` in violation of the Microsoft Services
Agreement §3.b. I am contributing to a community-maintained path
that I will not recommend to others for production use, and I will
not include anti-bot-circumvention code in this PR.
```

Without this section, the PR will be closed without review.

## Anti-bot policy

Path C must:

- Respect a default polling interval of ≥60 seconds
- Not include code to bypass CAPTCHAs (the maintainers reserve
  the right to ban any contribution that does)
- Not include code to forge or steal cookies from other browsers
  or sessions
- Not include code to manipulate Microsoft account security
  (password reset flows, MFA bypass, etc.)
- Not include code to scrape content the user did not explicitly
  request

If your PR would do any of the above, **don't open it**.

## Cookie / session storage rules

The encrypted cookie file (used to persist login between MCP calls)
must:

- Use AES-GCM with a key derived from `OUTLOOK_KEY` env var (set
  by the user via Bitwarden or similar secret manager)
- Have file permissions `0600`
- Not be committed to git (already covered by `.gitignore`)
- Be deletable by a single `lzt-outlook auth logout` command

If your PR proposes a different storage mechanism, open an issue
first to discuss; maintainers may reject changes that weaken the
storage model.

## Logging rules

- Do not log email content (subject, body, recipients) at INFO or
  higher. Debug-level is OK for development but must be opt-in via
  `LZT_OUTLOOK_DEBUG=1`.
- Do not log session cookies, even at DEBUG.
- Do not log full URLs (they may contain auth tokens in
  query strings); log only path + first 80 chars of query if needed.
- Log apply operations at INFO level so users can audit what the
  MCP did.

## Testing

- Unit tests must use mock HTML, not live `outlook.live.com`
- Live smoke tests are forbidden in CI (they would violate ToS).
  See `/COMPARISON.md` § Maintenance status.

## Maintenance

Path C is **not** a maintained path. PRs may sit for weeks without
review. If you need a fix urgently, fork the repo and patch your
fork; open a PR but don't expect quick turnaround.

## License

Apache-2.0, same as the repo root. See [`/LICENSE`](../LICENSE).