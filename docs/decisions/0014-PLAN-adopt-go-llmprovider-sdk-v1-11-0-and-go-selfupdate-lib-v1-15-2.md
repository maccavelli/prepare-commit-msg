---
status: in-progress
date: 2026-10-13
associated-madr: "0014-MADR-adopt-go-llmprovider-sdk-v1-11-0-and-go-selfupdate-lib-v1-15-2.md"
---
# Implement the move to go-llmprovider-sdk v1.11.0 and go-selfupdate-lib v1.15.2, and the API adoption

Associated MADR: [0014-MADR-adopt-go-llmprovider-sdk-v1-11-0-and-go-selfupdate-lib-v1-15-2.md](0014-MADR-adopt-go-llmprovider-sdk-v1-11-0-and-go-selfupdate-lib-v1-15-2.md)

## Goal

* **D1:** `go.mod` requires `go-llmprovider-sdk v1.11.0` and
  `go-selfupdate-lib v1.15.2`; both `ci.yml` workflow calls are pinned to
  `436a7f5976a6815491c7384527fef90869574a1b # v1.15.2`.
* **D2:** `configure` and `configure --yes` recommend the six
  `catalog.MaxListed` models; `DiscoveryTimeout` is gone.
* **D3:** one `wizard.TextPrompter` reads the whole interactive run, and
  piped operational values are saved.
* **D4:** `configure --yes` fails on a key the service refuses.
* **D5:** every provider and listing carries `WithClientInfo`.
* **D6:** the key rule, the base URL and the provider list come from the
  registry; an Ollama config runs the hook.
* **D7:** the tests derive their expectations from the SDK.
* **D8:** the updater takes the spec's `Unpacker()`.
* **D9:** the README and `docs/architecture.md` say what changed.
* **D10:** `v1.9.0` is released and the installed hook updated.

Done means every item under Verification holds.

## Scope

### In scope

| Phase | Who | What |
| :--- | :--- | :--- |
| 0 | agent, then owner | this pair and the index; the owner approves |
| 1 | agent | D1, D2, D7 (catalog drift), D9 (the README table): both modules, green |
| 2 | agent | D3, D4: configure's prompter and `--yes` |
| 2a | agent | MADR Amendment 1: go-llmprovider-sdk `v1.12.0` |
| 3 | agent | D6: key rule, base URL, provider list, help text |
| 4 | agent | D5, D8, D7 (`llmtest` fake), `session.Store` |
| 5 | agent | D9: the README section and `docs/architecture.md` |
| 6 | owner, then agent | CI on `main`, tag `v1.9.0`, live checks |
| 7 | agent | close-out |

### Out of scope

* The MADR's "Not decided here": acp-go-sdk2; mcplib's README section,
  code comments and records 0001 to 0008; update notices;
  `migration_test.go`'s selector; an end-to-end `update` test.
* The code the MADR lists as having no library equivalent.
* Any other dependency, toolchain or tool version.

## Rules for every phase

* **Branch.** All work is on `main`; no branch is created (owner,
  2026-10-13). The organization's policy forbids an agent commit on
  `main`, so the agent stages and the owner commits and pushes.
* **Fail first.** Every new test is run on a scratch copy of the tree
  without its fix, and the failure is recorded in the Execution Record.
  The tree itself is never dirtied for the experiment.
* **Pre-add.** Before each commit: `gofmt -l` on every staged `.go` file,
  `make verify-staged`, `make check-records`, `make markdownlint`. A file
  that fails is not committed.
* **Phase gate.** Before each commit: `go test -race -count=1 ./...`,
  `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...`, `go mod tidy
  -diff`, `make verify`. Each exit status is captured before any filter.
* **Stage, then the owner commits.** When a phase's gates pass, the agent
  stages exactly that phase's files and shows `git diff --cached --stat`.
  The owner runs `git commit --no-edit` (the global hook writes the
  message) and pushes. One commit per phase. The next phase starts only
  after the owner has committed the last one.
