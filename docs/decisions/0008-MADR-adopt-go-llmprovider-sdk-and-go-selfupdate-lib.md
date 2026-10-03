---
status: accepted
date: 2026-10-03
decision-makers: Project Owner
consulted: go-llmprovider-sdk (0002-PLAN Phase 10, this record's parent); go-selfupdate-lib (its migration guide §5)
informed: mcplib (it loses this consumer); mcp-server-magictools, mcp-server-magicdev (the consumers that follow)
---
# Drop mcplib: take providers and the wizard from go-llmprovider-sdk v1.0.0, and self-update from go-selfupdate-lib v1.5.0

## Context and Problem Statement

`prepare-commit-msg` requires `github.com/maccavelli/mcplib v1.6.0` for
three packages: `llmprovider`, `wizard` and `selfupdate`. All three have
left mcplib.

* **`llmprovider` and `wizard`** moved to
  `github.com/maccavelli/go-llmprovider-sdk`, released as `v1.0.0`
  (`b8ccb39`). go-llmprovider-sdk
  `docs/decisions/0002-PLAN-migrate-llmprovider-from-mcplib.md` Phase 10
  makes this repository the first consumer to adopt it. Phase 10 names this
  record as its companion.
* **`selfupdate`** moved to `github.com/maccavelli/go-selfupdate-lib`,
  first released under that path as `v1.5.0` (`6deaa52`). The library was
  called go-core-lib until then (go-selfupdate-lib
  `docs/decisions/0009-MADR-rename-to-go-selfupdate-lib.md`). It adds the
  canonical `update` command (`selfupdate/cli`) and build stamps
  (`buildinfo`).
* **The owner decided on 2026-09-29** that `prepare-commit-msg` drops mcplib
  entirely (go-llmprovider-sdk
  `docs/decisions/0002-MADR-migrate-llmprovider-from-mcplib.md`, "The
  owner's further decisions").

How should `prepare-commit-msg` take these three packages from their new
homes, and what does a user of the tool notice?

### What was measured

Measured on 2026-10-03 at `HEAD` `cfada6e`, on a scratch clone, with Go
1.27.1 on macOS. Both modules came from the public proxy, with no
`replace`. Nothing was committed.

| Check | Baseline | Self-update only | Providers only | Both |
| :--- | :--- | :--- | :--- | :--- |
| `go build`, `go vet` (and for linux, darwin, windows) | 0 | 0 | 0 | 0 |
| `go test ./...`, `go test -race ./...` | 0 | 0 | 0 | 0 |
| `go mod tidy -diff` | — | 0 | 0 | 0 |
| modules named mcplib in `go list -m all` | 1 | 1 | 1 | **0** |
| total coverage (minimum 80.0%) | 83.6% | 82.7% | 83.6% | 82.9% |
| golangci-lint, govulncheck | clean | clean | clean | clean |

* **In the "both" tree, the repository's own gates pass.** `make verify`
  and `scripts/go-precheck.py` both exit 0. `go mod verify` prints `all
  modules verified`.
* **The module graph shrinks.** `go.mod` requires the two libraries
  directly. The only indirect requirements left are `golang.org/x/mod`,
  `x/sys` and `x/term`. The MCP go-sdk, `jsonschema-go`, `segmentio` and
  `oauth2` leave the build.
* **The Go floor rises.** `go.mod` goes from `go 1.26.6` to `go 1.27.1`,
  because both libraries require 1.27.1. `Makefile`'s `MOD_VERSION` and
  `scripts/bootstrap-tools.sh`'s `GO_VERSION` follow. Today, at baseline,
  `make tools` already fails on a host pinned to go1.27.1: `expected
  go1.26.6, got go1.27.1`.
* **The self-update part, as go-selfupdate-lib's guide §5 shows:**
  * `update.go` keeps only the updater factory;
  * `main.go`'s `update` case becomes one call to `cli.Command`;
  * `version` prints `buildinfo.Identity()`.

  A byte-for-byte test drives the real `main()` with `update --check`. It
  matched go-selfupdate-lib's three fixtures (up to date, update available,
  failed), with exit codes 0, 10 and 1. Each planted break failed it:
  * a library whose summary reads `up-to-date`;
  * a banner written to stdout;
  * the Makefile's `buildinfo.kind` stamp removed.

  The trial added one test seam, `buildIdentity = buildinfo.Identity`, so
  that the test runs under plain `go test`, as CI runs it.
* **The release workflow.** `ci.yml`'s `uses:` moves from mcplib's
  `publish-selfupdate-release.yml` to go-selfupdate-lib's, at `58411f1`.
  * That workflow declares all four inputs the job passes.
  * It does not declare `bridge-release`, so that input goes.
  * Its permissions match.
  * The file is identical at `58411f1` and `v1.5.0`.
* **`scripts/verify-release.sh`** reads the `version` output of a release
  build, so it changes with that output.
* **The providers part** maps every call through go-llmprovider-sdk's
  `docs/guides/migrating-from-mcplib.md`:
  * provider names become `ProviderID`, converted at the configuration
    boundary, so the configuration file is unchanged;
  * construction goes through `providers.New` with options;
  * text goes through `GenerateText` with `WithRetry`;
  * descriptors come from a registry;
  * tokens come from `llmprovider/auth`;
  * the wizard loses `Orchestrated`, and no longer returns the tokens
    themselves.

  The three test seams keep their names.
* **Behaviour a user of the tool will notice:**

  | Area | Before | After |
  | :--- | :--- | :--- |
  | `update` progress and prompt | stdout | stderr; stdout carries only `--json` |
  | `update` flags | `--check`, `--yes`, `--force`, `--version` | adds `--dry-run`, `--json`, `--channel` |
  | `update` failure line | `Update failed:` | `update failed:` |
  | `version` output | `1.2.3 (release)` | `v1.2.3 (release) <revision>` |
  | providers offered by configure | 9 | 10, adding `together` |
  | Kilo | as mcplib `v1.6.0` offered it | adds two device logins; an OAuth session can now generate |
  | Claude key variable | `CLAUDE_API_KEY`, `ANTHROPIC_API_KEY` | the library reads `ANTHROPIC_API_KEY` only (D2) |
  | retries | every error except auth, invalid request, terminal, or `Retry-After` over 30 s | rate limits (not exhausted quota), service unavailable, failures to connect; the same attempt count, base delay and 30 s cap |
  | retry logging | a `slog.Warn` per retry | none; the library logs nothing |
  | error prefix | `llm:` | `llmprovider:` |
  | outgoing `User-Agent` and vendor identity headers | mcplib | go-llmprovider-sdk |
  | metadata environment overrides | `MCPLIB_*` | `LLMPROVIDER_*`, read only when the caller asks (D4) |
  | Grok OAuth issuer and client environment overrides | honoured during configure | not read |

* **What the trial did not cover:**
  * Phase 10 step 3, the folded-in 0008 P8. The ValidateOAuthSession
    probe failed exactly the two tests P8 predicts.
  * Phase 10 step 8's live check.
  * Anything on Windows.
  * A tag release.
  * Two new functions, `generateText` and the default updater factory,
    are not covered by any test.
* **`GONOSUMCHECK` is not a Go variable.** `scripts/go-precheck.py` refuses
  it today through `go env`, which always reports it empty, so that arm has
  never been able to fire.
* **Record 0007 is proposed and overlaps.**
  [0007-MADR-dependency-and-docs-link-refresh.md](0007-MADR-dependency-and-docs-link-refresh.md)
  would refresh the dependencies that arrive through mcplib. Once mcplib
  leaves, that half is moot. Its link and index halves are not.

## Decision Drivers

* **The owner's decision** that `prepare-commit-msg` drops mcplib
  entirely.
* **One canonical update command across the fleet:** the same flags, exit
  codes and stream rules as every other consumer of go-selfupdate-lib.
* **No silent break for an existing user.** A key that worked yesterday,
  and a release build that verified yesterday, keep working.
* **Each step lands green.** Every commit passes `make verify` and the
  precheck. That is possible because each part was shown to pass alone.
* **The owner prefers a complete, extensible surface now** to cuts made
  only for scope.

## Considered Options

* **A. One companion, in phases:** self-update first, then providers and
  the wizard, then the supply-chain gate. mcplib leaves in the last phase.
* **B. Self-update only.** Leave providers on mcplib for a later record.
* **C. Providers only.** Leave self-update on mcplib.
* **D. Stay on mcplib `v1.6.0`.**

## Decision Outcome

Chosen option: **"A. One companion, in phases"**, because it carries out
the owner's 2026-09-29 decision in the record go-llmprovider-sdk's Phase 10
already names. The trial shows each part passes alone, so each phase can
land and be checked on its own.

### 1. Self-update from go-selfupdate-lib

* Require `github.com/maccavelli/go-selfupdate-lib` at `v1.5.0`, its newest
  release when Phase 1 starts. Re-check the newest `v1.x` then.
* Adopt the canonical command as in go-selfupdate-lib's guide §5:
  * `update` is `cli.Command`;
  * `version` prints `buildinfo.Identity()`, behind a `buildIdentity` seam;
  * the Makefile stamps `buildinfo`'s two `-X` names, which are written
    out.
* `ci.yml` uses go-selfupdate-lib's reusable workflow, pinned to a commit.
  The `bridge-release` input goes.
* **D1. The `version` format** follows `buildinfo`: `v1.2.3 (release)
  <revision>`. `scripts/verify-release.sh` accepts that form, with or
  without a revision, and refuses `-dirty`, a wrong version and `(local)`.
* `update_test.go` and `TestRunUpdate_Flags` go with the code they test.
  `migration_test.go` replaces them, with the fixtures copied under
  `testdata/migration/`, and a test for the default updater factory.

### 2. Providers and the wizard from go-llmprovider-sdk

* Require `github.com/maccavelli/go-llmprovider-sdk` at `v1.0.0`, and map
  every call through its migration guide (Phase 10).
* **D2. The Claude key.** `ResolveAPIKey` keeps reading `CLAUDE_API_KEY`
  after `ANTHROPIC_API_KEY`, so an existing setup keeps working. The README
  names `ANTHROPIC_API_KEY` first and `CLAUDE_API_KEY` as a fallback.
* **D3. The Kilo organization.** A Kilo login that returns an organization
  saves it in the configuration. Generation passes it with
  `kilo.WithOrganization`, and model listing with
  `catalog.WithKiloOrganization`; both exist at `v1.0.0`. Otherwise an
  organization login would quietly use the personal account.
* **D4. Metadata overrides.** `catalog.OptionsFromEnv()` is passed wherever
  a provider or the catalog is built, so `LLMPROVIDER_*` overrides work.
* **D5. Phase 10 step 3 is in scope.** That is:
  * the folded-in 0008 P8;
  * `auth.ValidateOAuthSession` at configuration load, with the two test
    fixtures it fails corrected to hold valid sessions;
  * live-token-store isolation in `main_oauth_test.go`, including Windows'
    `%AppData%`.
* `Orchestrated` is removed, and its behaviour is unchanged (Phase 10
  step 7).
* `generateText` gets a test on the SDK's `llmtest` package.
* Comments that cite mcplib's records keep the citation, since it is
  history. A comment that states the current dependency is corrected.

### 3. The supply-chain gate

* `scripts/go-precheck.py`'s mcplib check becomes a check of the two new
  modules. Each must be:
  * required at a release version, not a pseudo-version;
  * free of any `replace`;
  * free of any `GOPRIVATE`, `GONOSUMDB` or `GOINSECURE` exemption;
  * matched by `go.sum`.
* The check also asserts that the module graph names no
  `github.com/maccavelli/mcplib`.
* **D6.** The dead `GONOSUMCHECK` arm is removed, not repaired, since Go
  has no such variable.

### 4. Record 0007

**D7.** This record supersedes 0007's dependency half (D-a). Its link and
index halves (L-a, I-a) stand as 0007 proposes them. 0007 gets an
amendment saying so.

### Consequences

* Good, because mcplib leaves this repository. Four indirect modules and
  the MCP go-sdk leave the binary.
* Good, because `update` matches the fleet: JSON output, a dry run,
  channels, and stdout kept clean.
* Good, because providers gain `together`, Kilo device logins, and
  generation on any OAuth source.
* Good, because each phase is green on its own, so a later phase can be
  reverted alone.
* Neutral, because coverage falls about 0.7 points, staying above the 80%
  floor.
* Bad, because the Go floor rises to 1.27.1 for anyone building from
  source.
* Bad, because retries become narrower, and quieter. An error mcplib
  retried, such as a plain 500 that is not "unavailable", is now returned
  at once.
* Bad, because scripts that parse `version`, or read `update`'s progress
  on stdout, must change. The README says so.
* Bad, because Grok's issuer and client environment overrides are no
  longer read during configure.

### Confirmation

* Each PLAN phase passes:
  * `make verify`;
  * `python3 scripts/go-precheck.py`;
  * `go test -race ./...`;
  * `go vet` for linux, darwin and windows;
  * `go mod tidy -diff`.
* The byte-for-byte migration test passes, and is seen to fail on a planted
  change.
* The precheck refuses, on a scratch copy:
  * a `replace`;
  * a pseudo-version;
  * a missing requirement;
  * mcplib re-imported;
  * a tampered `go.sum`;
  * each exemption variable.
* CI passes on Linux, macOS and Windows.
* Phase 10 step 8's live check passes: built from the tag, a ChatGPT-session
  model listing shows `gpt-6-sol`. The owner runs it.

## Pros and Cons of the Options

### A. One companion, in phases

* Good, because it is the record go-llmprovider-sdk Phase 10 names, and
  meets the owner's decision fully.
* Good, because the phases land separately, each green.
* Bad, because it is the largest change: about 27 files in the trial.

### B. Self-update only

* Good, because it is small, and the recipe is proven.
* Bad, because mcplib stays, contrary to the owner's decision, and Phase 10
  stays open.

### C. Providers only

* Good, because it closes most of Phase 10.
* Bad, because mcplib stays for `selfupdate`, which also has a new home.

### D. Stay on mcplib `v1.6.0`

* Good, because it is no work.
* Bad, because mcplib's `llmprovider` and `wizard` are frozen (mcplib
  0015), and its `selfupdate` is superseded. The tool would drift from the
  fleet.

