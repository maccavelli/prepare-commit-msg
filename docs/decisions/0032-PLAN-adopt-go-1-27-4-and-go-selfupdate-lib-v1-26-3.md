---
status: complete
date: 2026-11-19
associated-madr: "0032-MADR-adopt-go-1-27-4-and-go-selfupdate-lib-v1-26-3.md"
---
# Implement the move to Go 1.27.4 and go-selfupdate-lib v1.26.3

Associated MADR: [0032-MADR-adopt-go-1-27-4-and-go-selfupdate-lib-v1-26-3.md](0032-MADR-adopt-go-1-27-4-and-go-selfupdate-lib-v1-26-3.md)

## Goal

* **D1:** `go.mod` reads `go 1.27.4`; `scripts/bootstrap-tools.sh`,
  `Makefile` and `README.md` state Go 1.27.4.
* **D2:** `go.mod` requires go-selfupdate-lib `v1.26.3`; `go.sum` changes on
  that module's two lines alone.
* **D3:** both workflow pins name `1dffdc9d…` (`v1.26.3`), with the comment
  above them and the operations guide's two quotes to match.
* **D4:** no `.go` file changes, no README release section, no tag, no
  installed binary.

## Scope

### In scope

* `go.mod` and `go.sum`.
* `scripts/bootstrap-tools.sh` (`GO_VERSION`) and `Makefile`
  (`MOD_VERSION`).
* `.github/workflows/ci.yml`: the two `uses:` lines and the comment above the
  build job.
* `docs/guides/cicd-operations.md`: its two quotes of those lines.
* `README.md`: the "Building from source needs Go" sentence.
* `docs/README.md`: this pair's two index rows.
* This MADR and this PLAN, including the execution record.
* One dated pointer line each at the end of the "More Information" of
  `0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md`
  (the Go pin) and
  `0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md`
  (the library and its workflow pins), naming this MADR.

### Out of scope

* Every `.go` file (MADR D4).
* `go-llmprovider-sdk`'s requirement, `golangci-lint`'s pin and the CI
  trigger lines, reported and not changed.
* The text of earlier records beyond those two pointer lines, and of the
  README's earlier release sections.
* Any tag or release, and any install of the hook binary: `make install`, an
  installer script and `go install` of this module.
* Any write to a hooks directory or to Git configuration outside a
  per-test temporary repository.

### Permissions

The coordinator commits and pushes `main`; this PLAN's executor stages the
records and the docs index by path and does nothing else to the index. No
tag.

## Implementation Steps

### Phase 0: preconditions (read-only)

1. Confirm `main` at `2715f2d` with a clean tree.
2. Read `AGENTS.md`; survey every test for what it runs, reads, writes,
   dials or binds. Stop if any test could write outside a temporary
   directory or reach a real provider or credential with `HOME`, the four
   `XDG_*` variables, `GIT_CONFIG_GLOBAL` and `GIT_CONFIG_SYSTEM` thrown away.
3. On a `git archive` copy of `HEAD`: `go get` and `go mod tidy` with
   `GOPROXY=off`; note everything that moves.
4. Run `govulncheck ./...` under Go 1.27.3 on a copy of `HEAD`.
5. Read go-selfupdate-lib's changelog, records and compare for
   `v1.26.0...v1.26.3` against this repository's imports.

### Phase 1: the records

1. This MADR and this PLAN, `docs/README.md`'s two rows.
2. `make check-records` and `make markdownlint` pass.

### Phase 2: the move (D1-D3)

1. `go.mod:3` to `go 1.27.4`.
2. `go get github.com/maccavelli/go-selfupdate-lib@v1.26.3` and
   `go mod tidy`, `GOPROXY=off`; the diff is `go.mod:3`, `go.mod:7` and
   `go.sum`'s two lines.
3. `scripts/bootstrap-tools.sh:10` to `go1.27.4`; `Makefile:14` to `1.27.4`;
   `README.md` "Building from source needs Go 1.27.4."
4. `.github/workflows/ci.yml`: both `uses:` lines pin
   `1dffdc9dee24a7534513dc51c9039408be37c09b # v1.26.3`; the comment names
   `v1.26.3` and cites this record and 0031-MADR D2.
5. `docs/guides/cicd-operations.md`: the two quotes.
6. Search every other live statement of the Go version and of the
   library's version; history stays.

