---
status: proposed
date: 2026-09-14
---
<!-- markdownlint-disable MD013 MD024 MD033 MD036 MD060 -->

# PLAN 0005 — Windows compatibility suite as a Python pre-commit hook

Implements [0005-MADR-windows-on-demand-compat-tests.md](0005-MADR-windows-on-demand-compat-tests.md)
decisions D1–D11, closing findings F1–F11.

## Goal

After the last phase:

* Native untagged `go test ./... -count=1` still passes the five existing
  packages and does not compile any `wincompat` test.
* On this host, Windows Python `python scripts\run_wincompat.py` runs
  `go test -tags wincompat -count=1 ./...` and passes with **zero skips**
  for the required cases.
* WSL `python3 scripts/run_wincompat.py` prints one skip line and exits 0
  without invoking `go`.
* `.githooks/pre-commit` is Python 3 (shebang `#!/usr/bin/env python3`),
  not bash. On `win32` it runs the driver and fails the commit on
  non-zero. On Linux/macOS it skips the driver.
* After `make hooks-install`, Git for Windows in this repo has local
  `core.hooksPath` on the managed wrappers, so a Windows `git commit`
  invokes that Python hook.
* No new bash or PowerShell files exist. `scripts/verify-scripts.sh`
  `bash -n`s only `*.sh`.
* `make test` / `make verify` / CI `go-native` do not run wincompat.
* README and `docs/cicd-operations.md` document the Python command, the
  Windows-only pre-commit, the Unix skip, and `hooks-install`.

## Scope

### In scope (the only files any phase may touch)

* `docs/0005-MADR-windows-on-demand-compat-tests.md`
* `docs/0005-PLAN-windows-on-demand-compat-tests.md`
* `docs/README.md`
* `docs/cicd-operations.md`
* `README.md`
* `Makefile`
* `scripts/verify-scripts.sh`
* `scripts/run_wincompat.py`
* `.githooks/pre-commit`
* `wincompat_guard_test.go`
* `main_wincompat_test.go`
* `update_wincompat_test.go`
* `internal/git/wincompat_test.go`
* `internal/fsutil/wincompat_test.go`
* `internal/config/wincompat_test.go`

`go.mod` / `go.sum` may be added to **P3's commit only** if `go mod tidy`
promotes `golang.org/x/sys` to a direct test import. Do not upgrade it.

### Out of scope

* Production Go behavior unless a required test proves a bug — then stop,
  amend this pair, and re-approve.
* `.github/workflows/ci.yml` and any always-on CI job.
* `make test` / `make coverage` / `make verify` membership.
* Rewriting `scripts/install-hooks.sh`, `scripts/uninstall-hooks.sh`,
  `scripts/test-hooks.sh`, `scripts/go-precheck.sh`, or other existing
  bash into Python.
* Pointing `core.hooksPath` at `.githooks` alone (drops global
  `prepare-commit-msg.exe`).
* `make install` Windows `.exe` destination.
* Native `windows/arm64` execution.
* Adding pip dependencies, pytest, or a Python package.

## Stability rule

Every phase ends with **all** of:

```text
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test ./... -count=1"
python3 scripts/run_wincompat.py --self-test
git diff --check
```

From P1 onward, additionally:

```text
# WSL skip (must be exit 0):
python3 scripts/run_wincompat.py
# Native run (must exercise go test -tags wincompat):
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && python scripts\run_wincompat.py"
```

Do not `git push`. Do not create tags. Commit locally once per phase, with
only that phase's in-scope files. Do not stage unrelated dirty worktree
files. Do not add `.sh` or `.ps1` files.

If a required test cannot pass without a production change, **stop** and
amend this pair. Do not weaken the test.

## Cross-cutting contracts

1. **C1** — No production Go behavior change. No new bash or PowerShell
   source. The only permitted bash *edit* is restricting
   `verify-scripts.sh` to `*.sh`.
