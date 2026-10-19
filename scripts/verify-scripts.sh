#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
ACTIONLINT="${ACTIONLINT:-$REPO_ROOT/.tools/bin/actionlint}"
PYTHON="${PYTHON:-python3}"

while IFS= read -r script; do
	case "$(head -n 1 "$script")" in
	*python*) "$PYTHON" -c 'import ast, sys; ast.parse(open(sys.argv[1], encoding="utf-8").read(), sys.argv[1])' "$script" ;;
	*) bash -n "$script" ;;
	esac
done < <(
	find "$REPO_ROOT/scripts" "$REPO_ROOT/.githooks" \
		-type f 2>/dev/null | LC_ALL=C sort
)

# tools/ holds Python and a README; only its .py files are scripts (0016-MADR D1).
while IFS= read -r script; do
	"$PYTHON" -c 'import ast, sys; ast.parse(open(sys.argv[1], encoding="utf-8").read(), sys.argv[1])' "$script"
done < <(
	find "$REPO_ROOT/tools" -name '*.py' -not -path '*/__pycache__/*' -type f 2>/dev/null | LC_ALL=C sort
)

"$ACTIONLINT" "$REPO_ROOT"/.github/workflows/*.yml
