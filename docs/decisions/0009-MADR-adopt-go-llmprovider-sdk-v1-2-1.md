---
status: accepted
date: 2026-10-06
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its records 0020, 0021 and 0023, which made v1.1.0 to v1.2.1)
informed: users of the installed hook, who receive the change through `update`
---
# Move to go-llmprovider-sdk v1.2.1, and keep refusals, entitlement errors and `configure --yes` working as they did

## Context and Problem Statement

`prepare-commit-msg` requires `github.com/maccavelli/go-llmprovider-sdk
v1.0.0` (`go.mod:6`), the release
[0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md](0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md)
adopted. The installed hook, release `v1.5.0`, embeds the same version: its
Go build information lists `go-llmprovider-sdk v1.0.0`.

The SDK has since released four versions:

| Tag | Commit | Date | What it carries |
| :--- | :--- | :--- | :--- |
| `v1.1.0` | `0fb3e4d` | 2026-10-04 | 0020's 59 debugging-pass fixes; Windows token-store permissions (0010 D10); Kilo device sessions |
| `v1.1.1` | `e9083c3` | 2026-10-05 | 0021's hardening and tuning, and its L3: an OpenCode 403 is "not permitted" |
| `v1.2.0` | `d4f9fa5` | 2026-10-05 | 0021's close-out (records only) |
| `v1.2.1` | `64b82d5` | 2026-10-06 | 0023: Hugging Face's `ToolChoiceNone` sends no tools |

go-llmprovider-sdk's
`docs/decisions/0020-PLAN-remediate-v1-debugging-pass-findings.md`,
"Rollout", expects this repository to pick these releases up "with its own
dependency bump". This record decides how.

What should `prepare-commit-msg` do to move to v1.2.1, and which of the
SDK's behaviour changes need an answer in this repository?

### What was measured

Measured on 2026-10-06 at `HEAD` `ff8a6a1`, with Go 1.27.1 on macOS. Each
measurement ran on a scratch clone or in a throwaway module. Nothing was
committed, and the repository was not changed.

**The bump itself is two lines.** `go get
github.com/maccavelli/go-llmprovider-sdk@v1.2.1` and `go mod tidy` change
only the SDK's line in `go.mod` and its two lines in `go.sum` (`h1:zwxJDG+…`).
The SDK's own `go.mod` and `go.sum` are identical at `v1.0.0` and `v1.2.1`,
so no other module changes version.

| Check | `v1.0.0` (baseline) | `v1.2.1` |
| :--- | :--- | :--- |
| `go build ./...`; `CGO_ENABLED=0 go vet` for linux, darwin, windows | 0 | 0 |
| `go test -race -count=1 ./...` | 5 packages ok | 5 packages ok |
| `go mod tidy -diff` | 0 | 0 |
| `python3 scripts/go-precheck.py` (go.mod and go.sum staged for `v1.2.1`) | exit 0 | exit 0: `v1.2.1 resolved from GitHub`, `0 issues.`, `No vulnerabilities found.` |
| `make verify` | exit 0, total coverage 84.2% | exit 0, total coverage 84.1% |

**The 0.1-point coverage difference is noise.** It is all in
`internal/ui/open.go`'s `openBrowserDefault`: 80.0% or 73.3%, depending on
whether the launcher's reaping goroutine runs before the test ends. Both
values were seen at each SDK version, across six runs of each.

**The API is compatible.** apidiff (the SDK's `scripts/check_api.py`
pin) from `v1.0.0` to `v1.2.1` reports no incompatible change, and 18
additions. The repository's 8 Go files that import the SDK use 48 of its
identifiers, none of them among the additions.

**The curated catalogs and key variables are unchanged.** `catalog.Static`
and `ProviderEnvVars` print the same for all ten providers at both
versions. So `config.DefaultModelForProvider` (`internal/config/config.go:110`),
which takes the first curated model, picks the same default.