* **Deviations.** Stop, record a dated entry here (and in the MADR if a
  decision or a fact changes), and wait for approval.

## Implementation Steps

### Phase 0: records

1. Write this pair; add both rows to `docs/README.md`.
2. `make check-records`, `make markdownlint`.
3. Present the pair. The owner answers the MADR's Q1 to Q5, and the MADR
   becomes `accepted`.
4. The agent stages the records alone; the owner commits them.

### Phase 1: both modules (D1, D2, D7 drift, D9 table)

1. `go get github.com/maccavelli/go-llmprovider-sdk@v1.11.0
   github.com/maccavelli/go-selfupdate-lib@v1.15.2`, then `go mod tidy`.
   Assert `git diff --stat` touches only `go.mod` and `go.sum`.
2. `ci.yml:99`, `:114`: replace `5e199c831b5691ea687943e3c3fd495d50c739ed #
   v1.13.0` with `436a7f5976a6815491c7384527fef90869574a1b # v1.15.2`;
   update the comment at `:93-98` to name `v1.15.2` and this MADR. Run
   `make workflow-lint`.
3. `internal/ui/setup.go`:
   * `wizard.Options`: add `ModelLimit: catalog.MaxListed`; drop
     `DiscoverLimit`.
   * `discoverModels`: drop the `context.WithTimeout`; append
     `catalog.WithLimit(catalog.MaxListed)` to the listing options.
   * delete `DiscoveryTimeout` and its comment; fix `discoverModels`' doc
     comment.
4. Tests:
   * `TestConfig_TemplateDefaults`: for each of gemini, openai, claude,
     expect `Model == catalog.Static(id)[0]` and `FallbackModels ==
     Static(id)[1:min(len, config.MaxFallbacks+1)]`.
   * `TestRunSetupInteractive_Success` (and any other test typing a menu
     position): compute "Other" as
     `strconv.Itoa(min(len(catalog.Static(gemini)), catalog.MaxListed)+1)`.
   * ~~New `TestSetup_MenuIsShortList`: the offline interactive menu for a
     provider whose static catalog exceeds `MaxListed` (pick one by
     iterating the registry) lists exactly `MaxListed` models before
     "Other". Fail-first: a scratch copy without `ModelLimit`.~~
     Superseded by Deviation 1 (2026-10-13): `TestDiscoverModels_ShortList`
     and `TestRunSetupInteractive_ShortList` in
     `internal/ui/menu_limit_test.go`, over a live OpenCode Zen listing of
     eight known models.
5. `README.md:84-93`: each provider's models from `catalog.Static`, first
   marked "(recommended)"; fix the example models at `:72`, `:273-276`,
   `:292-293`. `TestReadmeCuratedModels` passes.
6. Phase gate; stage; the owner commits.

### Phase 2: configure reads through one prompter (D3, D4)

1. `runSetupInteractive`: build `p := &wizard.TextPrompter{In: in, Out:
   os.Stdout}`; pass `p` to `ConfigureLLM`; change `promptOperational` to
   take a `wizard.Prompter` and read each value with `p.Input(label,
   strconv.Itoa(current))`, keeping the "invalid value … keeping N" note.
   Delete `readLine`, `promptInt`, the `bufio` reader and imports left
   unused.
2. New `TestRunSetupInteractive_OperationalTyped`: pipes `99`, `12345`,
   `7`, `9` and asserts all four are saved. Fail-first: the scratch probe
   from the MADR (`f1c3523` kept `120`, `32000`, `3`, `3`).
3. `discoverModels` returns `([]string, error)`. When `cat.Err` satisfies
   `errors.Is(err, llmprovider.ErrAuthFailure) && !errors.Is(err,
   llmprovider.ErrNotPermitted)`, return it; `runSetupNonInteractive`
   fails with `configure: <provider> refused the API key: <err>`. Other
   listing failures keep the static fallback.
