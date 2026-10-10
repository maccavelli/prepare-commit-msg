---
status: in-progress
date: 2026-10-09
associated-madr: "0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md"
---
# Implement the move to Go 1.27.2 and go-selfupdate-lib v1.13.0, and the library's build, publish and install pipeline

Associated MADR: [0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md](0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md)

## Goal

* **T1, D1:** Go 1.27.2 everywhere:
  * `go.mod`'s `go` line is `1.27.2`;
  * `scripts/bootstrap-tools.sh` requires `go1.27.2`, and pins
    golangci-lint `v2.14.0`, govulncheck `v1.8.0` and actionlint
    `v1.7.12`;
  * `Makefile:1` and `README.md` name 1.27.2;
  * `make verify` and the pre-commit hook run on every host again.
* **L1:** `go.mod` requires `go-selfupdate-lib v1.13.0`, and
  `TestUpdateCheckJSONSchema` and the README say schema 4.
* **B1, P1:**
  * `selfupdate-release.json` at the root, embedded, is the one platform
    list `update` uses;
  * an `identity` command prints `buildinfo.Identity()`;
  * CI builds through `build-selfupdate-release.yml` and publishes through
    `publish-selfupdate-release.yml`, both at
    `5e199c831b5691ea687943e3c3fd495d50c739ed` (`v1.13.0`);
  * `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2's wording is amended.
* **I1:** each release carries `install.sh` and `install.ps1`. The README
  installs with `--dir` and leaves `core.hooksPath` to the user.
* **H1:**
  * the README's attestation check names the signer workflow;
  * `testfile.txt` is gone;
  * both rulesets are applied (by the owner).
* **The release:** `v1.8.0` carries all of it. The installed hook is
  updated to it, and the live checks pass.
* **0011:** `0011-PLAN-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md`
  stopped before its Phase 3, the `v1.8.0` release through `a0a26b6`.
  That phase, and its live checks, are carried into this PLAN's Phase 5,
  with this PLAN's pin.

Done means every item under Verification holds.

## Scope

### In scope

| Phase | Who | What |
| :--- | :--- | :--- |
| 0 | agent, then owner | this PLAN, the index, 0011-PLAN's note; the owner commits the records |
| 1 | agent | T1, L1, D1: the toolchain, the library, the tools |
| 2 | agent | B1, P1: the spec, `identity`, the build and publish workflows, 0010-MADR's wording |
| 3 | agent | I1: the installers and the README's install section |
| 4 | agent, then owner | H1: the README check, `testfile.txt`, and `configure-github.sh`'s workflow list (0012-MADR A1); the owner applies the rulesets |
| 5 | owner, then agent | push, CI and its rehearsal, tag `v1.8.0`, release checks, `update`, the live checks |
| 6 | agent | close-out |

### Out of scope

* MADR 0012's "Not decided here":
  * `AGENTS.md`, the agent pointers and a records tool;
  * moving `0001` to `0005` into `docs/decisions/`;
  * a default install directory in the library's spec;
  * `ghattest`.
* L2 (reporting the backup an interrupted update left), and Dependabot.
* Any change in go-selfupdate-lib or go-llmprovider-sdk.
* The Makefile's own `build-all`, which stays for local builds. Its
  platform list is the one copy B1 does not remove.
* Push, tags and the rulesets, which are the owner's.

## Rules for every phase

`0011-PLAN-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md`'s
"Rules for every phase" 1–7 apply as written:
* each phase ends staged for the owner's commit, and the agent does not
  commit;
* rule 2's checks: `make verify`, `make verify-staged`, `go test -race`,
  the cross-OS `go vet`, `go mod tidy -diff` and `gofmt -l`;
* proofs on scratch copies, with each new test seen to fail on a planted
  break;
* Python for session tooling;
* no identifiers;
* keys from the environment, never printed;
* deviations stop and prompt.

Two additions:

8. **Workflow checks.** A phase that changes `ci.yml` also runs
   `make workflow-lint`, which runs actionlint.
9. **Never `--no-verify`.** The pre-commit hook, `make verify-staged`, is
   the gate for every commit, records included.

## Implementation Steps

### Phase 0: records

1. **`docs/README.md`** indexes this PLAN after 0011-PLAN, in its list
   form, `Proposed`.
2. **0011-PLAN** gains a dated note:
   * Phases 0–2 landed;
   * Phase 3, the `v1.8.0` release and its live checks, and Phase 4 are
     carried into 0012-PLAN's Phases 5 and 6, with the pin and toolchain
     0012-MADR decided;
   * its status becomes `superseded`, with 0012-PLAN named.

   0011-PLAN's V2, "published through `a0a26b6`", does not hold, and the
   note says why.
3. **On approval,** this PLAN becomes `in-progress`, with an Approval
   entry.
4. **The commit order the hook forces.** Today the pre-commit hook fails
   every commit: `make verify-staged` needs `tools`, and
   `bootstrap-tools.sh` exits on `go1.27.2` ("expected go1.27.1, got
   go1.27.2", reproduced on a scratch clone). So:
   1. Phase 1 step 1's change to `bootstrap-tools.sh` is made in the work
      tree first, and left unstaged;
   2. the owner commits the records alone. The hook runs the work tree's
      bootstrap, so the commit passes on its own checks;
   3. Phase 1 continues.

   The record says so.

### Phase 1: the toolchain, the library and the tools (T1, L1, D1)

**Files:**
* `scripts/bootstrap-tools.sh`, `go.mod`, `go.sum`, `Makefile`
* `migration_test.go`, `README.md`

1. **`scripts/bootstrap-tools.sh`:**
   * `GO_VERSION="go1.27.2"`;
   * `GOLANGCI_LINT_VERSION="v2.14.0"`, with the expected string
     `"2.14.0"`;
   * `GOVULNCHECK_VERSION="v1.8.0"`;
   * `ACTIONLINT_VERSION` stays `v1.7.12`.

   Its existing rebuild rule rebuilds each tool, because they were built
   with another Go (`:31-35`).
2. **`go.mod`:**
   * `go get github.com/maccavelli/go-selfupdate-lib@v1.13.0`, then
     `go mod edit -go=1.27.2`, then `go mod tidy`;
   * no `toolchain` line, since one equal to the `go` line is redundant;
   * the diff must be the `go` line and the library's two lines in
     `go.mod`, and the library's two lines in `go.sum` (as probed in the
     MADR).
3. **`Makefile:1`:** `MOD_VERSION := 1.27.2`.
4. **The schema test, seen failing first.** Before step 5, with the
   library at `v1.13.0`, `go test -run TestUpdateCheckJSONSchema .` must
   fail with `schema_version 4; want result, 2`, as the MADR's probe
   recorded. Then `wantResultSchema` becomes 4.
5. **`README.md`:**
   * `:484` says Go 1.27.2;
   * `:370` says `"schema_version":4`;
   * a section "Changes with go-selfupdate-lib v1.13.0", after "Changes
     with go-llmprovider-sdk v1.3.2", in its form:
     * the toolchain (Go 1.27.2, and the advisories in 1.27.1);
     * schema 4 (`rolled_back`, `probes_skipped`,
       `replaced_before_stop`);
     * a backup an interrupted update leaves is kept as
       `.prepare-commit-msg.selfupdate-kept-<n>` beside the hook.
6. **Checks** (rule 2):
   * `make tools` rebuilds the three tools with go1.27.2. Each tool's
     `go version` names it;
   * `make verify` passes, with govulncheck "No vulnerabilities found.";
   * the same on the Linux test host. On the Windows test host,
     `go test ./...` and `go vet`, as CI's native job runs them.
7. **Staged** for the owner's commit.

### Phase 2: the spec, `identity`, and the workflows (B1, P1)

**Files:**
* `selfupdate-release.json` (new), `update.go`, `main.go`
* the tests (`main_test.go`, and a new `releasespec_test.go`)
* `.github/workflows/ci.yml`, `Makefile`, `scripts/verify-release.sh`
  (deleted)
* `README.md`
* `docs/decisions/0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md`

1. **`selfupdate-release.json`,** at the root, the main package's
   directory:

   ```json
   {
     "schema": 1,
     "products": [
       {"name": "prepare-commit-msg", "package": ".", "identity_args": ["identity"]}
     ],
     "platforms": [
       {"os": "linux", "arch": "amd64"}, {"os": "linux", "arch": "arm64"},
       {"os": "darwin", "arch": "amd64"}, {"os": "darwin", "arch": "arm64"},
       {"os": "windows", "arch": "amd64"}, {"os": "windows", "arch": "arm64"}
     ],
     "packaging": "binary"
   }
   ```

   These are the six platforms `ci.yml`'s `platforms-json` and
   `update.go:39` list today. The asset names stay
   `prepare-commit-msg-<os>-<arch>[.exe]`, so `v1.7.0` clients find the
   release.
2. **`update.go`:**
   * `//go:embed selfupdate-release.json`;
   * `releasespec.Parse`, `spec.Product("prepare-commit-msg")`, and
     `spec.AssetSelector()` replace `NewExactAssetSelector` and its
     hand-written list.

   A parse or product error fails `newUpdateUpdater`, as a selector error
   does now.