### Phase 3: the gates

Every run uses a throwaway environment: `HOME`, the four `XDG_*` variables,
`GIT_CONFIG_GLOBAL` and `GIT_CONFIG_SYSTEM` point into a scratch directory,
the environment is built from nothing, `GOMODCACHE`, `GOCACHE` and `GOPATH`
are the values `go env` reported first, and no live-test switch is set.
Heavy runs go through `run-heavy` on this macOS host, one at a time.

1. `make tools` (rebuilds the three tools with Go 1.27.4),
   `make mod-check`, `make fmt-check`, `make vet`, `make lint`.
2. `make test` (`go test -race -v ./...`), `make coverage`,
   `make vuln`, `make workflow-lint`, `make tools-test`, `make build-all`.
3. `make check-records`, `make markdownlint`, `make disclosure-check`, after
   the last record edit.
4. `govulncheck ./...` on the changed tree under Go 1.27.4.

### Phase 4: staging

Stage this MADR, this PLAN and `docs/README.md` by path. Everything else
stays unstaged and is listed. The coordinator commits and pushes.

## Verification

* `go.mod` and `go.sum` change only on D1 and D2's lines;
  `go mod tidy -diff` is empty; `go mod verify` passes.
* `1dffdc9dee24a7534513dc51c9039408be37c09b` occurs twice in `ci.yml` and
  twice in the operations guide; `7c677842dface4b14b65b2e8b46aa0a7543bed13`
  in neither.
* Every gate in Phase 3 exits 0.
* `govulncheck ./...` under Go 1.27.4 exits 0.

## Rollout and Rollback

Rollout is the coordinator's commit and push to `main`, on a green gate; CI
runs `ci.yml` in full, whose `build` job is a rehearsal on a non-tag ref and
whose `release` job is skipped. No tag, and the installed hook is unchanged.
Rollback is a revert commit of the move; nothing outside this repository
changes.

## Execution record

### Phase 0, 2026-11-18

* `main` at `2715f2d`, clean tree. `AGENTS.md` names no test-running
  method of its own beyond `make` targets; its standing rule is Python for
  multi-part work.
* **Safety survey, read-only.** No test starts, signals or lists a process,
  binds a fixed port (`net.Listen` is on port 0, `httptest` servers on
  loopback), reads a keychain, or calls a provider: `generateWithRetry`,
  `listingClient` and the update seams are replaced by fixtures; the only
  `git` runs are `internal/git/git_test.go`'s, in a `t.TempDir()` with
  `GIT_CEILING_DIRECTORIES` set, and `migration_test.go`'s
  `make -n build-all`. Tests that read the environment (`osGetenv =
  os.Getenv` in two vendor-login tests) read it after `HOME` and the
  per-user directories are redirected to a temporary directory, which the
  throwaway environment makes empty of any credential. The SDK's model
  metadata fetch is switched off with `LLMPROVIDER_DISABLE_MODELS_METADATA=1`,
  a disable switch and not a live-test switch. Result: safe to run.
* **The throwaway environment** is built from nothing: `HOME`, the four
  `XDG_*` variables, `GIT_CONFIG_GLOBAL` (an empty file) and
  `GIT_CONFIG_SYSTEM` (`/dev/null`) in a fresh scratch directory;
  `GOMODCACHE`, `GOCACHE` and `GOPATH` from `go env`, read first; `GOENV`
  off; `GOPROXY=off` unless a step below says otherwise.
* **Scratch move** on a `git archive` copy of `HEAD`, `GOPROXY=off`:
  `go get github.com/maccavelli/go-selfupdate-lib@v1.26.3` and `go mod tidy`
  change `go.mod:7` and `go.sum`'s two lines alone; nothing is downloaded and
  no other module moves. With `go 1.26.11` the same copy fails with
  "updates to go.mod needed"; `go-llmprovider-sdk v1.29.0` declares
  `go 1.27.3`.
* **`govulncheck ./...` under Go 1.27.3** (the toolchain extracted from the
  module cache, `GOPROXY=off`, the vulnerability database read from its
  public endpoint) on a copy of `HEAD`: exit 3, GO-2026-6859 in
  `net/http`, found in `go1.27.3`, fixed in `go1.27.4`.

### Phase 1, 2026-11-18

This MADR and this PLAN, `docs/README.md`'s two rows, and the two pointer
lines. `check-records`, `markdownlint` and `disclosure-check` are run after the
last record edit; see Phase 3.

