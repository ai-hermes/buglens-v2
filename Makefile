SHELL := /bin/bash

APP := buglens
MAIN_PKG := ./cmd/buglens
BIN_DIR := bin
OUT := $(BIN_DIR)/$(APP)

GO ?= go
GOTOOLCHAIN ?= auto
CGO_ENABLED ?= 0

.PHONY: help deps tidy fmt test test-race build run serve call release-snapshot clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-18s %s\n", $$1, $$2}'

deps: ## Download module dependencies
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) mod download

tidy: ## Tidy go modules
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) mod tidy

fmt: ## Format all Go files
	GOTOOLCHAIN=$(GOTOOLCHAIN) $(GO) fmt ./...

test: ## Run unit tests
	GOTOOLCHAIN=$(GOTOOLCHAIN) CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./...

test-race: ## Run tests with race detector
	GOTOOLCHAIN=$(GOTOOLCHAIN) CGO_ENABLED=1 $(GO) test -race ./...

build: ## Build CLI binary to ./bin/buglens
	mkdir -p $(BIN_DIR)
	GOTOOLCHAIN=$(GOTOOLCHAIN) CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(OUT) $(MAIN_PKG)

run: ## Run CLI directly (example: make run ARGS='version')
	GOTOOLCHAIN=$(GOTOOLCHAIN) CGO_ENABLED=$(CGO_ENABLED) $(GO) run $(MAIN_PKG) $(ARGS)

serve: ## Run MCP server (example: make serve ARGS='--transport stdio')
	GOTOOLCHAIN=$(GOTOOLCHAIN) CGO_ENABLED=$(CGO_ENABLED) $(GO) run $(MAIN_PKG) mcp serve $(ARGS)

call: ## Call a tool locally (example: make call TOOL=gitlab_list_projects ARGS_JSON='{\"search\":\"demo\"}')
	GOTOOLCHAIN=$(GOTOOLCHAIN) CGO_ENABLED=$(CGO_ENABLED) $(GO) run $(MAIN_PKG) mcp call --tool "$(TOOL)" --args-json '$(ARGS_JSON)'

release-snapshot: ## Build release artifacts via goreleaser snapshot
	GOTOOLCHAIN=$(GOTOOLCHAIN) goreleaser build --snapshot --clean

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) dist
