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
| 3a | agent | MADR Amendment 2: `configure --yes` key rule and `--base-url` |
| 4 | agent | D5, D8, D7 (`llmtest` fake), `session.Store` |
| 5 | agent | D9: the README section and `docs/architecture.md` |
| 5a | agent | Amendment 4: the unpushed `main` tested on a Windows and a Linux host |
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
   Delete `readLine`, ~~`promptInt`,~~ the `bufio` reader and imports left
   unused. *(Deviation 2, 2026-10-13: `promptInt` stays, reading through
   the prompter's `Input`.)*
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

### Phase 3a: `configure --yes` follows the registry (MADR Amendment 2)

Added 2026-10-13, after Phase 3 found `--yes` still requiring a key of
every provider. Runs after the owner commits Phase 3, and before Phase 4.

1. `main.go`: a `--base-url` flag, in `runConfigure` and in
   `printUsage`, passed as `ui.SetupOptions.BaseURL`.
2. `internal/ui/setup.go`:
   * `SetupOptions` gains `BaseURL`.
   * `runSetupNonInteractive`: take the descriptor already looked up;
     return `API key required` only when `RequiresAPIKey` (B1).
   * The endpoint (B2): `opts.BaseURL`, else the saved `pc.BaseURL`, else
     none. A flag value is checked by `checkBaseURL(d, raw)`, one function
     citing the SDK's `wizard/base_url.go` (0021-MADR Z7), which refuses a
     provider without `SupportsBaseURL`, a scheme other than `http` or
     `https`, a missing host, a user name or password, a query or a
     fragment, and plain `http` to a non-loopback host when the provider
     requires a key.
   * `discoverModels(ctx, provider, apiKey, baseURL)` appends
     `llmprovider.WithBaseURL(baseURL)` when it is set (B3); both `--yes`
     call sites and the tests pass it.
   * `savedBaseURL` takes the provider and the URL, so both paths save by
     D6's rule.
   * `runSetupInteractive`: `opts.BaseURL`, when set, seeds
     `existing.BaseURL`.
3. Tests:
   * `TestRunSetupNonInteractive_KeylessProvider`: README's Example 2
     arguments (`ollama`, `qwen2.5-coder:7b`, `NoEnv`, no key) save the
     configuration; Gemini with no key still fails.
   * `TestRunSetupNonInteractive_BaseURL`, on Phase 3's Ollama fixture,
     which records each request's `Host`: a typed `--base-url` is the
     listing's host and is saved with the listed model; the default URL
     saves none; with no flag, a saved `base_url` is listed and kept.
   * `TestRunSetupNonInteractive_BaseURLRefused`: a provider without
     `SupportsBaseURL`, a user name and password, a query, an `ftp` scheme
     and plain `http` to a remote host for a key-requiring provider each
     fail, and nothing is saved.
   * `TestMain_RunConfigure_BaseURL`: `configure --yes --provider ollama
     --model m --base-url <url>` saves `base_url`; `printUsage` lists
     `--base-url`.
   * Fail-first: each on a scratch copy without its step.
4. A scratch build runs README's Example 2 with an empty configuration:
   exit 0, `ollama` saved with no key.
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
   `--provider` row match the registry; add a `--base-url` row to the flag
   table (MADR Amendment 2, B5).
2. `docs/architecture.md`: the registry-driven provider list and the base
   URL.
3. `make markdownlint`, `make check-records`; stage; the owner commits.

### Phase 5a: the unpushed `main` on Windows and Linux (Amendment 4)

Added 2026-10-13, after the owner asked for a test on two of their hosts
before the push. Runs after the owner commits Phase 5, and before Phase 6.
The hosts are the owner's Windows laptop (x86-64, Git Bash, Ollama
running) and a Linux host (Ubuntu, x86-64), reached over `ssh`; both have
Go 1.27.2, `git`, `make` and `gcc`. They are not named in any record.

1. **Ship.** `git bundle create` of `main` into the scratch directory;
   `scp` it to `~/tmp/pcm-0014-5a/` on each host; `git clone` it there.
   Assert the clone's `HEAD` is the owner's Phase 5 commit.