### Phase 2, 2026-11-18

* `go.mod:3` `go 1.27.3` to `go 1.27.4`; `go.mod:7` `go-selfupdate-lib
  v1.26.0` to `v1.26.3`; `go.sum:3-4` the module's two lines. `go mod tidy`
  made no other change, and `go mod verify` printed "all modules verified".
* `scripts/bootstrap-tools.sh:10` `go1.27.3` to `go1.27.4`; `Makefile:14`
  `1.27.3` to `1.27.4`; `README.md:910` "Go 1.27.3" to "Go 1.27.4".
* `.github/workflows/ci.yml:110-117` the comment, and `:117` and `:132`
  (`uses:` lines) from `7c677842…` `# v1.26.0` to
  `1dffdc9dee24a7534513dc51c9039408be37c09b # v1.26.3`.
* `docs/guides/cicd-operations.md:181` and `:185`, the same two quotes.
* A search for `1.27.3`, `go1.27.3` and `7c677842` leaves only history: the
  README's `v1.10.1` section and the records.

### Phase 3, 2026-11-18

All in the throwaway environment, `GOPROXY=off` unless noted, exit status
before any filter.

| Gate | Exit | Start | End | Result |
| :--- | :--- | :--- | :--- | :--- |
| `make tools` | 2, then 0 | 18:44:46, 18:44:55 | 18:44:48, 18:45:10 | see Deviation 1 |
| `make mod-check` | 0 | 18:45:22 | 18:45:24 | tidy diff empty; all modules verified |
| `make fmt-check` | 0 | 18:45:24 | 18:45:26 | no output |
| `make vet` | 0 | 18:45:26 | 18:45:28 | no output |
| `make lint` | 0 | 18:45:28 | 18:46:04 | "0 issues." (golangci-lint v2.14.0 built with go1.27.4) |
| `make test` | 0 | 18:46:07 | 18:47:39 | `go test -race -v ./...`; every package `ok`; 0 `--- FAIL`; one skip (Windows-only) |
| `make coverage` | 0 | 18:47:44 | 18:47:50 | total 84.9% (minimum 80.0%) |
| `make vuln` | 0 | 18:47:50 | 18:47:52 | "No vulnerabilities found." under go1.27.4 |
| `make workflow-lint` | 0 | 18:49:30 | 18:49:32 | "ci filters: 11 patterns, 11 reads checked, 0 problems" (first run 18:47:52, exit 0) |
| `make tools-test` | 2, then 0 | 18:47:54, 18:48:37 | 18:48:26, 18:49:30 | see Deviation 2; second run 42 tools tests and 34 scripts tests OK |
| `make build-all` | 0 | 18:49:36 | 18:56:27 | six cross-builds (the interval includes 3 minutes in the heavy-run queue) |
| `make pre-add-check` on a scratch copy with the change staged | 0 | 19:00:17 | 19:00:25 | see Deviation 3; "0 issues.", "No vulnerabilities found.", `go-selfupdate-lib v1.26.3 resolved from GitHub` (a first run, 18:57, gave the same) |
| `make preflight` | 0 | 18:58:26 | 19:00:00 | `check-records`, `markdownlint` ("0 issues"), then `verify` (the targets above, `build-all` included); "preflight passed" |
| `make disclosure-check` | 0 | 19:00:17 | 19:00:17 | 143 files scanned, 0 findings |

`make check-records`, `make markdownlint` and `make disclosure-check` are run
once more after this record's last edit; the handoff gives that result.

### Deviation 1, 2026-11-18: `make tools` and the module proxy

* **Found.** With `GOPROXY=off`, `make tools` stops at
  `golangci-lint@v2.14.0`: "loading deprecation … module lookup disabled by
  GOPROXY=off". The three tool binaries were built with go1.27.3, and the
  bootstrap script rebuilds a tool built by another Go.
* **Resolution.** The same target, with the public module proxy, exit 0. No
  file was added to the module cache (a file listing newer than the run's
  start is empty): the proxy answered the deprecation lookup, and the
  sources were cached. The three tools are now built with go1.27.4;
  `.tools/` is ignored.

### Deviation 2, 2026-11-18: the first `make tools-test`

