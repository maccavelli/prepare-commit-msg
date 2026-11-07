# CI/CD Operations Runbook

This runbook implements
[`0003-MADR-layer-and-harden-ci-cd-quality-gates.md`](../decisions/0003-MADR-layer-and-harden-ci-cd-quality-gates.md)
and
[`0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md`](../decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md).
The repository-owned commands are the source of truth; CI invokes those
commands instead of maintaining separate check lists. The repository installs
no Git hooks: the global `prepare-commit-msg` hook and the disclosure guard
run from the global `core.hooksPath`.

## Routine verification

Before pushing, run every CI gate that runs off GitHub:

```bash
make preflight
```

It runs `make check-records`, `make markdownlint` and `make verify`, in that
order, printing `==> <gate>` before each, stopping at the first failure, and
ending with `preflight passed`. CI's `go` job runs the same three targets,
plus `make disclosure-check` after Markdown and before the contract. A
Markdown-only change under `ci.yml`'s `paths-ignore` starts only the Docs
workflow, which runs the records, Markdown and disclosure gates; a change to
anything else, `README.md` and this guide included, starts the full CI, and
so does every tag, because GitHub does not evaluate path filters for a tag
push
([`0022-MADR-align-docs-ci-and-disclosure-gates.md`](../decisions/0022-MADR-align-docs-ci-and-disclosure-gates.md)).
`make preflight` still runs all three gates locally either way.
`make verify` checks module tidiness and checksums, formatting and imports,
static analysis, race-enabled tests, the 80% aggregate coverage threshold,
reachable vulnerabilities, shell and workflow syntax, the maintainer tools'
tests, and all six release cross-builds.

When a change can behave differently per OS (Go sources, `go.mod`/`go.sum`,
build tags, paths, `os/exec`, permissions, the `Makefile`, scripts or
workflows), also run CI's jobs on the test hosts:

```bash
make preflight-hosts
```

It tests the committed `HEAD`, so commit first. Linux hosts run `make verify`;
Windows and this machine run `go test ./...`. The hosts are named in a config
outside the repository; [`tools/README.md`](../../tools/README.md) gives its
path and shape. A host that cannot be reached fails the run; say so when
handing off, rather than pushing past it. Changes to docs and records only are
exempt.

`make pre-add-check` checks the exact staged Go snapshot (or `FILES=...`)
with `scripts/go-precheck.py`; `make verify-staged` is its older name.

### A clone that still has the old hooks

Until 0017-MADR, `make hooks-install` pointed a clone's local
`core.hooksPath` at `.git/prepare-commit-msg-hooks`, whose `pre-push` ran
`.githooks/pre-push`. That file is gone, so such a clone fails every push.
Neither `make hooks-install` nor `make hooks-uninstall` exists any more, so
restore such a clone by hand. First see whether the clone had its own setting
before:

```bash
cat .git/prepare-commit-msg-hooks/.previous-local-present
```

If it prints `no`, remove the local setting:

```bash
git config --local --unset core.hooksPath
```

If it prints `yes`, restore the saved value instead:

```bash
git config --local core.hooksPath "$(cat .git/prepare-commit-msg-hooks/.previous-local-value)"
```

Then remove the managed directory and check where Git now looks; it should
print the global hooks directory:

```bash
rm -rf .git/prepare-commit-msg-hooks
git rev-parse --path-format=absolute --git-path hooks
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
- GitHub-owned actions, plus go-selfupdate-lib's reusable
  `build-selfupdate-release.yml` and `publish-selfupdate-release.yml`
  workflows, all pinned by full SHA;
- Dependabot security updates disabled by maintainer policy; and
- a default-branch ruleset that blocks deletion and non-fast-forward
  updates, with no bypass actor, so it binds every actor, the owner
  included.

No `v*` tag ruleset is created. Immutable releases are on, a repository
setting the script does not manage. They lock each published release's tag:
the tag cannot be moved, and cannot be deleted while its release exists.
Creating a `v*` tag is not restricted. The owner is the only account with
write access, and a tag pushed with a workflow's token starts no release. If
another account is ever given write access, restricting tag creation, as
go-selfupdate-lib's release guide advises, needs a new decision. The script
manages no deployment environment
([`0025-MADR-remove-ruleset-bypass-actors.md`](../decisions/0025-MADR-remove-ruleset-bypass-actors.md) D1, Q1, A1).

The apply path verifies the GitHub host, repository identity, administrator
access, default branch and origin URL, and that the local `ci.yml` is the
one on `origin/main` with no uncommitted workflow changes, before its first
mutation. It therefore refuses activation until `ci.yml` has been pushed to
`origin/main`.

Pushing is a separate, explicit maintainer action. After the implementation is
on `origin/main` and remote activation has been separately authorized:

```bash
scripts/configure-github.sh --apply
scripts/audit-github.sh
```

The first apply preserves a mode-`0600` baseline at
`.git/prepare-commit-msg-github-settings-before.json`; later applies preserve
that original snapshot. Named rulesets are updated in place, and repeated
runs converge without duplicates.

## Controlled release

Releases are tag-driven through
[`.github/workflows/ci.yml`](../../.github/workflows/ci.yml).
The workflow runs on pushes to `main`, pull requests, `workflow_dispatch`,
and `v*` tags. Only a `v*` tag publishes a GitHub Release. Every other ref
rehearses the build, stamped `rehearsal-<commit> (local)`, and publishes
nothing.

1. Ensure `main` is up to date and passes verification locally (`make verify`).

   When `main`'s commits since its last full CI run change only ignored
   Markdown, that run tested the same code: the docs workflow checked them,
   and the tag's unfiltered run is the release's proof.

2. Create and push an annotated semantic version tag:

```bash
git tag -a v1.2.3 -m "v1.2.3"
git push origin v1.2.3
```

The tag run then:

- runs the quality contract (`make verify`), which includes the six
  cross-builds;
- runs native `go test ./...` on Linux, macOS, and Windows;
- calls the pinned build workflow
  `maccavelli/go-selfupdate-lib/.github/workflows/build-selfupdate-release.yml@74ab5dd376a4c864debb32b2d8abfc7856059ba3 # v1.24.0` with
  `spec-path: selfupdate-release.json`, checks build identity, and writes
  `SHA256SUMS`;
