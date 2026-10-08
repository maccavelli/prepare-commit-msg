---
status: in-progress
date: 2026-10-08
associated-madr: "0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md"
---
# Implement the move to go-llmprovider-sdk v1.3.2 and go-selfupdate-lib v1.10.1

Associated MADR: [0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md](0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md)

## Goal

* **D1:** `go.mod` requires `go-llmprovider-sdk v1.3.2` and
  `go-selfupdate-lib v1.10.1`, nothing else in the module graph changes
  version, and no code or opt-in feature changes.
* **D2:** `ci.yml` publishes through `publish-selfupdate-release.yml` at
  `a0a26b6ecf66…` (`v1.10.0`), and its comment says why it is not
  `v1.10.1`'s.
* **D3:** a test sends Gemini's real refused-key reply through the SDK and
  the hook's model loop, and requires the run to stop at the first model.
* **D4:** the README describes what `v1.3.2` and `v1.10.1` change for a
  user.
* **D5:** release `v1.8.0` carries it, published through `a0a26b6`, and the
  installed hook is updated to it.

Done means every item under Verification holds.

## Scope

### In scope

| Phase | Who | What |
| :--- | :--- | :--- |
| 0 | owner, then agent | answer Q1–Q3, accept the MADR, index both records |
| 1 | agent | D1, D2 and D3: the bumps, the pin and the refused-key test |
| 2 | agent | D4: the README |
| 3 | owner, then agent | push, CI, tag `v1.8.0`, release checks, `update`, the live checks |
| 4 | agent | close-out |

### Out of scope

* Moving the pin to `v1.10.1`'s commit (0011-MADR D2's follow-up).
* go-selfupdate-lib `v1.11.0`.
* Adopting `releasespec`, `build-selfupdate-release.yml`, the installers,
  services, archives or codesign.
* Any change in go-llmprovider-sdk or go-selfupdate-lib.
* gobble-cli's versions.
* Push and tags, which are the owner's.

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
6. **Keys.** A live check reads a key from the environment, checked for
   presence only, never printed. A deliberately invalid key is a literal
   that is not a key.
7. **Deviations stop and prompt,** with evidence and resolutions, and are
   recorded here, and in the MADR when a decision or asserted fact changes.

## Implementation Steps

### Phase 0: accept the records

1. The owner answers Q1–Q3, or accepts the recommendations. The answers go
   into the MADR, which becomes `accepted`, and this PLAN `in-progress`.
2. `docs/README.md` indexes both records, in its list form, with their
   statuses.
3. The owner commits the records.

### Phase 1: the bumps, the pin and the refused-key test

1. **D3's test first, red at SDK `v1.2.1`.** In `main_test.go`,
   `TestRunStopsOnGeminiRefusedKey`:
   * an `httptest.Server` answers every request with HTTP 400 and Gemini's
     refused-key body, as go-llmprovider-sdk's live check captured it:
     `[{"error":{"code":400,"message":"API key not valid. Please pass a
     valid API key.","status":"INVALID_ARGUMENT","details":[{"@type":
     "type.googleapis.com/google.rpc.ErrorInfo","reason":"API_KEY_INVALID",
     "domain":"googleapis.com"}]}}]`;
   * a configuration with `gemini` active, a primary and one fallback
     model, and a key that is not a key;
   * `generateWithRetry` swapped, as `main_test.go:398-423` swaps it, for a
     function that builds `newKeyProvider(llmprovider.ProviderGemini, key,
     model, llmprovider.WithBaseURL(server.URL))` and calls the real
     `generateText` on it;
   * it requires the server to see exactly one request, and the error to
     contain "authentication failed for gemini".

   Run on the tree, still at `v1.2.1`: it fails with two requests and "all
   models for gemini failed". The FAIL line goes in the Execution Record.
   If it passes at `v1.2.1`, that is a deviation: the MADR's first SDK row
   would be wrong.
2. **D1.** `go get github.com/maccavelli/go-llmprovider-sdk@v1.3.2
   github.com/maccavelli/go-selfupdate-lib@v1.10.1`, then `go mod tidy`.
   * The diff must be exactly the two libraries' `go.mod` lines and their
     `go.sum` lines.
   * `go list -m all` must otherwise match the baseline: `x/mod v0.40.0`,
     `x/sys v0.47.0`, `x/term v0.43.0`, and `x/tools` as before.
   * Any other change is a deviation.
