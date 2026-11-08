---
status: complete
date: 2026-11-07
associated-madr: "0028-MADR-adopt-go-llmprovider-sdk-v1-28-0.md"
---
# Implement the move to go-llmprovider-sdk v1.28.0

Associated MADR: [0028-MADR-adopt-go-llmprovider-sdk-v1-28-0.md](0028-MADR-adopt-go-llmprovider-sdk-v1-28-0.md)

## Goal

* **D1:** `go.mod` requires `go-llmprovider-sdk v1.28.0`. `go.mod` and
  `go.sum` change on that module's lines alone.
* **D2:** no other code, test, workflow or toolchain change.
* **D3:** as Q1 is answered. With the recommended answer, no release and no
  README section. With B or C, Phase 5 runs and the README gains a section.

## Scope

### In scope

* `go.mod` and `go.sum`.
* `docs/README.md`: this pair's two index rows, and their status as it
  moves.
* This MADR and this PLAN, including the execution record.
* Only if Q1 is B or C: a `README.md` section and its contents line, the tag,
  and the release its run publishes.

### Out of scope

* `catalog.Cache.Close` and `auth.OAuthSession.Wait` (MADR D2, option B).
* go-selfupdate-lib, `.github/workflows/ci.yml`, the operations guide and
  the Go pins.
* Running `tools/prune_releases.py`. If Q1 releases, the pruning after it
  follows the operations guide's "Controlled release", on the owner's word.
* The text of earlier records and earlier README "Changes with" sections.

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
scratchpad.

### Phase 0: preconditions (read-only)

1. `git rev-parse --path-format=absolute --git-path hooks` names the global
   hooks directory.
2. `git status --short` lists only this pair, untracked, or nothing if the
   owner committed it. `git rev-parse HEAD origin/main` gives `870639c`, or
   the owner's commit of this pair on top of it.
3. `go list -m -versions github.com/maccavelli/go-llmprovider-sdk` ends at
   `v1.28.0`. `git ls-remote --tags` on go-selfupdate-lib ends at
   `v1.24.0`. If either is newer, stop and amend the MADR.
4. `gh api repos/maccavelli/prepare-commit-msg/releases/latest` names
   `v1.10.4`.

### Phase 1: the records

1. Set this PLAN to `status: in-progress`, and the MADR to `accepted` with
   Q1's answer recorded in D3.
2. Add the two index rows to `docs/README.md` after the 0027 rows:
   * MADR: "Adopt provider SDK v1.28.0", "accepted 2026-11-07";
   * PLAN: "Implement provider SDK v1.28.0", "in progress".
3. These go into Phase 2's commit, as `0027-PLAN-adopt-go-llmprovider-sdk-v1-27-0.md`
   Deviation 1 settled. If the owner has committed the pair alone, that is
   expected, not a deviation.

### Phase 2: the module (D1, D2)

1. **The module.** Run
   `go get github.com/maccavelli/go-llmprovider-sdk@v1.28.0`, then
   `go mod tidy`. Check:
   * `git diff go.mod` changes one line, `v1.27.0` to `v1.28.0`;
   * `git diff go.sum` replaces the module's two lines with
     `h1:Q1XMpmJzR8apF06whzRscIG2h3dRgLtvtJ65wWoaD34=` and the `/go.mod`
     hash `h1:dE+W7oDyvVmAxpSOuIJO3qSD3bDTh/olLt/TbByBE1M=`;
   * `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
     verified".

   Any other line changing stops the phase.
2. **Only if Q1 is B or C, the README.** A scratchpad Python script asserts
   the `v1.27.0` section's closing anchor once,
   `- **The hook, \`update\` and the installers** behave as in \`v1.10.3\`.`,
   followed by `---` and `## CLI Reference`. It inserts this section after
   that `---`, followed by its own `---`, with the release Q1 named:

   ```markdown
   ## Changes with go-llmprovider-sdk v1.28.0

   Release `<tag>` moves from go-llmprovider-sdk `v1.27.0` to `v1.28.0`
   ([0028-MADR](docs/decisions/0028-MADR-adopt-go-llmprovider-sdk-v1-28-0.md)).
   The SDK's changes are to caches and sign-in sessions this program does
   not close or wait for, so the hook, `configure`, `update` and the
   installers behave as in `v1.10.4`.
   ```

   It also inserts the contents line
   `- [Changes with go-llmprovider-sdk v1.28.0](#changes-with-go-llmprovider-sdk-v1280)`
   after `README.md:36`, asserted once, and asserts each new text once
   afterwards.
