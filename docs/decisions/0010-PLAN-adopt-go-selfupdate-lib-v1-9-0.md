---
status: in-progress
date: 2026-10-06
associated-madr: "0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md"
---
# Implement the move to go-selfupdate-lib v1.9.0

Associated MADR: [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md)

## Goal

* **D1:** `go.mod` requires `github.com/maccavelli/go-selfupdate-lib
  v1.9.0`, nothing else in the module graph changes version, and no opt-in
  feature is adopted.
* **D2:** `ci.yml` still publishes through `publish-selfupdate-release.yml`
  at `6deaa524cfb2…` (`v1.5.0`), and its comment says why.
* **D3:** a test holds `update --check --json`'s result at
  `schema_version` 2.
* **D4:** the README's "Self-Update" states the schema version and the
  warning rule.
* **D5:** release `v1.7.0` carries it, and the installed hook is updated to
  it.

Done means every item under Verification holds.

## Scope

### In scope

| Phase | Who | What |
| :--- | :--- | :--- |
| 0 | owner, then agent | answer Q1–Q5, accept the MADR, index both records |
| 1 | agent | D1, D2 and D3: the bump, the pin comment and the schema test |
| 2 | agent | D4: the README |
| 3 | owner, then agent | push, CI, tag `v1.7.0`, release checks, `update`, the live check |
| 4 | agent | close-out |

### Out of scope

