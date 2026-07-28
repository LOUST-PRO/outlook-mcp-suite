# Makefile — graph/outlook-mcp-go (Path A: Microsoft Graph)
#
# Stage 1 of Fase 1: scaffold + auth + config + list_messages proof.
# See ../Makefile (repo root) for cross-module targets.

GO            ?= go
GOFLAGS       ?= -trimpath
BINARY        := outlook-mcp-go
PKG           := ./cmd/outlook-mcp-go
INSTALL_DIR   ?= $(HOME)/.local/bin
SHELLCHECK    ?= shellcheck

.PHONY: all build test vet tidy run-auth serve clean install uninstall check help

all: vet build ## Default: vet + build

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Compile the MCP binary
	$(GO) build $(GOFLAGS) -o $(BINARY) $(PKG)

vet: ## Run go vet
	$(GO) vet ./...

test: ## Run unit tests
	$(GO) test -race ./...

tidy: ## Run go mod tidy
	$(GO) mod tidy

run-auth: ## Run device-code flow (usage: make run-auth ACCOUNT=personal)
ifndef ACCOUNT
	$(error ACCOUNT is required, e.g. make run-auth ACCOUNT=personal)
endif
	$(GO) run $(PKG) auth $(ACCOUNT)

serve: build ## Run the MCP server (stdio transport)
	./$(BINARY) serve

install: build ## Install binary to ~/.local/bin/
	install -m 0755 $(BINARY) $(INSTALL_DIR)/$(BINARY)

uninstall: ## Remove installed binary
	rm -f $(INSTALL_DIR)/$(BINARY)

clean: ## Remove build artifacts
	rm -f $(BINARY)

check: vet test ## Pre-commit gate: vet + test
