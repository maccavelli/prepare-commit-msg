---
status: accepted
date: 2026-11-10
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its CHANGELOG for v1.29.0, the compare v1.28.0...v1.29.0 by commit and file, its 0072-MADR, its source at v1.29.0, and its CI on the tagged commit); go-selfupdate-lib (its CHANGELOG for v1.26.0, the compare v1.25.0...v1.26.0 by commit and file, its 0041-MADR Amendment A1 and 0042-MADR, its source at v1.26.0, and its CI on the tagged commit); a scratch clone of this repository at 121879f moved to both; all 2026-11-10
informed: users of the installed hook, who receive the change through `update` once it is released
---
# Move to go-llmprovider-sdk v1.29.0 and go-selfupdate-lib v1.26.0, and release them with the v1.25.0 move

## Context and Problem Statement

The owner (2026-11-09), after the move to go-selfupdate-lib `v1.25.0`
(`0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md`) was on `main`: "2 of the
libraries this includes, including selfupdate-lib are being updated with
enhancements and bugfixes. i am going to wait until we can pin latest again
before i tag a release." 0030-MADR Amendment A1 holds that release until
then. And on 2026-11-10: "we can update the pins for the sibling libraries
now. i just released a new go-llmprovider-sdk and we need to ensure we pin
current tagged releases."

This repository pins its two libraries in four places **[read]**:

| Pin | Where | Pinned | Newest |
| :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `go.mod:6` | `v1.28.0` | `v1.29.0`, tagged 2026-11-10 |
| go-selfupdate-lib | `go.mod:7` | `v1.25.0` | `v1.26.0`, tagged 2026-11-09 |
| go-selfupdate-lib build workflow | `.github/workflows/ci.yml` | `875dec3c…` (`v1.25.0`) | `7c677842…` (`v1.26.0`) |
| go-selfupdate-lib publish workflow | `.github/workflows/ci.yml` | `875dec3c…` (`v1.25.0`) | `7c677842…` (`v1.26.0`) |

What has to be decided is whether the module lines and the two workflow pins
move, whether the app adapts to either release's changes in behavior, how the
README describes a release that carries 0030's move and this one, and when
that release is made.

Each claim is marked:

* **[read]**: from this repository, or from a library's changelog, source or
  git objects at the version named;
* **[probed]**: run in the session scratchpad on 2026-11-10, never in this
  tree;
* **[derived]**: reasoned from the two, and not run.

### The releases

* go-llmprovider-sdk `v1.29.0` is an annotated tag (`61ea66a0…`) on
  `4e338da59fe8c9037ec3d5b4c08b65982477edef`. `git ls-remote --tags` and
  `go list -m …@latest` both end at it. The library's CI (`42846065720`,
  `42844042192`) and Docs (`42844041991`) runs succeed on that commit
  **[probed]**.
* go-selfupdate-lib `v1.26.0` is an annotated tag (`df61822d…`) on
  `7c677842dface4b14b65b2e8b46aa0a7543bed13`. Both listings end at it. Its CI
  (`42756693188`) and Docs (`42756558914`) runs succeed on that commit
  **[probed]**.
* Neither release changes its `go.mod` **[read]**. The three `golang.org/x`
  modules and the Go toolchain stay where they are.
* No other module of the owner's is required **[read]**.

### What go-llmprovider-sdk v1.29.0 changes

From its CHANGELOG, its
`0072-MADR-keep-absolute-reset-times-and-limit-detail-on-apierror.md` and its
source **[read]**. `v1.28.0...v1.29.0` holds 10 commits and changes 38 files,
the Go changes in `llmprovider/` and its `internal/` packages:

* **Added:** `APIError.ResetAt`, the instant a service names for a limit to
  lift; `APIError.Limit`, of the new type `Limit`, what the service stated
  about the limit it hit; and `ClassifyStreamFailureBody`, which every
  built-in stream reader now uses.
* **Fixed:** an OpenCode `BudgetLimitError` (429) is `ErrQuotaExhausted` and
  is no longer retried, and `CreditLimitExceeded` is `ErrQuotaExhausted` at
  any status. A Gemini per-day quota reported inside a stream is
  `ErrQuotaExhausted`, as in an HTTP reply, and is no longer retried. A
  failure inside a stream carries the delay its error states.
* `APIError.Error()`'s format is unchanged; it shows `(retry-after …)`
  whenever `RetryAfter` is set, as before. The retry wrapper, `retry.go`, is
  unchanged: a stated wait above `RetryPolicy.MaxDelay` (30 seconds by
  default) returns the error at once.

