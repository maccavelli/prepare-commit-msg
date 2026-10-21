---
status: accepted
date: 2026-10-21
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its CHANGELOG, v1.16.0 to v1.21.0); go-selfupdate-lib (its CHANGELOG, v1.16.0 to v1.17.2, its migration guide, and its 0025-PLAN's S13 live rehearsal); the release runs of gobble-cli v0.1.1 and magic-cli-remote v0.22.2; magic-cli-remote's 0169-MADR, the fleet toolchain rule, as amended 2026-10-21, and its 0169-PLAN's P14; scratch clones of this repository moved to the new pins, 2026-10-20 and 2026-10-21
informed: users of the installed hook, who receive the change through `update`
---
# Move to Go 1.27.3, go-llmprovider-sdk v1.21.0 and go-selfupdate-lib v1.17.2, with the publish workflow at the newest live-published commit

*Revised 2026-10-21, before approval, twice. First, the owner asked that
the repository be pinned to Go 1.27.3 (D6). Then the owner asked again for
the libraries' current releases: go-llmprovider-sdk had released `v1.19.0`
to `v1.21.0` and go-selfupdate-lib `v1.17.1` and `v1.17.2` since the first
draft, which targeted `v1.18.0` and `v1.17.0`. This text replaces that
draft's. The filename keeps the first draft's slug, which
`0019-MADR-skip-full-ci-for-docs-only-pushes.md` and its PLAN already
link.*

## Context and Problem Statement

`prepare-commit-msg` pins Go in three places and two libraries in four
**[read]**:

| Pin | Where | Pinned | Newest |
| :--- | :--- | :--- | :--- |
| Go, the `go` directive | `go.mod:3` | `1.27.2` | `1.27.3` |
| Go, `MOD_VERSION` | `Makefile:14` | `1.27.2` | `1.27.3` |
| Go, the tools' toolchain check | `scripts/bootstrap-tools.sh:10` | `go1.27.2` | `go1.27.3` |
| go-llmprovider-sdk | `go.mod:6` | `v1.15.0` | `v1.21.0` (2026-10-21) |
| go-selfupdate-lib | `go.mod:7` | `v1.15.2` | `v1.17.2` (2026-10-21) |
| go-selfupdate-lib build workflow | `.github/workflows/ci.yml:105` | `436a7f5…` (`v1.15.2`) | `036289ecc701d84c7749f4950a57a154c1af2cf4` (`v1.17.2`) |
| go-selfupdate-lib publish workflow | `.github/workflows/ci.yml:120` | `436a7f5…` (`v1.15.2`) | `036289ec…` (`v1.17.2`) |

CI takes Go from `go.mod` (`ci.yml:42`, `:83`, `go-version-file: go.mod`),
and so does go-selfupdate-lib's build workflow for the release binaries
(`go-version-file: src/<module-dir>/go.mod`) **[read]**. The `go`
directive is therefore the toolchain every CI and release build uses.

The owner asked to assess the libraries' current releases and move the
pins to them, and to pin the repository to Go 1.27.3. What has to be
decided is whether both modules move to their newest releases, where the
publish workflow pin may move under the rule that governs it, how Go
moves, and what the user is told.

Every claim below is marked:

* **[read]**: from this repository, another repository's files, records or
  release runs, or a library's source, changelog or records at the version
  named;
* **[probed]**: run on a clone of this repository at `87f0e64` in the
  session scratchpad, never in this tree: on 2026-10-21 at `HEAD` as is;
  with the three Go pins moved to 1.27.3 alone; and with them and both
  libraries at their newest releases;
* **[derived]**: reasoned from the two, and not run.

### The toolchain

* **Go 1.27.3 fixes an advisory this module reaches.** GO-2026-6809
  (CVE-2026-111943, "Distrust compromised intermediate certificates in
  crypto/x509"), published 2026-10-21, is found in `crypto/x509@go1.27.2`
  and fixed in `go1.27.3` and `go1.26.10` (magic-cli-remote's
  `0169-MADR-standardize-toolchains-on-current-supported-advisory-free-releases.md`,
  "Amendment — 2026-10-21", from OSV) **[read]**. At `HEAD`, govulncheck
  under go1.27.2 exits 3 **[probed]**:

  ```text
  Vulnerability #1: GO-2026-6809
      Found in: crypto/x509@go1.27.2
      Fixed in: crypto/x509@go1.27.3
        #1: main.go:102:15: prepare.main calls fmt.Fprintf, which eventually calls x509.Certificate.Verify
  ```

  The hook verifies TLS certificates on every provider request and on
  `update` **[derived]**. Released `v1.10.0` was built from `go 1.27.2`
  **[read]**, so it carries the flaw **[derived]**. CI's `make verify`
  runs govulncheck under the toolchain `go.mod` names, so CI's `go` job
  fails on `HEAD` from 2026-10-21 **[derived]**.
* **The fleet standard is now Go 1.27.3.** The same amendment reads: "Go
  **1.27.3** on every host and in CI, with the `go` directive of Go modules
  moved to 1.27.3 repository by repository after each passes its own
  gates" **[read]**.
* **Every host is on it.** This Mac runs `go1.27.3`, with
  `GOTOOLCHAIN=go1.27.3` in Go's env file **[probed]**. The Linux and
  Windows test hosts moved to `go1.27.3` in magic-cli-remote's 0169-PLAN
  P14, each confirmed in a fresh session **[read]**.
* **The tools' check now fails on this Mac.** `make tools`, and so `make
  verify` and `make preflight`, stops at `HEAD` with `expected go1.27.2,
  got go1.27.3` (`scripts/bootstrap-tools.sh:14-17`) **[probed]**.
* **The move is three lines.** With `go.mod:3`, `Makefile:14` and
  `scripts/bootstrap-tools.sh:10` at 1.27.3, and no library moved, `go mod
  tidy -diff` is empty and `make verify` exits 0, with "No vulnerabilities
  found." **[probed]**. `MOD_VERSION` is read by nothing else in the
  repository **[read]**. `go.mod` has no `toolchain` line, and the move
  adds none **[read]**.
* **The libraries now require it too.** go-llmprovider-sdk `v1.20.1`
  requires `go 1.27.3` for the same advisory, and `go get` of it raises
  this module's `go` line **[read]**. go-selfupdate-lib `v1.17.2` requires
  `go 1.26.10`, with `toolchain go1.27.3` **[read]**.

### What the library move changes in the module graph

* With the Go pins at 1.27.3, `go get …go-llmprovider-sdk@v1.21.0
  …go-selfupdate-lib@v1.17.2` and `go mod tidy` change two `require` lines
  of `go.mod` and four lines of `go.sum`, the two modules' own **[probed]**.
* Both releases require the `golang.org/x/{mod,sys,term}` versions this
  module already does (`x/mod v0.40.0`, `x/sys v0.47.0`, `x/term v0.43.0`)
  **[read]**. No other module moves.
* `make verify` exits 0: mod-check, fmt-check, golangci-lint, `go vet`,
  `go test -race ./...`, the coverage threshold, `govulncheck` ("No
  vulnerabilities found."), workflow-lint, the tools' 25 tests, and all six
  release cross-builds **[probed]**. No test changes.

### go-llmprovider-sdk, v1.15.0 to v1.21.0

From its CHANGELOG **[read]**. Every release is additive to the exported
API, which its `make api-check` holds.

* **v1.16.0:** `catalog.Cache` / `catalog.WithCache` and
  `providers.DefaultWith`, a caller-owned cache; concurrency and error-kind
  docs on every export.
* **v1.17.0:** like failures match like sentinels, and those errors' text
  gains the sentinel's text first (browser and device sign-in denials and
  expiry, an unreachable Ollama URL, an unserved provider). A nil
  `*StaticToken` is refused where it panicked.
* **v1.18.0:** a typed-nil token source of any type is refused with
  `ErrInvalidRequest` wherever it panicked.
* **v1.19.0:** `llmprovider.WithStreamRetry`, opt-in; `WithRetry` and
  `WithRetryNotify` still buffer, unchanged.
* **v1.20.0:** a failure the service reports inside a 200 event stream
  (Responses' `response.failed`, Messages' `error` event) is retried at
  most once per call by `WithRetry`, where it was retried up to
  `MaxAttempts` times. Fixes: quadratic accumulation of streamed
  fragments; a panic on a negative stream index; concurrent `catalog.List`
  calls share one fetch; an Ollama `/api/show` reply is refetched after ten
  minutes or a digest change; concurrent listings share one run of the
  model probes.
* **v1.20.1:** requires Go 1.27.3, for GO-2026-6809.
* **v1.21.0:** `wizard.ResolveCredential`, `Options.KiloDeviceKey`,
  `Options.SkipMethods` and `Options.NoEndpointPrompt`, all opt-in. The
  wizard's notices that report a service's error show its class and HTTP
  status, such as "authentication failed, HTTP 401", never the service's
  message, which can echo a key; the errors it returns are unchanged.

**What reaches this app:**

* **Retries.** The hook calls `llmprovider.WithRetryNotify`
  (`main.go:423`) **[read]**. A failure reported inside a 200 stream is
  now retried at most once, whatever `retry_count` says; other failures
  are retried as before **[derived]**.
* **The wizard** (`wizard.ConfigureLLM`, `internal/ui/setup.go:269`)
  **[read]**: a refused key, a listing warning or a Kilo profile notice
  reads "<class>, HTTP <status>", not the service's text; sign-in and
  Ollama failures read with their kind first **[derived]**. No test here
  matches those texts, and every test passes **[probed]**.
* **Token sources.** Every one this app builds is
  `llmprovider.NewStaticToken(…)`, a non-nil value (`main.go:366`,
  `internal/ui/setup.go:166`) **[read]**: the typed-nil fixes change
  nothing it does **[derived]**.
* The curated model lists did not change: `TestReadmeCuratedModels`
  passes **[probed]**.

### go-selfupdate-lib, v1.15.2 to v1.17.2

From its CHANGELOG and migration guide **[read]**. Every exported change
is an addition, which its `make apicheck` holds.

* **v1.16.0:** `NewGitHubUpdater`; typed `HTTPError` and
  `service.CommandError`; context variants of the hand-off calls; clock,
  probe and slog seams; opt-in release signing (`signednote`, the publish
  workflow's `sign` input). Credentials print as `<redacted>`.
* **v1.17.0:** the `before_uninstall` installer hook, and the installers'
  `--force` / `-Force`. "Nothing changes without the hook."
* **v1.17.1:** an uninstall that cannot remove a program, such as a
  running one, reports it, removes the others and exits 3, where
  `install.ps1` reported it removed and exited 0; a nil `*Updater`'s
  `Checker` no longer panics; a managed install's session serializes its
  methods.
* **v1.17.2:** requires Go 1.26.10, built and tested with Go 1.27.3; "for
  a caller that pins `v1.17.2`, the publish workflow builds its release
  tools with Go 1.27.3."

**What reaches this app:**

* `update.go` builds its updater from `NewGitHubSource`,
  `NewStandaloneInstaller` and `selfupdate.New` (`update.go:66-97`), and
  `selfupdate-release.json` declares no installer hooks **[read]**. No call
  it makes changes behaviour, and the `--json` result's schema test passes
  **[probed]**.
* The released `install.sh` / `install.ps1` come from the build workflow's
  pinned commit. Their usage gains `--force` (`-Force`), which matters only
  with a hook; and an `--uninstall` that cannot remove a running
  `prepare-commit-msg` now says so and exits 3 **[read]**.

### The workflow pins, and the rule that governs the publish pin

`0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` Amendment A1 (decided in
`0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md`)
reads: "the publish pin is the newest go-selfupdate-lib commit whose publish
path, meaning the publish workflow file and the scripts it runs from the
pinned checkout, is unchanged from a commit that has published live."
0010-MADR D2 lets the publish pin differ from `go.mod`'s version, with a
comment that says why.

**The build workflow** file is the same blob (`56f95cd5…`) from `v1.11.0`
to `v1.17.2` **[read]**. What it runs from the checkout changed (the plan
step's `before_uninstall` floor; the installer templates). This repository's
CI runs it on every push to `main` as a rehearsal, so a push proves it
before any tag **[read]**.

**The publish path,** by git object **[read]**:

| Path the publish job uses | `adc2c75` (`v1.17.0`) | `e8d4904` (`v1.17.1`) | `036289e` (`v1.17.2`) |
| :--- | :--- | :--- | :--- |
| `.github/workflows/publish-selfupdate-release.yml` | `3e07c854` | `5048ad0c` | `5048ad0c` |
| `scripts/` | `23c47d2b` | `112fb099` | `829c22fd` |
| `internal/cmd/selfupdate-release/` | `22b7182d` | `2ec6408f` | `20594a2d` |
| `go.mod` (the job's `go-version-file`) | `3462ade1` | `3462ade1` | `24d7b31e` |

**Live publishes, read from the repositories and their runs:**

| Publish pin | Published by | Run | `sign` |
| :--- | :--- | :--- | :--- |
| `adc2c757…` (`v1.17.0`) | gobble-cli `v0.1.1`, 2026-10-20 | `39363019252`, success | not set |
| `e8d4904a…` (`v1.17.1`) | magic-cli-remote `v0.22.0`, 2026-10-20, and `v0.22.2`, 2026-10-21 | `39635656900`: "sign: skipped", "publish: success" | not set |
| `036289ec…` (`v1.17.2`) | none on record | | |

Both published runs called the workflow as this repository does, with
`sign` unset **[read]**.

**`e8d4904` against `036289e`:** the workflow gains a header comment; the
scripts the publish job runs change in comments and a usage string
(`refuse-existing-release.sh`, `verify-selfupdate-release.sh`); the
`selfupdate-release` package changes only in installer templates, which
the build stages, and tests; and `go.mod`'s `toolchain` moves from
`go1.27.2` to `go1.27.3`, so the publish job compiles its release tools
with Go 1.27.3 **[read]**. Under A1's words `v1.17.2`'s publish path is
changed, and nothing has published through it.

So:

* **By A1 as written,** the publish pin is `e8d4904a…` (`v1.17.1`): the
  newest commit with a live publish, in this repository's configuration.
* **`v1.17.2`'s** differences are comments, a usage string, and the Go
  that builds the publish job's tools. Pinning it reads A1 in substance,
  and makes this repository's next tag the first publish through it.

## Decision Drivers

* **No advisory this module reaches** ships in a release, or is left to
  fail CI.
* **The fleet toolchain standard** (magic-cli-remote's 0169-MADR D2, as
  amended 2026-10-21): Go 1.27.3.
* **Pins are current** unless a record names a reason to hold one.
* **The publish pin follows 0010-MADR A1:** proven by a live publish, not
  by this repository's tests.
* **No silent change for a user who changes nothing.** The README says what
  a user can see.
* **The change is a pin move.** New library features are separate
  decisions.

## Considered Options

Every option takes D6, the Go move, which the owner asked for, and moves
both modules to their newest releases; they differ in the workflow pins.

* **A. Build `v1.17.2`, publish `v1.17.1`:** the build pin follows
  `go.mod`; the publish pin is A1's, the newest live-published commit.
* **B. Both workflows at `v1.17.2`,** reading A1 in substance.
* **C. Both workflows at `v1.17.1`,** the newest live-published commit for
  both.

## Decision Outcome

Chosen option: "A", because it moves every pin as far as its own evidence
reaches: the modules and the build workflow to the newest releases, which
this repository's tests and CI rehearsal prove, and the publish workflow to
the newest commit that has published live in this configuration, which A1
requires. Accepted 2026-10-21 (see "Owner questions"). D6 is the
owner's (2026-10-21) and lands first.

### D1. Both modules move

* `go.mod` requires `go-llmprovider-sdk v1.21.0` and
  `go-selfupdate-lib v1.17.2`; `go mod tidy`. Only those two modules'
  lines change in `go.mod` and `go.sum`.

### D2. The workflow pins

* The build workflow's `uses:` pins
  `036289ecc701d84c7749f4950a57a154c1af2cf4 # v1.17.2`, the version
  `go.mod` requires.
* The publish workflow's `uses:` pins
  `e8d4904a09194d99761042a8996983e8d5b957eb # v1.17.1`.
* The comment above them is rewritten to say why they differ (0010-MADR
  D2): the publish pin is the newest commit with a live publish, magic-cli-remote
  `v0.22.2`'s (run `39635656900`), with `sign` unset as here; it moves to
  `v1.17.2`'s commit, in a separate change, once a publish through that
  commit is on record. It cites this record and 0010-MADR A1.
* The `release` job's `with:` is unchanged. `sign` is not passed and
  defaults to `false`; its permissions (`contents: write`, `id-token:
  write`, `attestations: write`) are what the publish job declares
  **[read]**.
* The build at `v1.17.2` and the publish at `v1.17.1` agree: the build
  workflow is the same blob at both, and the publish job's scripts differ
  between them only in comments and a usage string **[read]**.

### D3. No library feature is adopted here

* Not adopted: `catalog.WithCache`, `providers.DefaultWith`,
  `WithStreamRetry`, the new `wizard` entry point and options,
  `NewGitHubUpdater`, release signing, the `before_uninstall` hook.
* `NewGitHubUpdater` could replace part of `update.go`'s hand-built
  updater, which 0014-MADR's driver ("every SDK API that removes hand work
  is used") would favour. It is recorded as a candidate for its own record
  (Q2), not taken in a pin move.

### D4. The README states what changed

* A section, "Changes with Go 1.27.3, go-llmprovider-sdk v1.21.0 and
  go-selfupdate-lib v1.17.2", in the pattern of the two before it:
  * built with Go 1.27.3, which fixes GO-2026-6809 in `crypto/x509`, an
    advisory `v1.10.0`'s Go 1.27.2 carries (D6);
  * a failure a provider reports inside a successful stream is retried
    once, whatever `retry_count` is;
  * the wizard's notices of a service's error show its class and HTTP
    status, not the service's message; sign-in and Ollama failures read
    with their kind first;
  * the installer: `--uninstall` that cannot remove a running copy says
    so and exits 3; its usage lists `--force` (`-Force`), which only a
    `before_uninstall` hook uses, and this app has none;
  * `update` behaves as in the release before.
* The table of contents gains the section.
* "Building from source needs Go 1.27.2" (`README.md:705`) reads 1.27.3.
  The `v1.8.0` section's Go 1.27.2 text (`README.md:572-578`) is history,
  and stays.

### D5. The release (Q3)

* Recommended: `v1.10.1`, a patch: no new feature, and the visible changes
  are messages, a retry limit in one failure case, and the installer's
  uninstall report. It is the first release without GO-2026-6809, which
  `v1.10.0` carries (D6). The owner tags after CI is green on `main`.

### D6. Go moves to 1.27.3

* `go.mod:3` becomes `go 1.27.3`, `Makefile:14` `MOD_VERSION := 1.27.3`,
  and `scripts/bootstrap-tools.sh:10` `GO_VERSION="go1.27.3"`. No
  `toolchain` line is added.
* It lands in its own commit, before D1 and D2: it alone turns CI's
  govulncheck and this Mac's `make tools` green, and it passes `make
  verify` without the library moves **[probed]**. A hold on Q1 does not
  hold it.
* Records `0001` to `0017` keep their Go text as written.

### Consequences

* Good, because the hook is built with a Go that has no advisory it
  reaches, and CI and this Mac's tools check pass again (D6).
* Good, because both modules and the build workflow are current, and the
  publish workflow is at a commit proven by a live publish in this
  configuration.
* Neutral, because the hook's own behaviour changes only in the one retry
  case and in messages.
* Bad, because the two workflow pins differ by a patch release, and the
  publish job's tools are built with Go 1.27.2 (`v1.17.1`'s `toolchain`)
  until the publish pin moves (option A). Those tools check staged files;
  the uploads are `gh`'s **[read]**.
* Bad, because building from source now needs Go 1.27.3; a 1.27.2
  toolchain without `GOTOOLCHAIN=auto` refuses the module (D6).

### Confirmation

* D6: the three Go pins read 1.27.3 and no `1.27.2` pin remains outside
  the records and the README's `v1.8.0` section; `make tools` passes on
  this Mac; govulncheck reports no vulnerability; CI's setup-go logs Go
  1.27.3.
* `go.mod` and `go.sum` change only on the two modules' lines; `go mod
  tidy -diff` is empty.
* The build `uses:` line names `036289ec…`, the publish line `e8d4904a…`
  (option A), and `ci.yml` passes workflow-lint.
* `make preflight` and `make preflight-hosts` pass on the committed `HEAD`.
* CI on `main` is green, its build rehearsal included.
* If Q3 is yes: the tagged release publishes, with binaries, `SHA256SUMS`
  and installers and no `SHA256SUMS.note`; `update` on an installed copy
  moves to it, and `version` names the release.

## Pros and Cons of the Options

### A. Build `v1.17.2`, publish `v1.17.1`

* Good, because A1 is met to the letter, by a live publish in this
  repository's configuration.
* Good, because the build, whose installers and plan step carry the
  library's version, matches `go.mod`.
* Neutral, because it is how 0010-MADR D2 first split the pins.
* Bad, because the pins differ, and the publish job's tools build with Go
  1.27.2 until a later change moves the pin.

### B. Both at `v1.17.2`

* Good, because one version is pinned everywhere, and the publish job's
  tools build with Go 1.27.3, as the fleet standard says CI should.
* Bad, because A1's letter is not met: nothing has published through
  `036289e`, so this repository's tag would be the first.

### C. Both at `v1.17.1`

* Good, because both workflows are live-proven, and they match.
* Bad, because the build would stage `v1.17.1`'s installers for a binary
  that embeds `v1.17.2`, and the build pin no longer follows `go.mod`.

## Owner questions

* **Q1 (D2).** Option A, build `v1.17.2` and publish `v1.17.1`, by A1 as
  written (recommended); or option B, both at `v1.17.2`?
* **Q2 (D3).** Leave `NewGitHubUpdater` for a later record (recommended),
  or adopt it here?
* **Q3 (D5).** Release `v1.10.1` after the push (recommended), or leave the
  move unreleased?

**Answered 2026-10-21.** The owner asked that the current releases be
pinned and Go moved to 1.27.3 ("we are going to update"), which approves
the PLAN. That instruction named neither option, so Q1 takes the
recommendation, **A**, which A1 requires. Q2 takes the recommendation:
`NewGitHubUpdater` waits for its own record. **Q3 is open.** A tag needs
the owner's ask in the turn it is made, so Phase 5 waits for it.

## More Information

* The libraries' records, cited by repository and filename:
  go-llmprovider-sdk `0048-MADR-adopt-go-sdk-best-practices.md`,
  `0049-MADR-v1-16-0-follow-ups.md`,
  `0050-MADR-refuse-typed-nil-token-sources.md`,
  `0051-MADR-opt-in-streaming-retry.md`,
  `0053-MADR-stream-retry-and-cache-defects.md`,
  `0055-MADR-go-1-27-3-and-the-v1-20-0-release.md` and
  `0056-MADR-wizard-credential-only-entry-point.md`; go-selfupdate-lib
  `0025-MADR-align-with-go-sdk-design-practices.md`,
  `0026-MADR-before-uninstall-installer-hook.md`,
  `0027-MADR-documentation-pass-after-v1-17-0.md` and
  `0029-MADR-move-toolchain-to-go-1-27-3-and-floor-to-1-26-10.md`.
* The toolchain rule, cited by repository and filename: magic-cli-remote
  `0169-MADR-standardize-toolchains-on-current-supported-advisory-free-releases.md`,
  "Amendment — 2026-10-21: Go 1.27.3 is the standard (D2)", and its
  0169-PLAN's P14 (the test hosts). The last Go move here:
  [0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md](0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md).
* The pin rule:
  [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md)
  D2 and Amendment A1; the last move of both pins:
  [0014-MADR-adopt-go-llmprovider-sdk-v1-11-0-and-go-selfupdate-lib-v1-15-2.md](0014-MADR-adopt-go-llmprovider-sdk-v1-11-0-and-go-selfupdate-lib-v1-15-2.md)
  D1; the last SDK move:
  [0015-MADR-adopt-go-llmprovider-sdk-v1-15-0.md](0015-MADR-adopt-go-llmprovider-sdk-v1-15-0.md).
* The implementation:
  [0018-PLAN-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md](0018-PLAN-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md).
