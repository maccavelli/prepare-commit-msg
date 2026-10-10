---
status: accepted
date: 2026-10-06
decision-makers: Project Owner
consulted: go-selfupdate-lib (its records 0010 to 0013, which made v1.5.1 to v1.9.0)
informed: users of the installed hook, who receive the change through `update`
---
# Move to go-selfupdate-lib v1.9.0, and keep publishing on the proven workflow until v1.9.0's has published a release

## Context and Problem Statement

`prepare-commit-msg` requires `github.com/maccavelli/go-selfupdate-lib
v1.5.0` (`go.mod:7`), the release that
[0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md](0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md)
adopted. The library gives it four things:

* **the `update` command,** `cli.Command` (`main.go`);
* **the updater,** `update.go`: a GitHub source, an exact selector for six
  platforms, the standalone installer, and the strict version policy;
* **the `version` output,** `buildinfo.Identity`, stamped by the
  `Makefile`'s `buildinfo.version` and `buildinfo.kind`;
* **the release job,** `.github/workflows/ci.yml:115`, which calls the
  library's reusable `publish-selfupdate-release.yml`, pinned to
  `6deaa524cfb2…` (`v1.5.0`).

The library has since released five versions:

| Tag | Commit | Date | Record | What it carries |
| :--- | :--- | :--- | :--- | :--- |
| `v1.5.1` | `c7a8b4c` | 2026-10-03 | 0010, v1.5.1 PLAN | Fixes that keep every documented contract |
| `v1.6.0` | `b1f1caa` | 2026-10-04 | 0010, v1.6.0 PLAN | The contracts the owner decided (Q1–Q6, A14) |
| `v1.7.0` | `e825cda` | 2026-10-05 | 0011 | `selfupdate/service`: systemd, launchd and Windows SCM lifecycles |
| `v1.8.0` | `a99aa66` | 2026-10-05 | 0012 | `selfupdate/archive` and `selfupdate/codesign` |
| `v1.9.0` | `39b1294` | 2026-10-06 | 0013 | `selfupdate/releasespec` and `build-selfupdate-release.yml` |

go-selfupdate-lib's
`docs/decisions/0013-MADR-build-and-stage-release-workflow.md` lists this
repository among those it informs. Its consumer table shows this
repository's platform list written several times over: `build-all`,
`platforms-json`, the selector, and `scripts/verify-release.sh`.

How should `prepare-commit-msg` move to v1.9.0? And which of the library's
changes, new workflows and opt-in features should it take now?

### What was measured

Measured on 2026-10-06 at `HEAD` `abd4bc9`, with Go 1.27.1 on macOS. The
trial ran on scratch clones, and the reading used the library's tags
(`git show <tag>:…`). Nothing was committed.

**The bump is two lines.** `go get
github.com/maccavelli/go-selfupdate-lib@v1.9.0` and `go mod tidy` change
only the library's line in `go.mod` and its two lines in `go.sum`
(`h1:kP7ISSs+…`). The library's own `go.mod` is identical at `v1.5.0` and
`v1.9.0`: `go 1.27.1`, `x/mod v0.40.0`, `x/sys v0.47.0` and `x/term
v0.43.0`. So no other module moves.

| Check | `v1.5.0` (baseline) | `v1.9.0` |
| :--- | :--- | :--- |
| `go build ./...`; `CGO_ENABLED=0 go vet` for linux, darwin, windows | 0 | 0 |
| `go test -race -count=1 ./...` | 5 packages ok | 5 packages ok |
| `go mod tidy -diff` | 0 | 0 |
| `python3 scripts/go-precheck.py` (go.mod and go.sum staged for `v1.9.0`) | exit 0 | exit 0: `v1.9.0 resolved from GitHub`, `0 issues.`, `No vulnerabilities found.` |
| `make verify` | exit 0, total coverage 84.5% | exit 0, total coverage 84.5%, per function identical |
| `TestMigrationByteForByte` and the update tests | pass | pass |
| `update --help`, `--help`, the `version` form | — | identical to the baseline |

**The API is compatible.** apidiff, with go-llmprovider-sdk's
`scripts/check_api.py` pin, finds no incompatible change from `v1.5.0` to
`v1.9.0`. It finds 36 additions, 7 of them new packages. This repository
uses 33 of the library's identifiers.

