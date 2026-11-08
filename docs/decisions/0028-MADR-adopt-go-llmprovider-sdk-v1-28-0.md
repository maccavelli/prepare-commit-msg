---
status: accepted
date: 2026-11-07
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its CHANGELOG for v1.28.0, the compare v1.27.0...v1.28.0 by commit and file, its 0071-MADR, its source at v1.28.0, and its CI run on the tag); go-selfupdate-lib's tags; a scratch clone of this repository at 870639c moved to v1.28.0; all 2026-11-07
informed: users of the installed hook, who receive the change through `update` once it is released
---
# Move to go-llmprovider-sdk v1.28.0, whose additions and fixes this app does not reach, with no other change to the code

*Revised 2026-11-07, the day it was drafted. The owner answered Q1: the
move ships in `v1.11.0`, the release of
`0029-MADR-harden-the-hooks-generation-request.md`. D3 is rewritten to
that answer.*

## Context and Problem Statement

The owner (2026-11-07): "i just released a new go-llmprovider-sdk, we need
to assess it, and update our pin. it has a ton of bugfixes."

This repository pins its two libraries in four places **[read]**:

| Pin | Where | Pinned | Newest |
| :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `go.mod:6` | `v1.27.0` | `v1.28.0`, tagged 2026-11-07 |
| go-selfupdate-lib | `go.mod:7` | `v1.24.0` | `v1.24.0` |
| go-selfupdate-lib build workflow | `.github/workflows/ci.yml` | `74ab5dd3…` (`v1.24.0`) | unchanged |
| go-selfupdate-lib publish workflow | `.github/workflows/ci.yml` | `74ab5dd3…` (`v1.24.0`) | unchanged |

Only the first row moves. What has to be decided is whether the module
moves, whether the app adopts the release's two additions, and whether the
move is released.

Each claim is marked:

* **[read]**: from this repository, or from the SDK's changelog, source or
  git objects at the version named;
* **[probed]**: run in the session scratchpad on 2026-11-07, never in this
  tree;
* **[derived]**: reasoned from the two, and not run.

### The release

* `v1.28.0` is an annotated tag on `e98c24832115c5d4419d79d149a58643c9383e90`.
  The SDK's CI passed on that commit. `proxy.golang.org`'s `@latest` and
  `go list -m -versions` name it **[read]** **[probed]**.
* `v1.27.0...v1.28.0` holds 10 commits and changes 24 files. The Go changes
  are in `llmprovider/catalog/` (`cache.go`, `discovery.go`,
  `listing_cache.go`, `lookup.go`, `model_metadata.go`),
  `llmprovider/auth/oauth_session.go`, `llmprovider/internal/transport/probe.go`,
  and a comment in `llmprovider/command_token.go`. `wizard/`,
  `llmprovider/providers/` and `go.mod` are unchanged **[read]**.
* Its GitHub release carries no assets, and is not immutable **[read]**. As
  for `v1.27.0`, the module does not depend on that: `go mod verify` reports
  "all modules verified" **[probed]**.

### What v1.28.0 changes

From its CHANGELOG, its
`0071-MADR-give-a-cache-a-lifetime-and-wait-for-a-refresh.md` and its
source **[read]**. The CHANGELOG lists two additions and two fixes:

* **Added, `catalog.Cache.Close(ctx)`.** It ends a caller-owned cache's
  background work: listing fetches, models.dev refreshes and probe runs.
  Afterwards `List`, `LookupModel` and `LookupMetadataWith` given that cache
  refuse with an error matching `llmprovider.ErrInvalidRequest`. The
  process's cache, used when no `catalog.WithCache` is given, has no `Close`
  and behaves as before.
* **Added, `auth.OAuthSession.Wait(ctx)`.** It waits, without cancelling,
  for a token refresh that runs on after its caller has gone, so that a test
  or an ending process keeps the rotated session.
* **Fixed, the count of two background listing starts.** They counted their
  fetch after releasing their lock, so nothing could reliably wait for them.
  They now count it under the lock. Only a waiter, which is `Close`, sees
  the difference.
* **Fixed, a second listing request.** Two concurrent `List` calls of one key
  could send two listing requests, when the second claimed the key just after
  the first fetch stored its listing. The claim now re-checks for a fresh
  listing. The effect was one extra request, never a wrong answer.

