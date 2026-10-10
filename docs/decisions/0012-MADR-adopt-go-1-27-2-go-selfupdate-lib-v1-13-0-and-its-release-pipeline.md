---
status: accepted
date: 2026-10-09
decision-makers: Project Owner
consulted: go-selfupdate-lib (its records 0018, 0019 and 0020, which made v1.11.1 to v1.13.0, and its guides building-releases.md and extending-selfupdate.md); magic-cli-remote's 0169-MADR, the fleet toolchain rule, as amended 2026-10-08; probes on scratch clones of this repository, 2026-10-09
informed: users of the installed hook, who receive the change through `update`
---
# Move to Go 1.27.2 and go-selfupdate-lib v1.13.0, and build, publish and install through the library's release pipeline

## Context and Problem Statement

`prepare-commit-msg` builds with Go 1.27.1 and requires `go-selfupdate-lib
v1.10.1`. It releases through its own Makefile build and the library's
publish workflow pinned at `v1.10.0`, and it ships no installer. Since then:
* Go 1.27.1 gained published advisories;
* the library shipped `v1.11.0` to `v1.13.0`;
* the library's build workflow and installer templates matured.

This record assesses three moves and their order: the toolchain, the
library, and how the program is built, published and installed. It also
records how the repository measures against the fleet's standards.

Every claim is marked:
* **[read]**: from this repository, the library or a cited source;
* **[probed]**: run on 2026-10-09 on scratch clones of this repository at
  `aed7c60`, with a private module cache, never in its tree;
* **[derived]**: reasoned from the two, and not run.

### The toolchain

* **The module and the tools [read].**
  * `go.mod:3` is `go 1.27.1`, with no `toolchain` line.
  * `Makefile:1` is `MOD_VERSION := 1.27.1`.
  * `scripts/bootstrap-tools.sh:10` requires `GO_VERSION="go1.27.1"` and
    exits 1 on any other Go (`:14-17`).
  * The README says "Building from source needs Go 1.27.1" (`README.md:484`).
  * CI sets Go up from `go.mod` (`ci.yml:42-44`, `:96-98`), and runs
    `make verify`. That runs `tools`, `mod-check`, `fmt-check`, `lint`,
    `vet`, `test`, `coverage`, `vuln` (govulncheck) and `workflow-lint`,
    then `build-all` (`Makefile`).
* **The fleet rule [read].** magic-cli-remote's
  `docs/decisions/0169-MADR-standardize-toolchains-on-current-supported-advisory-free-releases.md`,
  "Amendment — 2026-10-08: Go 1.27.2 is the standard (D2)": "**D2 now
  reads:** Go **1.27.2** on every host and in CI, with the `go` directive
  of Go modules moved to 1.27.2 repository by repository after each passes
  its own gates." It lists 15 advisories, 13 in `stdlib` and 2 in
  `toolchain`, that 1.27.1 carries and 1.27.2 does not.
* **The development hosts [read].** This Mac, the Linux test host and the
  Windows test host run `go1.27.2`, per go-selfupdate-lib's
  `docs/decisions/0018-PLAN-move-toolchain-to-go-1-27-2.md` (V1).
* **What that does here [probed]:**

  | Probe | Result |
  | :--- | :--- |
  | P1: `make tools` with the host's Go, `go1.27.2` | exit 2: `expected go1.27.1, got go1.27.2`. `make verify` cannot start on any development host |
  | P2: `govulncheck` v1.8.0, source mode, at `aed7c60`, `GOTOOLCHAIN=go1.27.1` | exit 3: "Your code is affected by 10 vulnerabilities from the Go standard library." |
  | P3: the same with `GOTOOLCHAIN=go1.27.2` | exit 0: "No vulnerabilities found." |
  | B1: `govulncheck -mode=binary` on the hook installed on this Mac, `prepare-commit-msg version v1.7.0 (release) 141ad9b59590`, built with `go1.27.1` | exit 3: "Your code is affected by 12 vulnerabilities from the Go standard library.", GO-2026-6603 to -6617 among them |

* **CI has not caught it yet [derived].** The last CI run, on `aed7c60` on
  2026-10-08, passed before the advisories were published that day. The
  next run sets up `go1.27.1` from `go.mod` and runs `vuln`, so by P2 it
  fails.

