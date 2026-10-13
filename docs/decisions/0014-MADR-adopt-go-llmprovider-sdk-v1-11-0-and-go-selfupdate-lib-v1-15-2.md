---
status: accepted
date: 2026-10-13
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its records 0028 to 0038, which made v1.4.0 to v1.11.0); go-selfupdate-lib (its records 0021 to 0024, which made v1.14.0 to v1.15.2); probes on scratch clones of this repository, 2026-10-13
informed: users of the installed hook, who receive the change through `update`
---
# Move to go-llmprovider-sdk v1.11.0 and go-selfupdate-lib v1.15.2, and take from their APIs what this repository still does by hand

## Context and Problem Statement

`prepare-commit-msg` requires `go-llmprovider-sdk v1.3.2` and
`go-selfupdate-lib v1.13.0` (`go.mod:6-7`), and its release jobs call the
library's build and publish workflows at `5e199c8…` (`v1.13.0`,
`.github/workflows/ci.yml:99`, `:114`). Both libraries have released since.

The owner asked, on 2026-10-13, to pin and move to the latest release of
both, and to lean on the libraries' APIs as far as they go, so that no
hand-rolled code stands in for something a library already provides.

Two parts of the original request are settled and not decided here:

* **acp-go-sdk2:** nothing in this repository, and nothing in
  go-llmprovider-sdk `v1.11.0`, imports it. The owner set it aside.
* **mcplib:** `go mod graph` names no mcplib module, and
  `scripts/go-precheck.py:32` refuses one. The owner chose to change nothing
  for it. The README section and the code comments that cite mcplib's
  records stay, and so do records 0001 to 0008.

Every claim below is marked:

* **[read]**: from this repository, a library's source or record at the
  version named;
* **[probed]**: run on 2026-10-13 on scratch clones of this repository at
  `f1c3523`, never in its tree;
* **[derived]**: reasoned from the two, and not run.

### The versions

| Library | From | To | Tag commit | Records |
| :--- | :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `v1.3.2` | `v1.11.0` | (module proxy) | 0028 to 0038 |
| go-selfupdate-lib | `v1.13.0` | `v1.15.2` | `436a7f5976a6815491c7384527fef90869574a1b` | 0021 to 0024 |

The selfupdate tag commit was read with `git ls-remote` on
`refs/tags/v1.15.2^{}` **[read]**.

### What moving changes

**The module graph [probed].** `go get` of both versions and `go mod tidy`
change only the two libraries' lines in `go.mod` and `go.sum`. `x/mod`,
`x/sys` and `x/term` do not move; both libraries already need Go 1.27.2,
which `go.mod:3` names.

**Neither API breaks [read], [probed].** Nothing this repository imports
was removed, renamed or deprecated in either library. `go build ./...` and
`go vet ./...` exit 0 with both moved.

**go-selfupdate-lib `v1.15.2` changes nothing on this repository's paths
[read], [probed].**

* The only new exported symbol is `selfupdate.ErrBatchFile`, for Windows
  version probes, which this repository does not run.
* `build-selfupdate-release.yml` and `publish-selfupdate-release.yml` are
  byte-identical to `v1.13.0`'s (`cmp`), as are the publish scripts. The
  publish path is the one `v1.11.0` already published live through, so
  0010-MADR D2's rule, as 0012-MADR amended it, allows the pin to move.
* The nine `testdata/migration/` fixtures are byte-identical to `v1.15.2`'s
  `selfupdate/cli/testdata/migration/`.
* With go-selfupdate-lib alone moved, `go test -race ./...` passes in all
  five packages.

**go-llmprovider-sdk `v1.11.0` fails three tests [probed].**

