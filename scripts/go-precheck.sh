#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
GOLANGCI_LINT="$REPO_ROOT/.tools/bin/golangci-lint"
TEMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/prepare-commit-msg-staged.XXXXXX")"
trap 'rm -rf "$TEMP_ROOT"' EXIT

if [ ! -x "$GOLANGCI_LINT" ]; then
	"$REPO_ROOT/scripts/bootstrap-tools.sh"
fi

GO_FILES=()
if [ "$#" -gt 0 ]; then
	for candidate in "$@"; do
		case "$candidate" in
		*.go) GO_FILES+=("$candidate") ;;
		esac
	done
else
	while IFS= read -r candidate; do
		[ -n "$candidate" ] && GO_FILES+=("$candidate")
	done < <(git -C "$REPO_ROOT" diff --cached --name-only --diff-filter=ACM -- '*.go')
fi

if [ "${#GO_FILES[@]}" -eq 0 ]; then
	exit 0
fi

git -C "$REPO_ROOT" checkout-index --all --prefix="$TEMP_ROOT/"

STAGED_FILES=()
for candidate in "${GO_FILES[@]}"; do
	if [ -f "$TEMP_ROOT/$candidate" ]; then
		STAGED_FILES+=("$TEMP_ROOT/$candidate")
	fi
done

if [ "${#STAGED_FILES[@]}" -eq 0 ]; then
	exit 0
fi

UNFORMATTED="$(gofmt -l "${STAGED_FILES[@]}")"
if [ -n "$UNFORMATTED" ]; then
	echo "gofmt found unformatted staged files:" >&2
	printf '%s\n' "$UNFORMATTED" >&2
	exit 1
fi

(
	cd "$TEMP_ROOT"
	"$GOLANGCI_LINT" fmt --diff -c .golangci.yml
	"$GOLANGCI_LINT" run -c .golangci.yml ./...
)
