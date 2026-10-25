---
status: complete
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
   `v1.10.2`, ~~the first live run of D6's client,~~ and
   `prepare-commit-msg version` names it. *Deviation 2: that update runs
   `v1.10.1`'s client; `update --check` from `v1.10.2` is D6's first live
   run.*
5. This PLAN becomes `complete`, the index follows, in a records commit
   ~~on a branch, merged as above~~ *on `main`, committed by the owner
   (Deviation 1)*.

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
* Committed by the owner as `6e1fc11`. Its hook-written subject read
  "add v1.22.2 release notes"; the owner amended it, before any push, to
  `7c3cd79`, "add v1.10.2 release notes", with `git commit --amend -F`
  (an amend's source is `commit`, which the hook skips). `7c3cd79`'s tree
  is `6e1fc11`'s.

### Phase 4, 2026-10-24

* `make preflight` on `6e1fc11`: "preflight passed"; total coverage 83.7%
  (minimum 80.0%), "No vulnerabilities found.", the tools' 25 tests "OK".
* `make preflight-hosts` on the committed `HEAD`, `6e1fc11`:

  ```text
  linux: exit=0 (74s)
  wsl-ubuntu-24.04: exit=0 (35s)
  windows: exit=0 (3s)
  macos: exit=0 (4s)
  preflight-hosts: 4/4 hosts passed
  ```

  The host builds stamp `v1.10.1-6-g6e1fc11`. No host was unreachable.
* The disclosure guard, as AGENTS.md writes it, exits 0 with no output
  over `03df92d..6e1fc11`, and again over `03df92d..7c3cd79` after the
  amend.
* The owner pushed `main`. CI run `40149400018` on `7c3cd79`: success.
  "Go (quality contract)", the three native test jobs, "Build release /
  build" (its log: `Uses: …build-selfupdate-release.yml@84805563ed252df384f59ced45a3daf529702ab4`)
  and five identity jobs succeed; "Publish GitHub Release" is skipped, as
  on any push that is not a tag. The identity jobs cover linux/amd64,
  linux/arm64, windows/amd64, windows/arm64 and darwin/arm64, the same
  five as `v1.10.1`'s run `39709052042`; darwin/amd64 has none in either.

### Phase 5, 2026-10-24

* The owner tagged and pushed `v1.10.2` (tag object `68baa4c`, peeling to
  `7c3cd79`), where the PLAN had the agent do it on the owner's ask
  (Deviation 1 covers the owner's git).
* Release run `40150318042`:
  <https://github.com/maccavelli/prepare-commit-msg/actions/runs/40150318042>.
  Every job succeeds; "Publish GitHub Release / publish" success, its log
  `Uses: …publish-selfupdate-release.yml@84805563ed252df384f59ced45a3daf529702ab4`;
  "… / sign" skipped, as `sign` is unset. **The first publish through
  `84805563…`.**
* The release <https://github.com/maccavelli/prepare-commit-msg/releases/tag/v1.10.2>,
  published 2026-10-25T01:25:29Z, not a draft or pre-release, and
  `releases/latest` names it. Assets: the six binaries, `SHA256SUMS`,
  `install.sh`, `install.ps1`; no `SHA256SUMS.note`.
* `prepare-commit-msg-darwin-arm64`, downloaded to the session scratchpad:
  `shasum -a 256 -c` against `SHA256SUMS` "OK"; `gh attestation verify
  --signer-workflow maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml`
  exits 0, one attestation, signer
  `…publish-selfupdate-release.yml@84805563ed252df384f59ced45a3daf529702ab4`,
  source `refs/tags/v1.10.2` at `7c3cd79`.
* This Mac's installed hook was `v1.10.1 (release) 52ad5494ccfe`.
  `update --check`: "update available: v1.10.1 -> v1.10.2", exit 10.
  `update --yes`: "updated v1.10.1 -> v1.10.2", "release asset integrity
  verified", exit 0. `version`: `v1.10.2 (release) 7c3cd796b209`; the
  installed file's size, 9894866, is the asset's.
* D6 live: `update --check` from `v1.10.2`: "up to date (v1.10.2)", exit
  0; `update --check --json`'s last line is `kind` "result", schema 4.

### Deviation 2, 2026-10-24: the `v1.10.1` → `v1.10.2` update did not run D6's client

* **Found.** Phase 5 step 4 called that update "the first live run of
  D6's client". The update is performed by the running binary,
  `v1.10.1`, whose client is the old one.
* **Resolution.** D6's client was run live by `update --check` from the
  installed `v1.10.2` (Phase 5 above). The MADR's Confirmation is
  corrected by its Amendment A2. A slow or stalled GitHub reply, which is
  what D6 changes, was not provoked live; the unit test holds the
  client's shape.
* **Files added to scope:** none.