3. **`main.go`:**
   * `case "identity":` prints `buildIdentity()` alone, with `fmt.Println`;
   * the help text and the README's CLI reference gain the line;
   * `version` is unchanged.
4. **Tests,** each seen failing first on a scratch copy:
   * **`TestReleaseSpec`:** the embedded spec parses, names the product,
     and its targets are the six platforms. Plant: a misspelled key in the
     copy's JSON, which must fail the parse;
   * **`TestIdentityCommand`:** `identity`'s whole output is
     `buildIdentity().String() + "\n"`. Plant: `identity` printing
     `version`'s line;
   * the existing update tests pass with the spec's selector.
5. **`ci.yml`:**
   * **The `go` job:** loses its tag steps ("Build release binaries",
     "Upload release artifacts") and its `outputs`. It runs `make verify`
     on every ref;
   * **`go-native`:** unchanged;
   * **A new `build` job:** `needs: [go, go-native]`, `permissions:
     contents: read`,
     `uses: maccavelli/go-selfupdate-lib/.github/workflows/build-selfupdate-release.yml@5e199c831b5691ea687943e3c3fd495d50c739ed # v1.13.0`,
     with `spec-path: selfupdate-release.json`;
   * **`release`:** `needs: build`,
     `if: needs.build.outputs.rehearsal == 'false'`. It calls the publish
     workflow at the same commit with the build job's five outputs, as the
     library's building guide §4 shows. The comment above it names both
     records and the amended rule.
