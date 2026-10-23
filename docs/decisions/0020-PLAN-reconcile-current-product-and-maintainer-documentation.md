---
status: in-progress
date: 2026-10-23
associated-madr: "0020-MADR-reconcile-current-product-and-maintainer-documentation.md"
---
# Implement reconciliation of current product and maintainer documentation

Associated MADR: [0020-MADR-reconcile-current-product-and-maintainer-documentation.md](0020-MADR-reconcile-current-product-and-maintainer-documentation.md)

## Goal

Make all current-facing product and maintainer documentation factually match
the repository at implementation time, while preserving historical records.
Keep the README's public installation commands executable and add focused,
offline checks that prevent their documented shape and current CI facts from
silently drifting.

## Scope

### In scope

* `README.md`: update release wording after the Go 1.27.3 and dependency
  upgrade; retain the exact Unix and PowerShell one-line install commands and
  explain their scope accurately.
* `docs/README.md`: retain the completed 0019 and 0020 index rows and make
  their statuses agree with record metadata.
* `docs/architecture.md`: reconcile the current-system description with the
  implemented hook, configuration, updater, verification, and release
  boundaries, only where source evidence requires a change.
* `docs/guides/cicd-operations.md`: replace stale reusable-workflow
  identifiers and reconcile the current verification and release procedures
  with `.github/workflows/ci.yml` and the Makefile.
* A focused Go test in the root package (or a clearly named extension of
  `readme_test.go`) that reads only tracked documentation and workflow files
  and checks the documented install commands and release-workflow references.
* A non-destructive Unix live installer smoke test against the latest
  published release, performed only in a fresh temporary directory outside
  the repository.
* A Windows-host PowerShell installer smoke test using a fresh temporary
  install directory; record its exact result in this PLAN's execution record.

### Out of scope

* Changing hook behavior, provider catalog entries, installer templates,
  release assets, module versions, CI policy, or release publication.
* Rewriting historical MADRs, PLANs, or reports to make their contemporary
  facts look current.
* Running an update against an installed user hook or changing global Git,
  PATH, credential, or user configuration.
* Publishing commits, tags, releases, or documentation externally.

## Preconditions

* The owner has approved execution of this PLAN or one of its numbered
  phases.
* Before each phase, inspect `git status --short`; preserve unrelated changes.
* For any documentation fact derived from a release or workflow, obtain it
  from the current committed source or the published release being tested,
  rather than copying a value from an earlier record.
* Before claiming the Windows command is live-verified, obtain a Windows host
  through the repository's configured cross-host facility and use its native
  PowerShell. If no host is available, record that fact and leave the PLAN
  in progress rather than asserting success.

## Implementation Steps

### Phase 1: Add drift-detection tests and reconcile documented facts

1. Inspect the existing root-package test conventions, especially
   `readme_test.go`, and the exact `uses:` lines in `.github/workflows/ci.yml`.
   Add a focused test file or extend `readme_test.go`; do not introduce a new
   dependency.
2. Correct the known current-facing documentation drift before the new
   repository-facing assertions run: identify `v1.10.1` as released in the
   README; synchronize the operations guide's reusable-workflow references
   with `.github/workflows/ci.yml`; reconcile the architecture note with
   current source; and retain the 0019/0020 index rows with their
   metadata-consistent statuses. Preserve historical records.
3. Parse or match the README as text and assert all of the following:

   * the Unix command uses the project `releases/latest/download/install.sh`
     URL, pipes it to `sh -s --`, and passes
     `--dir ~/.global-git-hooks`;
   * the PowerShell command uses the project
     `releases/latest/download/install.ps1` URL and passes
     `-InstallDir "$env:USERPROFILE\.global-git-hooks" -NoPathUpdate`;
   * the Go 1.27.3 / dependency-change section refers to the published
     `v1.10.1` release rather than an unspecified future release.
4. Read `docs/guides/cicd-operations.md` and `.github/workflows/ci.yml` in
   the same test. Extract the build and publish reusable-workflow references
   from the workflow, then require the operations guide to contain exactly
   those references. Use literals only for stable file names and headings; do
   not duplicate the commit SHA in the test.
5. Prove every new assertion is live: copy the relevant documentation and
   workflow fixtures to a temporary directory, deliberately alter one
   expected install argument and one reusable-workflow revision, run the
   focused test against those fixtures, and confirm it fails with a diagnostic
   naming the altered fact. Do not dirty the repository to perform this
   experiment.
6. Run `make check-records`, `make markdownlint`, and `go test ./...`; record
   each command and exit status, plus the focused test result. Read the full
   changed documentation and inspect `git diff --check` before recording the
   phase as complete.

### ~~Phase 2: Reconcile current-facing documentation~~

Superseded before execution by the 2026-10-23 deviation below. Its work is
performed in Phase 1 so the new assertions are introduced only with the
current documentation they require.

1. ~~Update the README's upgrade/change section to state that `v1.10.1` was
   published, without implying that its release includes a library feature not
   present in `go.mod`. Retain the verified install commands verbatim unless
   a live installer test demonstrates a necessary correction.~~
2. ~~Update the operations guide's two reusable-workflow references and
   reconcile the current architecture note and documentation index.~~
3. ~~Run the documentation and Go test checks after the edits.~~

### Phase 3: Verify published installation paths

1. Download the latest published `install.sh` and `install.ps1` to a fresh
   temporary directory. Confirm their published interfaces still accept the
   README options and that their declared repository/product identities match
   this project. Do not place either script in the repository.
