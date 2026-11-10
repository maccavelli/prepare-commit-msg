---
status: complete
date: 2026-11-10
associated-madr: "0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md"
---
# Implement the move to go-llmprovider-sdk v1.29.0 and go-selfupdate-lib v1.26.0

Associated MADR: [0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md](0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md)

## Goal

* **D1:** `go.mod` requires go-llmprovider-sdk `v1.29.0` and
  go-selfupdate-lib `v1.26.0`. `go.mod` and `go.sum` change on those
  modules' lines alone.
* **D2:** both workflow pins name `7c677842…` (`v1.26.0`), with the comment
  above them and the operations guide's two quotes to match.
* **D3:** no Go code or test changes.
* **D4:** the README's "Changes with go-selfupdate-lib v1.25.0" section and
  its contents line become "Changes with go-llmprovider-sdk v1.29.0 and
  go-selfupdate-lib v1.26.0".
* **D5:** under Q1 A, Phase 5 releases `v1.11.1`, which also completes
  `0030-PLAN-adopt-go-selfupdate-lib-v1-25-0.md`.

## Scope

### In scope

* `go.mod` and `go.sum`.
* `.github/workflows/ci.yml`: the two `uses:` lines and the comment above
  the build job.
* `docs/guides/cicd-operations.md`: its two quotes of those lines.
* `README.md`: the section and contents line D4 names.
* `docs/README.md`: this pair's two index rows, and 0030-PLAN's row when it
  completes.
* This MADR and this PLAN, including the execution record; 0030-PLAN's
  execution record and status, from Phase 5.
* Under Q1 A: the tag, and the release its run publishes.

### Out of scope

* Every `.go` file (MADR D3, option C).
* The `golang.org/x` modules and the Go toolchain.
* The README's Security Note and Option B sentences, which stand.
* Running `tools/prune_releases.py`, except on the owner's word, as the
  operations guide's "Controlled release" gives it.
* The text of earlier records, other than 0030-PLAN's execution record and
  status.

### Permissions

Each of these needs the owner's word in the turn it happens, under
`sysadmin-git-branch-gates`:

* a commit, naming `main`;
* a push, naming `main`;
* making a tag, and pushing it.

The global hook writes every message through `git commit --no-edit`.

## Implementation Steps

Every phase ends with `make check-records` and `make markdownlint`. A
phase that stages `go.mod`/`go.sum` also runs `make pre-add-check`. Each
exit status is captured before any filter, with long output in the session
scratchpad. Every text edit is made by a scratchpad Python script that
asserts its anchor occurs exactly once (or the stated count) before
replacing it, and asserts the new text afterwards.

### Phase 0: preconditions (read-only)

1. `git rev-parse --path-format=absolute --git-path hooks` names the global
   hooks directory.
2. `git status --short` lists only the 0030 pair's two uncommitted record
   edits and this pair, or nothing the owner has not committed. `HEAD` and
   `origin/main` are `121879f`, or the owner's record commits on top of it.
3. `go list -m …@latest` and `git ls-remote --tags` end at
   go-llmprovider-sdk `v1.29.0` and go-selfupdate-lib `v1.26.0`;
   `v1.26.0` peels to `7c677842dface4b14b65b2e8b46aa0a7543bed13`. If either
   library has a newer tag, stop and amend the MADR.
4. `git diff --stat v1.25.0 v1.26.0 -- .github scripts Makefile` in a clone
   of go-selfupdate-lib is empty, so D2's publish-path claim holds.
5. `gh api repos/maccavelli/prepare-commit-msg/releases/latest` names
   `v1.11.0`.

### Phase 1: the records

1. Set this PLAN to `status: in-progress`, and the MADR to `accepted` with
   Q1's answer recorded under "Owner questions" and in D5.
2. Add the two index rows to `docs/README.md` after the 0030 rows:
   * MADR: "Adopt provider SDK v1.29.0 and self-update library v1.26.0",
     "accepted <date>";
   * PLAN: "Implement provider SDK v1.29.0 and self-update library v1.26.0",
     "in progress".
3. These, and the 0030 pair's uncommitted record edits, go into Phase 2's
   commit unless the owner commits them first. Either is expected, not a
   deviation.

