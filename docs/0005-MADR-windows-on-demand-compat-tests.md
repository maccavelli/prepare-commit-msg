---
status: proposed
date: 2026-09-14
decision-makers: Project Owner
consulted: none
informed: none
---
<!-- markdownlint-disable MD013 MD024 MD033 MD036 MD060 -->

# Windows compatibility suite as a Python pre-commit hook, skipped on Unix

## Context and Problem Statement

This repository already compiles Windows binaries and runs the portable Go
test suite on a GitHub `windows-2025` runner. That is not the same as proving
Windows-specific behavior on a real Windows workstation: Git for Windows hook
discovery, `CREATE_NO_WINDOW`, AppData config layout, exclusive-file rename
retries, Unicode paths, long paths, and `.exe` self-update replacement.

The owner asked for Windows-specific tests that run **on this host**. On
2026-09-14 that request was revised: the suite must **fire from this
repository's pre-commit hook** when committing with Windows Git; it must
**skip on Linux and macOS**; and **all new scripting for this work must be
Python, never bash** (and not PowerShell).

The suite still must not join `make test`, `make verify`, or the CI
`go-native` Windows job. Existing fleet bash (`install-hooks.sh`,
Makefile recipes) is not rewritten in this decision. The staged-Go gate is
already Python ([`scripts/go-precheck.py`](../scripts/go-precheck.py)).

### What was measured, not assumed

Measured on 2026-09-14 against worktree
`C:\Users\<user>\gitrepos\prepare-commit-msg` (WSL path
`/mnt/c/Users/<user>/gitrepos/prepare-commit-msg`).

**Host and toolchain**

* Agent default shell is WSL2 (`uname -a`: `Linux MAC420
  6.6.87.2-microsoft-standard-WSL2`). `go` is **not** on the WSL PATH.
  WSL `python3` is `/usr/bin/python3` **3.12.3**, `sys.platform=linux`.
* Native Windows Go, via `cmd.exe /C "go version && go env GOOS GOARCH"`:
  `go version go1.26.6 windows/amd64`, `GOOS=windows`, `GOARCH=amd64`.
* Native Git is Git for Windows `2.53.0.windows.2` at
  `C:\Program Files\Git\cmd\git.exe`.
* Native `make` is not on PATH. `mingw32-make.exe` exists at
  `C:\Users\<user>\toolchains\mingw64\bin\mingw32-make.exe`.
* Real native Python is
  `C:\Users\<user>\AppData\Local\Python\pythoncore-3.14-64\python.exe`
  **3.14.3**, `sys.platform=win32`. Git Bash `command -v python3` lists
  `WindowsApps` first, but `python -c` resolved to the 3.14 core
  (`sys.executable` was the `pythoncore-3.14-64` binary).
* `go test -race -count=1 ./internal/git` under native Windows passed in 2.076s.

**Git for Windows, hooks, and OS policy**

* System config: `core.autocrlf=true`, `core.symlinks=false`,
  `http.sslBackend=schannel`. Global config overrides `core.autocrlf=false`.
* System `core.longpaths` is `true`. Registry `LongPathsEnabled` is `0x1`.
* Native Git **global** `core.hooksPath` is
  `C:/Users/<user>/.global-git-hooks` (contains `prepare-commit-msg.exe` and
  `pre-push` / `post-*`). Native Git **local** `core.hooksPath` for this
  repo is **unset**. Effective hook dir is therefore the **global**
  directory, not `.githooks/` and not `.git/hooks`.
* WSL Git in the same worktree has **no** local `core.hooksPath`; effective
  hooks are `.git/hooks` (sample files only). WSL Git does not share Git
  for Windows' global `hooksPath`.
* [`.githooks/pre-commit`](../.githooks/pre-commit) is bash:
  `exec make -C "$REPO_ROOT" verify-staged`. It does not run today under
  Git for Windows because local hooks are not installed.
* [`scripts/install-hooks.sh`](../scripts/install-hooks.sh) sets **local**
  `core.hooksPath` to a managed directory that runs the previous hook then
  `.githooks/<name>`. That is the only in-repo mechanism that makes
  `.githooks/pre-commit` fire while still composing with the global
  `prepare-commit-msg.exe`.