### What go-selfupdate-lib v1.26.0 changes

From its CHANGELOG, its `0041-MADR-refuse-symlink-invocation-on-every-platform.md`
Amendment A1, its `0042-MADR-keep-go-identifiers-out-of-run-time-messages.md`
and its source **[read]**. `v1.25.0...v1.26.0` holds 5 commits and changes 47
files. It answers go-selfupdate-lib issue #10, raised from 0030-MADR:

* **Added:** `SymlinkInvocationError` (`Link`, `Target`, `Source`), which
  unwraps to `ErrSymlinkInvocation`; `SymlinkSource`; the sentinels
  `ErrSpecialModeBits`, `ErrNoAllowedRoot`, `service.ErrBinaryMismatch` and
  `taskscheduler.ErrLogonType`; and the `failed` event's classes
  `special-mode-bits` and `no-allowed-root`.
* **Changed:** the symlink refusal, for a program started through the link,
  reads `selfupdate: refusing to update through a symlink: <link> links to
  <target>; start <target> itself to update it`
  (`selfupdate/errors.go:210`). It no longer names
  `TargetPolicy.ExecutablePath`. Ten other run-time messages drop their Go
  identifiers. A setuid or setgid target, and a target under no allowed
  root, now report the classes above, where they were `error`.
* **Fixed:** `CleanupPending` and `KeptBackups` no longer fail for a program
  started through a link.
* `.github/`, `scripts/` and the `Makefile` are unchanged from `v1.25.0`
  (`git diff --stat v1.25.0 v1.26.0 -- .github scripts Makefile` is empty),
  and `v1.25.0`'s were unchanged from `v1.24.0`'s (0030-MADR) **[probed]**.

### What reaches this app

* **No code change is needed.** The app builds its installer with
  `selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{})`
  (`update.go:88`) and calls neither `CleanupPending` nor `KeptBackups`. It
  uses no `APIError` field that changed, matches no library message, and
  sets only `MaxAttempts` and `BaseDelay` on its `RetryPolicy`
  (`main.go:536`) **[read]**.
* **The symlink refusal now says what a user of this CLI can do.** The flaw
  0030-MADR recorded under "Consequences" is gone **[read]**.
* **`update` failures print through `cli.Exit`** as `update failed:
  <message>` (go-selfupdate-lib `selfupdate/cli/run.go:363`) **[read]**. So
  the refusal's stderr line begins `update failed: selfupdate: refusing to
  update through a symlink:`. 0030-PLAN Phase 5 step 7 says the stderr
  "starts with" the library's prefix; the line contains it after `update
  failed: ` **[derived]**.
* **The hook, for OpenCode's budget or credit limit and for a Gemini daily
  quota reported in a stream:** the attempt is not retried. The quota is not
  a `stopError` (`main.go:342`), so the hook tries the next model at once
  where it used to retry the model first **[read]** **[derived]**.
* **The hook, for a failure inside a stream that states a delay:** a retry
  waits that delay, up to 30 seconds, and a longer one moves to the next
  model at once. Before, the configured backoff applied **[derived]**.

### The move, on a scratch clone

On a clone of `121879f` **[probed]**:

* `go get github.com/maccavelli/go-llmprovider-sdk@v1.29.0
  github.com/maccavelli/go-selfupdate-lib@v1.26.0` and `go mod tidy` change
  two lines of `go.mod` and four of `go.sum`:
  * go-llmprovider-sdk `v1.29.0`
    `h1:KVuUfHpx6TtKwjH+XjSSkBU+PRRKg9aWa6p69iaOzy0=`, and for `/go.mod`
    `h1:dE+W7oDyvVmAxpSOuIJO3qSD3bDTh/olLt/TbByBE1M=`, the same as
    `v1.28.0`'s;
  * go-selfupdate-lib `v1.26.0`
    `h1:giqUiSV90pDZErgh+fEWR10KnM6CdDKOgcjDonVa1xM=`, and for `/go.mod`
    `h1:DIqb0hzmxkYhVTj6iyaHpntadoHkeOkHWyTsdw+MBHE=`, the same as
    `v1.25.0`'s.
* `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
  verified".
* `make verify` exits 0: "0 issues."; total coverage 84.9% (minimum 80.0%);
  "No vulnerabilities found."; "ci filters: 11 patterns, 11 reads checked, 0
  problems"; the tools' 42 tests and the scripts' 34 "OK"; the cross-builds.

