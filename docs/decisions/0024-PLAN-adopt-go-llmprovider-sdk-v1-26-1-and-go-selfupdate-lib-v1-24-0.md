---
status: in-progress
date: 2026-11-05
associated-madr: "0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md"
---
# Implement the move to go-llmprovider-sdk v1.26.1 and go-selfupdate-lib v1.24.0

Associated MADR: [0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md](0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md)

*Revised 2026-11-05, before execution, for the owner's answers ("carry
the workspace through, and release v1.10.3"): Phase 2b (MADR D3) and
Phase 5 (MADR D6) are unconditional. Execution waits for the owner's
approval of this PLAN.*

## Goal

* **D1:** `go.mod` requires `go-llmprovider-sdk v1.26.1` and
  `go-selfupdate-lib v1.24.0`.
* **D2:** both reusable-workflow lines in `ci.yml` and their quotes in the
  operations guide pin `74ab5dd376a4c864debb32b2d8abfc7856059ba3 #
  v1.24.0`, with the comment rewritten.
* **D3:** a Claude workspace the wizard returns is saved and used.
* **D5:** the README states what a user can see.
* **D6:** `v1.10.3` is released and checked.

## Scope

### In scope

* `go.mod`, `go.sum`.
* `.github/workflows/ci.yml`: the comment above the build job
  (`ci.yml:110-116`) and the two `uses:` lines (`:117`, `:132`).
* `docs/guides/cicd-operations.md:175`, `:179`.
* D3: `internal/config/config.go` (`ProviderConfig`),
  `internal/ui/setup.go` (`existing`, the saved result, `discoverModels`),
  `main.go` (`providerOptions`), and tests beside them.
* `README.md`: a "Changes with …" section and its table-of-contents line.
* `docs/README.md`: this pair's index rows, and their status as it moves.

### Out of scope

* D4's candidates: `chatcompat`, `LocalServers`, `catalog.Lister`,
  `NewHTTPClient`, reasoning options, images, `Updater.Plan` / `Apply`,
  every `service` addition.
* A `--workspace` flag for `configure --yes`.
* `retryKind`'s labels: the offline label's change is reported in the
  README, not altered.
* The Go pins, and any `golang.org/x/*` move beyond what `go mod tidy`
  takes from D1.
* `selfupdate-release.json`.
* The text of records `0001` to `0023`, and earlier README "Changes with"
  sections. 0023-PLAN's "seven" is amended when that PLAN executes.

## Implementation Steps

All work is on `main`; the agent checks and stages each phase, and the
owner commits with the global hook writing the message, and pushes
(`0021-MADR-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md`
A1). Each phase ends with `make check-records` and `make markdownlint`, and
`make pre-add-check` when Go files or `go.mod`/`go.sum` are staged.

### Phase 0: preconditions (read-only)

1. `git rev-parse --path-format=absolute --git-path hooks` names the
   global hooks directory.
2. `go list -m -versions` still ends at `v1.26.1` (go-llmprovider-sdk)
   and `v1.24.0` (go-selfupdate-lib), and `git ls-remote --tags` peels
   go-selfupdate-lib `v1.24.0` to `74ab5dd3…`. If either is newer, stop
   and amend the MADR.

### Phase 1: the records

1. This pair and `docs/README.md`'s two index rows. The MADR is
   `accepted`, with Q1 and Q2's answers recorded (2026-11-05).
2. On approval: this PLAN's `status` becomes `in-progress`; the index
   follows. Staged for the owner (bootstrap exception: records only).

### Phase 2: modules and workflow pins (D1, D2)

One commit, because `TestCurrentDocumentation` fails when `ci.yml` and the
guide disagree.

1. `go get github.com/maccavelli/go-llmprovider-sdk@v1.26.1
   github.com/maccavelli/go-selfupdate-lib@v1.24.0`, then `go mod tidy`.
   `git diff go.mod go.sum` shows only the lines the MADR's D1 names.
