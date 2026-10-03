---
status: in-progress
date: 2026-10-03
associated-madr: "0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md"
---
# Implement dropping mcplib for go-llmprovider-sdk v1.0.0 and go-selfupdate-lib v1.5.0

Associated MADR: [0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md](0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md)

## Goal

* `go.mod` requires `go-llmprovider-sdk` `v1.0.0` and `go-selfupdate-lib`
  `v1.5.0` (or newer `v1.x`), and `go list -m all` names no mcplib.
* `update` and `version` are go-selfupdate-lib's canonical command and
  build stamps.
* Every behaviour change in the MADR's table is either intended and
  documented in `README.md`, or kept by D2–D4.
* go-llmprovider-sdk's 0002-PLAN Phase 10 is satisfied, its live check
  included.

Done means every item under Verification holds, CI is green on Linux,
macOS and Windows, and a tag built from this work passes its release
verification.

## Scope

### In scope

| Phase | Who | What |
| :--- | :--- | :--- |
| 0 | owner, then agent | answer D1–D7, accept the MADR |
| 1 | agent | Go 1.27.1, and self-update from go-selfupdate-lib |
| 2 | agent | providers and the wizard from go-llmprovider-sdk, with D2–D5 |
| 3 | agent | the supply-chain gate, mcplib removed, D6 |
| 4 | agent | `README.md`, `docs/README.md`, 0007's amendment (D7) |
| 5 | owner, then agent | push, CI, tag, release verification, the live check |
| 6 | agent | close-out here and in go-llmprovider-sdk's 0002-PLAN |

### Out of scope

* **Any change in go-llmprovider-sdk, go-selfupdate-lib or mcplib** other
  than Phase 6's record entry in go-llmprovider-sdk.
* **0007's link and index work** (L-a, I-a). It stays under 0007.
* **Historical records.** Their mcplib and Go 1.26.6 text stays as
  written.
* **Push and tags,** which are the owner's.

## Rules for every phase

1. **Order.** Phases run in order. Each ends green and in its own commit.
2. **Commits.** `git commit --no-edit`, after the owner authorizes commits
   to `main` in that turn. The hooks chain to the global
   `prepare-commit-msg` hook.
3. **Checks before each commit that changes code:**
   * `make verify` (lint, coverage at least 80.0%, govulncheck, workflow
     lint, build-all);
   * `python3 scripts/go-precheck.py`;
   * `go test -race -count=1 ./...`;
   * `CGO_ENABLED=0 GOOS={linux,darwin,windows} go vet ./...`;
   * `go mod tidy -diff`.
4. **Proofs on scratch copies,** never in the tree. Each new check or test
   is seen failing on a planted break.
5. **Session tooling is Python.** The repository's own scripts and `make`
   targets run as they always do.

## Implementation Steps

### Phase 0: accept the records

1. The owner answers D1–D7, or accepts the recommendations. The answers go
   into the MADR, which becomes `accepted`, and this PLAN `in-progress`.
2. `docs/README.md` indexes both records.
3. Commit the records.

### Phase 1: Go 1.27.1 and self-update

1. **Go 1.27.1.** In `go.mod`, `go 1.26.6` becomes `go 1.27.1`. The
   Makefile's `MOD_VERSION` and `scripts/bootstrap-tools.sh`'s
   `GO_VERSION` follow. *(Done first, as its own commit `e735fa3`:
   deviation D1 below.)*
   * **1a. Proposed 2026-10-03, not approved.** `bootstrap-tools.sh`
     reinstalls a cached tool when the Go toolchain that built it is not
     `GO_VERSION`, read with `go version <binary>`. Proof: a cache built
     with another toolchain is rebuilt, and an up-to-date one is not.
2. **The module.** `go get github.com/maccavelli/go-selfupdate-lib@<newest
   v1.x>`, at `v1.5.0` or later, then `go mod tidy`. mcplib stays required
   for `llmprovider` and `wizard`.
3. **`update.go`.** It keeps the updater factory, `newUpdateUpdater`, built
   on `selfupdate.UserAgent`, `DiscardReporter` and
   `NonInteractiveConfirmer`. It adds:
   * `updateOptions`, defaulting to `cli.StdioOptions`;
   * `buildIdentity = buildinfo.Identity`, the seam.
