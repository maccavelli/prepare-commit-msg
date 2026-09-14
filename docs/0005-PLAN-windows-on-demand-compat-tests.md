---
status: proposed
date: 2026-09-14
---
<!-- markdownlint-disable MD013 MD024 MD033 MD036 MD060 -->

# PLAN 0005 — Windows on-demand compatibility tests

Implements [0005-MADR-windows-on-demand-compat-tests.md](0005-MADR-windows-on-demand-compat-tests.md)
decisions D1–D8, closing findings F1–F8.

## Goal

After the last phase:

* Native untagged `go test ./... -count=1` still passes the five existing
  packages and does not compile any `wincompat` test.
* Native `go test -tags wincompat -count=1 ./...` passes on this
  `windows/amd64` host with **zero skips** for the required cases listed in
  the MADR Confirmation.
* `./scripts/test-windows.sh` from WSL re-execs native `go.exe` and passes
  that tagged suite.
* `powershell.exe -NoProfile -File scripts\test-windows.ps1` does the same.
* `make test-windows` is a thin wrapper around the shell driver.
* `make verify` / CI `go-native` commands are byte-for-byte the same as
  before this plan, except `scripts/verify-scripts.sh` only `bash -n`s
  `*.sh`.
* README Developer Experience and `docs/cicd-operations.md` document the
  on-demand command.

## Scope

### In scope (the only files any phase may touch)

* `docs/0005-MADR-windows-on-demand-compat-tests.md`
* `docs/0005-PLAN-windows-on-demand-compat-tests.md`
* `docs/README.md`
* `docs/cicd-operations.md`
* `README.md`
* `Makefile`
* `scripts/verify-scripts.sh`
* `scripts/test-windows.sh`
* `scripts/test-windows.ps1`
* `wincompat_guard_test.go`
* `main_wincompat_test.go`
* `update_wincompat_test.go`
* `internal/git/wincompat_test.go`
* `internal/fsutil/wincompat_test.go`
* `internal/config/wincompat_test.go`

### Out of scope

* Production Go behavior (`cmd_windows.go`, `atomic.go`, hook runtime,
  config path algorithm) unless a required test proves a bug — then stop,
  amend this pair, and re-approve. Do not silently patch product code.
* `.github/workflows/ci.yml` and any always-on CI job.
* `make test`, `make coverage`, `make verify` membership (the suite stays
  out of those targets).
* Coverage threshold, golangci config, Go module version bumps.
* `make install` Windows `.exe` destination (deferred product fix).
* Porting `scripts/*.sh` hook helpers to PowerShell.
* Native `windows/arm64` execution (this host is amd64).

## Stability rule

Every phase ends with **all** of:

```text
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test ./... -count=1"
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test -tags wincompat -count=1 ./..."
git diff --check
```

From P1 onward the tagged command must exist and must fail or pass for the
reasons the phase names — never because the default suite broke.

Do not `git push`. Do not create tags. Commit locally once per phase, with
only that phase's in-scope files. Do not stage the currently dirty
unrelated worktree files.

If a required test cannot pass without a production change, **stop** and
amend this pair. Do not weaken the test.

## Cross-cutting contracts

1. **C1** — No production Go behavior change in this plan. Tests, drivers,
   Makefile wiring, and docs only (plus the `verify-scripts.sh` `*.sh`
   filter, which is a script-linter glob, not product behavior).
2. **C2** — Untagged `go test ./...` compiles none of the new tests
   (`//go:build windows && wincompat` or `//go:build wincompat` as specified
   per file).
3. **C3** — CI workflow YAML is not modified.
4. **C4** — `make test` / `make coverage` / `make verify` do not invoke
   `-tags wincompat`.
5. **C5** — No live network, GitHub API, or LLM. Isolate `APPDATA`,
   `AppData`, `USERPROFILE`, `HOME`, and `XDG_CONFIG_HOME` in every test
   that could load config or run the hook.
6. **C6** — On this host the required cases do not `t.Skip`. A skip on
   another machine must name the missing capability in the skip text.
7. **C7** — Tests do not assert Unix permission bits on Windows.
8. **C8** — No existing test is deleted or weakened.

**Most at risk:** C1. A failing hook-discovery or rename-retry test will
make it tempting to "just fix" `make install` or the retry loop. That is a
different change and needs an amended pair.

## Dependency and delivery order

P1 (entrypoints + fail-closed guard) → P2 (git flags + Git for Windows) →
P3 (fsutil) and P4 (config) in either order after P2 → P5 (hook `.exe`)
depends on P1 and should follow P2 → P6 (self-update) after P1 → P7 (docs
index and developer instructions) last so the commands it documents exist.

P3 and P4 may be one commit each and do not depend on each other.