* **Found.** Five failures in `tools/test_tools.py` (`HostsTest`), each
  `TypeError: rmtree() got an unexpected keyword argument 'onexc'`, from
  `python3` resolving to the system's 3.9.6 on the throwaway `PATH`, which
  held no newer Python. Not caused by the change.
* **Resolution.** The throwaway `PATH` gained the Python the owner's shell
  uses (3.14); `make tools-test` then passed (42 and 34 tests), and
  `make workflow-lint` was run again with it (exit 0). The first run's log is
  kept in the scratch directory.

### Deviation 3, 2026-11-18: `make pre-add-check` reads the index

* **Found.** `scripts/go-precheck.py` checks the staged snapshot. With
  nothing of the change staged (only records and the docs index may be) it
  checked `HEAD`'s `go.mod` (`go-selfupdate-lib v1.26.0`), and it refuses
  `GOSUMDB=off`.
* **Resolution.** The working tree was copied to a scratch repository, all of
  it staged there, and `make pre-add-check` run in it with the public module
  proxy and checksum database: exit 0. The real index was not touched. The
  coordinator runs the target again when `go.mod` is staged for the commit.

### Not run

`make preflight-hosts` tests the committed `HEAD`, and nothing is committed.
No push, tag, release or install was made, and no hooks directory or Git
configuration outside the temporary test repositories was written.

### Close-out, 2026-11-19

Added; no earlier line of this record is rewritten. The status line and the
`date` line above change from `in-progress` and 2026-11-18 to `complete` and
2026-11-19, and nothing else is removed.

* **Release from hold.** The owner's instruction was to commit the bump. The
  work was committed and pushed to `main` by the coordinator: the records in
  `d9a5a0e`, the change in `d377e88`.
* **Gate on the tree.** `make preflight` through `run-heavy` in the
  throwaway environment: exit 0, "preflight passed", ten packages ok under
  the race detector, no failure, the tree unchanged by the gate. After the
  record commit, with the seven change files staged, `make pre-add-check` on
  that staged snapshot through `run-heavy` with the proxy and checksum
  database on: exit 0, "0 issues.", "No vulnerabilities found." This is the
  run Deviation 3 asked for. Nothing was installed and no hooks directory was
  written.
* **CI on `d377e88`, run 44604053367.** Attempt 1: Go (quality contract),
  Native Tests (Linux) and Native Tests (Windows) succeeded; Native Tests
  (macOS) was cancelled with zero steps, GitHub's annotation reading "The job
  was not acquired by Runner of type hosted even after multiple attempts", a
  hosted-runner capacity fault and not a test result; Build release was
  skipped because of it. The failed job was re-run once. Attempt 2 succeeded:
  Native Tests on macOS, Linux and Windows (7 steps each), Go (quality
  contract, 10 steps), Build release / build (18 steps), Build release /
  identity for darwin/arm64, linux/arm64, windows/arm64, linux/amd64 and
  windows/amd64 (5 steps each); Publish GitHub Release skipped (no tag).
  Docs succeeded. This closes the "Not verified" items of the first handoff:
  the native Linux and Windows tests and the `build` rehearsal at the pinned
  `v1.26.3` workflow commit.
* **Acceptance criteria**, against `d377e88` and the evidence above:

| Criterion | Result | Evidence |
| :--- | :--- | :--- |
| `go.mod` and `go.sum` change only on D1 and D2's lines | met | the commit's diff of both files is `go.mod:3`, `go.mod:7` and `go.sum:3-4` alone |
| `go mod tidy -diff` is empty; `go mod verify` passes | met | `make mod-check` exit 0 (Phase 3); `go mod tidy -diff` exit 0 on `d377e88` |
| `1dffdc9d…` twice in `ci.yml` and in the operations guide; `7c677842…` in neither | met | counts 2 and 0 in each file at `d377e88` |
| Every gate in Phase 3 exits 0 | met | the Phase 3 table, with Deviations 1 to 3 resolved as recorded, and the coordinator's `make preflight` exit 0 |
| `govulncheck ./...` under Go 1.27.4 exits 0 | met | `make vuln` exit 0, "No vulnerabilities found." |
| No `.go` file changes (MADR D4) | met | the seven files of `d377e88` include none |
| No tag, release or installed binary | met | CI's Publish GitHub Release skipped; nothing installed |
| CI green on `main` | met | run 44604053367 attempt 2, with the macOS capacity fault re-run once |