4. **`main.go`.**
   * `update` is `cli.Command(ctx, args[1:], AppTitle, buildIdentity(),
     newUpdateUpdater, updateOptions())`.
   * `version` prints `buildIdentity().String()` (D1).
   * `Version` and `displayVersion` go, and the usage text names the new
     flags.
5. **`Makefile`.** `build` and `RELEASE_LDFLAGS` stamp
   `github.com/maccavelli/go-selfupdate-lib/buildinfo.version` and `.kind`.
   The `main.Version`, `main.RawVersion` and `main.RawBuildKind` stamps go.
6. **`scripts/verify-release.sh`** accepts `version vX.Y.Z (release)` with
   or without a 12-hex revision. It refuses `-dirty`, another version,
   `(local)`, and the old form without the `v`.
7. **`.github/workflows/ci.yml`.** `uses:` is
   `maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml`,
   pinned to the commit of the go-selfupdate-lib release chosen in step 2.
   That commit is resolved with `git ls-remote`, peeled, and must carry
   `publish-selfupdate-release.yml` unchanged from `58411f1`, or have its
   inputs re-read. The `bridge-release` input goes.
8. **Tests.**
   * `update_test.go` and `TestRunUpdate_Flags` (in `main_extra2_test.go`)
     are removed.
   * `migration_test.go` is new, with go-selfupdate-lib's
     `selfupdate/cli/testdata/migration/` fixtures copied to
     `testdata/migration/` at the chosen release. It holds:
     * `TestMigrationByteForByte`;
     * `TestUpdateRefusedBeforeBuild`;
     * `TestVersionPrintsIdentity`;
     * `TestMakefileStamps`, which skips without `make`.
   * A test covers the default `newUpdateUpdater`.
9. **Proofs** (scratch copy), each of which must fail:
   * a go-selfupdate-lib copy whose summary reads `up-to-date` fails only
     that subtest;
   * a banner written to stdout fails all three;
   * a Makefile without the `.kind` stamp fails `TestMakefileStamps`;
   * `verify-release.sh` is checked on six inputs: two it accepts and four
     it refuses.
10. Run the checks, then commit.

### Phase 2: providers and the wizard

1. **The module.** `go get github.com/maccavelli/go-llmprovider-sdk@v1.0.0`,
   then `go mod tidy`.
2. **Imports and calls,** mapped through go-llmprovider-sdk's
   `docs/guides/migrating-from-mcplib.md`, in:
   * `internal/config/config.go` and its test;
   * `internal/ui/setup.go` and its test;
   * `main.go`;
   * `main_test.go`;
   * `main_oauth_test.go`.

   The configuration keeps string provider IDs and converts at its
   boundary. The seams `newProvider`, `newProviderWithSource` and
   `generateWithRetry` keep their names. A `registry =
   providers.Default()` variable serves descriptors and the wizard.
3. **Generation.** `generateText` sends the prompt as one user
   `MessageItem`. It does so through `WithRetry`, with `RetryPolicy{MaxAttempts:
   retries+1, BaseDelay: delay}`. It gets a test on `llmtest` covering:
   * a success;
   * a rate limit retried, then a success;
   * an auth error not retried.
4. **D2.** `ResolveAPIKey` reads `ANTHROPIC_API_KEY`, then `CLAUDE_API_KEY`,
   for Claude. A test covers each, and both set.
5. **D3.** The configuration gains the Kilo organization. The wizard's
   result saves it. Generation and listing pass it with
   `kilo.WithOrganization` and `catalog.WithKiloOrganization`. A test
   covers each, and an existing configuration file without the field loads
   unchanged.
6. **D4.** `catalog.OptionsFromEnv()` is passed where providers and the
   catalog are built.
7. **D5, the folded-in P8.**
   * `ValidateOAuth` calls `auth.ValidateOAuthSession` at load.
   * The two tests it fails get fixtures holding valid sessions:
     `TestRunAnalyzer_OAuth…` and `TestValidateActive_OAuthWithoutKeyOK`.
   * `main_oauth_test.go` isolates the live token store on every OS,
     Windows' `%AppData%` included.
8. **Phase 10 step 7.** `orchestrated := false` and `Orchestrated:
   &orchestrated` are removed.
9. **Comments.** Citations of mcplib records stay. A comment that states
   the current dependency is corrected.
10. **Proofs:** each new test is seen failing on a planted break in a
    scratch copy.
11. Run the checks, then commit.

### Phase 3: the supply-chain gate

