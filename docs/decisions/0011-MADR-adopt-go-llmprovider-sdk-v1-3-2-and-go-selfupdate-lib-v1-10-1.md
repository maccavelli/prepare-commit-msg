---
status: accepted
date: 2026-10-08
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its records 0026 and 0027, which made v1.3.0 to v1.3.2); go-selfupdate-lib (its records 0013, 0014 and 0015, which made v1.10.0 and v1.10.1)
informed: users of the installed hook, who receive the change through `update`
---
# Move to go-llmprovider-sdk v1.3.2 and go-selfupdate-lib v1.10.1, and publish through the newest workflow with a live publish on record

## Context and Problem Statement

`prepare-commit-msg` requires two libraries (`go.mod:6-7`):

* `github.com/maccavelli/go-llmprovider-sdk v1.2.1`, adopted by
  [0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md](0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md);
* `github.com/maccavelli/go-selfupdate-lib v1.9.0`, adopted by
  [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md).

Its release job (`.github/workflows/ci.yml:120`) publishes through
go-selfupdate-lib's `publish-selfupdate-release.yml`, pinned to
`6deaa524cfb2…` (`v1.5.0`). 0010-MADR D2 kept it there until "a publish
through it is on record": the module is proven by this repository's tests,
the workflow only by a published release.

Both libraries have released since:

| Library | From | To | Commit | Records |
| :--- | :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `v1.2.1` | `v1.3.2` | `1cdfc06` | 0026 (`v1.3.0`, `v1.3.1`, `v1.3.2`); 0027 (tests only, no release) |
| go-selfupdate-lib | `v1.9.0` | `v1.10.1` | `0589232` | 0013 completed; 0014 (`v1.10.0`); 0015 (`v1.10.1`, in progress towards `v1.11.0`) |

How should `prepare-commit-msg` move to both, and which commit should its
publish pin name?

### What was measured

Measured on 2026-10-07 and 2026-10-08 at `HEAD` `c16b568`, with Go 1.27.1
on macOS. The trials ran on scratch copies; nothing in this repository was
changed.

**The bump is two lines in `go.mod`.** `go get` of both versions and `go mod
tidy` change only the two libraries' lines in `go.mod`, and their lines in
`go.sum`. `x/mod`, `x/sys` and `x/term` do not move.

| Check, with both bumped | Result |
| :--- | :--- |
| `go build ./...` | 0 |
| `CGO_ENABLED=0 go vet ./...`, for darwin, linux and windows | 0 each |
| `go test -count=1 -race ./...` | 5 packages ok |
| `golangci-lint run -c .golangci.yml ./...` (`.tools/bin`) | `0 issues.` |
| `govulncheck ./...` | `No vulnerabilities found.` |
| `testdata/migration/`, against `v1.10.1`'s `selfupdate/cli/testdata/migration/` | 9 files, byte for byte |

`make verify` needs a git checkout, which the scratch copies were not. With
go-selfupdate-lib at `v1.10.0`, its other steps were run one by one:
`make mod-check`, `make coverage` (84.5%, minimum 80.0%) and `make
build-all` passed.

**Both APIs are compatible.** apidiff finds no incompatible change in
go-llmprovider-sdk `v1.2.1` → `v1.3.2` (8 additions), or in
go-selfupdate-lib `v1.9.0` → `v1.10.0` (10 additions, all in
`selfupdate/releasespec`). go-selfupdate-lib's release notes for `v1.10.1`
record "No API change; `make apicheck` reports `v1.10.1` compatible with
`v1.10.0`".

**What go-llmprovider-sdk `v1.3` changes for the hook:**

| Case | `v1.2.1` | `v1.3.2` | SDK record |
| :--- | :--- | :--- | :--- |
| A refused Gemini key (HTTP 400 `API_KEY_INVALID`) | `ErrInvalidRequest`: the hook warns and tries each fallback model with the same key, then fails with "all models for gemini failed" | `ErrAuthFailure`: the hook stops at the first model with "authentication failed for gemini" (`main.go:284`) | 0026 F6, D12 |
| A key with a control character, or another input no request can carry | built; each send fails and is retried `retry_count` times | refused when the provider is built: the hook warns "failed to init" and moves on, without retries | 0026 F11, F19 |
| An answer cut off while the model was still reasoning, with no text | empty text: "returned unusable message after cleaning" | `ErrIncomplete`: "model … failed: … incomplete"; the next model is tried either way | 0026 F7 |
| An OAuth sign-in's refresh (`internal/config/config.go:101`, `auth.FileTokenStore`) | an `O_EXCL` lock file, `<provider>.lock` | an OS lock on `<provider>.oslock`; a `v1.2.1` process sharing the store is not excluded by it | 0026 F13, D2 |
| `configure`, when a live listing recommends no model | could re-prompt without end, ignoring cancellation | honours the context, and stops | 0026 F9 |
| `configure`, a masked key whose terminal write fails | could return a partial key | returns the write error, never a partial key | 0026 F49, F50 |

