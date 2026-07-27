# `graph/` — Path A: Microsoft Graph (PRIMARY)

This subtree contains the **maintained implementation path** for
`outlook-mcp-suite`, using Microsoft Graph API with OAuth 2.0
device-code flow against a personal app the user registers in their
own Azure portal.

## What lives here

```
graph/outlook-mcp-go/
├── cmd/outlook-mcp-go/main.go      # MCP server entry point (Phase 1)
├── internal/
│   ├── auth/                       # OAuth 2.0 device-code + token store
│   ├── mail/                       # Mail folder/list/read/search/conversation
│   ├── calendar/                   # Calendar list/get/free-busy
│   ├── attachments/                # List + extract attachments
│   ├── propose/                    # Shadow-mode previews for mutates
│   ├── apply/                      # Real mutates (Phase 2)
│   └── types/                      # Shared Graph type adapters
├── docs/                           # rate-limits, Graph deprecations, examples
├── tests/                          # Unit tests
└── smoke/                          # End-to-end against sacrificial account
```

## Why this path

See [`/COMPARISON.md`](../COMPARISON.md) for the full decision table.
TL;DR: Graph is the official, supported, documented API. Path C
(`playwright/`) is a community-maintained fallback for users who
genuinely cannot use Graph.

## Status

This subtree is **structurally scaffolded only** as of Phase 0. The
actual Go code lands in Phase 1 (~4 hours of work for the read-only
wrapper, then Phase 2 for mutates).

To verify the path structure is intact:

```bash
make phase0-check
```

## Onboarding (full instructions)

Phase 5 will publish `docs/onboarding-real-account.md`. Until then,
see [`/README.md`](../README.md) § Quick start for the 5-minute
Azure portal walkthrough.

## License

Apache-2.0, same as the repo root. See [`/LICENSE`](../LICENSE).