1. `scripts/go-precheck.py`'s mcplib check becomes a check of
   `go-llmprovider-sdk` and `go-selfupdate-lib`. Each must be:
   * required at a release version, not a pseudo-version;
   * free of any `replace`;
   * free of any `GOPRIVATE`, `GONOSUMDB` or `GOINSECURE` exemption;
   * matched by `go.sum`, checked with `go mod download`.
2. It also refuses a module graph that names
   `github.com/maccavelli/mcplib`.
3. **D6.** The `GONOSUMCHECK` arm is removed.
4. `go.mod` no longer requires mcplib; `go mod tidy -diff` is clean.
5. **Proofs** (scratch copy), each of which must fail with its own message:
   * a `replace` of either module;
   * a pseudo-version;
   * `go-selfupdate-lib` not required;
   * mcplib re-imported;
   * a tampered `go.sum` line;
   * each exemption variable set.

   The unchanged copy must pass.
6. Run the checks, then commit.

### Phase 4: documentation

1. **`README.md`:**
   * the provider count and engine (`:66`);
   * the Claude key row (`:92`), as D2 states it;
   * the self-update table and usage block (`:355`–`:375`): the new flags,
     stderr for progress, the `version` form;
   * a short "Changes from mcplib" note: retries, error prefix, removed
     Grok overrides, `LLMPROVIDER_*`.
2. **`docs/README.md`** indexes 0008 with its status.
3. **0007** gains an amendment. 0008 supersedes its dependency half (D-a);
   its L-a and I-a stand.
4. **Checks:**
   * every relative link in the changed files resolves, proven on a planted
     bad link;
   * `git grep -n mcplib` over `README.md` shows only historical mentions;
   * markdownlint if the repository configures it;
   * `git diff --check`.
5. Commit.

### Phase 5: release and the live check (owner, then agent)

1. **The owner** pushes. CI runs on Linux, macOS and Windows.
2. **The owner** tags the next minor release. On the tag:
   * CI builds;
   * `verify-release.sh` passes;
   * the reusable workflow publishes.
3. **The agent** checks each of these, and records the output:
   * the release assets, and their `SHA256SUMS`;
   * `update --check` from the previous release finds the new one;
   * the new binary's `version` prints the D1 form.
4. **Phase 10 step 8, the live check.** Built from the tag, a
   ChatGPT-session model listing shows `gpt-6-sol`. The owner runs it with
   their session, and the agent records the output.

### Phase 6: close-out

1. This PLAN is `complete`, and `docs/README.md` says so.
2. go-llmprovider-sdk's
   `docs/decisions/0002-PLAN-migrate-llmprovider-from-mcplib.md` gains a
   Phase 10 execution entry, with the live check's output. That is a
   records-only change there.

## Verification

* **V1.** `go list -m all` names no `github.com/maccavelli/mcplib`.
  `go.mod` requires the two libraries at release versions, and `go 1.27.1`.
* **V2.** Every check in rule 3 passes at the end of Phases 1, 2 and 3.
* **V3.** `TestMigrationByteForByte` passes, and fails on each planted
  break of Phase 1 step 9.
* **V4.** The precheck refuses each planted input of Phase 3 step 5.
* **V5.** D2–D5 each have a passing test that fails on a planted break.
* **V6.** CI is green on Linux, macOS and Windows, on `main` and on the tag.
  The tag's release passes `verify-release.sh`.
* **V7.** The live check shows `gpt-6-sol`.
* **V8.** Nothing committed carries a hostname, an account name or a
  real-machine path.

## Rollout and Rollback

* **Rollout.** Phases 1–4 are local commits. Phase 5 publishes.
* **Rollback.**
  * Before the tag, any phase reverts alone, newest first, because each is
    green alone.
  * After the tag, a problem is fixed forward in a patch release. Users on
    the previous release keep working: their `update` reads the same
    GitHub releases.

## Execution Record

The trial behind the MADR's measurements was a scratch clone; nothing was
committed from it.

### Phase 0: accept the records (2026-10-03)

* **Approval.** The owner answered "D1-D7 follow recommendations. proceed",
  and authorized a commit per phase to `main`, with no push.
* The MADR is `accepted`, and this PLAN `in-progress`.
* `docs/README.md` indexes both. Its older `file://` links are 0007's L-a
  work, out of this PLAN's scope, and are left as they are.

### Deviation D1 (2026-10-03): the toolchain bump lands before Phase 0