3. **D3 green.** `TestRunStopsOnGeminiRefusedKey` passes.
4. **D2, `.github/workflows/ci.yml`.**
   * The `uses:` line names
     `publish-selfupdate-release.yml@a0a26b6ecf66f51c19e9fea0f665c76ca5e99e4c
     # v1.10.0`; its `with:` and the job's permissions are unchanged.
   * The comment above it says that `go.mod` requires `v1.10.1`, that the
     workflow is `v1.10.0`'s, the newest with a live publish on record
     (go-selfupdate-lib 0014-PLAN I7 step 3), and that it moves to
     `v1.10.1`'s commit once a publish through that is on record
     (0010-MADR D2; 0011-MADR D2).
   * actionlint passes, through `make workflow-lint`.
5. **Unchanged:** `TestMigrationByteForByte`, `TestUpdateCheckJSONSchema`
   (`schema_version` 2), and the files under `testdata/migration/`.
6. **Proofs** (scratch copies), each of which must fail:
   * `go.mod` and `go.sum` held at SDK `v1.2.1` (step 1's run on the tree
     counts);
   * the stop rule at `main.go:284` removed: two requests, and "all models
     for gemini failed".
7. Run the checks, then stage the phase. `make verify-staged` must resolve
   both versions.

### Phase 2: documentation

1. **`README.md`:**
   * a section "Changes with go-llmprovider-sdk v1.3.2" after the `v1.2.1`
     one, naming the release `v1.8.0` and
     [0011-MADR](0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md),
     with the MADR's SDK table in a user's words: a refused Gemini key stops
     the run; a malformed key fails at once; a cut-off answer is reported
     as incomplete; the OAuth store's lock; `configure`'s cancellation and
     masked input;
   * "Self-Update": a backup the update could not restore is kept beside
     the binary as `.<base>.selfupdate-kept-<n>` (B1), and a binary
     replaced during the update is not overwritten (B2);
   * its line in the table of contents;
   * any other place the README names a library version.
2. **Checks:**
   * `readme_test.go` passes;
   * relative links and anchors in `README.md` resolve, proven on a planted
     bad link;
   * the identifier scan of the changed files.
3. Stage the phase.

### Phase 3: release and the live checks (owner, then agent)

1. **The owner** commits and pushes Phases 1–2. CI is green on Linux, macOS
   and Windows.
2. **The owner** tags `v1.8.0` and pushes it.
3. **The agent** checks, read-only:
   * CI on the tag, and its release job, published through `a0a26b6`;
   * the release: six binaries and `SHA256SUMS`, published, latest, not a
     draft;
   * `gh attestation verify` on one binary.
   A failed publish is a deviation: the tag stays, and the fix is a new
   commit and `v1.8.1`.
4. **The owner** runs `update` on the installed hook.
5. **The agent** checks the installed hook:
   * `version`, and its build information, name `v1.8.0`, SDK `v1.3.2` and
     go-selfupdate-lib `v1.10.1`;
   * `update --check` reports it up to date;
   * in a scratch repository, a staged change gets a message from the
     owner's configured provider;
   * in a scratch repository, with `GEMINI_API_KEY` set to a literal that
     is not a key and two Gemini models configured, the hook stops at the
     first model with "authentication failed for gemini".

### Phase 4: close-out

1. The Execution Record holds every phase's output and deviations.
2. This PLAN is `complete`, and `docs/README.md` says so.
3. Stage for the owner's commit.

## Verification

* **V1. D1:** the `go.mod` and `go.sum` diff is the two libraries only.
* **V2. D2:** the tag's release was published through `a0a26b6`.
* **V3. D3:** the test fails at SDK `v1.2.1` and with the stop rule
  removed, and passes at `v1.3.2`.
* **V4. The unchanged contracts:** the migration and `update` tests pass
  unchanged.
* **V5. The checks** of rule 2 at the end of Phases 1 and 2, and CI on
  `main` and the tag.
* **V6. Live:** the installed hook's versions, a real message, and the
  refused-key stop.

## Rollout and Rollback

* **Rollout.** Records, Phase 1 and Phase 2 commits, pushed by the owner;
  then the `v1.8.0` tag; users receive it through `update`.
* **Rollback, before the tag:** each phase reverts alone. Reverting Phase 1
  returns both modules and the pin, and removes the test, together.
* **Rollback, after the tag:** fix forward in `v1.8.x`. If the publish
  itself fails, the pin returns to `6deaa52` in a new commit, and `v1.8.1`
  is tagged. A user who returns to `v1.7.0` keeps a working hook; the OAuth
  store's files are compatible, though the two versions' locks do not
  exclude each other.

## Execution Record

### Phase 0: accept the records (2026-10-08)

