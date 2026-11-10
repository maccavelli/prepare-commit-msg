---
status: complete
date: 2026-11-10
associated-madr: "0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md"
---
# Implement the move to go-selfupdate-lib v1.25.0

Associated MADR: [0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md](0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md)

## Goal

* **D1:** `go.mod` requires `go-selfupdate-lib v1.25.0`. `go.mod` and
  `go.sum` change on that module's lines alone.
* **D2:** both workflow pins name `875dec3c…` (`v1.25.0`), with the comment
  above them and the operations guide's two quotes to match.
* **D3:** no Go code or test changes.
* **D4:** the README gains a "Changes with go-selfupdate-lib v1.25.0"
  section, its contents line, and one sentence each in "Self-Update"'s
  Security Note and in "Option B: Per-Repository Hook".
* **D5:** Q1 is answered A: Phase 5 releases `v1.11.1`, and the README
  section names it.

## Scope

### In scope

* `go.mod` and `go.sum`.
* `.github/workflows/ci.yml`: the two `uses:` lines and the comment above
  the build job.
* `docs/guides/cicd-operations.md`: its two quotes of those lines.
* `README.md`: under Q1 A, the new section and its contents line; under
  either answer, the Security Note and Option B sentences.
* `docs/README.md`: this pair's two index rows, and their status as it
  moves.
* This MADR and this PLAN, including the execution record.
* Under Q1 A: the tag, and the release its run publishes.

### Out of scope

* `update.go` and every other Go file (MADR D3, option B).
* go-llmprovider-sdk, the `golang.org/x` modules and the Go toolchain.
* Running `tools/prune_releases.py`. The pruning after a release follows the
  operations guide's "Controlled release", on the owner's word.
* The refusal's wording, which belongs to go-selfupdate-lib.
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
scratchpad. Every text edit is made by a scratchpad Python script that
asserts its anchor occurs exactly once before replacing it, and asserts the
new text once afterwards.

### Phase 0: preconditions (read-only)

1. `git rev-parse --path-format=absolute --git-path hooks` names the global
   hooks directory.
2. `git status --short` lists only this pair, untracked, or nothing if the
   owner committed it. `git rev-parse HEAD origin/main` gives `d0da9ae`, or
   the owner's commit of this pair on top of it.
3. `go list -m -versions` ends at go-selfupdate-lib `v1.25.0` and at
   go-llmprovider-sdk `v1.28.0`. `git ls-remote --tags` on go-selfupdate-lib
   peels `v1.25.0` to `875dec3c9d475ca47c4b012ca43742de832ef626`. If either
   library has a newer tag, stop and amend the MADR.
4. `git diff --stat v1.24.0 v1.25.0 -- .github scripts` in a clone of
   go-selfupdate-lib is empty, so D2's publish-path claim holds.
5. `gh api repos/maccavelli/prepare-commit-msg/releases/latest` names
   `v1.11.0`.

### Phase 1: the records

1. Set this PLAN to `status: in-progress`, and the MADR to `accepted` with
   Q1's answer recorded under "Owner questions" and in D5.
2. Add the two index rows to `docs/README.md` after the 0029 rows:
   * MADR: "Adopt go-selfupdate-lib v1.25.0", "accepted <date>";
   * PLAN: "Implement go-selfupdate-lib v1.25.0", "in progress".
3. These go into Phase 2's commit. If the owner has committed the pair
   alone, that is expected, not a deviation.

### Phase 2: the module, the pins and the README (D1–D4)

1. **The module.** Run
   `go get github.com/maccavelli/go-selfupdate-lib@v1.25.0`, then
   `go mod tidy`. Check:
   * `git diff go.mod` changes one line, `v1.24.0` to `v1.25.0`;
   * `git diff go.sum` replaces the module's two lines with
     `h1:K3IxUbldgxkssYQo3eW+b5gtJpBcz/CUjAMwGIMY8VE=` and the `/go.mod`
     hash `h1:DIqb0hzmxkYhVTj6iyaHpntadoHkeOkHWyTsdw+MBHE=`;
   * `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
     verified".

   Any other line changing stops the phase.
2. **The workflow pins (D2).** In `.github/workflows/ci.yml`, replace
   `74ab5dd376a4c864debb32b2d8abfc7856059ba3 # v1.24.0` with
   `875dec3c9d475ca47c4b012ca43742de832ef626 # v1.25.0`, asserting 2
   occurrences before and after. Replace the comment block at
   `ci.yml:110-116`, asserted once, with:

   ```yaml
       # Both workflows are pinned to the commit of go-selfupdate-lib v1.25.0,
       # the version go.mod requires. The publish path (the publish workflow
       # file and the scripts it runs) is unchanged from v1.24.0's, which has
       # published live (this repository's v1.11.0, run 42434050386), so
       # 0010-MADR D2's rule, as amended, holds
       # (docs/decisions/0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md
       # D2; docs/decisions/0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md D2, A1).
   ```

   In `docs/guides/cicd-operations.md`, make the same pin replacement,
   asserting 2 occurrences before and after.
