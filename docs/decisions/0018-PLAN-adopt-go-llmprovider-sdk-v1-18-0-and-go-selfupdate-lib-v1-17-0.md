---
status: in-progress
date: 2026-10-21
associated-madr: "0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md"
---
# Implement the move to Go 1.27.3, go-llmprovider-sdk v1.21.0 and go-selfupdate-lib v1.17.2

Associated MADR: [0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md](0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md)

*Revised 2026-10-21, before approval, twice: Phase 1, Go 1.27.3 (D6), is
added first and the phases after it renumbered; then the targets move to
go-llmprovider-sdk `v1.21.0` and go-selfupdate-lib `v1.17.2`, with the
workflow pins of the MADR's option A. The filename keeps its first slug.*

## Goal

* **D6:** the repository is pinned to Go 1.27.3: `go.mod`, `Makefile` and
  `scripts/bootstrap-tools.sh`.
* **D1:** `go.mod` requires `go-llmprovider-sdk v1.21.0` and
  `go-selfupdate-lib v1.17.2`.
* **D2:** `ci.yml`'s build workflow is pinned to
  `036289ecc701d84c7749f4950a57a154c1af2cf4` (`v1.17.2`) and its publish
  workflow to `e8d4904a09194d99761042a8996983e8d5b957eb` (`v1.17.1`), with
  a comment that says why and cites the MADR. If the owner answers Q1 "B",
  both are `036289ec…`. *Superseded by MADR Amendment A1 (2026-10-21):
  both are `036289ec…`. Phase 2a makes the change.*
* **D4:** the README states what a user can see.
* **D5:** if the owner answers Q3 yes, `v1.10.1` is released and checked.

## Scope

### In scope

* `go.mod`, `go.sum`; `Makefile:14`; `scripts/bootstrap-tools.sh:10`.
* `.github/workflows/ci.yml`: the two `uses:` lines and the comment above
  the build job (`ci.yml:99-105`, `:120`).
* `README.md`: a "Changes with …" section and its table-of-contents line;
  "Building from source needs Go …" (`README.md:705`).
* `docs/README.md`: this pair's index rows, and their status as it moves.

### Out of scope

* Any Go source or test. The MADR's probes needed none changed.
* D3's candidates: `catalog.WithCache`, `providers.DefaultWith`,
  `WithStreamRetry`, the new `wizard` entry point and options,
  `NewGitHubUpdater`, release signing, the `before_uninstall` hook.
* `selfupdate-release.json`, which declares no hooks.
* Moving the publish pin to `v1.17.2`'s commit under option A: a separate
  change, once a publish through that commit is on record (MADR D2).
* `golang.org/x/*`, which neither release moves.
* The Go text of records `0001` to `0017`, and the README's `v1.8.0`
  section (`README.md:572-578`), which are history.
* The hosts' toolchains: all three run `go1.27.3` already (MADR, "The
  toolchain"). *Corrected by MADR Amendment A2: four hosts, and WSL was
  on go1.27.2. It was moved under magic-cli-remote's 0169-PLAN, not
  here.*
* `0019-PLAN-skip-full-ci-for-docs-only-pushes.md`'s files and order. It
  waits for this pair's records commit and its `ci.yml` commit, both
  named below.

## Rules for every phase

* **Branch.** On `main`, as 0016-PLAN and 0017-PLAN did: the agent stages,
  the owner commits with `git commit --no-edit` and pushes. One commit per
  phase; the next phase starts after the owner has committed the last.
* **Pre-add,** before each stage:
  * `make check-records`;
  * `make markdownlint`;
  * with `go.mod` / `go.sum` staged, `make pre-add-check`.

  A file that fails is not staged.
* **Phase gate** for Phases 1 and 2: `make preflight`. Each exit status is
  captured to a scratch file before any filter.
* **Edits** to `go.mod`, `Makefile`, the script and `ci.yml` are made by a
  Python script that asserts each old string occurs exactly once before it
  replaces it, and that the new string is present after.
* **No new test or gate.** The phases add none. Phase 1 still shows its
  two existing gates failing on the old pin first, on a scratch clone
  (step 1). If a phase finds it needs a new gate, that is a deviation.
