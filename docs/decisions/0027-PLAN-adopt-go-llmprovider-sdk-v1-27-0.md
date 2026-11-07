---
status: complete
date: 2026-11-06
associated-madr: "0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md"
---
# Implement the move to go-llmprovider-sdk v1.27.0 and release v1.10.4

Associated MADR: [0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md](0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md)

## Goal

* **D1:** `go.mod` requires `go-llmprovider-sdk v1.27.0`. `go.mod` and
  `go.sum` change on that module's lines alone.
* **D2:** no other code, test, workflow or toolchain change.
* **D3:** the README has a "Changes with go-llmprovider-sdk v1.27.0"
  section and its table-of-contents line.
* **D4:** `v1.10.4` is released, immutable, and checked end to end.
* **D5:** the release is not held for
  `0026-MADR-keep-five-stable-releases.md`. That record's first pruning
  runs afterwards, by its own PLAN.

## Scope

### In scope

* `go.mod` and `go.sum`.
* `README.md`: the new section, after the `v1.26.1` / `v1.24.0` one, and its
  contents line after `README.md:35`.
* `docs/README.md`: this pair's two index rows, and their status as it
  moves.
* This MADR and this PLAN, including the execution record.
* The tag `v1.10.4`, and the release its run publishes.

### Out of scope

* `catalog.SortByRank`, `catalog.LabelName` and `catalog.LabelNote`
  (MADR D2).
* go-selfupdate-lib, `.github/workflows/ci.yml`, the operations guide and
  the Go pins.
* `0026-MADR-keep-five-stable-releases.md` and its PLAN. They stay
  untracked, and no step here stages them. Their first pruning is theirs to
  run after Phase 5 (MADR D5).
* The text of earlier records and earlier README "Changes with" sections.

### Permissions

Each of these needs the owner's word in the turn it happens, under
`sysadmin-git-branch-gates`:

* a commit, naming `main`;
* a push, naming `main`;
* making the tag `v1.10.4`, and pushing it.

The global hook writes every message through `git commit --no-edit`.

## Implementation Steps

Every phase ends with `make check-records` and `make markdownlint`. A
phase that stages `go.mod`/`go.sum` also runs `make pre-add-check`. Each
exit status is captured before any filter, with long output in the session
scratchpad.

### Phase 0: preconditions (read-only)

1. `git rev-parse --path-format=absolute --git-path hooks` names the global
   hooks directory.
2. `git status --short` lists only four untracked records:
   `0026-MADR-…`, `0026-PLAN-…`, `0027-MADR-…` and this PLAN.
   `git rev-parse HEAD origin/main` gives the same commit, `e56c281` or a
   later one the owner pushed.
3. `go list -m -versions github.com/maccavelli/go-llmprovider-sdk` ends at
   `v1.27.0`. `git ls-remote --tags` on go-selfupdate-lib ends at
   `v1.24.0`. If either is newer, stop and amend the MADR.
4. `gh api repos/maccavelli/prepare-commit-msg/releases/latest` names
   `v1.10.3`. The immutable-releases setting is enabled. The only ruleset
   is `prepare-commit-msg-main`. `proxy.golang.org`'s `@v/list` for this
   module does not list `v1.10.4`.

### Phase 1: the records

1. Set this PLAN to `status: in-progress`.
2. Add the two index rows to `docs/README.md` after the 0025 rows, leaving
   room for 0026's, which its own PLAN adds:
   * MADR: "Adopt provider SDK v1.27.0", "accepted 2026-11-06";
   * PLAN: "Implement provider SDK v1.27.0 and release v1.10.4",
     "in progress".
3. Run `make check-records`, `make markdownlint` and
   `make disclosure-check`. The disclosure check reads tracked files only,
   so also scan the two 0027 files directly for home paths, the account
   name and the host name. Each finds nothing.
4. Stage exactly `docs/decisions/0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md`,
   `docs/decisions/0027-PLAN-adopt-go-llmprovider-sdk-v1-27-0.md` and
   `docs/README.md`. `git diff --cached --name-only` lists those three.
