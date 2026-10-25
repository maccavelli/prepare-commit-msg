---
status: in-progress
date: 2026-10-24
associated-madr: "0021-MADR-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md"
---
# Implement the move to go-llmprovider-sdk v1.22.2 and go-selfupdate-lib v1.22.0

Associated MADR: [0021-MADR-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md](0021-MADR-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md)

*Revised 2026-10-24, before execution, for the owner's answers: Q1 "adopt
now" adds Phase 2b (MADR D6); Q2 "yes, we will release v1.10.2" makes
Phase 5 unconditional.*

## Goal

* **D1:** `go.mod` requires `go-llmprovider-sdk v1.22.2` and
  `go-selfupdate-lib v1.22.0`.
* **D2:** both reusable-workflow lines in `ci.yml` and their quotes in the
  operations guide pin `84805563ed252df384f59ced45a3daf529702ab4 #
  v1.22.0`, with the comment rewritten.
* **D6:** `update` passes `selfupdate.DefaultHTTPClient()`, with a test
  that holds it.
* **D4:** the README states what a user can see.
* **D5:** `v1.10.2` is released and checked.

## Scope

### In scope

* `go.mod`, `go.sum`.
* `.github/workflows/ci.yml`: the comment above the build job
  (`ci.yml:99-105`) and the two `uses:` lines (`:106`, `:121`).
* `docs/guides/cicd-operations.md:162`, `:166`.
* `update.go` (`:3-6`, `:54-55`, `:69`) and `migration_test.go` (its
  imports and one new test after `TestDefaultNewUpdateUpdater`) (D6).
* `README.md`: a "Changes with …" section and its table-of-contents line.
* `docs/README.md`: this pair's index rows, and their status as it moves.

### Out of scope

* Any Go source or test beyond D6's two files.
* D3's candidates: the new `wizard` options and the `service`, `launchd`
  and `taskscheduler` additions.
* A `Timeout` in `update`'s `cli.Options`: `cli.Run`'s 15-minute default
  stays.
* `golang.org/x/*` and the Go pins, which neither release moves.
* `selfupdate-release.json`.
* The text of records `0001` to `0020`, and earlier README "Changes with"
  sections, which are history.

## Implementation Steps

~~All work is on a branch, `deps/0021-pin-latest-libraries`, cut from
`main` at the commit that carries this pair. Nothing is committed on
`main`.~~ *Superseded by Deviation 1 (2026-10-24): all work is on `main`,
and the owner commits and pushes.* Each phase ends with `make
check-records` and `make markdownlint`, `make pre-add-check` when Go files
or `go.mod`/`go.sum` are staged, then the phase is staged for the owner's
`git commit`.

### Phase 0: preconditions (read-only)

1. `git rev-parse --path-format=absolute --git-path hooks` names the
   global hooks directory.
2. `go list -m -versions` still ends at `v1.22.2` (go-llmprovider-sdk)
   and `v1.22.0` (go-selfupdate-lib), and `git ls-remote --tags` peels
   go-selfupdate-lib `v1.22.0` to `84805563…`. If either is newer, stop
   and amend the MADR.

### Phase 1: the records

1. This pair and `docs/README.md`'s two index rows. The MADR is
   `accepted`, with Q1 and Q2's answers recorded (2026-10-24).
2. On approval: this PLAN's `status` becomes `in-progress`; the index
   follows.
3. Create the branch and commit the records alone (bootstrap exception).

### Phase 2: modules and workflow pins (D1, D2)

One commit, because `TestCurrentDocumentation` fails when `ci.yml` and the
guide disagree (MADR, "What the move changes").

1. `go get github.com/maccavelli/go-llmprovider-sdk@v1.22.2
   github.com/maccavelli/go-selfupdate-lib@v1.22.0`, then `go mod tidy`.
   `git diff go.mod go.sum` shows only those two modules' lines.
2. In `ci.yml` and `docs/guides/cicd-operations.md`, replace
   `036289ecc701d84c7749f4950a57a154c1af2cf4 # v1.17.2` with
   `84805563ed252df384f59ced45a3daf529702ab4 # v1.22.0`, asserting two
   occurrences in each before the replace.