6. **`Makefile`:** `release-artifacts` and `verify-release` go, with
   `scripts/verify-release.sh`. The build workflow's checks replace them.
   `build`, `build-all` and `verify` stay.
7. **`README.md`,** "Maintainer Release": a tag's CI builds through the
   library's build workflow, which checks each binary and runs `identity`
   on five platforms, then publishes. `make verify-release` is gone.
8. **0010-MADR,** an amendment under D2: the rule reads "the newest
   workflow commit whose publish path, the workflow file and the scripts it
   runs, is unchanged from one that has published live". The evidence is
   0012-MADR's: the same blob `a0c842a8…` at `v1.11.0` and `v1.13.0`.
9. **Checks:** rule 2, with rule 8, `make workflow-lint`. The rehearsal
   needs a push, so it is Phase 5 step 1's.
10. **Staged.**

### Phase 3: installers (I1)

**Files:** `selfupdate-release.json`, the spec test, `README.md`.

1. **The spec** gains `"installer": {"name": "prepare-commit-msg"}`. Its
   `env_prefix` is the default, `PREPARE_COMMIT_MSG`.
2. **`TestReleaseSpec`** also requires `InstallerScripts()` to be
   `install.sh` and `install.ps1`. Plant: the field removed.
3. **`README.md`, "Download or Build the Binary" and "Install as a Git
   Hook":**
   * **Unix:**

     ```sh
     curl -fsSL https://github.com/maccavelli/prepare-commit-msg/releases/latest/download/install.sh | sh -s -- --dir ~/.global-git-hooks
     git config --global core.hooksPath ~/.global-git-hooks
     ```

   * **Windows:** the scriptblock form with
     `-InstallDir "$env:USERPROFILE\.global-git-hooks" -NoPathUpdate`,
     then the same `git config`.
   * **The text says:**
     * the installer checks `SHA256SUMS` and the identity, and keeps
       `<name>.prev`;
     * a bare one-liner installs into `~/.local/bin`, which Git does not
       run as a hook;
     * the installer never sets `core.hooksPath`, because the hooks
       directory can hold other hooks;
     * `--uninstall` removes it, and `--verify-attestation` checks the
       attestation with `gh`.
   * Building from source and `make install` stay.
4. **Checks:** rule 2. Staged.

### Phase 4: hygiene (H1)

1. **`README.md`'s attestation check** gains
   `--signer-workflow maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml`.
2. **`git rm testfile.txt`.** 0009-PLAN and 0010-PLAN, which left it "for
   the owner", get no edit: their text records why it was kept then.
3. **Checks:** `go test ./...`, and the README test. Staged.
4. **`scripts/configure-github.sh`** (0012-MADR A1): the workflow list
   at `:79-82` and the loop at `:103` name `ci.yml` only, in place of
   `quality.yml`, `ci.yml` and `release.yml`. Nothing else in the script
   changes. Seen failing first: on a scratch clone, the unchanged script
   run without `--apply` reports `remote_ready` false. Checks:
   `make workflow-lint`, which covers shell scripts. Staged.