The first row was measured by the SDK's own live check
(`TestLive_GeminiCommandTokenRerunsOnInvalidKey`, go-llmprovider-sdk
0026-PLAN D11): Gemini's Interactions API, which `Generate` calls, sends
this refusal inside a one-element JSON array, which `v1.2.1` read as text,
so its classification fell back to the status, 400. The hook's own tests
cover the stop rule only with hand-built errors (`main_test.go:398-423`,
`provider_test.go:76`), so none of them sends Gemini's real reply through
the SDK.

The retry cap does not change for the hook: `v1.2.1` and `v1.3.2` both
return at once a 429 whose `Retry-After` is over `MaxDelay`, 30 seconds by
default, and the hook then tries the next model.

**What go-selfupdate-lib `v1.10.1` changes for `update`.** `v1.10.0` changed
nothing this repository imports (`buildinfo`, `selfupdate`,
`selfupdate/cli`, `selfupdate/selfupdatetest`): only
`selfupdate/releasespec` and `selfupdate/service/launchd` changed. `v1.10.1`
carries 0015's fixes; these, from its release notes, are on the paths
`update` takes here:

| Finding | What a user may notice |
| :--- | :--- |
| C1 | under `--json`, the result object's `exit_code` matches the process's exit status, written after any late error |
| C6 | a run that fails in discovery names its product, version and mode in the `--json` result (`"product":"prepare-commit-msg"`, `"checked":true`), where it gave empty values |
| B1 | a backup the run could not restore is kept beside the binary as `.<base>.selfupdate-kept-<n>`, and no sweep removes it |
| B2 | a binary replaced by something else during the update fails with `ErrConcurrentUpdate` instead of being overwritten |
| A1 | a release without this platform's asset fails as `unsupported-platform`, and the result is cached |
| A2 | a token from the environment is re-read on each run |
| A4 | a secondary rate limit with no reset time backs off |
| C4 | `--check` refuses what an update would refuse |

The `--json` result's `schema_version` is still 2. Two of the library's
goldens changed (`failed.json.stdout`, `contradiction.json.stdout`, C6); the
nine migration fixtures this repository copied did not.

**Which publish workflows have published live.**

| Workflow commit | Release | A live publish on record |
| :--- | :--- | :--- |
| `6deaa52` | `v1.5.0` | this repository's `v1.6.0` and `v1.7.0` |
| `39b1294` | `v1.9.0` | go-selfupdate-lib `0013-PLAN-build-and-stage-release-workflow.md`, Phase B6 step 3 (2026-10-06): a throwaway public repository, `v0.0.1`–`v0.0.4`, raw and archive, immutable; `gh attestation verify` exited 0; `update` ran on macOS, Linux and Windows |
| `a0a26b6` | `v1.10.0` | go-selfupdate-lib `0014-PLAN-shared-installer-templates.md`, Phase I7 step 3 (2026-10-07): both workflows pinned to it, `v0.0.1`–`v0.0.3`, raw and archive, immutable |
| `0589232` | `v1.10.1` | none: CI's release rehearsals passed (run 37716131684), and they create no release; `0015-PLAN`'s Phase P8 records step 7, "the live installer rehearsal, was not asked for, and was not run" |

No other repository of this owner has published a real release through
`v1.9.0` or later.

**The publish workflow still accepts this repository's call.** Its five
inputs, required inputs, defaults and permissions are unchanged from
`6deaa52` to `v1.10.1`; only `platforms-json`'s description adds an optional
`format`. From `6deaa52` to `a0a26b6` the job gains a gh 2.81.0 check,
archive checks that a raw release skips, and a script that keeps a backport
from being made latest. From `a0a26b6` to `0589232` it gains:

* a job `concurrency` group, `go-selfupdate-lib-publish-${{
  github.repository }}`, not cancelled in progress, so two publishes in
  one repository do not race (0015 F9);
* on an immutability timeout, a message naming the repository setting it
  needs (0015 F4); this is on the failure path only.