* **Deviations.** Stop, record a dated entry here (and an amendment in the
  MADR when a decision or a fact changes), and wait for approval.

## Implementation Steps

### Phase 0: records

1. The owner reviews the MADR, answers Q1–Q3, and accepts it, or chooses
   another option. An answer that changes a decision becomes a MADR
   amendment, and this PLAN is amended to match, before Phase 1.
2. The owner approves this PLAN. The MADR's status becomes `accepted` and
   this PLAN's `in-progress`; `docs/README.md`'s rows follow.
3. Stage the pair and `docs/README.md`. The owner commits. This is the
   "records committed" point 0019-PLAN waits for.

### Phase 1: Go 1.27.3 (D6)

1. **Fail first, on a scratch clone at the Phase 0 commit,** never in this
   tree: `make tools` exits non-zero with `expected go1.27.2, got
   go1.27.3`, and `.tools/bin/govulncheck ./...` under
   `GOTOOLCHAIN=go1.27.2` exits 3 naming GO-2026-6809. The MADR's probe saw
   both on 2026-10-21; this re-runs them on the commit being changed.
2. The three edits:
   * `go.mod`: `go 1.27.2` → `go 1.27.3`;
   * `Makefile`: `MOD_VERSION := 1.27.2` → `MOD_VERSION := 1.27.3`;
   * `scripts/bootstrap-tools.sh`: `GO_VERSION="go1.27.2"` →
     `GO_VERSION="go1.27.3"`.
3. **Assert:** `go mod tidy -diff` is empty; `git diff --numstat` lists
   exactly those three files, `1 1` each; `grep -rn -E '1\.27\.2'
   go.mod Makefile scripts .github` finds nothing.
4. **Gate:** `make preflight` exits 0, its govulncheck printing "No
   vulnerabilities found."
5. Pre-add checks; stage `go.mod`, `Makefile`,
   `scripts/bootstrap-tools.sh`. The owner commits.

### Phase 2: the library pins (D1, D2)

1. **Re-read the releases.** `go list -m -versions` for both modules and
   `git ls-remote --tags` for go-selfupdate-lib. If either has released
   past `v1.21.0` / `v1.17.2`, or a publish through `036289ec…` is now on
   record, stop: that is a deviation, and the MADR is amended first.
2. `go get github.com/maccavelli/go-llmprovider-sdk@v1.21.0
   github.com/maccavelli/go-selfupdate-lib@v1.17.2`, then `go mod tidy`.
3. **Assert the module diff:** `git diff --numstat go.mod go.sum` is
   `2 2 go.mod` and `4 4 go.sum`, and every changed line names one of the
   two modules; the `go` line stays `go 1.27.3`. `go mod tidy -diff` is
   empty.