2. **Tests,** each exit status captured on the host before any filter:
   * both hosts: `go test -count=1 ./...` (CI's native job), `go test
     -race -count=1 ./...`, `go vet ./...`;
   * Linux: `make verify` (CI's quality contract).
3. **A scratch build** in the clone (`go build -o pcm`, `pcm.exe` on
   Windows). Every live run below points the configuration at
   `~/tmp/pcm-0014-5a/home`: `HOME` and `XDG_CONFIG_HOME` on Linux,
   `APPDATA` and `USERPROFILE` on Windows. The hosts' own configuration,
   OAuth sessions and installed hooks are not read or written; the
   installed hook's SHA-256 is taken before and after, and must not change.
4. **Live, both hosts:** `version`; `help` names the 10 providers and
   `--base-url`; `configure --yes --provider gemini --base-url
   https://example.com` and a `--base-url` carrying a user name are
   refused, and nothing is saved.
5. **Live, Windows, against its Ollama:**
   * README's Example 2, `configure --yes --provider ollama --model
     qwen2.5-coder:7b`, with an empty configuration: exit 0, `ollama`
     saved with no key and no `base_url`;
   * `configure --yes --provider ollama --base-url http://127.0.0.1:11434`
     with no `--model`: the live listing answers, and `base_url` is saved;
   * the hook: in a scratch Git repository under the scratch directory,
     stage one small file and run the scratch binary as the hook on
     `.git/COMMIT_EDITMSG`. It writes a Conventional Commit message from
     Ollama with no key. No commit is made.
6. **Live, Linux:** `GEMINI_API_KEY` set in the environment (never as an
   argument) to a made-up value, `configure --yes --provider gemini`:
   fails with `gemini refused the API key`, and nothing is saved (D4).
7. **Clean up.** `rm -rf ~/tmp/pcm-0014-5a` on each host, by that literal
   path; the installed hook's SHA-256 is unchanged.
8. Record the results here, with the hosts described as above; stage this
   PLAN alone; the owner commits. The agent runs the disclosure guard over
   all outgoing commits before the owner pushes.

Not done here: anything on a host's own configuration (Phase 6 does that
after the release, through `update`); a remote `--base-url` to Ollama
across hosts, which depends on how the laptop's Ollama is bound.

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
* Phase 3a's four tests (MADR Amendment 2) pass, each recorded failing on
  a scratch copy without its step, and README's Example 2 runs from a
  scratch build.
* `TestMigrationByteForByte`, the `update` tests and
  `TestReadmeCuratedModels` pass.
* `grep -n 'DiscoveryTimeout\|analyzerTestProvider'` over the Go sources
  finds nothing. *(Deviation 2, 2026-10-13: `promptInt` left this grep;
  it stays, and `grep -n 'bufio\|readLine' internal/ui/setup.go` finds
  nothing.)*
* CI green on `main` and on the tag; `v1.9.0` published through
  `436a7f5` with attestations.
* The installed hook reports `v1.9.0` and writes a commit message.
* Phase 5a (Amendment 4): the tests pass on the Windows and Linux hosts,
  README's Example 2 and a keyless Ollama hook run succeed on Windows, the
  refusals hold on both, and neither host's installed hook changed.

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

### Amendment 3 (2026-10-13): Phase 3a, `configure --yes` follows the registry

The owner asked that the plan and the MADR be amended with the fix for
`--yes` found in Phase 3. The MADR's Amendment 2 records the evidence
(README's Example 2 failing on a scratch build of the Phase 3 tree), the
options and the proposed choice. Phase 3a is added between Phase 3 and
Phase 4, and Phase 5 step 1 gains the `--base-url` row. Nothing already
done changes.

Phase 3 was committed by the owner as `a4d929a`, with Amendment 3 left out
of that commit. The owner then said "follow recommendations. committed,
proceed", approving the MADR's Amendment 2 with option A and Phase 3a.

### Phase 3a: `configure --yes` follows the registry (2026-10-13)

**Step 1 (`main.go`).** `configure` takes `--base-url`, passed as
`ui.SetupOptions.BaseURL`; `printUsage` lists it after `--api-key`.

**Step 2 (`internal/ui/setup.go`).**

* `SetupOptions` gains `BaseURL`.
* `runSetupNonInteractive` keeps the descriptor it looks up and returns
  `API key required` only when `RequiresAPIKey` (B1).
* The endpoint is `opts.BaseURL`, else the saved `pc.BaseURL`; it is
  saved through `savedBaseURL`, which now takes the provider and the URL
  so both paths call it (B2).
* `checkBaseURL(d, raw)` cites the SDK's `wizard/base_url.go` (0021-MADR
  Z7). It refuses, in order: a provider without `SupportsBaseURL`; an
  unparseable URL; a user name or password; a scheme other than `http` or
  `https`, or no host; a query or fragment; plain `http` to a non-loopback
  host when the provider requires a key. `loopbackHost` repeats the
  wizard's. The user-name check comes before any check that quotes the URL,
  so no error quotes a password.