5. Commit on the owner's word naming `main`. This is records only, under
   the bootstrap exception.

### Phase 2: the module and the README (D1 to D3)

One commit: the section names the release that carries the pin.

1. **The module.** Run
   `go get github.com/maccavelli/go-llmprovider-sdk@v1.27.0`, then
   `go mod tidy`. Check:
   * `git diff go.mod` changes one line, `v1.26.1` to `v1.27.0`;
   * `git diff go.sum` replaces two lines, the module's `h1:` and `/go.mod`
     hashes, with the MADR's values:
     `h1:7tSn5DyOaGh+/eGC9wQe0xR497Vq2MtPPAMoCulTPfM=` and
     `h1:dE+W7oDyvVmAxpSOuIJO3qSD3bDTh/olLt/TbByBE1M=`;
   * `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
     verified".

   Any other line changing stops the phase.
2. **The README section.** A scratchpad Python script asserts this anchor
   occurs once in `README.md`, then inserts the section after its `---`:

   ```text
   - **The installers** behave as in `v1.10.2`.

   ---

   ## CLI Reference
   ```

   The section, between that `---` and `## CLI Reference`, followed by its
   own `---`:

   ```markdown
   ## Changes with go-llmprovider-sdk v1.27.0

   Release `v1.10.4` moves from go-llmprovider-sdk `v1.26.1` to `v1.27.0`
   ([0027-MADR](docs/decisions/0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md)).
   What you may notice:

   - **`configure`:**
     - Grok's `grok-4.20` models are ranked by when xAI released them, below
       `grok-4.3`, not above every Grok model, so the model menu's last
       place may show another Grok model. The default model, Grok 4.6, and
       the fallbacks `configure --yes` picks are as before;
     - some model descriptions in the menus are reworded so that none places
       a model in time: Grok 4.5 is no longer "previous flagship", and
       Claude Sonnet 5.5 and Opus 5.5 are no longer "latest".
   - **The hook, `update` and the installers** behave as in `v1.10.3`.
   ```

3. **The contents line.** The same script asserts once, then inserts after
   it, the line
   `- [Changes with go-llmprovider-sdk v1.26.1 and go-selfupdate-lib v1.24.0](#changes-with-go-llmprovider-sdk-v1261-and-go-selfupdate-lib-v1240)`.
   The new line is
   `- [Changes with go-llmprovider-sdk v1.27.0](#changes-with-go-llmprovider-sdk-v1270)`.
   The script asserts each new text occurs once afterwards.
4. **Read the whole diff:** `git diff` covers `go.mod`, `go.sum` and
   `README.md`, and nothing else.
5. **The gates:**
   * `make verify` exits 0, as in the MADR's probe: coverage at least
     80.0%, "No vulnerabilities found.", the ci-filters line, the tools'
     and the scripts' tests, and the six cross-builds;
   * `TestReadmeCuratedModels`, `TestCurrentDocumentation` and
     `TestCurrentDocumentationDetectsBrokenFixtures` pass (in
     `make verify`), and `make markdownlint` passes with the link-fragment
     rule, which checks the new contents line;
   * `grep -c 74ab5dd376a4c864debb32b2d8abfc7856059ba3` gives 2 in
     `ci.yml` and 2 in the operations guide, as before;
   * `make check-records` and `make disclosure-check`.
6. **Stage** `go.mod`, `go.sum` and `README.md`, then run
   `make pre-add-check` on the staged snapshot: exit 0, "all modules
   verified", go-llmprovider-sdk `v1.27.0` resolved.
7. **Commit** on the owner's word naming `main`. The hook writes the
   subject. In 0024 it named the SDK's version where the release was
   meant; if this one does, tell the owner before the push.

No new test is added (MADR D2). The behaviour that changes is the SDK's,
covered by its own `0070` tests. What this app passes to the SDK is
covered by its existing tests.

### Phase 3: before the push