5. **The owner,** after Phase 5 step 1's push, runs the script without
   `--apply` to read the plan (`remote_ready` must be true), then with
   `--apply`.
   * The script still refuses `--apply` while `ci.yml` differs from
     `origin/main` or has uncommitted changes (`:236-239`).
   * The agent then checks, read-only: `GET /repos/…/rulesets` lists
     `prepare-commit-msg-main` and `prepare-commit-msg-release-tags`.
   * This must hold before the tag.

### Phase 5: release `v1.8.0`, and the live checks

1. **The owner** commits and pushes Phases 1–4. The agent checks, read-only:
   * CI on `main` is green;
   * its `build` job ran the rehearsal, with five identity legs printing
     `rehearsal-<commit> (local)`;
   * `release` was skipped.
2. **Phase 4 step 5,** the rulesets.
3. **The owner** tags `v1.8.0` (annotated) and pushes it.
4. **The agent checks, read-only:**
   * the tag's CI: every job passes, and the identity legs print
     `v1.8.0 (release) <12-hex>`;
   * the release: immutable, latest, not a draft, with six binaries,
     `SHA256SUMS`, `install.sh` and `install.ps1`;
   * `gh attestation verify prepare-commit-msg-linux-amd64 --repo
     maccavelli/prepare-commit-msg --signer-workflow …` passes.

   A failed publish is a deviation: the tag stays, and the fix is
   `v1.8.1`.
5. **The owner** runs `prepare-commit-msg update` on the installed hook on
   each host.
6. **The agent checks the installed hook** (0011-PLAN Phase 3 step 5,
   carried):
   * `version` and `identity` name `v1.8.0`;
   * its build information names `go1.27.2`, SDK `v1.3.2` and
     go-selfupdate-lib `v1.13.0`;
   * `govulncheck -mode=binary` on it reports no vulnerabilities;
   * `update --check` reports it up to date;
   * in a scratch repository, a staged change gets a message from the
     owner's configured provider;
   * with `GEMINI_API_KEY` set to a literal that is not a key and two
     Gemini models configured, the hook stops at the first model with
     "authentication failed for gemini".
7. **The installers,** on this Mac and the Linux test host:
   * with `HOME` a scratch directory, the README's one-liner with
     `--dir <scratch>/.global-git-hooks` installs `v1.8.0`, its identity
     checked;
   * `--uninstall` removes it.

   On the Windows test host, the scriptblock form with a scratch
   `-InstallDir` and `-NoPathUpdate`, then `-Uninstall`. The user's real
   hooks directory is not touched.

### Phase 6: close-out

1. The Execution Record holds every phase's output and deviations.
2. 0012-MADR stays `accepted`. This PLAN is `complete`, and
   `docs/README.md` says so.
3. Staged for the owner's commit.

## Verification

* **V1 (T1, D1):**
  * `go.mod` says `go 1.27.2`;
  * the three tools report `go1.27.2` builds, at `v2.14.0`, `v1.8.0` and
    `v1.7.12`;
  * `make verify` and `make verify-staged` pass on this Mac and the Linux
    test host;
  * the pre-commit hook passes a commit again.
* **V2 (L1):** the `go.mod`/`go.sum` diff is as Phase 1 step 2 says.
  `TestUpdateCheckJSONSchema` failed at 2 and passes at 4.
  `govulncheck ./...` is clean.
* **V3 (B1):**
  * `TestReleaseSpec` and `TestIdentityCommand` failed on their plants and
    pass;
  * `update.go` holds no platform list;
  * the rehearsal on `main` passes with five identity legs.
* **V4 (P1):** both `uses:` name `5e199c83…`, and 0010-MADR D2 carries the
  amendment.
* **V5 (I1):** the release carries both installers, and the one-liners
  install and uninstall on three hosts into scratch directories.
* **V6 (H1):**
  * the README's check names the signer workflow;
  * `testfile.txt` is not tracked;
  * both rulesets are listed.
* **V7 (the release):**
  * `v1.8.0`'s CI, release and attestation hold (Phase 5 step 4);
  * the installed hook is `v1.8.0`, built with `go1.27.2`, and
    `govulncheck -mode=binary` is clean;
  * 0011's live checks pass.

## Rollout and Rollback

* **Rollout.** Phases 1–4 are committed by the owner and pushed. Then the
  rulesets, the `v1.8.0` tag, and `update` on each installed hook.