4. New `TestRunSetupNonInteractive_RefusedKey`: an `httptest` server
   answering Gemini's listing with HTTP 400 `API_KEY_INVALID` through
   `listingClient`; asserts the error and that no config is saved. A second
   case answers 403 and asserts the static fallback is saved. Fail-first:
   scratch copy without step 3.
5. Update the existing operational-prompt tests for the `(default N)`
   wording.
6. Phase gate; stage; the owner commits.

### Phase 2a: go-llmprovider-sdk v1.12.0 (MADR Amendment 1)

Added 2026-10-13, after the owner asked for the most current releases.
Runs after the owner commits Phase 2, and before Phase 3.

1. `go get github.com/maccavelli/go-llmprovider-sdk@v1.12.0`, then `go mod
   tidy`. Assert `git diff --stat` touches only `go.mod` and `go.sum`.
2. Re-run the curated-model dump; if any static catalog differs from
   `v1.11.0`'s, update the README table and examples as in Phase 1 step 5
   (`TestReadmeCuratedModels` decides).
3. `Request.Cache` is not set (Amendment 1, A2).
4. Phase gate, plus `make verify-staged` (go-precheck resolves `v1.12.0`
   from GitHub); stage; the owner commits.

### Phase 3: provider facts from the registry (D6)

1. `internal/config`:
   * `SupportedProviders`: built once from `providers.Default().
     Descriptors()` IDs, in registry order.
   * `ValidateActive`: look up the descriptor; require a key only when
     `RequiresAPIKey`; refuse an unknown provider by name.
   * `ProviderConfig` gains `BaseURL string \`json:"base_url,omitempty"\``.
2. `internal/ui/setup.go`: set `existing.BaseURL = pc.BaseURL`; after the
   wizard, save `res.BaseURL` when it differs from the descriptor's
   `DefaultBaseURL`, else clear it.
3. `main.go`: `providerOptions` appends `llmprovider.WithBaseURL(pc.BaseURL)`
   when set. The help text's `--provider` line and `runConfigure`'s flag
   usage list the registry's IDs (joined with `|` and `, `).
4. Tests:
   * `TestValidateActive_KeylessProvider`: Ollama with no key passes; Gemini
     with no key fails as now. Fail-first: the scratch probe from the MADR.
   * `TestNewActiveProvider_BaseURL`: a saved `base_url` pointing at an
     `httptest` server receives the hook's request.
   * `TestRunSetupInteractive_SavesBaseURL`: typing a non-default Ollama
     URL saves it; accepting the default saves none.
   * `TestApplyDefaults_EveryRegistryProvider`: a template exists for every
     descriptor ID.
   * Fail-first: each on a scratch copy without its step.
5. Phase gate; stage; the owner commits.

### Phase 4: client info, unpacker, SDK test fakes (D5, D8, D7)

1. Add one helper, `clientInfo() llmprovider.Option`, returning
   `llmprovider.WithClientInfo(AppTitle, buildIdentity().Current())`
   (`internal/ui` takes the name and version from `main` through
   `SetupOptions` or a package variable set in `main`, whichever leaves no
   import cycle). Pass it to `newSourceProvider`, the wizard's
   `ProviderOptions`, and `discoverModels`.
2. New `TestProvider_UserAgentNamesApp`: an `httptest` server records the
   `User-Agent`; assert it begins `prepare-commit-msg/`. Fail-first:
   scratch copy without step 1.
3. `update.go`: `releaseAssets` returns the spec too (or a sibling returns
   `spec.Unpacker()`); `selfupdate.Config` takes `Unpacker`. Extend
   `releasespec_test.go` to assert the embedded spec's `Unpacker()` is nil.
