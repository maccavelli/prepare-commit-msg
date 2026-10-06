---
status: in-progress
date: 2026-10-06
associated-madr: "0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md"
---
# Implement the move to go-llmprovider-sdk v1.2.1

Associated MADR: [0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md](0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md)

## Goal

* `go.mod` requires `github.com/maccavelli/go-llmprovider-sdk v1.2.1`, and
  nothing else in the module graph changes version.
* **D1:** a refused answer is never written as the commit message; the next
  model is tried.
* **D2:** "not permitted" tries the next model, and an authentication
  failure still stops the run.
* **D3:** `configure --yes` falls back to the curated catalog when the
  live listing recommends nothing.
* **D4:** the README says what changed, its provider table matches
  `catalog.Static`, and a test keeps it so.
* **D5:** release `v1.6.0` carries all of it, and the installed hook is
  updated to it.

Done means every item under Verification holds.

## Scope

### In scope

| Phase | Who | What |
| :--- | :--- | :--- |
| 0 | owner, then agent | answer Q1–Q5, accept the MADR, index both records |
| 1 | agent | D1–D3, on SDK `v1.0.0`, with tests |
| 2 | agent | the bump to `v1.2.1`, and the wire-level refusal test |
| 3 | agent | D4: the README section, table and wizard steps, and the table test |
| 4 | owner, then agent | push, CI, tag `v1.6.0`, release checks, `update`, the live check |
| 5 | agent | close-out |

### Out of scope

* **Any change in go-llmprovider-sdk.** No record there needs an entry: its
  0020-PLAN only says that this repository bumps on its own.
* go-selfupdate-lib `v1.5.0` → `v1.9.0`, acting on `FinishLength`, and
  gobble-cli's pin. The MADR names each as not decided here.
* Historical records. 0008's `v1.0.0` text stays as written.
* Push and tags, which are the owner's.

## Rules for every phase

1. **Order.** Phases run in order. Each ends green, staged for the owner's
   commit. The agent does not commit to `main`. The owner commits with `git
   commit --no-edit`, which the global `prepare-commit-msg` hook names.
2. **Checks before each phase that changes code:**
   * `make verify` (lint, coverage at least 80.0%, govulncheck, workflow
     lint, build-all);
   * `make verify-staged`, which runs `scripts/go-precheck.py` on the staged
     snapshot, as the pre-commit hook does;
   * `go test -race -count=1 ./...`;
   * `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...`;
   * `go mod tidy -diff`;
   * `gofmt -l` on the changed Go files.
3. **Proofs on scratch copies,** never in the tree. Each new test is seen to
   fail on a planted break. The record names the plant and the failure.
4. **Session tooling is Python.** The repository's own scripts and `make`
   targets run as they always do.
5. **Identifiers.** Nothing committed carries a hostname, an account name or
   a real-machine path.

## Implementation Steps

### Phase 0: accept the records

1. The owner answers Q1–Q5, or accepts the recommendations. The answers go
   into the MADR, which becomes `accepted`, and this PLAN `in-progress`.
2. `docs/README.md` indexes both records, each with its status, in the
   existing list form.
3. The owner commits the records.

### Phase 1: the adaptations, on SDK `v1.0.0`

Each identifier used below exists at `v1.0.0`: `FinishContentFilter`
(`llmprovider/contract.go:137` there), `ErrNotPermitted`
(`api_error.go:36`), and `Catalog.Recommended`. So this phase is green
before the bump.

1. **D1, `main.go`.**
   * `generateText` calls `llmprovider.WithRetry(p, policy).Generate(ctx,
     req)`, with the same request.
   * A nil response with no error is an error, as the SDK's `GenerateText`
     treats it.
   * A response whose `FinishReason` is `llmprovider.FinishContentFilter`
     returns `"", fmt.Errorf("%w: %s", errRefused, <the refusal text, cut to
     one line>)`. `errRefused` is a new unexported sentinel.
   * Otherwise it returns `resp.OutputText()`.
   * The run loop needs no change: the error is not an authentication
     failure, so the loop logs `model X failed: …` and tries the next model.
   * The comment on `generateText` says why it does not use `GenerateText`.
2. **D2, `main.go:282`.** The stop condition becomes
   `errors.Is(err, llmprovider.ErrAuthFailure) && !errors.Is(err,
   llmprovider.ErrNotPermitted)`. Its comment cites 0009-MADR D2.
