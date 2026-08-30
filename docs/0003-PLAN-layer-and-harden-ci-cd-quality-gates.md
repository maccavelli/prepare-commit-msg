---
status: completed
date: 2026-08-30
associated-madr: 0003-MADR-layer-and-harden-ci-cd-quality-gates.md
owner: Maintainers of prepare-commit-msg
---

# Implement Layered and Hardened CI/CD Quality Gates

Associated MADR:
[0003-MADR-layer-and-harden-ci-cd-quality-gates.md](0003-MADR-layer-and-harden-ci-cd-quality-gates.md)

## Post-Implementation Maintainer Directive

On 2026-08-30, after commit `f4fecb0` reached `origin/main`, the maintainer
directed that Dependabot be turned off. The follow-up removes
`.github/dependabot.yml`, changes repository-settings convergence from enabling
to disabling Dependabot security updates, and replaces automated update claims
with a manual review cadence. This directive supersedes the Dependabot-specific
items in the original completed phase record without changing its historical
account of what those phases implemented.

The first hosted native matrix also exposed Unix-only test assumptions on
Windows. The same follow-up makes config-path setup, filesystem error cases,
and executable-mode assertions portable without changing product behavior.

## Goal

Replace the repository's fragmented local checks, post-integration-only CI,
and tag-triggered release workflow with one repository-owned verification
contract used by staged-file checks, pre-push checks, CI, and release
validation. Harden GitHub Actions and release publication with deterministic
tool versions, native platform tests, least privilege, immutable action
references, dependency automation, controlled version creation, checksums, and
artifact attestations.

The completed implementation must catch the two defects that broke commit
`ce73a98` before commit, pass the existing `80.0%` coverage policy, report no
reachable known vulnerabilities, preserve the installed disclosure pre-push
hook, and make a failed release validation incapable of creating a version tag
or public release.

## Scope

### In scope

* Accepting the associated MADR and maintaining this executable plan.
* Repairing the current lint failure without changing intended CLI behavior.
* Updating the Go patch release and affected transitive modules to remediate
  the vulnerabilities recorded in the MADR.
* Adding tests until aggregate statement coverage is at least `80.0%` without
  reducing the existing threshold.
* Pinning development tools and exposing staged, full, and release
  verification entrypoints through the Makefile and repository scripts.
* Adding composable pre-commit and pre-push hooks plus installation,
  uninstallation, and isolated hook tests.
* Replacing duplicated CI/release validation with a reusable workflow,
  feature-branch CI, native platform tests, explicit timeouts, concurrency,
  minimal permissions, full-SHA action references, and workflow linting.
* Replacing tag-push publication with a manually dispatched, validated release
  workflow that creates a draft release and tag only after validation and then
  publishes the complete asset set.
* Adding GitHub Actions and Go module Dependabot configuration.
* Adding a deterministic, idempotent repository-settings script while
  deferring activation until the workflow commits exist on remote `main`.
* Updating developer and release documentation.
* Committing each completed phase locally. No phase pushes commits or tags.

### Out of scope

* Pushing commits, tags, or releases to any remote.
* Changing product behavior beyond removal of the obsolete password test seam
  and test-only additions required for coverage.
* Replacing the repository's existing disclosure policy or public-mirror
  topology.
* Enabling pull-request merging on the public mirror.
* Introducing GoReleaser, the Python `pre-commit` framework, containers, or a
  new application runtime dependency.
* Changing the `80.0%` coverage threshold.
* Publishing a replacement for the failed `v1.1.0` release.

## Fixed Inputs and Invariants

The implementation uses these reviewed values. Updating them is a separate,
reviewable dependency change:

| Input | Pinned value |
| --- | --- |
| Go toolchain | `1.26.6` |
| `golangci-lint` | `v2.13.1` |
| `govulncheck` | `v1.7.0` |
| `actionlint` | `v1.7.12` |
| `actions/checkout` | `3d3c42e5aac5ba805825da76410c181273ba90b1` (`v7`) |
| `actions/setup-go` | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (`v7`) |
| `actions/upload-artifact` | `043fb46d1a93c77aae656e7c1c64a875d1fc6a0a` |
| `actions/download-artifact` | `3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c` |
| `actions/attest` | `1e69f48acb82d1966a394da916b4c1698aa569d6` (`v4`) |
| Aggregate coverage minimum | `80.0%` |
| Release assets | Six OS/architecture binaries plus `SHA256SUMS` |

