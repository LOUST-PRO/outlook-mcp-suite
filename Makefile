# outlook-mcp-suite — Makefile
#
# Targets are organized per Phase. Phase 0 = scaffold only.
# Phase 0.5 = shell hooks.
# Phase 1   = Path A (graph) read-only Go wrapper.
# Phase 2   = Path A mutates (propose/apply) + audit ledger.
# Phase 3   = Path C (playwright) structure + LEGAL.
# Phase 4   = Unified CLI (lzt-outlook dispatcher).
# Phase 5   = Onboarding docs.

SHELL := /usr/bin/env bash
.SHELLFLAGS := -eu -o pipefail -c

REPO_ROOT := $(shell pwd)

# --- helpers ----------------------------------------------------------------

.PHONY: help
help: ## show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} \
	/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## remove build artifacts (gitignored)
	rm -rf bin/ dist/ coverage.* *.coverprofile playwright/.cache/

.PHONY: lint
lint: ## run all linters (placeholder until implementation lands)
	@echo "lint: TODO Phase 1"

.PHONY: test
test: ## run all tests (placeholder until implementation lands)
	@echo "test: TODO Phase 1"

.PHONY: fmt
fmt: ## format Go and shell sources (placeholder)
	@echo "fmt: TODO Phase 1"

.PHONY: vet
vet: ## go vet the Go wrappers (placeholder)
	@echo "vet: TODO Phase 1"

# --- phase 0: scaffold (already done by manual write) -------------------------

.PHONY: phase0-check
phase0-check: ## verify Phase 0 files exist
	@for f in LICENSE LEGAL-NOTICE.md README.md ARCHITECTURE.md COMPARISON.md CONTRIBUTING.md AGENTS.md .gitignore Makefile \
	         .claude/skills/outlook-assistant/SKILL.md \
	         .claude/hooks/README.md \
	         graph/README.md \
	         playwright/README.md playwright/LEGAL.md playwright/CONTRIBUTING.md \
	         cli/lzt-outlook/README.md \
	         docs/README.md; do \
	  test -f "$$f" || { echo "missing: $$f" >&2; exit 1; }; \
	done
	@echo "phase0-check: all required files present"

# --- phase 0.5: shell hooks --------------------------------------------------

HOOKS_DIR := .claude/hooks
HOOKS_SRC := $(shell ls $(HOOKS_DIR)/pre-tool-outlook-*.sh 2>/dev/null | grep -v _lib)

.PHONY: phase0.5-build
phase0.5-build: ## install the 4 hooks into ~/.claude/hooks (dry-run by default)
	@echo "phase0.5-build: TODO Phase 0.5 - implement after operator approval"

.PHONY: phase0.5-test
phase0.5-test: ## run hook tests under tests/hooks/
	@echo "phase0.5-test: TODO Phase 0.5"

# --- phase 1: path A read-only wrapper ---------------------------------------

.PHONY: phase1-build
phase1-build: ## build graph/outlook-mcp-go
	cd graph/outlook-mcp-go && go build ./...

.PHONY: phase1-test
phase1-test: ## run graph/outlook-mcp-go tests
	cd graph/outlook-mcp-go && go test -race -count=1 ./...

.PHONY: phase1-smoke
phase1-smoke: ## smoke graph/outlook-mcp-go against sacrificial account
	@echo "phase1-smoke: requires OUTLOOK_CLIENT_ID env var set; TODO Phase 1"

.PHONY: phase1-release
phase1-release: ## cut a release tag for Path A
	@echo "phase1-release: TODO Phase 1"

# --- phase 2: path A mutates + ledger ---------------------------------------

.PHONY: phase2-build
phase2-build: ## build with apply tools
	$(MAKE) phase1-build

.PHONY: phase2-test
phase2-test: ## test apply tools + ledger
	$(MAKE) phase1-test
	@echo "phase2-test: TODO Phase 2"

.PHONY: phase2-smoke
phase2-smoke: ## smoke apply path with sacrificial account
	@echo "phase2-smoke: TODO Phase 2"

# --- phase 3: path C scaffolding --------------------------------------------

.PHONY: phase3-build
phase3-build: ## build playwright/outlook-scraper-go (when implemented by community)
	cd playwright/outlook-scraper-go && go build ./...

.PHONY: phase3-legal-check
phase3-legal-check: ## verify playwright/LEGAL.md is present and prominent
	@test -f playwright/LEGAL.md && \
	  grep -q "ToS" playwright/LEGAL.md && \
	  echo "phase3-legal-check: LEGAL.md present and references ToS"

# --- phase 4: unified CLI ----------------------------------------------------

.PHONY: phase4-build
phase4-build: ## build cli/lzt-outlook
	cd cli/lzt-outlook && go build ./...

.PHONY: phase4-test
phase4-test: ## run unified CLI tests
	cd cli/lzt-outlook && go test -race -count=1 ./...

# --- phase 5: onboarding docs -----------------------------------------------

.PHONY: phase5-build
phase5-build: ## verify all 4 docs exist
	@for f in docs/onboarding-real-account.md docs/shadow-mode.md docs/ledger-format.md docs/performance-cost.md; do \
	  test -f "$$f" || { echo "missing: $$f" >&2; exit 1; }; \
	done
	@echo "phase5-build: all docs present"

# --- cross-phase targets ----------------------------------------------------

.PHONY: status
status: ## show current phase status
	@echo "Phase 0 scaffold:    $(shell test -f LICENSE && echo OK || echo PENDING)"
	@echo "Phase 0.5 hooks:    $(shell ls .claude/hooks/pre-tool-outlook-*.sh 2>/dev/null | wc -l) of 4 written"
	@echo "Phase 1 graph A:    $(shell test -f graph/outlook-mcp-go/main.go && echo OK || echo PENDING)"
	@echo "Phase 2 mutates:    $(shell test -f graph/outlook-mcp-go/internal/apply/apply.go && echo OK || echo PENDING)"
	@echo "Phase 3 path C:     $(shell test -f playwright/outlook-scraper-go/main.go && echo OK || echo PENDING)"
	@echo "Phase 4 CLI:        $(shell test -f cli/lzt-outlook/main.go && echo OK || echo PENDING)"
	@echo "Phase 5 docs:       $(shell ls docs/*.md 2>/dev/null | wc -l) of 4 written"

.PHONY: install-hooks-dry-run
install-hooks-dry-run: ## show what install-hooks would do (without doing it)
	@echo "Would symlink the following to ~/.claude/hooks/:"
	@for h in $(HOOKS_SRC); do \
	  echo "  $$h -> ~/.claude/hooks/$$(basename $$h)"; \
	done

.PHONY: ci
ci: phase0-check lint test phase5-build phase3-legal-check ## continuous-integration gate
