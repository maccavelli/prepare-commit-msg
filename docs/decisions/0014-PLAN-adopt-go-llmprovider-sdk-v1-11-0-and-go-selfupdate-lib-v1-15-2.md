---
status: proposed
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
   * New `TestSetup_MenuIsShortList`: the offline interactive menu for a
     provider whose static catalog exceeds `MaxListed` (pick one by
     iterating the registry) lists exactly `MaxListed` models before
     "Other". Fail-first: a scratch copy without `ModelLimit`.
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

* `go.mod` names `v1.11.0` and `v1.15.2`; `go mod tidy -diff` is empty;
  `scripts/go-precheck.py` passes.
* `ci.yml`'s two `uses:` name `436a7f5…`; `make workflow-lint` passes.
* `go test -race -count=1 ./...` passes on all five packages, including
  the eight new tests, each recorded failing on a scratch copy without its
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
