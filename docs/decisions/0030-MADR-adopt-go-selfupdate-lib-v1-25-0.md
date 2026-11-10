---
status: accepted
date: 2026-11-09
decision-makers: Project Owner
consulted: go-selfupdate-lib (its CHANGELOG for v1.25.0, the compare v1.24.0...v1.25.0 by commit and file, its 0041-MADR, its migration guide's section 25, its source at v1.25.0, and its CI on the tagged commit); go-llmprovider-sdk's tags; a scratch clone of this repository at d0da9ae moved to v1.25.0; all 2026-11-09
informed: users of the installed hook, who receive the change through `update` once it is released
---
# Move to go-selfupdate-lib v1.25.0, and accept its refusal to update through a symlink on every platform

## Context and Problem Statement

The owner (2026-11-09): "there is a new tagged release of go-selfupdate-lib.
i think the others are current. we need to pin to the new selfupdate-lib, it
has bugfixes." And, on asking for this pair: "ensure the other sibling
sdk/lib are current."

This repository pins its two libraries in four places **[read]**:

| Pin | Where | Pinned | Newest |
| :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `go.mod:6` | `v1.28.0` | `v1.28.0` |
| go-selfupdate-lib | `go.mod:7` | `v1.24.0` | `v1.25.0`, tagged 2026-11-09 |
| go-selfupdate-lib build workflow | `.github/workflows/ci.yml:117` | `74ab5dd3…` (`v1.24.0`) | `875dec3c…` (`v1.25.0`) |
| go-selfupdate-lib publish workflow | `.github/workflows/ci.yml:132` | `74ab5dd3…` (`v1.24.0`) | `875dec3c…` (`v1.25.0`) |

What has to be decided is whether the module and the two workflow pins
move, whether the app accepts the release's one change in behavior or works
around it, and whether the move is released.

Each claim is marked:

* **[read]**: from this repository, or from the library's changelog, source
  or git objects at the version named;
* **[probed]**: run in the session scratchpad on 2026-11-09, never in this
  tree;
* **[derived]**: reasoned from the two, and not run.

### The siblings are current

* go-llmprovider-sdk: `git ls-remote --tags` and `go list -m -versions` both
  end at `v1.28.0`, the version `go.mod` requires **[probed]**.
* go-selfupdate-lib: both end at `v1.25.0` **[probed]**.
* No other module of the owner's is required. The three `golang.org/x`
  modules are required at the versions go-selfupdate-lib requires, which
  `v1.25.0` leaves unchanged: `mod v0.42.0`, `sys v0.49.0`, `term v0.47.0`
  **[read]**. `go list -m -u all` shows newer `golang.org/x` releases; moving
  those is not asked for, and is out of scope here.

### The release

* `v1.25.0` is an annotated tag (`25d31a42…`) on
  `875dec3c9d475ca47c4b012ca43742de832ef626`. The library's CI
  (`42687360803`) and Docs (`42671188899`) runs succeed on that commit.
  `go list -m -versions` names it **[read]** **[probed]**.
* `v1.24.0...v1.25.0` holds 5 commits and changes 20 files. The Go changes
  are in `selfupdate/`: `errors.go`, `replace.go`, `target.go`, `types.go`,
  `updater.go`, and comments in `doc.go` and `standalone.go`, with their
  tests. `.github/`, `scripts/`, `internal/`, `buildinfo/` and `go.mod` are
  unchanged **[read]**.
* There is no GitHub release for the tag. As for earlier moves, the module
  does not depend on one: `go mod verify` reports "all modules verified"
  **[probed]**.

### What v1.25.0 changes

From its CHANGELOG, its
`0041-MADR-refuse-symlink-invocation-on-every-platform.md`, its migration
guide's "25. From v1.24 to v1.25" and its source **[read]**:

* **Fixed, on Linux: a program started through a symlink to its binary is
  refused** (the library's issue #9). macOS and Windows already refused it.
  On Linux `os.Executable` reads `/proc/self/exe`, which the kernel has
  already resolved, so the check never saw the link and the link's target
  was replaced. The path in `os.Args[0]` is now checked too: made absolute
  when it holds a separator, looked up on `PATH` when it does not. The update
  is refused only when that path is a symlink to the running binary. A
  symlinked directory on the way is still followed.
* **Added, `ErrSymlinkInvocation`,** which the refusal wraps, and the
  `failed` event's class `symlink-invocation`, where this refusal was
  `error`.
* **Changed, the refusal's message.** It keeps `selfupdate: refusing to
  update through a symlink` as its prefix, and goes on: `<link> links to
  <target>; run <target>, or set TargetPolicy.ExecutablePath to it`.
* The library's `make apicheck` reports it compatible with `v1.24.0`.
* To update through a link on purpose, a program sets
  `TargetPolicy.ExecutablePath` to the resolved path; `os.Args[0]` is then
  not read.

### What reaches this app

* **The check, unchanged in the app's code.** `update.go:88` builds the
  installer with `selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{})`.
  It sets no `TargetPolicy.ExecutablePath`, so `v1.25.0`'s check applies
  **[read]**.
* **No test depends on the old behavior.** No test here matches the
  refusal's text or the `error` class. `migration_test.go:63` sets
  `ExecutablePath`, so it never reads `os.Args[0]` **[read]**.
* **Who sees the change.** A Linux user who runs `update` through a link to
  the binary: a link on `PATH`, or the per-repository link that README
  "Option B" shows (`README.md:213`, `ln -sf … .git/hooks/prepare-commit-msg`)
  if `update` is run through it. Their `update` now fails with the message
  above, where before it replaced the link's target. Git running the hook
  through that link is unaffected: the hook never updates itself **[read]**
  **[derived]**.
* **Who does not.** The default install puts a regular file in
  `~/.global-git-hooks` (`README.md:191-203`). The installers, rendered from
  the library's templates by its build workflow, place a file and create no
  link **[read]**. macOS and Windows users see only the longer message
  **[read]** **[derived]**.
* **The message names a library field.** "set TargetPolicy.ExecutablePath to
  it" speaks to a program's author, not to the user of this CLI. The first
  half, "run `<target>`", is what a user can act on **[read]** **[derived]**.

### The move, on a scratch clone

On a clone of `d0da9ae` **[probed]**:

* `go get github.com/maccavelli/go-selfupdate-lib@v1.25.0` ("upgraded
  … v1.24.0 => v1.25.0") and `go mod tidy` change one line of `go.mod` and
  two of `go.sum`. The new hashes are
  `h1:K3IxUbldgxkssYQo3eW+b5gtJpBcz/CUjAMwGIMY8VE=`, and for `/go.mod`
  `h1:DIqb0hzmxkYhVTj6iyaHpntadoHkeOkHWyTsdw+MBHE=`, the same as
  `v1.24.0`'s. `go mod tidy -diff` is empty, and `go mod verify` prints "all
  modules verified". `go list -m all` is otherwise unchanged.
* `make verify` exits 0:
  * mod-check, fmt-check, golangci-lint ("0 issues."), `go vet`, and
    `go test -race ./...`;
  * total coverage 84.9%, against a minimum of 80.0%;
  * govulncheck: "No vulnerabilities found.";
  * workflow-lint: "ci filters: 11 patterns, 11 reads checked, 0
    problems";
  * the tools' 42 tests and the scripts' 34;
  * the six release cross-builds.

## Decision Drivers

* **Pins are current** unless a record names a reason to hold one.
* **The update path refuses what it cannot replace safely,** the same way on
  every platform.
* **The workflow pins name the version `go.mod` requires,** as the comment
  above them says (`ci.yml:110-116`, after
  `0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md`
  D2).
* **The change is a pin move.** New library features are separate
  decisions.
* **The README says what a user can see.**

## Considered Options

* **A. Move the module and both workflow pins, accept the refusal, and say
  in the README how to update a linked install.**
* **B. As A, but keep updating through a link:** set
  `TargetPolicy.ExecutablePath` to `filepath.EvalSymlinks(os.Executable())`
  in `update.go`, so the link's target is replaced on every platform.
* **C. Move the module only,** and leave both workflow pins at `v1.24.0`'s
  commit.
* **D. Hold at `v1.24.0`.**

## Decision Outcome

Chosen option: "A", because it keeps every pin current at the cost of one
module line and two workflow lines, and it takes the release's fix as the
library intends it. The refusal is what this app already does on macOS and
Windows, so Linux joins them. B would undo the fix on Linux and widen the
update path on macOS and Windows, a change of policy beyond a pin move. C
leaves the CI comment's claim false, or rewrites it to explain a split the
next move undoes. D holds a fix with no reason to hold.

### D1. The module moves

* `go.mod` requires `github.com/maccavelli/go-selfupdate-lib v1.25.0`,
  after `go mod tidy`.
* `go.mod` changes on that line alone. `go.sum` changes on that module's two
  lines alone. `go mod tidy -diff` is empty, and `go mod verify` passes.
* go-llmprovider-sdk, the `golang.org/x` modules and the Go toolchain do not
  change.

### D2. Both workflow pins move

* Both `uses:` lines in `.github/workflows/ci.yml` pin
  `875dec3c9d475ca47c4b012ca43742de832ef626 # v1.25.0`.
* The comment above the build job is rewritten: both are pinned to
  `v1.25.0`'s commit, the version `go.mod` requires; the publish path (the
  publish workflow file and the scripts it runs) is unchanged from
  `v1.24.0`'s, which has published live as this repository's `v1.11.0` (run
  `42434050386`); it cites this record, and
  `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2, A1.
* `docs/guides/cicd-operations.md:181` and `:185` quote the new lines, in the
  same commit.
* The `release` job's `with:` and permissions are unchanged.

### D3. The refusal is accepted, with no code change

* `update.go` is unchanged. It sets no `TargetPolicy.ExecutablePath`, so a
  program started through a symlink is refused on every platform.
* `ErrSymlinkInvocation` is not matched by the app. `cli.Run` prints the
  library's message, which names the file to run.
* No test is added. The behavior is the library's, covered by its
  `selfupdate/target_test.go` **[read]**. The app's tests already cover the
  installer it builds.

### D4. The README says how to update a linked install

* A section "Changes with go-selfupdate-lib v1.25.0" after the v1.28.0
  section, and its contents line, naming the release Q1 decides. It says:
  `update` started through a symlink is refused on Linux, as on macOS and
  Windows; the refusal names the link and its target, and says to run the
  target; `update --json` reports that refusal's class as
  `symlink-invocation`; the hook, `configure` and the installers behave as
  in `v1.11.0`.
* The "Security Note" under "Self-Update" adds that `update` refuses a
  binary started through a symlink, and to run the file the link points to.
* "Option B: Per-Repository Hook" adds that `update` runs on
  `/path/to/prepare-commit-msg`, not on the link.

### D5. The release (Q1)

A patch release, `v1.11.1`, once the move is on `main` and CI is green (Q1
A). `v1.11.0` is the latest release, and `main` holds only records after it
(`d0da9ae`) **[read]**. The README section D4 adds names `v1.11.1`.

### Consequences

* Good, because every pin is current, and the next move of either library
  starts from its newest tag.
* Good, because `update` treats a symlink the same way on Linux, macOS and
  Windows, and a refused user is told which file to run.
* Good, because the workflow pins and the module name the same version, and
  the publish path they run has published this repository's `v1.11.0`.
* Neutral, because the hook, `configure` and the installers behave as
  before **[probed]** **[derived]**.
* Bad, because a Linux user who updated through a link must now run the
  binary itself. No such user is reported, and the default install has no
  link.
* Bad, because the refusal's message ends with "or set
  TargetPolicy.ExecutablePath to it", a library field this CLI's user cannot
  set. The README's wording carries the part a user can act on. Changing the
  message belongs to the library, and is raised there as go-selfupdate-lib
  issue #10 (see "More Information").

### Confirmation

* `go.mod` and `go.sum` change only on D1's lines. `go mod tidy -diff` is
  empty, and `go mod verify` passes.
* `875dec3c9d475ca47c4b012ca43742de832ef626` occurs twice in `ci.yml` and
  twice in the operations guide, and `74ab5dd376a4c864debb32b2d8abfc7856059ba3`
  in neither.
* `update.go` is unchanged.
* `make verify`, `make preflight`, `make disclosure-check` and
  `make preflight-hosts` pass on the committed `HEAD`. CI is green on
  `main`, its build rehearsal included.
* `make check-records` and `make markdownlint` pass.
* Under Q1 A: the tag run publishes `v1.11.1` through
  `publish-selfupdate-release.yml@875dec3c…`, the release is immutable with
  nine assets, and `update` moves an installed `v1.11.0` to it. On a Linux
  host, `v1.11.1` started through a symlink in a scratch directory refuses
  `update` with `ErrSymlinkInvocation`'s message, naming the link and its
  target, and leaves the target unchanged.

## Pros and Cons of the Options

### A. Move the module and both pins, accept the refusal, and document it

* Good, because it is one module line, two workflow lines, and README text.
* Good, because the library's fix applies as its 0041-MADR intends, and
  matches macOS and Windows.
* Bad, because a Linux install updated through a link stops updating that
  way, and the message's second half names a library field.

### B. As A, but keep updating through a link

* Good, because no install that updated before is refused.
* Bad, because it undoes the library's fix for this app on Linux, and
  replaces through a link on macOS and Windows where `v1.24.0` refused.
* Bad, because it adds Go code and tests to a pin move, and changes the
  update path's policy, which deserves its own record.
* Neutral, because the target must still sit under the user's home
  directory, so a system path is refused either way.

### C. Move the module only

* Good, because CI's build and publish workflows do not change.
* Neutral, because the workflow files and the scripts they run are
  identical at both commits, so the release would build the same way
  **[read]**.
* Bad, because `ci.yml`'s comment says the pins name the version `go.mod`
  requires, and would be false, or rewritten to explain a split.

### D. Hold at `v1.24.0`

* Good, because nothing changes now.
* Bad, because Linux keeps replacing a link's target, which the library now
  calls a bug, and the next move must take this release anyway.

## Owner questions

* **Q1 (D5). Is the move released?**
  * A, recommended: release `v1.11.1`, a patch, once the move is on `main`
    and CI is green. The README section names it. The pruning after it
    follows the operations guide's "Controlled release" under
    `0026-MADR-keep-five-stable-releases.md`, on the owner's word.
  * B: no release now. Commit and push the move. It ships with the next
    release, whose README section names `v1.25.0` too; the section waits
    for it.

**Answered 2026-11-09.** The owner: "decisions, Q1. A, release v1.11.1."
D5 records it. The chosen option, A, was not questioned and stands. The
owner then accepted this record and approved its PLAN, the same day: "set
madr to accepted, approve the plan, proceed".

## Amendments

### A1 (2026-11-09): D5's release waits for the next library releases

*A1 changes when D5's release is made. D5 is left above as written. Found
after PLAN Phase 4, with the move on `main` and CI green.*

* **The owner (2026-11-09):** "2 of the libraries this includes, including
  selfupdate-lib are being updated with enhancements and bugfixes. i am
  going to wait until we can pin latest again before i tag a release. i will
  let you know."
* **Now:** no tag is made for this move until the owner says so. The
  release follows the moves to the next releases of go-selfupdate-lib and
  go-llmprovider-sdk, and ships this move with them.
* **The README section** this record added names `v1.11.1`
  (`README.md`, "Changes with go-selfupdate-lib v1.25.0"). If the release
  that ships this move has another tag, the record that makes that release
  corrects the section.
* **Unchanged:** D1–D4, and option A. `main` carries the pin from
  `121879f`. Until a release, a `go install …@main` build differs from
  `v1.11.0` in this move only.
* **Consequence:** go-selfupdate-lib's Linux symlink fix reaches users with
  the later release, not on its own.

## More Information

* The library's record, cited by repository and filename: go-selfupdate-lib
  `0041-MADR-refuse-symlink-invocation-on-every-platform.md`.
* The refusal's message, raised with the library on 2026-11-09 at the
  owner's request, outside this record's scope: go-selfupdate-lib issue #10,
  "ErrSymlinkInvocation: the refusal names a Go field to CLI users, gives
  wrong advice for an explicit ExecutablePath, and fails
  CleanupPending/KeptBackups"
  (<https://github.com/maccavelli/go-selfupdate-lib/issues/10>). This
  record's decisions do not depend on its outcome.
* The last move of this library:
  [0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md](0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md).
* The workflow-pin rule:
  [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md).
* The release window:
  [0026-MADR-keep-five-stable-releases.md](0026-MADR-keep-five-stable-releases.md).
* The implementation:
  [0030-PLAN-adopt-go-selfupdate-lib-v1-25-0.md](0030-PLAN-adopt-go-selfupdate-lib-v1-25-0.md).