3. **D3, `internal/ui/setup.go`, `discoverModels`.** When `catalog.List`
   returns no error and `cat.Recommended` is empty, it returns
   `defaultModels(provider)`. The existing error at `setup.go:341-343`
   stays, for a provider with no curated catalog (Ollama).
4. **Tests,** in the existing files and style:
   * **`TestGenerateText` gains two subtests,** on `llmtest.NewFake`:
     * "refusal is an error": `Reply(&llmprovider.Response{Output:
       []llmprovider.Item{MessageItem{Role: RoleAssistant, Text: "I'm sorry,
       …"}}, FinishReason: FinishContentFilter})`. It wants an error matching
       `errRefused`, and one request;
     * "stop keeps its text", with `FinishStop`.
   * **A run-loop test, in `main_test.go`'s style,** through the
     `generateWithRetry` seam:
     * model 1 refuses, model 2 answers `feat: x`. The message written is
       `feat: x`, and stderr names model 1;
     * model 1 fails with `&llmprovider.APIError{Status: 403, Kind:
       llmprovider.ErrNotPermitted}`, model 2 answers. The message is
       written;
     * model 1 fails with `&llmprovider.APIError{Status: 401, Kind:
       llmprovider.ErrAuthFailure}`. The run stops with `authentication
       failed`, and model 2 is never asked.
   * **`TestDiscoverModels_EmptyRecommendationFallsBack`,** in
     `internal/ui`. `listingClient` is set to a client whose transport sends
     every request to an `httptest` server. That server answers with the
     SDK's C13 fixture: a Kilo listing of `kilo-auto/small` and
     `kilo-auto/balanced`, in `kiloRankEntry`'s shape. It wants
     `catalog.Static(kilo)`.
     * At `v1.0.0` it passes without D3. The SDK substituted the static
       catalog itself.
     * Phase 2 is where this test is seen to fail without D3.
5. **Proofs** (scratch copy, at `v1.0.0`), each of which must fail:
   * `generateText` without the `FinishContentFilter` check: "refusal is an
     error" and the loop's refusal case;
   * the stop condition without `!errors.Is(…, ErrNotPermitted)`: the 403
     case;
   * the stop condition removed entirely: the 401 case.
6. Run the checks, then stage the phase.

### Phase 2: the bump

1. `go get github.com/maccavelli/go-llmprovider-sdk@v1.2.1`, then `go mod
   tidy`.
   * The diff must be exactly the SDK's `go.mod` line and its two `go.sum`
     lines.
   * `go list -m all` must otherwise match the baseline:
     * go-selfupdate-lib `v1.5.0`;
     * `x/mod` `v0.40.0`, `x/sys` `v0.47.0`, `x/term` `v0.43.0` and
       `x/tools` `v0.49.0`.
   * Any other change is a deviation.
2. **`TestGenerateText_ResponsesRefusal`,** in `provider_test.go`. It builds
   the real OpenAI provider with `providers.New`, `WithBaseURL` an
   `httptest` server, and `WithoutModelMetadata()`. The server answers 200
   with a Responses body whose message's only part is a `refusal`. It wants
   `generateText` to return an error matching `errRefused`.
3. **Proofs** (scratch copy, at `v1.2.1`), each of which must fail:
   * `generateText` without D1's check:
     `TestGenerateText_ResponsesRefusal`, with the refusal returned as the
     text;
   * `discoverModels` without D3's fallback:
     `TestDiscoverModels_EmptyRecommendationFallsBack`, which gets an empty
     list.
4. **The release build's dependency.** `go version -m` on the `build`
   target's binary lists `go-llmprovider-sdk v1.2.1`.
5. Run the checks, then stage the phase. `make verify-staged` must print
   `go-llmprovider-sdk v1.2.1 resolved from GitHub`.

### Phase 3: documentation

1. **`README.md`, "Supported Providers & Models"** (`:65`).
   * Each provider's curated column is `catalog.Static` at `v1.2.1`, in its
     order, with the first model marked "(recommended)".
   * Ollama's row stays "dynamically discovered".
   * The other columns are unchanged.
2. **`README.md`, "Interactive Wizard"** (`:191`): "all 9 supported" becomes
   10, and a step names the saved-session and preselected-fallback
   behaviour.