* **Answers.** "Questions follow recommendations, commit to main then
  proceed", 2026-10-08: Q1 the pin to `a0a26b6` (`v1.10.0`), Q2 the
  refused-key test, Q3 release `v1.8.0`. The MADR is `accepted`, and
  this PLAN `in-progress`.
* **Index.** `docs/README.md` lists both records with their statuses.
* **Checks.** Every relative link in the two records and the index
  resolves (a planted bad link was caught); the identifier scan of the
  three files: 0 hits.
* **Commit.** On the owner's ask in the same message, the agent
  committed the records to `main` as `56d7df7`, with `git commit
  --no-edit`, after checking that the repository's hooks directory
  chains to the global `prepare-commit-msg` and `pre-commit` hooks.

### Phase 1: the bumps, the pin and the refused-key test (2026-10-08)

**D3's test, red at SDK `v1.2.1`** (step 1). `TestRunStopsOnGeminiRefusedKey`
(`main_test.go`) on the tree, `go.mod` still at `v1.2.1`:

```text
--- FAIL: TestRunStopsOnGeminiRefusedKey (0.00s)
    main_test.go:490: err = all models for gemini failed, last error: model fallback failed: llmprovider: invalid request: gemini HTTP 400: [{"error":{"code":400,"message":"API key not valid. Please pass a valid API key.","status":"INVALID_ARGUMENT","details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo","reason":"API_KEY_INVALID","domain":"googleapis.com"}]}}] after 2 request(s); want the run stopped with authentication failed for gemini after 1
```

That is the MADR's first SDK row, measured in this repository: `v1.2.1`
classifies the refusal `invalid request`, and the fallback is asked with the
same key.

**D1, the bump** (step 2). `go get …go-llmprovider-sdk@v1.3.2
…go-selfupdate-lib@v1.10.1`, then `go mod tidy`:

* `go.mod`: the two `require` lines, `v1.2.1` → `v1.3.2` and `v1.9.0` →
  `v1.10.1`;
* `go.sum`: the two libraries' four lines;
* `go list -m all`, before and after: the two libraries' lines differ, and
  no other.

**D3 green** (step 3):

```text
--- PASS: TestRunAnalyzer_FailureKinds (0.03s)
--- PASS: TestRunStopsOnGeminiRefusedKey (0.00s)
--- PASS: TestMigrationByteForByte (0.00s)
--- PASS: TestUpdateCheckJSONSchema (0.00s)
```

**D2, the pin** (step 4). `ci.yml`'s `uses:` line names
`publish-selfupdate-release.yml@a0a26b6ecf66f51c19e9fea0f665c76ca5e99e4c #
v1.10.0`, the commit `git rev-parse 'v1.10.0^{commit}'` gives in
go-selfupdate-lib. The comment above it says why it is not `v1.10.1`'s.
Its `with:` and the job's permissions are unchanged: the diff's only
non-comment line is the `uses:` line.

**Proofs** (step 6):

| Proof | Result |
| :--- | :--- |
| SDK held at `v1.2.1` | step 1's run on the tree, above |
| the stop rule at `main.go:284` planted out (`if false && …`), on a scratch copy at `v1.3.2` | FAIL: `err = all models for gemini failed, last error: model fallback failed: llmprovider: authentication failed: gemini HTTP 400 API_KEY_INVALID: API key not valid. … after 2 request(s)` |

The plant shows the test needs both halves: the SDK's classification,
which on `v1.3.2` says `authentication failed`, and the loop's rule, which
stops on it.

**Checks** (step 7):

| Check | Result |
| :--- | :--- |
| `make verify` | exit 0: `0 issues.`, workflow lint, build-all of six binaries, `total coverage: 84.5% (minimum 80.0%)`, `No vulnerabilities found.` |
| `go test -race -count=1 ./...` | 5 packages ok |
| `CGO_ENABLED=0 go vet ./...`, for linux, darwin and windows | 0 each |
| `go mod tidy -diff` | 0 |
| `gofmt -l main_test.go` | no output |
| `make verify-staged` | exit 0: `go-llmprovider-sdk v1.3.2 resolved from GitHub (h1:GbDQ7ibn+…)`, `go-selfupdate-lib v1.10.1 resolved from GitHub (h1:ShT43ip2…)`, `0 issues.`, `No vulnerabilities found.` |
| identifier scan of the five staged files | 0 hits |

**Staged** for the owner's commit: `.github/workflows/ci.yml`, `go.mod`,
`go.sum`, `main_test.go` and this PLAN. No other file changed; `make
verify`'s `dist/` and `coverage.out` are ignored.
