---
status: in-progress
date: 2026-09-28
associated-madr: "0006-MADR-mcplib-1-6-canary.md"
decision-makers: Project Owner
---
# Implement the mcplib v1.6.0-rc1 Canary

Associated MADR: [0006-MADR-mcplib-1-6-canary.md](0006-MADR-mcplib-1-6-canary.md)
(`status: accepted`). The owner approved the MADR and this plan together on
2026-09-28. The release candidate is `v1.6.0-rc1`; §9 records the name
change.

This plan executes the MADR, and nothing else. If a fact contradicts the MADR
or this plan, **stop and prompt**. Add a dated entry to §9, amend the MADR when
a decision or an asserted fact changes, and only then continue.

## How this plan was proven

On 2026-09-28, in scratch copies only (`git archive` of `HEAD` `4b5dab4`):
* **Base.** `go.mod` was pointed at an archive of mcplib `4e1f9a5`, the commit
  `v1.6.0-rc1` will tag, and `go mod tidy` was run. Step S3's `go get` makes
  the same change with the real tag.
* **Red.** The tests diff was applied, and each red test was seen to fail on
  the unfixed code.
* **Mutants.** Each was applied alone to the fixed tree. Each mutant's test was
  first seen to pass unmutated, and a copy of the harness with a wrong
  assertion was seen to report the mutant as invalid.
* **Gate.** `make verify` passed on the fixed tree, and on the open.go fix
  alone against mcplib `v1.5.0`.
* **Diffs.** Appendix B's diffs were generated mechanically. Applied with
  `git apply` to a fresh `HEAD` archive, plus the dependency bump, they
  reproduce the proven tree byte for byte: 50 files, 0 mismatches.

## Goal

* `main` is green again (the open.go errcheck failure is fixed).
* `prepare-commit-msg` consumes mcplib `v1.6.0-rc1`: in `go.mod`, and in the
  release workflow's pin.
* The Codex and Grok CLI logins work, read through.
* Configure's tests are offline.
* `v1.4.0` ships, and the canary checks of MADR § Confirmation pass.
* Then, on the owner's decision, mcplib `v1.6.0` is promoted.

## Scope

**In scope:** MADR §1–§4, as phases S1–S8 below.

**Out of scope:**
* **Migrating sessions copied under older releases.** Configs with
  `auth_kind: "oauth"` are left alone (MADR, Consequences). The `v1.4.0`
  release notes tell users to re-run `configure` if a copied CLI session stops
  refreshing.
* **Non-interactive configuration of a CLI login.** `configure --yes` stays
  API-key-only, as the README says.
* **mcplib changes.** If the canary finds an mcplib defect, it is fixed under
  an mcplib record, and a new candidate (`v1.6.0-rc.2`) is tagged. That is a
  §9 deviation here.
* **The other consumers** (magictools, magicdev) follow after promotion,
  under their own records.

## 0. Preconditions and conventions

### 0.1 Baseline

* **prepare-commit-msg:** `main` at `4b5dab4`, level with `origin/main`, with a
  clean tree.
* **mcplib:** `main` contains `4e1f9a5`, and no release-candidate tag exists
  yet. The owner has since pushed `v1.6.0-rc1` (S2; §9, 2026-09-28).
* **Hooks:** `core.hooksPath` here is `.git/prepare-commit-msg-hooks`. Its
  `prepare-commit-msg` hook runs `~/.global-git-hooks/prepare-commit-msg`, and
  its `pre-commit` hook chains to the global one. Commits therefore go through
  the global message hook.

### 0.2 Gate (every code phase)

`make verify`, which runs:
* module tidiness and checksums;
* `fmt-check`;
* the pinned golangci-lint and `go vet`;
* `go test -race -v ./...`;
* the coverage threshold;
* govulncheck;
* the workflow and script lint;
* all six cross-builds.

### 0.3 Red first

Apply the **Tests** diff, and see each Appendix A red test fail with the
recorded message. Only then apply the **Fix** diff.

### 0.4 Commits, pushes, tags

* Each phase that changes files ends with one `git commit --no-edit`, after its
  gate.
* Every push and every tag (S1, S2, S4, S5, S7) needs the owner's explicit ask
  in that turn.

### 0.5 Credentials for the canary checks

* A Gemini API key (`GEMINI_API_KEY`).
* The owner's Codex CLI login, which is only ever read, never refreshed or
  revoked.
* None is printed.

## Phase S0 — Start

1. Confirm §0.1.
2. Set the MADR to `status: accepted`, and this plan to `status: in-progress`.
3. Commit the documents only.

## Phase S1 — `main` green first (MADR §4)

**Today.** `internal/ui/open.go:35` is `go func() { _ = command.Wait() }()`,
which errcheck rejects. CI run `36256206752` on `4b5dab4` failed on it, and
`make lint` fails on `HEAD`.

**Change.** Appendix B.S1: the launcher is still reaped in the background, and
a failure is reported on stderr instead of discarded.

**Verification.**
* **Red:** `make lint` on `HEAD` reports the errcheck issue (Appendix A).
* **Green:** `make verify` passes with mcplib `v1.5.0` unchanged.
* **Commit.** Pushing it, with the owner's ask, turns CI green before the
  canary work.

## Phase S2 — Tag mcplib `v1.6.0-rc1` (in mcplib)

> **Executed by the owner, 2026-09-28** (§9, §10). As approved, this phase
> named the tag ~~`v1.6.0-rc.1`~~. The owner's tag is `v1.6.0-rc1`, and every
> reference in this plan and the MADR now uses that name. Steps 3 and 4 were
> checked against it.

1. In mcplib, confirm `main` contains `4e1f9a5` and that CI run
   `36449357867` on `4e1f9a5` passed.
2. With the owner's ask:
   ```
   git tag -a v1.6.0-rc1 4e1f9a53e265808bbfa740e3e3b09a51ed7f56ce -m "mcplib v1.6.0-rc1"
   git push origin v1.6.0-rc1
   ```
   A tag message is not a commit message, so the hook rule does not apply to
   it.
3. Confirm mcplib's CI (it runs on `v*` tags) passes on the tag.
4. Confirm
   `go list -m github.com/maccavelli/mcplib@v1.6.0-rc1` resolves through the
   module proxy.

