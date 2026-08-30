#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
ACTIONLINT="${ACTIONLINT:-$REPO_ROOT/.tools/bin/actionlint}"

while IFS= read -r script; do
	bash -n "$script"
done < <(
	find "$REPO_ROOT/scripts" "$REPO_ROOT/.githooks" \
		-type f 2>/dev/null | LC_ALL=C sort
)

"$ACTIONLINT" "$REPO_ROOT"/.github/workflows/*.yml