- calls the pinned
  `maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml@74ab5dd376a4c864debb32b2d8abfc7856059ba3 # v1.24.0` only when
  `needs.build.outputs.rehearsal` is `'false'`.

Published assets are the six binaries, `SHA256SUMS`, `install.sh`,
`install.ps1`, and build-provenance attestations. The identity check and
the attestation verify command stay in the
[root README](../../README.md#maintainer-release).

Monitor the run with:

```bash
gh run list --workflow ci.yml --limit 1
```

Verify a downloaded checksum file with:

```bash
(cd download && sha256sum --check SHA256SUMS)
```

After the release is verified, prune the releases and tags below the eight
newest stable releases
([`0026-MADR-keep-five-stable-releases.md`](../decisions/0026-MADR-keep-five-stable-releases.md)):

```bash
python3 tools/prune_releases.py            # read the plan; exit 1 means work to do
python3 tools/prune_releases.py --apply
git fetch --prune --prune-tags origin
```

A deleted release cannot be restored, and its tag name cannot be used
again. `update` and the installers' default path follow the latest release,
which is always kept. Other clones drop the deleted tags with the same
fetch. [`tools/README.md`](../../tools/README.md#pruning-releases) lists what
the tool reports instead of deleting, and its refusals.

## Failed-release recovery

If a release tag's run fails before it publishes a release:

1. Fix the underlying issue on `main` and run `make verify`.
2. Check whether the Go module proxy already lists the version:

   ```bash
   curl -fsS https://proxy.golang.org/github.com/maccavelli/prepare-commit-msg/@v/list
   ```

   Use `@v/list` only. Asking the proxy for one version (`@v/$VERSION.info`)
   can make it fetch and record that version.
3. If the version is not listed, delete the remote and local failed tag, and
   tag the fixed commit with the same version:

   ```bash
   git push --delete origin "$VERSION"
   git tag -d "$VERSION"
   git tag -a "$VERSION" -m "$VERSION"
   git push origin "$VERSION"
   ```

4. If the version is listed, leave the tag. The checksum database already
   holds its content, so the same version on another commit would fail
   `go install` checks. Tag the fix as the next patch version (e.g.
   `v1.2.4`) instead.

If the release was published, its tag stays: immutable releases lock it,
and its name cannot be reused even if the release is deleted. Tag the fix
as the next patch version
([`0025-MADR-remove-ruleset-bypass-actors.md`](../decisions/0025-MADR-remove-ruleset-bypass-actors.md) A1-3).

## Dependency and action-pin updates

Dependabot version updates and security updates are disabled by maintainer
policy. Review Go module and GitHub Actions updates manually on a regular
cadence and run `make verify`. For an action update, confirm that the new full
SHA belongs to the official action and that the end-of-line release comment
matches it. Never replace a SHA with a mutable major tag.

This GitHub repository is the source of truth: accepted dependency updates
land on its `main`, and releases are tagged there.

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
Local rollback is independent: Git revert commits for version-controlled
CI/CD changes.

## References

- [GitHub Actions permission endpoints](https://docs.github.com/en/rest/actions/permissions)
- [GitHub deployment environment endpoints](https://docs.github.com/en/rest/deployments/environments)
- [GitHub deployment branch policy endpoints](https://docs.github.com/en/rest/deployments/branch-policies)
- [GitHub repository ruleset endpoints](https://docs.github.com/en/rest/repos/rules)
- [GitHub artifact attestation verification](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/verify-attestations)
- [Go vulnerability management](https://go.dev/doc/security/vuln/)