1. `make preflight` on the committed `HEAD`: "preflight passed".
2. `make preflight-hosts` on the committed `HEAD`, because `go.mod`
   changed. Each host prints `exit=0`, and the Windows host downloads
   go-llmprovider-sdk `v1.27.0`. A host that cannot be reached is
   reported, not skipped.
3. The disclosure guard over `origin/main..HEAD`, as `AGENTS.md` gives it.
   It exits 0 with no output.

### Phase 4: push and CI

1. Push `main` on the owner's word naming `main`. The push carries Phase
   1's and Phase 2's commits.
2. Wait for the runs on the pushed `HEAD`, with
   `gh run list --commit <sha>`. `README.md` is outside `ci.yml`'s
   `paths-ignore`, so `ci.yml` runs in full, and `docs.yml` runs for the
   records. Check:
   * "Go (quality contract)", the three native test jobs, "Build release
     / build" (`rehearsal=true`, using `…@74ab5dd3…`) and the identity
     jobs succeed;
   * "Publish GitHub Release" is skipped, as on any push that is not a
     tag;
   * the Docs run succeeds.

   A failure stops the phase, and is recorded as a deviation.

### Phase 5: the release `v1.10.4` (D4)

1. **Tag.** On the owner's word, make the tag on `main`'s `HEAD`, annotated
   as `v1.10.3` is: `git tag -a v1.10.4 -m v1.10.4`. On the owner's word
   to push it, run `git push origin v1.10.4`, saving its output.
2. **The push output** holds no "Bypassed rule violations". This is the
   check `0025-PLAN-remove-ruleset-bypass-actors.md` deferred to the next
   release. If one appears, stop and report: a ruleset on tags exists
   again.
3. **The run.** Find the tag's run with `gh run list --branch v1.10.4`, and
   wait for it. Check:
   * every job succeeds;
   * "Publish GitHub Release / publish" succeeds, its log naming
     `publish-selfupdate-release.yml@74ab5dd376a4c864debb32b2d8abfc7856059ba3`;
   * "… / sign" is skipped, as `sign` is unset.

   Record its URL.
4. **The release.** `gh api …/releases/tags/v1.10.4` shows:
   * not a draft and not a prerelease, and `immutable: true`;
   * nine assets: the six binaries, `SHA256SUMS`, `install.sh` and
     `install.ps1`;
   * `releases/latest` names `v1.10.4`.
5. **One binary.** Download `prepare-commit-msg-darwin-arm64` and
   `SHA256SUMS` to the scratchpad. Check:
   * `shasum -a 256 -c` against `SHA256SUMS` says "OK";
   * `gh attestation verify` with
     `--signer-workflow maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml`
     exits 0, signer `…@74ab5dd3…`, source `refs/tags/v1.10.4` at the
     tagged commit;
   * the binary carries the new labels: `grep -c -a -F` finds "second in
     the Grok CLI's built-in list" once and "previous flagship" not at all.
     That is the observable sign that `v1.27.0` was built in. On
     2026-11-06 the scratch build at `v1.27.0` gave 1 and 0, and the
     installed `v1.10.3` gave 0 and 1 **[probed]**.
6. **The update path.** On this Mac's installed hook, `v1.10.3`:
   * `update --check` reports `v1.10.3 -> v1.10.4` and exits 10;
   * `update --yes` reports the update and "release asset integrity
     verified", and exits 0;
   * `version` names `v1.10.4 (release)` and the tagged commit;
   * `update --check` then reports "up to date (v1.10.4)" and exits 0.
7. **If the run fails** before it publishes, follow the operations guide's
   "Failed-release recovery": check the proxy's `@v/list`, and re-tag only
   if `v1.10.4` is not listed. Otherwise tag the fix as `v1.10.5`. Record a
   deviation either way.

### Phase 6: the records

1. Write the execution record:
   * each phase's commands, exit statuses and the output the checks read;
   * the commit hashes, the CI and release run IDs, and the release URL;
   * a dated entry for each deviation.