All repository scripts use Bash with `set -euo pipefail`, quote path values,
resolve the repository root through Git, use `mktemp -d` for temporary trees,
and remove temporary state with traps. Verification commands are
non-interactive. Formatting checks display a diff and do not rewrite files.

The public repository is `maccavelli/prepare-commit-msg`, its default branch is
`main`, and the currently authenticated maintainer is the release operator.
The public repository remains a mirror. The settings phase must not require
GitHub pull requests for public-mirror updates.

## Implementation Steps

### Phase 0: Accept the decision and establish the executable plan

#### Phase 0 affected files

* `docs/0003-MADR-layer-and-harden-ci-cd-quality-gates.md`
* `docs/0003-PLAN-layer-and-harden-ci-cd-quality-gates.md`

#### Phase 0 steps

1. Change the MADR frontmatter from `status: proposed` to `status: accepted`.
2. Link the MADR to this mechanically named plan and replace language that
   says implementation is not yet authorized.
3. Create this complete plan with `status: ready`, exact phase boundaries,
   files, commands, acceptance criteria, rollout, and rollback.
4. Run Markdown lint and whitespace validation.
5. Commit only the MADR and PLAN with message
   `docs(cicd): accept hardening decision and implementation plan`.

#### Phase 0 verification

```bash
markdownlint-cli2 \
  docs/0003-MADR-layer-and-harden-ci-cd-quality-gates.md \
  docs/0003-PLAN-layer-and-harden-ci-cd-quality-gates.md
git diff --check
git diff --cached --name-only
```

#### Phase 0 acceptance criteria

* MADR status is `accepted`; PLAN status is `ready`.
* The filenames share identifier `0003` and the complete slug.
* The plan links to the MADR and the MADR links to the plan.
* Only the two documentation artifacts are committed in this phase.

### Phase 1: Restore a secure green application baseline

#### Phase 1 affected files

* `go.mod`
* `go.sum`
* `Makefile`
* `internal/ui/setup.go`
* `internal/ui/setup_test.go`
* `internal/selfupdate/client_test.go`

#### Phase 1 steps

1. Remove the obsolete `readPassword` package variable and its now-unused
   `golang.org/x/term` import from `internal/ui/setup.go`.
2. Apply `goimports` ordering to `internal/ui/setup_test.go`.
3. Change the Go directive and Makefile toolchain declaration from `1.26.5` to
   `1.26.6`.
4. Upgrade `golang.org/x/net` to at least `v0.55.0` if module selection still
   resolves an affected version, then run `go mod tidy` and `go mod verify`.
5. Use the coverage profile to add deterministic unit cases in
   `internal/ui/setup_test.go` and `internal/selfupdate/client_test.go` for
   currently uncovered error and boundary paths. Tests must use temporary
   directories and local `httptest` servers; they must not use live provider
   or GitHub endpoints.
6. Do not reduce the coverage threshold or add coverage exclusions.
7. Run all pre-commit checks on every changed Go file before committing.
8. Commit with message `fix(cicd): restore secure green quality baseline`.

#### Phase 1 verification

```bash
gofmt -d internal/ui/setup.go internal/ui/setup_test.go \
  internal/selfupdate/client_test.go
goimports -d internal/ui/setup.go internal/ui/setup_test.go \
  internal/selfupdate/client_test.go
golint internal/ui/setup.go
golint internal/ui/setup_test.go
golint internal/selfupdate/client_test.go
go mod tidy -diff
go mod verify
go test -race ./...
go test -coverprofile=/tmp/prepare-commit-msg-coverage.out ./...
go tool cover -func=/tmp/prepare-commit-msg-coverage.out
golangci-lint run -c .golangci.yml ./...
govulncheck ./...
make build-all
git diff --check
```

#### Phase 1 acceptance criteria

* The two original lint errors are absent.
* All tests and lint checks pass with Go `1.26.6` and `golangci-lint v2.13.1`.
* Aggregate statement coverage is at least `80.0%`.
* `govulncheck ./...` reports no reachable known vulnerabilities.
* All six target binaries compile.
* No intended runtime behavior changes.

### Phase 2: Add the repository-owned verification contract

#### Phase 2 affected files

* `.gitignore`
* `Makefile`
* `scripts/bootstrap-tools.sh`
* `scripts/go-precheck.sh`
* `scripts/verify-release.sh`
* `scripts/verify-scripts.sh`

#### Phase 2 steps