0015 F8 also makes publishing accept a legacy latest release; this
repository's latest is `v1.7.0`, a `vX.Y.Z` tag.

**The README names the SDK version.** "Changes with go-llmprovider-sdk
v1.2.1" (`README.md:24`, `:401`) describes the last SDK move.

## Decision Drivers

* **Take both libraries' fixes:** the hook stops on a refused Gemini key;
  the OAuth store's lock no longer rests on a file a killed process leaves
  behind; `update` keeps a backup it could not restore, and refuses to
  overwrite a binary replaced under it.
* **A release always publishes.** A tag whose publish fails costs a version
  number and a follow-up release. 0010-MADR D2's rule, that the workflow is
  proven by a published release, stands.
* **No user loses a working path.** `update`, `update --check`, `version`,
  `configure` and the hook behave as the README says.
* **Each step lands green, and reverts alone.**
* **The hook's own tests see the SDK's real classification** on the path
  that changed, not only a hand-built error.

## Considered Options

* **A. Move both modules (SDK `v1.3.2`, go-selfupdate-lib `v1.10.1`), and
  the publish pin to `a0a26b6` (`v1.10.0`), the newest workflow with a live
  publish on record.**
* **B. Move both modules, and the publish pin to `0589232` (`v1.10.1`).**
* **C. Move both modules, and keep the publish pin at `6deaa52`
  (`v1.5.0`).**
* **D. Move the SDK only.**

## Decision Outcome

Chosen option: **"A"**, because it takes every fix both libraries have
released, and moves the pin as far as 0010-MADR D2's rule allows: to the
newest workflow commit that has published live. The module and the
workflow are one patch release apart, and the workflow's difference across
that patch is a concurrency group and a message on a failure path.

### D1. Both modules move, and no opt-in feature is adopted

* `go.mod` requires `go-llmprovider-sdk v1.3.2` and `go-selfupdate-lib
  v1.10.1`; nothing else in the module graph moves.
* No code change is needed. None is made for go-selfupdate-lib's
  `releasespec`, installers, services, archives or codesign.

### D2. The publish pin moves to `a0a26b6…` (`v1.10.0`)

* `ci.yml`'s `uses:` names
  `publish-selfupdate-release.yml@a0a26b6ecf66f51c19e9fea0f665c76ca5e99e4c
  # v1.10.0`, and its call is unchanged.
* The comment above it says that `go.mod` requires `v1.10.1` while the
  workflow is `v1.10.0`'s, the newest with a live publish on record
  (go-selfupdate-lib 0014-PLAN I7 step 3), and that it moves to `v1.10.1`'s
  commit, or a later one, once a publish through that is on record.
* This applies 0010-MADR D2's rule; the rule is unchanged.

### D3. A test sends Gemini's refusal through the SDK and the hook's loop

* A test serves Gemini's reply to a refused key, as the SDK's live check
  captured it (HTTP 400, the one-element array, `API_KEY_INVALID`), from a
  local server.
* It runs the hook's model loop with a primary and a fallback model, with
  `generateWithRetry` calling the real `generateText` on a Gemini provider
  aimed at that server.
* It requires one request, and the run to stop with "authentication failed
  for gemini". At SDK `v1.2.1` the same test sends one request per model
  and ends with "all models for gemini failed".

### D4. The README states what changed

* A section "Changes with go-llmprovider-sdk v1.3.2" follows the `v1.2.1`
  one, with the SDK table's rows in a user's words.
* "Self-Update" names the kept backup (B1) and the refusal of a binary
  replaced during the update (B2).
* The table of contents and any library version the README names are
  updated.

### D5. The release is the next minor, `v1.8.0`

* The owner tags it after CI is green on `main`.
* It is minor, not a patch, because a refused Gemini key now stops the run
  where it went on to the fallbacks, the OAuth store's lock changes, and
  `update --json`'s failed result changes its values.

### Consequences

* Good, because a refused Gemini key fails once, with "authentication
  failed", instead of once per configured model.
* Good, because `update` keeps a backup it could not restore, and does not
  overwrite a binary something else replaced.
* Good, because the OAuth store's lock is released by the operating system
  when a process dies.
* Good, because the pin moves from `v1.5.0` to a workflow with a recorded
  live publish, without making this repository's tag a first publish.
* Neutral, because the module and the workflow name different patch
  releases until a publish through `v1.10.1`'s is on record; the pin's
  comment says why.