| Test | Failure | Cause |
| :--- | :--- | :--- |
| `TestReadmeCuratedModels` (`readme_test.go`) | the README's provider table (`README.md:84-93`) lists the old curated models for 8 providers | the test already derives its expectation from `catalog.Static`; the README is stale (SDK 0033 D4, `v1.7.1`; `v1.10.2`) |
| `TestConfig_TemplateDefaults` (`internal/config/config_test.go:257-286`) | hard-codes `gemini-3.7-flash`, `gpt-4.1-mini`, `claude-haiku-4-5` | the defaults are now `gemini-3.5-flash-lite`, `gpt-6-luna`, `claude-haiku-5-5` |
| `TestRunSetupInteractive_Success` (`internal/ui/setup_test.go:55-93`) | types `7` for "Other", assuming six Gemini models | Gemini's static catalog is now three, so "Other" is `4` |

Each is catalog drift; none is a code defect.

**v1.11.0 lengthens configure's model menu [read].** SDK
0038-MADR makes `catalog.List` and `wizard.ConfigureLLM` recommend the
whole curated catalog unless the caller asks for the short list, with
`wizard.Options.ModelLimit` (`wizard/configure.go:93`) or
`catalog.WithLimit` (`llmprovider/catalog/config.go:94`). The six it always
showed are `catalog.MaxListed` (`models_catalog.go:29`), and are a prefix of
the full list. 0038-MADR's drivers name `prepare-commit-msg` among the
consumers that want only the top six. This repository passes neither
(`internal/ui/setup.go:154`, `:231-243`).

**The SDK also changed what a user sees [read].** From SDK 0028
(`v1.4.0`): a 403 is `ErrNotPermitted` and is not renewed; Grok's refused
key is `ErrAuthFailure`; a request that failed after it was written is sent
at most twice; a 429 waits for the service's reset headers; the wizard asks
again for a refused typed key, up to three times; listings are cached for
10 minutes. SDK 0031 streams natively; `WithRetry`'s `Generate` path, which
the hook uses, is unchanged.

### What this repository still does by hand

Measured against both libraries' `v1.11.0` / `v1.15.2` exported APIs.

**Defects the APIs already fix [probed].**

* **Typed operational values are lost.** `runSetupInteractive` reads the
  wizard through `wizard.TextPrompter{In: in}` (`setup.go:231`) and the
  operational prompts through a second `bufio.NewReader(in)` (`setup.go:190`,
  `:86-142`). The prompter's own buffered reader drains the input, so the
  second reader sees end of input. A probe typed `99`, `12345`, `7` and `9`
  for the four operational values through a pipe at `f1c3523`; the saved
  config kept `120`, `32000`, `3` and `3`. No existing test types a
  non-default value. One prompter, with its `Input(prompt, def)`
  (`wizard/text_prompter.go:290`), reads both. On a terminal, where each read
  returns one line, the defect probably does not show **[derived]**.
* **An Ollama configuration cannot run the hook.** The README says Ollama
  needs no key (`README.md:89`, `:114`), and `configure` saves it without one
  (`setup.go:279`, which already reads `Descriptor.RequiresAPIKey`). But
  `config.ValidateActive` (`internal/config/config.go:300-315`) requires a
  key for every provider. A probe built the hook's provider for
  `{"active_provider":"ollama"}` at `f1c3523`: `no API key for provider
  "ollama"`. The registry's `Descriptor.RequiresAPIKey`
  (`llmprovider/registry.go:28`) is the rule.

**Wizard answers this repository drops [read].**

* `wizard.Result.BaseURL` is asked for Kilo, Together, Hugging Face,
  Ollama, OpenCode Zen and OpenCode Go (each descriptor's
  `SupportsBaseURL: true`; `wizard/configure.go:259-281`), validated, and
  returned. `config.ProviderConfig` has no field for it and
  `llmprovider.WithBaseURL` is never passed, so a non-default Ollama host or
  gateway is accepted at the prompt and then ignored. The README promises
  "Endpoint resolution (for Ollama or custom gateway endpoints)"
  (`README.md:244`).

**Hand-rolled code the APIs replace [read].**