3. Rewrite `ci.yml:99-105`: both workflows are pinned to the commit of
   go-selfupdate-lib v1.22.0, the version `go.mod` requires; the publish
   path is unchanged from v1.17.2's, which published live (this
   repository's v1.10.1, run 39709052042), so 0010-MADR D2's rule, as
   amended, holds; cite this MADR's D2 and 0010-MADR D2, A1.
4. Check: in each of the two files, `grep -c
   84805563ed252df384f59ced45a3daf529702ab4` gives 2 and `grep -c
   036289ecc701d84c7749f4950a57a154c1af2cf4` gives 0; `go mod tidy -diff`
   is empty; `make verify` exits 0.

### Phase 2b: `update`'s client (D6)

1. `update.go`: drop the `time` import; replace `updateTimeout` and its
   comment (`:54-55`) with `updateHTTPClient()`, returning
   `selfupdate.DefaultHTTPClient()`, its comment citing the MADR's D6;
   `Client: updateHTTPClient(),` at `:69`.
2. `migration_test.go`: import `net/http` after `io`; add
   `TestUpdateHTTPClientHasNoOverallTimeout` after
   `TestDefaultNewUpdateUpdater`, as the MADR's D6 describes.
3. Each replacement asserts one occurrence before it is made; the whole
   diff is read afterwards.
4. Check: `gofmt -l` and `golint` on both files are clean; `grep -rn
   updateTimeout --include='*.go' .` finds nothing; `make verify` exits 0.
   The new test's failure on the old client was seen in the MADR's probe;
   repeat it on a scratch copy of the branch, never in the tree, and
   record the output.

### Phase 3: the README (D4)

1. Add "Changes with go-llmprovider-sdk v1.22.2 and go-selfupdate-lib
   v1.22.0" after the `v1.10.1` section, in its shape, with the bullets the
   MADR's D4 lists, `update`'s included. Its first line is "Release
   `v1.10.2` moves …", and cites the MADR.
2. Add its table-of-contents line after README.md:33.
3. Check: `make markdownlint`; `TestCurrentDocumentation` and
   `TestReadmeCuratedModels` pass.

### Phase 4: before a push

1. `make preflight`.
2. `make preflight-hosts` on the committed `HEAD` (Go and `go.mod`
   changed). A host that cannot be reached is reported, not skipped.
3. The disclosure guard over the outgoing commits, ~~with the branch ref
   in place of `refs/heads/main` in~~ *as written in* AGENTS.md's command
   (Deviation 1).
4. ~~Ask the owner to push the branch and open a pull request into
   `main`. CI must be green on the branch, its build rehearsal included.
   The owner merges.~~ *Deviation 1:* the owner pushes `main`; CI must be
   green on `main`, its build rehearsal included.

### Phase 5: the release `v1.10.2` (D5)

1. After the merge, CI on `main` is green.
2. The owner asks for the tag in that turn; `v1.10.2` is tagged on
   `main`'s merge commit and pushed.
3. Record the run: its URL, that it published through `84805563…`, and the
   assets (six binaries, `SHA256SUMS`, `install.sh`, `install.ps1`).
4. `prepare-commit-msg update` on an installed `v1.10.1` reaches
   `v1.10.2`, the first live run of D6's client, and
   `prepare-commit-msg version` names it.
5. This PLAN becomes `complete`, the index follows, in a records commit
   on a branch, merged as above.

## Verification

* Phase 2: `make verify` exits 0, as in the MADR's probe; the `grep -c`
  counts hold.
* Phase 2b: `TestUpdateHTTPClientHasNoOverallTimeout` passes, and fails on
  a scratch copy with the old client; `make verify` exits 0.
* Phase 3: `make markdownlint` and `make check-records` exit 0.
* Phase 4: `make preflight` and `make preflight-hosts` exit 0; the
  disclosure guard reports nothing.
* CI green on the branch and on `main`.
* Phase 5, if run: the release run succeeds and `update` reaches it.

## Rollout and Rollback

* **Rollout:** through the pull request, then the `v1.10.2` tag; users
  take it with `update`.
