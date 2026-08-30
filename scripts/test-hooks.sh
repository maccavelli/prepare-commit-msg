#!/usr/bin/env bash
set -euo pipefail

SOURCE_ROOT="$(git rev-parse --show-toplevel)"
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/prepare-commit-msg-hooks-test.XXXXXX")"
trap 'rm -rf "$TEST_ROOT"' EXIT

make_hook() {
	local path="$1"
	local label="$2"
	local consume_stdin="$3"

	{
		printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail'
		printf 'printf '\''%%s\\n'\'' %q >> "$HOOK_TEST_LOG"\n' "$label"
		if [ "$consume_stdin" = "yes" ]; then
			printf 'cat >> "$HOOK_TEST_INPUT"\n'
		fi
		printf 'exit "${HOOK_TEST_EXIT:-0}"\n'
	} > "$path"
	chmod +x "$path"
}

TEST_REPO="$TEST_ROOT/repository"
GLOBAL_HOOKS="$TEST_ROOT/global-hooks"
GLOBAL_CONFIG="$TEST_ROOT/global.gitconfig"
mkdir -p "$TEST_REPO/scripts" "$TEST_REPO/.githooks" "$GLOBAL_HOOKS"
git -C "$TEST_REPO" init -q
TEST_REPO="$(cd "$TEST_REPO" && pwd -P)"
GLOBAL_HOOKS="$(cd "$GLOBAL_HOOKS" && pwd -P)"
cp "$SOURCE_ROOT/scripts/install-hooks.sh" "$TEST_REPO/scripts/install-hooks.sh"
cp "$SOURCE_ROOT/scripts/uninstall-hooks.sh" "$TEST_REPO/scripts/uninstall-hooks.sh"

make_hook "$GLOBAL_HOOKS/pre-commit" "previous-pre-commit" "no"
make_hook "$GLOBAL_HOOKS/pre-push" "previous-pre-push" "yes"
make_hook "$TEST_REPO/.githooks/pre-commit" "repository-pre-commit" "no"
make_hook "$TEST_REPO/.githooks/pre-push" "repository-pre-push" "yes"

git config --file "$GLOBAL_CONFIG" core.hooksPath "$GLOBAL_HOOKS"
export GIT_CONFIG_GLOBAL="$GLOBAL_CONFIG"
export HOOK_TEST_LOG="$TEST_ROOT/hook.log"
export HOOK_TEST_INPUT="$TEST_ROOT/hook.input"

(
	cd "$TEST_REPO"
	./scripts/install-hooks.sh
	./scripts/install-hooks.sh
)

MANAGED_DIR="$(git -C "$TEST_REPO" rev-parse --absolute-git-dir)/prepare-commit-msg-hooks"
LOCAL_PATH="$(git -C "$TEST_REPO" config --local --get core.hooksPath)"
[ "$LOCAL_PATH" = "$MANAGED_DIR" ] || {
	echo "installer did not set the managed local hooks path" >&2
	exit 1
}

"$MANAGED_DIR/pre-commit"
printf '%s\n' "ref-line" | "$MANAGED_DIR/pre-push" origin https://example.invalid/repo.git

EXPECTED_LOG="$(printf '%s\n' \
	previous-pre-commit \
	repository-pre-commit \
	previous-pre-push \
	repository-pre-push)"
[ "$(cat "$HOOK_TEST_LOG")" = "$EXPECTED_LOG" ] || {
	echo "hook invocation order mismatch" >&2
	exit 1
}

EXPECTED_INPUT="$(printf '%s\n' ref-line ref-line)"
[ "$(cat "$HOOK_TEST_INPUT")" = "$EXPECTED_INPUT" ] || {
	echo "pre-push stdin was not replayed to both hooks" >&2
	exit 1
}

: > "$HOOK_TEST_LOG"
if HOOK_TEST_EXIT=7 "$MANAGED_DIR/pre-commit"; then
	echo "a failing previous hook did not block pre-commit" >&2
	exit 1
fi
[ "$(cat "$HOOK_TEST_LOG")" = "previous-pre-commit" ] || {
	echo "repository hook ran after the previous hook failed" >&2
	exit 1
}

cp "$MANAGED_DIR/pre-commit" "$TEST_ROOT/pre-commit.expected"
printf '%s\n' "# unexpected modification" >> "$MANAGED_DIR/pre-commit"
if (cd "$TEST_REPO" && ./scripts/install-hooks.sh >/dev/null 2>&1); then
	echo "installer overwrote a modified managed hook" >&2
	exit 1
fi
mv "$TEST_ROOT/pre-commit.expected" "$MANAGED_DIR/pre-commit"

(
	cd "$TEST_REPO"
	./scripts/uninstall-hooks.sh
)

if git -C "$TEST_REPO" config --local --get core.hooksPath >/dev/null 2>&1; then
	echo "uninstaller did not restore the absent local hooks path" >&2
	exit 1
fi
[ "$(git -C "$TEST_REPO" rev-parse --path-format=absolute --git-path hooks)" = "$GLOBAL_HOOKS" ] || {
	echo "uninstaller did not restore the effective global hooks path" >&2
	exit 1
}

git -C "$TEST_REPO" config --local core.hooksPath "$GLOBAL_HOOKS"
(
	cd "$TEST_REPO"
	./scripts/install-hooks.sh
	./scripts/uninstall-hooks.sh
)
[ "$(git -C "$TEST_REPO" config --local --get core.hooksPath)" = "$GLOBAL_HOOKS" ] || {
	echo "uninstaller did not restore the previous local hooks path" >&2
	exit 1
}

echo "hook composition tests passed"
