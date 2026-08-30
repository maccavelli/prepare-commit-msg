#!/usr/bin/env bash
set -euo pipefail

GIT_DIR="$(git rev-parse --absolute-git-dir)"
MANAGED_DIR="$GIT_DIR/prepare-commit-msg-hooks"
MARKER="$MANAGED_DIR/.prepare-commit-msg-managed-hooks"
LOCAL_PRESENT_FILE="$MANAGED_DIR/.previous-local-present"
LOCAL_VALUE_FILE="$MANAGED_DIR/.previous-local-value"

case "$MANAGED_DIR" in
"$GIT_DIR"/*) ;;
*)
	echo "refusing hooks directory outside repository Git directory" >&2
	exit 1
	;;
esac

[ -f "$MARKER" ] || {
	echo "managed hooks are not installed" >&2
	exit 1
}
[ "$(cat "$MARKER")" = "prepare-commit-msg managed hooks v1" ] || {
	echo "managed hooks marker is invalid" >&2
	exit 1
}

CURRENT_LOCAL="$(git config --local --get core.hooksPath || true)"
[ "$CURRENT_LOCAL" = "$MANAGED_DIR" ] || {
	echo "local core.hooksPath no longer points to the managed directory" >&2
	exit 1
}

if [ "$(cat "$LOCAL_PRESENT_FILE")" = "yes" ]; then
	git config --local core.hooksPath "$(cat "$LOCAL_VALUE_FILE")"
else
	git config --local --unset-all core.hooksPath || true
fi

rm -rf -- "$MANAGED_DIR"
echo "restored the previous local core.hooksPath state"
