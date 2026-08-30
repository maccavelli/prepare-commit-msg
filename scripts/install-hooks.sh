#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
GIT_DIR="$(git rev-parse --absolute-git-dir)"
MANAGED_DIR="$GIT_DIR/prepare-commit-msg-hooks"
MARKER="$MANAGED_DIR/.prepare-commit-msg-managed-hooks"
PREVIOUS_DIR_FILE="$MANAGED_DIR/.previous-hooks-dir"
LOCAL_PRESENT_FILE="$MANAGED_DIR/.previous-local-present"
LOCAL_VALUE_FILE="$MANAGED_DIR/.previous-local-value"

case "$MANAGED_DIR" in
"$GIT_DIR"/*) ;;
*)
	echo "refusing hooks directory outside repository Git directory" >&2
	exit 1
	;;
esac

if [ -e "$MANAGED_DIR" ] && [ ! -f "$MARKER" ]; then
	echo "refusing unmanaged hooks directory: $MANAGED_DIR" >&2
	exit 1
fi

if [ ! -e "$MANAGED_DIR" ]; then
	PREVIOUS_HOOKS_DIR="$(git rev-parse --path-format=absolute --git-path hooks)"
	[ "$PREVIOUS_HOOKS_DIR" != "$MANAGED_DIR" ] || {
		echo "refusing recursive managed hooks path" >&2
		exit 1
	}

	mkdir -p "$MANAGED_DIR"
	printf '%s\n' "prepare-commit-msg managed hooks v1" > "$MARKER"
	printf '%s\n' "$PREVIOUS_HOOKS_DIR" > "$PREVIOUS_DIR_FILE"
	if LOCAL_VALUE="$(git config --local --get core.hooksPath)"; then
		printf '%s\n' "yes" > "$LOCAL_PRESENT_FILE"
		printf '%s\n' "$LOCAL_VALUE" > "$LOCAL_VALUE_FILE"
	else
		printf '%s\n' "no" > "$LOCAL_PRESENT_FILE"
		: > "$LOCAL_VALUE_FILE"
	fi
else
	[ "$(cat "$MARKER")" = "prepare-commit-msg managed hooks v1" ] || {
		echo "managed hooks marker is invalid" >&2
		exit 1
	}
	PREVIOUS_HOOKS_DIR="$(cat "$PREVIOUS_DIR_FILE")"
	[ "$PREVIOUS_HOOKS_DIR" != "$MANAGED_DIR" ] || {
		echo "refusing recursive managed hooks path" >&2
		exit 1
	}
fi

install_wrapper() {
	local hook_name="$1"
	local candidate="$MANAGED_DIR/.$hook_name.candidate.$$"
	local target="$MANAGED_DIR/$hook_name"
	local previous_hook="$PREVIOUS_HOOKS_DIR/$hook_name"
	local repository_hook="$REPO_ROOT/.githooks/$hook_name"

	[ -x "$repository_hook" ] || {
		echo "repository hook is not executable: $repository_hook" >&2
		exit 1
	}

	{
		printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail'
		printf 'PREVIOUS_HOOK=%q\n' "$previous_hook"
		printf 'REPOSITORY_HOOK=%q\n' "$repository_hook"
		if [ "$hook_name" = "pre-push" ]; then
			printf '%s\n' \
				'HOOK_INPUT="$(mktemp "${TMPDIR:-/tmp}/prepare-commit-msg-pre-push.XXXXXX")"' \
				'trap '\''rm -f "$HOOK_INPUT"'\'' EXIT' \
				'cat > "$HOOK_INPUT"' \
				'if [ -x "$PREVIOUS_HOOK" ]; then' \
				'  "$PREVIOUS_HOOK" "$@" < "$HOOK_INPUT"' \
				'fi' \
				'"$REPOSITORY_HOOK" "$@" < "$HOOK_INPUT"'
		else
			printf '%s\n' \
				'if [ -x "$PREVIOUS_HOOK" ]; then' \
				'  "$PREVIOUS_HOOK" "$@"' \
				'fi' \
				'"$REPOSITORY_HOOK" "$@"'
		fi
	} > "$candidate"
	chmod +x "$candidate"

	if [ -e "$target" ]; then
		if ! cmp -s "$candidate" "$target"; then
			rm -f "$candidate"
			echo "refusing to overwrite modified managed hook: $target" >&2
			exit 1
		fi
		rm -f "$candidate"
	else
		mv "$candidate" "$target"
	fi
}

install_wrapper pre-commit
install_wrapper pre-push

if [ -d "$PREVIOUS_HOOKS_DIR" ]; then
	for hook_file in "$PREVIOUS_HOOKS_DIR"/*; do
		[ -f "$hook_file" ] && [ -x "$hook_file" ] || continue
		hook_name="$(basename "$hook_file")"
		[ ! -e "$MANAGED_DIR/$hook_name" ] || continue
		candidate="$MANAGED_DIR/.$hook_name.candidate.$$"
		target="$MANAGED_DIR/$hook_name"
		{
			printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail'
			printf 'PREVIOUS_HOOK=%q\n' "$hook_file"
			printf '%s\n' 'exec "$PREVIOUS_HOOK" "$@"'
		} > "$candidate"
		chmod +x "$candidate"
		mv "$candidate" "$target"
	done
fi

git config --local core.hooksPath "$MANAGED_DIR"
echo "installed composable hooks in $MANAGED_DIR"