3. **`README.md`, a new section,** "Changes with go-llmprovider-sdk v1.2.1",
   after "Changes from the mcplib Releases" and in the table of contents. It
   says, briefly:
   * D1, D2 and D3;
   * a rejected Gemini key stops at once;
   * retries: 409 from OpenAI and Claude, one retry after a cut reply, none
     past the deadline, no re-send of a whole empty reply;
   * the wizard changes the MADR lists;
   * the OAuth directory: on Unix it must be a real directory you own, and
     on Windows it is restricted to you.
4. **`TestReadmeCuratedModels`,** in package `main`. It reads `README.md`,
   finds the provider table's rows by their `` **`id`** `` cell, and
   extracts the backquoted ids from the curated column. For every provider
   with a non-empty `catalog.Static`, it wants that list exactly. It also
   wants a row for every such provider.
5. **Proofs** (scratch copy), each of which must fail:
   * a model removed from one row;
   * two models swapped;
   * a row deleted.
6. **Checks:**
   * the phase's code checks, for the test;
   * every relative link and anchor in `README.md` resolves, proven on a
     planted bad link;
   * `git diff --check`;
   * markdownlint, if the repository configures it. 0008 found that it does
     not.
7. Stage the phase.

### Phase 4: release and the live check (owner, then agent)

1. **The owner** commits and pushes. CI runs on Linux, macOS and Windows.
   The Windows job runs the token-store tests that skip on macOS.
2. **The owner** tags `v1.6.0` on that commit. The tag's run builds, passes
   `make verify-release`, and publishes.
3. **The agent** checks, on a download into scratch space, and records the
   output:
   * `shasum -a 256 -c SHA256SUMS` for the six binaries;
   * `version` prints `v1.6.0 (release) <revision>`;
   * `go version -m` on the binary lists `go-llmprovider-sdk v1.2.1`.
4. **The owner** runs `prepare-commit-msg update` on the installed hook.
   * The agent reads the installed binary's build information. It must
     list prepare-commit-msg `v1.6.0` and `go-llmprovider-sdk v1.2.1`.
   * It records them without the install path.
5. **The live check.** The owner makes one commit in any repository, with
   their configured provider, and reports whether a message was written.

### Phase 5: close-out

1. This PLAN is `complete`, the MADR `accepted`, and `docs/README.md` says
   so.
2. Anything found along the way and not done is named, with why.

## Verification

* **V1.** `go list -m github.com/maccavelli/go-llmprovider-sdk` prints
  `v1.2.1`, and the rest of the module graph is unchanged.
* **V2.** Every check in rule 2 passes at the end of Phases 1, 2 and 3.
* **V3.** Each D1–D4 test passes, and fails on its plant (Phases 1.5, 2.3,
  3.5).
* **V4.** At `v1.2.1`, a real OpenAI provider's refusal is an error, not a
  message (`TestGenerateText_ResponsesRefusal`).
* **V5.** CI is green on Linux, macOS and Windows, on `main` and on the
  tag. The tag's release passes `verify-release`.
* **V6.** The installed hook's build information lists
  `go-llmprovider-sdk v1.2.1`. The owner's live commit got a message.
* **V7.** Nothing committed carries a hostname, an account name or a
  real-machine path.

## Rollout and Rollback

* **Rollout.**
  * Phases 1–3 are local commits by the owner.
  * Phase 4 publishes `v1.6.0`, and each user receives it through `update`.
* **Rollback.**
  * Before the tag, Phase 3, 2 or 1 reverts alone, newest first.
    * Reverting Phase 2 alone returns to `v1.0.0`, with D1–D3 still green
      on it.
    * Phase 2's refusal test goes with that revert.
  * After the tag, a problem is fixed forward in a patch release.
  * A user can return to the previous release with `prepare-commit-msg
    update --version v1.5.0 --yes`.

## Execution Record

The trial behind the MADR's measurements ran on scratch clones and
throwaway modules; nothing was committed from it.

### Phase 0: accept the records (2026-10-06)

* **Approval.** The owner answered "Q1 - Q5 follow recomendations".
  Execution of Phase 1 waits for the owner's explicit go-ahead.
* The MADR is `accepted`, and this PLAN `in-progress`.
* `docs/README.md` indexes both, in its list form, with those statuses.
* Both records and the index are staged for the owner's commit (rule 1).
