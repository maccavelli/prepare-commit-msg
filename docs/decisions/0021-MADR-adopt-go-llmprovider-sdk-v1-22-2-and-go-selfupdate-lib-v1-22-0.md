---
status: accepted
date: 2026-10-24
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its CHANGELOG, v1.22.0 to v1.22.2, and the compare v1.21.0...v1.22.2); go-selfupdate-lib (its CHANGELOG, v1.18.0 to v1.22.0, its migration guide's section 18, `DefaultHTTPClient` and `cli.Run` at v1.22.0, and the compare v1.17.2...v1.22.0 by git object); this repository's v1.10.1 release run; a scratch clone of this repository moved to the new pins, 2026-10-24
informed: users of the installed hook, who receive the change through `update`
---
# Move to go-llmprovider-sdk v1.22.2 and go-selfupdate-lib v1.22.0, with both release workflows at v1.22.0's commit and `update` on the library's default HTTP client

*Revised 2026-10-24, before execution. The owner answered Q1 "adopt now"
and Q2 "yes, we will release v1.10.2". D3 no longer holds back
`selfupdate.DefaultHTTPClient`; D6 adopts it, with the context subsection
"`update`'s HTTP client" that it rests on; D4 and D5 follow. The filename
keeps the first slug.*

## Context and Problem Statement

`prepare-commit-msg` pins its two libraries in four places **[read]**:

| Pin | Where | Pinned | Newest |
| :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `go.mod:6` | `v1.21.0` | `v1.22.2` (2026-10-24) |
| go-selfupdate-lib | `go.mod:7` | `v1.17.2` | `v1.22.0` (2026-10-24) |
| go-selfupdate-lib build workflow | `.github/workflows/ci.yml:106` | `036289ec…` (`v1.17.2`) | `84805563ed252df384f59ced45a3daf529702ab4` (`v1.22.0`) |
| go-selfupdate-lib publish workflow | `.github/workflows/ci.yml:121` | `036289ec…` (`v1.17.2`) | `84805563…` (`v1.22.0`) |

`docs/guides/cicd-operations.md:162` and `:166` quote both `uses:` lines,
and `TestCurrentDocumentation` (`documentation_test.go`) fails when the
guide and `ci.yml` disagree **[read]**.

The owner asked that the libraries' new releases be pinned. What has to be
decided is whether both modules move to their newest releases, where the
two workflow pins may move under the rule that governs the publish pin,
and what the user is told.

Every claim below is marked:

* **[read]**: from this repository, or a library's changelog, source or
  git objects at the version named;
* **[probed]**: run on a clone of this repository at `5d28b24` in the
  session scratchpad, never in this tree, on 2026-10-24;
* **[derived]**: reasoned from the two, and not run.

### What the move changes in the module graph

* `go get …go-llmprovider-sdk@v1.22.2 …go-selfupdate-lib@v1.22.0` and
  `go mod tidy` change two `require` lines of `go.mod` and four lines of
  `go.sum`, the two modules' own **[probed]**.
* go-llmprovider-sdk `v1.22.2` requires `go 1.27.3` and `golang.org/x/term
  v0.43.0`; go-selfupdate-lib `v1.22.0` requires `go 1.26.10`, `toolchain
  go1.27.3` and `x/mod v0.40.0`, `x/sys v0.47.0`, `x/term v0.43.0`. Those
  are the versions this module already requires **[read]**. No Go pin and
  no other module moves.
* With the modules moved and both workflow lines repinned in `ci.yml` and
  the operations guide, `make verify` exits 0: mod-check, fmt-check,
  golangci-lint, `go vet`, `go test -race ./...`, coverage 83.6% (minimum
  80.0%), govulncheck ("No vulnerabilities found."), workflow-lint, the
  tools' 25 tests, and the six release cross-builds **[probed]**.
* With `ci.yml` repinned and the guide not, `TestCurrentDocumentation`
  fails: "operations guide does not cite workflow reference
  `…build-selfupdate-release.yml@84805563… # v1.22.0`" **[probed]**. The
  guide moves in the same commit as `ci.yml`.

### go-llmprovider-sdk, v1.21.0 to v1.22.2

From its CHANGELOG **[read]**. Every release is additive to the exported
API, which its `make api-check` holds. Between the two tags, every non-test
Go file that changed is in `wizard/` **[read]**.

* **v1.22.0:** `wizard.Options.ShowDeviceCode` / `wizard.DeviceCode`,
  `Options.NoPastePrompt` and `wizard.InputerContext`, all opt-in. Changed:
  an empty answer to a browser sign-in's paste prompt is asked again,
  where it ended the sign-in with "oauth: empty authorization code"; end
  of input still ends it.
* **v1.22.1:** fixes from its
  `0059-MADR-wizard-debugging-pass-findings.md`. A typed key is no longer
  warned about for a shape the provider accepts (W1, W3); a sign-in the
  listing refuses no longer replaces the saved session in
  `Options.TokenStore` (W2); a pasted key no longer gains part of the
  terminal's paste marker (W4); a `Prompter` that answers empty at once no
  longer spins the paste prompt (W5); a saved session that cannot be used
  is not offered (W6); a saved fallback model the menu does not recommend
  is offered, preselected, instead of dropped (W7); `ConfigureLLM` on an
  ended context returns at once (W8); under `Options.NoEndpointPrompt` a
  held plain-http endpoint is used with a warning (W9); a pasted ChatGPT
  token's session keeps `Options.HTTPClient` (W11).
* **v1.22.2:** a refused Anthropic key's hint names only the prefix and a
  tag, and is shown only for a key without them (W14).

**What reaches this app:** `internal/ui/setup.go:269` calls
`wizard.ConfigureLLM` with a `TextPrompter`, `Discover: true`,
`NeedFallbacks: true`, `TokenStore`, `HTTPClient` and `AgentName`, and not
`NoEndpointPrompt` **[read]**. So `configure` gains W1, W2, W3, W4, W6,
W7, W8, W11, W14 and v1.22.0's paste-prompt change; W9 and the opt-in
options do not apply **[derived]**. Every test passes **[probed]**.
Nothing outside `configure` imports `wizard` **[read]**.

### go-selfupdate-lib, v1.17.2 to v1.22.0

From its CHANGELOG **[read]**. Every exported change is an addition, which
its `make apicheck` holds.

* **v1.18.0:** `selfupdate.DefaultHTTPClient`; `launchd.Job.Loaded` and
  `Status`. `NewGitHubSource`'s error for a nil client names
  `DefaultHTTPClient()`. The migration guide's `update.go` now passes
  `DefaultHTTPClient()`; "the guide passed `&http.Client{Timeout: …}`,
  which the field's documentation advises against."
* **v1.19.0:** Task Scheduler registration, status and one-shot handoff.
* **v1.20.0:** `service.ManagerEnvNames` / `ScrubManagerEnv`; the service
  handoffs no longer pass on the caller's unit's own variables.
* **v1.21.0:** `taskscheduler.Task.Enable` / `Disable`.
* **v1.22.0:** `service.NewSupervised` and `ErrNoSupervisor`.

**What reaches this app:** it imports `buildinfo`, `selfupdate`,
`selfupdate/cli`, `selfupdate/releasespec` and `selfupdate/selfupdatetest`,
and no `service` package **[read]**. In those, `github.go`,
`githubupdater.go` and `types.go` changed, by 1, 20 and 2 lines **[read]**.
`update.go:67-72` passes its own non-nil `&http.Client{Timeout:
updateTimeout}` to `NewGitHubSource` **[read]**, so the nil-client message
does not reach it, and with that client unchanged `update` behaves as
before **[derived]**; the `--json` schema test passes **[probed]**.

### `update`'s HTTP client

* **Today.** `update.go:54-55`: `updateTimeout = 15 * time.Minute`,
  "bounds each GitHub request. cli.Run bounds the whole run", passed as
  `&http.Client{Timeout: updateTimeout}` (`update.go:69`) **[read]**.
  `main.go:107` runs `cli.Command` with `cli.StdioOptions`, which sets no
  `Timeout`, so `cli.Run` bounds the run with `cli.DefaultTimeout`, 15
  minutes (`selfupdate/cli/run.go:21-22`, `:133-137`) **[read]**.
* **What the library advises.** `GitHubOptions.Client`'s doc at `v1.22.0`:
  "Client.Timeout bounds each request including the whole asset download;
  prefer a context deadline, or a Transport with ResponseHeaderTimeout, so
  a slow large download is not cut off." Its migration guide, section 18:
  under `&http.Client{Timeout: 15 * time.Minute}` "a request that stalls
  before its headers waits the whole 15 minutes"; "pass
  `selfupdate.DefaultHTTPClient()` and bound the run with a context
  instead; `cli.Run` already does" **[read]**.
* **`DefaultHTTPClient()`** returns a new client on each call, with a
  clone of `http.DefaultTransport` and `ResponseHeaderTimeout` 30 s, and no
  `Client.Timeout`; its doc gives 30 s to dial and 10 s for TLS
  (`selfupdate/githubupdater.go:104-125`) **[read]**.
* **So, with it:** a GitHub request that gets no response headers fails
  after 30 s, where it waited up to the run's 15 minutes; the whole run is
  still bounded at 15 minutes by `cli.Run`; a download is bounded only by
  that run, not by a per-request cap **[derived]**.
* **The probe.** With `update.go` calling `selfupdate.DefaultHTTPClient()`
  through a helper and `updateTimeout` removed, `make verify` exits 0
  **[probed]**. A test asserting that the helper's client has no
  `Client.Timeout` and a positive `ResponseHeaderTimeout` passes on it,
  and fails on a copy with the old client: "Client.Timeout = 15m0s, want
  0" **[probed]**.

### The workflow pins, and the rule that governs the publish pin

`0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` Amendment A1: "the publish
pin is the newest go-selfupdate-lib commit whose publish path, meaning the
publish workflow file and the scripts it runs from the pinned checkout, is
unchanged from a commit that has published live." The owner ruled in
`0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md`
A1 that `go.mod` is not part of that path.

`v1.22.0` peels to `84805563ed252df384f59ced45a3daf529702ab4` **[read]**.
By git object **[read]**:

| Path | `036289e` (`v1.17.2`) | `8480556` (`v1.22.0`) |
| :--- | :--- | :--- |
| `.github/workflows/publish-selfupdate-release.yml` | `5048ad0c` | `5048ad0c` |
| `.github/workflows/build-selfupdate-release.yml` | `56f95cd5` | `56f95cd5` |
| `scripts/` | `829c22fd` | `829c22fd` |
| `internal/cmd/selfupdate-release/` | `20594a2d` | `20594a2d` |
| `internal/cmd/selfupdate-note/` | `804438fe` | `804438fe` |
| `go.mod` | `24d7b31e` | `24d7b31e` |

The compare between the two commits touches none of `scripts/`,
`internal/cmd/` or the two reusable workflows; its only workflow change is
the library's own `ci.yml` **[read]**.

`036289ec…` has published live: this repository's `v1.10.1`, run
`39709052042`, success, `sign` unset
(`0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md`
A3) **[read]**. So `84805563…`'s publish path is unchanged from a commit
that has published live, and A1 makes it the publish pin with no
exception. The build workflow, the installer templates it stages and its
plan step are the same objects, so the released `install.sh` and
`install.ps1` do not change **[derived]**.

## Decision Drivers

* **Pins are current** unless a record names a reason to hold one.
* **The publish pin follows 0010-MADR A1:** proven by a live publish.
* **No silent change for a user who changes nothing.** The README says
  what a user can see.
* **The change is a pin move,** plus what the owner adopts by name (D6).
  Other new library features are separate decisions.
* **Follow the library's advice** for the calls this app already makes.
* **No commit on `main`** (the organization's Git rule). *Superseded by
  Amendment A1 (2026-10-24).*

## Considered Options

Both options move both modules to their newest releases.

* **A. Both workflows at `v1.22.0`'s commit,** `84805563…`.
* **B. Both workflows stay at `v1.17.2`'s commit,** `036289ec…`.

## Decision Outcome

Chosen option: "A", because the build and publish paths are the same git
objects at both commits, so 0010-MADR A1 is met by the live `v1.10.1`
publish, and one version is then pinned in `go.mod` and both workflows.

### D1. Both modules move

* `go.mod` requires `go-llmprovider-sdk v1.22.2` and
  `go-selfupdate-lib v1.22.0`; `go mod tidy`. Only those two modules'
  lines change in `go.mod` and `go.sum`.

### D2. Both workflow pins move

* Both `uses:` lines pin
  `84805563ed252df384f59ced45a3daf529702ab4 # v1.22.0`.
* The comment above the build job is rewritten: both are pinned to
  `v1.22.0`'s commit, the version `go.mod` requires; the publish path is
  unchanged from `v1.17.2`'s, which published live as this repository's
  `v1.10.1` (run `39709052042`); it cites this record and 0010-MADR D2, A1.
* `docs/guides/cicd-operations.md:162` and `:166` quote the new lines, in
  the same commit.
* The `release` job's `with:` and permissions are unchanged.

### D3. No other library feature is adopted here

* Not adopted: `wizard.Options.ShowDeviceCode`, `NoPastePrompt`,
  `InputerContext`; every `service`, `launchd` and `taskscheduler`
  addition.
* `selfupdate.DefaultHTTPClient` is adopted, by the owner's answer to Q1
  (D6).

### D4. The README states what changed

* A section, "Changes with go-llmprovider-sdk v1.22.2 and
  go-selfupdate-lib v1.22.0", in the pattern of the one before it, and its
  table-of-contents line. It says, for `configure`:
  * a key the provider accepts is no longer warned about for its shape;
    the expected shape is named only after a refusal;
  * a sign-in the model listing refuses no longer replaces the saved
    session;
  * a saved session that cannot be used is not offered;
  * a saved fallback model the menu does not recommend is kept,
    preselected;
  * a pasted key no longer picks up stray paste-marker characters;
  * an empty answer to the browser sign-in's paste prompt is asked again.
* For `update` (D6): a GitHub request that gets no response fails after
  30 seconds, where it could wait 15 minutes; a whole run is still limited
  to 15 minutes; a slow download is no longer cut off by a per-request
  limit.
* The installers behave as in `v1.10.1`.

### D5. The release (Q2)

* `v1.10.2`, a patch: what reaches the app are fixes, D6's included. The
  owner tags after the change is on `main` and CI there is green, and asks
  for the tag in the turn it is made.

### D6. `update` uses `selfupdate.DefaultHTTPClient()`

* `update.go` gains `updateHTTPClient()`, which returns
  `selfupdate.DefaultHTTPClient()`, and `defaultNewUpdateUpdater` passes it
  as `GitHubOptions.Client`. `updateTimeout` and the `time` import go.
* `cli.Run`'s default 15-minute run bound is kept; `update` passes no
  `Timeout` of its own.
* `migration_test.go` gains `TestUpdateHTTPClientHasNoOverallTimeout`:
  the helper's client has `Timeout` 0 and an `*http.Transport` with a
  positive `ResponseHeaderTimeout`. It was seen to fail on the old client
  (Context, "`update`'s HTTP client").
* It lands in its own commit, after D1, which brings `DefaultHTTPClient`
  in (`v1.18.0`).

### Consequences

* Good, because both modules and both workflows are current, and the
  publish workflow is at a commit whose publish path has published live.
* Good, because `configure` takes the wizard's fixes, among them that a
  refused sign-in no longer replaces a working session.
* Good, because `update` fails fast on a stalled request and no longer
  cuts off a slow download per request, as the library advises (D6).
* Neutral, because the installers and the hook itself behave as before.
* Bad, because a GitHub reply slower than 30 s to its headers, which the
  old client tolerated, now fails `update`; the user runs it again
  **[derived]**.

### Confirmation

* `go.mod` and `go.sum` change only on the two modules' lines; `go mod
  tidy -diff` is empty.
* `grep -c` gives `84805563…` 2 and `036289ec…` 0 in both `ci.yml` and
  the operations guide; `ci.yml` passes workflow-lint.
* `make verify`, `make preflight` and `make preflight-hosts` pass on the
  committed `HEAD`.
* D6: `grep -rn updateTimeout --include='*.go' .` finds nothing;
  `TestUpdateHTTPClientHasNoOverallTimeout` passes.
* CI is green on the branch and on `main`, the build rehearsal included.
* D5: the tagged `v1.10.2` publishes the six binaries, `SHA256SUMS` and
  both installers; `update` on an installed `v1.10.1` moves to it, which
  exercises D6's client against GitHub; and `version` names it.

## Pros and Cons of the Options

### A. Both at `v1.22.0`'s commit

* Good, because 0010-MADR A1 is met to the letter: the publish path is
  identical to `v1.17.2`'s, which published `v1.10.1` live.
* Good, because one version is pinned in `go.mod` and both workflows.
* Neutral, because nothing the workflows run changes; the rehearsal on the
  branch proves the pin resolves.

### B. Both stay at `v1.17.2`'s commit

* Good, because the workflow lines and the guide do not change.
* Bad, because the workflows name a version `go.mod` no longer requires,
  which `ci.yml`'s comment would have to explain, for no difference in
  what runs.

## Owner questions

* **Q1 (D3).** Leave `selfupdate.DefaultHTTPClient` for its own record
  (recommended), or adopt it here?
* **Q2 (D5).** Release `v1.10.2` once the change is on `main` and CI is
  green (recommended), or leave the move unreleased? A tag is asked for in
  the turn it is made.

**Answered 2026-10-24.** Q1: "adopt now", against the recommendation; D6
records it. Q2: "yes, we will release v1.10.2"; D5 records it. Option A
for the workflow pins was not questioned and stands.

## Amendments

### A1 (2026-10-24): the owner commits and pushes on `main`

*A1 supersedes the decision driver "No commit on `main`", which is left
above as written. Found at PLAN Phase 1.*

* **The decision (owner, 2026-10-24):** "i will commit and push you will
  not use branches. sysadmin-git-branch-gates is the mandate here". The
  owner's `sysadmin-git-branch-gates` skill takes priority over the
  organization's `enforcing-git-branch-gates` for this repository, which
  has no Jira project.
* **Consequences.** No branch and no pull request. The agent checks and
  stages each phase on `main`; the owner commits, with the global hook
  writing the message, and pushes. The Confirmation's "CI is green on the
  branch and on `main`" reads "on `main`". The release is tagged on `main`
  as before.
* The PLAN records it as Deviation 1.

## More Information

* The libraries' records, cited by repository and filename:
  go-llmprovider-sdk `0057-MADR-wizard-device-code-and-paste-prompt.md`
  and `0059-MADR-wizard-debugging-pass-findings.md`; go-selfupdate-lib
  `0031-MADR-export-default-http-client-and-launchd-status.md`,
  `0032-MADR-task-scheduler-registration-and-one-shot-handoff.md`,
  `0033-MADR-scrub-service-manager-environment-from-handoffs.md`,
  `0034-MADR-export-task-scheduler-enable-and-disable.md` and
  `0035-MADR-serve-supervisors-through-a-supervised-adapter.md`.
* The pin rule:
  [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md)
  D2 and Amendment A1; its reading and the last move:
  [0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md](0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md)
  A1 and A3.
* The implementation:
  [0021-PLAN-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md](0021-PLAN-adopt-go-llmprovider-sdk-v1-22-2-and-go-selfupdate-lib-v1-22-0.md).
