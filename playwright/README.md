# `playwright/` — Path C: web.outlook.com scraping (COMMUNITY)

> **⚠️ Read `playwright/LEGAL.md` before contributing or using this
> path. Microsoft ToS explicitly prohibits automated access except
> via official APIs. This subtree exists for users with genuine
> technical blockers to Path A, NOT as a blessed alternative.**

## What lives here (planned, not yet implemented)

```
playwright/outlook-scraper-go/
├── cmd/outlook-scraper-go/main.go   # MCP server entry point (Phase 3+)
├── internal/
│   ├── login/                       # Browser session bootstrap (interactive)
│   ├── scraper/                     # DOM-based data extraction
│   └── types/                       # Shared scraper type adapters
├── tests/                           # Unit tests (mock HTML only)
└── smoke/                           # NO live smoke; see LEGAL.md
```

## Why this path

The historical path-of-last-resort for no-Azure setups used to be
IMAP with Basic Auth. Microsoft removed Basic Auth for IMAP in
**October 2023**, eliminating that path. The next option for users
who genuinely cannot or will not register an Azure app is browser
automation against `outlook.live.com` / `outlook.office.com`.

This subtree exists because some users have that blocker (corporate
IT refuses, locked tenants, etc.) and benefit from a fragile but
functional fallback.

It is **NOT** the recommended path. See `COMPARISON.md` and
`LEGAL-NOTICE.md`.

## Status

This subtree is **structurally scaffolded only** as of Phase 0. No
Go code yet. The legal disclaimers (`LEGAL.md`, `CONTRIBUTING.md`,
and `playwright/README.md`) are intentionally written first so any
contributor or user encounters them before any code.

## Maintenance status

| Concern | Status |
|---|---|
| Maintained | No — community-driven |
| Issue triage | Best-effort, may be slow |
| Releases | Sporadic, when community drives |
| Security patches | Best-effort |
| Deprecation policy | No formal policy — may stop working without notice |

If your only option is Path C, please consider contributing back
fixes when the web UI changes. Microsoft ships UI changes without
notice; this path survives only as long as someone is willing to
chase the UI.

## License

Apache-2.0, same as the repo root. See [`/LICENSE`](../LICENSE).

## Contributing to this path

See [`playwright/CONTRIBUTING.md`](./CONTRIBUTING.md). Every PR
must acknowledge the ToS implications in the description.