#!/usr/bin/env bash
set -euo pipefail

sha256_file() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

write_checksums() {
	local dist_dir="$1"
	local temporary="$dist_dir/.SHA256SUMS.tmp"
	local asset=""

	[ -d "$dist_dir" ] || {
		echo "release directory does not exist: $dist_dir" >&2
		exit 1
	}

	rm -f "$dist_dir/SHA256SUMS" "$temporary"
	for asset in "$dist_dir"/*; do
		[ -f "$asset" ] || continue
		printf '%s  %s\n' "$(sha256_file "$asset")" "$(basename "$asset")"
	done | LC_ALL=C sort -k2 > "$temporary"
	mv "$temporary" "$dist_dir/SHA256SUMS"
}

if [ "${1:-}" = "--write-checksums" ]; then
	[ "$#" -eq 2 ] || {
		echo "usage: $0 --write-checksums DIST_DIR" >&2
		exit 2
	}
	write_checksums "$2"
	exit 0
fi

[ "$#" -eq 2 ] || {
	echo "usage: $0 VERSION DIST_DIR" >&2
	exit 2
}

VERSION="$1"
DIST_DIR="$2"
REPO_ROOT="$(git rev-parse --show-toplevel)"

if [[ ! "$VERSION" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
	echo "release version must match vMAJOR.MINOR.PATCH: $VERSION" >&2
	exit 1
fi

if ! git -C "$REPO_ROOT" diff --quiet || ! git -C "$REPO_ROOT" diff --cached --quiet; then
	echo "release source tree has tracked changes" >&2
	exit 1
fi

EXPECTED_ASSETS="$(cat <<'ASSETS' | LC_ALL=C sort
SHA256SUMS
prepare-commit-msg-darwin-amd64
prepare-commit-msg-darwin-arm64
prepare-commit-msg-linux-amd64
prepare-commit-msg-linux-arm64
prepare-commit-msg-windows-amd64.exe
prepare-commit-msg-windows-arm64.exe
ASSETS
)"

ACTUAL_ASSETS="$(
	for asset in "$DIST_DIR"/*; do
		[ -f "$asset" ] || continue
		basename "$asset"
	done | LC_ALL=C sort
)"

if [ "$ACTUAL_ASSETS" != "$EXPECTED_ASSETS" ]; then
	echo "release asset set mismatch" >&2
	echo "expected:" >&2
	printf '%s\n' "$EXPECTED_ASSETS" >&2
	echo "actual:" >&2
	printf '%s\n' "$ACTUAL_ASSETS" >&2
	exit 1
fi

while read -r expected name extra; do
	[ -z "${extra:-}" ] || {
		echo "malformed checksum line for $name" >&2
		exit 1
	}
	[ "$name" != "SHA256SUMS" ] || {
		echo "checksum manifest must not hash itself" >&2
		exit 1
	}
	actual="$(sha256_file "$DIST_DIR/$name")"
	[ "$actual" = "$expected" ] || {
		echo "checksum mismatch for $name" >&2
		exit 1
	}
done < "$DIST_DIR/SHA256SUMS"

if [ "$(wc -l < "$DIST_DIR/SHA256SUMS" | tr -d ' ')" -ne 6 ]; then
	echo "checksum manifest must contain exactly six entries" >&2
	exit 1
fi

GOOS="$(go env GOOS)"
GOARCH="$(go env GOARCH)"
SUFFIX=""
[ "$GOOS" != "windows" ] || SUFFIX=".exe"
NATIVE_BINARY="$DIST_DIR/prepare-commit-msg-$GOOS-$GOARCH$SUFFIX"
if [ -x "$NATIVE_BINARY" ]; then
	NATIVE_VERSION="$($NATIVE_BINARY version)"
	[ "$NATIVE_VERSION" = "prepare-commit-msg version ${VERSION#v} (release)" ] || {
		echo "embedded version mismatch: $NATIVE_VERSION" >&2
		exit 1
	}
fi