* [`scripts/verify-scripts.sh`](../scripts/verify-scripts.sh) runs `bash -n`
  on **every** file under `scripts/` and `.githooks/`. A Python
  `.githooks/pre-commit` or `scripts/*.py` would fail `make verify` unless
  `bash -n` is restricted to `*.sh`.

**Current tests (native `go test ./...`)**

* Five packages ok (0.45s–1.32s). Aggregate coverage **83.7%**.
  `newGitCmdContext` is 100% statements with **no** `HideWindow` /
  `CREATE_NO_WINDOW` assertion. `WriteFileAtomic` 77.8% /
  `ReplaceFileAtomic` 79.2%. No `//go:build windows` test file.
* CI `go-native` on `windows-2025` is untagged `go test ./...`.
* Makefile has no Windows test target. `install` is Unix-only (no `.exe`
  destination).

### Findings

1. **F1** — Default `go test ./...` on this Windows host is green, so the
   portable suite is not a Windows-compat suite.
2. **F2** — The suite must not join `make test`, `make verify`, or CI
   `go-native`. The owner now wants it **on every Windows commit in this
   repo**, not as a remembered extra command.
3. **F3** — `internal/git/cmd_windows.go` sets `HideWindow` and
   `CREATE_NO_WINDOW`. Nothing asserts those flags or Git-for-Windows
   `GatherInfo` with Unicode paths, spaces, and `core.autocrlf=true`.
4. **F4** — `internal/fsutil` retry-on-lock, long paths, and “Unix modes
   are not required” are untested.
5. **F5** — Config is `%AppData%\prepare-commit-msg\config.json`. Tests
   isolate `AppData` but never assert that layout.
6. **F6** — No test proves Git for Windows discovers
   `prepare-commit-msg.exe` and that the product hook exits 0 without a
   live LLM.
7. **F7** — Self-update tests name `.exe` assets but do not apply a
   fixture replace on this OS.
8. **F8** — Dual environment: WSL is Linux (`sys.platform=linux`, no
   `go`); Git for Windows is `win32` with `go.exe`. A Linux hook process
   must skip; a Windows hook process must run native `go test`.
9. **F9** — Git for Windows will **not** run `.githooks/pre-commit` until
   local `core.hooksPath` is installed (`make hooks-install`). Without
   that step the suite cannot fire on commit, because global hooksPath
   wins.
10. **F10** — New scripting must be Python. Git Bash `env python3` is
    usable on this host (resolves to Python 3.14 `win32`). `bash -n` on
    Python files would break `make verify`.
11. **F11** — Existing `.githooks/pre-commit` is bash and calls `make
    verify-staged`. Native Windows has no `make` on PATH. Replacing that
    file with Python is required; rewriting `install-hooks.sh` is not.
    The staged-Go gate `scripts/go-precheck.py` is already Python.

## Decision Drivers

* Run the Windows-compat assertions automatically on commit **from this
  repo** when the hook process is Windows.
* Skip those assertions when the hook process is Linux or macOS
  (including WSL Git).
* Author no new bash or PowerShell. Python 3 stdlib only.
* Keep `make verify` and CI `go-native` unchanged in meaning.
* Preserve 0003 hook composition (global `prepare-commit-msg.exe` still
  runs) via the existing installer, not a second `hooksPath`.
* No live network or LLM during the suite.

## Considered Options

* **A** — Go tests behind `-tags wincompat`; a stdlib Python runner;
  `.githooks/pre-commit` rewritten in Python to run the suite on
  `win32` and skip otherwise; `make hooks-install` so Git for Windows
  actually invokes it.
* **B** — Keep a manual `go test -tags wincompat` command only (the
  2026-09-14 morning draft). Rejected: the owner required a pre-commit.
* **C** — Put the suite in the **global** `pre-commit` under
  `C:/Users/<user>/.global-git-hooks`. Rejected: that would run for every
  repository on the machine.
* **D** — New bash/PowerShell drivers plus a bash one-liner in
  `.githooks/pre-commit`. Rejected: the owner forbade new bash.
* **E** — Always-on CI `windows-2025` job. Rejected: still not this host,
  and not a pre-commit.

## Decision Outcome

### The decisions