**The migration fixtures are unchanged.**
`selfupdate/cli/testdata/migration/` is identical at `v1.5.0` and `v1.9.0`.
Each of the nine files this repository copied (`testdata/migration/`)
matches both, byte for byte. So the `update --check` contract the hook
tests has not moved.

**What `update` does differently.** These come from the library's records
and goldens (`selfupdate/cli/testdata/golden/`). None of them is on the
text `--check` path:

| Case | `v1.5.0` | `v1.9.0` | Record |
| :--- | :--- | :--- | :--- |
| `update --json`: the result object | `"schema_version":1` | `"schema_version":2`, adding `service_started` and, when there are any, `warnings` | 0010 A3, A4 |
| An error after the install completed, such as `unlock failed` | `update failed: …`, exit 1 | a `warning: unlock failed` line on stderr, then `updated vX -> vY`; exit 0 | 0010 Q3 (`golden/warning.*`) |
| `update --check`: the summary itself cannot be written | exit 10 | exit 1, with the error | 0010 C3 |
| `--json` with a refused option | no result object | a result object | 0010 C4 |
| A URL in an error | with its query string | without it | 0010 A5 |
| Staging files and backups left by a crash | stay | removed by the next run | 0010 Q6 B11 |
| A setuid or setgid target | replaced | refused, unless `TargetPolicy.AllowSpecialModeBits` | 0010 Q6 B12 |
| Windows: a backup still in use | — | stays on a cleanup receipt, version 2. Releases before `v1.5.1` refuse that receipt. | 0010 B5, B6, A1 |

The rest of 0010 hardens checks a user never sees unless a release was
tampered with: A1, A8 and A12. The `version` form and the `buildinfo`
stamps are unchanged; `buildinfo/` has no diff.

**The publish workflow still accepts this repository's call.** Its
`workflow_call` inputs, required inputs, defaults and permissions are
unchanged between `6deaa52` and `v1.9.0`. Only `platforms-json`'s
description changed: it adds an optional `format`. Its job gains:

* a `gh` check that now actually fails below gh 2.81.0. The old check ran
  `--help` probes, which always exit 0.
* two Go steps that run only for an archive release
  (`steps.verify.outputs.packed == 'true'`). This repository's raw binaries
  skip them.
* `release-latest-flag.sh` on create and publish, so a stable tag lower
  than the current latest release, a backport, is not made latest (0010 D9).

**v1.9.0's publish workflow has not yet published a release.**
* go-selfupdate-lib's
  `docs/decisions/0013-PLAN-build-and-stage-release-workflow.md` is
  `in-progress` at `v1.9.0` and on its `origin/main`.
* Its CI rehearsal (run 37404279305) built and checked raw and archive sets
  on five runners. It does not create a release.
* B6 step 3, "the live publish rehearsal in a throwaway repository", is
  "still open … which waits for the owner".
* No consumer pins it: this repository pins `v1.5.0`'s commit, and five
  other repositories still pin mcplib's copy at `v1.4.1`.

So if this repository moved its pin, its next tag would be the workflow's
first live publish. If that publish failed, the tag would have no release,
and the fix would need a new commit and a new tag.

**A release-candidate tag cannot rehearse it here today.**
`scripts/verify-release.sh:48` accepts only `vX.Y.Z`. The tag build runs
`make verify-release`, so a `v1.7.0-rc.1` tag would fail before the
publish job.

**The new build workflow is optional.** go-selfupdate-lib's
`docs/guides/building-releases.md` §9 says "Nothing is required". The
0013 PLAN's rollout calls `build-selfupdate-release.yml` and `releasespec`
"new and opt-in". Adopting them takes three steps (0013-MADR §1):

1. a release spec;
2. an updater configured from the embedded spec;
3. the build, checksum and staging jobs replaced by the build workflow.

That would remove this repository's repeated platform list. go-selfupdate-lib
0014, shared installer templates in the build workflow, is accepted and
targets `v1.10.0`. It is not released.

**Opt-in features this repository does not need:**
* services (0011): it is a git hook, not a service;
* archives and macOS re-signing (0012): it publishes raw binaries;
* `TargetPolicy.AllowSpecialModeBits`: its binary lives under the user's
  home directory.

**The README now overstates one rule.** "Self-Update" says the exit
status is `1` "on any error" (`README.md:372`). From `v1.6.0` an error
after the install completed is a warning, with exit 0. The README does not
name `--json`'s `schema_version`.

## Decision Drivers

* **Take the library's fixes:** crash leftovers removed, special mode bits
  refused, integrity errors typed, URLs in errors trimmed.