**What the hook does differently.** Each case below drove a real SDK
provider against a local fake service, the way `generateText`
(`main.go:359`) does: one user message, through `WithRetry`, with 4
attempts. The 403 rows called `ClassifyHTTPError` directly. "Stops" means
`errors.Is(err, llmprovider.ErrAuthFailure)`: that is the test at
`main.go:282` which stops the run instead of trying the fallback models.

| Case | `v1.0.0` | `v1.2.1` | SDK record |
| :--- | :--- | :--- | :--- |
| OpenAI Responses answer whose only part is a `refusal` | 1 request; text `""`, no error | 1 request; **text is the refusal**, no error | 0021 W4 |
| Gemini HTTP 400 with reason `API_KEY_INVALID` | `invalid request`; does not stop | `authentication failed`; **stops** | 0020 F23 |
| OpenCode Go HTTP 403, untyped | `authentication failed`; stops | `not permitted`; **still stops** | 0021 L3 |
| Kilo HTTP 403 | `not permitted`; stops | `not permitted`; stops | — |
| OpenAI HTTP 409 | 1 request, `invalid request` | 4 requests, `provider unavailable` | 0021 T14 |
| A 200 whose body is cut short | 4 requests, `unexpected EOF` | 2 requests: retried once, then `provider unavailable: … read reply` | 0021 D1 |
| Hugging Face 200 with no choices | 4 requests | 1 request, `incomplete response` | 0020 F9 |
| Together 429 with `Retry-After: 5`, 1 s left before the deadline | waits out the deadline: `context deadline exceeded` | returns at once: `rate limited … (retry-after 5s)` | 0021 T6 |
| Claude 200 with `stop_reason: max_tokens` | partial text, no error | partial text, no error | — |

And through the hook's own code:

* **`configure --yes` on a Kilo listing of only `kilo-auto` tiers,** the
  SDK's own 0021 C13 fixture, served to `discoverModels`
  (`internal/ui/setup.go:147`):
  * at `v1.0.0` it returns the six curated Kilo models;
  * at `v1.2.1` it returns none, so `configure --yes` stops with `no live
    models listed for kilo; pass --model` (`setup.go:341-343`).

  0021 C13 keeps such a listing live, with an empty `Recommended`, so that
  the SDK's wizard opens its search. The hook's non-interactive path has no
  search.
* **The OAuth store on this host passes v1.2.1's new directory rule**
  (0021 T15): a real directory, owned by the user, mode `0700`. The rule
  refuses a symlinked or foreign-owned directory, and clears group and
  other write bits.

**Three of these change what a user gets:**

1. **A refusal can become the commit message.**
   * `cleanLLMOutput` (`main.go`) keeps a line such as "I'm sorry, but I
     can't help with that request.". That is 47 runes, over
     `minMessageRunes` (5), so `writeMessage` writes it.
   * With `git commit --no-edit`, it is committed.
   * At `v1.0.0` the refusal was dropped, so the empty text was "unusable"
     and the next model was tried.
   * This affects every route on the Responses wire: OpenAI keys, the
     ChatGPT session, Grok, and OpenCode's responses route. Claude's
     `refusal` stop reason and Gemini's `SAFETY` finish already returned
     whatever text came with them at `v1.0.0`. At `v1.2.1` they report
     `FinishContentFilter` too.
2. **An entitlement refusal still ends the run, and is now reported as two
   things at once.** `APIError.Unwrap` returns the error's kind and the
   status-only sentinel, so every 403 matches `ErrAuthFailure` at both
   versions (`api_error.go:133`, `statusSentinel`). So a model that the
   account may not use stops the run before its fallbacks are tried. At
   `v1.2.1` the message reads `authentication failed for opencode-go:
   llmprovider: not permitted: …`. 0021's L3 amendment says such a 403 "is
   an entitlement problem, not an auth failure".