1. Ignore `.tools/`, the repository-local pinned tool installation directory.
2. Add pinned version variables and repository-local binary paths to the
   Makefile. `make tools` installs exactly the versions in **Fixed Inputs and
   Invariants** using `GOBIN=<repo>/.tools/bin`.
3. Replace the mutating-only formatting workflow with separate `fmt` and
   `fmt-check` targets. `fmt-check` uses `gofmt` plus
   `golangci-lint fmt --diff` with `.golangci.yml`.
4. Add `mod-check`, `vuln`, `workflow-lint`, `coverage`, `verify`,
   `verify-staged`, and `verify-release` targets. Keep existing target names as
   compatible aliases where behavior is equivalent.
5. Define `make verify` as the single full contract: pinned tools, module
   checks, format/import checks, lint, vet, race tests, coverage, vulnerability
   scan, script/workflow validation, and all six cross-builds.
6. Implement `scripts/go-precheck.sh` so it accepts the changed filenames used
   by the external agent gate, exports the complete staged index into a clean
   temporary tree with `git checkout-index`, and runs format/import and lint
   checks there. Deleted files are ignored. The working tree must not mask the
   staged snapshot.
7. Implement `scripts/verify-release.sh VERSION DIST_DIR` to validate a strict
   `vMAJOR.MINOR.PATCH` version, clean source tree, exact seven-file asset set,
   executable names, embedded version output for the native binary when
   runnable, and a deterministic checksum manifest that excludes itself.
8. Implement `scripts/verify-scripts.sh` to run `bash -n` on every tracked shell
   script and the pinned `actionlint` over all workflow files.
9. Test that the staged gate rejects temporary fixtures matching both original
   CI defects, then restore the index and working tree without touching user
   changes.
10. Commit with message `build(cicd): add deterministic verification contract`.

#### Phase 2 verification

```bash
make tools
make fmt-check
make mod-check
make lint
make vet
make test
make coverage
make vuln
make workflow-lint
make build-all
make verify
make verify-release VERSION=v1.1.1
git diff --check
```

#### Phase 2 acceptance criteria

* A clean clone can obtain exact tool versions through `make tools`.
* `make verify` is green and includes every promised full-gate check.
* `make verify-staged` analyzes the index snapshot and catches both incident
  defect classes.
* `make verify-release` rejects invalid versions, dirty trees, missing assets,
  extra assets, and self-referential checksum manifests.
* Existing Make target users retain compatible focused commands.

### Phase 3: Install composable local Git hooks

#### Phase 3 affected files

* `.githooks/pre-commit`
* `.githooks/pre-push`
* `scripts/install-hooks.sh`
* `scripts/uninstall-hooks.sh`
* `scripts/test-hooks.sh`
* `README.md`

#### Phase 3 steps

1. Add repository hook entrypoints: pre-commit invokes `make verify-staged`;
   pre-push invokes `make verify`.
2. Implement an installer that resolves the effective pre-existing hooks path,
   creates a repository-local managed directory under `.git`, writes wrappers
   that invoke any pre-existing hook first, and then invokes the repository
   hook. Pre-push wrappers must preserve arguments and replay stdin to both
   hooks through a securely created temporary file.
3. Record whether a local `core.hooksPath` existed and its exact value before
   installation. Set a local hooks path only after all wrappers are complete.
4. Refuse installation if the target is outside this repository's `.git`
   directory, a managed file is unexpectedly modified, a recursive hook chain
   is detected, or a prior hook cannot be resolved safely.
5. Implement idempotent reinstallation and an uninstaller that restores the
   prior local configuration or unsets the local key so the global value is
   inherited again. Never delete or edit the pre-existing hook.
6. Add isolated shell tests using temporary Git repositories and fake hooks to
   prove invocation order, stdin replay, failure propagation, idempotence,
   refusal behavior, and exact restoration.
7. Document `make hooks-install`, `make hooks-test`, `make hooks-uninstall`, and
   `make verify` in the README.
8. Run the installer in this checkout. Confirm that the existing global
   disclosure `pre-push` still runs before the repository quality gate.
9. Commit with message `dev(cicd): add composable local quality hooks`.

#### Phase 3 verification

```bash
bash -n .githooks/pre-commit .githooks/pre-push scripts/install-hooks.sh \
  scripts/uninstall-hooks.sh scripts/test-hooks.sh
make hooks-test
make hooks-install
make hooks-install
git config --show-origin --get core.hooksPath
make verify-staged
make verify
make hooks-uninstall
git config --show-origin --get core.hooksPath
make hooks-install
git diff --check
```