2. In `ci.yml` and `docs/guides/cicd-operations.md`, replace
   `84805563ed252df384f59ced45a3daf529702ab4 # v1.22.0` with
   `74ab5dd376a4c864debb32b2d8abfc7856059ba3 # v1.24.0`, by a Python
   script asserting two occurrences in each before the replace.
3. Rewrite `ci.yml:110-116`: both workflows are pinned to the commit of
   go-selfupdate-lib v1.24.0, the version `go.mod` requires; the publish
   path is unchanged from v1.22.0's, which published live (this
   repository's v1.10.2, run 40150318042), so 0010-MADR D2's rule, as
   amended, holds; cite this MADR's D2 and 0010-MADR D2, A1.
4. Check: in each of the two files, `grep -c 74ab5dd376a4c864debb32b2d8abfc7856059ba3`
   gives 2 and `grep -c 84805563ed252df384f59ced45a3daf529702ab4` gives 0;
   `go mod tidy -diff` is empty; `make verify` exits 0. Read the whole
   diff.

### Phase 2b: the workspace (D3)

1. `internal/config/config.go`: `Workspace string
   \`json:"workspace,omitempty"\`` in `ProviderConfig`, documented as the
   Anthropic workspace a Claude key not scoped to one runs in.
2. `internal/ui/setup.go`: set `existing.Workspace` from the saved config
   (`:236-239`); save `pc.Workspace = res.Workspace` beside
   `pc.Organization` (`:293`); at the `--yes` path's two `discoverModels`
   calls (`:447`, `:464`), add `catalog.WithClaudeWorkspace(pc.Workspace)`
   to the options when it is not empty. The `--yes` path does not clear
   `pc.Workspace`.
3. `main.go`: `providerOptions` appends `claude.WithWorkspace` for
   `ProviderClaude` with a non-empty `pc.Workspace`.
4. Tests: a config with `workspace` round-trips through `Save` and load; a
   wizard result with a workspace is saved by `RunSetupWithOptions` (with
   the existing fake prompter seams); `providerOptions` yields the
   workspace option for Claude and none for others or for empty.
5. Each replacement asserts one occurrence before it is made; the whole
   diff is read afterwards.
6. Check: `make pre-add-check` on the changed files; `make verify` exits
   0. On a scratch copy of the tree with the step-2 save line removed, and
   again with the step-3 option removed, the new tests fail; record the
   output. Never in the tree.

### Phase 3: the README (D5)

1. Add "Changes with go-llmprovider-sdk v1.26.1 and go-selfupdate-lib
   v1.24.0" after the `v1.22.2` / `v1.22.0` section, in its shape, with
   the bullets the MADR's D5 lists, the workspace's included. Its first
   line is "Release `v1.10.3` moves …", and cites the MADR.
2. Add its table-of-contents line after README.md:34.
3. Check: `make markdownlint`; `TestCurrentDocumentation` and
   `TestReadmeCuratedModels` pass.

### Phase 4: before a push

1. `make preflight`.
2. `make preflight-hosts` on the committed `HEAD` (Go and `go.mod`
   changed). A host that cannot be reached is reported, not skipped.
3. The disclosure guard over the outgoing commits, as AGENTS.md writes it.
4. The owner pushes `main`; CI must be green on `main`, its build
   rehearsal included, with its log naming `…@74ab5dd3…`.

### Phase 5: the release `v1.10.3` (D6)

1. CI on `main` is green.
2. The owner tags `v1.10.3` on `main` and pushes it, or asks the agent to
   in that turn.
3. Record the run: its URL, that it published through `74ab5dd3…`, and the
   assets (six binaries, `SHA256SUMS`, `install.sh`, `install.ps1`);
   verify one binary's checksum and build attestation.
4. `prepare-commit-msg update` on an installed `v1.10.2` reaches `v1.10.3`,
   and `prepare-commit-msg version` names it. That update runs `v1.10.2`'s
   code; `update --check` from `v1.10.3` is the first live run of
   `v1.24.0`'s updater.
5. This PLAN becomes `complete`, the index follows, in a records commit
   on `main`.

## Verification

* Phase 2: `make verify` exits 0, as in the MADR's probe; the `grep -c`
  counts hold.