* Bad, because the follow-up that moves the pin to `v1.10.1`'s commit
  depends on evidence from outside this repository, as 0010-MADR D2's did.
* Bad, because a `v1.7.0` process and a `v1.8.0` process sharing the OAuth
  store do not exclude each other while both run. `update` replaces the one
  binary, so this lasts only while an old hook is still running.
* Bad, because a script that read `update --json`'s failed result and
  relied on its empty `product` must change.

### Confirmation

* **Every phase passes** `make verify`, `make verify-staged`, `go test
  -race -count=1 ./...`, `CGO_ENABLED=0 go vet` for linux, darwin and
  windows, and `go mod tidy -diff`.
* **D3's test passes at SDK `v1.3.2`,** and fails on a scratch copy held at
  `v1.2.1` (two requests, "all models … failed"), and on a scratch copy
  whose stop rule is removed.
* **`TestMigrationByteForByte` and the `update` tests pass unchanged.**
* **CI passes on Linux, macOS and Windows, on `main` and on the tag,** and
  the tag's release is published through `a0a26b6`, with its attestations.
* **The installed hook,** once updated, reports both library versions in
  its build information, writes a commit message on a real key, and, run
  with a deliberately invalid Gemini key, stops at the first model.

## Pros and Cons of the Options

### A. Both modules, and the pin to `a0a26b6` (`v1.10.0`)

* Good, because every released fix arrives, and the pin moves on recorded
  live publishes.
* Good, because the workflow change across the patch it skips is small and
  stated.
* Bad, because the pin and the module differ by a patch, and the last move
  waits on outside evidence.

### B. Both modules, and the pin to `0589232` (`v1.10.1`)

* Good, because the module and the workflow agree, and the workflow gains
  F9's concurrency group.
* Good, because the workflow's difference from `a0a26b6`, which published
  live, is small: a concurrency group and a failure-path message.
* Bad, because no publish through it is on record, so it sets 0010-MADR
  D2's rule aside, and this repository's tag would be its first live
  publish.

### C. Both modules, the pin kept at `6deaa52`

* Good, because the release path is the one proven here twice.
* Bad, because the pin stays five releases behind, past the evidence its
  own record set for moving it.

### D. The SDK only

* Good, because it is the smallest change.
* Bad, because `update` keeps the defects `v1.10.1` fixes on its path, and
  the pin stays at `v1.5.0`.

## Owner questions

Each has a recommendation, written into the outcome above.

**Answered 2026-10-08:** "Questions follow recommendations". The outcome
above stands as written, and this record is `accepted`.

* **Q1 (D2).** Pin the publish workflow to `a0a26b6`, `v1.10.0`, the newest
  with a live publish (recommended); to `0589232`, `v1.10.1` (B); or keep
  `6deaa52` (C)?
* **Q2 (D3).** Add the Gemini refused-key test through the real SDK
  (recommended), or rely on the hand-built-error tests?
* **Q3 (D5).** Release `v1.8.0` (recommended), or `v1.7.1`?

## More Information

* go-llmprovider-sdk records, by full filename in that repository:
  * `docs/decisions/0026-MADR-remediate-v1-2-debugging-pass-findings.md`
    and `0026-PLAN-remediate-v1-2-debugging-pass-findings.md`, whose Phase 7
    holds the release notes for `v1.3.0` to `v1.3.2`;
  * `docs/decisions/0027-MADR-live-test-skips-and-gemini-429-path.md`,
    which measured Gemini's 429 on the Interactions API.
* go-selfupdate-lib records, by full filename in that repository:
  * `docs/decisions/0013-PLAN-build-and-stage-release-workflow.md`, Phase
    B6 step 3;
  * `docs/decisions/0014-MADR-shared-installer-templates.md` and
    `0014-PLAN-shared-installer-templates.md`, Phase I7 step 3;
  * `docs/decisions/0015-MADR-remediate-third-debugging-pass-findings.md`
    and `0015-PLAN-remediate-third-debugging-pass-findings.md`: Phase P8
    and the release notes for `v1.10.1`;
  * `docs/guides/migrating-from-mcplib-selfupdate.md`, "From v1.10.0 to
    v1.10.1".
* [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md):
  D2, the pin rule this record applies.
* **Not decided here:**
  * moving the pin to `v1.10.1`'s commit, D2's follow-up;
  * go-selfupdate-lib `v1.11.0`;
  * adopting `releasespec`, the build workflow or the installers;
  * gobble-cli's versions.