#### Phase 3 acceptance criteria

* Hook tests pass in isolated repositories.
* Installation is idempotent and uninstallation restores the previous state.
* The effective hook path for this checkout is repository-managed after final
  installation.
* The original global disclosure hook remains byte-for-byte unchanged and is
  invoked by the managed pre-push wrapper.
* A failure in either the prior hook or repository gate blocks the operation.

### Phase 4: Replace and harden GitHub Actions workflows

#### Phase 4 affected files

* `.github/workflows/quality.yml`
* `.github/workflows/ci.yml`
* `.github/workflows/release.yml`
* `.github/dependabot.yml`
* `README.md`

#### Phase 4 steps

1. Add `quality.yml` as a reusable `workflow_call` workflow with top-level
   `contents: read`, explicit runner labels, job timeouts, and no write token.
2. Add a Linux `Quality Gate` job that checks out without persisted
   credentials, sets up Go from `go.mod`, caches Go and `.tools` inputs, and
   runs `make verify`.
3. Add a `Native Tests` matrix for explicit Linux, macOS, and Windows runner
   labels. Run `go test ./...` natively on every OS; keep race and coverage in
   the Linux quality job. Disable matrix fail-fast so all platform results are
   visible.
4. Reduce `ci.yml` to triggers, concurrency, read permissions, and a call to
   the reusable quality workflow. Trigger all branch pushes, pull requests,
   and manual dispatch. Cancel superseded non-release runs by ref.
5. Replace `release.yml`'s `v*` push trigger with `workflow_dispatch` inputs for
   a strict version and optional prerelease flag. Permit dispatch only from
   `main` and use a non-cancelling release concurrency group.
6. Call the reusable quality workflow before any release job. Build exactly
   once from `github.sha`, pass the requested version explicitly to
   `make build-all`, create `SHA256SUMS`, and run `make verify-release`.
7. Upload the complete `dist/` directory with a pinned official action. Attest
   all seven files in the build job with only `contents: read`,
   `id-token: write`, and `attestations: write`.
8. Give only the publish job `contents: write` and the `release` environment.
   Download the artifact, revalidate its exact contents and checksums without a
   source checkout, refuse an existing tag or release, create a draft release
   targeting the validated SHA with `gh release create`, upload all seven
   assets, and publish only after the complete draft is verified.
9. Pin every external action to the full SHA in **Fixed Inputs and Invariants**
   and retain its release tag in an end-of-line comment for Dependabot.
10. Add weekly Dependabot groups for `gomod` and `github-actions`; cap open pull
    requests and use conventional `chore(deps)` commit prefixes.
11. Update README release instructions to use the controlled dispatch and to
    document checksum plus `gh attestation verify` verification.
12. Commit with message `ci: harden verification and release workflows`.

#### Phase 4 verification

```bash
make workflow-lint
make verify
rg -n 'uses: [^#]+@(v|main|master|latest)' .github/workflows \
  .github/dependabot.yml
rg -n 'permissions:|persist-credentials|timeout-minutes|concurrency:' \
  .github/workflows
git diff --check
```

#### Phase 4 acceptance criteria

* `actionlint` and all local gates pass.
* CI runs on feature branches, pull requests, `main`, and manual dispatch.
* CI and release use the same reusable quality workflow and `make verify`.
* All action references are full SHAs; checkout credentials are not persisted.
* Validation and native-test jobs have read-only permissions and timeouts.
* Only publish has `contents: write`; only build attestation has OIDC and
  attestation write permissions.
* A validation/build failure occurs before any tag or release creation.
* Publication is draft-first and exposes exactly seven verified assets.

### Phase 5: Prepare and safely activate repository enforcement

#### Phase 5 affected files

* `scripts/configure-github.sh`
* `scripts/audit-github.sh`
* `docs/cicd-operations.md`
* `README.md`

#### Phase 5 steps

1. Add a read-only audit script that emits normalized JSON for repository
   Actions policy, default token permissions, `main` protection/rulesets,
   `v*` tag rulesets, release environments, Dependabot security updates,
   secret scanning, and push protection.
2. Add an idempotent configuration script with dry-run default and an explicit
   `--apply` mode. It must verify `gh` authentication, repository identity,
   admin access, default branch, and the presence of the hardened workflow SHA
   on the remote before changing settings.