* Phase 2b: the new tests pass, and fail on scratch copies without the
  change; `make verify` exits 0.
* Phase 3: `make markdownlint` and `make check-records` exit 0.
* Phase 4: `make preflight` and `make preflight-hosts` exit 0; the
  disclosure guard reports nothing; CI green on `main`.
* Phase 5: the release run succeeds and `update` reaches it.

## Rollout and Rollback

* **Rollout:** commits on `main`, then the `v1.10.3` tag; users take it
  with `update`.
* **Rollback:** revert Phase 2b's commit alone if only the workspace
  handling misbehaves, or Phases 2b and 2 together; if a release shipped,
  tag a new patch from the revert. A saved `workspace` field is then
  ignored by the reverted binary.
* **If the publish run fails** at `74ab5dd3…`: repin the publish workflow
  to `84805563…`, amend the MADR, and tag a new patch.

## Execution record

Execution approved by the owner, 2026-11-05: "approved to proceed".

### Phase 0, 2026-11-05

* `git rev-parse --path-format=absolute --git-path hooks` names the global
  hooks directory.
* `go list -m -versions` ends at `v1.26.1` (go-llmprovider-sdk) and
  `v1.24.0` (go-selfupdate-lib); `git ls-remote --tags` gives
  `v1.24.0^{}` = `74ab5dd376a4c864debb32b2d8abfc7856059ba3`.

### Phase 1, 2026-11-05

* The pair and its two index rows were staged on `main`, with `HEAD` at
  `6337fa0`; records only. `make check-records`, `make markdownlint` and
  `make disclosure-check` exit 0.
* Committed by the agent as `1a1b8ea`, `git commit --no-edit` with the
  global hook writing the message, on the owner's "Commit to main"
  (Deviation 1).

### Deviation 1, 2026-11-05: the agent commits on the owner's per-turn word

* **Found.** "Implementation Steps" has the owner commit each phase. After
  Phase 1 was staged, the owner wrote "You have my permission to commit to
  master"; the repository has no `master`, so the agent asked, and the
  owner answered "Commit to main".
* **Decision.** The agent commits a phase on `main` when the owner
  authorizes it by branch name in that turn, under
  `sysadmin-git-branch-gates`; otherwise it stages and stops. Each
  authorization covers one commit, not a push, which the owner still
  does. The message is the global hook's, through `git commit --no-edit`.
  No decision of the MADR changes.
* **Files added to scope:** none.

### Phase 2, 2026-11-05

* `go get …go-llmprovider-sdk@v1.26.1 …go-selfupdate-lib@v1.24.0` and
  `go mod tidy` exit 0. `go.mod` changes the two modules' `require` lines
  and `golang.org/x/mod` v0.42.0, `x/sys` v0.49.0, `x/term` v0.47.0
  (indirect); `go.sum` replaces those five modules' ten lines. `go mod
  tidy -diff` is empty. The diff is the MADR's probe's.
* The repin script asserted 2 occurrences of the old ref in `ci.yml` and
  in the operations guide, and 1 of the old comment in `ci.yml`, before
  replacing them. Afterwards, in each file: `74ab5dd3…` 2, `84805563…` 0.
  The comment at `ci.yml:110-116` names v1.24.0, v1.22.0's publish path,
  `v1.10.2` run 40150318042 and this MADR. The whole diff was read.
* `make verify` exits 0: total coverage 83.7% (minimum 80.0%), "No
  vulnerabilities found.", "ci filters: 11 patterns, 11 reads checked, 0
  problems", the tools' 25 tests and the scripts' 34 "OK", and the six
  cross-builds.
* `make pre-add-check` checks the staged snapshot; run before staging it
  verified `HEAD`'s v1.22.2 and v1.22.0, so it was run again on the staged
  phase: exit 0, "all modules verified", go-llmprovider-sdk v1.26.1 and
  go-selfupdate-lib v1.24.0 "resolved from GitHub", "0 issues.", "No
  vulnerabilities found.". `make check-records`, `make markdownlint` and
  `make disclosure-check` exit 0.