### Phase 2: the modules, the pins and the README (D1–D4)

1. **The modules.** Run
   `go get github.com/maccavelli/go-llmprovider-sdk@v1.29.0 github.com/maccavelli/go-selfupdate-lib@v1.26.0`,
   then `go mod tidy`. Check:
   * `git diff go.mod` changes two lines: `v1.28.0` to `v1.29.0`, and
     `v1.25.0` to `v1.26.0`;
   * `git diff go.sum` replaces the modules' four lines with the hashes the
     MADR records: go-llmprovider-sdk
     `h1:KVuUfHpx6TtKwjH+XjSSkBU+PRRKg9aWa6p69iaOzy0=` and `/go.mod`
     `h1:dE+W7oDyvVmAxpSOuIJO3qSD3bDTh/olLt/TbByBE1M=`; go-selfupdate-lib
     `h1:giqUiSV90pDZErgh+fEWR10KnM6CdDKOgcjDonVa1xM=` and `/go.mod`
     `h1:DIqb0hzmxkYhVTj6iyaHpntadoHkeOkHWyTsdw+MBHE=`;
   * `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
     verified".

   Any other line changing stops the phase.
2. **The workflow pins (D2).** In `.github/workflows/ci.yml`, replace
   `875dec3c9d475ca47c4b012ca43742de832ef626 # v1.25.0` with
   `7c677842dface4b14b65b2e8b46aa0a7543bed13 # v1.26.0`, asserting 2
   occurrences before and after. Replace the comment block above the build
   job, asserted once, with:

   ```yaml
       # Both workflows are pinned to the commit of go-selfupdate-lib v1.26.0,
       # the version go.mod requires. The publish path (the publish workflow
       # file and the scripts it runs) is unchanged from v1.24.0's, which has
       # published live (this repository's v1.11.0, run 42434050386), so
       # 0010-MADR D2's rule, as amended, holds
       # (docs/decisions/0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md
       # D2; docs/decisions/0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md D2, A1).
   ```

   In `docs/guides/cicd-operations.md`, make the same pin replacement,
   asserting 2 occurrences before and after.
3. **The README (D4).** Replace the contents line
   `- [Changes with go-selfupdate-lib v1.25.0](#changes-with-go-selfupdate-lib-v1250)`,
   asserted once, with
   `- [Changes with go-llmprovider-sdk v1.29.0 and go-selfupdate-lib v1.26.0](#changes-with-go-llmprovider-sdk-v1290-and-go-selfupdate-lib-v1260)`.
   Replace the section from `## Changes with go-selfupdate-lib v1.25.0` to
   its closing line `- **The hook, \`configure\` and the installers** behave
   as in \`v1.11.0\`.`, asserted once as one block, with:

   ```markdown
   ## Changes with go-llmprovider-sdk v1.29.0 and go-selfupdate-lib v1.26.0

   Release `v1.11.1` moves from go-llmprovider-sdk `v1.28.0` to `v1.29.0`,
   and from go-selfupdate-lib `v1.24.0` to `v1.26.0`
   ([0030-MADR](docs/decisions/0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md),
   [0031-MADR](docs/decisions/0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md)).
   What you may notice:

   - **`update` started through a symlink is refused on Linux,** as it
     already was on macOS and Windows. Before, Linux replaced the file the
     link points to. Run the binary itself, such as
     `~/.global-git-hooks/prepare-commit-msg update`.
   - **The refusal names the link and its target,** and says to start the
     target to update it.
   - **`update --json`** reports that refusal's `failed` event with the
     class `symlink-invocation`, a target with a setuid or setgid bit with
     `special-mode-bits`, and a target outside the allowed directories with
     `no-allowed-root`. Each was `error`.
   - **An OpenCode budget or credit limit is not retried,** nor is a Gemini
     daily quota reported partway through an answer. The hook tries the
     next model at once.
   - **A failure partway through an answer** that states how long to wait
     is retried after that wait, up to 30 seconds. A longer one moves to
     the next model at once.
   - **`configure` and the installers** behave as in `v1.11.0`.
   ```

   The `---` rules around the section are unchanged.