2. Execute the README-equivalent Unix installation on macOS or Linux with a
   unique temporary `--dir`; do not use `~/.global-git-hooks`. Confirm all of
   the following observable results:

   * installer exit status is zero;
   * only the expected binary is retained in the target directory;
   * `<target>/prepare-commit-msg identity` exits zero and begins with the
     latest release identity; and
   * the installer reports a checksum-verified installation.

3. On a Windows host, execute the README-equivalent PowerShell installation
   with a unique temporary `-InstallDir` and `-NoPathUpdate`. Confirm the
   installer exits zero, produces `prepare-commit-msg.exe`, and that
   `identity` reports the latest release identity. Remove only the temporary
   directory created for this step after recording its result.
4. If a latest-release change makes an existing README statement false,
   stop before altering scope, amend this PLAN and (if necessary) the MADR
   with the observed evidence, and request renewed approval.

### Phase 4: Close out

1. Run the final documentation and focused test checks:

   ```bash
   make check-records
   make markdownlint
   go test ./...
   git diff --check
   ```

2. Review the full resulting documentation, tests, and PLAN execution record.
   Confirm every current-facing claim is sourced from current code, workflow,
   or an execution result in this PLAN.
3. Record phase outputs verbatim or as concise command/output pairs below.
   Record any deviation with date, evidence, owner decision, added scope, and
   verification. Change the PLAN status to `complete` only after all
   acceptance criteria hold; otherwise leave it `in-progress` and state what
   remains.

## Verification

The implementation is accepted only when all of these are true:

* The documentation index lists every numbered record, including both 0019
  records and the 0020 MADR/PLAN, with metadata-consistent statuses.
* README installer commands match the tested published installer interfaces;
  the Unix command has passed a live isolated installation test and the
  PowerShell command has passed its corresponding Windows test.
* The README identifies `v1.10.1` as released, not future work.
* The operations guide's build and publish workflow references exactly match
  `.github/workflows/ci.yml`.
* The architecture note contains no current-state claim contradicted by the
  inspected hook, updater, configuration, Makefile, or workflow code.
* New documentation checks have been observed failing on deliberately broken
  temporary fixtures, then passing against the repository.
* `make check-records`, `make markdownlint`, `go test ./...`, and
  `git diff --check` pass at close-out.

## Rollout and Rollback

The change is documentation and test coverage only. It has no runtime rollout
or migration. If a documentation edit or assertion proves inaccurate, revert
only that approved-plan change in a follow-up commit, restore the last
evidence-backed wording, and amend this PLAN with the reason. Do not modify a
published historical record as rollback.

## Execution Record

### Deviation, 2026-10-23: combine the dependent tests and documentation edits

* **Found:** the original Phase 1 required assertions for facts that its
  original Phase 2 deliberately described as stale, then required a passing
  `go test ./...` result before Phase 2 could correct them.
* **Resolution (owner):** combine the two phases. Correct the documented
  facts and add their assertions in the same phase; use deliberately broken
  temporary fixtures, not the repository, to demonstrate the new assertions
  fail.
* **Scope change:** Phase 1 now includes the current-facing documentation
  edits. No product behavior, dependency, workflow, or release scope was
  added.

### Phase 1, 2026-10-23

* Updated `README.md` to identify `v1.10.1` as the published Go 1.27.3 and
  dependency update; retained both installation commands unchanged.
* Updated the operations guide to cite the complete build and publish reusable
  workflow references from `.github/workflows/ci.yml` at `v1.17.2`.
* Expanded the architecture note with the implemented skip, fallback, retry,
  soft-fail, configuration-authentication, and updater boundaries.
* Added `documentation_test.go`. `TestCurrentDocumentation` checks both
  README installation commands, the released-version wording, and that the
  operations guide cites the two workflow references extracted from CI.
  `TestCurrentDocumentationDetectsBrokenFixtures` changes the Unix install
  target and a workflow SHA in separate in-memory fixtures; each run fails
  with the assertion's named diagnostic before the unmodified files pass.
* `go test -run '^TestCurrentDocumentation' .` exited 0. `make check-records`
  exited 0. `make markdownlint` exited 0. `go test ./...` exited 0 when run
  outside the local sandbox, which blocks existing loopback-listener tests.

### Phase 3, 2026-10-23

* Downloaded the published latest `install.sh` and `install.ps1`; both identify
  this repository and expose the README's directory options. The latest
  release is `v1.10.1`.
* A Linux-host Unix installation into a unique temporary directory exited 0,
  retained one `prepare-commit-msg` binary, and its `identity` output was
  `v1.10.1 (release) 52ad5494ccfe`. The temporary directory was removed.
* A Windows-host PowerShell installation into a unique temporary directory
  with `-NoPathUpdate` exited 0, retained one `prepare-commit-msg.exe`, and
  its `identity` output was `v1.10.1 (release) 52ad5494ccfe`. The temporary
  directory was removed.

### Phase 4, 2026-10-23

* `gofmt -d documentation_test.go` produced no diff.
* `make check-records`, `make markdownlint`, and `git diff --check` exited 0.
* The full changed README, architecture note, operations guide, documentation
  index, test, and this execution record were reviewed. The acceptance
  criteria hold pending the required phase commit.

### Deviation, 2026-10-23: phase commit awaits protected-branch authorization

* **Found:** `HEAD` is `main`. The branch gate requires an explicit current-
  turn instruction to commit to `main`; approval to execute this PLAN is not
  commit authorization for that protected branch.
* **Resolution:** leave the verified changes unstaged and the PLAN
  `in-progress` until the owner explicitly authorizes a commit to `main` or
  chooses a branch. No push was requested or attempted.
