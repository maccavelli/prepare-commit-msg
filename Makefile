MOD_VERSION := 1.26.6
BINARY_NAME=prepare-commit-msg
DIST_DIR=dist
GIT_VERSION=$(shell git describe --tags --always --dirty 2>/dev/null)
VERSION?=$(GIT_VERSION)
TOOLS_BIN       := $(CURDIR)/.tools/bin
GOLANGCI_LINT   ?= $(TOOLS_BIN)/golangci-lint
GOVULNCHECK     ?= $(TOOLS_BIN)/govulncheck
ACTIONLINT      ?= $(TOOLS_BIN)/actionlint
FLEET_LINT_CFG := .golangci.yml

.PHONY: all build clean test coverage test-coverage run install version build-all \
	linux linux-amd64 linux-arm64 darwin-arm64 darwin-amd64 windows-amd64 windows-arm64 \
	help tools fmt fmt-check mod-check vet lint vuln workflow-lint verify verify-staged \
	release-artifacts verify-release hooks-install hooks-uninstall hooks-test

all: help build-all

build: ## Compiles the Go application for the local OS/Arch
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=0 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) .

build-all: linux-amd64 linux-arm64 darwin-arm64 darwin-amd64 windows-amd64 windows-arm64 ## Compiles for all 6 target platforms (Linux, macOS, Windows on AMD64 and ARM64)

linux: linux-amd64 ## Alias for linux-amd64

linux-amd64: ## Compiles for Linux x86_64 (AMD64)
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .

linux-arm64: ## Compiles for Linux ARM64 (aarch64)
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .

darwin-arm64: ## Compiles for macOS ARM64 (Apple Silicon)
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .

darwin-amd64: ## Compiles for macOS x86_64 (Intel)
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .

windows-amd64: ## Compiles for Windows x86_64 (AMD64)
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .

windows-arm64: ## Compiles for Windows ARM64
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe .

clean: ## Removes all build artifacts
	rm -rf $(DIST_DIR)

COVERAGE_MIN ?= 80.0

test: ## Runs all tests with verbose output and the race detector
	go test -race -v ./...

coverage: ## Enforces the aggregate statement coverage threshold
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | awk -v min=$(COVERAGE_MIN) '/total:/ { gsub(/%/,"",$$NF); printf "total coverage: %s%% (minimum %.1f%%)\n", $$NF, min; if (($$NF+0) < min) exit 1 }'

test-coverage: coverage ## Backward-compatible alias for coverage

tools: ## Installs repository-pinned development tools under .tools/bin
	./scripts/bootstrap-tools.sh

fmt: tools ## Formats Go source and imports
	$(GOLANGCI_LINT) fmt -c $(FLEET_LINT_CFG)

fmt-check: tools ## Checks Go formatting and imports without rewriting files
	@files="$$(git ls-files '*.go')"; \
		unformatted="$$(gofmt -l $$files)"; \
		if [ -n "$$unformatted" ]; then \
			echo "gofmt found unformatted files:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi
	$(GOLANGCI_LINT) fmt --diff -c $(FLEET_LINT_CFG)

mod-check: ## Verifies module tidiness and downloaded module checksums
	go mod tidy -diff
	go mod verify

vet: ## Runs go vet on the project
	go vet ./...

lint: tools ## Runs the repository-pinned golangci-lint configuration
	$(GOLANGCI_LINT) run -c $(FLEET_LINT_CFG) ./...

vuln: tools ## Reports reachable vulnerabilities using the current Go database
	$(GOVULNCHECK) ./...

workflow-lint: tools ## Checks shell scripts and GitHub Actions workflows
	ACTIONLINT=$(ACTIONLINT) ./scripts/verify-scripts.sh

verify: tools mod-check fmt-check lint vet test coverage vuln workflow-lint build-all ## Runs the complete local and CI quality contract

verify-staged: tools ## Checks the exact staged Go snapshot
	./scripts/go-precheck.sh

hooks-install: ## Installs composable repository-local Git hook wrappers
	./scripts/install-hooks.sh

hooks-uninstall: ## Restores the hooks-path state from before installation
	./scripts/uninstall-hooks.sh

hooks-test: ## Tests hook composition in isolated temporary repositories
	./scripts/test-hooks.sh

release-artifacts: clean build-all ## Builds and checksums the complete release asset set
	./scripts/verify-release.sh --write-checksums $(DIST_DIR)

verify-release: release-artifacts ## Builds and validates release assets (set VERSION=vX.Y.Z)
	./scripts/verify-release.sh "$(VERSION)" "$(DIST_DIR)"

run: build ## Builds and executes the local binary
	@BIN_NAME=$(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) ; \
	$$BIN_NAME

install: build ## Installs the local binary to ~/.global-git-hooks/ (Unix)
	@mkdir -p $(HOME)/.global-git-hooks
	@BIN_NAME=$(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) ; \
	cp $$BIN_NAME $(HOME)/.global-git-hooks/$(BINARY_NAME) ; \
	chmod +x $(HOME)/.global-git-hooks/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(HOME)/.global-git-hooks/"
	@echo "Ensure git core.hooksPath points at that directory (or symlink into .git/hooks)."

version: build ## Displays the version of the local binary
	@BIN_NAME=$(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) ; \
	$$BIN_NAME version

help: ## Displays this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
