#!/usr/bin/env bash
set -euo pipefail

SOURCE_ROOT="$(git rev-parse --show-toplevel)"
TEST_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/prepare-commit-msg-hooks-test.XXXXXX")"
trap 'rm -rf "$TEST_ROOT"' EXIT

make_hook() {
	local path="$1"
	local label="$2"
	local consume_stdin="$3"

	# Literal dollars are emitted into the generated fixture hook.
	# shellcheck disable=SC2016
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
make_hook "$GLOBAL_HOOKS/post-commit" "previous-post-commit" "no"
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
[ -x "$MANAGED_DIR/pre-commit" ] || {
	echo "installer did not preserve the previous pre-commit hook" >&2
	exit 1
}
grep -q '^PREVIOUS_HOOK=' "$MANAGED_DIR/pre-commit" || {
	echo "managed pre-commit does not delegate to the previous hook" >&2
	exit 1
}
if grep -q 'REPOSITORY_HOOK\|\.githooks/pre-commit' "$MANAGED_DIR/pre-commit"; then
	echo "managed pre-commit still invokes the repository hook" >&2
	exit 1
fi

"$MANAGED_DIR/pre-commit"
printf '%s\n' "ref-line" | "$MANAGED_DIR/pre-push" origin https://example.invalid/repo.git

EXPECTED_LOG="$(printf '%s\n' \
	previous-pre-commit \
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

cp "$MANAGED_DIR/post-commit" "$TEST_ROOT/post-commit.expected"
printf '%s\n' "# unexpected modification" >> "$MANAGED_DIR/post-commit"
if (cd "$TEST_REPO" && ./scripts/install-hooks.sh >/dev/null 2>&1); then
	echo "installer overwrote a modified carried-forward hook" >&2
	exit 1
fi
mv "$TEST_ROOT/post-commit.expected" "$MANAGED_DIR/post-commit"

: > "$HOOK_TEST_LOG"
if HOOK_TEST_EXIT=7 "$MANAGED_DIR/pre-commit"; then
	echo "a failing previous hook did not block pre-commit" >&2
	exit 1
fi
[ "$(cat "$HOOK_TEST_LOG")" = "previous-pre-commit" ] || {
	echo "unexpected hook ran after the previous hook failed" >&2
	exit 1
}

# Literal dollars are emitted into the generated legacy hook fixture.
# shellcheck disable=SC2016
{
	printf '%s\n' '#!/usr/bin/env bash' 'set -euo pipefail'
	printf 'PREVIOUS_HOOK=%q\n' "$GLOBAL_HOOKS/pre-commit"
	printf 'REPOSITORY_HOOK=%q\n' "$TEST_REPO/.githooks/pre-commit"
	printf '%s\n' \
		'if [ -x "$PREVIOUS_HOOK" ]; then' \
		'  "$PREVIOUS_HOOK" "$@"' \
		'fi' \
		'"$REPOSITORY_HOOK" "$@"'
} > "$MANAGED_DIR/pre-commit"
chmod +x "$MANAGED_DIR/pre-commit"
(
	cd "$TEST_REPO"
	./scripts/install-hooks.sh
)
if grep -q 'REPOSITORY_HOOK\|\.githooks/pre-commit' "$MANAGED_DIR/pre-commit"; then
	echo "installer did not migrate the legacy pre-commit wrapper" >&2
	exit 1
fi

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

NO_PRECOMMIT_REPO="$TEST_ROOT/no-precommit-repository"
NO_PRECOMMIT_HOOKS="$TEST_ROOT/no-precommit-global-hooks"
NO_PRECOMMIT_CONFIG="$TEST_ROOT/no-precommit-global.gitconfig"
mkdir -p "$NO_PRECOMMIT_REPO/scripts" "$NO_PRECOMMIT_REPO/.githooks" "$NO_PRECOMMIT_HOOKS"
git -C "$NO_PRECOMMIT_REPO" init -q
NO_PRECOMMIT_REPO="$(cd "$NO_PRECOMMIT_REPO" && pwd -P)"
NO_PRECOMMIT_HOOKS="$(cd "$NO_PRECOMMIT_HOOKS" && pwd -P)"
cp "$SOURCE_ROOT/scripts/install-hooks.sh" "$NO_PRECOMMIT_REPO/scripts/install-hooks.sh"
make_hook "$NO_PRECOMMIT_HOOKS/pre-push" "no-precommit-previous-pre-push" "yes"
make_hook "$NO_PRECOMMIT_REPO/.githooks/pre-push" "no-precommit-repository-pre-push" "yes"
git config --file "$NO_PRECOMMIT_CONFIG" core.hooksPath "$NO_PRECOMMIT_HOOKS"
(
	export GIT_CONFIG_GLOBAL="$NO_PRECOMMIT_CONFIG"
	cd "$NO_PRECOMMIT_REPO"
	./scripts/install-hooks.sh
)
NO_PRECOMMIT_MANAGED="$(git -C "$NO_PRECOMMIT_REPO" rev-parse --absolute-git-dir)/prepare-commit-msg-hooks"
[ ! -e "$NO_PRECOMMIT_MANAGED/pre-commit" ] || {
	echo "installer created a managed pre-commit without a host pre-commit" >&2
	exit 1
}

echo "hook composition tests passed"