## Decision Drivers

* **Pins are current** unless a record names a reason to hold one.
* **The workflow pins name the version `go.mod` requires**
  (0030-MADR D2; `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2, A1).
* **The change is a pin move.** New library features are separate
  decisions.
* **The README says what a user of the release can see,** from the release
  before it.
* **The release the owner held (0030-MADR A1) ships once the pins are
  current.**

## Considered Options

* **A. Move both modules and both workflow pins, change no code, fold the
  unreleased v1.25.0 README section into one section for both libraries,
  and release `v1.11.1`.**
* **B. As A, but keep the v1.25.0 section and add a second one** for this
  move.
* **C. As A, and also use the new SDK fields:** stop the run at
  `ErrQuotaExhausted`, or name `APIError.ResetAt` in the hook's warning.
* **D. Move the modules only,** and leave both workflow pins at `v1.25.0`'s
  commit.

## Decision Outcome

Chosen option: "A", because it makes every pin current at the cost of two
module lines and two workflow lines, needs no code, and lets the held
release ship as one step from `v1.11.0`. B would describe a library version,
`v1.25.0`, that no release carries. C changes the hook's behavior, which a
pin move should not decide: a Gemini quota is per model, so stopping the run
on it would skip models that could answer. D leaves `ci.yml`'s comment false.

### D1. The modules move

* `go.mod` requires `github.com/maccavelli/go-llmprovider-sdk v1.29.0` and
  `github.com/maccavelli/go-selfupdate-lib v1.26.0`, after `go mod tidy`.
* `go.mod` changes on those two lines alone, and `go.sum` on those modules'
  four lines alone. `go mod tidy -diff` is empty, and `go mod verify`
  passes.
* The `golang.org/x` modules and the Go toolchain do not change.

### D2. Both workflow pins move

* Both `uses:` lines in `.github/workflows/ci.yml` pin
  `7c677842dface4b14b65b2e8b46aa0a7543bed13 # v1.26.0`.
* The comment above the build job names `v1.26.0` and says the publish path
  is unchanged from `v1.24.0`'s, which published live as this repository's
  `v1.11.0` (run `42434050386`). It cites this record, and 0010-MADR D2, A1.
* `docs/guides/cicd-operations.md` quotes the new lines, in the same commit.
* The `release` job's `with:` and permissions are unchanged.

### D3. No code change

* No `.go` file changes. The app matches none of the new sentinels or
  types, and the hook's run loop and retry policy are unchanged.
* No test is added. The behaviors are the libraries', covered by their own
  tests **[read]**.

### D4. One README section for the release

The section 0030-MADR D4 added, "Changes with go-selfupdate-lib v1.25.0",
describes a version no release will carry, so it becomes the release's one
section, as earlier two-library moves have
("Changes with go-llmprovider-sdk v1.26.1 and go-selfupdate-lib v1.24.0").
This is the correction 0030-MADR A1 foresaw.

* **Heading and contents line:** "Changes with go-llmprovider-sdk v1.29.0
  and go-selfupdate-lib v1.26.0", in place of the v1.25.0 ones.
* **Opening:** release `v1.11.1` moves from go-llmprovider-sdk `v1.28.0` to
  `v1.29.0`, and from go-selfupdate-lib `v1.24.0` to `v1.26.0`, citing
  0030-MADR and this record.
* **Kept from 0030's section:** the Linux symlink refusal, and the
  `symlink-invocation` class.
* **Changed:** the refusal says to start the target to update it.
* **Added:** the `special-mode-bits` and `no-allowed-root` classes; that an
  OpenCode budget or credit limit, and a Gemini daily quota reported partway
  through an answer, are not retried and the next model is tried; that a
  failure partway through an answer waits the delay the service states, up
  to 30 seconds.
* **Closing line:** `configure` and the installers behave as in `v1.11.0`.
* The "Security Note" and "Option B" sentences 0030 added stand unchanged.

### D5. The release (Q1)

A patch release, `v1.11.1`, once this move is on `main` and CI is green
(Q1 A). It carries 0030's move and this one. Its checks are 0030-PLAN
Phase 5's, run once by this record's PLAN, with step 7's stderr read as
"What reaches this app" gives it. That run completes 0030-PLAN.

### Consequences

* Good, because every pin is current, and the held release can go out.
* Good, because the symlink refusal tells a user of this CLI what to do, and
  the issue 0030 raised is closed by the library.
* Good, because an exhausted OpenCode budget or Gemini daily quota no longer
  spends retries that cannot succeed.
* Neutral, because `configure`, the installers and the hook's success path
  behave as before **[probed]** **[derived]**.
* Bad, because a user who relied on a retry riding out an OpenCode budget
  limit now sees the next model tried at once. The budget does not lift
  within a retry's wait.
* Bad, because a README section committed in `121879f` is rewritten before
  any release carried it. Its record, 0030-MADR, says so in A1.

### Confirmation

* `go.mod` and `go.sum` change only on D1's lines. `go mod tidy -diff` is
  empty, and `go mod verify` passes.
* `7c677842dface4b14b65b2e8b46aa0a7543bed13` occurs twice in `ci.yml` and
  twice in the operations guide, and `875dec3c9d475ca47c4b012ca43742de832ef626`
  in neither.
* No `.go` file changes.
* `make verify`, `make preflight`, `make disclosure-check` and
  `make preflight-hosts` pass on the committed `HEAD`. CI is green on
  `main`, its build rehearsal included.
* `make check-records` and `make markdownlint` pass.
* Under Q1 A: the tag run publishes `v1.11.1` through
  `publish-selfupdate-release.yml@7c677842…`; the release is immutable with
  nine assets; `update` moves an installed `v1.11.0` to it; and on a Linux
  host `v1.11.1` started through a symlink refuses `update` with
  `update failed: selfupdate: refusing to update through a symlink: <link>
  links to <target>; start <target> itself to update it`, leaving the target
  unchanged.

## Pros and Cons of the Options

### A. Move everything, one README section, release v1.11.1

* Good, because it is four pin lines and README text.
* Good, because a reader of the release sees one step from `v1.11.0`.
* Bad, because it rewrites a committed, unreleased README section.

### B. Keep the v1.25.0 section and add another

* Good, because no committed text is rewritten.
* Bad, because the first section names a version no release carries, and
  its refusal text is no longer what `v1.11.1` prints.

### C. Also use the new SDK fields

* Good, because the hook could stop sooner on a quota, or say when it
  resets.
* Bad, because it changes the hook's run loop in a pin move; and a Gemini
  quota is per model, so stopping on it is wrong for that provider. It
  deserves its own record if wanted.

### D. Move the modules only

* Good, because CI's workflows do not change.
* Neutral, because the workflow files and scripts are identical at both
  commits **[probed]**.
* Bad, because `ci.yml`'s comment says the pins name the version `go.mod`
  requires, and would be false.

## Owner questions

* **Q1 (D5). Is `v1.11.1` released once this move is on `main` and CI is
  green?**
  * A, recommended: yes. It carries 0030's move and this one, and the
    README section names it. The pruning after it follows the operations
    guide's "Controlled release" under
    `0026-MADR-keep-five-stable-releases.md`, on the owner's word.
  * B: not yet. Commit and push the move, and hold the tag until the owner
    says so; the README section waits for it.

**Answered 2026-11-10.** The owner: "we will release v1.11.1 after ci
passes. accept the madr and plan approved to proceed." That is Q1 A, and
D5 records it. The chosen option, A, was not questioned and stands. The
same message accepted this record and approved its PLAN.

## More Information

* The libraries' records, cited by repository and filename:
  go-llmprovider-sdk
  `0072-MADR-keep-absolute-reset-times-and-limit-detail-on-apierror.md`;
  go-selfupdate-lib `0041-MADR-refuse-symlink-invocation-on-every-platform.md`
  (Amendment A1) and
  `0042-MADR-keep-go-identifiers-out-of-run-time-messages.md`.
* go-selfupdate-lib issue #10
  (<https://github.com/maccavelli/go-selfupdate-lib/issues/10>), raised from
  0030-MADR.
* The move this release also carries, and its hold:
  [0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md](0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md)
  D5, A1.
* The last move of go-llmprovider-sdk:
  [0028-MADR-adopt-go-llmprovider-sdk-v1-28-0.md](0028-MADR-adopt-go-llmprovider-sdk-v1-28-0.md).
* The workflow-pin rule:
  [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md).
* The release window:
  [0026-MADR-keep-five-stable-releases.md](0026-MADR-keep-five-stable-releases.md).
* The implementation:
  [0031-PLAN-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md](0031-PLAN-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md).