2. Set this PLAN to `status: complete`, and its index row to complete.
3. Note for 0026: `v1.10.4` is verified, so
   `0026-PLAN-keep-five-stable-releases.md` Phase 3 step 1 will find it.
   Its expected plan then keeps `v1.10.0` to `v1.10.4`, and deletes 8
   releases and 23 tags (MADR D5). Edit nothing in 0026.
4. Run `make check-records`, `make markdownlint` and
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
* Phase 5: the push shows no bypass. The release is immutable, with nine
  assets, and is the latest. The checksum, the attestation and the label
  text check out. `update` moves `v1.10.3` to `v1.10.4`.

## Rollout and Rollback

* **Rollout:** the commits on `main`, then the tag `v1.10.4`. Users take
  it with `update`, or the installers' latest download.
* **Rollback before the tag:** revert Phase 2's commit, and push on the
  owner's word.
* **Rollback after the release:** `v1.10.4` stays. Its release is
  immutable, and its tag name cannot be reused. Revert Phase 2's commit,
  and release the revert as `v1.10.5`. Under 0026, `v1.10.4` stays in the
  window until five newer stable releases exist.
* **A user** can return to `v1.10.3` with `update --version v1.10.3 --yes`
  while that release is kept. Nothing in the configuration changes between
  the two, so either reads the other's `config.json`.

## Execution record

Execution approved by the owner, 2026-11-06: "committed, proceed". The
owner had committed the two 0027 records as `33bc3d0`.

### Phase 0, 2026-11-06

A scratchpad script, read-only, exit 0, "clean":

* `git rev-parse --path-format=absolute --git-path hooks` names the global
  hooks directory.
* `git status --short` lists only the two untracked 0026 records. `HEAD`
  is `33bc3d0`, one commit ahead of `origin/main` (`e56c281`) and none
  behind.
* `go list -m -versions` ends at go-llmprovider-sdk `v1.27.0`, and
  go-selfupdate-lib's newest tag is `v1.24.0`.
* The latest release is `v1.10.3`. The immutable-releases setting is
  `{"enabled": true, "enforced_by_owner": false}`. The only ruleset is
  `24842254`, `prepare-commit-msg-main`, on a branch. The proxy's
  `@v/list` has 27 versions, without `v1.10.4`, and the remote has no tag
  `v1.10.4`.

### Phase 1, 2026-11-06

* `33bc3d0`, committed by the owner, holds the MADR and this PLAN only, with
  this PLAN `proposed` and no index rows (Deviation 1).
* In the tree: this PLAN's status became `in-progress`. The two index rows
  went into `docs/README.md` after the 0025 rows: MADR "Adopt provider SDK
  v1.27.0", accepted 2026-11-06; PLAN "Implement provider SDK v1.27.0 and
  release v1.10.4", in progress.

### Phase 2, 2026-11-06