3. **Read the whole diff.** `git diff` covers `go.mod`, `go.sum`,
   `docs/README.md`, the two 0028 records' status lines, and `README.md`
   only under Q1 B or C.
4. **The gates:**
   * `make verify` exits 0, as in the MADR's probe: coverage at least
     80.0%, "No vulnerabilities found.", the ci-filters line, the tools'
     and the scripts' tests, and the six cross-builds;
   * `grep -c 74ab5dd376a4c864debb32b2d8abfc7856059ba3` gives 2 in
     `ci.yml` and 2 in the operations guide, as before;
   * `make check-records`, `make markdownlint` and `make disclosure-check`.
5. **Stage** those files, then run `make pre-add-check` on the staged
   snapshot: exit 0, "all modules verified", go-llmprovider-sdk `v1.28.0`
   resolved.
6. **Commit** on the owner's word naming `main`.

No new test is added (MADR D2).

### Phase 3: before the push

1. `make preflight` on the committed `HEAD`: "preflight passed".
2. `make preflight-hosts` on the committed `HEAD`, because `go.mod`
   changed. Each host prints `exit=0`. A host that cannot be reached is
   reported, not skipped.
3. The disclosure guard over `origin/main..HEAD`, as `AGENTS.md` gives it.
   It exits 0 with no output.

### Phase 4: push and CI

1. Push `main` on the owner's word naming `main`.
2. Wait for the runs on the pushed `HEAD` with `gh run list --commit <sha>`.
   `go.mod` is outside `ci.yml`'s `paths-ignore`, so `ci.yml` runs in full,
   and `docs.yml` runs for the records. Check:
   * "Go (quality contract)", the three native test jobs, "Build release
     / build" (`rehearsal=true`) and the identity jobs succeed;
   * "Publish GitHub Release" is skipped;
   * the Docs run succeeds.

   A failure stops the phase, and is recorded as a deviation.

### Phase 5: the release (only if Q1 is B or C)

As `0027-PLAN-adopt-go-llmprovider-sdk-v1-27-0.md` Phase 5, with the tag Q1
named in place of `v1.10.4` and `v1.10.4` in place of `v1.10.3`:

1. The owner's word makes and pushes an annotated tag on `main`'s `HEAD`.
2. The rule-suites API, or the push's output, shows no bypass of a ruleset.
3. Every job of the tag's run succeeds. "Publish GitHub Release / publish"
   names `publish-selfupdate-release.yml@74ab5dd376a4c864debb32b2d8abfc7856059ba3`.
4. The release is not a draft or prerelease, `immutable: true`, has the nine
   assets, and `releases/latest` names it.
5. `prepare-commit-msg-darwin-arm64` matches `SHA256SUMS`, and
   `gh attestation verify` with the publish workflow as signer exits 0, its
   source the tag at the tagged commit.
6. On this Mac's installed `v1.10.4`: `update --check` exits 10 naming the
   new tag, `update --yes` reports "release asset integrity verified",
   `version` names the tag, and `update --check` then exits 0.
7. A failed run follows the operations guide's "Failed-release recovery",
   recorded as a deviation.

### Phase 6: the records

1. Write the execution record: each phase's commands, exit statuses and the
   output the checks read; the commit hashes and run IDs; a dated entry for
   each deviation.