* **Found.** The Phase 0 commit was refused by the repository's
  pre-commit hook.
  * `.githooks/pre-commit` runs `make verify-staged`, which runs `make
    tools`.
  * `scripts/bootstrap-tools.sh:10` pinned `GO_VERSION="go1.26.6"`. The host
    runs go1.27.1, so the hook stopped with `expected go1.26.6, got
    go1.27.1`.
  * This was pre-existing: the MADR's trial saw the same failure at
    `cfada6e`. No commit can pass the hook on this host until the pin moves.
* **Decision.** The owner chose "Bump toolchain first". Phase 1 step 1 lands
  alone, then the Phase 0 records, then the rest of Phase 1.
* **What the bump needed.**
  * **`go mod tidy`.** At `go 1.27.1`, `go mod tidy -diff` regrouped
    `go.mod`'s `require` blocks. It moved `mcplib` into its own block and
    `x/term` beside `x/sys`. The requirements and versions are unchanged,
    and `go.sum` is unchanged.
  * **The tool cache.** `.tools/bin` held golangci-lint v2.13.1,
    govulncheck v1.7.0 and actionlint v1.7.12, all built with go1.26.6.
    golangci-lint refused the module: `the Go language version (go1.26)
    used to build golangci-lint is lower than the targeted Go version
    (1.27.1)`.
    * `bootstrap-tools.sh` compares only each tool's own version, so it
      never rebuilds them.
    * The three were rebuilt in this host's ignored cache, at the same
      pinned versions, with `GOBIN=.tools/bin go install`. `go version` then
      reports go1.27.1 for each.
    * CI starts from an empty cache, so it is unaffected.
    * Step 1a proposes the fix in the tree.
* **Checks on `e735fa3`:**

  | Check | Result |
  | :--- | :--- |
  | `make verify` | exit 0: 0 lint issues; `total coverage: 83.5% (minimum 80.0%)`; `No vulnerabilities found.`; build-all |
  | `go test -race -count=1 ./...` | 5 packages ok |
  | `CGO_ENABLED=0 go vet` for linux, darwin, windows | 0 each |
  | `go mod tidy -diff` | 0 |
  | `python3 scripts/go-precheck.py` | exit 0; mcplib `v1.6.0` resolved from GitHub |
  | the pre-commit hook | passed |

### Phase 1: Go 1.27.1 and self-update (2026-10-03)

* **Step 1** landed first, as `e735fa3` (deviation D1).
* **Steps 2–8.** The trial's self-update tree was the source. Each file it
  replaced was first checked to be identical to `cfada6e` in this tree.
  * `go-selfupdate-lib` `v1.5.0` is required, its newest release. mcplib
    `v1.6.0` stays for `llmprovider` and `wizard`.
  * `update.go` holds `defaultNewUpdateUpdater`, `buildIdentity`,
    `newUpdateUpdater` and `updateOptions`.
  * In `main.go`, `update` is one `cli.Command` call, and `version` prints
    `buildIdentity()`. The usage text appends `cli.HelpText`.
  * The `Makefile` stamps `buildinfo.version` and, for releases,
    `buildinfo.kind`.
  * `verify-release.sh` takes D1's form.
  * `ci.yml`'s `uses:` is go-selfupdate-lib's workflow at `6deaa52`
    (`v1.5.0`). That file is identical to `58411f1`'s, whose inputs the
    trial read. `bridge-release` is gone.
  * `update_test.go` and `TestRunUpdate_Flags` are removed.
  * `migration_test.go` and the nine fixtures under `testdata/migration/`
    are added. Beyond the trial, it adds `TestDefaultNewUpdateUpdater`.

**Deviation D2 (2026-10-03): the test now sees the process's stdout.**

* **Found.** Step 9's banner proof printed a line with `fmt.Println` before
  `cli.Command`. The tests still passed. `runMain` captured only the update
  command's own stdout stream, so a stray write to the process's stdout,
  which would corrupt `--json` output, went unseen.
* **Fix, within this phase.** `runMain` also captures the process's
  stdout, through a pipe it restores in `t.Cleanup`. It returns both, the
  process output first. `TestVersionPrintsIdentity` uses that capture,
  instead of a pipe of its own. No assertion was loosened.

**Proofs** (scratch copies; `pcm_phase1_proofs.py`):