4. **Read the whole diff.** `git diff` covers `go.mod`, `go.sum`,
   `.github/workflows/ci.yml`, `docs/guides/cicd-operations.md`,
   `README.md`, `docs/README.md`, this pair, and the 0030 pair's record
   edits if still uncommitted. No `.go` file changes.
5. **The gates:**
   * `make verify` exits 0, as in the MADR's probe. `ci.yml` and `README.md`
     are read by `make workflow-lint` and `readme_test.go`, so both run here;
   * `7c677842dface4b14b65b2e8b46aa0a7543bed13` occurs 2 times in `ci.yml`
     and 2 in the operations guide; `875dec3c9d475ca47c4b012ca43742de832ef626`
     0 in both;
   * `make check-records`, `make markdownlint` and `make disclosure-check`.
6. **Stage** those files, then run `make pre-add-check` on the staged
   snapshot: exit 0, "all modules verified", both modules resolved at the
   new versions.
7. **Commit** on the owner's word naming `main`.

### Phase 3: before the push

1. `make preflight` on the committed `HEAD`: "preflight passed".
2. `make preflight-hosts` on the committed `HEAD`, because `go.mod` and a
   workflow changed. It needs a clean tree. Each host prints `exit=0`. A
   host that cannot be reached is reported, not skipped.
3. The disclosure guard over `origin/main..HEAD`, as `AGENTS.md` gives it.
   It exits 0 with no output.

### Phase 4: push and CI

1. Push `main` on the owner's word naming `main`.
2. Wait for the runs on the pushed `HEAD` with `gh run list --commit <sha>`.
   `ci.yml` runs in full, and `docs.yml` for the records. Check:
   * "Go (quality contract)", the three native test jobs, "Build release
     / build" (`rehearsal=true`) and the identity jobs succeed;
   * the build job's log names
     `build-selfupdate-release.yml@7c677842dface4b14b65b2e8b46aa0a7543bed13`;
   * "Publish GitHub Release" is skipped;
   * the Docs run succeeds.

   A failure stops the phase, and is recorded as a deviation.

### Phase 5: the release (only if Q1 is A)

0030-PLAN Phase 5, run once for both records, with these changes: the
publish pin is `7c677842…`; and step 7's refused run prints, on stderr, the
line `update failed: selfupdate: refusing to update through a symlink:
<link> links to <target>; start <target> itself to update it`, with the
scratch link and target as its paths. *Amended by Deviation 1: the line
is the MADR's A1, with `selfupdate: prepare-commit-msg: ` before the
refusal.* In full:

1. The owner's word makes and pushes an annotated tag `v1.11.1` on `main`'s
   `HEAD`.
2. The rule-suites API, or the push's output, shows no bypass of a ruleset.
3. Every job of the tag's run succeeds. "Publish GitHub Release / publish"
   names `publish-selfupdate-release.yml@7c677842dface4b14b65b2e8b46aa0a7543bed13`.
4. The release is not a draft or prerelease, `immutable: true`, has the nine
   assets, and `releases/latest` names it.
5. `prepare-commit-msg-darwin-arm64` matches `SHA256SUMS`, and
   `gh attestation verify` with the publish workflow as signer exits 0, its
   source the tag at the tagged commit.
6. On this Mac's installed `v1.11.0`: `update --check` exits 10 naming
   `v1.11.1`, `update --yes` reports "release asset integrity verified",
   `version` names `v1.11.1`, and `update --check` then exits 0.
7. **The symlink refusal, on the Linux host** named by
   `available-dev-hosts`, in a scratch directory under its home that is
   removed afterwards. Download `prepare-commit-msg-linux-<arch>` of
   `v1.11.1`, check it against that release's `SHA256SUMS`, make it
   executable, and link to it from a second scratch directory. Run
   `update --force --yes` through the link by its path: it exits 1, stderr
   holds the line above, and the target's checksum is unchanged. Then run
   `update --force --yes` on the target itself: it exits 0, and `version`
   names `v1.11.1`. Remove the scratch directories, and confirm they are
   gone.
8. A failed run follows the operations guide's "Failed-release recovery",
   recorded as a deviation here and in 0030-PLAN.
9. The pruning, `tools/prune_releases.py`, runs only on the owner's word.