2. Set this PLAN to `status: complete`, and its index row to complete.
3. Run `make check-records`, `make markdownlint` and
   `make disclosure-check`. Commit and push on the owner's word, naming
   `main`. These are records only, so `docs.yml` runs.

## Verification

* Phase 0: the preconditions hold.
* Phase 2: `go.mod`/`go.sum` change on D1's lines alone; `make verify`,
  `make pre-add-check`, `make check-records`, `make markdownlint` and
  `make disclosure-check` exit 0; the whole diff was read.
* Phase 3: `make preflight` and `make preflight-hosts` pass on the
  committed `HEAD`, and the disclosure guard reports nothing.
* Phase 4: CI and Docs are green on `main`.
* Phase 5, if it runs: the release is immutable, has nine assets, and is the
  latest. Its checksum and attestation check out, and `update` moves
  `v1.10.4` to it.

## Rollout and Rollback

* **Rollout:** the commit on `main`. Under Q1 A, users receive it with the
  next release. Under B or C, with the tag, through `update` or the
  installers.
* **Rollback before a tag:** revert Phase 2's commit, and push on the
  owner's word.
* **Rollback after a release:** the release stays, immutable. Revert
  Phase 2's commit and release the revert as the next patch. A user can
  return to `v1.10.4` with `update --version v1.10.4 --yes` while 0026 keeps
  it.

## Execution record

Execution approved by the owner, 2026-11-07: "0028 q1: confirm ships in
v1.11.0 2. approved to execute 0028, mark madr accepted and approved to
proceed." The MADR is accepted, with Q1 answered A (no release of its own;
the move ships in `v1.11.0`). So D3's README section and Phase 5 do not run.

### Phase 0, 2026-11-07

A scratchpad script, read-only:

* `git rev-parse --path-format=absolute --git-path hooks` names the global
  hooks directory.
* `git status --short` lists this pair untracked, and also
  `0029-MADR-harden-the-hooks-generation-request.md` and
  `0029-PLAN-harden-the-hooks-generation-request.md` untracked
  (Deviation 1).
* `HEAD` and `origin/main` are both `870639c`.
* `go list -m -versions` ends at go-llmprovider-sdk `v1.28.0`;
  go-selfupdate-lib's newest tag is `v1.24.0`.
* The latest release is `v1.10.4`.

### Deviation 1, 2026-11-07: the 0029 pair is untracked too

* **Found.** Phase 0 step 2 expects only this pair in `git status`. The
  0029 pair was written after this PLAN, in the same session, and is
  untracked as well.
* **Decision.** The owner chose to leave the 0029 pair untracked. No step of
  this PLAN stages or edits it; its own PLAN's Phase 1 commits it. The other
  resolutions offered were committing the 0029 pair first, or including it
  in Phase 2's commit.
* **Effect.** None on this PLAN's steps. Phase 2 step 3's diff and step 5's
  staging name this PLAN's files only. No decision of the MADR changes.
* **Files added to scope:** none.

### Phase 1, 2026-11-07

* This PLAN's status became `in-progress`. The MADR was set to `accepted`,
  with Q1's answer under "Owner questions" and in D3.
* The two index rows went into `docs/README.md` after the 0027 rows: MADR
  "Adopt provider SDK v1.28.0", accepted 2026-11-07; PLAN "Implement
  provider SDK v1.28.0", in progress.

### Phase 2, 2026-11-07