3. **The README (D4).**
   * **Security Note**, `README.md:519`, asserted once: after "System paths
     (`/usr`, Homebrew Cellar, Nix store) are rejected." append " It also
     refuses a binary started through a symlink: run `update` on the file
     the link points to." *Amended by Deviation 1: the sentence goes on its
     own `> ` line in the same blockquote, not appended to that line.*
   * **Option B**, after the closing fence of the `ln -sf` block
     (`README.md:212-215`), asserted once, insert the paragraph "Run
     `update` on `/path/to/prepare-commit-msg`, not on the link: `update`
     refuses a binary started through a symlink."
   * **Under Q1 A only, the section.** Assert the v1.28.0 section's closing
     anchor once, `- **\`update\` and the installers** behave as in
     \`v1.10.4\`.`, followed by a blank line, `---` and `## CLI Reference`.
     Insert after that `---`, followed by its own `---`:

     ```markdown
     ## Changes with go-selfupdate-lib v1.25.0

     Release `v1.11.1` moves from go-selfupdate-lib `v1.24.0` to `v1.25.0`
     ([0030-MADR](docs/decisions/0030-MADR-adopt-go-selfupdate-lib-v1-25-0.md)).
     What you may notice:

     - **`update` started through a symlink is refused on Linux,** as it
       already was on macOS and Windows. Before, Linux replaced the file the
       link points to. Run the binary itself, such as
       `~/.global-git-hooks/prepare-commit-msg update`.
     - **The refusal names the link and its target,** and says to run the
       target.
     - **`update --json`** reports that refusal's `failed` event with the
       class `symlink-invocation`, where it was `error`.
     - **The hook, `configure` and the installers** behave as in `v1.11.0`.
     ```

     Insert the contents line
     `- [Changes with go-selfupdate-lib v1.25.0](#changes-with-go-selfupdate-lib-v1250)`
     after the v1.28.0 contents line, asserted once.
4. **Read the whole diff.** `git diff` covers `go.mod`, `go.sum`,
   `.github/workflows/ci.yml`, `docs/guides/cicd-operations.md`,
   `README.md`, `docs/README.md` and the two 0030 records. `update.go` and
   every other `.go` file are unchanged.
5. **The gates:**
   * `make verify` exits 0, as in the MADR's probe: "0 issues.", coverage at
     least 80.0%, "No vulnerabilities found.", the ci-filters line, the
     tools' and the scripts' tests, and the six cross-builds. `ci.yml` and
     `README.md` are read by `make workflow-lint` and `readme_test.go`, so
     both run here;
   * `grep -c 875dec3c9d475ca47c4b012ca43742de832ef626` gives 2 in `ci.yml`
     and 2 in the operations guide; `74ab5dd376a4c864debb32b2d8abfc7856059ba3`
     gives 0 in both;
   * `make check-records`, `make markdownlint` and `make disclosure-check`.
6. **Stage** those files, then run `make pre-add-check` on the staged
   snapshot: exit 0, "all modules verified", go-selfupdate-lib `v1.25.0`
   resolved.
7. **Commit** on the owner's word naming `main`.

No new test is added (MADR D3).

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
   `go.mod` and `ci.yml` are outside `ci.yml`'s `paths-ignore`, so `ci.yml`
   runs in full, and `docs.yml` runs for the records. Check:
   * "Go (quality contract)", the three native test jobs, "Build release
     / build" (`rehearsal=true`) and the identity jobs succeed;
   * the build job's log names
     `build-selfupdate-release.yml@875dec3c9d475ca47c4b012ca43742de832ef626`;
   * "Publish GitHub Release" is skipped;
   * the Docs run succeeds.

   A failure stops the phase, and is recorded as a deviation.