3. **`configure --yes` can fail where it used to succeed,** in the C13 case
   above.

**The rest is better for the hook, with no change here:**

* a bad Gemini key stops at once, instead of failing every fallback model;
* a retry no longer sleeps past the hook's deadline (`main.go:241`,
  default 120 s);
* a reply cut short is retried once, not three times;
* a well-formed empty reply is not re-sent and re-billed;
* OAuth refreshes, token-store locking and redaction are hardened (0020
  F1, F12, F14 and F8; 0021 T2–T10, Z1–Z3).

The interactive `configure` changes as the SDK's wizard does:

* a saved session is offered before the method menu;
* saved fallbacks are preselected;
* a typed base URL is validated, and a plain `http` one to a remote host is
  confirmed;
* a pasted OpenAI credential is checked;
* masked entry handles paste, Escape and Ctrl-C on legacy consoles;
* a read at end of input no longer loops (0020 F6, F18, F19 and F50; 0021
  Z6, Z7 and Z10).

**The README is out of date, and was before the SDK moved.** Its
"Supported Providers & Models" table (`README.md:67`-`80`) disagrees with
`catalog.Static` at both versions:

* for `grok`, `kilo`, `huggingface`, `opencode-zen` and `opencode-go`, the
  "(recommended)" model and the list differ. The first curated model is the
  default `configure` picks;
* the `gemini` row lacks `gemini-2.5-flash-lite`.

The wizard steps (`README.md:191`) say "all 9 supported" backends; there
are 10. 0008's Phase 4 rewrote only the `together` row from the SDK.

**Seen, and out of this record's scope:**

* go-selfupdate-lib's newest release is `v1.9.0`; this repository requires
  `v1.5.0`.
* The Claude `max_tokens` row above. A cut answer is written as the message
  at both versions. `v1.2.1` now reports it as `FinishLength`, which a later
  record could act on.

## Decision Drivers

* **A commit message is never a model's refusal.** The hook may run under
  `git commit --no-edit`, where nobody reads the message before it is
  committed.
* **No user loses a working path.** A setup that ran `configure --yes`
  yesterday runs it today.
* **Take the SDK's fixes.** Most of v1.1.0 to v1.2.1 is fixes to retries,
  OAuth sessions, token storage and redaction that the hook uses as it is.
* **Each step lands green, and can be reverted alone.**
* **The README tells a user the truth about defaults,** and stays true when
  the SDK changes them.

## Considered Options

* **A. Move to v1.2.1, and adapt the three call sites the move affects.**
* **B. Move to v1.2.1 and change nothing else.**
* **C. Move to v1.1.0 only,** taking 0020 without 0021.
* **D. Stay on v1.0.0.**

## Decision Outcome

Chosen option: **"A. Move to v1.2.1, and adapt the three call sites"**,
because only A takes the SDK's fixes without letting a refusal become a
commit message or `configure --yes` fail where it worked. Each adaptation
uses identifiers that exist at `v1.0.0` (`FinishContentFilter`,
`ErrNotPermitted`, `Catalog.Recommended`). So the adaptations land first,
green on `v1.0.0`, and the bump lands alone after them.

### D1. A refused answer is unusable, and the next model is tried

* `generateText` calls `Generate` through `WithRetry`, instead of
  `GenerateText`.
* A response whose `FinishReason` is `llmprovider.FinishContentFilter`
  returns an error naming the refusal. The run loop then reports `model X
  failed: … refused`, and tries the next model, as it does for an empty
  answer.
* Otherwise it returns `OutputText()`, as now.

This restores `v1.0.0`'s result on the Responses wire, and extends it to
Claude's and Gemini's refusals.

### D2. "Not permitted" tries the next model; an authentication failure still stops

* The test at `main.go:282` becomes "matches `ErrAuthFailure` and does not
  match `ErrNotPermitted`".