## Implementation Steps

### P1 — On-demand entrypoints and fail-closed guard (D1, D2; closes F1, F2, F8)

**Files:** `wincompat_guard_test.go`, `scripts/test-windows.sh`,
`scripts/test-windows.ps1`, `scripts/verify-scripts.sh`, `Makefile`.

1. Add `wincompat_guard_test.go` with `//go:build wincompat` (not `windows`).
   If `runtime.GOOS != "windows"` the test fatals with the canonical native
   command and the WSL re-exec hint. On Windows it is a no-op success so the
   tag always compiles at least one test in package `main`.
2. Change `scripts/verify-scripts.sh` so `bash -n` iterates only `*.sh`
   under `scripts/` and `.githooks/` (still sorted, still `set -euo
   pipefail`). Keep the `actionlint` invocation unchanged.
3. Add `scripts/test-windows.sh`:
   * `set -euo pipefail`; `cd` to repo root via `git rev-parse --show-toplevel`.
   * If `go env GOOS` is `windows`, exec
     `go test -tags wincompat -count=1 ./...`.
   * Else if `cmd.exe` is on PATH, exec
     `cmd.exe /C "cd /d <windows-path-of-root> && go test -tags wincompat -count=1 ./..."`.
     Resolve the Windows path with `pwd -W` when Git Bash provides it,
     otherwise `wslpath -w` when present, otherwise a documented failure.
   * Else print the canonical command and exit 1.
4. Add `scripts/test-windows.ps1` that fails unless
   `[System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform(Windows)`
   (or `$env:OS -eq 'Windows_NT'`) and then runs
   `go test -tags wincompat -count=1 ./...` from the repo root. `$ErrorActionPreference = 'Stop'`.
5. Add Makefile target `test-windows` with a `##` help string, listed in
   `.PHONY`, invoking `./scripts/test-windows.sh`. Do not add it as a
   prerequisite of `verify` or `test`.

**Verification:**

```text
# Untagged suite still ignores the guard file.
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test ./... -count=1"
# Tagged suite compiles the guard on Windows.
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test -tags wincompat -count=1 ./..."
bash -n scripts/test-windows.sh scripts/verify-scripts.sh
# verify-scripts must not bash -n the ps1:
./scripts/verify-scripts.sh
```

**Phase acceptance:** untagged test run output does not mention
`TestWincompatRequiresWindows` (or whatever the guard is named). Tagged
run does. `bash -n` on the ps1 is not performed by `verify-scripts.sh`.

### P2 — Git subprocess Windows contract (D3; closes F3)

**Files:** `internal/git/wincompat_test.go` (`//go:build windows && wincompat`).

1. Assert `newGitCmdContext(ctx, "git", "version")` returns a cmd whose
   `SysProcAttr` is non-nil, `HideWindow == true`, and
   `CreationFlags & createNoWindow == createNoWindow` (`0x08000000`).
2. Isolated temp repo (`t.TempDir`, `GIT_CEILING_DIRECTORIES`, local
   `user.name` / `user.email`, `core.autocrlf=true`): stage a Unicode path
   with spaces (for example `docs/你好 world.txt`) plus a `.ps1` and a
   `.go` file; `GatherInfo` must list the Unicode path unquoted, classify
   the script, and return a non-empty diff. `cmd.SysProcAttr` still applies
   because `GatherInfo` uses `newGitCmdContext`.
3. Do not `os.Chdir` if a helper can pass `cmd.Dir`; if Chdir is used,
   restore the previous directory in `t.Cleanup` and do not parallelize
   that test.