### Phase 6: the records

1. Write this PLAN's execution record: each phase's commands, exit statuses
   and the output the checks read; the commit hashes and run IDs; a dated
   entry for each deviation.
2. In 0030-PLAN, add a Phase 5 execution record pointing to this PLAN's
   Phase 5 and the tag, and a dated entry that its step 7's stderr was read
   as this PLAN gives it. Set 0030-PLAN, and its index row, to complete.
3. Set this PLAN to `status: complete`, and its index row to complete.
4. Run `make check-records`, `make markdownlint` and
   `make disclosure-check`. Commit and push on the owner's word, naming
   `main`. These are records only, so `docs.yml` runs.

## Verification

* Phase 0: the preconditions hold, and both libraries are at their newest
  tags.
* Phase 2: `go.mod`/`go.sum` change on D1's lines alone; both pins and both
  quotes name `7c677842…`; no `.go` file changes; `make verify`,
  `make pre-add-check`, `make check-records`, `make markdownlint` and
  `make disclosure-check` exit 0; the whole diff was read.
* Phase 3: `make preflight` and `make preflight-hosts` pass on the
  committed `HEAD`, and the disclosure guard reports nothing.
* Phase 4: CI and Docs are green on `main`, and the build rehearsal ran the
  new pin.
* Phase 5, if it runs: the release is immutable, has nine assets, and is the
  latest. Its checksum and attestation check out. `update` moves `v1.11.0`
  to it on this Mac. On Linux, `update` through a link is refused with the
  new text and leaves the target unchanged, and `update` on the target
  succeeds.
* Phase 6: both PLANs are complete, with their index rows.

## Rollout and Rollback

* **Rollout:** the commit on `main`. Under Q1 A, users receive it, with
  0030's move, in `v1.11.1`, through `update` or the installers.
* **Rollback before a tag:** revert Phase 2's commit, and push on the
  owner's word. `main` returns to 0030's state.
* **Rollback after a release:** the release stays, immutable. Revert the
  offending commit and release the revert as the next patch. A user can
  return to `v1.11.0` with `update --version v1.11.0 --yes` while 0026's
  window keeps it.

## Execution record

Execution approved by the owner, 2026-11-10: "we will release v1.11.1
after ci passes. accept the madr and plan approved to proceed." Q1 is A.

### Phase 0, 2026-11-10

A scratchpad script, read-only, every check passing:

* `git rev-parse --path-format=absolute --git-path hooks` names the global
  hooks directory.
* `git status --short` lists the 0030 pair's two record edits and this
  pair, untracked.
* `HEAD` and `origin/main` are both
  `121879f54b1de17367cef6073769b0329199cc2a`.
* `git ls-remote --tags` and `go list -m …@latest` end at go-llmprovider-sdk
  `v1.29.0` and go-selfupdate-lib `v1.26.0`. `v1.26.0` peels to
  `7c677842dface4b14b65b2e8b46aa0a7543bed13`.
* `git diff --stat v1.25.0 v1.26.0 -- .github scripts Makefile` in the
  library's clone is empty.
* The latest release is `v1.11.0`.

### Phase 1, 2026-11-10

* The MADR was set to `accepted`, with the owner's answer quoted under
  "Owner questions" and D5 marked Q1 A; this PLAN to `in-progress`.
* The two index rows went into `docs/README.md` after the 0030 rows: MADR
  "Adopt provider SDK v1.29.0 and self-update library v1.26.0", accepted
  2026-11-10; PLAN "Implement provider SDK v1.29.0 and self-update library
  v1.26.0", in progress.

### Phase 2, 2026-11-10

* `go get github.com/maccavelli/go-llmprovider-sdk@v1.29.0 github.com/maccavelli/go-selfupdate-lib@v1.26.0`
  and `go mod tidy` exit 0. The script asserted that `git diff go.mod
  go.sum` holds exactly the two `require` lines and the four `go.sum` lines,
  with the hashes the MADR records. `go mod tidy -diff` is empty, and
  `go mod verify` prints "all modules verified".
