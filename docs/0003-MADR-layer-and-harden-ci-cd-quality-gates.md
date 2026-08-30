---
status: accepted
date: 2026-08-30
decision-makers:
  - Maintainers of prepare-commit-msg
---

# Layer and Harden CI/CD Quality Gates

## Context and Problem Statement

The repository's CI and release workflows provide useful checks, but they do
not form a complete or consistently enforced quality boundary. A refactor in
commit `a5e796e` left an unused package variable and a `goimports` violation.
The feature branch had no GitHub Actions run or pull request, so the defects
were first detected after merge commit `ce73a98` reached `main`. The
[main CI run](https://github.com/maccavelli/prepare-commit-msg/actions/runs/33266225563)
then failed, and the same lint failure caused the
[`v1.1.0` release run](https://github.com/maccavelli/prepare-commit-msg/actions/runs/33323570738)
to skip its build and publication jobs. The `v1.1.0` tag remains present even
though no corresponding GitHub release was published.

This incident exposed gaps at four boundaries: developer feedback, pre-merge
or pre-export enforcement, GitHub Actions hardening, and release provenance.
The repository needs one quality contract that developers can run locally,
that local Git hooks can invoke without hiding staged-file errors, and that CI
and release automation enforce without duplicating commands.

### Repository and workflow facts

The following facts were verified on 2026-08-30:

* [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) runs only for pushes
  to `main` and pull requests. It does not run for an ordinary pushed feature
  branch. The repository's [mirror notice](../README.md#L1-L4) also says that
  GitHub pull requests are not the integration path, making the pull-request
  trigger insufficient as the only pre-merge trigger.
* CI runs tests, `go vet`, `golangci-lint`, and six cross-compilations on one
  `ubuntu-latest` runner. Cross-compilation confirms that platform-specific
  files compile, but it does not execute the Windows or macOS implementations.
* [`.github/workflows/release.yml`](../.github/workflows/release.yml) duplicates
  the test, vet, and lint sequence instead of calling a shared repository
  contract. Git history shows that the CI linter was pinned in commit `b08a9df`
  and the duplicated release linter required a separate correction in commit
  `d47c860`.
* The workflows have no job timeouts or concurrency policy. CI relies on the
  repository's default read-only `GITHUB_TOKEN`, while the release workflow
  grants `contents: write` at workflow scope, including to validation and build
  jobs that do not need write access.
* Every external action is referenced by a movable major-version tag. The
  repository currently allows all actions and does not require full commit-SHA
  pinning. The checkout steps do not explicitly disable persisted credentials.
* There is no repository-owned Dependabot configuration for Go modules or
  GitHub Actions. Dependabot security updates are disabled. Secret scanning and
  push protection are enabled, which should be retained.
* GitHub reports no branch protection, repository ruleset, protected release
  environment, or tag ruleset for this public repository. A direct update can
  therefore reach `main`, and a `v*` tag immediately starts release automation,
  without a server-enforced quality precondition.
* The release publishes SHA-256 checksums, which protect downloaded bytes after
  a user has obtained a trusted manifest. It does not publish a cryptographic
  build-provenance attestation binding those bytes to the repository, workflow,
  event, and source commit.

### Local quality-gate facts

* The [Makefile](../Makefile) exposes separate `fmt`, `vet`, `test`, `lint`,
  `test-coverage`, and `build-all` targets, but no single authoritative target
  equivalent to CI. The README consequently asks developers to remember four
  separate commands.
* The `fmt` target runs `go fmt`, while CI's configured formatter also enforces
  `goimports`. The `lint` target requires a separately installed binary and its
  error message recommends `@latest`, while CI pins version `v2.13.1`. This can
  make local and CI results depend on different tool versions.
* The Makefile declares an `80.0%` aggregate coverage minimum, but neither CI
  nor release automation invokes it. A read-only diagnostic run produced
  `79.0%` aggregate statement coverage, so activating the documented threshold
  today would fail. Package coverage ranged from `62.9%` in `internal/ui` to
  `96.5%` in `internal/git`.
* `go mod verify` and `go mod tidy -diff` pass, but neither is part of a local
  aggregate target or CI.
* The repository contains no version-controlled pre-commit hook or safe hook
  installer. This workstation uses a global `core.hooksPath`; its `pre-push`
  hook is a disclosure guard, not a code-quality gate. That guard must continue
  to run after any repository-specific hook integration.
* The external agent pre-commit fallback checks staged Go files with `gofmt` and
  the deprecated `golint`. Both failing files from the incident pass those two
  checks: `gofmt` does not enforce `goimports` grouping, and `golint` does not
  report the unused package variable. Workstation-specific automation is also
  absent for developers and automation that do not share that configuration.

### Security diagnostic facts

The pinned Go toolchain is `1.26.5`. A diagnostic `govulncheck ./...` run found
four reachable vulnerabilities through the release client's HTTP path:

* [`GO-2026-6218`](https://pkg.go.dev/vuln/GO-2026-6218) in `net/url`;
* [`GO-2026-6090`](https://pkg.go.dev/vuln/GO-2026-6090) in `crypto/tls`;
* [`GO-2026-5972`](https://pkg.go.dev/vuln/GO-2026-5972) in `encoding/asn1`;
* [`GO-2026-5026`](https://pkg.go.dev/vuln/GO-2026-5026) in `net/http` and
  `golang.org/x/net/idna`.

All four reports identify Go `1.26.6` as containing the relevant standard
library fixes. This is evidence that dependency and toolchain verification must
include reachability-aware vulnerability scanning rather than relying only on
compilation, tests, and static linting.

The authoritative private source repository and its server-side protections
were not available for this audit. The mirror notice is therefore treated as a
constraint, while any claims about private-host enforcement remain assumptions
to verify during implementation planning.

## Decision Drivers

* Detect formatting, lint, test, dependency, and build failures before code is
  integrated or exported.
* Give developers a fast default check and an exact local reproduction of the
  authoritative CI gate.
* Keep tool versions and commands deterministic across workstations, agents,
  CI, and release automation.
* Preserve the existing global disclosure guard and avoid silently overwriting
  a developer's `core.hooksPath` or installed hooks.
* Exercise supported operating-system behavior, not only cross-compilation.
* Apply least privilege and reduce mutable supply-chain inputs in GitHub
  Actions.
* Prevent a failed validation from leaving a release-shaped tag without a
  published, verifiable release.
* Preserve the one-way mirror model while enforcing equivalent checks in the
  authoritative source and independently revalidating exported code.
* Keep the solution understandable and maintainable for a small Go repository.

## Considered Options

* Layer repository-owned local gates, server-side CI, and hardened release
  automation
* Rely on comprehensive GitHub Actions only
* Rely primarily on local hooks and retain lightweight CI
* Adopt external hook and release frameworks as the quality contract
* Retain the current workflows with targeted fixes

## Decision Outcome

Chosen option: "Layer repository-owned local gates, server-side CI, and
hardened release automation", because no single boundary is sufficient: local
checks provide early feedback, server-side checks provide non-bypassable
enforcement, and a separately constrained release path protects published
binaries. A repository-owned command contract gives all three layers parity
without making a workstation-specific hook or GitHub YAML the source of truth.

### Repository-owned verification contract

The repository will define composable, non-interactive verification entrypoints
with these semantic tiers:

* A fast staged-change gate will inspect the staged snapshot, not merely the
  working tree. It will reject `gofmt` and `goimports` differences and run the
  repository's pinned lint rules over the relevant Go packages. Checking the
  index prevents unstaged edits from masking what will actually be committed.
* A full verification gate will run module tidiness and checksum verification,
  formatting/import checks, `go vet`, the pinned `golangci-lint` configuration,
  race-enabled tests, the aggregate coverage policy, `govulncheck`, and all
  supported cross-builds.
* A release verification gate will include the full gate and release-specific
  invariants: a clean exact source commit, valid semantic version, expected
  platform asset set, embedded version consistency, deterministic checksum
  manifest, and successful native smoke tests where applicable.

The full gate is the contract CI invokes. Individual Make targets may remain
for focused work, but they will compose into that gate instead of duplicating
its command list in YAML. Verification targets will check formatting without
rewriting files. Tool binaries will be version-pinned in repository-controlled
metadata or bootstrap logic; instructions and automation will not use
`@latest`.

The existing `80.0%` aggregate coverage policy will become enforced after tests
restore the current baseline to at least that value. The threshold will not be
silently lowered to make activation pass. Future decreases will fail unless a
reviewed decision intentionally changes the policy.

The `govulncheck` binary will be pinned, but its vulnerability database will be
current by design. A new advisory can therefore make an unchanged source tree
fail. That nondeterminism is an intentional security signal: reachable known
vulnerabilities must block releases and protected-branch integration until
fixed or covered by a documented, time-bounded exception.

### Local hooks and developer workflow

The repository will provide a version-controlled hook entrypoint and an
idempotent installer or documented dispatcher integration. It must detect an
existing `core.hooksPath` and existing hook files, compose with them, and fail
safely with instructions when composition cannot be proven. It must never
replace, disable, or reconfigure the existing disclosure `pre-push` guard
without explicit operator action.

The local policy will be:

* pre-commit runs the staged-change gate and is optimized for quick feedback;
* pre-push runs the full verification gate before source integration or public
  export, in addition to the existing disclosure checks;
* hooks remain bypassable emergency feedback mechanisms, while server-side
  rules remain authoritative;
* the README documents one setup command and one manual command that exactly
  reproduces CI.

Workstation-wide agent hooks may call the repository entrypoint when it exists,
but they are adapters rather than the quality policy itself.

### Continuous integration

CI will use the repository-owned commands and will:

* validate feature-branch pushes as well as pull requests and `main`, because
  this repository's documented integration model does not depend on GitHub pull
  requests;
* cancel superseded branch and pull-request runs while never cancelling an
  active release;
* set explicit job timeouts and explicit minimal permissions;
* use checkout without persisted credentials in jobs that do not push;
* run format, module, lint, vulnerability, race, coverage, and cross-build
  checks through the shared verification contract;
* execute tests natively on Linux, macOS, and Windows for platform-specific
  behavior, with the race detector in a supported primary job and
  cross-compilation retained for the complete six-target asset matrix;
* validate workflow and embedded shell syntax before relying on changed
  workflow files;
* expose stable, uniquely named required checks suitable for server-side
  rules; and
* use explicit supported runner images rather than mutable `*-latest` aliases
  where GitHub provides an appropriate stable label.

CI and release automation will share a reusable workflow or call the same
repository commands. The release workflow will not maintain its own copy of
test and lint steps.

### Repository and mirror enforcement

The authoritative source repository will require the full CI check before
integration into its protected default branch. The public GitHub mirror will
independently rerun the full gate on every exported snapshot because private
CI results and squashed export commits are not automatically trustworthy or
portable across repository boundaries.

The public repository will add rules appropriate to a one-way mirror:

* block force-pushes and deletion of `main`;
* restrict updates to the designated publication identity or application;
* restrict creation, update, and deletion of `v*` tags to the release path; and
* retain read-only default workflow permissions.

If the public repository later becomes a normal collaboration repository, the
rules will be tightened to require pull requests, review, and the stable CI
checks before merge. That topology change does not require changing the
verification contract.

### Workflow and release hardening

All reusable actions will be pinned to reviewed full commit SHAs. Repository
Actions policy will require SHA pinning and allow only GitHub-authored actions
plus explicitly reviewed exceptions. Go module and GitHub Actions references
will be reviewed and updated manually so immutability does not cause silent
staleness.

#### Maintainer amendment: Dependabot disabled

On 2026-08-30, after the initial implementation reached `main`, the maintainer
directed that Dependabot be turned off. This amendment supersedes the original
automation choice: the repository will have no Dependabot update configuration,
and the settings configurator will converge Dependabot security updates to
disabled. Vulnerability scanning remains mandatory in `make verify`; dependency
and action-pin updates remain a manual, reviewed maintenance responsibility.

Workflow permissions will default to `contents: read`. Only the job that
creates the release will receive `contents: write`; only the attestation step
will receive `id-token: write` and `attestations: write`. Third-party actions
with release credentials will be avoided when the GitHub CLI or API provides a
small, auditable equivalent. Any retained third-party action requires a pinned
SHA and explicit review.

Release initiation will move from "publish whenever any `v*` tag is pushed" to
a controlled release operation against an already verified `main` commit. The
operation will validate first, then create the version tag and release, so a
validation failure cannot create another orphan release tag. A protected
release environment may add approval without granting validation jobs write
access.

Release binaries and `SHA256SUMS` will continue to be produced from one clean
build job. In addition, the workflow will create GitHub artifact attestations
for the binaries and checksum manifest. Consumers will be able to verify both
the checksum and the GitHub/Sigstore provenance linking artifacts to the exact
repository, workflow, event, and commit. Publication must be atomic from the
user's perspective: an incomplete asset set remains a draft or is not exposed
as the latest release.

### Consequences

* Good, because the two defects from the `v1.1.0` incident will be caught by
  the staged gate before commit and by the same full gate before integration.
* Good, because local, CI, and release checks will use the same commands and
  pinned tools, eliminating the current linter and workflow drift.
* Good, because native platform tests cover behavior that cross-compilation
  cannot execute.
* Good, because least-privilege jobs, immutable action references, restricted
  publishers, vulnerability scanning, and attestations reduce compromise and
  release-tampering risk.
* Good, because validation occurs before tag creation, avoiding a repeat of the
  orphaned `v1.1.0` release tag.
* Neutral, because SHA-pinned actions and tools require routine automated
  update review rather than implicit upgrades.
* Neutral, because vulnerability results intentionally change as the Go
  database changes even when repository code does not.
* Bad, because the full gate will take longer than the present Linux-only job
  and will consume macOS and Windows runner capacity.
* Bad, because composing with global Git hooks is more complex than assuming an
  empty `.git/hooks` directory.
* Bad, because the existing coverage deficit and reachable Go vulnerabilities
  must be remediated before all proposed gates can become green.
* Bad, because protected source, mirror, and release identities require
  administrative configuration outside the Git repository and periodic audit.

### Confirmation

This decision is considered implemented only when all of the following are
demonstrated:

* A clean clone can install pinned development tools and run one documented
  full-verification command with the same result as CI.
* A staged fixture containing the incident's `goimports` error or unused
  package variable is rejected before commit, including when unstaged working
  tree changes are present.
* Existing global hook behavior still executes after local hook installation,
  and the installer refuses destructive hook replacement.
* Feature-branch, `main`, and release-candidate changes receive the intended
  stable CI checks; superseded branch runs cancel; release runs do not.
* Linux, macOS, and Windows tests pass natively, all six release targets build,
  aggregate coverage is at least `80.0%`, module files are tidy and verified,
  and `govulncheck` reports no reachable known vulnerabilities.
* Repository settings show read-only default workflow permissions, full-SHA
  action enforcement, active default-branch and tag rules, and a restricted
  release identity or environment.
* A deliberately failing release validation creates neither a new version tag
  nor a public release.
* A successful release contains exactly the expected binaries and checksum
  manifest, and `gh attestation verify` succeeds for every published artifact.

## Pros and Cons of the Options

### Layer repository-owned local gates, server-side CI, and hardened release automation

* Good, because it provides fast feedback and authoritative enforcement while
  sharing one repository-controlled quality contract.
* Good, because it accommodates both the private-source/public-mirror topology
  and a future pull-request workflow.
* Good, because release security is addressed separately from source quality.
* Bad, because it has the highest initial implementation and administration
  cost of the considered options.

### Rely on comprehensive GitHub Actions only

* Good, because GitHub is a consistent, centrally visible execution
  environment.
* Good, because no local hook installation or composition is required.
* Bad, because failures are discovered after a push or export, as occurred with
  `ce73a98`.
* Bad, because GitHub is documented as a mirror rather than the authoritative
  integration surface, and public pull requests cannot currently enforce the
  source repository's merge policy.

### Rely primarily on local hooks and retain lightweight CI

* Good, because developers receive failures early and runner usage stays low.
* Bad, because hooks are machine-specific and bypassable with `--no-verify`.
* Bad, because global `core.hooksPath` configurations and agent-only hooks make
  installation inconsistent.
* Bad, because local-only policy cannot safely hold release credentials or
  attest public artifacts.

### Adopt external hook and release frameworks as the quality contract

This option would make a framework such as `pre-commit` and a release tool such
as GoReleaser authoritative rather than using repository-native entrypoints.

* Good, because mature frameworks provide broad ecosystems and conventional
  setup flows.
* Good, because a release framework can reduce custom workflow scripting.
* Neutral, because these tools could still be adopted later as adapters to the
  repository-owned contract.
* Bad, because they introduce additional runtimes, configuration surfaces, and
  supply-chain dependencies for a small Go project.
* Bad, because a framework installer may conflict with the existing global
  hook path unless custom composition is still designed.

### Retain the current workflows with targeted fixes

* Good, because removing the unused variable and applying `goimports` is quick.
* Good, because the existing pipeline already runs race tests, vet, lint, and
  cross-builds.
* Bad, because it fixes the current symptoms without addressing how they
  bypassed local and branch checks.
* Bad, because duplicated workflows, broad release permissions, movable action
  tags, missing vulnerability scanning, and unprotected publication remain.

## More Information

* The accepted implementation sequence is defined by
  [0003-PLAN-layer-and-harden-ci-cd-quality-gates.md](0003-PLAN-layer-and-harden-ci-cd-quality-gates.md).
* The current CI failure is repaired in the first implementation phase before
  a green run is used as the hardened-pipeline baseline.
* GitHub recommends explicitly minimizing workflow permissions and pinning
  actions to full commit SHAs in its
  [secure use reference](https://docs.github.com/en/actions/reference/security/secure-use).
* GitHub documents rulesets and required checks in
  [Available rules for rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets).
* GitHub documents binary build provenance and the required permissions in
  [Using artifact attestations to establish provenance for builds](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations).
* The Go project describes `govulncheck` as a reachability-aware, low-noise
  scanner in [Go Vulnerability Management](https://go.dev/doc/security/vuln/).