## Phase S3 — Adopt the candidate (MADR §1–§3)

1. Run
   `go get github.com/maccavelli/mcplib@v1.6.0-rc1 && go mod tidy`. Only
   `go.mod` and `go.sum` change.
2. Confirm the three tests of MADR "What was measured" fail, as recorded in
   Appendix A.
3. ~~Apply Appendix B.S3 **Tests**, then B.S3 **Fix**.~~ The step-A tests
   need step A's fix to compile (§9, 2026-09-28), so apply them with
   `git apply`, in this order:
   1. B.S3a **Tests**, then B.S3a **Fix**: offline listing, the search-first
      menu, and the CI pin;
   2. B.S3b **Tests**. The five red tests must fail as recorded, and the base
      guard must pass.
4. Apply B.S3b **Fix** with `git apply`. With B.S3a **Fix**, it covers:
   * `internal/config/config.go`: `AuthKindVendorCLI`, `VendorAuthPath`,
     `IsVendorCLI`, `ValidateVendorCLI`, and `ApplyDefaults` keeping the kind;
   * `main.go`: validation, and the `VendorCLISession` provider;
   * `internal/ui/setup.go`: the `CredVendorCLI` result, stale-session
     cleanup, and `listingClient`;
   * the README paragraph on CLI logins;
   * the release workflow pin.
5. Run §0.2, and the six mutants of Appendix A on a scratch copy.
6. Commit.

## Phase S4 — Push (owner's ask)

1. Push `main`.
2. CI must pass on every job: `Go (test; build on tag)`, and Native Tests on
   Linux, Windows and macOS.

## Phase S5 — Release `v1.4.0` (owner's ask)

1. Follow `docs/cicd-operations.md` "Controlled release":
   `git tag -a v1.4.0 -m "prepare-commit-msg v1.4.0"`, then
   `git push origin v1.4.0`.
2. The release job runs mcplib's workflow at `4e1f9a5`. Confirm the release
   is published, is immutable, and passes `gh release verify v1.4.0`.
3. The release notes add, after the generated notes:
   * **Built on mcplib `v1.6.0-rc1`.**
   * **CLI logins read through.** Choosing "Use the Codex/Grok CLI login"
     now reads the CLI's file in place.
   * **Sessions copied by older releases.** A CLI session copied into
     `prepare-commit-msg` by an older release keeps working until it next
     fails to refresh. Then run `prepare-commit-msg configure` and choose the
     CLI login.

## Phase S6 — Canary checks (MADR § Confirmation)

Each check is recorded in §10 with its output:
1. **macOS self-update.** Install `v1.3.0`, run `prepare-commit-msg update`,
   and confirm `prepare-commit-msg version` reports `v1.4.0`.
2. **Windows self-update.** Repeat on the owner's Windows laptop. That machine
   fails mcplib's `TestNativeReplaceRunningCopy`, so a failure here is a
   finding for mcplib, not a workaround target (§9).
3. **Gemini.** With a Gemini API key configured, a real `git commit` in a
   scratch repository gets a generated message.
4. **Codex CLI login.** Configure the Codex CLI login (it reads through), and
   a real `git commit` in a scratch repository gets a generated message. The
   Codex `auth.json` is unchanged afterwards (its checksum is compared).

## Phase S7 — Promote (owner's decision and ask)

Only after S6 passes:
1. In mcplib, run
   `git tag -a v1.6.0 4e1f9a53e265808bbfa740e3e3b09a51ed7f56ce -m "mcplib v1.6.0"`,
   then push the tag.
2. Here, run `go get github.com/maccavelli/mcplib@v1.6.0 && go mod tidy`.
   Change the pin's comment to `# mcplib v1.6.0`; the SHA stays the same.
3. Run §0.2, commit, and push.

## Phase S8 — Records and close-out

1. Record each phase in §10.
2. Set this plan to `status: complete` once §7 holds.
3. Add 0006 to `docs/README.md`.
4. Commit the documents.

## 7. Acceptance criteria

* Every Appendix A red test fails before its fix and passes after it.
* Every mutant is killed.
* The gate passes after S1, S3 and S7.
* CI passes after S4 and S7.
* `v1.4.0` is published and verified.
* The four canary checks of S6 pass.
* mcplib `v1.6.0` is promoted, and this repository requires it.

## 8. Rollout and rollback

* **Behaviour changes in `v1.4.0`:**
  * CLI logins read through;
  * configure's model step searches first;
  * Gemini is called on the Interactions API;
  * the mcplib `0012` provider conformance applies.
* **Rollback:**
  * Before S5: revert the S3 commit. S1 stands on its own.
  * After S5: ship `v1.4.1` with the revert. Never replace a published
    release, per `docs/cicd-operations.md` "Failed-release recovery".
* **If the canary fails:** do not promote. Fix mcplib under its own record, tag
  `v1.6.0-rc.2`, and repeat S3–S6 here (a §9 deviation).

## 9. Deviation log

* **2026-09-28: the release candidate is `v1.6.0-rc1`, not `v1.6.0-rc.1`.**
  * **Found:**
    * the owner tagged and pushed `v1.6.0-rc1`, annotated, on `4e1f9a5`;
    * mcplib CI run `36461010700` passed on it;
    * the Go module proxy serves it (`go list -m -json` reports time
      `2026-09-28T16:11:38Z`);
    * `v1.6.0-rc.1` does not exist (`unknown revision`).
  * **Decision (owner):** keep `v1.6.0-rc1`. The proxy has cached it, so
    replacing it would do more harm than the difference in name.
  * **Changes to this plan and the MADR:**
    * every reference now reads `v1.6.0-rc1`, including the CI pin comment in
      Appendix B.S3;
    * that diff was regenerated;
    * `make verify` and the diff reproduction were re-run on it, and passed;
    * the red and mutant runs were not re-run, because only a YAML comment
      changed, and no test reads it;
    * S2 is annotated as executed by the owner.
  * **Note:** semver compares `rc10` before `rc2`. A further candidate is
    `v1.6.0-rc2`.