## Owner questions

Each has a recommendation, written into the outcome above.

**Answered 2026-10-03:** "D1-D7 follow recommendations. proceed". The
outcome above stands as written, and this record is `accepted`.

* **D1.** Adopt `buildinfo`'s `version` form (recommended), or keep the old
  `1.2.3 (release)` form?
* **D2.** Keep the `CLAUDE_API_KEY` fallback (recommended), or read
  `ANTHROPIC_API_KEY` only?
* **D3.** Save the Kilo organization now (recommended), or record it as a
  limitation?
* **D4.** Pass `catalog.OptionsFromEnv()` (recommended), or drop the
  metadata overrides?
* **D5.** Fold in P8 here (recommended, as Phase 10 says), or leave it for
  a later record?
* **D6.** Remove the dead `GONOSUMCHECK` arm (recommended), or check
  `os.environ` for it?
* **D7.** Supersede 0007's dependency half (recommended), or withdraw 0007
  entirely?

## More Information

* go-llmprovider-sdk
  `docs/decisions/0002-PLAN-migrate-llmprovider-from-mcplib.md`, Phase 10,
  and its amendment of 2026-10-03.
* go-llmprovider-sdk `docs/guides/migrating-from-mcplib.md`.
* go-selfupdate-lib `docs/guides/migrating-from-mcplib-selfupdate.md` §5, and
  `docs/decisions/0004-PLAN-v1-4-0-command-surface.md` Step 7, the first
  proof of this migration, at the same commit `cfada6e`.
* [0006-MADR-mcplib-1-6-canary.md](0006-MADR-mcplib-1-6-canary.md): how this
  repository came to mcplib `v1.6.0`.
