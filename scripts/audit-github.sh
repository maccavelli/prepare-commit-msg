#!/usr/bin/env bash
set -euo pipefail

EXPECTED_REPOSITORY="maccavelli/prepare-commit-msg"
REPOSITORY="${GITHUB_REPOSITORY:-$EXPECTED_REPOSITORY}"
API_VERSION="2026-03-10"

for command_name in gh jq; do
	command -v "$command_name" >/dev/null 2>&1 || {
		echo "required command is not installed: $command_name" >&2
		exit 1
	}
done

if ! gh auth status >/dev/null 2>&1; then
	echo "GitHub CLI is not authenticated" >&2
	exit 1
fi

AUDIT_TMP="$(mktemp -d "${TMPDIR:-/tmp}/prepare-commit-msg-github-audit.XXXXXX")"
trap 'rm -rf "$AUDIT_TMP"' EXIT

api_get() {
	local endpoint="$1"
	local output="$2"
	local error_file="$AUDIT_TMP/api-error"

	if ! gh api -H "X-GitHub-Api-Version: $API_VERSION" "$endpoint" \
		>"$output" 2>"$error_file"; then
		echo "GitHub API request failed: $endpoint" >&2
		sed -n '1,5p' "$error_file" >&2
		exit 1
	fi
}

api_optional() {
	local endpoint="$1"
	local output="$2"
	local error_file="$AUDIT_TMP/api-error"
	local response="$AUDIT_TMP/api-response"

	if gh api -H "X-GitHub-Api-Version: $API_VERSION" "$endpoint" \
		>"$response" 2>"$error_file"; then
		jq -S '{state: "present", data: .}' "$response" >"$output"
	elif grep -q '(HTTP 404)' "$error_file"; then
		printf '%s\n' '{"state":"absent"}' | jq -S . >"$output"
	else
		echo "GitHub API request failed: $endpoint" >&2
		sed -n '1,5p' "$error_file" >&2
		exit 1
	fi
}

api_get "repos/$REPOSITORY" "$AUDIT_TMP/repository.json"
api_get "repos/$REPOSITORY/actions/permissions" "$AUDIT_TMP/actions.json"
api_get "repos/$REPOSITORY/actions/permissions/workflow" \
	"$AUDIT_TMP/workflow-permissions.json"

if [ "$(jq -r '.allowed_actions' "$AUDIT_TMP/actions.json")" = "selected" ]; then
	api_optional "repos/$REPOSITORY/actions/permissions/selected-actions" \
		"$AUDIT_TMP/selected-actions.json"
else
	printf '%s\n' '{"state":"not_applicable"}' | jq -S . \
		>"$AUDIT_TMP/selected-actions.json"
fi

api_optional "repos/$REPOSITORY/branches/main/protection" \
	"$AUDIT_TMP/main-protection.json"
api_optional "repos/$REPOSITORY/automated-security-fixes" \
	"$AUDIT_TMP/dependabot-security-updates.json"
api_optional "repos/$REPOSITORY/environments/release" \
	"$AUDIT_TMP/release-environment.json"

if [ "$(jq -r '.state' "$AUDIT_TMP/release-environment.json")" = "present" ]; then
	api_get "repos/$REPOSITORY/environments/release/deployment-branch-policies" \
		"$AUDIT_TMP/release-branch-policies.json"
else
	printf '%s\n' '{"branch_policies":[]}' \
		>"$AUDIT_TMP/release-branch-policies.json"
fi

api_get "repos/$REPOSITORY/rulesets?includes_parents=true" \
	"$AUDIT_TMP/ruleset-index.json"
mkdir "$AUDIT_TMP/rulesets"
while IFS= read -r ruleset_id; do
	[ -n "$ruleset_id" ] || continue
	api_get "repos/$REPOSITORY/rulesets/$ruleset_id" \
		"$AUDIT_TMP/rulesets/$ruleset_id.json"
done < <(jq -r '.[].id' "$AUDIT_TMP/ruleset-index.json")

if compgen -G "$AUDIT_TMP/rulesets/*.json" >/dev/null; then
	jq -s -S '
		map({
			id,
			name,
			target,
			source_type,
			source,
			enforcement,
			bypass_actors: ((.bypass_actors // []) | sort_by(.actor_type, .actor_id)),
			conditions,
			rules
		}) | sort_by(.name, .id)
	' "$AUDIT_TMP"/rulesets/*.json >"$AUDIT_TMP/rulesets-normalized.json"
else
	printf '%s\n' '[]' >"$AUDIT_TMP/rulesets-normalized.json"
fi

jq -n -S \
	--slurpfile repository "$AUDIT_TMP/repository.json" \
	--slurpfile actions "$AUDIT_TMP/actions.json" \
	--slurpfile workflow "$AUDIT_TMP/workflow-permissions.json" \
	--slurpfile selected "$AUDIT_TMP/selected-actions.json" \
	--slurpfile protection "$AUDIT_TMP/main-protection.json" \
	--slurpfile rulesets "$AUDIT_TMP/rulesets-normalized.json" \
	--slurpfile environment "$AUDIT_TMP/release-environment.json" \
	--slurpfile policies "$AUDIT_TMP/release-branch-policies.json" \
	--slurpfile dependabot "$AUDIT_TMP/dependabot-security-updates.json" '
	{
		repository: {
			name_with_owner: $repository[0].full_name,
			default_branch: $repository[0].default_branch,
			visibility: $repository[0].visibility,
			admin_access: ($repository[0].permissions.admin // false)
		},
		actions: {
			policy: ($actions[0] | {
				enabled,
				allowed_actions,
				sha_pinning_required
			}),
			selected_actions: (
				if $selected[0].state == "present" then
					{
						state: "present",
						data: ($selected[0].data | {
							github_owned_allowed,
							verified_allowed,
							patterns_allowed: ((.patterns_allowed // []) | sort)
						})
					}
				else $selected[0] end
			),
			workflow_permissions: ($workflow[0] | {
				default_workflow_permissions,
				can_approve_pull_request_reviews
			})
		},
		main_branch_protection: (
			if $protection[0].state == "absent" then $protection[0]
			else {
				state: "present",
				data: ($protection[0].data | {
					required_status_checks,
					enforce_admins,
					required_pull_request_reviews,
					restrictions,
					required_signatures,
					allow_force_pushes,
					allow_deletions
				})
			} end
		),
		rulesets: $rulesets[0],
		release_environment: (
			if $environment[0].state == "absent" then $environment[0]
			else {
				state: "present",
				data: {
					name: $environment[0].data.name,
					protection_rules: ($environment[0].data.protection_rules // []),
					deployment_branch_policy: $environment[0].data.deployment_branch_policy,
					branch_policies: (($policies[0].branch_policies // []) |
						map({id, name, type}) | sort_by(.type, .name))
				}
			} end
		),
		security: {
			dependabot_security_updates: $dependabot[0],
			secret_scanning: ($repository[0].security_and_analysis.secret_scanning.status // "unavailable"),
			secret_scanning_push_protection: (
				$repository[0].security_and_analysis.secret_scanning_push_protection.status // "unavailable"
			)
		}
	}
'