* **Moving the publish pin** (D2's follow-up). It waits for a published
  release through `v1.9.0`'s workflow, and gets its own change.
* **Adopting `releasespec` and `build-selfupdate-release.yml`,** services,
  archives, codesign, or `AllowSpecialModeBits`.
* **Any change in go-selfupdate-lib,** including its 0013-PLAN B6
  rehearsal, which is that repository's work.
* **gobble-cli's pin.**
* **Push and tags,** which are the owner's.

## Rules for every phase

1. **Order.** Phases run in order. Each ends green, staged for the owner's
   commit. The agent does not commit to `main`. The owner commits with `git
   commit --no-edit`.
2. **Checks before each phase that changes code:**
   * `make verify` (lint, coverage at least 80.0%, govulncheck, workflow
     lint, build-all);
   * `make verify-staged` on the staged snapshot;
   * `go test -race -count=1 ./...`;
   * `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...`;
   * `go mod tidy -diff`;
   * `gofmt -l` on the changed Go files.
3. **Proofs on scratch copies,** never in the tree. Each new test is seen to
   fail on a planted break.
4. **Session tooling is Python.**
5. **Identifiers.** Nothing committed carries a hostname, an account name or
   a real-machine path.

## Implementation Steps

### Phase 0: accept the records

1. The owner answers Q1–Q5, or accepts the recommendations. The answers go
   into the MADR, which becomes `accepted`, and this PLAN `in-progress`.
2. `docs/README.md` indexes both records, in its list form, with their
   statuses.
3. The owner commits the records.

### Phase 1: the bump, the pin comment and the schema test

1. **D1.** `go get github.com/maccavelli/go-selfupdate-lib@v1.9.0`, then
   `go mod tidy`.
   * The diff must be exactly the library's `go.mod` line and its two
     `go.sum` lines.
   * `go list -m all` must otherwise match the baseline:
     * go-llmprovider-sdk `v1.2.1`;
     * `x/mod` `v0.40.0`, `x/sys` `v0.47.0`, `x/term` `v0.43.0` and
       `x/tools` `v0.49.0`.
   * Any other change is a deviation.
2. **D2, `.github/workflows/ci.yml`.**
   * The `uses:` line is unchanged.
   * The comment above it says that the publish workflow stays at
     `v1.5.0`'s commit while `go.mod` requires `v1.9.0`, until a release
     has been published through `v1.9.0`'s workflow (0010-MADR D2).
   * actionlint passes, through `make workflow-lint`.
3. **D3, `migration_test.go`.** `TestUpdateCheckJSONSchema` runs
   `runMain(t, []string{"update", "--check", "--json"}, releaseID,
   fixtureUpdater(t, fake))`, with the up-to-date fixture source. It wants:
   * exit 0;
   * stdout's last line to decode as `{"kind":"result", …}`;
   * the `result` object's `schema_version` to be `2`.
4. **The migration fixtures** under `testdata/migration/` stay byte-for-byte
   the library's. `TestMigrationByteForByte` passes unchanged.
5. **Proofs** (scratch copies), each of which must fail:
   * a copy whose `go.mod` and `go.sum` are held at `v1.5.0`:
     `TestUpdateCheckJSONSchema` fails with `schema_version` 1;
   * the test's wanted version changed to 3: the test fails. This shows it
     reads the field rather than passing on any value.
6. Run the checks, then stage the phase. `make verify-staged` must print
   `go-selfupdate-lib v1.9.0 resolved from GitHub`.

### Phase 2: documentation

1. **`README.md`, "Self-Update"** (`:350`):
   * the `--json` row says the result object carries `schema_version` 2;
   * the exit-status sentence says that an error after the new binary is
     in place is a `warning:` line, with exit 0, and that `1` is for an
     error before that.
2. **Checks:**
   * relative links and anchors in `README.md` resolve, proven on a planted
     bad link;
   * `TestReadmeCuratedModels` still passes;
   * `git diff --check`;
   * markdownlint is not configured here (0008, 0009).
3. Stage the phase.

### Phase 3: release and the live check (owner, then agent)

1. **The owner** commits and pushes. CI runs on Linux, macOS and Windows.
2. **The owner** tags `v1.7.0` on that commit. The tag's run builds,
   passes `make verify-release`, and publishes through the `v1.5.0`
   workflow.
3. **The agent** checks, on a download into scratch space, and records the
   output:
   * `shasum -a 256 -c SHA256SUMS` for the six binaries;
   * `version` prints `v1.7.0 (release) <revision>`;
   * `go version -m` lists `go-selfupdate-lib v1.9.0` and
     `go-llmprovider-sdk v1.2.1`.
4. **The owner** runs `prepare-commit-msg update` on the installed `v1.6.0`
   hook. This is the old library's updater installing the new release.
   * The agent reads the installed binary's build information, and records
     it without the install path.
   * The owner runs `prepare-commit-msg update --check`, which is the new
     library's updater, and reports its output. It should say up to date,
     with exit 0.
5. **The live check.** The owner makes one commit in a scratch repository,
   and reports whether the hook wrote a message. A throwaway repository
   avoids adding a file to this one, as 0009's check did.

### Phase 4: close-out

1. This PLAN is `complete`, the MADR `accepted`, and `docs/README.md` says
   so.
2. D2's follow-up is named, with what it waits for.

## Verification

* **V1.** `go list -m github.com/maccavelli/go-selfupdate-lib` prints
  `v1.9.0`, and the rest of the module graph is unchanged.
* **V2.** Every check in rule 2 passes at the end of Phases 1 and 2.
* **V3.** `TestUpdateCheckJSONSchema` passes, and fails on both Phase 1.5
  plants. `TestMigrationByteForByte` passes with the fixtures unchanged.
* **V4.** `ci.yml`'s `uses:` line is unchanged, and its comment states D2.
* **V5.** CI is green on Linux, macOS and Windows, on `main` and on the
  tag, and the tag's release is published.
* **V6.** The installed hook's build information lists
  `go-selfupdate-lib v1.9.0`. Its `update --check` reports up to date, and
  the owner's live commit got a message.
* **V7.** Nothing committed carries a hostname, an account name or a
  real-machine path.

## Rollout and Rollback

* **Rollout.**
  * Phases 1 and 2 are local commits by the owner.
  * Phase 3 publishes `v1.7.0`, and each user receives it through
    `update`.
* **Rollback.**
  * Before the tag, Phase 2 or Phase 1 reverts alone. Phase 1 takes the
    module back to `v1.5.0`, and its schema test goes with it.
  * After the tag, a problem is fixed forward in a patch release.
  * A user can return to `v1.6.0` with `prepare-commit-msg update
    --version v1.6.0 --yes`. On Windows, if a backup was still in use,
    `v1.6.0` may refuse the newer cleanup receipt (MADR, "What `update`
    does differently").

## Execution Record

The trial behind the MADR's measurements ran on scratch clones; nothing
was committed from it.

### Phase 0: accept the records (2026-10-06)

* **Approval.** The owner answered "questions: follow the recommendations." Execution of
  Phase 1 waits for the owner's explicit go-ahead.
* The MADR is `accepted`, and this PLAN `in-progress`.
* `docs/README.md` indexes both, in its list form, with those statuses.
* Both records and the index are staged for the owner's commit (rule 1).