| Repository | What it does | API |
| :--- | :--- | :--- |
| `setup.go:27-28`, `DiscoveryTimeout = 45s`, passed as `DiscoverLimit` and to `context.WithTimeout` | bounds listing | every listing is capped at 10 s (`catalog/discovery.go:57`, `:115`); `Options.DiscoverLimit`'s own comment says a larger value has no effect; zero uses the wizard's default, 10 s (`wizard/configure.go:22`) |
| `setup.go:71-77`, `:86-142`, `readLine`, `promptInt`, `promptOperational` on a private `bufio.Reader` | reads four integers | `wizard.TextPrompter.Input`, on the prompter the wizard used |
| `setup.go:147-163`, `discoverModels` ignores `cat.Err` | `--yes` saves a refused key silently with the static list | `catalog.Catalog.Err` (`discovery.go:77`) with the wizard's rule, `errors.Is(err, ErrAuthFailure) && !errors.Is(err, ErrNotPermitted)` (`wizard/configure.go:310`, unexported, so the two-clause check is repeated) |
| `config.go:27-32`, `SupportedProviders`, 4 of the registry's 10 providers; `main.go:53`, `:116`, help and flag text naming 3 or 4 | the provider list | `providers.Default().Descriptors()` (`llmprovider/registry.go:95`) |
| `config.go:300-315`, `ValidateActive` requires a key | the key rule | `Descriptor.RequiresAPIKey` |
| no client info on any provider or listing; the User-Agent leads with `go-llmprovider-sdk` | names the application | `llmprovider.WithClientInfo(name, version)` (`llmprovider/identification.go:15`) |
| `main.go:328`, `session.Store = store` | sets the session's store | `auth.FileTokenStore.Load` already sets it (`auth/tokenstore_file.go:86`) |
| `main_oauth_test.go:218-228`, `analyzerTestProvider` | a provider that returns one text | `llmtest.NewFake(id, caps).ReplyText(text)` (`llmtest/fake.go:26`, `:51`) |
| `config_test.go:257-286`, `setup_test.go:75`, hard-coded model IDs and menu positions | test expectations | `catalog.Static`, `catalog.MaxListed`, `config.MaxFallbacks` |
| `update.go:57-74`, the updater has no `Unpacker` | follows the release's packaging | `releasespec.Spec.Unpacker()` (`selfupdate/releasespec/spec.go:322`); nil for `"packaging":"binary"`, so nothing changes today |

**Hand-rolled code with no library equivalent [read].** These stay:

* `generateText` (`main.go:377-393`): `llmprovider.GenerateText`
  (`convenience.go:10`) returns text only and cannot tell a
  `FinishContentFilter` refusal from a message. `WithRetry` is already used.
* `config.LookupEnv`'s `CLAUDE_API_KEY` fallback: the SDK knows only
  `ANTHROPIC_API_KEY`.
* `internal/fsutil/atomic.go`, `internal/config/save_atomic.go`: both
  libraries' atomic writers are `internal/`.
* `internal/ui/open.go`: neither library opens a browser; the SDK takes an
  `OpenURL` hook, which this file fills.
* `internal/git`: go-selfupdate-lib's windowless runner is `internal/`, and
  its exported `service.ExecRunner` requires an absolute path, caps output
  at 1 MiB and refuses `git.cmd` shims.
* `config.DefaultModelForProvider` and `DefaultFallbacksForProvider`: they
  already read `catalog.Static`; the SDK has no "default fallbacks" API.
* `update.go:35-36`, `updateTimeout`: equals `cli.DefaultTimeout` in value,
  but bounds one HTTP request where the library's bounds the whole run.

## Decision Drivers

* **Lean on the libraries.** Where a library exports what this repository
  does by hand, call it; where it does not, say so.
* **No user loses a working path, and the broken ones work.** `configure`,
  `configure --yes`, the hook, `update`, `version` and `identity` behave as
  the README says, including for Ollama.
* **Tests follow the libraries' data,** so the next catalog refresh fails
  the README check, not the unit tests.
* **A release always publishes.** The workflow pin moves only on recorded
  evidence (0010-MADR D2, as amended by 0012-MADR).
* **Each phase lands green, and reverts alone.**

## Considered Options

* **A. Move both libraries, keep the six-model menu, and replace every
  hand-rolled piece the APIs cover (D1 to D10).**
