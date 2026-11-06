---
status: accepted
date: 2026-11-05
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its CHANGELOG, v1.23.0 to v1.26.1, and its source at v1.26.1); go-selfupdate-lib (its CHANGELOG, v1.22.1 to v1.24.0, its migration guide's sections 23 and 24, and the compare v1.22.0...v1.24.0 by git object); this repository's v1.10.2 release run; a scratch clone of this repository moved to the new pins, 2026-11-05
informed: users of the installed hook, who receive the change through `update`
---
# Move to go-llmprovider-sdk v1.26.1 and go-selfupdate-lib v1.24.0, with both release workflows at v1.24.0's commit

*Revised 2026-11-05, before execution. The owner answered Q1 and Q2:
"carry the workspace through, and release v1.10.3". Option A is chosen; D3
and D6 hold without condition.*

## Context and Problem Statement

`prepare-commit-msg` pins its two libraries in four places **[read]**:

| Pin | Where | Pinned | Newest |
| :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `go.mod:6` | `v1.22.2` | `v1.26.1` (2026-11-04) |
| go-selfupdate-lib | `go.mod:7` | `v1.22.0` | `v1.24.0` (2026-11-02) |
| go-selfupdate-lib build workflow | `.github/workflows/ci.yml:117` | `84805563…` (`v1.22.0`) | `74ab5dd376a4c864debb32b2d8abfc7856059ba3` (`v1.24.0`) |
| go-selfupdate-lib publish workflow | `.github/workflows/ci.yml:132` | `84805563…` (`v1.22.0`) | `74ab5dd3…` (`v1.24.0`) |

`docs/guides/cicd-operations.md:175` and `:179` quote both `uses:` lines,
and `TestCurrentDocumentation` (`documentation_test.go`) fails when the
guide and `ci.yml` disagree
(`0021-MADR-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md`,
"What the move changes") **[read]**.

The owner asked that both libraries' newer tagged releases be pinned, for
their fixes and enhancements. What has to be decided is whether both
modules move to their newest releases, where the two workflow pins may
move under the rule that governs the publish pin, what is done about the
one wizard addition this app would otherwise discard, and what the user is
told.

Every claim below is marked:

* **[read]**: from this repository, or a library's changelog, source or
  git objects at the version named;
* **[probed]**: run on a clone of this repository at `6337fa0` in the
  session scratchpad, never in this tree, on 2026-11-05;
* **[derived]**: reasoned from the two, and not run.

### What the move changes in the module graph

* `go get …go-llmprovider-sdk@v1.26.1 …go-selfupdate-lib@v1.24.0` and
  `go mod tidy` change five `require` lines of `go.mod` and replace ten
  lines of `go.sum` **[probed]**: the two modules, and `golang.org/x/mod` v0.40.0
  → v0.42.0, `x/sys` v0.47.0 → v0.49.0, `x/term` v0.43.0 → v0.47.0, all
  `// indirect`.
* The `x/*` moves come from go-selfupdate-lib `v1.23.1`, which requires
  them; its changelog: "No exported identifier and no behaviour a program
  can see changes" **[read]**. go-llmprovider-sdk `v1.26.1` still requires
  `go 1.27.3` and `x/term v0.43.0`; go-selfupdate-lib `v1.24.0` requires
  `go 1.26.10` and `toolchain go1.27.3` **[read]**. No Go pin moves.
* With the modules moved and both workflow lines repinned in `ci.yml` and
  the operations guide, `make verify` exits 0: mod-check, fmt-check,
  golangci-lint, `go vet`, `go test -race ./...`, total coverage 83.7%
  (minimum 80.0%), govulncheck ("No vulnerabilities found."),
  workflow-lint ("ci filters: 11 patterns, 11 reads checked, 0 problems"),
  the tools' 25 tests, the scripts' 34 tests, and the six release
  cross-builds **[probed]**. `go mod tidy -diff` is empty **[probed]**.

### go-llmprovider-sdk, v1.22.2 to v1.26.1

From its CHANGELOG **[read]**. Every release is additive to the exported
API.

* **v1.23.0:** `providers/chatcompat` (any Chat Completions server, under a
  caller's id) and `chatcompat.Discover`; `catalog.Catalog.Search` over a
  model's facts; `ErrEmptyAnswer`, `ErrOffline`, `EffortMax`, reasoning
  refusals and `ImageItem`. Changed: the wizard's model search searches
  facts and shows ten at a time with a "More" row; a request's reasoning
  effort and thinking shape follow the model's facts and learned refusals,
  not rules of model names; a request that could not be sent for want of a
  network matches `ErrProviderUnavailable`, through `ErrOffline`, on every
  send path.
* **v1.24.0 / v1.24.1:** `claude.WithWorkspace`,
  `catalog.WithClaudeWorkspace`, `ErrWorkspaceRequired`,
  `ErrWorkspaceNotFound`, and `wizard.Result.Workspace`: "the wizard asks a
  Claude key that needs a workspace for one, up to three answers, keeps the
  one that lists". `v1.24.1` is a release fix with `v1.24.0`'s code.
* **v1.25.0:** `catalog.Lister`, `wizard.Options.LocalServers`,
  `llmprovider.NewHTTPClient`. Changed: a saved session the listing
  refuses no longer ends the run, the sign-in menu follows. Fixed: a model
  id holding a terminal control character no longer reaches a menu, the
  terminal or a `Result`; the wizard's notices show no service's text and
  no key a URL carries; "offline" is called so.
* **v1.26.0:** the shared transport pings an idle HTTP/2 connection after
  30 s and closes it after 15 s unanswered, so a dead pooled connection
  fails its next request within 45 s, where it held it for the 300 s
  first-byte bound; `WithRetry` no longer retries HTTP 501 or 505, or a
  refused TLS certificate; `FileTokenStore`'s refresh lock waits 60 s,
  where it waited 40 s and could give up on a holder within its bounds.
* **v1.26.1:** an event the stream ends inside is no longer read; a
  stream cut short is `ErrProviderUnavailable`, retried at most once,
  where it was a decode error never retried.

**What reaches this app** **[read]**:

* `configure` calls `wizard.ConfigureLLM` with a `TextPrompter`,
  `Discover`, `ModelLimit`, `Registry: providers.Default()`, `TokenStore`
  and `NeedFallbacks`, and not `LocalServers`, `Providers` or
  `NoEndpointPrompt` (`internal/ui/setup.go:269-282`). So it gains the
  model-search change, the control-character filter, the cleaned notices
  and the refused-saved-session change; `chatcompat`, `LocalServers` and
  the opt-in options do not apply **[derived]**.
* `listingClient` is nil outside tests (`internal/ui/setup.go:34-36`) and
  `newSourceProvider` passes no client (`main.go:370-374`), so listing and
  generation use the SDK's shared transport, and v1.26.0's HTTP/2 ping
  applies to both **[derived]**.
* `generateText` wraps the provider in `WithRetryNotify`
  (`main.go:421-426`), so v1.26.0's 501/505 and TLS changes apply: such a
  failure is reported at once instead of after the configured retries
  **[derived]**. `retryKind` tests `ErrProviderUnavailable` before
  `*url.Error` (`main.go:404-410`), so an offline retry is now labelled
  "unavailable" where it was "cannot reach"; it is retried as before
  **[derived]**.
* The app calls `Generate`, never `Stream`, and passes no `Reasoning`
  **[read]**. `Generate` reads an event stream on one path: OpenAI's
  ChatGPT session, `decoder.ReadStream` at the SDK's
  `llmprovider/providers/openai/openai.go:502` **[read]**. So v1.26.1's fix
  reaches a ChatGPT-session user only **[derived]**.
* `auth.FileTokenStore` holds OAuth sessions (`internal/config/config.go:114-119`),
  so v1.26.0's lock wait applies **[read]**.

### The wizard's workspace, which this app would discard

* At `v1.26.1`, `wizard.ConfigureLLM` asks for an Anthropic workspace when
  the listing refuses a Claude key with `ErrWorkspaceRequired` or
  `ErrWorkspaceNotFound`, and returns it as `Result.Workspace`, whose doc
  says "pass it to the provider with claude.WithWorkspace"
  (the SDK's `wizard/configure.go:56-60`, `:416-436`) **[read]**.
* This app saves no workspace: `config.ProviderConfig` has no such field
  (`internal/config/config.go:64-81`), `RunSetupWithOptions` copies
  `Model`, `Fallbacks`, `Organization`, `BaseURL` and the credential only
  (`internal/ui/setup.go:291-325`), and `providerOptions` passes no
  `claude.WithWorkspace` (`main.go:353-363`) **[read]**.
* So, with the move alone, a user whose Claude key is not scoped to a
  workspace is asked for one by `configure`, which then accepts the key and
  saves it without it; the hook's generation then fails with
  `ErrWorkspaceRequired` **[derived]**. What `v1.22.2`'s wizard did with
  such a key was not probed. A key scoped to its workspace, which needs
  none, is not affected **[derived]**.

### go-selfupdate-lib, v1.22.0 to v1.24.0

From its CHANGELOG and migration guide **[read]**. Every exported change is
an addition, which its `make apicheck` holds.

* **v1.22.1, v1.22.2:** `service`, systemd, launchd and Task Scheduler
  stop fixes.
* **v1.22.3:** its own tests only.
* **v1.23.0:** `Updater.Plan` / `Apply` (pinned update plans) and the
  handoff that carries one; `servicetest`. Changed: "the install resolves
  every name through the `os.Root` of the directory it locked … A
  directory swapped in after the lock is never written to"; on Windows the
  rename is unchanged. "`Run`'s events, call order and output are as in
  `v1.22.3`."
* **v1.23.1:** the `x/*` requirements above; no behaviour change.
* **v1.24.0:** `fmt` and `log/slog` redact the `Env` of four `service` and
  `ghattest` option types.

**What reaches this app:** it imports `buildinfo`, `selfupdate`,
`selfupdate/cli`, `selfupdate/releasespec` and
`selfupdate/selfupdatetest`, and no `service` package **[read]**. `update`
runs `cli.Command` with `cli.StdioOptions` and no `HandOff`
(`main.go:107`, `update.go:68`), so `cli.Run` takes its `RunWith` path,
not the new plan path (`selfupdate/cli/run.go` at `v1.24.0`) **[read]**.
The change it sees is v1.23.0's `os.Root` install on Unix **[derived]**.
`NewGitHubSource` and `DefaultHTTPClient`, which `update.go:57-73` calls,
changed only by moving test seams **[read]**; the `--json` schema test
passes **[probed]**.

### The workflow pins, and the rule that governs the publish pin

`0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` Amendment A1: "the publish
pin is the newest go-selfupdate-lib commit whose publish path, meaning the
publish workflow file and the scripts it runs from the pinned checkout, is
unchanged from a commit that has published live." The owner ruled in
`0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md`
A1 that `go.mod` is not part of that path.

`v1.24.0` peels to `74ab5dd376a4c864debb32b2d8abfc7856059ba3` **[read]**.
By git object **[read]**:

| Path | `8480556` (`v1.22.0`) | `74ab5dd` (`v1.24.0`) |
| :--- | :--- | :--- |
| `.github/workflows/publish-selfupdate-release.yml` | `5048ad0c` | `5048ad0c` |
| `.github/workflows/build-selfupdate-release.yml` | `56f95cd5` | `56f95cd5` |
| `internal/cmd/selfupdate-release/` | `20594a2d` | `20594a2d` |
| `internal/cmd/selfupdate-note/` | `804438fe` | `804438fe` |
| `scripts/` | `829c22fd` | `619882b8` |
| `go.mod` | `24d7b31e` | `0d6f1d93` |

`scripts/` differs, unlike in 0021. The publish workflow runs
`check-release-tag.sh`, `verify-selfupdate-release.sh`,
`refuse-existing-release.sh` and `release-latest-flag.sh` from it, and the
build workflow the first two **[read]**. The compare changes
`check-docs.sh`, `gate.sh` and `check_records.py` and three tests in
`scripts/`, none of the four **[read]**. `go.mod` differs by v1.23.1's
`x/*` requirements, with which the two `internal/cmd/` programs the publish
workflow `go run`s are built; 0018-MADR A1 holds that outside the path
**[read]**.

`84805563…` has published live: this repository's `v1.10.2`, run
`40150318042`, success, `sign` unset
(`0021-PLAN-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md`,
Phase 5) **[read]**. So `74ab5dd3…`'s publish path is unchanged from a
commit that has published live, and A1 makes it the publish pin with no
exception. The installer templates come from `internal/cmd/` and are the
same objects, so the released `install.sh` and `install.ps1` do not change
**[derived]**.

## Decision Drivers

* **Pins are current** unless a record names a reason to hold one.
* **The publish pin follows 0010-MADR A1:** proven by a live publish.
* **No configure flow that accepts a value and drops it.** A prompt the
  user answers must have an effect.
* **The change is a pin move,** plus what the owner adopts by name. Other
  new library features are separate decisions.
* **The README says what a user can see.**

## Considered Options

Every option moves both modules to their newest releases, and both
workflows to `74ab5dd3…`. They differ in the wizard's workspace.

* **A. Move the pins and carry the workspace through:** `configure` saves
  `Result.Workspace`, and generation and the `--yes` listing pass it.
* **B. Move the pins only,** and record the dropped workspace as a known
  gap for a later record.
* **C. Hold go-llmprovider-sdk at `v1.23.0`,** the last release without
  the workspace prompt, and move go-selfupdate-lib alone.

## Decision Outcome

Chosen option: "A", by the owner's answer to Q1, because the move
otherwise introduces a `configure` prompt whose answer is discarded and a
saved configuration that cannot generate, and carrying one string through
is the smallest change that keeps the pins current.

### D1. Both modules move

* `go.mod` requires `go-llmprovider-sdk v1.26.1` and
  `go-selfupdate-lib v1.24.0`; `go mod tidy`. `go.mod` and `go.sum`
  change on those two modules' lines and `golang.org/x/{mod,sys,term}`'s,
  and no others.

### D2. Both workflow pins move

* Both `uses:` lines pin
  `74ab5dd376a4c864debb32b2d8abfc7856059ba3 # v1.24.0`.
* The comment above the build job is rewritten: both are pinned to
  `v1.24.0`'s commit, the version `go.mod` requires; the publish path is
  unchanged from `v1.22.0`'s, which published live as this repository's
  `v1.10.2` (run `40150318042`); it cites this record and 0010-MADR D2, A1.
* `docs/guides/cicd-operations.md:175` and `:179` quote the new lines, in
  the same commit.
* The `release` job's `with:` and permissions are unchanged.

### D3. The wizard's workspace is saved and used (Q1)

* `config.ProviderConfig` gains `Workspace string`
  (`json:"workspace,omitempty"`). An existing `config.json` has none and
  loads as before **[derived]**.
* `RunSetupWithOptions` passes a saved Claude workspace as
  `existing.Workspace`, and saves `res.Workspace` into `pc.Workspace`. The
  `--yes` path keeps a saved workspace, as it keeps a saved `BaseURL`
  (`internal/ui/setup.go:431-443` clears only the credential fields and
  `Organization`) **[read]**.
* `providerOptions` adds `claude.WithWorkspace(pc.Workspace)` for
  `ProviderClaude` when it is not empty, as it adds Kilo's organization.
* The `--yes` path's `discoverModels` passes
  `catalog.WithClaudeWorkspace` when a workspace is given. A `--workspace`
  flag is **not** added here; `--yes` lists with a saved workspace only.
* Tests: a wizard result with a workspace is saved and reloaded; a saved
  workspace reaches `providers.New` as an option; each is seen to fail on
  a scratch copy without the change.
* It lands in its own commit, after D1.

### D4. No other library feature is adopted here

* Not adopted: `chatcompat`, `wizard.Options.LocalServers`,
  `catalog.Lister`, `NewHTTPClient`, `EffortMax` and reasoning refusals,
  `ImageItem`, `Updater.Plan` / `Apply`, and every `service` addition.

### D5. The README states what changed

* A section, "Changes with go-llmprovider-sdk v1.26.1 and
  go-selfupdate-lib v1.24.0", in the pattern of the one before it, and its
  table-of-contents line. For `configure`:
  * model search matches a family or a feature as well as a name, and
    shows ten matches at a time;
  * an Anthropic key that needs a workspace is asked for one, which is
    saved and used (D3);
  * a saved sign-in the model listing refuses no longer ends the run; the
    sign-in menu follows;
  * a model name holding terminal control characters is not shown, and
    notices no longer quote a service's reply.
* For the hook: a dead connection after sleep or a network change fails
  within 45 seconds, where it could wait 5 minutes; an HTTP 501 or 505, or
  a refused TLS certificate, is reported at once, not retried; an offline
  retry is labelled "unavailable".
* For `update`: on macOS and Linux the binary is replaced only in the
  directory it locked. The installers behave as in `v1.10.2`.

### D6. The release (Q2)

* `v1.10.3`, a patch, by the owner's answer to Q2: what reaches the app
  are fixes, and D3 completes a flow the move brings in. The owner tags after
  the change is on `main` and CI there is green, and asks for the tag in
  the turn it is made.

### Consequences

* Good, because both modules and both workflows are current, and the
  publish workflow is at a commit whose publish path has published live.
* Good, because the hook recovers from a dead pooled connection in 45 s,
  and stops retrying failures that cannot succeed.
* Good, because a Claude key that needs a workspace can be configured end
  to end (D3).
* Neutral, because the installers, `update`'s output and the hook's
  message format are as before.
* Neutral, because three `golang.org/x` indirect requirements move with
  go-selfupdate-lib, with no identifier this app uses changed.
* Bad, because D3 adds a configuration field, which a downgrade to
  `v1.10.2` ignores; that release then generates without the workspace
  **[derived]**.
* Bad, because
  `0023-PLAN-split-user-documentation-into-guides.md` Phase 1 step 6 moves
  "the seven 'Changes with …'" sections; D5 makes them eight, and that
  PLAN is amended when it is executed.

### Confirmation

* `go.mod` and `go.sum` change only on the lines D1 names; `go mod tidy
  -diff` is empty.
* `grep -c` gives `74ab5dd3…` 2 and `84805563…` 0 in both `ci.yml` and
  the operations guide; `ci.yml` passes workflow-lint.
* D3's tests pass, and were seen to fail on a scratch copy without D3.
* `make verify`, `make preflight` and `make preflight-hosts` pass on the
  committed `HEAD`.
* CI is green on `main`, the build rehearsal included.
* D6: the tagged `v1.10.3` publishes the six binaries,
  `SHA256SUMS` and both installers through `74ab5dd3…`, its first publish;
  `update` on an installed `v1.10.2` moves to it; `version` names it.

## Pros and Cons of the Options

### A. Move the pins and carry the workspace through

* Good, because every prompt `configure` shows has an effect.
* Good, because the pins are current.
* Bad, because it changes Go source and the config schema, where 0021's
  move changed one helper.

### B. Move the pins only

* Good, because the change stays a pin move and the README.
* Bad, because a user with an unscoped Claude key is asked for a
  workspace, told the configuration is saved, and then every commit fails
  **[derived]**.

### C. Hold go-llmprovider-sdk at `v1.23.0`

* Good, because there is no workspace prompt to handle.
* Bad, because it forgoes v1.25.0's terminal-control fix and v1.26.0's
  connection and retry fixes, and leaves a known reason to hold a pin that
  the next move must clear anyway.

## Owner questions

* **Q1 (D3).** Carry the wizard's workspace through in this move
  (Option A, recommended), or move the pins only and record the gap
  (Option B)?
* **Q2 (D6).** Release `v1.10.3` once the change is on `main` and CI is
  green (recommended), or leave the move unreleased? A tag is asked for in
  the turn it is made.

**Answered 2026-11-05.** "carry the workspace through, and release
v1.10.3". Q1: Option A, as recommended; D3 records it. Q2: release
`v1.10.3`, as recommended; D6 records it. The workflow pins were not
questioned and stand.

## Amendments

### A1 (2026-11-05): D5 reaches the reference documentation

*A1 widens D5, "The README states what changed", which is left above as
written. Found after PLAN Phase 5; the owner asked "Update the docs".*

* **The gap.** D5 put the workspace in the release notes only. The
  README's reference sections describe each saved field and variable a
  user meets, and `docs/architecture.md` describes `base_url`'s effect;
  none names `workspace` or `ANTHROPIC_WORKSPACE_ID` **[read]**.
* **What reads `ANTHROPIC_WORKSPACE_ID`.** Only the wizard: it is the
  first workspace prompt's default under `Options.AllowEnv` (the SDK's
  `wizard/configure.go:636-655`), which `configure` sets unless `--no-env`
  (`internal/ui/setup.go:273-274`); `config.LookupEnv` passes it through
  unchanged. Neither the hook nor the Claude provider reads it **[read]**.
* **The decision.** The current-system documentation names the workspace
  where it names its neighbours:
  * README "Interactive Wizard": step 3 says an Anthropic key not scoped
    to a workspace is asked for its workspace ID;
  * README "LLM Provider Credentials": a row for `ANTHROPIC_WORKSPACE_ID`,
    read by `configure` only, as the prompt's default, and not under
    `--no-env`;
  * README "Non-Interactive CLI Flags": `--yes` keeps and lists with a
    saved workspace, and has no flag to set one;
  * README "Configuration File": the `claude` example carries a
    placeholder `workspace`, and a paragraph after `base_url`'s says what
    it is, when it is present and where it is sent;
  * `docs/architecture.md` "Runtime flow": a saved Claude `workspace` is
    sent with that provider's requests.
* **Not changed.** The release-notes section, which already says it; the
  code; the earlier README "Changes with" sections.
* The PLAN records it as Phase 6.

## More Information

* The libraries' records, cited by repository and filename:
  go-llmprovider-sdk `0061-MADR-consumer-gaps-from-gobble-issue-9.md`,
  `0062-MADR-learn-thinking-parameters-from-refusals.md`,
  `0064-MADR-anthropic-workspace-for-unscoped-keys.md`,
  `0066-MADR-wizard-v1-24-debugging-pass-findings.md`,
  `0068-MADR-five-internal-fixes-from-the-0054-audit.md` and
  `0069-MADR-discard-an-sse-event-the-stream-ends-inside.md`;
  go-selfupdate-lib
  `0037-MADR-honour-contexts-in-backend-fakes-and-stop-commands.md`,
  `0038-MADR-pinned-plans-lifecycle-conformance-and-bubble-tests.md`,
  `0039-MADR-refresh-golang-x-requirements.md` and
  `0040-MADR-close-four-open-items-after-v1-23-1.md`.
* The pin rule:
  [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md)
  D2 and Amendment A1; its reading:
  [0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md](0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md)
  A1; the last move:
  [0021-MADR-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md](0021-MADR-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md).
* The implementation:
  [0024-PLAN-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md](0024-PLAN-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md).