4. **The workflow pins:**
   * the build line's `@436a7f5976a6815491c7384527fef90869574a1b # v1.15.2`
     becomes `@036289ecc701d84c7749f4950a57a154c1af2cf4 # v1.17.2`;
   * the publish line's becomes
     `@e8d4904a09194d99761042a8996983e8d5b957eb # v1.17.1` (option A), or
     `036289ec… # v1.17.2` if Q1 is "B";
   * the comment at `ci.yml:99-104` is replaced by one that says: the
     build workflow is pinned to go-selfupdate-lib `v1.17.2`'s commit, the
     version `go.mod` requires; the publish workflow to `v1.17.1`'s, the
     newest commit with a live publish (magic-cli-remote `v0.22.2`, run
     `39635656900`, with `sign` unset as here), until a publish through
     `v1.17.2`'s is on record; and it cites this MADR and
     `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2 and A1. Under "B",
     it says both are `v1.17.2`'s and why, per the MADR.
5. **Assert the workflow edit landed:** `grep -c
   036289ecc701d84c7749f4950a57a154c1af2cf4 .github/workflows/ci.yml` is
   `1` and `grep -c e8d4904a09194d99761042a8996983e8d5b957eb …` is `1`
   (option A; `2` and `0` under "B"), and `grep -c 436a7f59 …` is `0`.
   Read the whole edited block.
6. **Gate:** `make preflight` exits 0. It runs `make verify`, whose
   `workflow-lint` lints `ci.yml`.
7. Pre-add checks; stage `go.mod`, `go.sum`, `.github/workflows/ci.yml`.
   The owner commits. This is the "`ci.yml` commit" 0019-PLAN waits for.

### Phase 2a: both workflows at `v1.17.2` (MADR A1)

Added 2026-10-21, after Phase 2's commit `c57e9de`, by the owner's
"option 1".

1. The MADR gains Amendment A1 and this PLAN gains this phase, in the
   same commit as step 2.
2. **The edit** is made by a script that checks each step:
   * the publish line's
     `@e8d4904a09194d99761042a8996983e8d5b957eb # v1.17.1` becomes
     `@036289ecc701d84c7749f4950a57a154c1af2cf4 # v1.17.2`;
   * the comment above the build job is replaced by one that says both
     workflows are pinned to `v1.17.2`'s commit, the version `go.mod`
     requires. The publish path is unchanged from `v1.17.1`'s, which
     has published live (magic-cli-remote `v0.22.2`, run
     `39635656900`). The comment cites this MADR's A1 and
     `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2 and A1.
3. **Assert:**
   * `grep -c 036289ecc701d84c7749f4950a57a154c1af2cf4
     .github/workflows/ci.yml` is `2`;
   * `grep -c e8d4904a .github/workflows/ci.yml` is `0`;
   * `git diff --numstat` names only `ci.yml` and this pair;
   * the whole edited block is read back.
4. **Gate:** `make preflight` exits 0.
5. Pre-add checks; stage `ci.yml` and the pair. The owner commits.

### Phase 3: the README (D4)

1. `README.md` gains, after "Changes with go-llmprovider-sdk v1.15.0", a
   section "Changes with Go 1.27.3, go-llmprovider-sdk v1.21.0 and
   go-selfupdate-lib v1.17.2", in the shape of the two before it. Its first
   line names the release (`v1.10.1` if Q3 is yes, otherwise "the next
   release") and links the MADR. It lists:
   * **Built with Go 1.27.3,** which fixes GO-2026-6809 in `crypto/x509`;
     `v1.10.0` was built with Go 1.27.2;
   * **Retries:** a failure a provider reports inside a successful stream
     is retried once, whatever `retry_count` is; other failures as before;
   * **Wizard messages:** a refused key, a listing warning or a Kilo
     profile notice shows the error's class and HTTP status, not the
     service's message; a failed browser or device sign-in, an expired
     device code, and an unreachable Ollama server read with their kind
     first;
   * **The installer:** an `--uninstall` that cannot remove a running copy
     says so and exits 3; the usage lists `--force` (`install.ps1`:
     `-Force`), which applies only to a `before_uninstall` hook, and this
     app has none;
   * **`update`** behaves as in `v1.10.0`.
2. The table of contents (`README.md:31-32`) gains the section's line.
3. `README.md:705`: "Building from source needs Go 1.27.2." → "… Go
   1.27.3."
4. `make markdownlint` and `go test ./...` (`readme_test.go`) pass.
5. Pre-add checks; stage `README.md`. The owner commits.

### Phase 4: before the push

1. `make preflight` on the committed `HEAD`.
2. `make preflight-hosts`: required, since Go's version, `go.mod` /
   `go.sum`, the `Makefile`, a script and a workflow changed. Each host's
   output shows `go1.27.3`. A host that cannot be reached is reported,
   never skipped silently.
3. The disclosure guard from `AGENTS.md` over the outgoing commits.
4. The agent asks the owner to push. CI on `main` is green, its setup-go
   steps log Go 1.27.3, and its build rehearsal runs through the `v1.17.2`
   build workflow.

### Phase 5: release `v1.10.1` (D5; only if Q3 is yes)

1. The owner tags `v1.10.1` and pushes the tag. CI's `release` job
   publishes through `036289ec…` (MADR A1).
2. **The release, checked by the agent:**
   * the build job's setup-go logs Go 1.27.3;
   * the run's `sign` job is skipped and `publish` succeeds;
   * the assets are the six binaries, `SHA256SUMS`, `install.sh` and
     `install.ps1`, and no `SHA256SUMS.note`; `shasum -a 256 -c
     SHA256SUMS` passes on the downloaded binaries;
   * `go version -m` on a downloaded binary reports `go1.27.3`;
   * the downloaded `install.sh --help` lists `--force`;
   * the run's URL is recorded as the first publish through
     `036289ec…` (MADR A1).
3. **`update`, the owner present,** on a scratch copy of `v1.10.0`'s
   binary under a scratch directory; the installed hook is not touched:
   `update --check` reports `v1.10.1`, `update --yes` exits 0, and
   `version` prints `v1.10.1 (release) …`.
4. The run URL and each result go into the Execution Record.

## Verification

* Phase 1: step 1's two failures; step 3's tidy diff, numstat and grep;
  `make preflight`'s exit status, its "No vulnerabilities found." and its
  last line, `preflight passed`.
* Phase 2: step 1's version listing; step 3's numstat and `go mod tidy
  -diff`; step 5's counts; `make preflight`'s exit status and last line.
* Phase 3: `make markdownlint` and `go test ./...` exit 0.
* Phase 4: `make preflight`, `make preflight-hosts` per host, the guard's
  exit status, and the CI run's conclusion and Go version.
* Phase 5: the run's job results, the asset list, the checksum check, the
  binary's Go version, and the `update` and `version` output.

Each is recorded in the Execution Record as output, not as a summary.

## Rollout and Rollback

* **Rollout** is the push (Phase 4), then the tag (Phase 5). The hook's
  users get the change through `update` only once a release exists; until
  then, the released `v1.10.0` keeps GO-2026-6809.
* **Before a tag,** each phase reverts on its own: Phase 2's commit
  restores both modules and both workflow pins together, and Phase 3's
  README section goes with it. Phase 1's revert would bring back the
  advisory and the failing tools check, so it is not a rollback this PLAN
  offers; a problem there is a deviation.
* **If the `v1.10.1` publish fails:** the tag has no release. Stop and
  record a deviation. The resolutions to offer: under "B", move the publish
  pin to `e8d4904a…` (option A) and publish a new patch tag; under A, fix
  forward to a go-selfupdate-lib commit that publishes. A published tag is
  not moved or deleted without the owner's ask.

## Execution Record

The MADR's probes ran on scratch clones, not this tree; their results
are in the MADR's Context.

### Phase 0, 2026-10-21

* The owner approved: "pin the current releases of go-selfupdate-lib
  and go-llmprovider-sdk as well as changing your go.mod and toolchain
  to go 1.27.3 … we are going to update". Q1 is A, and Q2 leaves
  `NewGitHubUpdater` for later. Q3 is open (the MADR's "Owner
  questions").
* Before staging, the releases were re-read: `go list -m -versions`
  ends at `v1.21.0` and `v1.17.2`, and the tags `v1.17.1^{}` and
  `v1.17.2^{}` peel to `e8d4904a…` and `036289ec…`, as the MADR says.
* The MADR is `accepted`, this PLAN is `in-progress`, and the index rows
  match. The pair and `docs/README.md` are staged. The untracked 0019
  pair is not.
* The owner committed Phase 0 as `75cb0c3`. The 0019 records followed as
  `6eb837a`.

### Phase 1, 2026-10-21

* **Step 1, fail first,** on a scratch clone at `75cb0c3`, never in this
  tree:
  * `make tools` exited 2: `expected go1.27.2, got go1.27.3`.
  * `govulncheck ./...` (`v1.8.0`, built and run under
    `GOTOOLCHAIN=go1.27.2`) exited 3. It reported `Vulnerability #1:
    GO-2026-6809`, `Found in: crypto/x509@go1.27.2`, `Fixed in:
    crypto/x509@go1.27.3`.
* **Step 2:** the three edits were made by a script that checks each one.
  It found each old string once and the new string after the edit.
* **Step 3:** `go mod tidy -diff` exited 0 with empty output.
  `git diff --numstat` showed `1 1` each for `Makefile`, `go.mod` and
  `scripts/bootstrap-tools.sh`, and nothing else. `grep -rn -E '1\.27\.2'
  go.mod Makefile scripts .github` exited 1, finding nothing.
* **Step 4:** `make preflight` exited 0. `make tools` rebuilt
  golangci-lint, govulncheck and actionlint under go1.27.3, govulncheck
  printed `No vulnerabilities found.`, no test failed, and the last line
  was `preflight passed`.
* **Step 5:** with the files staged, `make check-records`,
  `make markdownlint` and `make pre-add-check` each exited 0. The last
  printed `0 issues.` and `No vulnerabilities found.`.
* The owner committed Phase 1 as `4f9a175`.

### Phase 2, 2026-10-21

* **Step 1, the releases re-read:**
  * `go list -m -versions` ends at `v1.21.0` (go-llmprovider-sdk) and at
    `v1.17.2` (go-selfupdate-lib).
  * `git ls-remote --tags`: `v1.17.1^{}` is `e8d4904a…` and `v1.17.2^{}`
    is `036289ec…`. No tag is newer.
  * The owner had published `v1.17.2`'s GitHub Release (no assets) that
    day. It ran no workflow, so it is not a publish through `036289ec…`'s
    publish workflow.
  * The fleet's consumers still pin the publish workflow at `e8d4904a…`
    (magic-cli-remote) and `adc2c757…` (gobble-cli). No publish through
    `036289ec…` is on record. Option A stands.
* **Steps 2–3:** `go get` reported `go-llmprovider-sdk v1.15.0 =>
  v1.21.0` and `go-selfupdate-lib v1.15.2 => v1.17.2`, and `go mod tidy`
  exited 0. `git diff --numstat` showed `2 2 go.mod` and `4 4 go.sum`,
  every changed line naming one of the two modules. `go.mod` keeps
  `go 1.27.3`. `go mod tidy -diff` exited 0 with empty output.
* **Steps 4–5:** the pin edits were made by a script that checks each one,
  and the whole block was re-read. `grep -c` on `ci.yml`: `036289ec…` 1,
  `e8d4904a…` 1, `436a7f59` 0.
* **Step 6:** `make preflight` exited 0. Lint showed `0 issues.`, total
  coverage was `83.7% (minimum 80.0%)`, and govulncheck printed `No
  vulnerabilities found.`. `workflow-lint` passed on `ci.yml`, and the
  last line was `preflight passed`.
* **Step 7:** with the files staged, `make check-records`,
  `make markdownlint` and `make pre-add-check` each exited 0.
* The owner committed Phase 2 as `c57e9de`.

### Deviation, 2026-10-21: the publish pin moves to `v1.17.2`

* **Evidence.** The owner asked why the two pins differed. The commits
  were compared file by file on GitHub. Every part of `036289ec…`'s
  publish path is the same blob as at `e8d4904a…`: the workflow, the four
  scripts and `internal/cmd/selfupdate-note`. Only test files and
  `go.mod`'s `go` and `toolchain` lines differ. The MADR's Context had
  said the scripts changed in comments and a usage string. That was
  wrong.
* **Resolution chosen by the owner** ("option 1"): `go.mod`'s toolchain
  lines are not part of A1's publish path, so 0010-MADR A1 gives
  `036289ec…` for both workflows (MADR Amendment A1). Phase 2a was added
  to make the change.

### Phase 2a, 2026-10-21

* **Step 1:** the MADR gained Amendment A1, plus pointers in D2 and the
  owner answers, and this PLAN gained Phase 2a. Phase 5 now names
  `036289ec…`.
* **Steps 2–3:** the edit was made by a script that checks each step, and
  the whole block was re-read. `grep -c` on `ci.yml`: `036289ec…` 2,
  `e8d4904a` 0. `git diff --numstat` named `ci.yml` (`7 8`) and this pair
  only.
* **Step 4:** `make preflight` exited 0. Lint showed `0 issues.`, total
  coverage was `83.7% (minimum 80.0%)`, govulncheck printed `No
  vulnerabilities found.`, and the last line was `preflight passed`.
* **Step 5:** with the files staged, `make check-records`,
  `make markdownlint` and `make pre-add-check` each exited 0.
* The owner committed Phase 2a as `0d50074`.

### Phase 3, 2026-10-21

* **Step 1:** `README.md` gained "Changes with Go 1.27.3,
  go-llmprovider-sdk v1.21.0 and go-selfupdate-lib v1.17.2" after the
  `v1.15.0` section. Q3 is open, so its first line says "The next
  release".
  * Before the installer bullet was written, the installers it describes
    were read at `036289ec…`. Both print `could not remove <path>` and
    exit 3 (`install.sh`'s uninstall function; `install.ps1`'s
    `-Uninstall` branch, `return 3`).
  * `install.ps1` at `436a7f59…` (`v1.15.2`) has no such check.
* **Step 2:** the table of contents gained the section's line, anchored
  `#changes-with-go-1273-go-llmprovider-sdk-v1210-and-go-selfupdate-lib-v1172`.