1. **D1** — Add an opt-in Go test suite compiled only with
   `-tags wincompat`. Untagged `go test ./...` must not compile those tests.
2. **D2** — The only driver is Python 3 stdlib:
   `scripts/run_wincompat.py`. Canonical test command (already on
   Windows): `python scripts/run_wincompat.py` →
   `go test -tags wincompat -count=1 ./...`. No new `.sh` or `.ps1`.
   On `win32` it runs the suite and non-zero fails. On Linux/macOS it
   prints one skip line to stderr and exits 0. A `--self-test` mode
   checks skip/run branching without requiring `go`.
3. **D3** — Same-package Windows tests must assert `newGitCmdContext`
   sets `HideWindow` and `CREATE_NO_WINDOW`, and that `GatherInfo` using
   real `git.exe` accepts Unicode names, names with spaces, and a test repo
   with `core.autocrlf=true`.
4. **D4** — Same-package tests must assert atomic replace over an existing
   file, retry-and-succeed when a sharing violation is released during the
   retry window, long-path write (>260 characters) on this host, and that
   Unix file modes are not required for success on Windows.
5. **D5** — Config tests must assert
   `GetConfigPath()` resolves under `%AppData%\prepare-commit-msg\config.json`
   when `AppData`/`APPDATA` is isolated, and that Save/Load round-trips
   there.
6. **D6** — An integration test must build `prepare-commit-msg.exe`, prove
   Git for Windows discovers a `prepare-commit-msg.exe` hook, and prove the
   real binary exits 0 against an isolated empty config with no network LLM.
7. **D7** — Self-update tests under the tag must apply a fixture release to
   a Windows `.exe` target on this OS and accept the mcplib Windows replace
   (including a `.old` leftover if the installer creates one).
8. **D8** — Document the Python command, the Windows-only pre-commit, the
   Unix skip, and `make hooks-install` in README and
   `docs/cicd-operations.md`. Do **not** add the suite to CI `go-native`
   or `make verify`.
9. **D9** — Replace [`.githooks/pre-commit`](../.githooks/pre-commit) with
   a Python 3 script (`#!/usr/bin/env python3`) that: (1) on all
   platforms, runs the existing staged-Go gate when it can (`make
   verify-staged` if `make` exists; otherwise `scripts/go-precheck.py`
   run with `sys.executable`, so no bash is involved); (2) on `win32`
   only, then runs
   `scripts/run_wincompat.py` and fails the commit on non-zero; (3) on
   Linux/macOS, skips (2) with one stderr line and does not re-exec
   Windows `go.exe`.
10. **D10** — Rollout on this host includes `make hooks-install` (existing
    installer) so Git for Windows' local `core.hooksPath` points at the
    managed wrappers that invoke `.githooks/pre-commit`. Without that,
    the hook cannot fire (F9). Do not point `core.hooksPath` at
    `.githooks` alone (that would drop the global `prepare-commit-msg.exe`).
11. **D11** — Restrict `scripts/verify-scripts.sh` `bash -n` to `*.sh`
    so Python files do not fail `make verify`. Do not add new bash
    programs. `Makefile` `test-windows` may invoke the Python driver;
    it is not a `verify` prerequisite.

### Consequences

* Good, because a Windows commit in this repo runs the extra assertions
  without relying on memory, CI, or bash.
* Good, because a Linux/macOS/WSL commit skips the suite and does not
  fail closed for missing Windows `go`.
* Good, because `install-hooks.sh` composition keeps the global product
  hook.
* Neutral, because tagged tests do not count toward `make coverage`.
* Neutral, because GitHub `windows-2025` still will not run the extra suite.
* Bad, because every Windows commit pays the wincompat runtime (includes a
  `go build` in D6). `git commit --no-verify` remains the emergency bypass.
* Bad, because until `hooks-install` is run on this host, Git for Windows
  still uses only the global hooks directory. D10 is mandatory for the
  owner's stated trigger.
* Good, because the staged-Go gate `scripts/go-precheck.py` is Python, so
  the hook runs it with `sys.executable` without starting Git Bash.

### Confirmation

