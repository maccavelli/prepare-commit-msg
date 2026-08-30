# CI/CD Operations Runbook

This runbook implements
[`0003-MADR-layer-and-harden-ci-cd-quality-gates.md`](0003-MADR-layer-and-harden-ci-cd-quality-gates.md).
The repository-owned commands are the source of truth; workflows and hooks
invoke those commands instead of maintaining separate check lists.

## Routine verification

Install the pinned tools and run the complete gate before exporting or pushing
a change:

```bash
make tools
make verify
```

`make verify` checks module tidiness and checksums, formatting and imports,
static analysis, race-enabled tests, the 80% aggregate coverage threshold,
reachable vulnerabilities, shell and workflow syntax, and all six release
cross-builds. Use `make verify-staged` for the exact staged Go snapshot.

Install the composable local hooks with:

```bash
make hooks-install
make hooks-test
```

The installer preserves and invokes the previously effective hooks. To remove
only the managed repository layer and restore the prior `core.hooksPath` state:

```bash
make hooks-uninstall
```

## Repository-settings audit

The audit is read-only, prints normalized JSON, and reports optional resources
as `present` or `absent`. API failures terminate the command rather than being
misreported as absent resources.

```bash
scripts/audit-github.sh > github-settings.json
scripts/audit-github.sh > github-settings.second.json
diff -u github-settings.json github-settings.second.json
```

The output includes the Actions policy, default workflow token permissions,
selected-action policy, legacy `main` protection, complete rulesets, the
`release` environment and branch policies, Dependabot security updates, secret
scanning, and push protection. It contains no credentials.

## Repository-settings activation

Configuration is dry-run by default and emits every intended endpoint, method,
and JSON body:

```bash
scripts/configure-github.sh | jq .
```

The intended converged state is:

- read-only default `GITHUB_TOKEN` permissions and no pull-request approvals;
- GitHub-authored and repository-local actions only, with full-SHA pinning;
- Dependabot security updates enabled;
- a `release` environment restricted to the `main` branch;
- a default-branch ruleset that blocks deletion and non-fast-forward updates,
  with repository administrators able to publish mirror snapshots; and
- a `v*` tag ruleset that restricts creation to repository administrators and
  the GitHub Actions integration and blocks deletion and non-fast-forward
  updates for other actors.

The apply path verifies the GitHub host, repository identity, administrator
access, default branch, origin URL, deployed workflow commit, and remote
workflow blob IDs before its first mutation. It therefore refuses activation
until the implementation has been pushed to `origin/main`.

Pushing is a separate, explicit maintainer action. After the implementation is
on `origin/main` and remote activation has been separately authorized:

```bash
scripts/configure-github.sh --apply
scripts/audit-github.sh
```

The first apply preserves a mode-`0600` baseline at
`.git/prepare-commit-msg-github-settings-before.json`; later applies preserve
that original snapshot. Named rulesets are updated in place, the environment
branch policy is reconciled, and repeated runs converge without duplicates.

Required reviewers are intentionally not configured because no reviewer
identity was supplied. A repository administrator may add reviewers later,
but must preserve the `main` deployment branch policy and capture audits before
and after the change.

## Controlled release

Choose a new strict `vMAJOR.MINOR.PATCH` value and dispatch from `main`:

```bash
VERSION=v1.2.3
gh workflow run release.yml --ref main \
  -f version="$VERSION" \
  -f prerelease=false
gh run list --workflow release.yml --limit 1
```

The workflow runs the shared quality workflow before any publication step,
builds the six binaries once from the dispatched SHA, creates and validates
`SHA256SUMS`, and records build-provenance attestations. The publish job then
downloads and independently verifies exactly seven files, refuses an existing
tag or release, creates a draft, verifies its metadata and assets, and only
then publishes it.

Verify downloaded artifacts with:

```bash
(cd download && sha256sum --check SHA256SUMS)
gh attestation verify download/prepare-commit-msg-linux-amd64 \
  --repo maccavelli/prepare-commit-msg
```

Repeat `gh attestation verify` for each binary consumed. The checksum proves
the downloaded byte set matches the release manifest; the attestation binds an
artifact to the repository, workflow, event, and source commit.

## Failed-release recovery

Validation and build failures occur before tag or release creation. Fix the
source problem, rerun `make verify`, and dispatch the same unused version.