* **Step 3:** "Building from source needs Go 1.27.3." The remaining
  `Go 1.27.2` hits in `README.md` are the `v1.8.0` section, which is
  history, and the new section's own statement that `v1.10.0` was built
  with it.
* **Step 4:** `make markdownlint` exited 0 (`0 issues`), and
  `go test ./...` exited 0 across all five packages, `readme_test.go`
  included.
* **Step 5:** with the files staged, `make check-records`,
  `make markdownlint` and `make pre-add-check` each exited 0.
* The owner committed Phase 3 as `f75153e`.

### Phase 4, 2026-10-21

* **Step 1:** `make preflight` on `f75153e`, a clean tree, exited 0.
  Lint showed `0 issues.`, total coverage was `83.7% (minimum 80.0%)`,
  govulncheck printed `No vulnerabilities found.`, and the last line was
  `preflight passed`.
* **Step 2, first run:** `make preflight-hosts` exited 2. It reported
  `linux: exit=0`, `wsl-ubuntu-24.04: exit=2`, `windows: exit=0`,
  `macos: exit=0`, and `3/4 hosts passed`. WSL's log ends
  `expected go1.27.2, got go1.27.3`.

### Deviation, 2026-10-21: the WSL test host was on go1.27.2

* **Evidence.** `go version` and `go env` were read on each host,
  changing nothing:
  * Linux, Windows and this Mac report `go1.27.3`.
  * WSL reports `go version go1.27.2 linux/amd64`, with `GOTOOLCHAIN`
    `go1.27.2` and `GOROOT` `~/sdk/go1.27.2`.
  * The failure was there before this PLAN. magic-cli-remote's 0169-PLAN
    P14 moved two test hosts, not the WSL distro. This PLAN's Scope said
    all hosts were on go1.27.3 (corrected by MADR A2).