* `go get github.com/maccavelli/go-llmprovider-sdk@v1.27.0` ("upgraded
  … v1.26.1 => v1.27.0") and `go mod tidy` exit 0. `go.mod` changes the one
  `require` line. `go.sum` replaces the module's two lines with
  `h1:7tSn5DyOaGh+/eGC9wQe0xR497Vq2MtPPAMoCulTPfM=` and the `/go.mod` hash
  `h1:dE+W7oDyvVmAxpSOuIJO3qSD3bDTh/olLt/TbByBE1M=`, as the MADR found.
  `go mod tidy -diff` is empty, and `go mod verify` prints "all modules
  verified".
* A scratchpad script asserted the README's anchor (the `v1.26.1`
  section's last bullet, `---` and `## CLI Reference`) and the `v1.26.1`
  contents line once each. It inserted the section and the contents line,
  and asserted each new text once. The whole diff was read: `go.mod`,
  `go.sum`, `README.md`, `docs/README.md` and this PLAN's status.
* `make verify` exits 0: total coverage 83.9% (minimum 80.0%), "No
  vulnerabilities found.", "ci filters: 11 patterns, 11 reads checked, 0
  problems", the tools' 25 tests and the scripts' 34 "OK", and the six
  cross-builds. `TestReadmeCuratedModels`, `TestCurrentDocumentation` and
  `TestCurrentDocumentationDetectsBrokenFixtures` pass.
* `make check-records`, `make markdownlint` (8 files, 0 issues) and
  `make disclosure-check` (129 files, 0 findings) exit 0.
  `74ab5dd376a4c864debb32b2d8abfc7856059ba3` occurs 2 times in `ci.yml`
  and 2 times in the operations guide, as before.
* `make pre-add-check` on the staged `go.mod`, `go.sum` and `README.md`
  exits 0: "all modules verified"; go-llmprovider-sdk `v1.27.0` "resolved
  from GitHub" with the `h1:` above; "0 issues."; "No vulnerabilities
  found.".
* The agent stopped before committing, and offered three resolutions for
  the Phase 1 records (Deviation 1).
* The owner committed `cb8558b`, "chore(deps): bump go-llmprovider-sdk to
  v1.27.0". It holds Phase 2's three files, the index rows, this PLAN's
  status, and both 0026 records.

### Deviation 1, 2026-11-06: the owner's commits carry the records differently

* **Found.** Phase 1 has one records commit: the pair, the index rows and
  this PLAN's `in-progress` status. The owner's `33bc3d0` holds the pair
  alone. The agent offered three resolutions: (1) the index rows and status
  in Phase 2's commit, recommended; (2) a separate records commit first;
  (3) leaving them to Phase 6.
* **Decision.** The owner committed `cb8558b`, which is resolution 1. It
  also holds `0026-MADR-keep-five-stable-releases.md` and
  `0026-PLAN-keep-five-stable-releases.md`, both `proposed`. This PLAN's
  Scope had kept them untracked.
* **Effect.** The 0027 records are indexed from `cb8558b` onward. The 0026
  records are committed with no index rows, which `0026-PLAN` Phase 0 adds.
  `make check-records` passes, because it checks names, pairs and links,
  not the index. No decision of the MADR changes.
* **Files added to scope:** the two 0026 records, as committed by the
  owner, unchanged by the agent.

### Phase 3, 2026-11-06

* `make preflight` on `cb8558b`: "preflight passed". Total coverage 83.9%,
  "No vulnerabilities found.", the ci-filters line, and the tools' 25 and
  the scripts' 34 tests "OK".
* `make preflight-hosts` on `cb8558b`, exit 0:

  ```text
  linux: exit=0 (100s)
  wsl-ubuntu-24.04: exit=0 (36s)
  windows: exit=0 (3s)
  macos: exit=0 (5s)
  preflight-hosts: 4/4 hosts passed
  ```

  The Linux build stamps `v1.10.3-8-gcb8558b`. windows and macos run
  `go test ./...` on the bundle of `HEAD`, 5 packages `ok` each; linux 8.
  No host was unreachable. Phase 3 step 2 expected the Windows host to
  download go-llmprovider-sdk `v1.27.0`. Its log shows no download, so the
  module was already in that host's cache. The bundle's `go.mod` requires
  `v1.27.0`, so the tests ran against it.
* The disclosure guard, as `AGENTS.md` gives it, exits 0 with no output
  over `e56c281..cb8558b` (`33bc3d0`, `cb8558b`).

### Phase 4, 2026-11-06

* The owner committed the execution record of Phases 0 to 3 as `28e1a7e`,
  and pushed `main`, `e56c281..28e1a7e`. The agent ran the disclosure
  guard over that range afterwards, since `28e1a7e` followed its first run:
  exit 0, no output.
* CI run `42319794531` on `28e1a7e`: success. The push's diff holds
  `go.mod` and `README.md`, so `ci.yml` ran in full, though `28e1a7e`
  alone is a record. "Go (quality contract)", the three native test jobs,
  "Build release / build" (`rehearsal=true`, its log naming
  `build-selfupdate-release.yml@74ab5dd376a4c864debb32b2d8abfc7856059ba3`)
  and the five identity jobs succeed. "Publish GitHub Release" is skipped.
  Docs run `42319794328`: success.

### Phase 5, 2026-11-06

* The owner made and pushed the tag `v1.10.4` (Deviation 2). It is
  annotated, tag object `58f0c04`, message `v1.10.4`, and peels to
  `28e1a7e`, which is `origin/main`.
* **The bypass check** that `0025-PLAN-remove-ruleset-bypass-actors.md`
  deferred: the agent did not see the push's output. The rule-suites API
  records each push a ruleset evaluates. It holds no suite for
  `refs/tags/v1.10.4`, and every suite of the day is a `refs/heads/main`
  push with result `pass`. No ruleset applies to the tag, so none could be
  bypassed. `v1.10.3`'s push, by contrast, was recorded as a bypass.
* **Release run `42321034534`:**
  <https://github.com/maccavelli/prepare-commit-msg/actions/runs/42321034534>.
  It succeeds. "Go (quality contract)", the three native test jobs, "Build
  release / build" and the five identity jobs succeed. "Publish GitHub
  Release / publish" succeeds, and its log names
  `publish-selfupdate-release.yml@74ab5dd376a4c864debb32b2d8abfc7856059ba3`.
  "… / sign" is skipped, as `sign` is unset.
* **The release:**
  <https://github.com/maccavelli/prepare-commit-msg/releases/tag/v1.10.4>.
  * Published 2026-11-07T00:59:03Z; not a draft and not a prerelease;
    `immutable: true`. `releases/latest` names it.
  * Its nine assets: the six binaries, `SHA256SUMS`, `install.sh` and
    `install.ps1`.
* **One binary,** `prepare-commit-msg-darwin-arm64`, downloaded to the
  session scratchpad:
  * its SHA-256, `8f824503c5a84533…`, is `SHA256SUMS`'s;
  * `gh attestation verify --signer-workflow
    maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml`
    exits 0. The signer is
    `…publish-selfupdate-release.yml@74ab5dd376a4c864debb32b2d8abfc7856059ba3`,
    and the source is `refs/tags/v1.10.4` at
    `28e1a7ecfa9df9ba1e8188bbf1fe9e917c598991`;
  * it holds "second in the Grok CLI's built-in list" once and "previous
    flagship" not at all, so `v1.27.0`'s labels are built in.
* **The update path,** on this Mac's installed hook, `v1.10.3 (release)
  79f894aea2e6`:
  * `update --check`: "update available: v1.10.3 -> v1.10.4", exit 10;
  * `update --yes`: "release asset integrity verified", "updated v1.10.3
    -> v1.10.4", exit 0;
  * `version`: `v1.10.4 (release) 28e1a7ecfa9d`. The installed file is
    10265122 bytes with SHA-256 `8f824503c5a84533…`, the asset's;
  * `update --check` again: "up to date (v1.10.4)", exit 0.
* The proxy's `@v/list` still has 27 versions, without `v1.10.4`, an hour
  after the tag. It lists a version once something asks for it.

### Deviation 2, 2026-11-06: the owner tagged and pushed

* **Found.** Phase 5 step 1 has the agent make and push the tag on the
  owner's word, and step 2 reads the push's output for a bypass message.
  The owner made and pushed `v1.10.4` themselves, so the agent had no push
  output.
* **Decision.** Step 2's check was made from the rule-suites API, which
  records what a push's ruleset evaluation found. It shows no evaluation
  of the tag, and so no bypass. The tag's shape was checked from the
  repository: annotated, message `v1.10.4`, on `origin/main`. No decision
  of the MADR changes.
* **Files added to scope:** none.

### Phase 6, 2026-11-06

* This record, and this PLAN and its index row set to complete.
* **For 0026:** `v1.10.4` is verified.
  `0026-PLAN-keep-five-stable-releases.md` Phase 3 step 1 will find 13
  releases and 28 tags. By its MADR's D1 to D4, the expected plan keeps
  `v1.10.0` to `v1.10.4`. It deletes the 8 releases `v1.2.0` to `v1.9.0`
  and their tags, and the 15 tags with no release: 8 releases and 23
  tags. Nothing in 0026 was edited.
* Every Confirmation bullet of the MADR holds. This PLAN is complete.