* **2026-09-28: the first push of the documents failed on the pre-push gate.**
  * **Found:** the owner's commit `800bf0e` (this MADR and plan, documents
    only) could not be pushed. This repository's pre-push hook runs
    `make verify`, which fails on `HEAD` with the `open.go:35` errcheck issue
    S1 fixes. The global disclosure guard passed.
  * **Decision (owner):** no bypass. Land the amendment, S0 and S1, then push
    them together with `800bf0e`.
  * **Scope:** unchanged.

* **2026-09-28: S3's diffs split into the proven order.**
  * **Found:** before S3 ran, B.S3's single **Tests** diff was found to
    include step A's tests, which use `listingClient`. Only B.S3's **Fix**
    adds it. Applied alone, the tests stop `internal/ui` compiling, so its
    two red tests would fail on a build error, not with the recorded
    messages. The proof had run step B's red check on a tree that already
    had step A's fix.
  * **Decision (owner, option 1):** split B.S3 into B.S3a (step A's tests and
    fix) and B.S3b (step B's tests and fix). S3 step 3 applies them in the
    proven order.
  * **Re-proven on the harness, now pinned to `4b5dab4`:**
    * the five diffs reproduce the proven tree (50 files, 0 mismatches);
    * the red run gives the recorded failures;
    * `make verify` passes;
    * B.S1 is unchanged, byte for byte.
  * **Code:** unchanged.

## 10. Execution record

* **S0**, 2026-09-28:
  * The owner committed the documents as `800bf0e`.
  * Commit `dc79464` holds this plan's §9 amendment (the tag name), the MADR
    set to `accepted`, and this plan set to `in-progress`.
* **S2**, 2026-09-28, executed by the owner (§9):
  * the annotated tag `v1.6.0-rc1` is on `4e1f9a5`;
  * mcplib CI run `36461010700` on the tag passed;
  * `go list -m github.com/maccavelli/mcplib@v1.6.0-rc1` resolves through the
    module proxy (time `2026-09-28T16:11:38Z`).
* **S1**, commit `86b21ef`:
  * Appendix B.S1 was applied with `git apply`, taken from this document and
    checked equal to the proven diff.
  * The red check (`make lint` on `HEAD`) had failed as recorded in
    Appendix A.
  * `make verify` passed in the repository, with mcplib `v1.5.0`.
  * **Pushed** with `800bf0e`, `dc79464` and `3106f7d` (`4b5dab4..3106f7d`).
    The pre-push `make verify` passed.
  * **CI** run `36463370832` on `3106f7d` passed: `Go (test; build on tag)`,
    and Native Tests on Linux, macOS and Windows. `main` is green again.
* **S3**, commit `1ed8a74`, run after the §9 split (`8241b8a`). Each diff was
  taken from this document and checked equal to the proven diff.
  1. **Bump.** `go get …@v1.6.0-rc1 && go mod tidy` changed only `go.mod` and
     `go.sum`. `go list -m` reports `github.com/maccavelli/mcplib v1.6.0-rc1`.
  2. **The bump alone** fails exactly the three tests of MADR "What was
     measured": `…_Success`, `…_ImportGrokSession` and
     `…_ChatGPTDoesNotCopyAccessIntoAPIKey`.
  3. **B.S3a Tests and Fix, then B.S3b Tests.** All five red tests failed with
     the messages recorded in Appendix A, and
     `TestRunAnalyzer_OAuthDoesNotCallNewProviderWithAccessToken` passed.
  4. **B.S3b Fix.** `make verify` passed in the repository.
  5. **Mutants** ran on an archive of `HEAD` plus the working diff. Each test
     passed unmutated first, and all six were killed, with the messages of
     Appendix A.
  6. **Committed** 10 files: `go.mod`, `go.sum`, `ci.yml`, `README.md`,
     `main.go`, `main_oauth_test.go`, and in `internal/config/` and
     `internal/ui/`, `config.go`, `config_test.go`, `setup.go` and
     `setup_test.go`.
  * **Pending:** S4, the push and CI.

## Appendix A — Proof record (2026-09-28)

### A.S1 `main` green first

**Red** (`make lint` on `HEAD` `4b5dab4`, the pinned golangci-lint):

```text
internal/ui/open.go:35:14: Error return value of `command.Wait` is not checked (errcheck)
1 issues:
* errcheck: 1
```

**Green:** `make verify` passes with Appendix B.S1 applied and mcplib `v1.5.0` unchanged.

### A.S3 Adopt the candidate

**The bump alone** (`HEAD` against mcplib `4e1f9a5`): three tests fail, as MADR "What was measured" records.

**Red** (tests diff applied, fix not): each test failed as required.

| Test | Package | First failure message |
|---|---|---|
| `TestRunSetupInteractive_GrokCLILoginReadsThrough` | `internal/ui` | runSetupInteractive() error = API key is required |
| `TestRunSetupInteractive_CodexCLILoginReadsThrough` | `internal/ui` | runSetupInteractive() error = API key is required |
| `TestApplyDefaults_KeepsVendorCLILogin` | `internal/config` | grok config = {"api_key":"","model":"grok-4.6","fallback_models":["grok-4.5","grok-3-mini-fast","grok-3-mini"]}, want "auth_kind":"vendor_cli" |
| `TestRunAnalyzer_VendorCLIReadsThrough` | `main` | runAnalyzer() error = no API key for provider "grok"; run 'prepare-commit-msg configure' or set XAI_API_KEY |
| `TestRunAnalyzer_VendorCLIWithoutPathFails` | `main` | runAnalyzer() error = no API key for provider "grok"; run 'prepare-commit-msg configure' or set XAI_API_KEY, want the missing CLI login path |

**Guard before the fix:**

* `TestRunAnalyzer_OAuthDoesNotCallNewProviderWithAccessToken`: passed on the unfixed code.

**Step A has no red run.** Its tests name new API (`listingClient`), so the mutants below prove them. The rewritten `TestRunSetupInteractive_Success` passes only with the search step scripted.

**Mutants** (each applied alone to the fixed tree; each test first seen to pass unmutated):

| Mutant | Test | Result | Failure message |
|---|---|---|---|
| `a-wizard-default-client` | `TestRunSetupInteractive_ListingUsesInjectedClient` | killed | configure listed models without the injected client |
| `a-discover-default-client` | `TestDiscoverModels_UsesInjectedClient` | killed | discoverModels listed models without the injected client |
| `b-keep-stale-session` | `TestRunSetupInteractive_GrokCLILoginReadsThrough` | killed | stored Grok session = &{Provider:grok Access:copied-by-an-older-release Refresh: Expiry:0001-01-01 00:00:00 +0000 UTC Issuer: ClientID: AccountID: FedRAMP:fals… |
| `b-runtime-drops-path` | `TestRunAnalyzer_VendorCLIReadsThrough` | killed | NewProviderWithSource() source = &llmprovider.VendorCLISession{Provider:"grok", Path:"", mu:sync.Mutex{_:sync.noCopy{}, mu:sync.Mutex{state:0, sema:0x0}}, acco… |
| `b-validate-skips-path` | `TestRunAnalyzer_VendorCLIWithoutPathFails` | killed | a vendor CLI login without a path built a provider |
| `b-setup-drops-path` | `TestRunSetupInteractive_CodexCLILoginReadsThrough` | killed | saved openai vendor_auth_path = "set", want "<tmp>" |

The validity check was itself seen to fail. In a copy of the harness with the wrong assertion that first hid a failing test, it reported `a-discover-default-client -> TestDiscoverModels_UsesInjectedClient: INVALID (fails unmutated)`, and the run exited 1.

**Gate** (`make verify`, fixed tree, including the CI pin): passed. That covers module tidiness and checksums, `fmt-check`, golangci-lint, `go vet`, `go test -race`, the coverage threshold, govulncheck, actionlint and the script checks, and the six cross-builds.

**Diffs** (regenerated 2026-09-28 in S3's split order, from `4b5dab4`; §9):

```text
c1-fix.diff: 26 lines
c2a-tests.diff: 106 lines
c2a-fix.diff: 55 lines
c2b-tests.diff: 298 lines
c2b-fix.diff: 151 lines
git apply c1-fix: exit=0 
git apply c2a-tests: exit=0 
git apply c2a-fix: exit=0 
git apply c2b-tests: exit=0 
git apply c2b-fix: exit=0 
compared 50 files with the proven green tree: mismatches=[] extra=[]
```

## Appendix B — Diffs

Generated from the proof. Apply them with `git apply` in this order:
1. B.S1, on the §0.1 baseline;
2. B.S3a Tests, after S3 step 1;
3. B.S3a Fix;
4. B.S3b Tests;
5. B.S3b Fix.

### B.S1 `main` green first

**Fix** (`c1-fix.diff`, 26 lines):

```diff
diff --git a/internal/ui/open.go b/internal/ui/open.go
--- a/internal/ui/open.go
+++ b/internal/ui/open.go
@@ -3,6 +3,7 @@
 import (
 	"fmt"
 	"net/url"
+	"os"
 	"os/exec"
 	"runtime"
 	"strings"
@@ -32,7 +33,13 @@
 	if err := command.Start(); err != nil {
 		return fmt.Errorf("open browser: %w", err)
 	}
-	go func() { _ = command.Wait() }()
+	// Reap the launcher without blocking the sign-in flow. The URL is already
+	// printed, so a launcher that fails only needs saying so.
+	go func() {
+		if err := command.Wait(); err != nil {
+			fmt.Fprintf(os.Stderr, "could not open a browser (%v); open the URL above instead\n", err)
+		}
+	}()
 	return nil
 }
 
```

### B.S3a Adopt the candidate, step A: offline listing and the pin

**Tests** (`c2a-tests.diff`, 106 lines):

```diff
diff --git a/internal/ui/setup_test.go b/internal/ui/setup_test.go
--- a/internal/ui/setup_test.go
+++ b/internal/ui/setup_test.go
@@ -4,10 +4,13 @@
 	"bufio"
 	"bytes"
 	"context"
+	"errors"
+	"net/http"
 	"os"
 	"path/filepath"
 	"runtime"
 	"strings"
+	"sync/atomic"
 	"testing"
 
 	"github.com/maccavelli/mcplib/llmprovider"
@@ -15,13 +18,27 @@
 	"github.com/maccavelli/prepare-commit-msg/internal/config"
 )
 
-func isolate(t *testing.T) {
+// offlineTransport fails every request, so configure's unit tests never
+// reach a live listing endpoint; requests counts the attempts.
+type offlineTransport struct{ requests atomic.Int32 }
+
+func (o *offlineTransport) RoundTrip(*http.Request) (*http.Response, error) {
+	o.requests.Add(1)
+	return nil, errors.New("offline: unit tests make no network calls")
+}
+
+func isolate(t *testing.T) *offlineTransport {
 	t.Helper()
+	offline := &offlineTransport{}
+	previous := listingClient
+	listingClient = &http.Client{Transport: offline}
+	t.Cleanup(func() { listingClient = previous })
 	tmp := t.TempDir()
 	t.Setenv("HOME", tmp)
 	t.Setenv("USERPROFILE", tmp)
 	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
 	t.Setenv("AppData", filepath.Join(tmp, "AppData", "Roaming"))
+	return offline
 }
 
 func TestRunSetupInteractive_Success(t *testing.T) {
@@ -41,11 +58,12 @@
 
 	// 1: gemini
 	// y: use env key
-	// Static gemini catalog has 6 models → 7 is Other
-	// Enter custom model
-	// fallbacks: enter = recommended
+	// search: enter; the listing is offline, so the menu is the built-in
+	// catalog: 6 models, then 7 Other
+	// 7, then the custom model id
+	// fallbacks: enter to search, enter for none
 	// operational: all enter (defaults)
-	input := "1\ny\n7\nmy-custom-model\n\n\n\n\n\n"
+	input := "1\ny\n\n7\nmy-custom-model\n\n\n\n\n\n\n"
 	r := strings.NewReader(input)
 
 	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, r); err != nil {
@@ -62,6 +80,42 @@
 	}
 	if conf.TimeoutSeconds != config.DefaultTimeoutSeconds {
 		t.Errorf("timeout defaults: %d", conf.TimeoutSeconds)
+	}
+}
+
+// TestRunSetupInteractive_ListingUsesInjectedClient: configure's live model
+// listing goes through listingClient, so tests can keep it offline.
+func TestRunSetupInteractive_ListingUsesInjectedClient(t *testing.T) {
+	offline := isolate(t)
+
+	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
+	config.ApplyDefaults(conf)
+
+	oldEnv := osGetenv
+	defer func() { osGetenv = oldEnv }()
+	osGetenv = func(k string) string {
+		if k == "GEMINI_API_KEY" {
+			return "test-key"
+		}
+		return ""
+	}
+
+	input := "1\ny\n\n1\n\n\n\n\n\n\n"
+	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
+		t.Fatalf("unexpected error: %v", err)
+	}
+	if offline.requests.Load() == 0 {
+		t.Fatal("configure listed models without the injected client")
+	}
+}
+
+// TestDiscoverModels_UsesInjectedClient: the non-interactive listing goes
+// through listingClient too.
+func TestDiscoverModels_UsesInjectedClient(t *testing.T) {
+	offline := isolate(t)
+	discoverModels(context.Background(), providerGemini, "test-key")
+	if offline.requests.Load() == 0 {
+		t.Fatal("discoverModels listed models without the injected client")
 	}
 }
 
```

**Fix** (`c2a-fix.diff`, 55 lines):

```diff
diff --git a/.github/workflows/ci.yml b/.github/workflows/ci.yml
--- a/.github/workflows/ci.yml
+++ b/.github/workflows/ci.yml
@@ -112,7 +112,7 @@
       contents: write
       id-token: write
       attestations: write
-    uses: maccavelli/mcplib/.github/workflows/publish-selfupdate-release.yml@d13f89cf6ee385bc76f8bf3d3c11155276c2af31 # mcplib v1.5.0
+    uses: maccavelli/mcplib/.github/workflows/publish-selfupdate-release.yml@4e1f9a53e265808bbfa740e3e3b09a51ed7f56ce # mcplib v1.6.0-rc1
     with:
       artifact-name: prepare-commit-msg-${{ needs.go.outputs.version || github.sha }}
       products-json: '["prepare-commit-msg"]'
diff --git a/internal/ui/setup.go b/internal/ui/setup.go
--- a/internal/ui/setup.go
+++ b/internal/ui/setup.go
@@ -5,6 +5,7 @@
 	"context"
 	"fmt"
 	"io"
+	"net/http"
 	"os"
 	"strconv"
 	"strings"
@@ -26,6 +27,10 @@
 )
 
 var osGetenv = os.Getenv
+
+// listingClient carries configure's live model listing. Nil uses mcplib's
+// default client; tests set one that never reaches the network.
+var listingClient *http.Client
 
 // SetupOptions holds non-interactive / flag-driven configure settings.
 // Zero values mean "unset" and interactive mode will prompt (unless Yes is set).
@@ -137,7 +142,11 @@
 func discoverModels(ctx context.Context, provider, apiKey string) []string {
 	dCtx, cancel := context.WithTimeout(ctx, DiscoveryTimeout)
 	defer cancel()
-	models, err := llmprovider.ListAvailableModels(dCtx, provider, apiKey)
+	var opts []llmprovider.ProviderOption
+	if listingClient != nil {
+		opts = append(opts, llmprovider.WithHTTPClient(listingClient))
+	}
+	models, err := llmprovider.ListAvailableModels(dCtx, provider, apiKey, opts...)
 	if err != nil {
 		return nil
 	}
@@ -216,6 +225,7 @@
 		LookupEnv:     osGetenv,
 		Discover:      true,
 		DiscoverLimit: DiscoveryTimeout,
+		HTTPClient:    listingClient,
 		NeedFallbacks: true,
 		TokenStore:    store,
 		OpenURL:       openBrowser,
```

### B.S3b Adopt the candidate, step B: CLI logins read through

**Tests** (`c2b-tests.diff`, 298 lines):

```diff
diff --git a/internal/config/config_test.go b/internal/config/config_test.go
--- a/internal/config/config_test.go
+++ b/internal/config/config_test.go
@@ -3,6 +3,7 @@
 import (
 	"bytes"
 	"context"
+	"encoding/json"
 	"fmt"
 	"os"
 	"path/filepath"
@@ -442,3 +443,24 @@
 		t.Errorf("expected 1 fallback model 'gpt-4o-mini', got %v", pc.FallbackModels)
 	}
 }
+
+// TestApplyDefaults_KeepsVendorCLILogin: a vendor CLI login, and the path to
+// its auth file, survive ApplyDefaults and the JSON round trip.
+func TestApplyDefaults_KeepsVendorCLILogin(t *testing.T) {
+	raw := `{"active_provider":"grok","providers":{"grok":{"auth_kind":"vendor_cli",` +
+		`"vendor_auth_path":"/x/grok/auth.json","model":"grok-4.6"}}}`
+	var c Config
+	if err := json.Unmarshal([]byte(raw), &c); err != nil {
+		t.Fatalf("decode: %v", err)
+	}
+	ApplyDefaults(&c)
+	out, err := json.Marshal(c.Providers["grok"])
+	if err != nil {
+		t.Fatalf("encode: %v", err)
+	}
+	for _, want := range []string{`"auth_kind":"vendor_cli"`, `"vendor_auth_path":"/x/grok/auth.json"`} {
+		if !bytes.Contains(out, []byte(want)) {
+			t.Errorf("grok config = %s, want %s", out, want)
+		}
+	}
+}
diff --git a/internal/ui/setup_test.go b/internal/ui/setup_test.go
--- a/internal/ui/setup_test.go
+++ b/internal/ui/setup_test.go
@@ -4,11 +4,11 @@
 	"bufio"
 	"bytes"
 	"context"
+	"encoding/json"
 	"errors"
 	"net/http"
 	"os"
 	"path/filepath"
-	"runtime"
 	"strings"
 	"sync/atomic"
 	"testing"
@@ -230,49 +230,57 @@
   }
 }`
 
-func TestRunSetupInteractive_ImportGrokSession(t *testing.T) {
+// TestRunSetupInteractive_GrokCLILoginReadsThrough: choosing the Grok CLI
+// login saves only the path to its auth file. No token is copied, and a
+// session an older release copied from the CLI is removed, so it is never
+// refreshed against the CLI's own (mcplib MADR 0012 §5.1).
+func TestRunSetupInteractive_GrokCLILoginReadsThrough(t *testing.T) {
 	isolate(t)
 	vendorHome := t.TempDir()
 	t.Setenv("GROK_HOME", vendorHome)
-	if err := os.WriteFile(filepath.Join(vendorHome, "auth.json"), []byte(grokAuthFixture), 0o600); err != nil {
+	authPath := filepath.Join(vendorHome, "auth.json")
+	if err := os.WriteFile(authPath, []byte(grokAuthFixture), 0o600); err != nil {
 		t.Fatalf("write Grok fixture: %v", err)
 	}
 	originalEnv := osGetenv
 	osGetenv = os.Getenv
 	t.Cleanup(func() { osGetenv = originalEnv })
 
-	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
-	config.ApplyDefaults(conf)
-	ctx, cancel := context.WithCancel(context.Background())
-	cancel()
-	input := "4\n5\ny\n1\n\n\n\n\n\n"
-	if err := runSetupInteractive(ctx, conf, SetupOptions{}, strings.NewReader(input)); err != nil {
+	store, err := config.NewOAuthStore()
+	if err != nil {
+		t.Fatalf("NewOAuthStore() error = %v", err)
+	}
+	stale := &llmprovider.OAuthSession{Provider: llmprovider.ProviderGrok, Access: "copied-by-an-older-release"}
+	if err := store.Save(context.Background(), llmprovider.ProviderGrok, stale); err != nil {
+		t.Fatalf("save stale session: %v", err)
+	}
+
+	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
+	config.ApplyDefaults(conf)
+	input := "4\n5\ny\n\n1\n\n\n\n\n\n\n"
+	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
 		t.Fatalf("runSetupInteractive() error = %v", err)
 	}
 	pc := conf.Providers[llmprovider.ProviderGrok]
-	if pc.AuthKind != "oauth" || pc.APIKey != "" {
-		t.Fatalf("Grok auth kind/key = %q/%q", pc.AuthKind, pc.APIKey)
-	}
-	oauthDir, err := config.OAuthDir()
-	if err != nil {
-		t.Fatalf("OAuthDir() error = %v", err)
-	}
-	tokenPath := filepath.Join(oauthDir, llmprovider.ProviderGrok+".json")
-	info, err := os.Stat(tokenPath)
-	if err != nil {
-		t.Fatalf("stat Grok OAuth session: %v", err)
-	}
-	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
-		t.Fatalf("Grok OAuth mode = %04o, want 0600", info.Mode().Perm())
+	if pc.AuthKind != "vendor_cli" || pc.APIKey != "" {
+		t.Fatalf("Grok auth kind/key = %q/%q, want vendor_cli and no key", pc.AuthKind, pc.APIKey)
+	}
+	assertSavedVendorPath(t, llmprovider.ProviderGrok, authPath)
+	if session, loadErr := store.Load(context.Background(), llmprovider.ProviderGrok); loadErr != nil || session != nil {
+		t.Fatalf("stored Grok session = %+v, %v; want none", session, loadErr)
 	}
 	assertConfigOmitsOAuthTokens(t)
 }
 
-func TestRunSetupInteractive_ChatGPTDoesNotCopyAccessIntoAPIKey(t *testing.T) {
+// TestRunSetupInteractive_CodexCLILoginReadsThrough: choosing the Codex CLI
+// login saves only the path to its auth file; neither its access token nor
+// its OPENAI_API_KEY becomes the configured API key.
+func TestRunSetupInteractive_CodexCLILoginReadsThrough(t *testing.T) {
 	isolate(t)
 	vendorHome := t.TempDir()
 	t.Setenv("CODEX_HOME", vendorHome)
-	if err := os.WriteFile(filepath.Join(vendorHome, "auth.json"), []byte(openAIAuthFixture), 0o600); err != nil {
+	authPath := filepath.Join(vendorHome, "auth.json")
+	if err := os.WriteFile(authPath, []byte(openAIAuthFixture), 0o600); err != nil {
 		t.Fatalf("write OpenAI fixture: %v", err)
 	}
 	originalEnv := osGetenv
@@ -281,15 +289,41 @@
 
 	conf := &config.Config{Providers: make(map[string]config.ProviderConfig)}
 	config.ApplyDefaults(conf)
-	input := "2\n5\ny\n1\n\n\n\n\n\n"
+	input := "2\n5\ny\ngpt-5.4\n\n\n\n\n\n\n"
 	if err := runSetupInteractive(context.Background(), conf, SetupOptions{}, strings.NewReader(input)); err != nil {
 		t.Fatalf("runSetupInteractive() error = %v", err)
 	}
 	pc := conf.Providers[llmprovider.ProviderOpenAI]
-	if pc.AuthKind != "oauth" || pc.APIKey != "" {
-		t.Fatalf("OpenAI auth kind/key = %q/%q", pc.AuthKind, pc.APIKey)
-	}
+	if pc.AuthKind != "vendor_cli" || pc.APIKey != "" {
+		t.Fatalf("OpenAI auth kind/key = %q/%q, want vendor_cli and no key", pc.AuthKind, pc.APIKey)
+	}
+	assertSavedVendorPath(t, llmprovider.ProviderOpenAI, authPath)
 	assertConfigOmitsOAuthTokens(t)
+}
+
+// assertSavedVendorPath reads the saved configuration file and checks the
+// provider's vendor_auth_path.
+func assertSavedVendorPath(t *testing.T, provider, want string) {
+	t.Helper()
+	path, err := config.GetConfigPath()
+	if err != nil {
+		t.Fatalf("GetConfigPath() error = %v", err)
+	}
+	data, err := os.ReadFile(path)
+	if err != nil {
+		t.Fatalf("read config: %v", err)
+	}
+	var saved struct {
+		Providers map[string]struct {
+			VendorAuthPath string `json:"vendor_auth_path"`
+		} `json:"providers"`
+	}
+	if err := json.Unmarshal(data, &saved); err != nil {
+		t.Fatalf("decode config: %v", err)
+	}
+	if got := saved.Providers[provider].VendorAuthPath; got != want {
+		t.Fatalf("saved %s vendor_auth_path = %q, want %q", provider, got, want)
+	}
 }
 
 func assertConfigOmitsOAuthTokens(t *testing.T) {
diff --git a/main_oauth_test.go b/main_oauth_test.go
--- a/main_oauth_test.go
+++ b/main_oauth_test.go
@@ -2,8 +2,10 @@
 
 import (
 	"context"
+	"encoding/json"
 	"os"
 	"path/filepath"
+	"strings"
 	"testing"
 	"time"
 
@@ -74,6 +76,103 @@
 	}
 }
 
+// TestRunAnalyzer_VendorCLIReadsThrough: a vendor CLI login generates through
+// a VendorCLISession on the saved auth file, never through an API key.
+func TestRunAnalyzer_VendorCLIReadsThrough(t *testing.T) {
+	t.Setenv("HOME", t.TempDir())
+	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
+	var pc config.ProviderConfig
+	if err := json.Unmarshal([]byte(`{"auth_kind":"vendor_cli","vendor_auth_path":"/x/grok/auth.json",`+
+		`"model":"grok-4.6"}`), &pc); err != nil {
+		t.Fatalf("decode: %v", err)
+	}
+
+	originalNewProvider := newProvider
+	originalNewProviderWithSource := newProviderWithSource
+	originalGenerate := generateWithRetry
+	originalEnv := osGetenv
+	t.Cleanup(func() {
+		newProvider = originalNewProvider
+		newProviderWithSource = originalNewProviderWithSource
+		generateWithRetry = originalGenerate
+		osGetenv = originalEnv
+	})
+	osGetenv = func(string) string { return "" }
+	newProvider = func(string, string, string, ...llmprovider.ProviderOption) (llmprovider.Provider, error) {
+		t.Fatal("vendor CLI generation called NewProvider")
+		return nil, nil
+	}
+	var capturedSource llmprovider.TokenSource
+	newProviderWithSource = func(
+		_ string,
+		source llmprovider.TokenSource,
+		_ string,
+		_ ...llmprovider.ProviderOption,
+	) (llmprovider.Provider, error) {
+		capturedSource = source
+		return analyzerTestProvider{}, nil
+	}
+	generateWithRetry = func(context.Context, llmprovider.Provider, string, int, time.Duration) (string, error) {
+		return "feat: use the Grok CLI login", nil
+	}
+
+	conf := &config.Config{
+		ActiveProvider: llmprovider.ProviderGrok,
+		Providers:      map[string]config.ProviderConfig{llmprovider.ProviderGrok: pc},
+		TimeoutSeconds: 5,
+	}
+	messagePath := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
+	if err := os.WriteFile(messagePath, nil, 0o600); err != nil {
+		t.Fatalf("write message file: %v", err)
+	}
+	info := &git.Info{Files: []string{"main.go"}, Additions: 1}
+	if err := runAnalyzer(messagePath, conf, info); err != nil {
+		t.Fatalf("runAnalyzer() error = %v", err)
+	}
+	session, ok := capturedSource.(*llmprovider.VendorCLISession)
+	if !ok || session.Provider != llmprovider.ProviderGrok || session.Path != "/x/grok/auth.json" {
+		t.Fatalf("NewProviderWithSource() source = %#v, want the Grok VendorCLISession on /x/grok/auth.json", capturedSource)
+	}
+}
+
+// TestRunAnalyzer_VendorCLIWithoutPathFails: a vendor CLI login with no saved
+// path sends the user back to configure, and builds no provider.
+func TestRunAnalyzer_VendorCLIWithoutPathFails(t *testing.T) {
+	t.Setenv("HOME", t.TempDir())
+	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
+	var pc config.ProviderConfig
+	if err := json.Unmarshal([]byte(`{"auth_kind":"vendor_cli","model":"grok-4.6"}`), &pc); err != nil {
+		t.Fatalf("decode: %v", err)
+	}
+	originalNewProvider := newProvider
+	originalNewProviderWithSource := newProviderWithSource
+	originalEnv := osGetenv
+	t.Cleanup(func() {
+		newProvider = originalNewProvider
+		newProviderWithSource = originalNewProviderWithSource
+		osGetenv = originalEnv
+	})
+	osGetenv = func(string) string { return "" }
+	newProvider = func(string, string, string, ...llmprovider.ProviderOption) (llmprovider.Provider, error) {
+		t.Fatal("a vendor CLI login without a path built an API-key provider")
+		return nil, nil
+	}
+	newProviderWithSource = func(string, llmprovider.TokenSource, string, ...llmprovider.ProviderOption) (llmprovider.Provider, error) {
+		t.Fatal("a vendor CLI login without a path built a provider")
+		return nil, nil
+	}
+	conf := &config.Config{
+		ActiveProvider: llmprovider.ProviderGrok,
+		Providers:      map[string]config.ProviderConfig{llmprovider.ProviderGrok: pc},
+		TimeoutSeconds: 5,
+	}
+	info := &git.Info{Files: []string{"main.go"}, Additions: 1}
+	err := runAnalyzer(filepath.Join(t.TempDir(), "COMMIT_EDITMSG"), conf, info)
+	if err == nil || !strings.Contains(err.Error(), "no CLI login path") {
+		t.Fatalf("runAnalyzer() error = %v, want the missing CLI login path", err)
+	}
+}
+
 type analyzerTestProvider struct{}
 
 func (analyzerTestProvider) Name() string { return llmprovider.ProviderOpenAI }
```

**Fix** (`c2b-fix.diff`, 151 lines):

```diff
diff --git a/README.md b/README.md
--- a/README.md
+++ b/README.md
@@ -197,6 +197,12 @@
 main configuration in its private `oauth/` directory; access and refresh tokens
 are never copied into `config.json`. Non-interactive `configure --yes` remains
 API-key-only.
+
+Choosing **Use the Codex CLI login** or **Use the Grok CLI login** stores only
+the path to that CLI's `auth.json` (`vendor_auth_path` in `config.json`). The
+file is read on every run and never refreshed here: the CLI keeps its own
+login current. When it expires, sign in again with the CLI; there is no need
+to re-run `configure`.
 
 ---
 
diff --git a/internal/config/config.go b/internal/config/config.go
--- a/internal/config/config.go
+++ b/internal/config/config.go
@@ -44,10 +44,19 @@
 	RetryDelaySeconds int `json:"retry_delay_seconds"`
 }
 
+// AuthKindVendorCLI selects a vendor CLI's own login (Codex or Grok), read in
+// place through llmprovider.VendorCLISession (mcplib MADR 0012 §5.1).
+const AuthKindVendorCLI = "vendor_cli"
+
 // ProviderConfig stores credentials and model selection for a single LLM provider.
 type ProviderConfig struct {
-	// AuthKind is empty for legacy/API-key auth and "oauth" for a saved session.
-	AuthKind       string   `json:"auth_kind,omitempty"`
+	// AuthKind is empty for legacy/API-key auth, "oauth" for a saved session,
+	// and AuthKindVendorCLI for a vendor CLI's own login.
+	AuthKind string `json:"auth_kind,omitempty"`
+	// VendorAuthPath is the vendor CLI's auth file when AuthKind is
+	// AuthKindVendorCLI. It holds no token: the file is read on every request,
+	// and the CLI keeps refreshing its own login.
+	VendorAuthPath string   `json:"vendor_auth_path,omitempty"`
 	APIKey         string   `json:"api_key"`
 	Model          string   `json:"model"`
 	FallbackModels []string `json:"fallback_models,omitempty"`
@@ -127,9 +136,12 @@
 		c.ActiveProvider = llmprovider.ProviderGemini
 	}
 	for provider, pc := range c.Providers {
-		if IsOAuth(pc) {
+		switch {
+		case IsOAuth(pc):
 			pc.AuthKind = "oauth"
-		} else {
+		case IsVendorCLI(pc):
+			pc.AuthKind = AuthKindVendorCLI
+		default:
 			pc.AuthKind = ""
 		}
 		c.Providers[provider] = pc
@@ -282,6 +294,26 @@
 // IsOAuth reports whether a provider config selects subscription authentication.
 func IsOAuth(pc ProviderConfig) bool {
 	return strings.EqualFold(pc.AuthKind, "oauth")
+}
+
+// IsVendorCLI reports whether a provider config reads a vendor CLI's login.
+func IsVendorCLI(pc ProviderConfig) bool {
+	return strings.EqualFold(pc.AuthKind, AuthKindVendorCLI)
+}
+
+// ValidateVendorCLI checks that a vendor CLI login has a model and an auth
+// file path. The file, and the token in it, are read on use.
+func ValidateVendorCLI(provider string, pc ProviderConfig) error {
+	if strings.TrimSpace(provider) == "" {
+		return fmt.Errorf("no active provider configured; please run 'prepare-commit-msg configure'")
+	}
+	if strings.TrimSpace(pc.Model) == "" {
+		return fmt.Errorf("no model configured for provider %q; run 'prepare-commit-msg configure'", provider)
+	}
+	if strings.TrimSpace(pc.VendorAuthPath) == "" {
+		return fmt.Errorf("no CLI login path for provider %q; run 'prepare-commit-msg configure'", provider)
+	}
+	return nil
 }
 
 // ValidateOAuth checks that an OAuth provider has both a model and a saved session.
diff --git a/internal/ui/setup.go b/internal/ui/setup.go
--- a/internal/ui/setup.go
+++ b/internal/ui/setup.go
@@ -241,13 +241,27 @@
 	}
 	pc.Model = res.Model
 	pc.FallbackModels = res.Fallbacks
-	if res.Kind == wizard.CredOAuth {
+	pc.VendorAuthPath = ""
+	switch res.Kind {
+	case wizard.CredOAuth:
 		pc.AuthKind = string(wizard.CredOAuth)
 		pc.APIKey = ""
 		if err := config.ValidateOAuth(ctx, res.Provider, pc, store); err != nil {
 			return err
 		}
-	} else {
+	case wizard.CredVendorCLI:
+		pc.AuthKind = config.AuthKindVendorCLI
+		pc.APIKey = ""
+		pc.VendorAuthPath = res.VendorAuthPath
+		if err := config.ValidateVendorCLI(res.Provider, pc); err != nil {
+			return err
+		}
+		// A session an older release copied from the CLI shares the CLI's
+		// refresh token; drop it so it is never refreshed again.
+		if err := store.Delete(ctx, res.Provider); err != nil {
+			return fmt.Errorf("delete stale OAuth session: %w", err)
+		}
+	default:
 		pc.AuthKind = ""
 		pc.APIKey = res.APIKey
 		d, _ := llmprovider.DescriptorFor(res.Provider)
@@ -313,6 +327,7 @@
 	}
 	pc.APIKey = apiKey
 	pc.AuthKind = ""
+	pc.VendorAuthPath = ""
 
 	model := strings.TrimSpace(opts.Model)
 	if model == "" {
diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -269,6 +269,10 @@
 		if err := config.ValidateOAuth(ctx, conf.ActiveProvider, pc, store); err != nil {
 			return err
 		}
+	} else if config.IsVendorCLI(pc) {
+		if err := config.ValidateVendorCLI(conf.ActiveProvider, pc); err != nil {
+			return err
+		}
 	} else {
 		apiKey := config.ResolveAPIKey(pc, conf.ActiveProvider, true, osGetenv)
 		if err := config.ValidateActive(conf.ActiveProvider, pc, apiKey); err != nil {
@@ -333,6 +337,12 @@
 		session.Store = store
 		return newProviderWithSource(conf.ActiveProvider, session, model)
 	}
+	if config.IsVendorCLI(pc) {
+		// The CLI's auth file is read on every request and never refreshed here,
+		// so the CLI keeps its refresh token (mcplib MADR 0012 §5.1).
+		source := &llmprovider.VendorCLISession{Provider: conf.ActiveProvider, Path: pc.VendorAuthPath}
+		return newProviderWithSource(conf.ActiveProvider, source, model)
+	}
 	apiKey := config.ResolveAPIKey(pc, conf.ActiveProvider, true, osGetenv)
 	if err := config.ValidateActive(conf.ActiveProvider, pc, apiKey); err != nil {
 		return nil, err
```