2. **C2** — Untagged `go test ./...` compiles none of the new Go tests.
3. **C3** — CI workflow YAML is not modified.
4. **C4** — `make test` / `make coverage` / `make verify` do not invoke
   `-tags wincompat`.
5. **C5** — No live network, GitHub API, or LLM. Isolate `APPDATA`,
   `AppData`, `USERPROFILE`, `HOME`, and `XDG_CONFIG_HOME` in every Go
   test that could load config or run the product hook.
6. **C6** — On this host, required Go cases do not `t.Skip` when the
   process is native Windows.
7. **C7** — Tests do not assert Unix permission bits on Windows.
8. **C8** — No existing test is deleted or weakened.
9. **C9** — Python is 3.12-compatible stdlib only (`from __future__`
   not required). The hook and driver must not import anything outside
   the stdlib.
10. **C10** — Linux/macOS (including WSL) skip wincompat with exit 0.
    They must not re-exec `go.exe` / `cmd.exe` to force a Windows run
    from the hook.

**Most at risk:** C1 (new bash “just for the hook”) and C10 (re-exec from
WSL because that is how the morning draft worked). The owner forbade
both.

## Dependency and delivery order

P1 (Python driver, Python pre-commit, verify-scripts glob, guard test) →
P2 (git flags) → P3 (fsutil) and P4 (config) in either order → P5 (hook
`.exe`) after P2 → P6 (self-update) after P1 → P7 (docs + hooks-install
on this host).

## Implementation Steps

### P1 — Python driver, Python pre-commit, fail-closed guard (D1, D2, D9, D11; closes F1, F2, F8, F10, F11)

**Files:** `scripts/run_wincompat.py`, `.githooks/pre-commit`,
`scripts/verify-scripts.sh`, `Makefile`, `wincompat_guard_test.go`.

1. `scripts/run_wincompat.py` (UTF-8, stdlib):
   * Resolve repo root: `git rev-parse --show-toplevel` via
     `subprocess`, else walk parents from `__file__`.
   * If `--self-test`: assert skip/run helpers; do not call `go`; exit 0.
   * If `sys.platform` is not `win32`: print exactly one stderr line
     starting with `wincompat: skipped` and containing `not Windows`;
     exit 0.
   * Else: `subprocess.run(["go", "test", "-tags", "wincompat",
     "-count=1", "./..."], cwd=root, check=False)` and exit with that
     return code. Use `os.environ` copy; do not add provider API keys.
2. Replace `.githooks/pre-commit` with Python 3:
   * Shebang `#!/usr/bin/env python3`.
   * Keep the file executable (`100755`).
   * Run staged-Go gate first: if `shutil.which("make")`,
     `make -C root verify-staged`; else if `scripts/go-precheck.sh`
     exists and a `bash` executable exists, invoke that **existing**
     script; else print a one-line warning that the staged gate could
     not run. Non-zero from the staged gate fails the hook.
   * Then import/run the wincompat driver the same way as
     `python scripts/run_wincompat.py` (subprocess of
     `sys.executable` on that file, not a bash wrapper).
   * Do not re-exec Windows from Linux.
3. `scripts/verify-scripts.sh`: `bash -n` only files matching `*.sh`
   under `scripts/` and `.githooks/`. Keep `actionlint` unchanged.
4. Makefile: add `.PHONY` `test-windows` with help text, invoking
   `python3 scripts/run_wincompat.py` (or `python` if that is what
   `$(PYTHON)` detects). Do **not** add it to `verify` or `test`.
5. `wincompat_guard_test.go` with `//go:build wincompat` (not
   `windows`): if `runtime.GOOS != "windows"` fatal with the Python
   command. On Windows it is a no-op success.

**Verification:**

```text
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test ./... -count=1"
python3 scripts/run_wincompat.py --self-test
python3 scripts/run_wincompat.py
# expect skip, exit 0
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && python scripts\run_wincompat.py"
# expect go test -tags wincompat (guard only in P1)
python3 -m py_compile scripts/run_wincompat.py .githooks/pre-commit
./scripts/verify-scripts.sh
# must not bash -n the python files
git ls-files -s .githooks/pre-commit
# mode 100755
```