* The same script made steps 2 and 3's edits, asserting each anchor's count
  before and the new text's after: both `uses:` pins and the comment in
  `ci.yml`; both quotes in the operations guide; the README's contents line
  and section, with the `---` rules on either side of it unchanged, and no
  `v1.25.0` heading or anchor left.
* The whole diff was read. The changes are `go.mod`, `go.sum`,
  `.github/workflows/ci.yml`, `docs/guides/cicd-operations.md`,
  `README.md`, `docs/README.md`, this pair, and the 0030 pair's record
  edits (its MADR A1 and PLAN Phase 2–4 records and Deviation 2, written
  2026-11-09). No `.go` file changed.
* `make verify` exits 0: "0 issues."; total coverage 84.9% (minimum
  80.0%); "No vulnerabilities found."; "ci filters: 11 patterns, 11 reads
  checked, 0 problems"; the tools' 42 tests and the scripts' 34 "OK"; the
  six cross-builds. `make check-records`, `make markdownlint` and
  `make disclosure-check` ("0 finding(s)") exit 0.
* `7c677842dface4b14b65b2e8b46aa0a7543bed13` occurs 2 times in `ci.yml` and
  2 times in the operations guide; `875dec3c9d475ca47c4b012ca43742de832ef626`
  0 times in each.
* Staged: the ten files above. `make pre-add-check` on the staged snapshot
  exits 0: "all modules verified"; go-llmprovider-sdk `v1.29.0` and
  go-selfupdate-lib `v1.26.0` "resolved from GitHub" with the `h1:` hashes
  above; "0 issues."; "No vulnerabilities found.".
* The owner committed `423321d`, "chore(deps): bump llmprovider-sdk v1.29.0
  and selfupdate-lib v1.26.0". It holds exactly the ten staged files.

### Phase 3, 2026-11-10

* `make preflight` on `423321d`: "preflight passed". Total coverage 84.9%,
  "No vulnerabilities found.", the ci-filters line, and the tools' 42 and
  the scripts' 34 tests.
* `make preflight-hosts` on `423321d`, with a clean tree, exits 0:

  ```text
  linux: exit=0 (118s)
  wsl-ubuntu-24.04: exit=0 (48s)
  windows: exit=0 (32s)
  macos: exit=0 (5s)
  preflight-hosts: 4/4 hosts passed
  ```

  No host was unreachable. The Windows and macOS logs each show `ok`, not
  cached, for all five packages.
* The disclosure guard, as `AGENTS.md` gives it, exits 0 with no output
  over `121879f..423321d`.

### Phase 4, 2026-11-10

* The owner pushed `main`, `121879f..423321d`.
* CI run `42866344999` on `423321d`: success. "Go (quality contract)", the
  three native test jobs, "Build release / build" and the five identity
  jobs succeed; "Publish GitHub Release" is skipped. The build job's log
  names `build-selfupdate-release.yml@7c677842dface4b14b65b2e8b46aa0a7543bed13`,
  with `rehearsal=true`.
* Docs run `42866343919`: success.

### Phase 5, steps 1–6, 2026-11-10

* **Step 1.** The owner made and pushed the annotated tag `v1.11.1`. It
  peels to `423321d6a575ee318cddf5335fc356ec46b03f54`, `origin/main`'s
  `HEAD`.
* **Step 2.** The repository's one ruleset, `prepare-commit-msg-main`,
  targets branches. The rule-suites API lists no suite for the tag's ref,
  only the `main` push of 17:34 UTC, which passed: no ruleset was evaluated
  for the tag, so none was bypassed.
* **Step 3.** CI run `42869640645` on the tag: success. Every job succeeds
  but "Publish GitHub Release / sign", skipped as `sign` is unset, as in
  `v1.11.0`'s run `42434050386`. The log names
  `build-selfupdate-release.yml@7c677842dface4b14b65b2e8b46aa0a7543bed13`
  and `publish-selfupdate-release.yml@7c677842dface4b14b65b2e8b46aa0a7543bed13`,
  with `rehearsal=false`. The check script first counted the skipped `sign`
  job as a failure; that was a fault of the check, not of the run.
* **Step 4.** The release is not a draft or prerelease, `immutable: true`,
  and has the nine assets: `SHA256SUMS`, `install.ps1`, `install.sh` and the
  six binaries. `releases/latest` names `v1.11.1`.
