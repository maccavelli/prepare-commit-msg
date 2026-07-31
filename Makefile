MOD_VERSION := 1.26.5
BINARY_NAME=prepare-commit-msg
DIST_DIR=dist
GIT_VERSION=$(shell git describe --tags --always --dirty 2>/dev/null)
VERSION?=$(GIT_VERSION)
# Prefer the user's Go toolchain install (go install ...), then PATH.
GOPATH_BIN     := $(shell go env GOPATH)/bin
GOLANGCI_LINT  ?= $(GOPATH_BIN)/golangci-lint
FLEET_LINT_CFG := .golangci.yml

.PHONY: all build clean test run install version build-all \
	linux linux-arm64 darwin-arm64 darwin-amd64 windows-amd64 windows-arm64 \
	help fmt vet lint

all: help build-all

build: ## Compiles the Go application for the local OS/Arch
	@mkdir -p $(DIST_DIR)
	@CGO_ENABLED=0 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(if $(filter windows,$(shell go env GOOS)),.exe,) .

build-all: linux linux-arm64 darwin-arm64 darwin-amd64 windows-amd64 windows-arm64 ## Compiles for multiple platforms

linux: ## Compiles for Linux AMD64
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .

linux-arm64: ## Compiles for Linux ARM64
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -tags netgo -ldflags "-extldflags '-static' -s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .

darwin-arm64: ## Compiles for macOS ARM64 (Apple Silicon)
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .

darwin-amd64: ## Compiles for macOS AMD64 (Intel)
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.Version=$(VERSION)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .

windows-amd64: ## Compiles for Windows AMD64
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

test-coverage: ## Runs tests and fails if overall statement coverage is below COVERAGE_MIN (default 80.0)
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | awk -v min=$(COVERAGE_MIN) '/total:/ { gsub(/%/,"",$$NF); printf "total coverage: %s%% (minimum %.1f%%)\n", $$NF, min; if (($$NF+0) < min) exit 1 }'

fmt: ## Formats all Go source files
	go fmt ./...

vet: ## Runs go vet on the project
	go vet ./...

lint: ## Runs golangci-lint from $(go env GOPATH)/bin (override with GOLANGCI_LINT=)
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint not found at $(GOLANGCI_LINT)"; \
		echo "Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run -c $(FLEET_LINT_CFG) ./...

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