* **Rollback, before the tag:** each phase reverts alone.
  * Reverting Phase 1 returns the toolchain, and the hook's commit failure
    with it.
  * Reverting Phase 2 restores the Makefile build, `verify-release.sh` and
    the `v1.10.0` pin together.
* **Rollback, after the tag:** fix forward in `v1.8.x`. A user who returns
  to `v1.7.0` keeps a working hook, built with Go 1.27.1.

## Execution Record

### Approval (2026-10-09)

The owner answered 0012-MADR's nine questions with the recommendations,
and Dependabot "Leave off". They then said "prepare-commit-msg should be
using go 1.27.2 for everything now", and approved this PLAN: "proceed".

### Phase 0: records (2026-10-09)

1. `docs/README.md` indexes both 0012 records: the MADR `Accepted`, this
   PLAN `In Progress`. 0011-PLAN is `Superseded`.
2. 0011-PLAN gains "Superseded (2026-10-09)", and `status: superseded`.
3. **The hook.** The owner's first commit attempt failed. Reproduced on a
   scratch clone: `make verify-staged` exits 2 with
   `expected go1.27.1, got go1.27.2`, from `make tools`. As step 4
   planned:
   * Phase 1 step 1's change to `scripts/bootstrap-tools.sh` is in the
     work tree, unstaged;
   * the records are staged alone, for the owner's commit.

### Phase 1: the toolchain, the library and the tools (2026-10-09)

1. **`scripts/bootstrap-tools.sh`** (step 1) set `GO_VERSION="go1.27.2"`,
   golangci-lint `v2.14.0` (expected `"2.14.0"`) and govulncheck `v1.8.0`;
   actionlint stays `v1.7.12`. Made in Phase 0, as its step 4 planned. The
   owner's records commit, `332bb70`, included it with the records. So it
   is on `main` already, and this phase stages the rest.
   * `make verify-staged` then rebuilt the three tools. `go version` on
     each in `.tools/bin` gives `go1.27.2`.
2. **`go.mod`** (step 2): `go get …@v1.13.0`, `go mod edit -go=1.27.2`,
   `go mod tidy`. The diff is:
   * `go 1.27.1` → `go 1.27.2`;
   * `go-selfupdate-lib v1.10.1` → `v1.13.0`;
   * in `go.sum`, the library's two lines.

   There is no `toolchain` line.
3. **`Makefile:1`** (step 3): `MOD_VERSION := 1.27.2`.
4. **Seen failing first** (step 4). With the library at `v1.13.0` and the
   test unchanged:

   ```text
   --- FAIL: TestUpdateCheckJSONSchema (0.00s)
       migration_test.go:186: last stdout line is kind "result", schema_version 4; want result, 2
   ```

   `wantResultSchema` is now 4, with its comment naming 0012-MADR L1.
5. **`README.md`** (step 5):
   * `:370` says `"schema_version":4`;
   * "Building from source needs Go 1.27.2";
   * a new section, "Changes with go-selfupdate-lib v1.13.0", in the table
     of contents. It covers Go 1.27.2 and `update`, schema 4, and the
     backup an interrupted update leaves.
   * **One line not in the step:** "Changes with go-llmprovider-sdk
     v1.3.2" said release `v1.8.0` moves go-selfupdate-lib "to `v1.10.1`".
     That no longer holds, so it now says "to `v1.13.0`, whose changes are
     in the next section".
6. **Checks** (step 6, rule 2):

   | Check | Host | Result |
   | :--- | :--- | :--- |
   | `make verify` | this Mac | exit 0: `0 issues.`, `total coverage: 81.9% (minimum 80.0%)`, `No vulnerabilities found.` |
   | `make verify` | the Linux test host, `go1.27.2`, on a copy of the work tree | exit 0: `0 issues.`, `total coverage: 81.8% (minimum 80.0%)`, `No vulnerabilities found.` |
   | `go test -count=1 ./...`, `go vet ./...` | the Windows test host, `go1.27.2` | five `ok`; vet exit 0 |
   | `go test -race -count=1 ./...` | this Mac | exit 0 |
   | `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...` | this Mac | exit 0, 0, 0 |
   | `go mod tidy -diff` | this Mac | exit 0 |
   | `gofmt -l migration_test.go` | this Mac | empty |

7. **Staged:** `go.mod`, `go.sum`, `Makefile`, `migration_test.go`,
   `README.md` and this PLAN. `make verify-staged` on the staged snapshot:
   exit 0. Both libraries resolved from GitHub (go-selfupdate-lib `v1.13.0`,
   `h1:5gMIuQvC…`), golangci-lint `0 issues.`, govulncheck
   `No vulnerabilities found.`
