---
status: accepted
date: 2026-10-19
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its records 0040 to 0046, which made v1.13.0 to v1.15.0); GitHub issues 1 and 2 of this repository; probes on a scratch copy of this repository moved to v1.15.0, 2026-10-18
informed: users of the installed hook, who receive the change through `update`
---
# Move to go-llmprovider-sdk v1.15.0, and offer the ChatGPT plan sign-in under this app's own name

## Context and Problem Statement

`prepare-commit-msg` requires `go-llmprovider-sdk v1.12.0` (`go.mod:6`)
**[read]**. The SDK has released `v1.13.0`, `v1.14.0` and `v1.15.0` since
**[read]**. Its issue on this repository, issue 2 ("Adopt
go-llmprovider-sdk v1.15.0: the ChatGPT plan sign-in in the wizard"), asks
for the move. It names one field that enables OpenAI's Sign in with ChatGPT
for agents in the wizard this repository already runs.

Issue 1 ("Adopt go-llmprovider-sdk v1.11.0 and keep the six-model menu") is
still open, though what it asks for shipped with
`0014-MADR-adopt-go-llmprovider-sdk-v1-11-0-and-go-selfupdate-lib-v1-15-2.md`
D2: `ModelLimit: catalog.MaxListed` is in the wizard's options
(`internal/ui/setup.go:270`) and `catalog.WithLimit(catalog.MaxListed)` in
`--yes`'s listing (`:154`) **[read]**.

What has to be decided is how far to take the new releases, what to do
about the tests the move breaks, and what the user is told.

Every claim below is marked:

* **[read]**: from this repository, or the SDK's source or records at the
  version named;
* **[probed]**: run on 2026-10-18 on a scratch copy of this repository at
  `a04b617`, moved to `v1.15.0` by the SDK's `tools/consumer.py
  --dependent` (which copies the repository without `.git` or any secret
  file, and never changes it), never in this tree;
* **[derived]**: reasoned from the two, and not run.

### The versions

| Library | Pinned | Newest | Source |
| :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `v1.12.0` | `v1.15.0` | its tags and module proxy **[read]** |
| go-selfupdate-lib | `v1.15.2` | `v1.15.2` | its tags **[read]** |

### What moving changes

**[probed]** The move itself is clean: `go get …@v1.15.0` and `go mod
tidy` change only the SDK's lines; `go build`, and `go vet` for darwin,
linux and windows, exit 0; golangci-lint reports 0 issues. Two tests fail,
both in `internal/ui`, under `go test` and under `-race`:
*(Amendment 1, 2026-10-19: three tests fail; the third,
`TestReadmeCuratedModels`, is in the root package.)*

1. **`TestRunSetupInteractive_ShortList`** (`menu_limit_test.go:93`): "the
   wizard did not list through the injected client".
   * It passes run alone, and fails after
     `TestDiscoverModels_ShortList` **[probed]**. That test lists the same
     provider, endpoint and key just before it.
   * `v1.14.0` caches a listing for ten minutes, process-wide, keyed by
     provider, endpoint and credential (the SDK's
     `0041-MADR-decide-per-model-behaviour-from-metadata.md` D6;
     `catalog.listingKeyFor`) **[read]**. So the wizard is served the
     cached listing, and sends no request through the test's client.
   * The product is right; the test assumes no cache.
2. **`TestDiscoverModels_EmptyRecommendationFallsBack`**
   (`setup_test.go:187`): a Kilo listing of only `kilo-auto/small` and
   `kilo-auto/balanced` now gives `["kilo-auto/balanced" "kilo-auto/small"]`,
   where it gave nothing and the test wants the curated catalog.
   * 0041-MADR A4, decision 2: "The profile's exclusions … are admitted to
     `Recommended` past the band"; the band itself is unchanged **[read]**.
     The `kilo-auto` tiers are one of those exclusions
     (`utilityExcluded`, `catalog/model_ranking.go:306`) **[read]**.
   * The fallback itself still works. A Kilo listing whose only model
     writes embeddings gives the curated catalog, through
     `discoverModels`'s empty-list path **[probed]**.
     *(Amendment 2, 2026-10-19: the curated catalog, yes, but from the
     SDK, not through that path.)*

**What the user would see,** from the SDK's release notes (its 0041-PLAN,
0043-PLAN and 0044-PLAN) **[read]**:

* **The hook's request.** It sets no output limit and no reasoning
  (`main.go:394-396`) **[read]**. Under `v1.14.0`, its output limit is the
  configured default, cut to the model's own where that is smaller, and
  never raised.
* **Listings:**
  * a listing is cached for ten minutes;
  * after a transient listing failure, the last good one (within 24
    hours) is served;
  * Claude's listing is live;
  * Ollama's listing leaves out models that cannot write text.
* **Ollama.** An install whose `/api/tags` reports capabilities (Ollama
  `v0.30.0` and later) no longer offers an embedding model to `--yes`
  (`["llama3.3:latest"]` from an install of `llama3.3` and
  `nomic-embed-text`) **[probed]**. That is the SDK fix that
  0014-PLAN's Deviation 3 waited for. The README's note "With Ollama, name
  the models" (`README.md:578`) still holds only for older Ollama versions
  **[derived]**.
* **Kilo.** A Kilo listing with no model in the profile's band now offers
  the `kilo-auto` tiers to `--yes` (item 2 above). A real Kilo listing
  carries hundreds of models, so the band fills first **[derived]**.
* **OpenAI's sign-in.** The wizard offers "Use your ChatGPT plan with this
  app" (`llmprovider.AuthChatGPTPlan`) when `wizard.Options.AgentName` and
  a `TokenStore` are set (the SDK's `wizard/auth.go`) **[read]**. This
  repository passes a `TokenStore` (`internal/ui/setup.go:275`), and no
  `AgentName` **[read]**.
* **Retries.** `llmprovider.WithRetryNotify` calls a function before each
  wait that `WithRetry` takes, with the attempt, the wait and the error
  (the SDK's 0044-MADR) **[read]**. The README says "a retry is no longer
  logged" (`README.md:461-463`), a loss from leaving mcplib **[read]**.

### What the ChatGPT plan sign-in asks of this app

OpenAI's terms for plan usage ask things that the SDK cannot do for the
app (the SDK's `auth` package doc, 0043-MADR D8) **[read]**:

* requests from the user's own runtime, for that user's activity;
* express consent before background use;
* the plan used only for the app the user connected;
* no pooling, sharing, transfer or resale;
* no charge.

This hook runs on the user's machine, once per `git commit` the user makes,
to write that commit's message. It runs nothing in the background, and
charges nothing **[derived]**. Whether a commit hook counts as the user's
own activity is the owner's call (Q1).

## Decision Drivers

* **Every SDK API that removes hand work is used.** That is 0014-MADR's
  rule, which this record keeps.
* **No silent change for a user who changes nothing.** What changes, the
  README says.
* **Tests assert the product, not the SDK's caching.**
* **The app signs in under its own name.** `AppTitle`, which D5 of 0014
  already sends (`main.go:35`).

## Considered Options

* **A. Move to `v1.15.0`;** offer the plan sign-in; show each retry; fix
  the two tests to the SDK's behaviour.
* **B. Move to `v1.15.0` and fix the tests only;** no new feature.
* **C. As A, without the retry line.**
* **D. Amend 0014 rather than write a new record,** as 0014's Amendment 1
  did for `v1.12.0`.

## Decision Outcome

Chosen option: "A", because it closes both issues and takes the two APIs
the release adds for an app like this one: the sign-in under its own name,
and the retry it can show. A new record, not an amendment (D), because the
sign-in is a decision of its own, and 0014's PLAN is complete.

### D1. go-llmprovider-sdk moves to `v1.15.0`

* `go.mod` requires `v1.15.0`; `go mod tidy`. go-selfupdate-lib stays at
  `v1.15.2`, its newest. No other module changes.

### D2. The wizard offers the ChatGPT plan sign-in

* `ui.SetupOptions` gains `AgentName`. `main.go` sets it to `AppTitle`,
  as it builds `ClientOptions` today, and `runSetupInteractive` passes it
  to `wizard.Options.AgentName`.
* The plan session is saved in the OAuth store, as every sign-in is, and
  the hook uses it as it uses any stored session (`main.go:325-332`). The
  `openai` provider sends it to the platform under OpenAI's rules.
* Before the first plan sign-in, the wizard shows OpenAI's notice, "You're
  using your ChatGPT plan", and the usage page.
* `configure --yes` stays API-key-only.

### D3. A retry is shown again

* `generateText` uses `WithRetryNotify` with a function that writes one
  line to stderr before each wait:
  `prepare-commit-msg: <provider>: retrying in <wait> (attempt <n> of <max>): <error kind>`.
  The kind is the error's class (rate limited, unavailable, cannot reach),
  never the provider's message.
* Retries are decided as before. Only the line is new.

### D4. The tests follow the SDK

* **The listing cache.** Each test that lists through an injected client
  lists with a key of its own. Two tests then never share a cached listing.
* **The fallback.** `TestDiscoverModels_EmptyRecommendationFallsBack`
  lists ~~a Kilo model that writes embeddings, which still gives an empty
  list~~ a Kilo model past its expiration date, which the SDK lists and
  recommends none of, and so the curated catalog **[probed]**
  *(Amendment 2, 2026-10-19)*. A new test pins the
  `kilo-auto` result of 0041-MADR A4.
* **The curated rows** *(Amendment 1, 2026-10-19).* The README's
  provider rows for `kilo`, `huggingface`, `together`, `opencode-zen`
  and `opencode-go`
  take `v1.15.0`'s curated lists, which `TestReadmeCuratedModels` pins.

### D5. The README states what changed

* **The sign-in:** the plan method in the wizard, its notice, and the
  terms the MADR lists, and how the app keeps each, with D2's reasoning for this hook.
* **The retry line.**
* **"With Ollama, name the models"** is narrowed to Ollama before
  `v0.30.0`.
* **A section, "Changes with go-llmprovider-sdk v1.15.0":**
  * the facts-based output limit;
  * the listing cache, and the stale listing;
  * Claude's live listing, and Ollama's filter;
  * the `kilo-auto` tiers;
  * the curated lists of five providers *(Amendment 1, 2026-10-19)*;
  * the sign-in and the retry line;
  * the downgrade cost: a plan session saved by this version, read by an
    older one, fails to refresh until the user signs in again (the SDK's
    0043-MADR D3) **[read]**.

### D6. The issues close with the release

* Issue 1 closes as done by 0014-MADR D2. Issue 2 closes with this release.
  Each gets a comment, and is closed on the owner's ask.

### D7. The release is `v1.10.0`

* Minor, because the wizard offers a new sign-in and the hook prints a
  line it did not. The owner tags it after CI is green on `main`.

### Consequences

* Good, because a ChatGPT plan subscriber can use this hook under its own
  name, without the Codex CLI's client id.
* Good, because a user sees why a commit waits.
* Good, because an Ollama user no longer has to name the models, on a
  current Ollama.
* Neutral, because the tests change to match the SDK, not the product.
* Bad, because a plan session record breaks under a downgrade, which the
  README states.

### Confirmation

* The ~~two~~ three failing tests pass *(Amendment 1, 2026-10-19)*, and the new ones (the sign-in option, the
  retry line, the `kilo-auto` result) are each seen to fail on a planted
  break first.
* `go test -race ./...`, `scripts/go-precheck.py` and CI pass on all
  three platforms.
* A live check with an OpenAI key, run by the owner: one commit message,
  and a forced retry's line on stderr.
* A live plan sign-in through the wizard, by the owner.

## Pros and Cons of the Options

### A. Move, sign-in, retry line, tests

* Good, because both issues close, and both new APIs are used.
* Bad, because the change is larger than a pin bump.

### B. Move and fix the tests only

* Good, because it is the smallest change.
* Bad, because issue 2 stays open, and a retry stays silent.

### C. As A, without the retry line

* Good, because the hook's stderr is unchanged.
* Bad, because a commit can wait a minute with no word, the gap the SDK's
  issue 3 named.

### D. Amend 0014

* Good, because it is the pattern of 0014's Amendment 1.
* Bad, because 0014's PLAN is complete, and a sign-in method is a decision
  of its own.

## Owner questions

* **Q1.** Does a commit hook that runs on each `git commit` meet the
  terms' "user's own activity", and need no separate consent? D2 assumes
  yes.
  **Answered 2026-10-18: yes.** D2 stands.
* **Q2.** Should the retry line go to stderr on every retry (D3), or only
  when stderr is a terminal?
  **Answered 2026-10-18: on every retry.** D3 stands.
* **Q3.** Is the release `v1.10.0` (D7)?
  **Answered 2026-10-18: `v1.10.0`.** D7 stands.

## More Information

* The SDK's records, cited by repository and filename:
  go-llmprovider-sdk `0041-MADR-decide-per-model-behaviour-from-metadata.md`
  (D6, A4), `0042-MADR-ollama-models-that-write-text.md`,
  `0043-MADR-openai-chatgpt-plan-sign-in.md` (D3, D7, D8) and
  `0044-MADR-observe-retries.md`.
* The issues: [issue 1](https://github.com/maccavelli/prepare-commit-msg/issues/1)
  and [issue 2](https://github.com/maccavelli/prepare-commit-msg/issues/2).
* The implementation:
  [0015-PLAN-adopt-go-llmprovider-sdk-v1-15-0.md](0015-PLAN-adopt-go-llmprovider-sdk-v1-15-0.md).

### Amendment 1 (2026-10-19): a third test fails, on the README's curated rows

"What moving changes" said two tests fail. Three do. 0015-PLAN's Phase 1,
on a scratch copy at `v1.15.0` with no test changed, found the third:

* `TestReadmeCuratedModels` (`readme_test.go:51`) compares each provider's
  row in the README with `catalog.Static`. `v1.15.0` changed five lists
  **[probed]**:
  * `kilo`: `meta/muse-spark-1.2` and `openai/gpt-6.1-sol` leave;
    `minimax/minimax-m3` and `nvidia/nemotron-3-ultra-550b-a55b` join; and
    `google/gemini-3.6-flash` now comes before `google/gemini-3.8-flash`;
  * `huggingface`: `thinkingmachines/Inkling-Small`,
    `stepfun-ai/Step-3.7-Flash` and `stepfun-ai/Step-3.5-Flash` leave;
    `tencent/Hy3`, `google/gemma-4-26B-A4B-it` and `google/gemma-4-31B-it`
    join;
  * `together`: the same six, with `openai/gpt-oss-120b` now before
    `Qwen/Qwen3.6-Plus`;
  * `opencode-zen`: `gemini-3.5-flash-lite` leaves; `gpt-6-luna` joins;
  * `opencode-go`: the same six, with `gpt-6-luna` and `space-bunny` now
    third and fourth.
* At `1dc2434`, on `v1.12.0`, the test passes **[probed]**. The move
  causes it.
* The MADR's own probe copy fails it too **[probed]**: the draft
  miscounted, and read only `internal/ui`'s failures.

The owner chose (2026-10-19): the five rows change in Phase 1, with the
move; D4 and D5 say so. No other decision changes.

### Amendment 2 (2026-10-19): the embedding listing does not reach the fallback

"What moving changes" and D4 said a Kilo listing whose only model writes
embeddings gives an empty `Recommended`, and so `discoverModels`'s
curated catalog. The catalog is right; the route is not. 0015-PLAN's
Phase 1, on a scratch copy with that fallback removed **[probed]**:

| Kilo listing of one… | `Live` | `Recommended` | `discoverModels`, no fallback |
| :--- | :--- | :--- | :--- |
| model that writes embeddings | false | the SDK's curated list | the curated list |
| text model without tools | false | the SDK's curated list | the curated list |
| tool-using text model, `expiration_date` 2026-01-01 | true | `[]` | `[]` |
| tool-using text model, an id with `-contributor` | true | `[]` | `[]` |

* A model that writes no text, or uses no tools, is not usable. With none
  usable, the SDK gives its static catalog itself (`listingResult.catalog`,
  `catalog/discovery.go:474-482`) **[read]**.
* A usable model with a caveat, expiring (`caveatsOf`,
  `catalog/menu.go:66`) or unavailable (`excludedAtRequest`,
  `catalog/model_ranking.go:182`), stays out of `Recommended`, so a live
  listing of only such models recommends nothing **[read]**. That is the
  case `discoverModels`'s fallback serves.

The owner chose (2026-10-19): the fallback test lists one model past its
expiration date. D4 says so. No other decision changes.