4. `main.go:328`: delete `session.Store = store`; the OAuth tests pass.
5. `main_oauth_test.go`: replace `analyzerTestProvider` with
   `llmtest.NewFake(llmprovider.ProviderOpenAI,
   llmprovider.Capabilities{}).ReplyText("feat: use OAuth session")`.
6. Phase gate; stage; the owner commits.

### Phase 5: documents (D9)

1. `README.md`: add "Changes with go-llmprovider-sdk v1.11.0 and
   go-selfupdate-lib v1.15.2" after the `v1.13.0` section, covering D2 to
   D6 and SDK 0028's user-visible behaviour; add it to the table of
   contents; document `base_url` under "Configuration File"; make the
   `--provider` row match the registry.
2. `docs/architecture.md`: the registry-driven provider list and the base
   URL.
3. `make markdownlint`, `make check-records`; stage; the owner commits.

### Phase 6: release `v1.9.0` (owner, then agent)

1. Before each of the owner's pushes, the agent runs the pre-push
   disclosure guard from `AGENTS.md` over the outgoing commits and
   reports the result.
2. After Phase 5's push, CI on `main` is green on Linux, macOS and
   Windows, including the build rehearsal through `436a7f5`.
3. The owner tags `v1.9.0` and pushes the tag; CI publishes through
   `436a7f5`.
4. Agent checks the release assets and attestations, runs `update` on the
   installed hook, and confirms `version` names both libraries'
   versions, a real commit message, and a keyless Ollama run if Ollama is
   available on the host.

### Phase 7: close-out

1. Execution Record complete; this PLAN `complete`; the index updated;
   stage; the owner commits and pushes.

## Verification

* `go.mod` names go-llmprovider-sdk `v1.12.0` (MADR Amendment 1) and
  go-selfupdate-lib `v1.15.2`; `go mod tidy -diff` is empty;
  `scripts/go-precheck.py` passes.
* `ci.yml`'s two `uses:` name `436a7f5…`; `make workflow-lint` passes.
* `go test -race -count=1 ./...` passes on all five packages, including
  the nine new tests (eight as planned, with Deviation 1's two in place of
  one), each recorded failing on a scratch copy without its
  fix.
* `TestMigrationByteForByte`, the `update` tests and
  `TestReadmeCuratedModels` pass.
* `grep -n 'DiscoveryTimeout\|promptInt\|analyzerTestProvider'` over the Go
  sources finds nothing.
* CI green on `main` and on the tag; `v1.9.0` published through
  `436a7f5` with attestations.
* The installed hook reports `v1.9.0` and writes a commit message.

## Rollout and Rollback

* **Rollout:** users receive `v1.9.0` through `update`. A `config.json`
  written by `v1.8.0` loads unchanged; `ApplyDefaults` adds the new
  templates on the next save.
* **Rollback:** each phase is one commit and reverts alone. Reverting
  Phase 3 leaves a `base_url` field that `v1.8.0`'s loader ignores. A bad
  release is superseded by `v1.9.1`, not deleted; users go back with
  `update` to the previous tag only through a new release.

## Execution Record

### Amendment (2026-10-13): no branch; the agent stages, the owner commits

As written, the PLAN cut a branch and had the agent commit each phase and
merge by pull request. The owner uses no branches. The organization's
policy forbids an agent commit, push or merge on `main`, so the owner
chose: work on `main`, the agent stages each phase, the owner commits and
pushes. "Rules for every phase", Phase 0 step 4, Phase 6 and Phase 7 were
changed to match. No decision in the MADR changes.

### Phase 0: records (2026-10-13)

The pair and the index rows were written; the owner answered Q1 to Q5
"follow recommendations", and the MADR is `accepted`. Staged for the
owner's commit.

Committed by the owner as `703ea27`. The owner then said "proceed".

### Phase 1: both modules (2026-10-13)