### What reaches this app

* **The process's cache only.** The app passes no `catalog.WithCache`
  (`main.go:376`, `internal/ui/setup.go:159`, `:279`). So no cache it uses
  can be closed. The new closed checks find the process's store open on
  every call. Its probe store's new lifetime context is never cancelled
  **[read]**.
* **No waiter.** Nothing in the app waits on the SDK's background work, so
  the first fix changes nothing it can see **[read]** **[derived]**.
* **One `List` call per run.** `configure` lists a provider's models once,
  through `discoverModels` (`internal/ui/setup.go:158-184`). The hook lists
  nothing. So the app makes no concurrent `List` calls of one key, and the
  second fix changes nothing it sends **[read]** **[derived]**.
* **An `auth.OAuthSession` is the hook's token source** for a signed-in
  provider (`main.go:324-336`). The SDK runs a refresh detached from its
  caller. The hook's process ends once it has written the message. So if the
  hook's overall timeout, 120 s by default (`internal/config/config.go:19`),
  ended while a refresh was in flight, the process could end before the
  rotated session was saved. That needs a refresh that outlasts the whole
  generation budget. It is the same at `v1.27.0`, and this move neither
  causes nor fixes it **[read]** **[derived]**. `Wait` is what an app would
  call to close it, and D2 leaves it out of this move.

No observable behavior of `configure`, the hook, `update` or the installers
changes, and no label, model list or default changes **[read]**
**[derived]**.

### The move, on a scratch clone

On a clone of `870639c` **[probed]**:

* `go get github.com/maccavelli/go-llmprovider-sdk@v1.28.0` ("upgraded
  … v1.27.0 => v1.28.0") and `go mod tidy` change one line of `go.mod` and
  two of `go.sum`. The new hashes are
  `h1:Q1XMpmJzR8apF06whzRscIG2h3dRgLtvtJ65wWoaD34=`, and for `/go.mod`
  `h1:dE+W7oDyvVmAxpSOuIJO3qSD3bDTh/olLt/TbByBE1M=`, the same as
  `v1.27.0`'s. `go mod tidy -diff` is empty, and `go mod verify` prints "all
  modules verified".
* `make verify` exits 0:
  * mod-check, fmt-check, golangci-lint, `go vet`, and
    `go test -race ./...`, `TestReadmeCuratedModels` included;
  * total coverage 83.9%, against a minimum of 80.0%;
  * govulncheck: "No vulnerabilities found.";
  * workflow-lint: "ci filters: 11 patterns, 11 reads checked, 0
    problems";
  * the tools' 42 tests and the scripts' 34;
  * the six release cross-builds.

## Decision Drivers

* **Pins are current** unless a record names a reason to hold one.
* **The change is a pin move.** New library features are separate
  decisions.
* **A release carries something a user would want.** Under
  `0026-MADR-keep-five-stable-releases.md`, each release also moves the
  rollback window on by one.
* **The README says what a user can see.**

## Considered Options

* **A. Move the pin, and keep the app's code as it is.**
* **B. Move the pin, and call `OAuthSession.Wait` before the hook exits.**
  Keep each OAuth session the run loop builds, and wait for it, bounded,
  after the message is written or the run fails.
* **C. Hold at `v1.27.0`.** Record the reason, and move with the SDK's next
  release.

## Decision Outcome

Chosen option: "A", because it keeps the pin current at the cost of one
module line, and nothing in the release reaches the app's behavior. B
answers an exposure that predates this release, and needs a run that
outlasts the hook's whole timeout. It is a code change with its own tests,
and belongs in its own record if the owner wants it. C holds a move the next
one must take anyway, with no reason to hold.

### D1. The module moves

* `go.mod` requires `github.com/maccavelli/go-llmprovider-sdk v1.28.0`,
  after `go mod tidy`.
* `go.mod` changes on that line alone. `go.sum` changes on that module's
  two lines alone. `go mod tidy -diff` is empty, and `go mod verify` passes.
* go-selfupdate-lib, both workflow pins, the operations guide and the Go
  toolchain do not change.

### D2. No other code changes

* Not adopted: `catalog.Cache.Close` and `auth.OAuthSession.Wait`. The app
  owns no `catalog.Cache`, and `Wait` is option B.
* No test is added. The app's tests already cover what reaches it. The two
  fixes are the SDK's, covered by its `0071` tests **[read]**.

### D3. The release and the README (Q1)

No release of its own (Q1 A). The move ships in `v1.11.0`, the release of
`0029-MADR-harden-the-hooks-generation-request.md` (its D9), whose "Changes
with go-llmprovider-sdk v1.28.0" section names this move. This record adds
no README section, because nothing a user sees changes.

### Consequences

* Good, because the pin is current, and the next SDK move starts from
  `v1.28.0`.
* Good, because nothing else moves: one module line, no code, no workflow.
* Neutral, because the hook, `configure`, `update` and the installers behave
  as before **[probed]** **[derived]**.
* Neutral, because the hook's exposure to an OAuth refresh still in flight
  at exit stays as it is at `v1.27.0`.
* Bad, because if Q1 holds the release, `main` carries an unreleased pin
  until the next release. A `go install …@main` build differs from
  `v1.10.4` in this module only.

### Confirmation

* `go.mod` and `go.sum` change only on D1's lines. `go mod tidy -diff` is
  empty, and `go mod verify` passes.
* `grep` finds `74ab5dd376a4c864debb32b2d8abfc7856059ba3` twice in
  `ci.yml` and twice in the operations guide, as before.
* `make preflight`, `make disclosure-check` and `make preflight-hosts` pass
  on the committed `HEAD`. CI is green on `main`, its build rehearsal
  included.
* `make check-records` and `make markdownlint` pass.
* Q1 A: no tag is made for this record. The release checks run for
  `v1.11.0` under `0029-PLAN-harden-the-hooks-generation-request.md`.

## Pros and Cons of the Options

### A. Move the pin, and keep the app's code as it is

* Good, because the change is one module line.
* Neutral, because neither addition is used, and neither fix is reached.

### B. Move the pin, and call `OAuthSession.Wait` before the hook exits

* Good, because a refresh still in flight at exit would be saved before the
  process ends.
* Bad, because it adds Go code and tests to a pin move. The run loop builds
  a session per model (`main.go:321-336`), so each must be kept and waited
  for.
* Bad, because the wait must be bounded, or a slow refresh holds the commit.
  Choosing that bound is its own decision.
* Neutral, because the case it covers needs a refresh that outlasts the
  hook's 120 s default timeout, and no failure of that kind is reported.

### C. Hold at `v1.27.0`

* Good, because nothing changes now.
* Bad, because the next move must take this release anyway, with no reason
  recorded to hold it.

## Owner questions

* **Q1 (D3). Is the move released?**
  * A, recommended: no release now. Commit and push the move. It ships with
    the next release that a user would notice, whose README section names
    `v1.28.0` too. No README section now.
  * B: release `v1.10.5`, a patch, once the move is on `main` and CI is
    green, with a README section, "Changes with go-llmprovider-sdk
    v1.28.0", saying the hook, `configure`, `update` and the installers
    behave as in `v1.10.4`. Under 0026, the pruning after it keeps
    `v1.8.0` to `v1.10.5`, and deletes `v1.7.0` and its tag **[derived]**.
  * C: as B, but `v1.11.0`, to mark the SDK's minor version.

**Answered 2026-11-07.** The owner: "0028 q1: confirm ships in v1.11.0".
That is option A: no release of its own, and the move ships in the next
release, `v1.11.0`, which 0029-MADR's Q4 answer named. D3 records it. The
chosen option, A, was not questioned and stands.

## More Information

* The SDK's record, cited by repository and filename: go-llmprovider-sdk
  `0071-MADR-give-a-cache-a-lifetime-and-wait-for-a-refresh.md`.
* The last move:
  [0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md](0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md).
* The release pruning:
  [0026-MADR-keep-five-stable-releases.md](0026-MADR-keep-five-stable-releases.md).
* The implementation:
  [0028-PLAN-adopt-go-llmprovider-sdk-v1-28-0.md](0028-PLAN-adopt-go-llmprovider-sdk-v1-28-0.md).
