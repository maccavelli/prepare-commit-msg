#!/usr/bin/env bash
set -euo pipefail

EXPECTED_REPOSITORY="maccavelli/prepare-commit-msg"
REPOSITORY="${GITHUB_REPOSITORY:-$EXPECTED_REPOSITORY}"
API_VERSION="2026-03-10"
MODE="dry-run"

case "${1:-}" in
	"") ;;
	--apply) MODE="apply" ;;
	*)
		echo "usage: $0 [--apply]" >&2
		exit 2
		;;
esac

for command_name in gh git jq; do
	command -v "$command_name" >/dev/null 2>&1 || {
		echo "required command is not installed: $command_name" >&2
		exit 1
	}
done

[ "$REPOSITORY" = "$EXPECTED_REPOSITORY" ] || {
	echo "refusing unexpected repository: $REPOSITORY" >&2
	exit 1
}

if ! gh auth status --hostname github.com >/dev/null 2>&1; then
	echo "GitHub CLI is not authenticated for github.com" >&2
	exit 1
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

ORIGIN_URL="$(git remote get-url origin)"
ORIGIN_WITHOUT_SUFFIX="${ORIGIN_URL%.git}"
case "$ORIGIN_WITHOUT_SUFFIX" in
	"https://github.com/$REPOSITORY" | "git@github.com:$REPOSITORY" | \
		"ssh://git@github.com/$REPOSITORY") ;;
	*)
		echo "origin does not identify $REPOSITORY: $ORIGIN_URL" >&2
		exit 1
		;;
esac

CONFIG_TMP="$(mktemp -d "${TMPDIR:-/tmp}/prepare-commit-msg-github-config.XXXXXX")"
trap 'rm -rf "$CONFIG_TMP"' EXIT

api() {
	gh api -H "X-GitHub-Api-Version: $API_VERSION" "$@"
}

api "repos/$REPOSITORY" >"$CONFIG_TMP/repository.json"
if [ "$(jq -r '.full_name' "$CONFIG_TMP/repository.json")" != "$REPOSITORY" ]; then
	echo "authenticated API repository identity does not match $REPOSITORY" >&2
	exit 1
fi
if [ "$(jq -r '.permissions.admin // false' "$CONFIG_TMP/repository.json")" != "true" ]; then
	echo "authenticated account lacks administrator access to $REPOSITORY" >&2
	exit 1
fi
DEFAULT_BRANCH="$(jq -r '.default_branch' "$CONFIG_TMP/repository.json")"
[ "$DEFAULT_BRANCH" = "main" ] || {
	echo "expected default branch main, found $DEFAULT_BRANCH" >&2
	exit 1
}

api "apps/github-actions" >"$CONFIG_TMP/github-actions-app.json"
if [ "$(jq -r '.slug' "$CONFIG_TMP/github-actions-app.json")" != \
	"github-actions" ]; then
	echo "could not resolve the GitHub Actions integration" >&2
	exit 1
fi
GITHUB_ACTIONS_APP_ID="$(jq -r '.id' "$CONFIG_TMP/github-actions-app.json")"

HARDENED_WORKFLOW_SHA="$(git log -1 --format=%H -- \
	.github/workflows/quality.yml \
	.github/workflows/ci.yml \
	.github/workflows/release.yml)"
[ -n "$HARDENED_WORKFLOW_SHA" ] || {
	echo "could not resolve the hardened workflow commit" >&2
	exit 1
}
REMOTE_MAIN_SHA="$(api "repos/$REPOSITORY/commits/$DEFAULT_BRANCH" --jq '.sha')"

REMOTE_READY=true
if ! api "repos/$REPOSITORY/commits/$HARDENED_WORKFLOW_SHA" \
	>/dev/null 2>&1; then
	REMOTE_READY=false
else
	COMPARE_STATUS="$(api \
		"repos/$REPOSITORY/compare/$HARDENED_WORKFLOW_SHA...$REMOTE_MAIN_SHA" \
		--jq '.status')"
	case "$COMPARE_STATUS" in
		ahead | identical) ;;
		*) REMOTE_READY=false ;;
	esac
fi

for workflow_file in quality.yml ci.yml release.yml; do
	LOCAL_BLOB="$(git hash-object ".github/workflows/$workflow_file")"
	REMOTE_BLOB="$(api \
		"repos/$REPOSITORY/contents/.github/workflows/$workflow_file?ref=$DEFAULT_BRANCH" \
		--jq '.sha' 2>/dev/null || true)"
	if [ "$LOCAL_BLOB" != "$REMOTE_BLOB" ]; then
		REMOTE_READY=false
	fi