**Phase acceptance:** no new `.sh`/`.ps1`. Untagged `go test` does not
list the guard. WSL driver skips. Native driver runs tagged tests.

### P2 — Git subprocess Windows contract (D3; closes F3)

**Files:** `internal/git/wincompat_test.go` (`//go:build windows && wincompat`).

1. Assert `newGitCmdContext` has `SysProcAttr.HideWindow == true` and
   `CreationFlags & createNoWindow == createNoWindow`.
2. Isolated temp repo with `core.autocrlf=true`, Unicode path with
   spaces (for example `docs/你好 world.txt`), a `.ps1` and a `.go`
   file staged; `GatherInfo` lists the Unicode path unquoted and returns
   a non-empty diff.
3. Restore any `Chdir` in `t.Cleanup`; do not parallelize a Chdir test.

**Verification:** native `python scripts\run_wincompat.py` includes
these tests with no skips. Untagged `go test ./internal/git` unchanged.

### P3 — Atomic replace, sharing-violation retry, long path (D4; closes F4)

**Files:** `internal/fsutil/wincompat_test.go` (`//go:build windows && wincompat`).

1. Replace-over-existing succeeds.
2. Exclusive share-mode `0` handle via `golang.org/x/sys/windows.CreateFile`;
   release inside the retry window; write succeeds. Never-released case
   returns a rename error and cannot hang (timeout).
3. Destination path length `> 260` under `t.TempDir()` succeeds on this
   host. Do not skip. If it fails, **stop** (C1).
4. Do not require mode `0600`.

**Verification:** tagged fsutil tests pass with no skips.

### P4 — AppData config layout (D5; closes F5)

**Files:** `internal/config/wincompat_test.go` (`//go:build windows && wincompat`).

1. Isolate `HOME`, `USERPROFILE`, `AppData`, `APPDATA`,
   `XDG_CONFIG_HOME`.
2. `GetConfigPath()` equals
   `filepath.Join(appData, "prepare-commit-msg", "config.json")`.
   Not an XDG `\.config\` layout.
3. Save/Load round-trip; `os.Stat` that path.

**Verification:** tagged config tests pass with no skips.

### P5 — Git for Windows hook discovery and soft-fail (D6; closes F6)

**Files:** `main_wincompat_test.go` (`//go:build windows && wincompat`).

1. Wrapper `prepare-commit-msg.exe` (tiny Go program) writes a marker
   then exits 0; `core.hooksPath` points at it;
   `git commit --allow-empty -m test` creates the marker.
2. Real `go build -o prepare-commit-msg.exe .` invoked as the hook
   against a comments-only `COMMIT_EDITMSG` with isolated
   `APPDATA`/`USERPROFILE`/`HOME`; process exit 0 within 15s; no
   provider env vars; no network.

**Verification:** tagged package-main tests pass with no skips, timeout 60s.

### P6 — Self-update Windows apply (D7; closes F7)

**Files:** `update_wincompat_test.go` (`//go:build windows && wincompat`).

1. Reuse `fixtureRelease`, `installTestUpdater`, `withTempTarget` from
   `update_test.go` (same package; untagged file still compiles under
   the extra tag).
2. Apply fixture `v1.3.0` with `--yes` from `v1.2.0` release; target
   `.exe` bytes become the fixture body. `.old` leftover optional.

**Verification:** tagged apply test passes.

### P7 — Docs, index, and this-host hook install (D8, D10; closes F9)

**Files:** `README.md`, `docs/cicd-operations.md`, `docs/README.md`.

1. README Developer Experience: Python driver, Windows-only pre-commit,
   Unix skip, `make hooks-install` required for Git for Windows, not
   part of `make verify`, `--no-verify` emergency bypass.
2. `docs/cicd-operations.md`: same facts; CI `go-native` remains
   untagged.