**Step 1.** `go get` of both versions, then `go mod tidy`: both exited 0;
`go: upgraded github.com/maccavelli/go-llmprovider-sdk v1.3.2 => v1.11.0`
and `go: upgraded github.com/maccavelli/go-selfupdate-lib v1.13.0 =>
v1.15.2`. `git diff --stat`: `go.mod` (2 lines) and `go.sum` (4 lines)
only.

**Step 2.** Both `uses:` in `ci.yml` name
`436a7f5976a6815491c7384527fef90869574a1b # v1.15.2`; the comment names
`v1.15.2` and this MADR's D1. A grep of `.github/` for `5e199c8` and
`v1.13.0` finds nothing. `make workflow-lint`: 0.

**Step 3.** `wizard.Options` takes `ModelLimit: catalog.MaxListed` in
place of `DiscoverLimit`; `discoverModels` passes
`catalog.WithLimit(catalog.MaxListed)` and no longer wraps the context;
`DiscoveryTimeout` and the `time` import are gone.

**Step 4.** `TestConfig_TemplateDefaults` loops over gemini, openai and
claude against `catalog.Static`; `TestRunSetupInteractive_Success` types
"Other" at `min(len(catalog.Static(gemini)), catalog.MaxListed) + 1`. Of
the other interactive inputs, none depended on the menu's length, and all
pass unchanged.

**Step 5.** The provider table's eight drifted rows (gemini, openai,
claude, grok, kilo, opencode-zen, opencode-go, together) were rewritten
from a dump of `catalog.Static` for every registry provider at `v1.11.0`,
made by a scratch program. The example at `README.md:72`, the gemini and
claude `--yes` examples, and the example `config.json`'s gemini, openai
and claude entries follow the same catalog. The example `config.json` was
not named in step 5; it is one of the "example models" of MADR D9.

### Deviation 1 (2026-10-13): no static catalog exceeds `MaxListed`

* **Found:** step 4's `TestSetup_MenuIsShortList` needs "a provider whose
  static catalog exceeds `MaxListed`". There is none: `MaxListed` "also
  bounds each static catalog" (go-llmprovider-sdk `v1.11.0`,
  `llmprovider/catalog/models_catalog.go:26-29`). Offline, the menu is
  never longer than six, with or without `ModelLimit`, so that test could
  not fail. The error was in this PLAN, not the code.
* **Decision (owner, "approve the deviation"):** two tests over a live
  OpenCode Zen listing of eight models the SDK's catalog knows, routed to
  a local server through the existing `redirectTransport`:
  `TestDiscoverModels_ShortList` (`--yes` recommends exactly `MaxListed`)
  and `TestRunSetupInteractive_ShortList` (choice `MaxListed + 1` is
  "Other").
* **Files added to the phase:** `internal/ui/menu_limit_test.go`.
* **The MADR** is unchanged: D2 and its Confirmation hold.

**Fail-first,** on a scratch copy of the tree with `ModelLimit` and
`WithLimit` removed from `setup.go` (the tree was not touched):

```text
menu_limit_test.go:63: discoverModels recommends 8 models ["deepseek-v4.1-flash" "qwen3.8-flash" "glm-5.3-flash" "deepseek-v4-flash" "gemini-3.5-flash-lite" "gemini-3.8-flash" "claude-sonnet-5" "claude-opus-5"]; want catalog.MaxListed (6)
--- FAIL: TestDiscoverModels_ShortList
menu_limit_test.go:93: model = "claude-sonnet-5"; want my-custom-model, chosen as Other at 7
--- FAIL: TestRunSetupInteractive_ShortList
```

The menu there ended `9) Other (enter a model id)`. With the fix, it
lists six models and `7) Other (enter a model id)`, and both tests pass.

**Phase gate,** each exit status captured before any filter:

| Check | Result |
| :--- | :--- |
| `gofmt -l` on the changed `.go` files | 0, no output |
| `go test -race -count=1 ./...` | 0; 5 packages ok |
| `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | 0 each |
| `go mod tidy -diff` | 0, empty |
| `make verify` | 0; `golangci-lint run`: `0 issues.`; total coverage 81.9% (minimum 80.0%); `govulncheck`: `No vulnerabilities found.`; `build-all` built six targets |

Staged for the owner's commit.

Committed by the owner as `bd5d46d`. The owner then said "proceed".

### Phase 2: configure reads through one prompter (2026-10-13)

**Step 1 (D3).** `runSetupInteractive` builds one `wizard.TextPrompter`
and passes it to `wizard.ConfigureLLM` and to `promptOperational`.
`promptInt` takes a `wizard.Prompter` and reads with `Input(label,
strconv.Itoa(current))`. `readLine`, the private `bufio.Reader` and the
`bufio` import are gone; a grep of `internal/ui` for `readLine` and
`bufio` finds nothing.

* **Noted, within the step:** the SDK's prompter returns an error wrapping
  `io.EOF` once input is exhausted (`errExhausted`,
  `wizard/text_prompter.go:63`). The old `readLine` answered end of input
  with a blank, which kept the current value. `promptInt` keeps that: an
  `io.EOF` keeps the current value, so a piped run that stops after the
  wizard's answers still saves the defaults. `TestPromptInt` gains the
  case "end of input keeps current".

**Step 2.** `TestRunSetupInteractive_OperationalTyped` pipes `99`,
`12345`, `7`, `9` and requires all four saved.

**Step 3 (D4).** `discoverModels` returns `([]string, error)`: a refused
credential, from `List`'s error or `cat.Err`, by the wizard's rule
(`refusedCredential`, the SDK's unexported two-clause check repeated);
any other failure keeps the static fallback. Both `--yes` call sites
return `refusedKeyError` and save nothing.

* **Noted, within the step:** the message is `<provider> refused the API
  key: <err>`, without the step's `configure:` prefix, because `main`
  already prints configure's errors as `Setup failed: <err>`.

**Step 4.** `TestRunSetupNonInteractive_RefusedKey` serves Gemini's
`API_KEY_INVALID` refusal (the SDK's captured fixture, unwrapped from the
Interactions API's array) with HTTP 400, and requires `ErrAuthFailure`,
the message, and no `config.json`. Its control case serves a 403
(`PERMISSION_DENIED`) and requires the static catalog's first model
saved.

**Step 5.** No existing test asserted the `[N]` wording, so none changed.
The three `discoverModels` call sites in the tests take its error.

**Fail-first,** on a scratch copy of the tree with D3 undone (the
operational prompts on a second `TextPrompter` over `in`) and D4 undone
(`cat.Err` ignored); the tree was not touched:

```text
setup_test.go:623: operational values = [120 32000 3 3]; want the typed [99 12345 7 9]
--- FAIL: TestRunSetupInteractive_OperationalTyped
setup_test.go:679: err = <nil>; want gemini's refusal, ErrAuthFailure
--- FAIL: TestRunSetupNonInteractive_RefusedKey/invalid_key
--- PASS: TestRunSetupNonInteractive_RefusedKey/not_permitted
```

The 403 control passes both with and without the fix, as it should.

**Phase gate,** each exit status captured before any filter:

| Check | Result |
| :--- | :--- |
| `gofmt -l` on the changed `.go` files | 0, no output |
| `go test -race -count=1 ./...` | 0; 5 packages ok |
| `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | 0 each |
| `go mod tidy -diff` | 0, empty |
| `make verify` | 0; `golangci-lint run`: `0 issues.`; total coverage 82.2% (minimum 80.0%); `govulncheck`: `No vulnerabilities found.` |

Staged for the owner's commit.

### Amendment 2 (2026-10-13): Phase 2a, go-llmprovider-sdk v1.12.0

The owner asked that the most current releases be pinned.
go-llmprovider-sdk `v1.12.0` was tagged after this pair was written; the
MADR's Amendment 1 records what it is and the measurements. Phase 2a is
added between Phase 2 and Phase 3; nothing already done changes.