* **Resolution (owner): "Move it under 0169",** keeping go1.27.2
  installed. Declined: recording the move in this pair, and the owner
  moving it by hand.
  * The move is magic-cli-remote 0169-PLAN's Deviation 21, which is
    written there and left for the owner to commit in that repository.
  * No file in this repository is added to the phase.
* **Q3** was answered at the same time: yes, release `v1.10.1` (MADR A2).

### Phase 4, continued

* **The WSL move,** per 0169-PLAN Deviation 21: a dated backup; the
  archive's checksum `OK`; `~/sdk/go1.27.3`; one PATH entry swapped in
  `devenv.sh`; `GOTOOLCHAIN=go1.27.3`. A fresh login shell reported
  `go version go1.27.3 linux/amd64`.
* **Step 2, second run,** on `f75153e`: `make preflight-hosts` exited 0.
  It reported `linux: exit=0 (29s)`, `wsl-ubuntu-24.04: exit=0 (67s)`,
  `windows: exit=0 (3s)`, `macos: exit=0 (4s)`, and `4/4 hosts passed`.
  * Linux and WSL ran `make verify`: `0 issues.`, total coverage `83.7%`,
    `No vulnerabilities found.`.
  * Windows and this Mac ran the native tests (0017-MADR). Each listed
    `ok` for all five packages, with timings, so none was cached.
* **Step 3:** the disclosure guard over `origin/main..f75153e` exited 0.
* **Step 4** waits for this record's commit. Then the push needs the
  owner's ask.
