# Contributing

Thanks for your interest in `outlook-mcp-suite`. This document explains
how to open issues, propose changes, and what we expect from
contributors.

> **⚠️ This repo is AGENT-AUTOMATED.** Read § **AGENT-AUTOMATED REPOSITORY**
> before opening a PR. Some rules are different from typical repos.
> Most importantly: **the AI agent operating this repo self-merges PRs
> that meet strict criteria**, without waiting for human approval per-PR.

## TL;DR — what kind of contribution for what path

| I want to... | Path to use | Where to open it |
|---|---|---|
| Report a bug in Path A | `graph/` | issue against this repo, label `path:A-graph` |
| Report a bug in Path C | `playwright/` | issue against this repo, label `path:C-playwright`, see `playwright/LEGAL.md` first |
| Request a feature for Path A | `graph/` | issue, label `path:A-graph feature` |
| Request a feature for Path C | `playwright/` | issue, label `path:C-playwright community` — understand first that this path is community-maintained |
| Propose a code change | depends on path | PR against `dev` branch, see `AGENT-AUTOMATED` below |
| Improve docs in root (`README`, `ARCHITECTURE`, etc) | this repo | PR against `dev` |
| File a security disclosure | this repo | security advisory, see `SECURITY.md` (Phase 5) |
| Send a DMCA / ToS complaint | this repo | email maintainer, see `LEGAL-NOTICE.md` |

## Branch layout

```
main     ← stable releases only; never direct commits
dev      ← active development; PRs merge here, then squash-merged to main for release
chore/*  ← maintenance (docs, ci, deps)
feat/*   ← new features (per-path prefix: feat/graph/* or feat/playwright/*)
fix/*    ← bug fixes (per-path prefix)
```

PRs target `dev`, never `main`. Releases are cut by maintainers
using `make release` (Phase 5 target).

## Branch protection

The repo will eventually have GitHub branch protection rules on
`main` (and possibly `dev`). When active, the rules will be:

- **Required status checks**: CI green is mandatory before merge
- **Required reviews**: 1 approving review from a maintainer
- **No force-push**: rebases only via `gh pr update-branch` or
  re-opening the PR
- **No direct pushes**: all changes via PR
- **Dismiss stale approvals on push**: re-review required after
  any force-update

### Admin bypass

The repo owner (`@louzt` at time of writing) holds admin rights and
may bypass branch protection via the `--admin` flag on `gh pr merge`,
or by pushing directly to a protected branch. This is a **privilege,
not a workflow default** — admins are expected to use the same PR
flow as everyone else, with admin bypass reserved for genuine
emergencies (security fixes, recovery from a broken main, etc.).

Other maintainers (when added) do not hold admin bypass unless
explicitly granted. AI agents operating on behalf of the repo owner
may use admin bypass **only** when the owner has issued a literal
`approve PR <#>: <reason>` phrase in the same turn.

### What this means for contributors

- **Non-admin contributors** cannot self-approve their own PRs. They
  must wait for a maintainer review.
