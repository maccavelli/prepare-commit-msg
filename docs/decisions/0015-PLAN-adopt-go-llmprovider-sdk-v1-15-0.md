---
status: in-progress
date: 2026-10-19
associated-madr: "0015-MADR-adopt-go-llmprovider-sdk-v1-15-0.md"
---
# Implement the move to go-llmprovider-sdk v1.15.0 and the ChatGPT plan sign-in

Associated MADR: [0015-MADR-adopt-go-llmprovider-sdk-v1-15-0.md](0015-MADR-adopt-go-llmprovider-sdk-v1-15-0.md)

## Goal

* **D1:** `go.mod` requires `go-llmprovider-sdk v1.15.0`.
* **D2:** the wizard offers "Use your ChatGPT plan with this app",
  under the name `prepare-commit-msg`.
* **D3:** each retry the hook waits for prints one line to stderr.
* **D4:** the two tests the move breaks assert the product again, and a
  new test pins the `kilo-auto` result.
* **D5:** the README states every change a user sees.
* **D6, D7:** `v1.10.0` is released, and issues 1 and 2 are closed with a
  comment each, on the owner's ask.

## Scope

### In scope

* `go.mod`, `go.sum`; `main.go`; `internal/ui/setup.go`; the tests in
  `internal/ui` and the root package that the steps name.
* `README.md` (its five curated rows in Phase 1, Deviation 1), `docs/README.md`.
  `docs/architecture.md` stays as it is:
  its runtime flow (`docs/architecture.md:16-32`) names neither the
  wizard's options nor retries.

### Out of scope

* go-selfupdate-lib, which stays at `v1.15.2`, and the workflow pins.
* `configure --yes`'s credentials: it stays API-key-only.
* A logout command; the store and the hook's provider path, which take a
  plan session unchanged.

## Rules for every phase

* **Branch.** All work is on `main`, and no branch is created. The agent
  stages, and the owner commits and pushes (0014-PLAN's amendment of
  2026-10-13).
* **Fail first.** Every new or changed test is run on a scratch copy of
  the tree without its fix, and the failure is recorded in the Execution
  Record. The tree is never dirtied for the experiment.
* **Pre-add.** Before each commit:
  * `gofmt -l` on every staged `.go` file;
  * `make verify-staged`;
  * `make check-records`;
  * `make markdownlint`.

  A file that fails is not committed.
* **Phase gate.** Before each commit:
  * `go test -race -count=1 ./...`;
  * `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...`;
  * `go mod tidy -diff`;
  * `make verify`.

  Each exit status is captured before any filter.
* **Stage, then the owner commits.** The agent stages exactly the phase's
  files and shows `git diff --cached --stat`. The owner runs `git commit
  --no-edit` and pushes. One commit per phase, and the next phase starts
  after the owner has committed the last.
* **Deviations.** Stop, record a dated entry here (and in the MADR when a
  decision or a fact changes), and wait for approval.

## Implementation Steps

### Phase 0: records

1. The owner reviews the MADR, answers Q1–Q3, accepts it, and approves
   this PLAN. Answers that change a decision become MADR amendments first.
2. `docs/README.md` indexes the pair.

### Phase 1: the move, and the tests that follow the SDK (D1, D4)

1. `go get github.com/maccavelli/go-llmprovider-sdk@v1.15.0`, then `go mod
   tidy`. Assert that only the SDK's lines change in `go.mod` and `go.sum`.