3. Configure read-only default workflow permissions, disallow Actions from
   approving pull requests, require full-SHA action references, and restrict
   allowed actions to GitHub-authored actions plus repository-local workflows.
4. Enable Dependabot security updates while retaining secret scanning and push
   protection.
5. Create or update a `release` environment bound to the default branch. Do
   not invent reviewer identities; document optional reviewer configuration as
   a separately auditable administrative enhancement.
6. Create or update a `main` ruleset that blocks deletion and non-fast-forward
   updates. Because this is a one-way mirror, do not require public GitHub pull
   requests. Preserve administrator publication access.
7. Create or update a `v*` tag ruleset that blocks deletion and
   non-fast-forward updates and restricts creation to administrators and the
   GitHub Actions integration used by the controlled release workflow.
8. Document the release dispatch, failure recovery for an unpublished draft,
   settings audit, action-pin update review, Dependabot handling, vulnerability
   exceptions, and rollback commands.
9. Run the audit and configuration script in dry-run mode. Do not run `--apply`
   until the hardened workflow commit exists on `origin/main`; this is
   impossible without a push, which this plan explicitly excludes.
10. Commit with message `ops(cicd): codify repository enforcement and runbook`.

#### Phase 5 verification

```bash
bash -n scripts/configure-github.sh scripts/audit-github.sh
scripts/audit-github.sh
scripts/configure-github.sh
make verify
markdownlint-cli2 docs/cicd-operations.md README.md
git diff --check
```

After a maintainer separately pushes the implementation and authorizes remote
activation in a later turn:

```bash
scripts/configure-github.sh --apply
scripts/audit-github.sh
```

#### Phase 5 acceptance criteria

* Audit output is deterministic, contains no token material, and distinguishes
  absent settings from API failures.
* Dry-run output lists exact intended mutations and makes none.
* `--apply` refuses to run until the hardened commit is present on
  `origin/main`.
* Repeated `--apply` runs converge without duplicate rulesets or environments.
* The runbook includes exact recovery and rollback procedures.
* Remote activation remains explicitly pending because this implementation
  turn does not authorize `git push`.

### Phase 6: Final verification and implementation record

#### Phase 6 affected files

* `docs/0003-PLAN-layer-and-harden-ci-cd-quality-gates.md`

#### Phase 6 steps

1. Start from a clean index after Phase 5 and run the complete verification
   suite twice to detect generated-file or ordering instability.
2. Confirm the original global disclosure hook checksum matches the value
   recorded before hook installation.
3. Confirm no tracked file contains a floating action reference, `@latest`
   tool installation, stale Go `1.26.5` declaration, or duplicated release
   validation command list.
4. Confirm the local commit history contains one commit per completed phase and
   that no remote ref changed.
5. Update this plan from `status: ready` to `status: completed`, append the
   actual commit IDs and verification results, and explicitly record remote
   repository-setting activation as a post-push operational task rather than a
   completed claim.
6. Commit with message `docs(cicd): record hardening implementation results`.

#### Phase 6 verification

```bash
make verify
make verify
make hooks-test
scripts/audit-github.sh
rg -n '@(latest|main|master|v[0-9]+)([[:space:]#]|$)' \
  .github/workflows scripts Makefile
rg -n '1\.26\.5' --glob '!docs/0003-*'
git log --oneline --decorate -8
git status --short --branch
```

#### Phase 6 acceptance criteria

* The full gate passes twice from an unchanged tree.
* Hook composition and isolated hook tests pass.
* Local phase commits are present and no push occurred.
* The plan truthfully distinguishes completed repository work from pending
  remote activation.

## Verification

The authoritative local acceptance command is:

```bash
make verify
```

The implementation is accepted locally when that command passes twice, the
phase-specific criteria above pass, and the only remaining work is the
explicitly deferred post-push repository-setting activation. A GitHub-hosted
run is not claimed until the local commits are pushed by the maintainer.

The implementation is accepted end to end in a later authorized rollout when:

1. The pushed feature-branch CI run is green on all native runners.
2. The hardened commit is integrated into the authoritative source and
   exported to public `main`.
3. `scripts/configure-github.sh --apply` succeeds and the audit matches the
   desired state.
4. A deliberately invalid release dispatch fails before tag creation.
5. A valid release candidate produces seven assets whose checksums and
   attestations verify, without using `v1.1.0` as the test version.

## Rollout and Rollback

