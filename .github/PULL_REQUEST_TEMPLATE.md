## Summary

<!-- One-paragraph description of the change. -->

## Scope

<!-- What this PR does (concern-by-concern). -->

- [ ] Concern 1
- [ ] Concern 2

## Out of scope

<!-- Explicit non-goals so reviewers don't expect them. -->

## Test evidence

```text
$ go -C ./graph/outlook-mcp-go test -race -v ./...
# paste real output
```

## Checklist

- [ ] Tests pass locally (`go test -race ./...`)
- [ ] `go vet ./...` clean
- [ ] No secrets, tokens, or internal paths committed
- [ ] New public API documented in code comments
- [ ] Conventional commits used (`feat:`, `fix:`, `chore:`, etc.)

## Related

<!-- Link to issues, design docs, or other PRs. -->