* **A release always publishes.** A tag whose publish fails costs a version
  number and a follow-up release.
* **No user loses a working path.** `update`, `update --check` and
  `version` behave as the README says.
* **Each step lands green, and reverts alone.**
* **Keep the version bump small.** The release-spec build is a separate
  change to how releases are built.

## Considered Options

* **A. Move the module to v1.9.0, and keep the publish pin at `v1.5.0`'s
  commit until v1.9.0's workflow has published a release.**
* **B. Move the module and the publish pin to v1.9.0 together.**
* **C. Move both, and rehearse the publish first with a release-candidate
  tag.**
* **D. Move to v1.9.0, and adopt the release spec and the build workflow
  in the same change.**
* **E. Stay on v1.5.0.**

## Decision Outcome

Chosen option: **"A"**, because it takes every library fix with a two-line
module change, and publishes the next release through the workflow that
already published `v1.5.0` and `v1.6.0` here. It does not make this
repository's tag the first live publish of a workflow whose own record is
still waiting for one.

### D1. The module moves to `v1.9.0`, and no opt-in feature is adopted

* `go.mod` requires `go-selfupdate-lib v1.9.0`.
* `update.go`, `main.go` and the `Makefile` are unchanged. The selector,
  installer, version policy and stamps are as they are.
* None of services, archives, codesign, `releasespec` or
  `AllowSpecialModeBits` is used.

### D2. The publish pin stays at `6deaa524cfb2…` (`v1.5.0`) for now

* **This amends 0008's rule** that `ci.yml` pins the commit "of the
  go-selfupdate-lib release chosen". The module and the workflow are two
  pins with separate evidence:
  * the module is proven by this repository's tests;
  * the workflow is proven by a published release.
* The comment on the pin says why it differs from `go.mod`.
* **The pin moves to `39b1294` (`v1.9.0`), or to a later release, in a
  separate change,** once a publish through it is on record. That is
  go-selfupdate-lib 0013-PLAN B6 step 3, or another repository's release.
  That change re-reads the workflow's inputs, as 0008 did.

### D3. A test holds the `--json` result's schema version

* A test drives `update --check --json` through the real `main`, with the
  existing fixture source, and requires the result object's
  `schema_version` to be 2.
* It is the contract the README will state. A later release that changes
  it then fails here first.

### D4. The README states what changed

"Self-Update" says:

* `--json`'s result carries `schema_version` 2;
* an error after the new binary is in place is reported as a `warning:`
  line, and the run still exits 0;
* `1` is for an error before that.

### D5. The release is the next minor, `v1.7.0`

* The owner tags it after CI is green on `main`.
* It is minor, not a patch, because `update --json`'s documented output
  and the exit status of a late error change.

### Consequences

* Good, because the hook gets the library's fixes from five releases.
* Good, because the next tag publishes through a workflow that has
  published here twice.
* Good, because `update --json`'s schema version becomes a tested
  contract.
* Neutral, because the module and the publish pin name different library
  versions for a while. The pin's comment says why.
* Bad, because a follow-up change is needed to move the pin, and it depends
  on evidence from outside this repository.
* Bad, because a script that parses `update --json` and requires
  `schema_version` 1 must change.
* Bad, on Windows only: a user who returns to `v1.6.0` or earlier while a
  backup is still in use may see that release refuse the newer cleanup
  receipt.

### Confirmation

* **Every phase passes:**
  * `make verify`;
  * `make verify-staged`;
  * `go test -race -count=1 ./...`;
  * `CGO_ENABLED=0 go vet` for linux, darwin and windows;
  * `go mod tidy -diff`.
* **D3's test passes at `v1.9.0`,** and fails on a scratch copy held at
  `v1.5.0` (schema version 1).
* **`TestMigrationByteForByte` passes unchanged.**
* **CI passes on Linux, macOS and Windows, on `main` and on the tag.**
  The tag's release is published by the `v1.5.0` workflow.
* **The installed hook** reports `go-selfupdate-lib v1.9.0` in its build
  information once `update` has run. Run by the owner, `update --check`
  then reports it up to date.

## Pros and Cons of the Options

### A. Module to v1.9.0, publish pin kept until v1.9.0's has published

* Good, because every library fix arrives, and the release path is the one
  already proven here.
* Good, because the module bump reverts alone.
* Bad, because the two pins differ until a follow-up, which waits on
  evidence from outside this repository.