### The library

* **Today [read].** `go.mod:7` requires `go-selfupdate-lib v1.10.1`.
  `update` is the library's `cli.Command` (`main.go:97`). It uses:
  * a GitHub source;
  * an exact asset selector over a hand-written list of six platforms
    (`update.go:39`);
  * a `StandaloneInstaller` with default options (`update.go:50`), so its
    target is the running executable under the home directory: the global
    hooks directory, `~/.global-git-hooks`, in the README's install.
* **What `v1.11.0` to `v1.13.0` change for a program like this one [read]**
  (the library's migration guide, §11–§13):
  * the result document's `schema_version` goes from 2 to 3, adding
    `rolled_back` and `probes_skipped`, then to 4, adding
    `replaced_before_stop`;
  * an interrupted update keeps the previous binary as
    `.<base>.selfupdate-kept-<n>`, which `KeptBackups` lists;
  * opt-in provenance checking (`selfupdate/verify/ghattest`);
  * opt-in replace-before-stop for services;
  * archive refusals;
  * `toolchain go1.27.2` in the library's own `go.mod`.

  The library's API only adds in these releases; its `make apicheck`
  passed each one.
* **The upgrade [probed],** on a scratch clone:
  * `go get github.com/maccavelli/go-selfupdate-lib@v1.13.0`;
  * `go mod edit -toolchain=go1.27.2`;
  * `go mod tidy`.

  Results:
  * `go.mod` and `go.sum` change in the library's lines only;
  * `go build ./...` and `go vet ./...` exit 0;
  * `govulncheck ./...` reports "No vulnerabilities found.";
  * `go test ./...` fails in one test only:

    ```text
    --- FAIL: TestUpdateCheckJSONSchema (0.00s)
        migration_test.go:186: last stdout line is kind "result", schema_version 4; want result, 2
    ```

    That is the test 0010-MADR added to catch exactly this. The other
    four packages pass. With `go 1.27.2` as the `go` line instead, the
    module also builds.
* **Lint and modernizers [probed].** On that clone, golangci-lint v2.14.0
  with this repository's `.golangci.yml` reports "0 issues." for linux,
  darwin and windows. `go fix -diff ./...` is empty for all three.
* **The README states schema 2** (`README.md:370`) [read].

### The release pipeline today

* **Build [read].**
  * The `go` job's tag steps run `make verify-release VERSION=$TAG`:
    `clean`, `build-all`, then `scripts/verify-release.sh`.
  * They upload `dist/` (`ci.yml:46-63`).
  * The Makefile cross-compiles six platforms with
    `-X …/buildinfo.version=$(VERSION) -X …/buildinfo.kind=release`, and
    on Linux with `-tags netgo` and `-extldflags '-static'`.
  * `verify-release.sh` checks:
    * the asset set by a fixed list of names;
    * `SHA256SUMS`;
    * a clean tree;
    * the *native* binary's `version` output, which on the Linux runner
      is linux/amd64 only.
* **Publish [read].** `ci.yml:123` calls
  `publish-selfupdate-release.yml@a0a26b6ecf66f51c19e9fea0f665c76ca5e99e4c
  # v1.10.0`, with a hand-written `platforms-json`. Its comment applies
  `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2's rule. It is restated
  in `0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md`
  D2 as "the newest workflow commit that has published live".
* **Releases [read],** from the GitHub API: `v1.5.0`, `v1.6.0` and `v1.7.0`
  are immutable, with 7 assets each (six binaries and `SHA256SUMS`), and
  no installer.
* **Install [read].** The README's install is manual:
  1. download or `make install`;
  2. copy the binary into `~/.global-git-hooks`;
  3. `git config --global core.hooksPath ~/.global-git-hooks`.

  `scripts/install-hooks.sh` installs this repository's *developer*
  hooks; it is not a user installer.

### What the library's pipeline offers, and what it asks

The library's `docs/guides/building-releases.md` (read at `v1.13.0`):
* **One spec,** `selfupdate-release.json`, embedded by the program and read
  by the build workflow, replaces each hand-kept platform list: the
  Makefile, `ci.yml` and `update.go` (§1–§2, §11).
* **The build workflow:**
  * builds every platform on one runner with a fixed recipe;
  * checks each binary's build information: the platform, cgo off,
    `-trimpath`, the commit, a clean tree, the toolchain, and the tag as
    the module version;
  * runs an identity command on the five platforms with a GitHub-hosted
    runner, every one of the six here but darwin/amd64;
  * writes `SHA256SUMS`, and stages;
  * rehearses on every non-tag push (§3, §5, §8).

  "With cgo off, `netgo`, `osusergo` and `-extldflags -static` change
  nothing" (§8).
* **The identity command** must print `buildinfo.Identity()` first:
  `<tag> (release)`, optionally with the commit (§3). This program's
  `version` prints `prepare-commit-msg version <identity>` (`main.go:85`),
  so its first line is not the identity. It would fail the check as it
  stands [derived].
* **Both workflows are pinned to the same commit** (§4). The build job's
  plan step requires the module to require a library version that reads
  the spec (§4, §12).
* **The installers.** `"installer": {}` makes the build workflow render
  `install.sh` and `install.ps1` into each release. They are published
  and attested, but not in `SHA256SUMS` (§12). They:
  * download over HTTPS only, and check `SHA256SUMS` before installing;
  * check the identity, and put the previous binary back on a failure;
  * keep `<product>.prev`;
  * take `--dir`/`-InstallDir`, `--version`, `--verify-attestation` and
    `--uninstall`.

  Their default directory is `~/.local/bin`, or
  `%LOCALAPPDATA%\Programs\<name>` on Windows. The spec's `installer`
  holds only `name`, `env_prefix` and `hooks`, so a program cannot set
  another default (§12) [read].
* **The live record [read]:**
  * go-selfupdate-lib's
    `docs/decisions/0015-PLAN-remediate-third-debugging-pass-findings.md`,
    "Release procedure step 7: the `v1.11.0` live rehearsal (2026-10-08)",
    published `v0.1.1` to `v0.1.3` of a throwaway program;
  * both workflows were pinned to `a0612da` (`v1.11.0`), with installers,
    raw and archive packaging;
  * every one-liner was run on macOS, Linux, Windows PowerShell 5.1 and
    PowerShell 7.
* **The publish workflow since that record [read],** from go-selfupdate-lib:
  * `git rev-parse` gives the same blob for
    `.github/workflows/publish-selfupdate-release.yml` at `v1.11.0` and
    `v1.13.0`, `a0c842a8…`;
  * the four scripts it runs from the pinned checkout
    (`check-release-tag.sh`, `refuse-existing-release.sh`,
    `verify-selfupdate-release.sh`, `release-latest-flag.sh`) are
    unchanged;
  * `go.mod` gained `toolchain go1.27.2`;
  * the release tool's non-test sources changed in four lines, by
    `go fix` (`docs/decisions/0019-MADR-apply-go-fix-modernizers.md`);
  * the installer templates changed in `v1.12.1`, `install.ps1`'s identity
    restore, which was not live-rehearsed.
  * The library's own CI rehearses the build workflow with ten identity
    legs on every push, run 37979214344 for the `v1.13.0` tag among them.

### What a release must keep

* **Users update from `v1.7.0` [read].** That client's exact selector asks
  for `prepare-commit-msg-<os>-<arch>[.exe]` (`update.go:39`). The build
  workflow's `"packaging": "binary"` ships exactly those names (§1). So a
  `v1.7.0` client can update to a release built by it [derived].
* **The installed binary lives in Git's hooks directory, which is shared
  [read].** On this Mac, `~/.global-git-hooks` also holds the global
  pre-push disclosure guard (`github-disclosure.py`, `pre-push`).
  Anything that sets `core.hooksPath` for the user would decide which
  other hooks run.

### The repository against the fleet's standards

Each is [read] unless marked:
* **No `AGENTS.md`,** and no agent pointers (`.claude/rules/`,
  `.grok/rules/`, `.opencode/rules.md`). The fleet's other Go repositories
  have them, go-selfupdate-lib's `AGENTS.md` among them.
* **Records live in three places:**
  * `docs/0003-*` to `docs/0005-*` at the `docs/` root;
  * `docs/plans/0001-*` and `0002-*`;
  * `docs/decisions/` for `0001`, `0002` and `0006` to `0011`.
* **Tools:**
  * `bootstrap-tools.sh` pins golangci-lint `v2.13.1`, govulncheck `v1.7.0`
    and actionlint `v1.7.12`;
  * go-selfupdate-lib pins golangci-lint `v2.14.0` and govulncheck
    `v1.8.0` (its 0006 and 0007 records), and actionlint `v1.7.12`;
  * the fleet rule names golangci-lint `v2.13.2` as its host standard
    (0169-MADR).
* **GitHub settings,** from the API:
  * immutable releases are on (`"enabled":true`);
  * the repository is public;
  * secret scanning and its push protection are on;
  * Dependabot security updates are off;
  * there is **no ruleset at all**:
    `GET /repos/…/rulesets?includes_parents=true` returns 0. Yet
    `scripts/configure-github.sh --apply` defines two:
    * `prepare-commit-msg-main`, which refuses deletion and non-fast-forward
      on the default branch;
    * `prepare-commit-msg-release-tags`, which restricts `v*` tags.

    The building guide asks for the tag ruleset before the first release
    (§4).
* **The README's attestation check,** `gh attestation verify … --repo …`
  (`README.md:519`), omits the `--signer-workflow` the building guide uses
  (§10).
* **`testfile.txt`,** two lines ("hello", "hi there"), is tracked. It was
  added and appended by the live checks of
  `0009-PLAN-adopt-go-llmprovider-sdk-v1-2-1.md` and
  `0010-PLAN-adopt-go-selfupdate-lib-v1-9-0.md`. Both PLANs leave its
  removal "for the owner", and nothing references it.

## Decision Drivers

* **Ship nothing built with a toolchain with published advisories**
  (0169-MADR D1). The installed `v1.7.0` carries 12, and only a new
  release reaches users.
* **The gates must run.** `make verify` cannot start on a development host
  today (P1).
* **Every existing install keeps updating.** `v1.7.0` clients must find
  the next release's assets under the same names.
* **One source of truth for the platforms and the release recipe**,
  instead of four hand-kept copies.
* **A release always publishes** (0010-MADR D2): a workflow is trusted
  when a published release proves it.
* **Nothing changes a user's Git configuration without asking.** The hooks
  directory is shared.
* **Each step lands green, and reverts alone.**

## Considered Options

**T, the toolchain:**
* **T1. `go 1.27.2` as the `go` line,** and the tools, the Makefile and the
  README at 1.27.2.
* **T2. `go 1.27.1` with `toolchain go1.27.2`,** as go-selfupdate-lib did
  (its 0018-MADR).
* **T3. Stay on 1.27.1.**

**L, the library:**
* **L1. Require `v1.13.0`;** take the schema-4 result and the
  kept-backup behaviour; adopt no opt-in feature.
* **L2. Require `v1.13.0`, and also report kept backups** (`KeptBackups`)
  after `update`.
* **L3. Stay on `v1.10.1`.**

**P, the publish pin:**
* **P1. `5e199c83…` (`v1.13.0`),** the commit the module requires.
* **P2. `a0612da…` (`v1.11.0`),** the newest commit with a live publish
  on record, word for word under 0010-MADR D2.
* **P3. Keep `a0a26b6…` (`v1.10.0`).**

**B, the build:**
* **B1. Adopt the build workflow** with an embedded
  `selfupdate-release.json`, and an `identity` command printing
  `buildinfo.Identity()`. The CI build steps and `verify-release.sh` go;
  the Makefile keeps its local build targets.
* **B2. Keep the Makefile build,** and only move the pins.

**I, installers:**
* **I1. `"installer": {"name": "prepare-commit-msg"}`** with no hooks. The
  README's one-liners pass `--dir ~/.global-git-hooks` (or `-InstallDir`),
  and leave `core.hooksPath` to the user, as today.
* **I2. I1 plus an `after_install` hook** that sets `core.hooksPath`.
* **I3. No installers.**

**D, the developer tools:**
* **D1. Align with go-selfupdate-lib:** golangci-lint `v2.14.0`,
  govulncheck `v1.8.0`, actionlint `v1.7.12`.
* **D2. Only rebuild the current pins with 1.27.2.**

**H, hygiene and settings:**
* **H1. In this change:**
  * apply the two rulesets with `scripts/configure-github.sh --apply`;
  * add `--signer-workflow` to the README's check;
  * delete `testfile.txt`.
* **H2. Record them, and do them separately.**

## Decision Outcome

Chosen options: **"T1", "L1", "P1", "B1", "I1", "D1" and "H1"**, the
owner's answers of 2026-10-09 (Owner questions), in four steps, each its own commit and each green. Then a
release, `v1.8.0`.

1. **Toolchain, library and tools (T1, L1, D1):**
   * `go.mod` gets `go 1.27.2` and `go-selfupdate-lib v1.13.0`;
   * `bootstrap-tools.sh`, `Makefile:1` and `README.md:484` move to 1.27.2;
   * the tool pins move to D1's versions;
   * `migration_test.go`'s `wantResultSchema` becomes 4, and
     `README.md:370` says schema 4.
2. **Pipeline (P1, B1):**
   * `selfupdate-release.json` at the repository root (the main package),
     embedded;
   * `update.go` takes its selector from `spec.AssetSelector()`;
   * an `identity` subcommand prints `buildinfo.Identity()`, and the spec
     names it in `identity_args`;
   * `ci.yml` calls the build workflow and the publish workflow, both at
     `5e199c83…` (`v1.13.0`), with the build `need`ing `go` and
     `go-native`;
   * the tag build steps, the upload and `verify-release.sh` go.
3. **Installers (I1):** `"installer": {"name": "prepare-commit-msg"}`, and
   the README's install section rewritten around the one-liners with
   `--dir`.
4. **Hygiene (H1):**
   * the rulesets are applied: an owner action, since it needs
     administrator rights;
   * the README's attestation check is fixed;
   * `testfile.txt` is removed.

Why each:
* **T1 over T2.** This is an application with no module depending on it,
  so raising the floor costs no one. The fleet rule asks for the `go` line
  itself, and the toolchain check in `bootstrap-tools.sh` then matches the
  module. T2 keeps a floor no consumer needs.
* **L1.** It takes every library fix. Two kinds of backup differ in how
  they are reported [read]:
  * a backup the run could not restore is reported, in the failure's
    detail ("pending backup …", `selfupdate/updater.go:403-404`) and in
    the JSON result's `pending_backup`;
  * a backup that an *interrupted* update left is kept as
    `.prepare-commit-msg.selfupdate-kept-<n>` by the next `update`, and
    no `Result` reports it. Only `KeptBackups` lists it
    (`selfupdate/types.go`, `Result.PendingBackup`'s comment). This
    program does not call it.

  L2 would close that gap. It is not recommended because an interrupted
  update of a hook binary is rare, and Git ignores a file in the hooks
  directory that is not named after a hook. The cost is a stray file, not
  a broken hook [derived]. Owner question 2 asks.
* **P1 over P2.** Since the live-published `v1.11.0` commit:
  * the publish workflow file is the same blob;
  * the scripts it runs are unchanged;
  * its tool differs by four `go fix` lines and Go 1.27.2.

  B1 also needs both workflows on one commit, and the module must require
  that commit's version. 0010-MADR D2's rule is met in substance; this
  record amends its wording to "the newest workflow commit whose publish
  path is unchanged from one that has published live" [derived]. P2 would
  pin the build workflow at `v1.11.0` too, behind the module's `v1.13.0`,
  which the guide asks not to do.
* **B1.** It replaces three platform lists and a native-only identity
  check with one spec and five identity runs. Every non-tag push
  rehearses the release recipe.
* **I1, not I2.** The hooks directory is shared (above), so the installer
  never sets `core.hooksPath`; the README tells the user to, as now.
* **D1.** The probe found no new finding with golangci-lint v2.14.0, and
  govulncheck v1.8.0 is what every host runs.
* **H1.** Each item is small, and the tag ruleset is a precondition the
  library's guide sets for any release.

### Consequences

* Good, because the next release, `v1.8.0`, is built with 1.27.2, and an
  `update` takes every user off the 12 advisories in `v1.7.0`.
* Good, because `make verify` runs again on every host.
* Good, because the platform list, build recipe, checks and installers
  come from one spec, and each push rehearses the release.
* Good, because users get verified one-line installs and uninstalls, on
  Windows too.
* Neutral, because `update --json`'s result is schema 4. A reader that
  ignores unknown keys sees no change.
* Neutral, because the installers' default directory is not where a git
  hook lives. The README must lead with `--dir`, and a user who copies a
  bare one-liner gets the binary on `PATH`, not in the hooks directory.
* Bad, because B1 moves the release build out of the repository's
  Makefile into a called workflow that this repository does not control.
  It is pinned by SHA, and the library's records cover it.
* Bad, because P1 amends 0010-MADR D2's wording rather than following it
  literally.
* Bad, because the installer templates' `v1.12.1` change has no live
  rehearsal. This repository's first release through them is that
  rehearsal for `install.ps1`.

### Confirmation

* `make verify` passes on this Mac, the Linux test host and the Windows
  test host, and in CI. `govulncheck` reports "No vulnerabilities found."
* `TestUpdateCheckJSONSchema` fails against `wantResultSchema = 2` with
  the new library, and passes at 4.
* A non-tag push runs the build workflow's rehearsal, with five identity
  legs printing `rehearsal-<commit> (local)`.
* **`v1.8.0`'s tag run:**
  * passes every job, with identity legs `v1.8.0 (release) <12-hex>`;
  * publishes an immutable release with six binaries, `SHA256SUMS`,
    `install.sh` and `install.ps1`;
  * `gh attestation verify … --signer-workflow …` passes on each.
* **On each host:**
  * `prepare-commit-msg update` takes an installed `v1.7.0` to `v1.8.0`,
    and `govulncheck -mode=binary` on the result reports no
    vulnerabilities;
  * the one-liner with `--dir` installs into a scratch hooks directory,
    and `--uninstall` removes it.
* **The rulesets:** `GET /repos/…/rulesets` lists both after the owner
  applies them.

## Pros and Cons of the Options

### T1 / T2 / T3

* T1: Good, because it meets the fleet rule as written, and its floor is
  the toolchain. Neutral, because no other module depends on this one.
* T2: Good, because it keeps a 1.27.1 floor. Bad, because the floor
  serves no consumer, and `bootstrap-tools.sh` must then accept two
  versions.
* T3: Bad, because releases keep shipping the advisories, and the gates
  cannot run on the hosts.

### L1 / L2 / L3

* L1: Good, because it takes every fix with one test change.
* L2: Good, because the backup an interrupted update left is named,
  which no `Result` does. Bad, because it adds a call and output for a
  rare case.
* L3: Bad, because B1 needs at least `v1.11.0`, and the library's fixes
  since `v1.10.1` are lost.

### P1 / P2 / P3

* P1: Good, because one commit serves the module, the build and the
  publish. Bad, because it rests on a by-content reading of the
  live-publish rule.
* P2: Good, because it follows the rule word for word. Bad, because the
  build workflow would be pinned behind the module.
* P3: Bad, because B1 cannot use it, and the publish job keeps building
  its tools with 1.27.1.

### B1 / B2

* B1: Good, because it gives one spec, checked builds and rehearsals on
  every push. Bad, because the build recipe becomes the library's.
* B2: Good, because nothing moves. Bad, because there are three platform
  lists, and only a native identity check.

### I1 / I2 / I3

* I1: Good, because installs are verified and reversible, with the hooks
  directory the user's choice. Bad, because the default directory is
  wrong for a hook, so the README carries `--dir`.
* I2: Bad, because it would rewrite a shared, global Git setting from a
  piped script.
* I3: Bad, because users keep a manual copy-and-chmod install with no
  checksum check.

### D1 / D2

* D1: Good, because it matches the library and the hosts, with no new
  finding. Neutral, because it differs from the fleet rule's v2.13.2.
* D2: Good, because it is a smaller change. Bad, because the tool pins
  drift from the library's.

### H1 / H2

* H1: Good, because the release precondition (the tag ruleset) is met
  before `v1.8.0`. Neutral, because the rulesets need the owner's
  administrator token.
* H2: Bad, because `v1.8.0` would ship without the tag ruleset the
  library's guide asks for.

## Amendments

### A1 (2026-10-09): `configure-github.sh --apply` cannot run as it stands

*Found while writing the PLAN, before any execution.*

* **Found [read].** `scripts/configure-github.sh` takes its "hardened
  workflow commit" from `git log -- .github/workflows/quality.yml
  .github/workflows/ci.yml .github/workflows/release.yml` (`:79-82`). It
  sets `REMOTE_READY=false` unless each of those three files' blobs matches
  `origin/main`'s (`:103-109`), and `--apply` refuses while it is false
  (`:231-234`). `.github/workflows/` holds only `ci.yml`, so `--apply` can
  never run. H1's "apply the two rulesets with `scripts/configure-github.sh
  --apply`" cannot be done as written.
* **What the script defines [read]** (`:157-193`):
  * `prepare-commit-msg-main`: deletion and non-fast-forward refused on the
    default branch, with administrators able to bypass;
  * `prepare-commit-msg-release-tags`: creation, deletion and
    non-fast-forward refused for `refs/tags/v*`, with administrators and the
    GitHub Actions integration able to bypass.
* **Dependabot [read].** Commit `c50de30`, "ci: disable dependabot and fix
  Windows tests", turned it off on purpose. The owner's "Leave off" agrees.
* **Decided, as H1's method (for the owner's approval with the PLAN):** the
  PLAN's Phase 4 changes the script's workflow list to the one file,
  `ci.yml`, and changes nothing else in it. The owner runs it without
  `--apply` to read the plan, then with `--apply`. The rulesets are those
  the script already defines.

## More Information

### Owner questions

Asked and answered on 2026-10-09. Every answer is the recommended one:

1. **T:** **"T1: go 1.27.2 line"**.
2. **L:** **"L1: v1.13.0, no opt-ins"**. The backup an interrupted update
   leaves stays unreported (L2 not taken).
3. **P:** **"P1: v1.13.0, amend rule"**. The PLAN amends
   `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2's wording, as Decision
   Outcome says, in the same commit that moves the pin.
4. **B:** **"B1: adopt it"**.
5. **I:** **"I1: yes, no hooks"**.
6. **D:** **"D1: align with the library"**.
7. **H:** **"H1: in this change"**. Dependabot security updates:
   **"Leave off"**.
8. **The identity command:** **"New `identity` command"**. `version`'s
   output is unchanged.
9. **The release:** **"One v1.8.0 at the end"**.

### Not decided here

* **An `AGENTS.md`, the agent pointers and a records tool** (the fleet's
  workspace scaffold), **and moving `0001` to `0005` into
  `docs/decisions/`.** Each is its own record.
* **A default install directory in the installer spec:** a library change
  for go-selfupdate-lib's records, if the README's `--dir` proves not
  enough.
* **`ghattest` in `update`.** It needs a logged-in `gh` on every machine
  that updates, which suits a developer's machine. It is opt-in, and not
  adopted.

### Not verified

* The build workflow and the installers run against this repository. The
  first non-tag push after step 2 is that run.
* `install.ps1` from `v1.12.1` on, live. It is covered by the library's
  unit and Windows-host tests, not by a published release.
* Whether golangci-lint v2.14.0 finds anything new after step 2's code
  changes. The probe ran on step 1's tree.

### Sources

* This repository: `go.mod`, `Makefile`, `ci.yml`, `update.go`,
  `main.go`, `migration_test.go`, `README.md`, `scripts/*.sh`,
  `docs/decisions/0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md`,
  `docs/decisions/0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md`.
* go-selfupdate-lib at `v1.13.0`:
  * `docs/guides/building-releases.md` and
    `docs/guides/migrating-from-mcplib-selfupdate.md`;
  * `docs/decisions/0015-PLAN-remediate-third-debugging-pass-findings.md`,
    the live rehearsal;
  * `docs/decisions/0018-MADR-move-toolchain-to-go-1-27-2.md`.
* magic-cli-remote,
  `docs/decisions/0169-MADR-standardize-toolchains-on-current-supported-advisory-free-releases.md`,
  its 2026-10-08 amendment.
* Go's toolchain rules for the `go` and `toolchain` lines:
  <https://go.dev/doc/toolchain>.