done

jq -n '{
	enabled: true,
	allowed_actions: "selected",
	sha_pinning_required: true
}' >"$CONFIG_TMP/actions-policy.json"

jq -n '{
	github_owned_allowed: true,
	verified_allowed: false,
	patterns_allowed: []
}' >"$CONFIG_TMP/selected-actions.json"

jq -n '{
	default_workflow_permissions: "read",
	can_approve_pull_request_reviews: false
}' >"$CONFIG_TMP/workflow-permissions.json"

if api "repos/$REPOSITORY/environments/release" \
	>"$CONFIG_TMP/current-environment.json" 2>"$CONFIG_TMP/environment-error"; then
	jq '{
		wait_timer: ([.protection_rules[]? | select(.type == "wait_timer") |
			.wait_timer][0] // 0),
		prevent_self_review: ([.protection_rules[]? |
			select(.type == "required_reviewers") | .prevent_self_review][0] // false),
		reviewers: ([.protection_rules[]? | select(.type == "required_reviewers") |
			.reviewers[]? | {type, id: .reviewer.id}]),
		deployment_branch_policy: {
			protected_branches: false,
			custom_branch_policies: true
		}
	}' "$CONFIG_TMP/current-environment.json" >"$CONFIG_TMP/environment.json"
elif grep -q '(HTTP 404)' "$CONFIG_TMP/environment-error"; then
	jq -n '{
		deployment_branch_policy: {
			protected_branches: false,
			custom_branch_policies: true
		}
	}' >"$CONFIG_TMP/environment.json"
else
	echo "could not read the release environment" >&2
	sed -n '1,5p' "$CONFIG_TMP/environment-error" >&2
	exit 1
fi

jq -n '{
	name: "prepare-commit-msg-main",
	target: "branch",
	enforcement: "active",
	bypass_actors: [{
		actor_id: 5,
		actor_type: "RepositoryRole",
		bypass_mode: "always"
	}],
	conditions: {ref_name: {include: ["~DEFAULT_BRANCH"], exclude: []}},
	rules: [{type: "deletion"}, {type: "non_fast_forward"}]
}' >"$CONFIG_TMP/main-ruleset.json"

jq -n --argjson actions_app_id "$GITHUB_ACTIONS_APP_ID" '{
	name: "prepare-commit-msg-release-tags",
	target: "tag",
	enforcement: "active",
	bypass_actors: [
		{
			actor_id: 5,
			actor_type: "RepositoryRole",
			bypass_mode: "always"
		},
		{
			actor_id: $actions_app_id,
			actor_type: "Integration",
			bypass_mode: "always"
		}
	],
	conditions: {ref_name: {include: ["refs/tags/v*"], exclude: []}},
	rules: [
		{type: "creation"},
		{type: "deletion"},
		{type: "non_fast_forward"}
	]
}' >"$CONFIG_TMP/tag-ruleset.json"

jq -n -S \
	--arg mode "$MODE" \
	--arg repository "$REPOSITORY" \
	--arg hardened_workflow_sha "$HARDENED_WORKFLOW_SHA" \
	--arg remote_main_sha "$REMOTE_MAIN_SHA" \
	--argjson remote_ready "$REMOTE_READY" \
	--slurpfile actions "$CONFIG_TMP/actions-policy.json" \
	--slurpfile selected "$CONFIG_TMP/selected-actions.json" \
	--slurpfile workflow "$CONFIG_TMP/workflow-permissions.json" \
	--slurpfile environment "$CONFIG_TMP/environment.json" \
	--slurpfile main_ruleset "$CONFIG_TMP/main-ruleset.json" \
	--slurpfile tag_ruleset "$CONFIG_TMP/tag-ruleset.json" '{
	mode: $mode,
	repository: $repository,
	hardened_workflow_sha: $hardened_workflow_sha,
	remote_main_sha: $remote_main_sha,
	remote_ready: $remote_ready,
	mutations: [
		{method: "PUT", endpoint: "actions/permissions", body: $actions[0]},
		{method: "PUT", endpoint: "actions/permissions/selected-actions", body: $selected[0]},
		{method: "PUT", endpoint: "actions/permissions/workflow", body: $workflow[0]},
		{method: "DELETE", endpoint: "automated-security-fixes", body: null},
		{method: "PUT", endpoint: "environments/release", body: $environment[0]},
		{method: "RECONCILE", endpoint: "environments/release/deployment-branch-policies",
			body: {name: "main", type: "branch"}},
		{method: "UPSERT_BY_NAME", endpoint: "rulesets", body: $main_ruleset[0]},
		{method: "UPSERT_BY_NAME", endpoint: "rulesets", body: $tag_ruleset[0]}
	]
}'