Phase 2 was committed by the owner as `8df87c7`, with Amendment 2 left out
of that commit. The owner then said "committed, proceed", approving the
MADR's Amendment 1 and Phase 2a.

### Phase 2a: go-llmprovider-sdk v1.12.0 (2026-10-13)

**Step 1.** `go get github.com/maccavelli/go-llmprovider-sdk@v1.12.0`
(exit 0: `upgraded … v1.11.0 => v1.12.0`), then `go mod tidy` (exit 0).
Outside the two records, `git diff --stat` names only `go.mod` (1 line)
and `go.sum` (the SDK's 2 lines); go-selfupdate-lib stays at `v1.15.2`.

**Step 2.** The curated-model dump (each registered provider's
`catalog.Static`, encoded as JSON) was run on scratch copies at `v1.11.0`
and `v1.12.0`: both outputs are 1,136 bytes and `cmp` exits 0. The README
table and examples therefore do not change, and `TestReadmeCuratedModels`
passes.

**Step 3.** `Request.Cache` is not set: a grep of the repository's `.go`
files for `Cache`, `CacheShort`, `CacheLong` and `CacheRetention` finds
nothing (exit 1).

**Fail-first:** not applicable; the phase adds no gate, assertion or test.

**Phase gate,** each exit status captured before any filter:

| Check | Result |
| :--- | :--- |
| `go test -race -count=1 ./...` | 0; 5 packages ok |
| `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | 0 each |
| `go mod tidy -diff` | 0, empty |
| `go test -count=1 -run TestReadmeCuratedModels ./...` | 0; PASS |
| `make verify` | 0; `golangci-lint run`: `0 issues.`; total coverage 82.2% (minimum 80.0%); `govulncheck`: `No vulnerabilities found.` |
| `make verify-staged` | 0; go-precheck resolves go-llmprovider-sdk `v1.12.0` and go-selfupdate-lib `v1.15.2` from GitHub; `0 issues.`; `No vulnerabilities found.` |
| `make check-records`, `make markdownlint` | 0 each |

Staged for the owner's commit.

Committed by the owner as `ea1ff9c`. Before the next phase the agent ran
the disclosure guard over the four commits not yet on `origin/main`
(`703ea27`, `bd5d46d`, `8df87c7`, `ea1ff9c`): exit 0, no output. The owner
then said "proceed".

### Phase 3: provider facts from the registry (2026-10-13)

The SDK's registry (`providers.Default().Descriptors()`, at `v1.12.0`)
lists 10 providers. Only Ollama has `RequiresAPIKey` false; OpenCode Zen,
OpenCode Go, Hugging Face, Kilo, Together and Ollama have
`SupportsBaseURL` and a `DefaultBaseURL` **[probed]**.

**Step 1 (`internal/config`).** `SupportedProviders` is built once from
the registry's descriptor IDs, in its order. `ValidateActive` looks the
provider up, refuses an unknown one by name (`unsupported provider %q
(known: …)`), and requires a key only when `RequiresAPIKey` is true.
`ProviderConfig` gains `BaseURL` (`json:"base_url,omitempty"`).

**Step 2 (`internal/ui/setup.go`).** `existing.BaseURL` is the saved
`base_url`; after the wizard, `savedBaseURL` keeps `res.BaseURL` only when
it differs from the descriptor's `DefaultBaseURL`, and clears it otherwise.

**Step 3 (`main.go`).** `providerOptions` adds
`llmprovider.WithBaseURL(pc.BaseURL)` when it is set, before the Kilo
organization option. The help text's `--provider` line joins the
registry's IDs with `|`, and the flag's usage with `, `.

**Step 4 (tests).**

* `TestValidateActive_KeylessProvider`: Ollama with no key passes, Gemini
  with no key fails, and an unknown provider is refused by name.
* `TestNewActiveProvider_BaseURL` runs `runAnalyzer` against an Ollama
  config whose `base_url` is an `httptest` server answering
  `/v1/chat/completions`, and requires the request and the written
  message. It covers the hook's key rule too: Ollama reaches the server
  with no key.
* `TestRunSetupInteractive_SavesBaseURL` (Ollama, its requests redirected
  to a fixture answering `/api/version` and `/api/tags`): a typed endpoint
  is saved; accepting the default saves none.
  * **Noted, within the step:** a third case, "saved endpoint kept", seeds
    a saved `base_url` and accepts the offered default. It is the only case
    that tests `existing.BaseURL`.
* `TestApplyDefaults_EveryRegistryProvider`: every descriptor ID has a
  template, holding `DefaultModelForProvider`'s model.

**Fail-first,** each on a scratch copy of the tree with one step undone;
the tree was not touched:

```text
key rule undone (every provider needs a key, no unknown-provider check):
--- FAIL: TestNewActiveProvider_BaseURL
    provider_test.go:198: runAnalyzer: no API key for provider "ollama"; run 'prepare-commit-msg configure'
--- FAIL: TestValidateActive_KeylessProvider
    config_test.go:569: Ollama without a key: no API key for provider "ollama"; run 'prepare-commit-msg configure'; want nil, its descriptor needs no key
    config_test.go:576: unknown provider: <nil>; want it refused by name
provider list undone (the four hard-coded providers):
--- FAIL: TestApplyDefaults_EveryRegistryProvider
    config_test.go:588: no template for opencode-zen
    (likewise opencode-go, huggingface, kilo, together, ollama)
step 2 undone (endpoint neither offered back nor saved):
    --- FAIL: TestRunSetupInteractive_SavesBaseURL/typed_endpoint
        setup_test.go:766: base_url = ""; want "http://ollama.test:11434"
existing.BaseURL alone undone:
    --- FAIL: TestRunSetupInteractive_SavesBaseURL/saved_endpoint_kept
        setup_test.go:766: base_url = ""; want "http://ollama.test:11434"
step 3 undone (no WithBaseURL):
--- FAIL: TestNewActiveProvider_BaseURL
    provider_test.go:198: runAnalyzer: all models for ollama failed, last error: model llama3.3 failed: llmprovider: invalid request: ollama HTTP 404 not_found_error: model 'llama3.3' not found
```

With step 2 wholly undone, "saved endpoint kept" passes, because the old
`base_url` is never overwritten; the `existing.BaseURL` run is the one
that shows that case failing. With step 3 undone, the request went to the
default `localhost:11434`, where an Ollama on the test machine answered
404; with no Ollama running, the request is refused instead. Either way,
the fixture gets no request.

**Phase gate,** each exit status captured before any filter:

| Check | Result |
| :--- | :--- |
| `gofmt -l` on the six changed `.go` files | 0, no output |
| `go test -race -count=1 ./...` | 0; 5 packages ok |
| `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | 0 each |
| `go mod tidy -diff` | 0, empty |
| `make verify` | 0; `golangci-lint run`: `0 issues.`; total coverage 82.5% (minimum 80.0%); `govulncheck`: `No vulnerabilities found.` |
| `make verify-staged` | 0; `0 issues.`; `No vulnerabilities found.` |
| `make check-records`, `make markdownlint` | 0 each |

**Found, out of scope, not done.** `configure --yes` still requires a key
for every provider: `runSetupNonInteractive` returns `API key required:
pass --api-key or set ` with an empty variable name for Ollama
(`internal/ui/setup.go`, the `apiKey == ""` check), and it neither takes
nor saves a base URL. D6 names the hook's key rule and the wizard's
endpoint, not the `--yes` path, so this phase leaves it as it was. It
waits for an amendment and the owner's approval.

Staged for the owner's commit.