* `go get github.com/maccavelli/go-llmprovider-sdk@v1.28.0` ("upgraded
  … v1.27.0 => v1.28.0") and `go mod tidy` exit 0. `go.mod` changes the one
  `require` line. `go.sum` replaces the module's two lines with
  `h1:Q1XMpmJzR8apF06whzRscIG2h3dRgLtvtJ65wWoaD34=` and the `/go.mod` hash
  `h1:dE+W7oDyvVmAxpSOuIJO3qSD3bDTh/olLt/TbByBE1M=`, as the MADR found.
  `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
  verified".
* Step 2, the README section, did not run: Q1 is A.
* The whole diff was read. The tracked changes are `go.mod`, `go.sum` and
  `docs/README.md`, besides this pair.
* `make verify` exits 0: "0 issues."; total coverage 83.9% (minimum 80.0%);
  "No vulnerabilities found."; "ci filters: 11 patterns, 11 reads checked, 0
  problems"; the tools' 42 tests and the scripts' 34 "OK"; the six
  cross-builds.
* `74ab5dd376a4c864debb32b2d8abfc7856059ba3` occurs 2 times in `ci.yml` and
  2 times in the operations guide, as before.
* `make check-records`, `make markdownlint` and `make disclosure-check`
  (134 files, 0 findings) exit 0. A direct scan of this pair for home paths,
  the account name and the host name finds nothing.
* Staged: `go.mod`, `go.sum`, `docs/README.md` and this pair; the 0029 pair
  stays untracked (Deviation 1). `make pre-add-check` on the staged snapshot
  exits 0: "all modules verified"; go-llmprovider-sdk `v1.28.0` "resolved
  from GitHub" with the `h1:` above; "0 issues."; "No vulnerabilities
  found.".
* The owner committed `3af1a8d`, "chore(deps): bump go-llmprovider-sdk to
  v1.28.0". It holds exactly the five staged files.

### Phase 3, 2026-11-07

* `make preflight` on `3af1a8d`: "preflight passed". Total coverage 83.9%,
  "No vulnerabilities found.", the ci-filters line, and the tools' 42 and
  the scripts' 34 tests.
* `make preflight-hosts` on `3af1a8d` refused: "commit first: preflight-hosts
  tests HEAD, and the tree has changes". `tools/hosts.py:161-167` rejects any
  `git status --porcelain` output, and the untracked 0029 pair counted
  (Deviation 2).
* After the owner committed the 0029 pair as `0184093`, `make
  preflight-hosts` on `0184093` exits 0:

  ```text
  linux: exit=0 (71s)
  wsl-ubuntu-24.04: exit=0 (41s)
  windows: exit=0 (9s)
  macos: exit=0 (10s)
  preflight-hosts: 4/4 hosts passed
  ```

  No host was unreachable. `0184093` adds only records to `3af1a8d`, so the
  Go code tested is this PLAN's.
* The disclosure guard, as `AGENTS.md` gives it, exits 0 with no output over
  `870639c..3af1a8d`, and again over `870639c..0184093`.

### Deviation 2, 2026-11-07: the cross-host run needs a clean tree

* **Found.** Phase 3 step 2 runs `make preflight-hosts` on the committed
  `HEAD`. The tool refuses while `git status --porcelain` lists anything,
  and Deviation 1 had left the 0029 pair untracked.
* **Decision.** The owner committed the 0029 pair alone as `0184093`, records
  only, under the bootstrap exception. The other resolutions offered were
  moving the two files aside for the run, or stopping before the push.
* **Effect.** The push carries `3af1a8d` and `0184093`. Deviation 1's
  resolution holds for this PLAN's own commits: no step of it staged or
  edited the 0029 pair. `0029-PLAN-harden-the-hooks-generation-request.md`
  Phase 1 step 3 expects its pair committed alone. No decision of the MADR
  changes.
* **Files added to scope:** none.

### Phase 4, 2026-11-07

* The owner pushed `main`, `870639c..0184093`.
* CI run `42429506312` on `0184093`: success. `go.mod` changed in the push,
  so `ci.yml` ran in full. "Go (quality contract)", the three native test
  jobs, "Build release / build" and the five identity jobs succeed;
  "Publish GitHub Release" is skipped.
* Docs run `42429505992`: success.

### Phase 5

Not run: Q1 is A. The move ships in `v1.11.0`, released by
`0029-PLAN-harden-the-hooks-generation-request.md` Phase 8.

### Phase 6, 2026-11-07

* This record, and this PLAN and its index row set to complete.
* Every Confirmation bullet of the MADR holds. Its release bullet, Q1 A, is
  met by making no tag here.