```text
# Default gate unchanged (native Windows).
cmd.exe /C "cd /d C:\Users\<user>\gitrepos\prepare-commit-msg && go test ./... -count=1"
# Expected: five packages ok; no wincompat tests.

# Python driver on Windows runs the tagged suite.
cmd.exe /C "cd /d C:\Users\<user>\gitrepos\prepare-commit-msg && python scripts\run_wincompat.py"
# Expected: extra Windows tests run; required cases do not skip on this host.

# Python driver on WSL skips.
python3 scripts/run_wincompat.py
# Expected: one skip line on stderr, exit 0, no go test.

# python3 scripts/run_wincompat.py --self-test
# Expected: exit 0.

# After make hooks-install, a Windows git commit in this repo invokes
# .githooks/pre-commit (Python) and runs wincompat. A WSL git commit
# in this repo skips wincompat.

# make verify still lints only *.sh with bash -n.
# make verify does not run wincompat.
```

Required (non-skippable on this host, when the process is `win32`): D3
flags, D3 Git-for-Windows GatherInfo, D4 sharing-violation retry, D4 long
path, D5 AppData path, D6 hook `.exe` soft-fail, D7 Windows apply.

## Pros and Cons of the Options

### A — Python pre-commit + tagged Go tests (chosen)

* Good, because it matches the revised trigger, skip rule, and
  Python-only scripting constraint.
* Good, because same-package Go tests can still see unexported
  `newGitCmdContext`.
* Neutral, because `install-hooks.sh` remains bash (already shipped).
* Bad, because Windows commits get slower.

### B — Manual command only

* Good, because it is the smallest change and stays out of the hook path.
* Bad, because the owner required a pre-commit.

### C — Global hooksPath pre-commit

* Good, because it would fire without `hooks-install`.
* Bad, because every other repo on the machine would pay for it.

### D — New bash/PowerShell drivers

* Good, because Git for Windows already runs bash hooks.
* Bad, because the owner forbade new bash.

### E — Always-on CI job

* Good, because it would catch regressions without a local hook.
* Bad, because GitHub Windows is not this host and is not a pre-commit.

## More Information

### Evidence index

| Claim | Source |
| --- | --- |
| Native tests pass on windows/amd64 | `cmd.exe /C "go test ./... -count=1"` 2026-09-14 |
| `newGitCmdContext` 100% statements, no flag assert | native `go tool cover -func`; `rg SysProcAttr` only in `cmd_windows.go` |
| CI Windows job is untagged `go test ./...` | `.github/workflows/ci.yml` `go-native` |
| Native local hooksPath unset; global is `C:/Users/<user>/.global-git-hooks` | `git config --local/--global --get core.hooksPath` via `cmd.exe` |
| WSL Git effective hooks are `.git/hooks` | `git rev-parse --git-path hooks` in WSL |
| `.githooks/pre-commit` is bash `make verify-staged` | `.githooks/pre-commit` |
| `install-hooks.sh` sets local managed `core.hooksPath` | `scripts/install-hooks.sh` last lines |
| `verify-scripts.sh` bash-n's every file | `scripts/verify-scripts.sh` `find ... -type f` |
| Native Python 3.14 `win32`; WSL Python 3.12 `linux` | `sys.executable` / `sys.version` / `sys.platform` |
| Git Bash `python` reaches 3.14 core | Git Bash `python -c` `sys.executable` |
| Long paths enabled | `LongPathsEnabled=0x1`; system `core.longpaths=true` |

### Related records

* [0003-MADR-layer-and-harden-ci-cd-quality-gates.md](0003-MADR-layer-and-harden-ci-cd-quality-gates.md)
  — composable hooks and `verify-staged`; this decision adds a Windows-only
  extra step and does not replace that gate's meaning.
* [0004-MADR-align-cicd-with-magic-cli-remote.md](0004-MADR-align-cicd-with-magic-cli-remote.md)
  — CI `go-native` remains untagged.

### Open questions for the plan

* Exact stderr skip text (must be one line, must mention Windows).
* Whether the staged gate on Windows (no `make`) should always run
  `scripts/go-precheck.py` directly. Prefer running it with
  `sys.executable` whenever `make` is absent; it needs no bash, so the
  staged gate never has to be skipped for a missing shell.
* Python shebang `#!/usr/bin/env python3` vs `python` — use `python3`
  first with a fallback inside the hook if `env` cannot find it.
