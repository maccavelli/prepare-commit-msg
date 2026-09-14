#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
GOLANGCI_LINT="$REPO_ROOT/.tools/bin/golangci-lint"
TEMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/prepare-commit-msg-staged.XXXXXX")"
SNAPSHOT_ROOT="$TEMP_ROOT/${REPO_ROOT##*/}"
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

mkdir -p "$SNAPSHOT_ROOT"
git -C "$REPO_ROOT" checkout-index --all --prefix="$SNAPSHOT_ROOT/"

# Preserve the sibling topology required by the temporary mcplib replacement.
# The consumer remains a staged snapshot; only its explicitly replaced module
# is read from the local checkout used by the real build.
if grep -Eq '^[[:space:]]*replace[[:space:]]+github\.com/maccavelli/mcplib[[:space:]]+=>[[:space:]]+\.\./mcplib([[:space:]]|$)' "$SNAPSHOT_ROOT/go.mod"; then
	LOCAL_MCPLIB="$REPO_ROOT/../mcplib"
	if [ ! -d "$LOCAL_MCPLIB" ]; then
		echo "local mcplib replacement not found: $LOCAL_MCPLIB" >&2
		exit 1
	fi
	ln -s "$LOCAL_MCPLIB" "$TEMP_ROOT/mcplib"
fi

STAGED_FILES=()
for candidate in "${GO_FILES[@]}"; do
	if [ -f "$SNAPSHOT_ROOT/$candidate" ]; then
		STAGED_FILES+=("$SNAPSHOT_ROOT/$candidate")
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
	cd "$SNAPSHOT_ROOT"
	"$GOLANGCI_LINT" fmt --diff -c .golangci.yml
	"$GOLANGCI_LINT" run -c .golangci.yml ./...
)