* **Step 5.** `prepare-commit-msg-darwin-arm64` matches `SHA256SUMS`
  (`66b9debdf26a23d28182964ba10f249402af5c8ca0b4ef58e55a738daacfa576`), and
  `gh attestation verify` with the publish workflow as signer, the tag as
  source ref and `423321d` as source digest, exits 0.
* **Step 6.** On this Mac's installed `v1.11.0`: `update --check` exits 10,
  "update available: v1.11.0 -> v1.11.1"; `update --yes` exits 0, "release
  asset integrity verified", "updated v1.11.0 -> v1.11.1"; `version` prints
  `v1.11.1 (release) 423321d6a575`; `update --check` then exits 0, "up to
  date (v1.11.1)".

### Deviation 1, 2026-11-10: the refused run's stderr line

* **Found.** Phase 5 step 7, on the Linux host. `update --force --yes`
  through the link exits 1, the `failed` event's class is
  `symlink-invocation`, the target's checksum is unchanged, and on the
  target itself `update --force --yes` exits 0 and `version` names
  `v1.11.1`. But the stderr line is

  ```text
  update failed: selfupdate: prepare-commit-msg: selfupdate: refusing to update through a symlink: <link> links to <target>; start <target> itself to update it
  ```

  where the step expected it without `selfupdate: prepare-commit-msg: `.
  The Updater adds that prefix to every error of a run (go-selfupdate-lib
  `selfupdate/updater.go:772`), as it did at `v1.24.0` and `v1.25.0`: not
  new, and not a fault of the release. The expected line was derived in the
  MADR, never run.
* **Decision.** The owner chose option 2 (2026-11-10): correct the expected
  line to the one printed, rerun step 7 matching it exactly, and raise the
  doubled `selfupdate:` with go-selfupdate-lib as an issue. The other option
  offered was the correction alone. Loosening the match was not offered.
* **Effect.** Phase 5's introduction is annotated above. The MADR records
  the corrected fact as A1. Step 7 is rerun. 0030-PLAN's step 7, which this
  Phase runs, takes the same line.
* **Files added to scope:** none. The issue is outside this repository.

### Phase 5, steps 7–9, 2026-11-10

* **Step 7**, on the Linux host, by a scratchpad script over SSH. The host
  reports `Linux x86_64`. `prepare-commit-msg-linux-amd64` of `v1.11.1` was
  downloaded here, matched `SHA256SUMS`
  (`2779d7b489001052a7f839da7c6e9462d1acb16328906b3f4a292da33ea98c1f`),
  copied to a `mktemp -d` directory under the host's home, and matched it
  again there. It was made executable in `bin/`, with a symlink to it in
  `link/`.
  * Through the link, `update --force --yes` exits 1; the `failed` event's
    class is `symlink-invocation`; stderr holds, exactly, the line
    Deviation 1 gives, with the scratch link and target as its paths; and
    the target's checksum is unchanged.
  * On the target, `update --force --yes` exits 0, and `version` prints
    `v1.11.1 (release) 423321d6a575`.
  * The scratch directory was removed, and `test ! -e` confirms it.

  The first run matched the line as first written and failed on it, the
  failure that Deviation 1 records; the rerun, matching the corrected line,
  passes every check.
* **The issue.** go-selfupdate-lib issue #11, "Run errors repeat the
  package prefix" (<https://github.com/maccavelli/go-selfupdate-lib/issues/11>), proposes that `wrapRun` drop one leading
  `selfupdate: ` from the error it wraps, keeping `errors.Is` and
  `errors.As`. It names no host or path.
* **Step 8.** Not needed: the run succeeded.
* **Step 9.** Not run. The pruning waits for the owner's word.

### Phase 6, 2026-11-10

* 0030-PLAN gained its Phase 5 record, pointing here, and Deviation 3 for
  step 7's line; it and its index row are complete.
* This PLAN and its index row are complete. Every acceptance criterion
  under "Verification" is met; the pruning of step 9 is the owner's to ask
  for, and is not one of them.
* `make check-records`, `make markdownlint` and `make disclosure-check` are
  run before the commit, which the owner makes and pushes.