* `discoverModels` takes the endpoint and appends
  `llmprovider.WithBaseURL` when it is set (B3); both `--yes` call sites
  pass it, and the three existing test call sites pass `""`.
* `runSetupInteractive`: `opts.BaseURL`, when set, seeds
  `existing.BaseURL`.
* **Noted, within the step:** `make verify` first failed on two goconst
  findings: the literal `"http"` appeared three times in the package,
  twice in `checkBaseURL` and once in `internal/ui/open.go:51`.
  `checkBaseURL` now names the schemes as local constants (`plain`,
  `secure`), and `open.go` is unchanged. The fail-first runs below were
  made before this rename, which changes no behaviour; the tests passed
  again after it.

**Step 3 (tests).**

* `TestRunSetupNonInteractive_KeylessProvider`: README's Example 2
  arguments save `ollama` with no key; Gemini with no key fails with `API
  key required` and saves nothing.
* `TestRunSetupNonInteractive_BaseURL`, on Phase 3's Ollama fixture, which
  now records each request's `Host` (`hostLog`): a typed `--base-url` is
  the only host listed and is saved with the listed model; the default URL
  saves none; with no flag, a saved `base_url` is listed and kept.
* `TestRunSetupNonInteractive_BaseURLRefused`: a provider without an
  endpoint, a user name and password, a query, a fragment, an `ftp` scheme
  and plain `http` to a remote host for Together each fail with nothing
  saved, and no error quotes the password. Its control saves plain `http`
  to `127.0.0.1`.
* `TestMain_RunConfigure_BaseURL`: `configure --yes --provider=ollama
  --model=m --fallback=f --base-url=…` saves `base_url`, and `printUsage`
  lists `--base-url`.
* **Noted, within the step:** `TestRunSetupInteractive_SavesBaseURL`
  (Phase 3) gains the case "flag seeds the endpoint", the one test of the
  seeding bullet above. The plan named no test for it.

**Fail-first,** each on a scratch copy of the tree with one step undone;
the tree was not touched:

```text
key rule (B1) undone:
--- FAIL: TestMain_RunConfigure_BaseURL
    main_extra2_test.go:176: configure unexpectedly called osExit(1)
--- FAIL: TestRunSetupNonInteractive_KeylessProvider
    setup_test.go:823: Ollama with no key: API key required: pass --api-key or set ; want it saved
    (and each TestRunSetupNonInteractive_BaseURL case: API key required)
checkBaseURL accepting everything:
    --- FAIL: TestRunSetupNonInteractive_BaseURLRefused/provider_without_an_endpoint
        setup_test.go:915: err = <nil>; want "takes no --base-url"
    (likewise the other five cases)
the plain-http case removed:
    --- FAIL: TestRunSetupNonInteractive_BaseURLRefused/plain_http_to_a_remote_host
        setup_test.go:915: err = <nil>; want "plain http to a remote host"
the user-name refusal quoting the URL:
    --- FAIL: TestRunSetupNonInteractive_BaseURLRefused/user_name_and_password
        setup_test.go:918: err quotes the password: --base-url https://user:<password>@example.test/v1 carries no user name or password
the endpoint not saved:
    main_extra2_test.go:186: base_url = ""; want the flag's
    --- FAIL: TestRunSetupNonInteractive_BaseURL/flag
        setup_test.go:885: base_url = ""; want "http://ollama.test:11434"
    setup_test.go:933: base_url = ""; want the loopback endpoint
the saved endpoint not used:
    --- FAIL: TestRunSetupNonInteractive_BaseURL/saved_endpoint_kept
        setup_test.go:878: listing hosts = ["localhost:11434"]; want only "ollama.test:11434"
        setup_test.go:885: base_url = ""; want "http://ollama.test:11434"
the listing without WithBaseURL (B3):
    --- FAIL: TestRunSetupNonInteractive_BaseURL/flag
        setup_test.go:878: listing hosts = ["localhost:11434"]; want only "ollama.test:11434"
    (and saved_endpoint_kept, the same)
the wizard not seeded:
    --- FAIL: TestRunSetupInteractive_SavesBaseURL/flag_seeds_the_endpoint
        setup_test.go:794: base_url = ""; want "http://ollama.test:11434"
the flag removed from main:
--- FAIL: TestMain_RunConfigure_BaseURL
    main_extra2_test.go:176: configure unexpectedly called osExit(1)
the help line alone removed:
--- FAIL: TestMain_RunConfigure_BaseURL
    main_extra2_test.go:189: help does not list --base-url:
```