3. `docs/README.md`: index 0005.
4. On this host, run existing `make hooks-install` (or
   `scripts/install-hooks.sh` from WSL / Git Bash) so native
   `git config --local --get core.hooksPath` is the managed directory.
   This step **mutates local git config only**, not a tracked file.
   Record the resulting path in the PLAN execution record. Do not
   commit `.git/` state.

**Verification:** documented commands match P1. Native
`git config --local --get core.hooksPath` is non-empty and under this
repo's `.git/`. `git diff --check` clean on tracked docs.

## Verification (whole plan)

Native Windows:

```text
go test ./... -count=1
python scripts\run_wincompat.py --self-test
python scripts\run_wincompat.py
```

WSL:

```text
python3 scripts/run_wincompat.py --self-test
python3 scripts/run_wincompat.py
# skip, exit 0
python3 -m py_compile scripts/run_wincompat.py .githooks/pre-commit
```

Hook wiring (Windows Git, after P7 install):

```text
git config --local --get core.hooksPath
# managed prepare-commit-msg-hooks directory
```

A dry invocation of `.githooks/pre-commit` with Windows Python must
start `go test -tags wincompat`. The same file under WSL Python must
skip wincompat.

### Acceptance criteria (mapped to MADR Confirmation)

| # | Criterion | MADR |
| --- | --- | --- |
| A1 | Untagged native `go test ./...` passes and does not compile wincompat tests | D1 |
| A2 | Windows `python scripts\run_wincompat.py` passes tagged tests with zero required-case skips | D2, D3–D7 |
| A3 | WSL `python3 scripts/run_wincompat.py` skips, exit 0, no `go` | D2, D9, C10 |
| A4 | `.githooks/pre-commit` is Python 3, mode `100755`, not bash | D9, F11 |
| A5 | No new `.sh` or `.ps1` files | D2, C1 |
| A6 | `verify-scripts.sh` `bash -n`s only `*.sh` | D11, F10 |
| A7 | `newGitCmdContext` HideWindow + CREATE_NO_WINDOW asserted | D3 |
| A8 | GatherInfo parses Unicode + spaces + autocrlf=true | D3 |
| A9 | Sharing-violation retry + long path >260; no Unix mode assert | D4 |
| A10 | GetConfigPath is `%AppData%\prepare-commit-msg\config.json` | D5 |
| A11 | Git discovers `prepare-commit-msg.exe`; real binary exits 0, no network | D6 |
| A12 | Fixture self-update apply replaces a `.exe` target | D7 |
| A13 | CI workflow untouched; `make verify` does not run wincompat | D8, C3, C4 |
| A14 | After hooks-install, native local `core.hooksPath` is the managed dir | D10, F9 |
| A15 | Makefile `test-windows` calls the Python driver and is not a `verify` prereq | D11 |

**Most likely to be quietly dropped:** A14 (hooks-install on this host).
Without it the owner will commit with Git for Windows and the suite will
never run. Do not close the plan without recording that local config.

## Rollout and Rollback

Rollout is local commits only, one per phase, no push, no tag. After P7,
`make hooks-install` on this host (local git config only).

Rollback of tracked files is `git revert` of that phase's commit. To
stop the Windows pre-commit trigger without reverting tests:
`make hooks-uninstall` (restores previous local `core.hooksPath`).
`git commit --no-verify` remains the emergency bypass.

## Deferred (named, so they are not mistaken for oversights)

* Always-on CI `-tags wincompat` on `windows-2025` — different machine;
  needs its own decision.
* Rewriting 0003 bash (`install-hooks.sh`, `go-precheck.sh`,
  `verify-scripts.sh` itself) into Python — owner rule applies to **new**
  scripting in this pair, not a fleet-wide conversion.
* `make install` writing `prepare-commit-msg.exe` — product change.
* `syscall` → `golang.org/x/sys/windows` in `cmd_windows.go`.
* `.gitattributes` `eol=lf`.
* Native `windows/arm64`.
* Re-executing Windows `go.exe` from WSL — forbidden by D9/C10; maintainers
  who are in WSL and want the suite run Windows Python / Git for Windows
  instead.