| Planted | Result |
| :--- | :--- |
| none (control) | exit 0 |
| go-selfupdate-lib `v1.5.0` copy, through `replace`, summary `up-to-date` | exit 1: `TestMigrationByteForByte/up-to-date` only |
| `fmt.Println("banner")` before `cli.Command` | exit 1: all three byte-for-byte subtests and `TestUpdateRefusedBeforeBuild` |
| the `Makefile` without the `buildinfo.kind` stamp | exit 1: `TestMakefileStamps` |
| the factory with an empty repository name | exit 1: `TestDefaultNewUpdateUpdater` |
| `verify-release.sh`, six versions | the two forms with and without a revision accepted; `-dirty`, `v1.2.4`, `(local)` and the form without `v` refused |
| the same script with a catch-all `*) ;;` first | exit 1 |

**Checks:**

| Check | Result |
| :--- | :--- |
| `make verify` | exit 0: 0 lint issues; `total coverage: 83.5% (minimum 80.0%)`; `No vulnerabilities found.`; build-all |
| `go test -race -count=1 ./...` | 5 packages ok |
| `CGO_ENABLED=0 go vet` for linux, darwin, windows | 0 each |
| `go mod tidy -diff` | 0 |
| `python3 scripts/go-precheck.py` | exit 0 |
| actionlint v1.7.12 on `ci.yml`; `shellcheck` and `bash -n` on `verify-release.sh` | 0 each |
| `gofmt -l .` | empty |

### Phase 2: providers and the wizard (2026-10-03)

* **From the trial.** The trial's provider changes were the source, in:
  * `main.go` and `main_test.go`;
  * `main_oauth_test.go`;
  * `internal/config/config.go` and its test;
  * `internal/ui/setup.go` and its test.

  Each file was first checked to be identical to the trial's self-update
  tree, which is this repository's Phase 1. `go-llmprovider-sdk` `v1.0.0` is
  required.
* **mcplib left the module graph in this phase.** Phase 1 removed its
  `selfupdate` import, and Phase 2 its last imports, so `go mod tidy`
  dropped the requirement. `go list -m all` names no mcplib. The precheck
  still downloads mcplib `v1.6.0` by name and passes. Phase 3 replaces that
  check.
* **D2.** `config.LookupEnv` reads `CLAUDE_API_KEY` when `ANTHROPIC_API_KEY`
  is unset. `ResolveAPIKey` and the wizard's `LookupEnv` use it.
* **D3.** `ProviderConfig.Organization` (`omitempty`) holds the
  organization.
  * The wizard's result saves it, and configure passes it back to the
    wizard as the existing value.
  * The non-interactive API-key path clears it.
  * `providerOptions` adds `kilo.WithOrganization` for Kilo only. Listing
    needs nothing here (MADR amendment A1).
* **D4.** `catalog.OptionsFromEnv()` is passed where providers, the
  catalog and the wizard are built.
* **D5, the folded-in P8:**
  * `ValidateOAuth` calls `auth.ValidateOAuthSession`, and wraps its error
    with the configure hint;
  * `isolateHome`, the ui `isolate` and the new `isolateUserDirs` in
    `main_oauth_test.go` set `HOME`, `USERPROFILE`, `XDG_CONFIG_HOME`,
    `APPDATA`, `AppData` and `LOCALAPPDATA`;
  * `TestValidateActive_OAuthWithoutKeyOK` holds a refreshable session;
  * the analyzer OAuth test holds an access-only ChatGPT session with its
    own fixture token, and fails if the live session file ever holds it.
* **Step 7.** `Orchestrated` is gone. Step 9: the drift-guard comment in
  `setup_test.go` names go-llmprovider-sdk. The three citations of mcplib
  MADR 0012 stay.
* **New tests:**
  * `TestGenerateText`, on `llmtest.Fake`: success; a rate limit retried;
    an authentication failure not retried.
  * `TestNewActiveProvider_KiloOrganization`.
  * `TestResolveAPIKey_ClaudeFallback`.
  * `TestProviderConfig_Organization`.
  * `TestValidateOAuth_RejectsChatGPTAccessFixture`.
  * `TestRedirectUserConfig_APPDATARequired`.
  * `TestIsolateHome_RedirectsWindowsUserConfigDir`, which runs on Windows
    only, so it skips on this macOS host. CI's Windows job runs it.

**Proofs** (scratch copies; `pcm_phase2_proofs.py`). Each test passed on
the unchanged copy, then failed on its plant:

| Planted | Failed |
| :--- | :--- |
| no Claude fallback | `TestResolveAPIKey_ClaudeFallback/claude_only` |
| the organization never passed | `TestNewActiveProvider_KiloOrganization` |
| an empty organization written (no `omitempty`) | `TestProviderConfig_Organization` |
| no session validation | `TestValidateOAuth_RejectsChatGPTAccessFixture` |
| the helper without `APPDATA` | `TestRedirectUserConfig_APPDATARequired` |
| `MaxAttempts: 1` | `TestGenerateText/rate_limit_retried` |
| the prompt sent as the assistant | `TestGenerateText/success` |

**Not proven here, and why:**

* **The live-session canary in the analyzer test.** Planting a failure
  would mean letting the test write the live profile, which is what it
  guards against. The isolation it relies on is proven by the `APPDATA`
  plant.
* **The Windows `UserConfigDir` test.** It skips on macOS. It runs, and
  must pass, in CI's `windows-2025` job.

**Checks:**

| Check | Result |
| :--- | :--- |
| `make verify` | exit 0: 0 lint issues; `total coverage: 84.2% (minimum 80.0%)`; `No vulnerabilities found.`; build-all |
| `go test -race -count=1 ./...` | 5 packages ok |
| `CGO_ENABLED=0 go vet` for linux, darwin, windows | 0 each |
| `go mod tidy -diff` | 0 |
| `python3 scripts/go-precheck.py` | exit 0 |
| `go list -m all \| grep -c mcplib` | 0 |
| `gofmt -l .` | empty |

### Deviation D3 (2026-10-03): Phases 2 and 3 are one commit

* **Found.** The Phase 2 commit was refused by the pre-commit hook. `make
  verify-staged` runs `scripts/go-precheck.py` on the staged snapshot, and
  that stopped with `go-precheck: go.mod does not require
  github.com/maccavelli/mcplib`.
  * The earlier precheck run, without staged files, had passed. It looked
    at the working tree.
  * Phase 2 drops mcplib from `go.mod` (see Phase 2 above). The precheck
    that accepts this tree is Phase 3's.
* **Decision.** The owner chose "Merge Phases 2 and 3". Phase 3 ran at
  once, and both are committed together. MADR amendment A2 records it.
  Nothing else in either phase changed.

### Phase 3: the supply-chain gate (2026-10-03)

* **`scripts/go-precheck.py`.** The trial's rewrite was the source, and the
  file was first checked to be identical to `cfada6e`'s.
  * `check_dependencies` runs `check_module` for go-llmprovider-sdk and
    go-selfupdate-lib. Each must be required at a released version, not
    replaced, not exempted, matched by `go.sum`, and downloaded to the
    module cache.
  * It then refuses a module graph that names mcplib.
* **D6.** The `GONOSUMCHECK` arm is removed from the `go env` query and from
  the exemption loop. No mention is left.
* **`go.mod`** no longer requires mcplib, which happened in Phase 2.

**Proofs** (copies of the repository, staged index included;
`pcm_phase3_proofs.py`):

| Case | Result |
| :--- | :--- |
| control (unchanged) | exit 0 |
| `replace` go-llmprovider-sdk with a local copy | exit 1: `go.mod replaces … the dependency must resolve from GitHub` |
| `replace` go-selfupdate-lib with `v1.4.1` | exit 1, the same |
| go-llmprovider-sdk at a pseudo-version | exit 1: `… is a pseudo-version; pin a released tag` |
| go-selfupdate-lib not required | exit 1: `go.mod does not require github.com/maccavelli/go-selfupdate-lib` |
| mcplib imported again | exit 1: `the module graph names github.com/maccavelli/mcplib; it must not be a dependency` |
| go-selfupdate-lib's `go.sum` hash tampered | exit 1: `command failed (exit 1): go mod download -json …` |
| `GOPRIVATE`, `GONOSUMDB`, `GOINSECURE` each set to cover the modules | exit 1 each: `… exempts … from checksum verification` |

**Checks on the Phase 2 and 3 tree:**

| Check | Result |
| :--- | :--- |
| `make verify` | exit 0: 0 lint issues; `total coverage: 84.2% (minimum 80.0%)`; `No vulnerabilities found.` |
| `make verify-staged` | exit 0; both modules `resolved from GitHub` |
| `go test -race -count=1 ./...` | 5 packages ok |
| `CGO_ENABLED=0 go vet` for linux, darwin, windows | 0 each |
| `go mod tidy -diff` | 0 |
| `go list -m all \| grep -c mcplib` | 0 |