A failure after draft creation can leave an unpublished draft and its tag.
Inspect both before cleanup:

```bash
VERSION=v1.2.3
gh release view "$VERSION" \
  --repo maccavelli/prepare-commit-msg \
  --json isDraft,tagName,targetCommitish,assets
git ls-remote --tags origin "refs/tags/$VERSION"
```

Never delete a published release through this recovery path. If and only if
`isDraft` is `true` and the draft is the failed workflow's incomplete output,
remove the draft and its generated tag, then redispatch:

```bash
gh release delete "$VERSION" \
  --repo maccavelli/prepare-commit-msg \
  --cleanup-tag \
  --yes
```

If the release was published, choose a new patch version; do not replace its
tag or assets.

## Dependency and action-pin updates

Dependabot groups Go-module and GitHub Actions changes weekly. Review grouped
updates as normal source changes and run `make verify`. For an action update,
confirm that the new full SHA belongs to the official action and that the
end-of-line release comment matches it. Never replace a SHA with a mutable
major tag.

This GitHub repository is a one-way mirror, so accepted dependency updates must
also land in the authoritative private source before the next export. Do not
merge a public pull request that the mirror workflow cannot propagate.

## Vulnerability failures and exceptions

`make vuln` uses the pinned `govulncheck` client with the current Go
vulnerability database. Prefer upgrading the affected toolchain or dependency.
Do not silence a finding in workflow YAML.

If remediation is temporarily impossible, record a separately reviewed
decision containing the vulnerability ID, reachable call path, affected
artifacts, compensating control, accountable owner, and an explicit expiration
date. Any change that bypasses `make vuln` requires its own accepted MADR and
implementation plan; this runbook grants no standing exception.

## Settings rollback

Rollback is an administrator action and must use the baseline captured before
apply. First preserve current state and compare it to the baseline:

```bash
scripts/audit-github.sh > /tmp/github-settings-before-rollback.json
SNAPSHOT=.git/prepare-commit-msg-github-settings-before.json
jq . "$SNAPSHOT"
```

For the audited pre-implementation baseline—Actions allowed, SHA enforcement
off, read-only token, no approval permission, Dependabot security updates off,
and no managed rulesets or release environment—the exact rollback is:

```bash
REPOSITORY=maccavelli/prepare-commit-msg

jq -n '{enabled:true, allowed_actions:"all", sha_pinning_required:false}' \
  >/tmp/actions-policy.json
gh api --method PUT "repos/$REPOSITORY/actions/permissions" \
  --input /tmp/actions-policy.json

jq -n '{default_workflow_permissions:"read",
  can_approve_pull_request_reviews:false}' >/tmp/workflow-permissions.json
gh api --method PUT "repos/$REPOSITORY/actions/permissions/workflow" \
  --input /tmp/workflow-permissions.json

gh api --method DELETE "repos/$REPOSITORY/automated-security-fixes"
gh api --method DELETE "repos/$REPOSITORY/environments/release"

for name in prepare-commit-msg-main prepare-commit-msg-release-tags; do
  id="$(gh api "repos/$REPOSITORY/rulesets" \
    --jq ".[] | select(.name == \"$name\") | .id")"
  if [ -n "$id" ]; then
    gh api --method DELETE "repos/$REPOSITORY/rulesets/$id"
  fi
done
```

Run the static rollback only when the saved snapshot confirms that exact
baseline. If a resource existed before activation, restore its complete saved
body instead of deleting it. Stop if duplicate managed ruleset names are found.
Finally run `scripts/audit-github.sh` and compare every field with the snapshot.
Local rollback is independent: use `make hooks-uninstall` for hooks and Git
revert commits for version-controlled CI/CD changes.

## References

- [GitHub Actions permission endpoints](https://docs.github.com/en/rest/actions/permissions)
- [GitHub deployment environment endpoints](https://docs.github.com/en/rest/deployments/environments)
- [GitHub deployment branch policy endpoints](https://docs.github.com/en/rest/deployments/branch-policies)
- [GitHub repository ruleset endpoints](https://docs.github.com/en/rest/repos/rules)
- [GitHub artifact attestation verification](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/verify-attestations)
- [Go vulnerability management](https://go.dev/doc/security/vuln/)