* A 401, or Gemini's `API_KEY_INVALID`, still stops the run, as now.
* An entitlement refusal is reported as `model X failed: …`, and the
  fallbacks run. That covers an OpenCode or Kilo 403, OpenAI's
  `usage_not_included`, or an OpenCode `RegionError`, `DataPolicyError` or
  `FreeTierError`.

### D3. `configure --yes` falls back to the curated catalog when the live listing recommends nothing

When `catalog.List` succeeds with an empty `Recommended`, `discoverModels`
returns `catalog.Static` for the provider. That is `v1.0.0`'s result. The
interactive wizard keeps the SDK's new search.

### D4. The README says what changed, and its model table is checked

* **A new section,** "Changes with go-llmprovider-sdk v1.2.1", beside
  "Changes from the mcplib Releases". It lists what a user notices:
  * D1–D3;
  * a bad Gemini key stops at once;
  * the retry changes;
  * the wizard changes;
  * the Unix and Windows token-directory rules.
* **The provider table** lists `catalog.Static` for each provider, and
  marks its first model as "(recommended)". The wizard steps say 10
  backends.
* **A test** reads the table from `README.md` and requires each curated
  column to equal `catalog.Static` for its provider. A later SDK release
  that changes a curated list then fails CI until the README follows.

### D5. The release is the next minor, `v1.6.0`

* The owner tags it after CI is green on `main`.
* Users receive it through `prepare-commit-msg update`.
* It is minor, not a patch, because D1–D3 and the SDK's own changes alter
  what a user sees.

### Consequences

* Good, because the hook gets the SDK's fixes:
  * OAuth refreshes that do not revoke a token family;
  * one holder of the token-store lock;
  * redaction of this module's key formats;
  * retries that respect the deadline and do not re-bill a reply that
    arrived.
* Good, because a refusal is never written as a commit message, on any
  provider. That is stricter than `v1.0.0`, which wrote Claude's and
  Gemini's refusal text.
* Good, because one model's entitlement refusal no longer hides a
  fallback that would have worked.
* Good, because the README's defaults become true, and a test keeps them
  so.
* Neutral, because coverage stays about 84%, against the 80% floor.
* Bad, because D2 tries the fallbacks on an account-wide entitlement
  problem. Each costs one refused request, and the hook writes one warning
  per model.
* Bad, because D1 re-runs a request on the next model when a provider
  refuses. That is one more billed request, where `v1.0.0` already did the
  same for the Responses wire.
* Bad, because the README test makes an SDK curated-list change a README
  change too. That is its purpose.

### Confirmation

* **Every phase passes:**
  * `make verify`;
  * `python3 scripts/go-precheck.py` on the staged snapshot;
  * `go test -race -count=1 ./...`;
  * `CGO_ENABLED=0 go vet` for linux, darwin and windows;
  * `go mod tidy -diff`.
* **D1–D4 each have a test that is seen to fail on a planted break,** on a
  scratch copy. D1 also has a test on a real OpenAI provider against a local
  fake service that answers with a refusal. At `v1.2.1` that test fails
  without D1.
* `go list -m github.com/maccavelli/go-llmprovider-sdk` prints `v1.2.1`.
* **CI** passes on Linux, macOS and Windows, on `main` and on the tag.
* **The installed hook** reports `go-llmprovider-sdk v1.2.1` in its build
  information once `update` has run. It writes a commit message in a real
  repository, run by the owner.

## Pros and Cons of the Options

### A. Move to v1.2.1, and adapt the three call sites

* Good, because it takes every SDK fix, and the only regressions measured
  are answered.
* Good, because the adaptations land before the bump, each green, so the
  bump is a two-line commit that reverts alone.
* Bad, because it is more than a version bump: three small code changes,
  their tests, and documentation.

### B. Move to v1.2.1 and change nothing else

* Good, because it is two lines, and every gate passes.
* Bad, because a refusal from OpenAI, ChatGPT, Grok or OpenCode's
  responses route is written, and under `--no-edit` committed, as the
  message.