**Verification:** tagged `go test -tags wincompat -count=1 ./internal/git`
passes with no skips. Untagged `./internal/git` does not compile the new
file (the test binary's `-test.list` without the tag omits the new names).

### P3 — Atomic replace, sharing-violation retry, long path (D4; closes F4)

**Files:** `internal/fsutil/wincompat_test.go` (`//go:build windows && wincompat`).

1. Replace-over-existing file succeeds and content matches.
2. Sharing-violation retry: hold an exclusive share-mode `0` handle via
   `golang.org/x/sys/windows.CreateFile` on the destination; start
   `WriteFileAtomic` (or `ReplaceFileAtomic`) in this goroutine or a child;
   release the handle inside the retry window (~100ms + 200ms + 400ms as
   implemented); the write must succeed. A second case that never releases
   must return a rename error after retries (bound the test with
   `t.Deadline` / a timeout so it cannot hang).
3. Long path: nest a destination whose path length is `> 260` under
   `t.TempDir()`; `WriteFileAtomic` must succeed on this host. Do not skip
   here. Do not require `\\?\` prefix in the caller if Go/OS long-path
   support is enough — if it is not, **stop** (C1) rather than changing
   production in this phase.
4. After a successful write, do not `Stat` and require `0600`.

If `golang.org/x/sys` becomes a **direct** module because of the test
import, run `go mod tidy` in this phase only and keep it in this phase's
commit. That is a `go.mod` / `go.sum` touch not listed above — if tidy
changes them, add those two files to this phase's commit and record it in
the execution record. Do not upgrade the version.

**Verification:** tagged tests in `./internal/fsutil` pass with no skips.
Untagged `go test ./internal/fsutil` still passes.

### P4 — AppData config layout (D5; closes F5)

**Files:** `internal/config/wincompat_test.go` (`//go:build windows && wincompat`).

1. Isolate `HOME`, `USERPROFILE`, `AppData`, `APPDATA`,
   `XDG_CONFIG_HOME` to a temp tree.
2. `GetConfigPath()` must equal
   `filepath.Join(appData, "prepare-commit-msg", "config.json")` where
   `appData` is the isolated roaming dir. The path must use `filepath`
   separators (backslash on this OS) and must not be a `\.config\` XDG
   layout unless `userConfigDir` failed (it must not fail in this fixture).
3. `Save` then `Load` round-trips `ActiveProvider` and an API key through
   that path. `os.Stat` the path to prove the file exists where asserted.

**Verification:** tagged `go test -tags wincompat -count=1 ./internal/config`
passes with no skips.

### P5 — Git for Windows hook discovery and soft-fail (D6; closes F6)

**Files:** `main_wincompat_test.go` (`//go:build windows && wincompat`).

1. `go build` the module to
   `t.TempDir()/hooks/prepare-commit-msg.exe` (use
   `exec.Command("go", "build", "-o", exe, ".")` with `t.Helper` and the
   module root as `Dir`).
2. Isolated git repo: `core.hooksPath` points at that hooks dir (or copy
   the exe into `repo/.git/hooks/prepare-commit-msg.exe`). Isolate
   `APPDATA` / `AppData` / `USERPROFILE` / `HOME` so `config.Load` cannot
   see the owner's real config.
3. Stage a file and `git -c user.name=... -c user.email=... commit -m
   unused` **without** `-n`/`--no-verify` so `prepare-commit-msg` runs.
   The commit must exit 0. The test must not set provider API keys.
4. Assert Git actually invoked the exe: either the commit succeeds only
   when the exe is present (control: missing exe still commits, which is
   Git's default — so prefer a side effect such as the exe being
   `os.Stat`-able plus `GIT_TRACE` or a wrapper). Minimum acceptable proof:
   replace the built exe with a tiny Go program that writes a marker file
   and execs the real binary, **or** set `GIT_TRACE2_EVENT` and require a
   hook record. Do not talk to a network LLM. If the real binary would
   call an LLM, the isolated empty config must take the `softFail` / empty
   staged-or-unconfigured path and still exit 0.

Prefer: hooks dir contains only `prepare-commit-msg.exe`; `git commit`
with an empty message file source (`prepare-commit-msg` hook args) via
`git commit -m "x"` still runs the hook then skips generation because
source is `message`. That would **not** exercise generation. To exercise
the hook body, commit **without** `-m` is interactive. Use
`git commit -F msgfile`? That is also source `message`.

Use Git's `prepare-commit-msg` invocation directly:

```text
exe COMMIT_EDITMSG
```

in the isolated repo (empty `COMMIT_EDITMSG` with only comments), with
`cmd.Dir = repo`, isolated env, and assert exit code 0. Separately, prove
Git discovers `.exe` by running:

```text
git -C repo commit --allow-empty -F NUL
```

No: `-F` is source `message` and the hook should skip. That still proves
Git **spawned** the hook if we wrap it.

**Required shape:** a wrapper `prepare-commit-msg.exe` (small Go program
written to a temp file and `go build`'d) that appends one line to a marker
file then `os.Exit(0)`. Place it on `core.hooksPath`. `git commit --allow-empty
-m test` must create the marker. That proves Git for Windows discovers
`.exe`. A second subtest runs the **real** built product against
`COMMIT_EDITMSG` with empty comments-only content and isolated config, and
asserts process exit 0 (soft-fail, no panic, no hang) with a 15s timeout.

**Verification:** tagged `go test -tags wincompat -count=1 -count=1 .` in
package main (timeout 60s) passes with no skips. No HTTP listen / no
provider env vars in the test process.

### P6 — Self-update Windows apply (D7; closes F7)

**Files:** `update_wincompat_test.go` (`//go:build windows && wincompat`).

1. Reuse the existing fixture helpers (`fixtureRelease`,
   `installTestUpdater`, `withTempTarget`) if they stay unexported in
   package `main` — they are already in `update_test.go` and are available
   to another `_test.go` in the same package even when the other file has a
   tighter build tag, **but only when the tag is on**. When `-tags
   wincompat` is set, `update_test.go` (untagged) still compiles. Call the
   helpers; do not duplicate the fixture.
2. `RawBuildKind=release`, `RawVersion=v1.2.0`, apply `--yes` to fixture
   `v1.3.0`. Assert the target `.exe` bytes become the fixture body.
   If `target.exe.old` exists, that is accepted; if not, that is also
   accepted (mcplib may unlink it). Failure is: apply error, or target
   bytes unchanged.

**Verification:** tagged tests in package main include this case and pass.

### P7 — Developer docs (D8)

**Files:** `README.md`, `docs/cicd-operations.md`, `docs/README.md`,
and the MADR/PLAN status lines if the owner accepted the pair in the same
turn (`status: accepted` / `status: in-progress` then `completed` only
after P1–P6 are done).

1. README Developer Experience table: add `make test-windows` / the
   canonical `go test -tags wincompat -count=1 ./...` command, stating it
   is **on-demand**, **native Windows**, not part of `make verify`.
2. `docs/cicd-operations.md`: one short subsection under routine
   verification, same commands, plus the WSL re-exec note for this host.
   State explicitly that CI `go-native` Windows remains untagged
   `go test ./...`.
3. `docs/README.md`: index 0005 MADR (Proposed or Accepted as instructed)
   and 0005 PLAN.

**Verification:** the documented commands are the commands P1 implemented.
`git diff --check` clean on the doc files.

## Verification (whole plan)

On this host, from native Windows:

```text
go test ./... -count=1
go test -tags wincompat -count=1 ./...
powershell.exe -NoProfile -File scripts\test-windows.ps1
```

From WSL in the repo:

```text
./scripts/test-windows.sh
make test-windows
```

Untagged `go test -list .` must not include `Wincompat` test names.
Tagged `-list` must include the guard, git flag test, fsutil retry test,
config AppData test, hook discovery test, and update apply test.

### Acceptance criteria (mapped to MADR Confirmation)

| # | Criterion | MADR |
| --- | --- | --- |
| A1 | Untagged native `go test ./... -count=1` passes and does not compile wincompat tests | D1, Confirmation default gate |
| A2 | Tagged native `go test -tags wincompat -count=1 ./...` passes with zero required-case skips on this host | D1, D3–D7, Confirmation on-demand suite |
| A3 | WSL `./scripts/test-windows.sh` re-execs native `go.exe` and passes | D2, F8 |
| A4 | `scripts/test-windows.ps1` passes under native PowerShell | D2 |
| A5 | `make test-windows` is documented and is not a `verify` prerequisite | D2, D8 |
| A6 | `scripts/verify-scripts.sh` still `bash -n`s every `*.sh` and does not `bash -n` the ps1 | D2, F8 |
| A7 | `newGitCmdContext` HideWindow + CREATE_NO_WINDOW asserted | D3 |
| A8 | GatherInfo parses Git-for-Windows Unicode + spaces + autocrlf=true | D3 |
| A9 | Atomic write retries through a released sharing violation; long path >260 works; no Unix mode assert | D4 |
| A10 | GetConfigPath is `%AppData%\prepare-commit-msg\config.json` under isolation | D5 |
| A11 | Git discovers `prepare-commit-msg.exe`; real binary exits 0 with isolated empty config and no network | D6 |
| A12 | Fixture self-update apply replaces a `.exe` target | D7 |
| A13 | CI workflow file is untouched | D8, C3 |

**Most likely to be quietly dropped:** A11 (hook `.exe` discovery). It is
the slowest test and the one that most closely matches how this host
actually installs the product. Do not ship the plan without it.

## Rollout and Rollback

Rollout is local commits only, one per phase, no push, no tag, no CI
workflow change. After P7 the owner runs the Confirmation commands on this
host.

Rollback of any phase is `git revert` of that phase's commit. Because C1
forbids production Go changes, revert cannot break the default suite.

## Deferred (named, so they are not mistaken for oversights)

* Always-on or `workflow_dispatch` CI job for `-tags wincompat` on
  `windows-2025` — different machine than this host; needs its own decision.
* `make install` writing `prepare-commit-msg.exe` into a Windows hooks
  directory — product change; P5 tests Git discovery, not Makefile install.
* Replacing `syscall` in `cmd_windows.go` with `golang.org/x/sys/windows` —
  production cleanup, not required to assert `CREATE_NO_WINDOW`.
* Adding `.gitattributes` `eol=lf` — repo hygiene, not a runtime test.
* Porting hook installer scripts to PowerShell — adjacent DX, not this suite.
* Native `windows/arm64` execution — no ARM64 Windows host here.
