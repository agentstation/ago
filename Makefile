# ago Makefile
MAKEFLAGS += --no-print-directory

BINARY_NAME = ago
MAIN_PATH   = ./cmd/ago
GO          = go

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
LDFLAGS  = -ldflags "-s -w -X github.com/agentstation/ago.Version=$(VERSION)"

GOLANGCI_LINT_VERSION = v2.12.2
GORELEASER_VERSION    = 2.17.1
TECHNICAL_WRITING    ?= $(HOME)/.agents/skills/technical-writing/scripts/technical-writing

# testdata holds intentionally unparseable Go. Point every tool that walks
# the tree at real source instead.
GOFILES = $(shell find . -name '*.go' -not -path './testdata/*')

.PHONY: all
all: help

##@ General

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[1mUsage:\033[0m\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } \
		/^###/ { printf "  \033[90m%s\033[0m\n", substr($$0, 4) }' $(MAKEFILE_LIST)

##@ Build

.PHONY: build
build: ## Build the ago binary
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: install
install: ## Install ago into GOBIN
	$(GO) install $(LDFLAGS) $(MAIN_PATH)

.PHONY: clean
clean: ## Remove build and test output
	@rm -f $(BINARY_NAME) coverage.txt ago.sarif
	@rm -rf dist/
	$(GO) clean -testcache

##@ Test

.PHONY: test
test: ## Run tests with the race detector
	$(GO) test -race ./...

.PHONY: test-short
test-short: ## Run tests without the race detector
	$(GO) test ./...

.PHONY: cover
cover: ## Run tests and write a coverage profile
	$(GO) test -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GO) tool cover -func=coverage.txt | tail -1

##@ Lint

.PHONY: fmt
fmt: ## Format all source (excluding testdata)
	@gofmt -w $(GOFILES)

.PHONY: fmt-check
fmt-check: ## Fail if gofmt lists unformatted source
	@out=$$(gofmt -l $(GOFILES)); \
	if [ -n "$$out" ]; then echo "unformatted files:"; echo "$$out"; exit 1; fi

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: fmt-check vet ## Run gofmt, go vet, and golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not installed; run: make tools"; exit 1; \
	fi
	@# A different version on PATH finds a different set of issues, so a clean
	@# run here can still fail CI. Warn rather than stop. The run is still useful.
	@have=$$(golangci-lint version 2>/dev/null | sed -n 's/.*has version \([^ ]*\).*/\1/p'); \
	want=$(GOLANGCI_LINT_VERSION:v%=%); \
	if [ "$$have" != "$$want" ]; then \
		echo "warning: golangci-lint $$have on PATH, CI runs $$want; run: make tools"; \
	fi
	golangci-lint run

.PHONY: prose
prose: ## Lint developer-facing prose in strict mode
	@if [ ! -x "$(TECHNICAL_WRITING)" ]; then \
		echo "technical-writing helper not found at $(TECHNICAL_WRITING)"; \
		echo "set TECHNICAL_WRITING to the helper path"; exit 1; \
	fi
	$(TECHNICAL_WRITING) glossary check
	$(TECHNICAL_WRITING) lint . --mode strict --format text

.PHONY: tidy
tidy: ## Tidy and verify go.mod
	$(GO) mod tidy
	$(GO) mod verify

##@ Dogfood

.PHONY: ago
ago: build ## Run ago against its own source
	./$(BINARY_NAME) ./...

.PHONY: check
check: fmt-check vet test ago ## Run everything CI runs

##@ Release

.PHONY: snapshot
snapshot: ## Build a local release snapshot with goreleaser
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not installed; run: make tools"; exit 1; }
	goreleaser release --snapshot --clean

.PHONY: release-check
release-check: ## Validate the goreleaser configuration
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not installed; run: make tools"; exit 1; }
	goreleaser check

##@ Tooling

.PHONY: tools
tools: ## Install development tools
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO) install github.com/goreleaser/goreleaser/v2@v$(GORELEASER_VERSION)