* **B. Move both libraries and fix only the three tests and the README.**
* **C. As A, but take the SDK's full curated menu.**
* **D. Move go-selfupdate-lib only.**

## Decision Outcome

Chosen option: **"A"**, because it is the only option that takes both
releases and does what the owner asked: the two defects and the dropped
base URL are fixed by calling APIs the repository already imports, and the
rest of the hand-rolled code either goes or is named as having no
equivalent.

### D1. Both modules move, and the workflows follow

* `go.mod` requires `go-llmprovider-sdk v1.11.0` and `go-selfupdate-lib
  v1.15.2`; nothing else in the graph moves. *(Amended 2026-10-13: the SDK
  moves on to `v1.12.0`; see Amendment 1 under More Information.)*
* `ci.yml`'s two `uses:` name `@436a7f5976a6815491c7384527fef90869574a1b #
  v1.15.2`. The comment above them names `v1.15.2` and keeps its reasoning:
  the publish path is unchanged from `v1.11.0`'s, which has published live.
* `scripts/go-precheck.py` is unchanged; it reads the versions from
  `go.mod`.

### D2. The menu stays at six

* `wizard.Options.ModelLimit` is `catalog.MaxListed`.
* `--yes`'s listing passes `catalog.WithLimit(catalog.MaxListed)`.
* `DiscoveryTimeout` goes. The wizard takes its default limit, and
  `discoverModels` takes the listing's own 10 s cap.

### D3. One prompter reads the whole configure run

* `runSetupInteractive` builds one `*wizard.TextPrompter` on `in` and
  `os.Stdout`, passes it to `wizard.ConfigureLLM`, and reads the four
  operational values with its `Input`.
* `readLine`, `promptInt` and the private `bufio.Reader` go. An invalid or
  non-positive value still keeps the current value, with the same note.
* The prompt shows `Timeout seconds (default 120):` where it showed
  `Timeout seconds [120]:`.

### D4. `configure --yes` refuses a refused key

* `discoverModels` returns `cat.Err` when the wizard's rule calls it a
  refused credential, and `--yes` fails with it, naming the provider. Any
  other listing failure still falls back to the static catalog, as now.

### D5. Every SDK call names the application

* `llmprovider.WithClientInfo(AppTitle, buildIdentity().Current())` is
  added to the hook's provider options, the wizard's `ProviderOptions`, and
  `--yes`'s listing.

### D6. Provider facts come from the registry

* **The key rule:** `ValidateActive` requires a key only when the
  provider's `Descriptor.RequiresAPIKey` is true. An unknown provider is
  refused, as `providers.New` already would.
* **The base URL:** `config.ProviderConfig` gains `base_url` (omitted when
  empty). `configure` saves `wizard.Result.BaseURL` when it differs from the
  descriptor's `DefaultBaseURL`, and passes the saved one back as
  `Existing.BaseURL`. The hook adds `llmprovider.WithBaseURL` when it is
  set. A config without the field behaves as now.
* **The provider list:** `config.SupportedProviders` is built from
  `providers.Default().Descriptors()`, in the registry's order, so
  `ApplyDefaults` writes a template for each of the 10 providers. The help
  text and `--provider`'s usage list the registry's IDs.

### D7. Tests follow the SDK's data

* `TestConfig_TemplateDefaults` expects, for each provider,
  `catalog.Static(id)[0]` and the next `config.MaxFallbacks`.
* `TestRunSetupInteractive_Success` types
  `min(len(catalog.Static(gemini)), catalog.MaxListed) + 1` for "Other".
* `analyzerTestProvider` becomes an `llmtest.NewFake(...).ReplyText(...)`.
* `TestReadmeCuratedModels` is unchanged; the README table is fixed (D9).
* `main.go:328`'s redundant `session.Store = store` goes.

### D8. The updater follows the spec's packaging

* `defaultNewUpdateUpdater` passes `Unpacker:` from the embedded spec's
  `Unpacker()`. It is nil for this repository's binary releases, so
  `update` behaves as now; an archive release would need no code change.

### D9. The documents state what changed

* The README's provider table and its example models follow the SDK's
  catalog.
* A section "Changes with go-llmprovider-sdk v1.11.0 and go-selfupdate-lib
  v1.15.2" follows the `v1.13.0` one, with D3 to D6 and SDK 0028's
  user-visible behaviour in a user's words, and the table of contents names
  it.
* `docs/architecture.md` names the base URL and the registry-driven
  provider list.

### D10. The release is `v1.9.0`

* The owner tags it after CI is green on `main`.
* Minor, not a patch, because `--yes` now refuses a refused key, the config
  gains a field and six provider templates, and the operational prompt's
  wording changes.

### Consequences

* Good, because typed operational values, Ollama, and a configured base
  URL work, through APIs the repository already imports.
* Good, because the next catalog refresh breaks only the README check,
  which names the rows to change.
* Good, because the menu and the defaults a user sees stay six models, the
  ones the SDK ranks best.
* Good, because the application names itself to every provider.
* Neutral, because `config.json` grows by six provider templates on the
  next save; a template is never read unless that provider is active.
* Bad, because a script that pipes `configure --yes` with a key the
  service refuses now fails, where it saved the key.
* Bad, because a hook older than `v1.9.0` sharing a `config.json` written
  by `v1.9.0` ignores `base_url` and still refuses Ollama.

### Confirmation

* **Every phase passes** `make verify`, `make verify-staged`, `go test
  -race -count=1 ./...`, `CGO_ENABLED=0 go vet` for linux, darwin and
  windows, `go mod tidy -diff`, `make check-records` and `make
  markdownlint`.
* **Each new test is seen to fail** on a scratch copy without its fix:
  the piped operational values (D3), the refused `--yes` key (D4), the
  keyless Ollama config (D6), the saved and applied base URL (D6), the
  client info in the User-Agent (D5), and the six-model menu (D2).
* **`TestMigrationByteForByte` and the `update` tests pass unchanged.**
* **CI passes on Linux, macOS and Windows, on `main` and on the tag,** and
  the tag's release is published through `436a7f5`.
* **The installed hook,** once updated, reports both library versions in
  its build information and writes a commit message on a real key.

## Pros and Cons of the Options

### A. Both libraries, six-model menu, every API the libraries cover

* Good, because it fixes two probed defects and one dropped answer with
  library calls, not new code.
* Good, because it removes the four-provider list, the private reader and
  a dead timeout.
* Bad, because it is the largest change: config, configure, the hook and
  their tests all move.

### B. Both libraries, the tests and the README only

* Good, because it is two lines in `go.mod` and data edits.
* Bad, because v1.11.0 then lengthens the menu, Ollama still fails, typed
  operational values are still lost, and the hand-rolled code stays.

### C. As A, with the full curated menu

* Good, because the menu shows every model the SDK knows works.
* Bad, because a commit-message hook wants a short list of fast models, as
  0038-MADR's drivers say for this repository.

### D. go-selfupdate-lib only

* Good, because every test passes with no other change.
* Bad, because it leaves the SDK eight minor releases behind, and does
  nothing the owner asked of the SDK.

## Owner questions

Each has a recommendation, written into the outcome above.

**Answered 2026-10-13:** "follow recommendations". The outcome above
stands as written, and this record is `accepted`.

* **Q1 (D2).** Keep the six-model menu (recommended), or take the full
  curated list (C)?
* **Q2 (D6).** Persist and apply the wizard's base URL, adding `base_url`
  to the config (recommended), or leave it dropped?
* **Q3 (D6).** Build the provider list from the registry, so `config.json`
  carries 10 templates (recommended), or keep the four?
* **Q4 (D4).** Make `configure --yes` fail on a refused key (recommended),
  or keep saving it with a warning?
* **Q5 (D10).** Release `v1.9.0` (recommended), or `v1.8.1`?

## More Information

### Amendment 1 (2026-10-13): go-llmprovider-sdk v1.12.0

**Status: accepted** by the owner on 2026-10-13 ("committed, proceed",
in answer to the request to approve it). D1 to D10 otherwise stand.

The owner asked that the most current releases be pinned. Checked on
2026-10-13 against both the module proxy and the repositories' tags
**[read]**:

| Library | Pinned (Phase 1) | Newest | Tagged |
| :--- | :--- | :--- | :--- |
| go-selfupdate-lib | `v1.15.2` | `v1.15.2` | 2026-10-13 00:46 UTC |
| go-llmprovider-sdk | `v1.11.0` | `v1.12.0` | 2026-10-13 19:32 UTC, after this record |

**What `v1.12.0` is [read].** One record, the SDK's
`0039-MADR-request-side-prompt-caching.md`. Its only exported change is
`llmprovider.CacheRetention` (`CacheShort`, `CacheLong`) and the field
`Request.Cache`; `Request.Validate` refuses any other value. "The zero
value asks for nothing: the request is sent as it would be without the
field" (`llmprovider/contract.go`; 0039-MADR D1). Its `go.mod` is
unchanged from `v1.11.0`'s.

**Measured on a scratch copy of the tree with Phases 1 and 2 [probed]:**
`go get …@v1.12.0` and `go mod tidy` change only the SDK's lines in
`go.mod` and `go.sum`; `go build`, `CGO_ENABLED=0 go vet` for linux,
darwin and windows, `go test -race -count=1 ./...` (5 packages) and `go
mod tidy -diff` all exit 0.

**A1.** `go.mod` requires `go-llmprovider-sdk v1.12.0`. go-selfupdate-lib
and the workflow pins stay at `v1.15.2`, its newest.

**A2. `Request.Cache` is left at its zero value.** The SDK's own record
names this repository as the caller that would only pay for it: one
request per commit, each with a new diff, so "a write it never reads costs
a quarter more" on Anthropic (0039-MADR, "What the consumers need" and
option D). The prompt's fixed prefix (`main.go:457-458`) is two sentences,
far below the 512 to 4,096 tokens below which a prompt "will be processed
without caching" (0039-MADR, Anthropic). This is the one new API the
release adds, and it is not used, deliberately.

**A3. A cost the SDK cannot switch off [read].** OpenAI caches
automatically on GPT-5.6 and later, at 1.25 times the input price for a
write (0039-MADR, OpenAI). D2's catalog makes `gpt-6-luna` OpenAI's
default where `v1.3.2` made `gpt-4.1-mini`, which has no write charge. The
SDK sends no field that turns this off (0039-MADR D3), so a user on
OpenAI's default pays it on each commit's prompt. The README's changes
section says so (D9).

**Confirmation** is D1's, at `v1.12.0`; `scripts/go-precheck.py` resolves
`v1.12.0` from GitHub.

* go-llmprovider-sdk records, by full filename in that repository:
  `docs/decisions/0028-MADR-heuristics-and-performance-from-the-research-pass.md`,
  `0029-MADR-go-1-27-2-for-standard-library-fixes.md`,
  `0031-MADR-native-streaming-on-the-responses-wire.md`,
  `0033-MADR-public-model-metadata-and-current-catalogs.md`,
  `0038-MADR-full-curated-catalog-by-default.md`.
* go-selfupdate-lib records, by full filename in that repository:
  `docs/decisions/0021-MADR-task-scheduler-lifecycle-and-windowless-children.md`,
  `0022-MADR-documentation-pass-after-v1-14-0.md`,
  `0024-MADR-documentation-pass-after-v1-15-1.md`.
* [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md)
  D2 and
  [0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md](0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md):
  the pin rule D1 applies.
* **Not decided here:**
  * acp-go-sdk2, set aside by the owner on 2026-10-13;
  * mcplib's README section, its code comments and records 0001 to 0008;
  * update-available notices (`Checker.CheckCached`), which nothing here
    does by hand today;
  * `migration_test.go`'s exact asset selector, a test that already
    passes;
  * an end-to-end `update` test on `selfupdatetest.NewGitHubServer`.
