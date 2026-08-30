#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
TOOLS_BIN="$REPO_ROOT/.tools/bin"

GOLANGCI_LINT_VERSION="v2.13.1"
GOVULNCHECK_VERSION="v1.7.0"
ACTIONLINT_VERSION="v1.7.12"
GO_VERSION="go1.26.6"

mkdir -p "$TOOLS_BIN"

if [ "$(go env GOVERSION)" != "$GO_VERSION" ]; then
	echo "expected $GO_VERSION, got $(go env GOVERSION)" >&2
	exit 1
fi

install_tool() {
	local name="$1"
	local module="$2"
	local version="$3"
	local expected="$4"
	local binary="$TOOLS_BIN/$name"
	local output=""

	if [ -x "$binary" ]; then
		case "$name" in
		golangci-lint) output="$($binary version 2>&1)" ;;
		govulncheck) output="$($binary -version 2>&1)" ;;
		actionlint) output="$($binary -version 2>&1)" ;;
		esac
	fi

	if [[ "$output" != *"$expected"* ]]; then
		echo "installing $module@$version"
		GOBIN="$TOOLS_BIN" go install "$module@$version"
	fi
}

install_tool \
	"golangci-lint" \
	"github.com/golangci/golangci-lint/v2/cmd/golangci-lint" \
	"$GOLANGCI_LINT_VERSION" \
	"2.13.1"
install_tool \
	"govulncheck" \
	"golang.org/x/vuln/cmd/govulncheck" \
	"$GOVULNCHECK_VERSION" \
	"govulncheck@$GOVULNCHECK_VERSION"
install_tool \
	"actionlint" \
	"github.com/rhysd/actionlint/cmd/actionlint" \
	"$ACTIONLINT_VERSION" \
	"$ACTIONLINT_VERSION"