### B. Module and publish pin to v1.9.0 together

* Good, because one change, and the pins agree, as 0008 had them.
* Good, because this repository's tag would supply the live publish that
  go-selfupdate-lib 0013 is waiting for.
* Bad, because a defect in the publish job's new steps (the gh check, the
  latest-flag script) would surface on a real tag. The fix then costs a
  new commit and a new version.

### C. Both, rehearsed first with a release-candidate tag

* Good, because the publish is proven on a prerelease, which stable
  clients never read (go-selfupdate-lib 0005-MADR §5).
* Bad, because `verify-release.sh` refuses a suffixed tag today. So it adds
  a change to the release checks, and a permanent prerelease in this
  repository's history.

### D. v1.9.0 with the release spec and the build workflow

* Good, because one platform list instead of four, and a fixed build
  recipe.
* Bad, because it replaces how releases are built in the same change as the
  version bump, and depends on the same unproven publish path.
* Bad, because 0014's installer templates, in `v1.10.0`, change the build
  workflow again. Adopting once, after `v1.10.0`, costs less.

### E. Stay on v1.5.0

* Good, because it is no work.
* Bad, because the hook keeps:
  * crash leftovers in its install directory;
  * replacing setuid or setgid targets;
  * untrimmed URLs in errors;
  * the drift from the library's documented contracts.

## Owner questions

Each has a recommendation, written into the outcome above.

**Answered 2026-10-06:** "questions: follow the recommendations." The outcome above stands as
written, and this record is `accepted`.

* **Q1 (D2).** Keep the publish pin at `v1.5.0`'s commit until v1.9.0's
  workflow has published a release (recommended); move it now (B); or
  rehearse with a release candidate first (C)?
* **Q2 (D1).** Adopt no opt-in feature now (recommended), or adopt the
  release spec and build workflow too (D)?
* **Q3 (D3).** Add the `--json` schema test (recommended), or rely on the
  library's goldens?
* **Q4 (D4).** Update the README's "Self-Update" (recommended), or leave
  it?
* **Q5 (D5).** Release `v1.7.0` (recommended), or `v1.6.1`?

## Amendments

### A1 (2026-10-09): D2's rule reads by the publish path

*Decided in
[0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md](0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md)
(P1), and applied in its PLAN's Phase 2. D2's text above is left as
written.*

* **The rule now reads:** the publish pin is the newest go-selfupdate-lib
  commit whose publish path, meaning the publish workflow file and the
  scripts it runs from the pinned checkout, is unchanged from a commit that
  has published live.
* **Why:**
  * `v1.13.0`'s publish workflow is the same blob (`a0c842a8…`) as
    `v1.11.0`'s, which published live in go-selfupdate-lib 0015-PLAN's
    `v1.11.0` rehearsal;
  * the four scripts it runs are unchanged;
  * the build workflow, adopted by 0012-MADR (B1), must be pinned to the
    same commit as the publish workflow, and that commit's version is the
    one `go.mod` requires.
* **The pins** are both
  `5e199c831b5691ea687943e3c3fd495d50c739ed` (`v1.13.0`).

## More Information

* go-selfupdate-lib records, by full filename in that repository:
  * `docs/decisions/0010-MADR-remediate-second-debugging-pass-findings.md`
    and its PLANs `0010-PLAN-v1-5-1-contract-preserving-fixes.md`,
    `0010-PLAN-v1-6-0-owner-contracts.md` and `0010-PLAN-tooling-fixes.md`;
  * `docs/decisions/0011-MADR-reference-service-lifecycles.md`;
  * `docs/decisions/0012-MADR-archive-assets-and-macos-codesign.md`;
  * `docs/decisions/0013-MADR-build-and-stage-release-workflow.md` and
    `0013-PLAN-build-and-stage-release-workflow.md`, Phase B6;
  * `docs/decisions/0014-MADR-shared-installer-templates.md`, not released;
  * `docs/guides/building-releases.md` and
    `docs/guides/migrating-from-mcplib-selfupdate.md` §6.
* [0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md](0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md):
  the first adoption, the migration fixtures, and the pin rule D2 amends.
* [0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md](0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md):
  the release before this one.
* **Not decided here:**
  * moving the publish pin, D2's follow-up;
  * adopting the release spec and build workflow, after `v1.10.0`;
  * gobble-cli's go-selfupdate-lib pin at `v1.7.0`.