* **Rollback before merge:** close the pull request; `main` is untouched.
* **Rollback after merge:** revert Phase 2b's commit alone if only
  `update`'s client misbehaves, or Phases 2b and 2 together, on a branch,
  and merge it. If a release shipped, tag a new patch from the revert.
* **If the publish run fails** at `84805563…`: repin the publish workflow
  to `036289ec…`, amend the MADR, and tag a new patch.

## Execution record

### Deviation 1, 2026-10-24: no branch; the owner commits and pushes on `main`

* **Found.** The owner committed this pair on `main` as `03df92d`, before
  Phase 1 step 3's branch existed. On the question of a branch, the owner
  answered: "i will commit and push you will not use branches.
  sysadmin-git-branch-gates is the mandate here", and chose that the owner
  commits each phase.
* **Decision.** No branch is created. Each phase is checked and staged on
  `main` by the agent and committed and pushed by the owner, whose global
  `prepare-commit-msg` hook writes the message. Phase 4's disclosure guard
  runs as AGENTS.md writes it; there is no pull request. Phase 5 tags
  `main`. The MADR's driver on `main` is amended (MADR A1).
* **Files added to scope:** none.

### Phase 0, 2026-10-24

* `git rev-parse --path-format=absolute --git-path hooks` names the global
  hooks directory.
* `go list -m -versions` ends at `v1.22.2` (go-llmprovider-sdk) and
  `v1.22.0` (go-selfupdate-lib); `git ls-remote --tags` gives
  `v1.22.0^{}` = `84805563ed252df384f59ced45a3daf529702ab4`.

### Phase 1, 2026-10-24

* The pair and its index rows were committed by the owner as `03df92d`,
  on `main` (Deviation 1).

### Phase 2, 2026-10-24

* `go get …@v1.22.2 …@v1.22.0` and `go mod tidy` exit 0. `go.mod` changes
  the two `require` lines, `go.sum` those modules' four lines; `go mod
  tidy -diff` is empty.
* The repin script asserted 2 occurrences of the old ref in `ci.yml` and
  in the operations guide before replacing them. Afterwards, in each:
  `84805563…` 2, `036289ec…` 0. The comment at `ci.yml:99-105` was
  rewritten as step 3 says; the whole diff was read.
* `make verify` exits 0: total coverage 83.6% (minimum 80.0%), "No
  vulnerabilities found.", workflow-lint through actionlint, the tools'
  25 tests "OK", and the six cross-builds.
* Committed by the owner as `af7fb4e`.

### Phase 2b, 2026-10-24

* The edit script asserted one occurrence of each replaced text in
  `update.go` (the `time` import, `updateTimeout` and its comment, the
  `Client:` line) and of the import anchor and test anchor in
  `migration_test.go`; the whole diff was read. `gofmt -l` on both files
  prints nothing; `grep -rn updateTimeout --include='*.go' .` finds
  nothing.
* On a scratch copy of the tracked tree with `updateHTTPClient` returning
  `&http.Client{Timeout: 15 * time.Minute}`, the new test fails:

  ```text
  --- FAIL: TestUpdateHTTPClientHasNoOverallTimeout (0.00s)
      migration_test.go:210: Client.Timeout = 15m0s, want 0
  ```

  In the tree it passes.
* `make verify` exits 0: total coverage 83.6% (minimum 80.0%), "No
  vulnerabilities found.", the tools' 25 tests "OK", and the six
  cross-builds.
* Committed by the owner as `c1347a2`.

### Phase 3, 2026-10-24

* `README.md` gains "Changes with go-llmprovider-sdk v1.22.2 and
  go-selfupdate-lib v1.22.0" after the `v1.10.1` section, opening
  "Release `v1.10.2` moves …" and citing the MADR, with the MADR's D4
  bullets for `configure`, `update` and the installers, and its
  table-of-contents line after the `v1.10.1` one.
* `make markdownlint` exits 0, with "0 issues", the link-fragment rule
  included; `TestCurrentDocumentation`,
  `TestCurrentDocumentationDetectsBrokenFixtures` and
  `TestReadmeCuratedModels` pass.