The test's fixed password is replaced by `<password>` above.

**Step 4.** A scratch build of the tree ran README's Example 2,
`configure --yes --provider ollama --model qwen2.5-coder:7b`, with an
empty configuration under a scratch home directory: exit 0, `Active
provider: ollama (qwen2.5-coder:7b)`, saved with no key and no
`base_url`. The three fallbacks it saved came from the fallback listing
of an Ollama running on the test machine at the default endpoint.

**Phase gate,** each exit status captured before any filter:

| Check | Result |
| :--- | :--- |
| `gofmt -l` on the five changed `.go` files | 0, no output |
| `go test -race -count=1 ./...` | 0; 5 packages ok |
| `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | 0 each |
| `go mod tidy -diff` | 0, empty |
| `make verify` | 0, after the goconst fix above; `golangci-lint run`: `0 issues.`; total coverage 83.7% (minimum 80.0%); `govulncheck`: `No vulnerabilities found.` |
| `make verify-staged` | 0; `0 issues.`; `No vulnerabilities found.` |
| `make check-records`, `make markdownlint` | 0 each |

Staged for the owner's commit.

Committed by the owner as `7a2b7b4`. The owner then said "committed,
proceed".

### Phase 4: client info, unpacker, SDK test fakes (2026-10-13)

**Step 1 (D5).** `main.clientInfo()` returns
`llmprovider.WithClientInfo(AppTitle, buildIdentity().Current())`.
`newSourceProvider` passes it on every provider the hook builds.
`runConfigure` hands it to `internal/ui` as
`SetupOptions.ClientOptions`, which the wizard's `ProviderOptions` and
both `--yes` listings (`discoverModels`, now variadic in extra options)
carry. `internal/ui` imports nothing from `main`.

* **Noted, within the step:** the step named `clientInfo` an
  `llmprovider.Option` passed through `SetupOptions`. `llmprovider.Option`
  is a struct (`llmprovider/settings.go:19`), so a single field cannot be
  left unset; the field is a slice, `ClientOptions []llmprovider.Option`,
  and none leaves the SDK naming itself.

**Step 2.** `TestProvider_UserAgentNamesApp` builds the hook's provider
on an `httptest` server and requires a User-Agent beginning
`prepare-commit-msg/<build> (`.

* **Noted, within the step:** step 1 reaches three call sites and step 2
  tested one. Two more tests cover the rest:
  `TestSetup_UserAgentNamesApp` (`internal/ui`: the wizard's and `--yes`'s
  requests to Phase 3's Ollama fixture, which now records each request's
  User-Agent) and `TestMain_RunConfigure_UserAgentNamesApp` (`configure
  --yes --provider=ollama --base-url=<fixture>`, which shows that `main`
  passes the name).

**Step 3 (D8).** `update.go`: `embeddedSpec()` parses the spec and
requires it to name this program; `releaseAssets()` and the new
`releaseUnpacker()` (`spec.Unpacker()`) are built on it, and
`defaultNewUpdateUpdater` passes `Unpacker:` to `selfupdate.Config`.
`TestReleaseSpec` requires `releaseUnpacker()` to be nil for the
embedded spec.

* **Not tested, and why:** that `defaultNewUpdateUpdater` passes the
  unpacker it gets is visible only when an update runs against GitHub.
  `TestDefaultNewUpdateUpdater` still builds the updater, and the
  assertion covers the value it passes.

**Step 4.** `session.Store = store` is gone from `newActiveProvider`;
`auth.FileTokenStore.Load` sets it (`auth/tokenstore_file.go:86`). The
OAuth tests pass.

**Step 5 (D7).** `analyzerTestProvider` is gone; its three uses (two in
`main_oauth_test.go`, one in `provider_test.go`) are
`llmtest.NewFake(llmprovider.ProviderOpenAI,
llmprovider.Capabilities{}).ReplyText("feat: use OAuth session")`. A grep
of the Go sources for `analyzerTestProvider` and `session.Store` finds
nothing.

**Fail-first,** each on a scratch copy of the tree; the tree was not
touched:

```text
the hook without clientInfo():
--- FAIL: TestProvider_UserAgentNamesApp
    provider_test.go:233: User-Agent = "go-llmprovider-sdk/v1.12.0 (darwin; arm64) go-llmprovider-sdk/v1.12.0"; want it to begin "prepare-commit-msg/dev ("
main not passing ClientOptions:
--- FAIL: TestMain_RunConfigure_UserAgentNamesApp
    main_extra2_test.go:230: User-Agent = "go-llmprovider-sdk/v1.12.0 (darwin; arm64) go-llmprovider-sdk/v1.12.0"; want it to begin "prepare-commit-msg/dev ("
the wizard's ProviderOptions without them:
    --- FAIL: TestSetup_UserAgentNamesApp/wizard
        setup_test.go:980: User-Agent = "go-llmprovider-sdk/v1.12.0 (darwin; arm64) go-llmprovider-sdk/v1.12.0"; want it to name prepare-commit-msg/v0.0.0-test
discoverModels dropping them:
    --- FAIL: TestSetup_UserAgentNamesApp/--yes
        setup_test.go:980: (as above)
--- FAIL: TestMain_RunConfigure_UserAgentNamesApp
    main_extra2_test.go:230: (as above)
the embedded spec's packaging set to "archive" (a broken input):
--- FAIL: TestReleaseSpec
    releasespec_test.go:42: releaseUnpacker() = &{ 4096}, <nil>; want nil for "archive" packaging
```

**Phase gate,** each exit status captured before any filter:

| Check | Result |
| :--- | :--- |
| `gofmt -l` on the eight changed `.go` files | 0, no output |
| `go test -race -count=1 ./...` | 0; 5 packages ok |
| `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | 0 each |
| `go mod tidy -diff` | 0, empty |
| `make verify` | 0; `golangci-lint run`: `0 issues.`; total coverage 83.5% (minimum 80.0%); `govulncheck`: `No vulnerabilities found.` |
| `make verify-staged` | 0; `0 issues.`; `No vulnerabilities found.` |
| `make check-records`, `make markdownlint` | 0 each |

**Found, from Phase 2, not done.** MADR D3 says "`readLine`, `promptInt`
and the private `bufio.Reader` go", and Phase 2 step 1 says to delete
them. Phase 2 (`8df87c7`) deleted `readLine` and the reader but kept
`promptInt`, rewired to read through `wizard.Prompter.Input`, and its
record describes the rewiring without naming it a deviation. The
Verification grep for `promptInt` therefore still finds it
(`internal/ui/setup.go:85`). It waits for the owner's choice of
resolution.

Staged for the owner's commit.

Committed by the owner as `c8d935a`. The owner then said "committed,
proceed", without choosing a resolution for the finding above; it still
waits.

### Phase 5: documents (2026-10-13)

**Step 1, `README.md`:**

* The section "Changes with go-llmprovider-sdk v1.12.0 and
  go-selfupdate-lib v1.15.2" follows the `v1.13.0` one, and the table of
  contents names it. It covers D2 to D6 (all 10 providers; Ollama with no
  key in the hook and in `--yes`; `base_url` and `--base-url`; `--yes`
  stopping on a refused key; the six-model menu; typed operational values
  and the `(default N)` wording), D5's User-Agent, Amendment 1 A3's
  prompt-cache charge on `gpt-6-luna`, SDK 0028's user-visible behaviour,
  and that `update` is unchanged.
* "Configuration File": the example's `ollama` entry carries a `base_url`
  (`192.0.2.10`, an RFC 5737 documentation address), and a paragraph says
  when it is saved and which providers take one.
* The `--provider` row lists the registry's 10 IDs in registry order; a
  `--base-url` row follows `--api-key` (MADR Amendment 2, B5).
* **Noted, within the step:** the step names the section "Changes with
  go-llmprovider-sdk v1.11.0 and go-selfupdate-lib v1.15.2". Amendment 1
  moved the SDK to `v1.12.0` and the release ships it, so the heading
  names `v1.12.0`.

**Step 2, `docs/architecture.md`:** "Runtime flow" says the provider list,
the key rule and each default endpoint come from the SDK's registry, which
`internal/config` builds its list from, and that a saved `base_url`
replaces the default endpoint.

**Checked against the code, not the records:** v1.8.0's `ValidateActive`
(`703ea27:internal/config/config.go`) refused an empty key for every
provider with `no API key for provider %q`, and its `--provider` help
named `gemini|openai|claude|grok`; the SDK's `TextPrompter.Input` prints
`%s (default %s): ` (`wizard/text_prompter.go:292`); the registry's
`SupportsBaseURL` is true for exactly the six providers the README names.

**A new check, seen to fail:** the README's in-page links were checked
against its headings by a scratch script (`#` links resolved to GitHub's
heading slugs): none missing. On a scratch copy with the new table of
contents link pointing at `#changes-with-go-llmprovider-sdk-v1-12-0`, it
reported that anchor missing and exited 1. The script is not committed.

**Phase gate,** each exit status captured before any filter:

| Check | Result |
| :--- | :--- |
| `go test -race -count=1 ./...` | 0; 5 packages ok (`TestReadmeCuratedModels` among them) |
| `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | 0 each |
| `go mod tidy -diff` | 0, empty |
| `make verify` | 0; `golangci-lint run`: `0 issues.`; total coverage 83.5% (minimum 80.0%); `govulncheck`: `No vulnerabilities found.` |
| `make check-records`, `make markdownlint` | 0 each |

**Found, and done in this phase.** "3. Configure Provider & Model"
(`README.md:230`) listed nine providers, in the provider table's order
but without `together`; Phase 5 named the `--provider` row only. The
owner chose to fix it here ("follow recommendations", 2026-10-13): the
sentence names the registry's 10 IDs in registry order, as the
`--provider` row does. `README.md` was already in this phase.

### Deviation 2 (2026-10-13): `promptInt` stays

**Found** in Phase 4 (recorded there): MADR D3 and Phase 2 step 1 say
`promptInt` is deleted. Phase 2 (`8df87c7`) deleted `readLine` and the
`bufio.Reader` but kept `promptInt` (`internal/ui/setup.go:85`), reading
through `wizard.Prompter.Input`, and did not name that a deviation. The
Verification grep for `promptInt` would have failed at close-out.

**Why it stays [read].** `wizard.Prompter` has `Select`, `MultiSelect`,
`Confirm`, `Input`, `Secret` and `Notify`, and no integer prompt. The
string from `Input` has to be parsed, and an invalid value has to keep the
current one with the "keeping N" note, somewhere in this repository.
Inlining it into `promptOperational` would repeat the parse four times or
move it under another name, with no change in behaviour.

**Decision.** The owner chose to record it ("follow recommendations",
2026-10-13). Records only, no code:

* the MADR's D3 carries a dated annotation;
* Phase 2 step 1 strikes `promptInt` from the deletion;
* Verification drops `promptInt` from the grep and checks instead that
  `bufio` and `readLine` are gone from `internal/ui/setup.go`.

**Checked:** `grep -n 'bufio\|readLine' internal/ui/setup.go` finds
nothing (exit 1).

Phase 5, with the README line above and Deviation 2's records, is staged
for the owner's commit.

Committed by the owner as `adee10a`. The disclosure guard over the eight
outgoing commits (`703ea27` to `adee10a`) exited 0 with no output.

### Amendment 4 (2026-10-13): Phase 5a, the unpushed `main` on Windows and Linux

The owner asked that the unpushed work be tested on their Windows laptop
and on a Linux host, over `ssh`, before the push. The PLAN tested only on
this macOS host before the push, leaving Windows and Linux to CI. Phase 5a
is added between Phase 5 and Phase 6, and one Verification item with it.
Read-only probes on 2026-10-13: both hosts answer `ssh` with key-only
login; both have Go 1.27.2, `git`, `make` and `gcc`; the laptop's Ollama
answers on `localhost:11434` and lists `qwen2.5-coder:7b`, README's
Example 2 model. No decision in the MADR changes. The amendment waits for
the owner's approval.

Approved and committed by the owner as `f90c8f7` ("accepted, committed,
proceed").

### Phase 5a: the unpushed `main` on Windows and Linux (2026-10-13)

The hosts are as Amendment 4 describes them: the owner's Windows laptop
and a Linux host.

**Step 1.** `git bundle create` of `main`, copied by `scp` to
`~/tmp/pcm-0014-5a/` on each host and cloned there.

* **Noted, within the step:** the clones' `HEAD` is `f90c8f7`, not
  Phase 5's `adee10a`; the owner committed Amendment 4 first, and `git diff
  --stat adee10a f90c8f7` touches only this PLAN.

**Step 2,** each exit status captured on the host before any filter:

| Check | Windows | Linux |
| :--- | :--- | :--- |
| `go test -count=1 ./...` | 0; 5 packages ok | 0; 5 packages ok |
| `go test -race -count=1 ./...` | 0; 5 packages ok | 0; 5 packages ok |
| `go vet ./...` | 0 | 0 |
| `make verify` | not run (Linux only, as CI) | 0; `0 issues.`; total coverage 83.5% (minimum 80.0%); `No vulnerabilities found.` |

**Step 3.** Both scratch builds exit 0. Every live run set `APPDATA`,
`USERPROFILE` and `LOCALAPPDATA` (Windows), or `HOME` and
`XDG_CONFIG_HOME` (Linux), to a directory under the scratch directory.
The installed hook's SHA-256 was taken on each host before step 1.

**Step 4, both hosts:**

* `version`: `prepare-commit-msg version v0.0.0-20261014000753-f90c8f7afc9f
  (local) f90c8f7afc9f`; `help` lists the 10 IDs in registry order and
  `--base-url`.
* `--base-url http://user@127.0.0.1:11434` for Ollama: `Setup failed:
  --base-url carries no user name or password`, exit 1, nothing saved.
* `--provider gemini --base-url https://example.com`: the first Windows run
  had no key, so it was refused by the key check (`API key required`)
  before the endpoint was checked. **Noted, within the step:** rerun on
  both hosts with a made-up `GEMINI_API_KEY` in the environment and
  `--model` (so no listing, and nothing sent): `Setup failed: gemini takes
  no --base-url`, exit 1, nothing saved.
* Added within the step, on both: `--provider opencode-zen --base-url
  http://example.com/zen/v1` with a made-up `OPENCODE_API_KEY`: `Setup
  failed: --base-url http://example.com/zen/v1 is plain http to a remote
  host, so the key would travel unencrypted; use https`, exit 1, nothing
  saved.

**Step 5, Windows, against its Ollama:**

* README's Example 2 on an empty configuration: exit 0, `Active provider:
  ollama (qwen2.5-coder:7b)`, saved with `api_key` empty and no `base_url`.
* `--base-url http://127.0.0.1:11434` with no `--model`: exit 0; the live
  listing answered, and `base_url` is saved. Example 2 run again keeps that
  `base_url` and sets the model back (MADR Amendment 2, B2).
* The hook, on one staged Go file in a scratch repository: exit 0 in 38 s,
  `prepare-commit-msg: generating via ollama (qwen2.5-coder:7b)...`, and
  the message file reads `` feat(greet): add new `greet` function ``, with
  `# AI-generated (1 files: +4 -0)`. No key; no commit made.

**Step 6, Linux:** a made-up `GEMINI_API_KEY` in the environment,
`configure --yes --provider gemini`: `Setup failed: gemini refused the API
key: llmprovider: authentication failed: gemini HTTP 400 API_KEY_INVALID:
API key not valid. Please pass a valid API key.`, exit 1, nothing saved.

**Step 7.** `rm -rf ~/tmp/pcm-0014-5a` on each host; `test -e` then exits
1. Each installed hook's SHA-256 is the same as before step 1. The logs
were copied to the agent's scratch directory first; they are not
committed.

**Found, not caused by this work, waiting for the owner:**

* **F1. `--yes` can save an Ollama embedding model.** Example 2 saved
  `granite-embedding:30m` as the first fallback, and `--base-url` with no
  `--model` saved it as the model, because the laptop's Ollama lists it
  first. For Ollama, `catalog.Static` is nil and the SDK passes the
  installed models through as listed, in `v1.3.2` as in `v1.12.0`
  (`llmprovider/catalog/models_catalog.go:276`); `--yes` takes the first
  as the model when `--model` is absent and the next three as fallbacks
  when `--fallback` is absent. Before Amendment 2, `--yes` could not
  configure Ollama at all; the wizard shows the same list and lets the user
  choose. What the hook does when it reaches an embedding model was not
  tested.
* **F2. The README's sample message is not what the hook writes.**
  "Daily Workflow & Hook Behavior" shows `# Go: 2, JSON: 1`; the hook
  counts `.go` files as `Scripts` (`internal/git/git.go:98`) and wrote
  `# Scripts: 1`. That has been so since the code was imported; it is
  outside this PLAN.

Staged for the owner's commit.