* Bad, because `configure --yes` fails on a listing that curates to
  nothing.

### C. Move to v1.1.0 only

* Good, because it avoids 0021's W4 and C13, the two changes behind the
  regressions.
* Bad, because it drops the rest of 0021:
  * the OpenCode 403 classification;
  * the deadline-aware retry;
  * the single retry after a reply;
  * the OAuth refresh and lock hardening;
  * the redaction fixes.
* Bad, because it pins the hook to a release the SDK has moved past. The
  next bump meets the same two changes.

### D. Stay on v1.0.0

* Good, because it is no work.
* Bad, because the hook keeps the defects 0020 lists as high severity:
  * F1, a failed save that revokes the user's ChatGPT or Grok sign-in;
  * F12, two holders of the refresh lock;
  * F8, keys that are not redacted in error text.

## Owner questions

Each has a recommendation, written into the outcome above.

**Answered 2026-10-06:** "Q1 - Q5 follow recomendations". The
outcome above stands as written, and this record is `accepted`.

* **Q1 (D1).** A refused answer tries the next model (recommended), or is
  written as the message?
* **Q2 (D2).** "Not permitted" tries the next model (recommended), or still
  stops the run?
* **Q3 (D3).** `configure --yes` falls back to the curated catalog
  (recommended), takes the first model of the live listing, or keeps the
  new error?
* **Q4 (D4).** The README section, table fix and table test (recommended);
  the section and fix without the test; or the section only?
* **Q5 (D5).** Release `v1.6.0` (recommended), or `v1.5.1`?

## More Information

* go-llmprovider-sdk records, by full filename in that repository:
  * `docs/decisions/0020-MADR-remediate-v1-debugging-pass-findings.md`:
    F1, F6, F8, F9, F12, F14, F18, F19, F23 and F50, and its amendment of
    2026-10-04 (a failure after a 200 is retried once);
  * `docs/decisions/0021-MADR-harden-and-tune-after-the-v1-1-review.md`:
    D1, T2–T10, T6, T7, T14, T15, W4, C13, Z1–Z3, Z6, Z7, Z10, and its
    amendment for L3;
  * `docs/decisions/0010-MADR-windows-stdio-oauth-tokenstore-ci.md`: D10,
    the Windows token-directory DACL;
  * `docs/decisions/0023-MADR-huggingface-tool-choice-none.md`: tools only.
    The hook sends none.
* [0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md](0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md):
  how this repository came to `v1.0.0`, and the supply-chain gate that
  checks the new pin.
* Not decided here:
  * go-selfupdate-lib `v1.5.0` to `v1.9.0`;
  * acting on `FinishLength`;
  * gobble-cli's pin. gobble-cli is another consumer of the SDK, at
    `v1.1.1`.

## Amendment 2026-10-08: the Gemini row did not hold until SDK `v1.3.2`

Made by
[0011-PLAN-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md](0011-PLAN-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md),
its deviation D1.

* **Corrects the table's row** "Gemini HTTP 400 with reason
  `API_KEY_INVALID`", which says `v1.2.1` classifies it `authentication
  failed` and the run stops. Through `Generate` it did not: Gemini's
  Interactions API sends that refusal in a one-element JSON array, which
  `v1.2.1` read as text, so the reply was `invalid request`, and the hook
  tried every fallback with the same key. `TestRunStopsOnGeminiRefusedKey`
  measured it at `v1.2.1`: "all models for gemini failed … invalid request
  … after 2 request(s)".
* **So** releases `v1.6.0` and `v1.7.0` did not stop on a rejected Gemini
  key, though this record and the README said they would. From `v1.8.0`,
  on SDK `v1.3.2`, they do (go-llmprovider-sdk 0026-MADR, its amendment
  "Gemini's Interactions API wraps its errors in an array").