### Rollout

1. Complete and commit Phases 0 through 6 locally.
2. Review the commit series and MADR/PLAN consistency.
3. In a later turn with explicit push authorization, push the implementation
   branch and observe feature-branch CI.
4. Integrate through the authoritative private source, then export the verified
   snapshot to public `main`.
5. Run the repository-settings script in dry-run mode again, authorize
   `--apply`, and audit the result.
6. Use a new patch version for the first controlled release and verify every
   checksum and attestation before announcing it.

### Rollback

* Code and workflow phases are separate commits and can be reverted in reverse
  order without rewriting history.
* `scripts/uninstall-hooks.sh` restores the exact prior local hooks-path state;
  the global disclosure hook is never modified.
* Before applying GitHub settings, the configuration script records current
  normalized settings. The runbook maps each mutation to its inverse API call.
* If hardened CI is too slow, revert the workflow orchestration commit while
  retaining the local verification contract and security baseline; do not
  weaken coverage or vulnerability policy silently.
* If release publication fails after draft creation, keep the release draft,
  diagnose it, and delete it only through the documented explicit recovery
  command. Never move an existing version tag to different source.
* If a ruleset blocks the mirror publisher, use the documented administrator
  bypass to restore the captured prior ruleset, then correct the actor policy
  in a reviewed change.

## Phase Commit Ledger

The commit IDs and final results were populated only after each phase actually
completed. A Git commit cannot contain its own eventual object ID, so the
Phase 6 row identifies the commit containing this ledger as `this commit`.

| Phase | Commit | Result |
| --- | --- | --- |
| 0 | `85e350d` | Accepted MADR and ready implementation plan |
| 1 | `faea9f3` | Secure green application baseline |
| 2 | `7a3ce27` | Deterministic repository verification contract |
| 3 | `ba717de` | Composable local hooks |
| 4 | `c511501` | Hardened GitHub workflows |
| 5 | `8aad997` | Enforcement scripts and operations runbook |
| 6 | This commit | Completed plan and final verification record |

## Implementation Results

Local implementation completed on 2026-08-30 with these observed results:

* The original CI failures were repaired: the unused `readPassword` seam was
  removed and `internal/ui/setup_test.go` now satisfies `goimports`.
* Go was raised to `1.26.6`. The final `govulncheck ./...` runs reported no
  reachable vulnerabilities.
* Aggregate statement coverage is `82.2%`, above the unchanged `80.0%`
  threshold. The full gate reported zero `golangci-lint` issues and completed
  all six cross-builds.
* `make verify-staged` was exercised against synthetic staged `goimports` and
  unused-variable defects and rejected both without allowing unstaged content
  to mask the index.
* Release validation accepted a complete `v1.1.1` test asset set and rejected
  a missing asset. It enforces strict version syntax, six binaries, a
  six-entry non-self-referential manifest, checksums, and embedded native
  version consistency.
* Hook installation, repeated installation, failure propagation, pre-push
  stdin replay, and exact uninstallation restoration passed in isolated test
  repositories. The final effective hooks path is repository-managed. The
  original global disclosure `pre-push` SHA-256 remains
  `8522a697596001a2d4778e8e878de1647d43a863f0d4c5e5a98093101717cff6`.
* `actionlint` accepts the reusable quality, CI, and controlled release
  workflows. All external action references are full commit SHAs, and a scan
  found no `@latest`, mutable branch, or mutable major-tag references in the
  workflows, scripts, or Makefile.
* Two consecutive `make verify` runs completed from and returned to an
  unchanged tracked tree. `make hooks-test`, the deterministic settings audit,
  Markdown lint, shell syntax checks, and `git diff --check` also passed.
* The live read-only audit still reports Actions allowed without enforced SHA
  pinning, no `main` protection or rulesets, no `release` environment, and
  Dependabot security updates disabled. Secret scanning and push protection
  remain enabled.
* `scripts/configure-github.sh` produced the same normalized eight-mutation
  dry-run twice and reported `remote_ready: false`. It made no mutation.
  `origin/main` remained
  `ce73a982163b250231dc6af22f2fc851ca032f5a` before and after final
  verification, and no push, tag, release, or settings activation occurred.

Repository implementation is therefore complete within the approved local
scope. GitHub-hosted native runs, remote ruleset and environment activation,
invalid-dispatch testing, and a valid controlled release remain the explicitly
deferred rollout steps listed above; they require a later push and separate
authorization.