### Phase 5: the release (only if Q1 is A)

As `0028-PLAN-adopt-go-llmprovider-sdk-v1-28-0.md` Phase 5, with `v1.11.1`
in place of the tag and `v1.11.0` in place of `v1.10.4`:

1. The owner's word makes and pushes an annotated tag `v1.11.1` on `main`'s
   `HEAD`.
2. The rule-suites API, or the push's output, shows no bypass of a ruleset.
3. Every job of the tag's run succeeds. "Publish GitHub Release / publish"
   names `publish-selfupdate-release.yml@875dec3c9d475ca47c4b012ca43742de832ef626`.
4. The release is not a draft or prerelease, `immutable: true`, has the nine
   assets, and `releases/latest` names it.
5. `prepare-commit-msg-darwin-arm64` matches `SHA256SUMS`, and
   `gh attestation verify` with the publish workflow as signer exits 0, its
   source the tag at the tagged commit.
6. On this Mac's installed `v1.11.0`: `update --check` exits 10 naming
   `v1.11.1`, `update --yes` reports "release asset integrity verified",
   `version` names `v1.11.1`, and `update --check` then exits 0.
7. **The fix, on the Linux host** named by `available-dev-hosts`, in a
   scratch directory under its home that is removed afterwards. The check
   lives in the running binary, so the binary run is `v1.11.1`, the first
   built with go-selfupdate-lib `v1.25.0`. Download
   `prepare-commit-msg-linux-<arch>` of `v1.11.1`, check it against that
   release's `SHA256SUMS`, make it executable, and link to it from a second
   scratch directory. Run `update --force --yes` through the link by its
   path. It exits 1, its stderr starts with `selfupdate: refusing to update
   through a symlink:` and names the link and its target, and the target's
   checksum is unchanged. Then run `update --force --yes` on the target
   itself: it exits 0, and `version` names `v1.11.1`. Remove the scratch
   directories, and confirm they are gone.
8. A failed run follows the operations guide's "Failed-release recovery",
   recorded as a deviation.
9. The pruning, `tools/prune_releases.py`, runs only on the owner's word,
   as the operations guide's "Controlled release" gives it.

### Phase 6: the records

1. Write the execution record: each phase's commands, exit statuses and the
   output the checks read; the commit hashes and run IDs; a dated entry for
   each deviation.
2. Set this PLAN to `status: complete`, and its index row to complete.
3. Run `make check-records`, `make markdownlint` and
   `make disclosure-check`. Commit and push on the owner's word, naming
   `main`. These are records only, so `docs.yml` runs.

## Verification

* Phase 0: the preconditions hold, and both libraries are at their newest
  tags.
* Phase 2: `go.mod`/`go.sum` change on D1's lines alone; both pins and both
  quotes name `875dec3c…`; no `.go` file changes; `make verify`,
  `make pre-add-check`, `make check-records`, `make markdownlint` and
  `make disclosure-check` exit 0; the whole diff was read.
* Phase 3: `make preflight` and `make preflight-hosts` pass on the
  committed `HEAD`, and the disclosure guard reports nothing.
* Phase 4: CI and Docs are green on `main`, and the build rehearsal ran the
  new pin.
* Phase 5, if it runs: the release is immutable, has nine assets, and is the
  latest. Its checksum and attestation check out. `update` moves `v1.11.0`
  to it on this Mac. On Linux, `update` through a link is refused and leaves
  the target unchanged, and `update` on the target succeeds.

## Rollout and Rollback

* **Rollout:** the commit on `main`. Under Q1 A, users receive it with
  `v1.11.1`, through `update` or the installers. Under B, with the next
  release.
* **Rollback before a tag:** revert Phase 2's commit, and push on the
  owner's word.
* **Rollback after a release:** the release stays, immutable. Revert
  Phase 2's commit and release the revert as the next patch. A user can
  return to `v1.11.0` with `update --version v1.11.0 --yes` while 0026's
  window keeps it.
* **A user refused by the new check** runs the target the message names.
  No rollback is needed for that.

## Execution record

Execution approved by the owner, 2026-11-09: "set madr to accepted, approve
the plan, proceed." Q1 was answered A the same day (release `v1.11.1`).

### Phase 0, 2026-11-09

A scratchpad script, read-only:

* `git rev-parse --path-format=absolute --git-path hooks` names the global
  hooks directory.
* `git status --short` lists only this pair, untracked.
* `HEAD` and `origin/main` are both `d0da9aeba298e8bba92b9a9bd85c14a270494c49`.
  The script first reported this check failed: it passed both revisions to
  `git rev-parse --short`, which takes one ("fatal: Needed a single
  revision"). Run without `--short`, both name `d0da9ae`. A fault of the
  check, not of the precondition.
* `go list -m -versions` and `git ls-remote --tags` both end at
  go-selfupdate-lib `v1.25.0` and go-llmprovider-sdk `v1.28.0`.
  `v1.25.0` peels to `875dec3c9d475ca47c4b012ca43742de832ef626`.
* `git diff --stat v1.24.0 v1.25.0 -- .github scripts` in the library's
  clone is empty.
* The latest release is `v1.11.0`.

### Phase 1, 2026-11-09

* The MADR was set to `accepted`, with the owner's approval quoted under
  "Owner questions"; this PLAN to `in-progress`.
* The two index rows went into `docs/README.md` after the 0029 rows: MADR
  "Adopt go-selfupdate-lib v1.25.0", accepted 2026-11-09; PLAN "Implement
  go-selfupdate-lib v1.25.0", in progress.

### Deviation 1, 2026-11-09: the Security Note line exceeds MD013

* **Found.** Phase 2 step 5: `make markdownlint` exited 2 with
  `README.md:523:201 error MD013/line-length Line length [Expected: 200;
  Actual: 266]`. Step 3 appended 82 characters to the Security Note, a line
  of 184. Not pre-existing: `make markdownlint` passed before the edit.
* **Decision.** The owner chose option 1: the new sentence goes on its own
  `> ` line in the same blockquote, so it renders as one paragraph and the
  existing line is unchanged. The other resolution offered was rewrapping
  the whole Security Note at about 100 characters. Raising or disabling the
  rule was not offered.
* **Effect.** Step 3's Security Note bullet is annotated above. MADR D4
  says the note "adds" the sentence and does not change. No other step
  changes.
* **Files added to scope:** none.

### Phase 2, 2026-11-09

* `go get github.com/maccavelli/go-selfupdate-lib@v1.25.0` ("upgraded
  … v1.24.0 => v1.25.0") and `go mod tidy` exit 0. `go.mod` changes the one
  `require` line. `go.sum` replaces the module's two lines with
  `h1:K3IxUbldgxkssYQo3eW+b5gtJpBcz/CUjAMwGIMY8VE=` and the `/go.mod` hash
  `h1:DIqb0hzmxkYhVTj6iyaHpntadoHkeOkHWyTsdw+MBHE=`, as the MADR found.
  `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
  verified".
* A scratchpad Python script made steps 2 and 3's edits, asserting each
  anchor once and each new text once: both `uses:` pins and the comment in
  `ci.yml`, both quotes in the operations guide, and the README's contents
  line, section, Security Note and Option B sentences. Deviation 1 then
  moved the Security Note sentence to its own line.
* The whole diff was read. The changes are `go.mod`, `go.sum`,
  `.github/workflows/ci.yml`, `docs/guides/cicd-operations.md`,
  `README.md`, `docs/README.md` and this pair. No `.go` file changed.
* After Deviation 1, `make verify` exits 0: "0 issues."; total coverage
  84.9% (minimum 80.0%); "No vulnerabilities found."; "ci filters: 11
  patterns, 11 reads checked, 0 problems"; the tools' 42 tests and the
  scripts' 34 "OK"; the six cross-builds. `make check-records`,
  `make markdownlint` and `make disclosure-check` exit 0.
* `875dec3c9d475ca47c4b012ca43742de832ef626` occurs 2 times in `ci.yml` and
  2 times in the operations guide; `74ab5dd376a4c864debb32b2d8abfc7856059ba3`
  0 times in each.
* Staged: the eight files above. `make pre-add-check` on the staged
  snapshot exits 0: "all modules verified"; go-selfupdate-lib `v1.25.0`
  "resolved from GitHub" with the `h1:` above; "0 issues."; "No
  vulnerabilities found.".
* The owner committed `121879f`, "chore(deps): pin go-selfupdate-lib to
  v1.25.0". It holds exactly the eight staged files.

### Phase 3, 2026-11-09

* `make preflight` on `121879f`: "preflight passed". Total coverage 84.9%,
  "No vulnerabilities found.", the ci-filters line, and the tools' 42 and
  the scripts' 34 tests.
* `make preflight-hosts` on `121879f`, with a clean tree, exits 0:

  ```text
  linux: exit=0 (68s)
  wsl-ubuntu-24.04: exit=0 (53s)
  windows: exit=0 (4s)
  macos: exit=0 (5s)
  preflight-hosts: 4/4 hosts passed
  ```

  No host was unreachable. The Windows and macOS logs each show `ok`, not
  cached, for all five packages.
* The disclosure guard, as `AGENTS.md` gives it, exits 0 with no output
  over `d0da9ae..121879f`.

### Phase 4, 2026-11-09

* The owner pushed `main`, `d0da9ae..121879f`.
* CI run `42724069596` on `121879f`: success. "Go (quality contract)", the
  three native test jobs, "Build release / build" and the five identity
  jobs succeed; "Publish GitHub Release" is skipped. The build job's log
  names `build-selfupdate-release.yml@875dec3c9d475ca47c4b012ca43742de832ef626`,
  with `rehearsal=true`.
* Docs run `42724069203`: success.

### Deviation 2, 2026-11-09: the release waits for the next library releases

* **Found.** After Phase 4, before Phase 5's tag. The owner: "2 of the
  libraries this includes, including selfupdate-lib are being updated with
  enhancements and bugfixes. i am going to wait until we can pin latest
  again before i tag a release. i will let you know."
* **Decision.** The owner's. Phase 5 does not run now. No tag is made until
  the owner says so, after go-selfupdate-lib and go-llmprovider-sdk are
  moved to their next releases. The MADR records it as A1.
* **Effect.** Phase 5 is on hold, and with it Phase 6's completion: this
  PLAN stays `in-progress`. When the owner releases, the release ships this
  move with the next library moves. If it is not `v1.11.1`, the record that
  makes the release corrects this pair's README section and runs Phase 5's
  checks against that tag, step 7's Linux check included. Phases 1–4 stand.
* **Files added to scope:** none.

### Phase 5, 2026-11-10

Run once, with `0031-PLAN-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md`
Phase 5, as Deviation 2 foresaw. The release is `v1.11.1`, as this pair's
README section named it; 0031-MADR D4 folds that section into one for both
libraries, with go-selfupdate-lib `v1.26.0` in place of `v1.25.0`.

* The owner made and pushed the annotated tag `v1.11.1` on `423321d`. CI run
  `42869640645` succeeds, "… / sign" skipped as `sign` is unset, and
  publishes through
  `publish-selfupdate-release.yml@7c677842dface4b14b65b2e8b46aa0a7543bed13`
  (`v1.26.0`), whose publish path is this PLAN's `875dec3c…`'s.
* The release is immutable, has the nine assets and is the latest; its
  darwin-arm64 checksum and attestation check out; `update` moved this
  Mac's installed `v1.11.0` to `v1.11.1`.
* Step 7, on the Linux host: `update` through a symlink is refused, exit 1,
  class `symlink-invocation`, the target unchanged; on the target itself it
  succeeds. The line it prints is Deviation 3's.
* The pruning waits for the owner's word.

The commands and the output each check read are in 0031-PLAN's execution
record.

### Deviation 3, 2026-11-10: step 7's stderr line

* **Found.** Step 7 says stderr "starts with `selfupdate: refusing to update
  through a symlink:`". The released `v1.11.1` prints
  `update failed: selfupdate: prepare-commit-msg: selfupdate: refusing to
  update through a symlink: <link> links to <target>; start <target> itself
  to update it`: `cli.Exit` adds `update failed: `, and the Updater adds
  `selfupdate: <product>: `, as at `v1.25.0`.
* **Decision.** The owner's, 2026-11-10, "option 2", recorded as 0031-PLAN
  Deviation 1 and 0031-MADR A1: the line printed is the one expected,
  matched exactly; the doubled prefix is raised as go-selfupdate-lib
  issue #11.
* **Effect.** Step 7 is met with that line. This MADR's Confirmation says
  the refusal names the link and its target, which holds.
* **Files added to scope:** none.

### Phase 6, 2026-11-10

* This PLAN and its index row are complete, in 0031-PLAN's Phase 6.