2. **The listing cache.** A helper in `internal/ui`'s tests, `testKey(t)`,
   returns ~~`"test-key-" + t.Name()`~~ `"test-key-<name>-<n>"`, `n` a
   process-wide counter *(Deviation 2, 2026-10-19)*. Every test that lists through an
   injected client uses it in place of a shared literal key, so that no
   two tests share a cached listing (the SDK's 0041-MADR D6).
3. **The fallback.** `TestDiscoverModels_EmptyRecommendationFallsBack`
   lists ~~one Kilo model whose `architecture.output_modalities` is
   `["embeddings"]`, the body probed for the MADR~~ one tool-using Kilo
   text model whose `expiration_date` is 2026-01-01 *(Deviation 3,
   2026-10-19)*. It still requires the curated catalog.
4. **The `kilo-auto` result.** `TestDiscoverModels_KiloTiersOnly_0015`
   lists `kilo-auto/small` and `kilo-auto/balanced`, and requires
   `["kilo-auto/balanced", "kilo-auto/small"]` (0041-MADR A4).
5. **Fail first:**
   * at `v1.15.0` without steps 2 and 3, the two original failures recur;
   * at `v1.12.0`, the new test of step 4 fails, since that version gave
     the curated catalog;
   * on a scratch copy with `discoverModels`'s empty-list fallback
     removed, step 3's test fails.
6. **Order independence.** `go test -count=3 -shuffle=on ./internal/ui/`
   passes.
7. **The curated rows** *(Deviation 1, 2026-10-19).* `README.md`'s rows
   for `kilo`, `huggingface`, `together`, `opencode-zen` and
   `opencode-go` take
   `catalog.Static`'s lists at `v1.15.0`, in its order, the first marked
   "(recommended)" as now. `TestReadmeCuratedModels` failed first: at
   `v1.15.0` with the old rows (step 5's first run).

### Phase 2: the ChatGPT plan sign-in (D2)

1. `ui.SetupOptions` gains `AgentName string`, documented as the app's
   own name for OpenAI's Sign in with ChatGPT for agents.
2. `main.go` sets `AgentName: AppTitle` where it builds `SetupOptions`,
   beside `ClientOptions`.
3. `runSetupInteractive` passes `AgentName: opts.AgentName` to
   `wizard.Options`.
4. **Tests:**
   * `TestSetup_OffersChatGPTPlan_0015`: an interactive run that chooses
     OpenAI reaches the method menu, which lists the SDK's label for
     `llmprovider.AuthChatGPTPlan`, last.
   * `TestSetup_NoAgentNameNoPlan_0015`: with `AgentName` empty, the menu
     does not list it.
   * The existing OpenAI method choices keep their indices. The new method
     is last; the tests that choose by index pass unchanged.
5. **Fail first:** on a scratch copy without step 3, the first test fails.

### Phase 3: the retry line (D3)

1. `generateText` builds its wrapper with `llmprovider.WithRetryNotify(p,
   policy, notify)`. `notify` writes one line to a package variable
   `retryOut` (`os.Stderr` by default):

   ```text
   prepare-commit-msg: <provider>: retrying in <wait> (attempt <n> of <max>): <kind>
   ```

   * `<wait>` is rounded to tenths of a second.
   * `<kind>` is `rate limited` for `ErrRateLimited`, `unavailable` for
     `ErrProviderUnavailable`, `cannot reach` for a `*url.Error`, and
     `error` otherwise. It never quotes the provider's message.
   * ~~If the owner answers Q2 "only when stderr is a terminal", the line
     is written only then.~~ *(Phase 0, 2026-10-18: the owner answered
     "on every retry"; the line is always written.)*
2. **Tests,** with a fake provider and `retryOut` captured:
   * a 429, then a success: one line, matching the format, with
     `attempt 1 of <max>` and `rate limited`;
   * a success: no line;
   * a terminal failure: no line.
3. **Fail first:** on a scratch copy with `WithRetry` restored, the first
   test fails.

### Phase 4: documents (D5)

1. **`README.md`:**
   * **Configuration:** the plan sign-in in the wizard; the notice; the
     plan terms the MADR lists, and how the app keeps each.
   * **"With Ollama, name the models"** (`README.md:578`) is narrowed to
     Ollama before `v0.30.0`.
   * **The retry line,** where the README describes retries.
   * **A section, "Changes with go-llmprovider-sdk v1.15.0":**
     * the facts-based output limit;
     * the ten-minute listing cache, and the stale listing after a
       transient failure;
     * Claude's live listing; Ollama's filter;
     * the `kilo-auto` tiers;
     * the sign-in and the retry line;
     * the downgrade cost of a plan session record.
2. `readme_test.go` (`TestReadmeCuratedModels`) passes.

### Phase 5: release `v1.10.0`, the live checks and the issues (D6, D7)

1. Before each push, the agent runs `AGENTS.md`'s pre-push disclosure
   guard over the outgoing commits.
2. CI on `main` is green on Linux, macOS and Windows.
3. **Live checks, the owner present,** on a scratch build with the
   configuration under a scratch home (`HOME` / `XDG_CONFIG_HOME`, or
   `APPDATA` / `USERPROFILE` on Windows). The installed hook and its
   configuration are not read or written; the installed hook's SHA-256 is
   taken before and after, and must not change.
   * the wizard: OpenAI, then "Use your ChatGPT plan with this app", with
     the owner signing in. The notice is shown, and the session is saved
     in the scratch OAuth store;
   * the hook, on a scratch repository with one staged file: a
     Conventional Commit message through the plan session;
   * the retry line, against a local server that answers 429 once:
     `configure --yes --provider openai --base-url <server>` with a
     made-up key in the environment, then the hook.
4. The owner tags `v1.10.0` and pushes the tag. CI publishes. The agent
   checks the release assets, runs `update` on the installed hook, and
   confirms `version` names the SDK at `v1.15.0`.
5. **The issues, on the owner's word,** each comment written to a file,
   scanned for identifiers, and shown first:
   * issue 1: done by 0014-MADR D2 (`ModelLimit`, `catalog.WithLimit`),
     released in `v1.9.0`; closed;
   * issue 2: released in `v1.10.0`; closed.

   Each URL goes into the Execution Record.

## Verification

| # | Criterion | Check | Phase |
| :--- | :--- | :--- | :--- |
| V1 | `go.mod` pins `v1.15.0`, and only the SDK's lines changed | `git diff go.mod go.sum`; `go mod tidy -diff` | 1 |
| V2 | The three broken tests pass (Deviation 1), and the new ones failed first | the phase gate; the fail-first records | 1 |
| V3 | The tests pass in any order | `go test -count=3 -shuffle=on ./internal/ui/` | 1 |
| V4 | The wizard offers the plan sign-in only with the app's name | Phase 2's tests | 2 |
| V5 | A retry prints one line, in the format; nothing else does | Phase 3's tests | 3 |
| V6 | The README states the changes | review; `readme_test.go`; `make markdownlint` | 4 |
| V7 | CI passes on all three platforms | the CI run on `main` | 5 |
| V8 | The plan sign-in, the hook and the retry line work live | Phase 5 step 3's log | 5 |
| V9 | `v1.10.0` is released, and issues 1 and 2 are closed | the release; the comment URLs | 5 |

## Rollout and Rollback

* Users receive `v1.10.0` through `update`.
* Rollback before the tag: revert the phase's commit.
* After the tag, a `v1.10.1` restores the previous behaviour. A user who
  signed in with the plan signs in again, or uses an API key, since the
  older SDK cannot refresh a plan session (the MADR's D5).

## Execution Record

### Phase 0: records (2026-10-18)

* The pair and its index rows were committed by the owner as `1dc2434`,
  who then said "proceed, prompt me questions".
* **The owner's answers,** each the recommended one: Q1 yes, offer the
  sign-in; Q2 the line on every retry; Q3 `v1.10.0`; no host test
  before the release, CI's three platforms being enough. No decision
  changes. The MADR records each answer and is `accepted`; this PLAN
  is `in-progress`.
* **Corrections, no decision changed:**
  * the records' date is the local date, 2026-10-18, where the drafts
    had 2026-10-19;
  * Phase 4 said "four terms" where the MADR lists five; it now names
    the MADR's list;
  * Phase 3's step for a Q2 answer of "only on a terminal" is struck.

### Deviation 1 (2026-10-19): a third failing test, the README's curated rows

* **Found.** Phase 1 step 5's first run, on a scratch copy at `v1.15.0`
  with no test changed: `go test -count=1 ./...` exit 1, with
  `TestRunSetupInteractive_ShortList` (`menu_limit_test.go:93`, "the wizard
  did not list through the injected client"),
  `TestDiscoverModels_EmptyRecommendationFallsBack` (`setup_test.go:187`,
  `["kilo-auto/balanced" "kilo-auto/small"]`), and a third the MADR did
  not name: `TestReadmeCuratedModels` (`readme_test.go:51`), on the rows
  of `kilo`, `huggingface`, `together`, `opencode-zen` and `opencode-go`.
* **Not pre-existing.** At `1dc2434` (`v1.12.0`) it passes. The MADR's
  probe copy fails it too: the MADR miscounted.
* **Resolution, the owner's (2026-10-19):** the rows change in Phase
  1, as step 7. `README.md` joins Phase 1's files; Phase 4 still writes
  the rest. The MADR's Amendment 1 records the fact.
* **Corrected the same day.** The deviation first named four rows: the
  read of the failing run's log was cut at four lines. The full log has
  five `readme_test.go:51` lines; the fifth is `opencode-go`, reordered.
  The resolution is unchanged; the records now name five. The MADR's
  D5 also said "four terms", as Phase 4 did before Phase 0; it now
  names the MADR's list.

### Deviation 2 (2026-10-19): a key per run, not per test name

* **Found.** Step 6, `go test -count=3 -shuffle=on ./internal/ui/`, exit 1:
  `menu_limit_test.go:63`, `menu_limit_test.go:93` and `setup_test.go:219`
  ("did not list through the injected client"), each on its second and
  third runs. `go test -count=2 -run 'TestDiscoverModels_KiloTiersOnly_0015$'`
  fails once, on the second run.
* **Cause.** Step 2's key, `"test-key-" + t.Name()`, is the same for each
  run of one test, so a repeat is served the listing the first run cached.
  Step 2 as written caused it; nothing before it.
* **Resolution, the owner's (2026-10-19):** `testKey` appends a
  process-wide atomic counter, so every run of every test lists with a key
  of its own. Step 2 is annotated. No MADR decision changes; D4's "a key of
  its own" holds.

### Deviation 3 (2026-10-19): the fallback test reached the SDK's catalog, not the fallback

* **Found.** Step 5's third check: on a scratch copy with
  `discoverModels`'s `len(cat.Recommended) == 0` branch removed, the
  fallback test, with step 3's embedding listing, passed. It guarded
  nothing.
* **Cause.** At `v1.15.0`, a model that writes no text is not usable, and
  with none usable the SDK gives its static catalog itself (`Live` false).
  The MADR's probe saw the catalog and credited the fallback. A text model
  without tools goes the same way. A usable model with a caveat does not:
  one past its `expiration_date`, or with `-contributor` in its id, is
  listed (`Live` true) and recommended nothing; with the fallback removed,
  `discoverModels` gives `[]`.
* **Resolution, the owner's (2026-10-19):** the test lists one tool-using
  Kilo text model whose `expiration_date` is 2026-01-01. Step 3 is
  annotated; the MADR's Amendment 2 corrects the fact.

### Phase 1: the move, and the tests (2026-10-19)

* **Deviations 1, 2 and 3,** above, each on the owner's choice.
* **Commits.** The owner committed the phase through Deviation 2 as
  `e63b118`, before Deviation 3 was found. Its fallback test still listed
  the embedding model, which guarded nothing. A second commit carries
  Deviation 3, its MADR amendment, and this record; `e63b118` is not
  rewritten.
* **Step 1.** `go get …@v1.15.0`, `go mod tidy`: exit 0. `go.mod` changes
  one line, `go.sum` two, all the SDK's.
* **Steps 2–4, 7:** `testKey` (with Deviation 2's counter) in the six
  calls that list through an injected client and assert on it:
  * `menu_limit_test.go`'s two short-list tests;
  * `TestRunSetupInteractive_ListingUsesInjectedClient`;
  * `TestDiscoverModels_UsesInjectedClient`;
  * `TestRunSetupNonInteractive_RefusedKey`, in place of its shared
    `bad-key`;
  * the two Kilo tests, through `kiloListing`.

  A key that is only saved, or that is used with no listing, stays as it was.
  The five README rows take `catalog.Static`'s lists.
* **Step 5, fail first,** each on a scratch copy, never in the tree:
  * at `v1.15.0`, no test changed: exit 1, the three failures of
    Deviation 1;
  * at `v1.12.0` (a clone of `1dc2434`) with these tests:
    `TestDiscoverModels_KiloTiersOnly_0015` fails (`setup_test.go:226`,
    the curated catalog where the tiers are wanted);
  * with `discoverModels`'s empty-list branch removed:
    `TestDiscoverModels_EmptyRecommendationFallsBack` fails
    (`setup_test.go:211`, `discoverModels(kilo) = []`). With step 3's first
    fixture it passed; that is Deviation 3;
  * Deviation 2's counter: without it, `go test -count=2 -run
    'TestDiscoverModels_KiloTiersOnly_0015$'` fails once; with it, exit 0.
* **Step 6.** `go test -count=3 -shuffle=on ./internal/ui/`: exit 0.
* **The phase gate,** each exit status captured before any filter:
  * `go test -race -count=1 ./...`: 0;
  * `go vet` for linux, darwin, windows: 0, 0, 0;
  * `go mod tidy -diff`: 0;
  * `make verify`: 0 (golangci-lint 0 issues; coverage 83.5%, minimum 80%;
    govulncheck "No vulnerabilities found."; the six builds);
  * `make check-records`: 0; `make markdownlint`: 0.

### Phase 2: the ChatGPT plan sign-in (2026-10-19)

* **No deviation.**
* **Steps 1–3.** `SetupOptions.AgentName`; `main.go` sets it to
  `AppTitle`; `runSetupInteractive` passes it to `wizard.Options`.
* **Step 4.** The two tests run the wizard to OpenAI's method menu, read
  it from stdout (`captureStdout`, a pipe; no product change for the test),
  choose the API key, and compare the labels with the SDK's own label for
  `llmprovider.AuthChatGPTPlan`. The existing tests that choose a method by
  number pass unchanged.
* **Step 5, fail first,** each on a scratch copy:
  * without step 3 (a clone of `3c28d39`, the phase applied but the
    pass-through): `TestSetup_OffersChatGPTPlan_0015` fails
    (`setup_test.go:1119`, five methods, no plan sign-in);
  * with the wizard always given a name (`AgentName: "planted"`):
    `TestSetup_NoAgentNameNoPlan_0015` fails (`setup_test.go:1129`, the
    plan sign-in sixth).
* **Not covered by a unit test:** `main.go`'s `AgentName: AppTitle`. It is
  one field in the configure command's options. Phase 5's live wizard run
  checks it: the menu must offer the plan sign-in.
* **The phase gate:**
  * `go test -race -count=1 ./...`: 0;
  * `go vet` for linux, darwin, windows: 0, 0, 0;
  * `go mod tidy -diff`: 0;
  * `go test -count=3 -shuffle=on ./internal/ui/`: 0;
  * `make verify`: 0 (0 issues; coverage 83.5%; no vulnerabilities).