if [ "$MODE" = "dry-run" ]; then
	if [ "$REMOTE_READY" != "true" ]; then
		echo "dry-run only: hardened workflows are not yet present on origin/main" >&2
	fi
	exit 0
fi

if [ "$REMOTE_READY" != "true" ]; then
	echo "refusing --apply: hardened workflows are not present on origin/main" >&2
	exit 1
fi

if ! git diff --quiet -- .github/workflows || \
	! git diff --cached --quiet -- .github/workflows; then
	echo "refusing --apply: hardened workflow files have uncommitted changes" >&2
	exit 1
fi

SNAPSHOT="$REPO_ROOT/.git/prepare-commit-msg-github-settings-before.json"
if [ ! -e "$SNAPSHOT" ]; then
	"$REPO_ROOT/scripts/audit-github.sh" >"$SNAPSHOT"
	chmod 600 "$SNAPSHOT"
	echo "saved pre-apply audit to $SNAPSHOT" >&2
else
	echo "preserving existing pre-apply audit at $SNAPSHOT" >&2
fi

api --method PUT "repos/$REPOSITORY/actions/permissions" \
	--input "$CONFIG_TMP/actions-policy.json" >/dev/null
api --method PUT "repos/$REPOSITORY/actions/permissions/selected-actions" \
	--input "$CONFIG_TMP/selected-actions.json" >/dev/null
api --method PUT "repos/$REPOSITORY/actions/permissions/workflow" \
	--input "$CONFIG_TMP/workflow-permissions.json" >/dev/null
api --method DELETE "repos/$REPOSITORY/automated-security-fixes" >/dev/null
api --method PUT "repos/$REPOSITORY/environments/release" \
	--input "$CONFIG_TMP/environment.json" >/dev/null

api "repos/$REPOSITORY/environments/release/deployment-branch-policies" \
	>"$CONFIG_TMP/branch-policies.json"
FOUND_MAIN_POLICY=false
while IFS=$'\t' read -r policy_id policy_name policy_type; do
	if [ "$policy_name" = "main" ] && [ "$policy_type" = "branch" ]; then
		FOUND_MAIN_POLICY=true
	else
		api --method DELETE \
			"repos/$REPOSITORY/environments/release/deployment-branch-policies/$policy_id" \
			>/dev/null
	fi
done < <(jq -r '.branch_policies[]? | [.id, .name, (.type // "branch")] | @tsv' \
	"$CONFIG_TMP/branch-policies.json")

if [ "$FOUND_MAIN_POLICY" != "true" ]; then
	jq -n '{name: "main", type: "branch"}' >"$CONFIG_TMP/main-policy.json"
	api --method POST \
		"repos/$REPOSITORY/environments/release/deployment-branch-policies" \
		--input "$CONFIG_TMP/main-policy.json" >/dev/null
fi

api "repos/$REPOSITORY/rulesets?includes_parents=false" \
	>"$CONFIG_TMP/rulesets.json"

upsert_ruleset() {
	local name="$1"
	local payload="$2"
	local count=""
	local ruleset_id=""

	count="$(jq --arg name "$name" '[.[] | select(.name == $name)] | length' \
		"$CONFIG_TMP/rulesets.json")"
	if [ "$count" -gt 1 ]; then
		echo "refusing duplicate managed rulesets named $name" >&2
		exit 1
	elif [ "$count" -eq 1 ]; then
		ruleset_id="$(jq -r --arg name "$name" \
			'.[] | select(.name == $name) | .id' "$CONFIG_TMP/rulesets.json")"
		api --method PUT "repos/$REPOSITORY/rulesets/$ruleset_id" \
			--input "$payload" >/dev/null
	else
		api --method POST "repos/$REPOSITORY/rulesets" \
			--input "$payload" >/dev/null
	fi
}

upsert_ruleset "prepare-commit-msg-main" "$CONFIG_TMP/main-ruleset.json"
upsert_ruleset "prepare-commit-msg-release-tags" "$CONFIG_TMP/tag-ruleset.json"

echo "GitHub repository enforcement converged; resulting audit follows" >&2
"$REPO_ROOT/scripts/audit-github.sh"