- **AI agents** (e.g. Claude Code opened on a contributor's behalf)
  must request a review from a maintainer; the agent may NOT use
  `--admin` even if the contributor's token has admin rights,
  because that would be circumvention of the trust model.
- **Cosmetic-only PRs** are rejected at review time regardless of
  branch protection status.

## AGENT-AUTOMATED REPOSITORY

> **Read this section before opening a PR. If you are an AI agent
> operating this repo, read AGENTS.md for the agent-side equivalent.**

This repo delegates **per-PR human approval** to the AI agent under
strict conditions. The agent may auto-merge a PR when ALL of the
following are true:

- [ ] **Single concern atomic** (rollbackable in isolation; one bug, one feature, one refactor per PR)
- [ ] **Diff size**: ≤ 1500 LOC median, ≤ 3000 LOC absolute
- [ ] **CI green** before `gh pr create --ready` (no `--no-verify` commits)
- [ ] **Commit message cites semantic rationale** (manual, behavior observed, test) — NOT "convention" or "matches v1 pattern"
- [ ] **No secrets / no absolute paths / no host-specific values** in diff (run `lzt-secret-scrubber` pre-push)
- [ ] **PR template filled**: §Description, §Redundancy check, §Compatibility, §Test plan, §Security implications

### PRs that ALWAYS require explicit human approval

The agent may NOT auto-merge any of the following — they require a
human (the repo owner, in this repo's case) to act:

- Any PR where even one of the 6 criteria above fails
- PRs targeting `main` (only `dev`)
- PRs with `--no-verify` commits
- PRs with `WIP` or "draft for visibility" body
- Any PR touching `.claude/hooks/` (security-critical surface)
- Any PR changing `LEGAL-NOTICE.md` or this `AGENT-AUTOMATED`
  section (operator-controlled boundary)
- Any PR opened by a contributor who is not the repo owner — those
  require a maintainer review regardless of the 6 criteria

### Per-invocation bypass (repo owner only)

The repo owner may override a single PR with the literal phrase
**`approve PR <#>: <reason>`** in their next message. This:

- Applies to that PR only
- Does NOT extend to future PRs
- Does NOT extend to any other repo
- Does NOT extend to PRs opened by other contributors (those still
  require a maintainer review)
- Allows the agent to use the `--admin` flag on `gh pr merge` if
  branch protection is active and would otherwise block

The agent records this phrase in the commit message audit trail and
proceeds with the merge. Other contributors do not have this bypass
— they must follow the standard PR + review flow.

### Why this rule exists

This rule balances:

- Rapid iteration on a brand-new repo
- Review-quality preservation (small atomic PRs are reviewable in
  minutes; mega-PRs collapse review quality)
- Audit traceability — every merged PR has a git commit hash, CI run,
  and commit message containing semantic rationale, which is more
  auditable than a human-vetted PR with "LGTM"
- Anti-farming safeguards: cosmetic-only PRs ("para inflar contador")
  and `gh pr merge --auto` without CI green are the two patterns
  that turn fast iteration into account-level risk

The 6 criteria above are the minimum viable constraints to keep
auto-merge from drifting into farming. If you think a criterion is
wrong, open an issue with `discussion:` prefix and propose a
replacement. Don't just bypass.

### Cadence — unrestricted (per-PR value, not per-PR count)

This repo is operated with an explicit **high-cadence PR slicing**
model: 5+ PRs in a 24h window is fine, multiple PRs touching the
same concern in 24h is fine, when each PR is genuinely atomic, has
tests, and passes CI. The constraint that does NOT loosen is the
**per-PR value** requirement: cosmetic-only PRs (renaming imports,
adding blank lines, reformatting whitespace for its own sake) are
rejected regardless of how small the diff is.

Likewise, `gh pr merge --auto` requires CI green. `--auto` is the
green-CI fast-path; bypassing the CI wait is exactly the risk being
gated.

## Code style

- Go: `gofmt` + `go vet` + `staticcheck`. The agent will auto-format
  PRs that fail `gofmt`; pre-existing style in a file is honored.
- Shell: `shellcheck -S warning`. Tests use `bats` (Phase 0.5).
- Markdown: `markdownlint` (lints run but do not block).

## Communication

- **Bug reports**: be specific about path (A or C), OS, Go version,
  and reproduction. Include the actual JSON envelope from the failing
  tool call (redact credentials first).
- **Feature requests**: explain the use case, not just the solution.
  Especially for Path C: explain why Path A doesn't fit.
- **Security**: see `SECURITY.md` (Phase 5) — do NOT open public
  issues for undisclosed vulnerabilities.

## What maintainers will REJECT

- PRs that game the auto-merge criteria (e.g. splitting one feature
  across 5 PRs to keep each under 1500 LOC)
- PRs that bypass a hook (`/dev/null`-piping, env-var override of
  `LZT_*` policy flags, etc.)
- PRs that introduce a new top-level dependency without prior
  discussion in an issue
- PRs that change LEGAL-NOTICE.md or this section without maintainer
  approval
- PRs to Path C that don't acknowledge the ToS implications in the
  description

## Commit signing

All commits must be signed (GPG or SSH). The repo enforces this via
branch protection on `main` (Phase 5). For `dev` it's encouraged
but not enforced.

## License

By contributing, you agree that your contributions are licensed under
the Apache License 2.0 — same as this repo's LICENSE file. You
retain copyright on your original work; the contribution is a
non-exclusive grant.

## Code of conduct

Be excellent to each other. This is a small OSS project; conflicts
are between maintainers and issue authors, not between contributors.
Harrassment, doxxing, or threats result in permanent ban without
notice.
