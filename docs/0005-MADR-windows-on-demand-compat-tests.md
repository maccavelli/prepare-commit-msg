---
status: proposed
date: 2026-09-14
decision-makers: Project Owner
consulted: none
informed: none
---
<!-- markdownlint-disable MD013 MD024 MD033 MD036 MD060 -->

# Add an on-demand native Windows compatibility suite without changing the default quality gate

## Context and Problem Statement

This repository already compiles Windows binaries and runs the portable Go
test suite on a GitHub `windows-2025` runner. That is not the same as proving
Windows-specific behavior on a real Windows workstation: Git for Windows hook
discovery, `CREATE_NO_WINDOW`, AppData config layout, exclusive-file rename
retries, Unicode paths, long paths, and `.exe` self-update replacement.

The owner asked for Windows-specific **on-demand** tests that can be run on
**this host** to raise Windows compatibility and platform-standards
adherence. The suite must be opt-in. It must not become another always-on
gate inside `make test`, `make verify`, or the CI `go-native` Windows job.

### What was measured, not assumed

Measured on 2026-09-14 against worktree
`C:\Users\macsm\gitrepos\prepare-commit-msg` (WSL path
`/mnt/c/Users/macsm/gitrepos/prepare-commit-msg`).

**Host and toolchain**

* Agent default shell is WSL2 (`uname -a`: `Linux MAC420
  6.6.87.2-microsoft-standard-WSL2`). `go` is **not** on the WSL PATH.
* Native Windows Go, via `cmd.exe /C "go version && go env GOOS GOARCH"`:
  `go version go1.26.6 windows/amd64`, `GOOS=windows`, `GOARCH=amd64`.
  `GOROOT` is the auto toolchain
  `C:\Users\macsm\go\pkg\mod\golang.org\toolchain@v0.0.1-go1.26.6.windows-amd64`.
* Native Git is Git for Windows `2.53.0.windows.2` at
  `C:\Program Files\Git\cmd\git.exe`.
* Native `make` is not on PATH. `mingw32-make.exe` exists at
  `C:\Users\macsm\toolchains\mingw64\bin\mingw32-make.exe`. WSL has GNU
  `make`. `gcc.exe` from the same MinGW tree is on the native PATH.
* `go test -race -count=1 ./internal/git` under native Windows passed in 2.076s
  (race detector is usable on this amd64 host).

**Git for Windows and OS policy**

* System config: `core.autocrlf=true`, `core.symlinks=false`,
  `http.sslBackend=schannel`. Global config overrides `core.autocrlf=false`.
* System `core.longpaths` is `true`. Registry
  `HKLM\SYSTEM\CurrentControlSet\Control\FileSystem\LongPathsEnabled` is
  `REG_DWORD 0x1`.
* Global `core.hooksPath` is `C:/Users/macsm/.global-git-hooks`. That
  directory already contains `prepare-commit-msg.exe` plus
  `pre-push` / `post-*` scripts. This host is a real Windows install target.
* The repository has **no** `.gitattributes` (search for `eol` / `crlf` /
  `gitattributes` in tracked files returned no matches).

**Current tests and coverage (native `go test ./...`)**

* `cmd.exe /C "go test ./... -count=1"` passed all five packages:
  `.` 0.687s, `internal/config` 0.467s, `internal/fsutil` 0.450s,
  `internal/git` 1.197s, `internal/ui` 1.320s.
* Native coverage (`go test -coverprofile` then `go tool cover -func`):
  aggregate **83.7%**. `internal/git/cmd_windows.go:newGitCmdContext` reports
  **100.0%** because `GatherInfo` calls it. There is **no** assertion that
  `SysProcAttr.HideWindow` or `CreationFlags` include `CREATE_NO_WINDOW`
  (`0x08000000`). `internal/fsutil.WriteFileAtomic` is **77.8%** and
  `ReplaceFileAtomic` is **79.2%** — the Windows retry-on-lock loop is the
  uncovered path.
* `rg 'runtime\.GOOS|t\.Skip|windows' --glob '*_test.go'` found Windows
  branching only in `update_test.go` (asset `.exe` suffix) and one
  `t.Skipf` in `internal/config/config_test.go` when a blocking-file fixture
  cannot be created. There is no `//go:build windows` test file.
* `internal/config.isolateHome` sets `AppData` so `os.UserConfigDir` works,
  but no test asserts the Windows path
  `%AppData%\prepare-commit-msg\config.json`.
* `TestIsCommitMsgEmpty_CRLF` exists and is portable. `GatherInfo` is not
  exercised with Git-for-Windows `core.autocrlf=true`, Unicode filenames, or
  paths containing spaces.

**Quality-gate surface**

* [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) job `go-native`
  on `windows-2025` runs `go test ./...` only — no race, no extra tags, no
  hook-discovery check.
* [`Makefile`](../Makefile) has `test`, `coverage`, `verify`, and
  `windows-amd64` / `windows-arm64` **build** targets. It has no
  `test-windows` target. `install` is documented as Unix and copies to
  `~/.global-git-hooks/$(BINARY_NAME)` **without** a `.exe` suffix.
* [`scripts/verify-scripts.sh`](../scripts/verify-scripts.sh) runs `bash -n`
  on **every** file under `scripts/` and `.githooks/`. A `*.ps1` added there
  would fail `make workflow-lint` / `make verify` unless the finder is
  restricted to shell scripts.
* [`docs/0003-MADR-layer-and-harden-ci-cd-quality-gates.md`](0003-MADR-layer-and-harden-ci-cd-quality-gates.md)
  already recorded that cross-compilation does not execute Windows
  implementations, and 0003's follow-up made portable test **assumptions**
  (config-path setup, filesystem error cases, executable-mode assertions).
  That follow-up removed Unix-only failures; it did not add Windows-only
  behavioral tests.

### Findings

1. **F1** — Default `go test ./...` on this Windows host is green, so the
   portable suite is not a Windows-compat suite. Coverage of
   `newGitCmdContext` is statement coverage without a Windows-contract
   assertion.
2. **F2** — The owner asked for **on-demand** tests on **this host**. Folding
   extra tests into `make test`, `make verify`, or CI `go-native` would make
   them always-on and would run host-sensitive cases on GitHub-hosted
   Windows, which is a different machine than this one.
3. **F3** — `internal/git/cmd_windows.go` sets `HideWindow` and
   `CREATE_NO_WINDOW` for GUI Git clients. Nothing asserts those flags, and
   nothing asserts `GatherInfo` still parses Git-for-Windows output with
   Unicode paths, spaces, and `core.autocrlf=true`.
4. **F4** — `internal/fsutil` documents Windows rename retries under AV /
   sharing locks, but tests only the happy path and parent-is-a-file errors.
   Retry success after a sharing violation, long paths (>260 with
   `LongPathsEnabled=1`), and “do not require Unix `0600`” are untested.
5. **F5** — Config is specified as
   `%AppData%\prepare-commit-msg\config.json`. Tests isolate `AppData` but
   never assert that layout, nor that `APPDATA` / `AppData` case folding
   works on Windows.
6. **F6** — Git for Windows discovers `prepare-commit-msg.exe` under
   `core.hooksPath` (this host already uses that layout). No test builds the
   Windows binary, places it as a hook, and proves `git commit` still
   succeeds (soft-fail, exit 0) without a live LLM.
7. **F7** — Self-update tests already append `.exe` to Windows asset and
   target names. They do not assert that a native apply leaves a replaceable
   target (and the documented `.old` rollover) on this OS.
8. **F8** — This workstation is dual-environment: WSL2 has `make` and no
   `go`; native Windows has `go.exe` / `git.exe` and no GNU `make` on PATH.
   The on-demand entrypoint must work from native `cmd`/`pwsh` and must be
   re-invokable from WSL via `cmd.exe` + `go.exe`. A PowerShell driver cannot
   land under `scripts/` until `verify-scripts.sh` stops `bash -n`-ing every
   file.

## Decision Drivers

* Prove Windows-specific product contracts on the maintainer's Windows host,
  not only that the portable tests compile and pass there.
* Keep the existing quality gate (`go test ./...`, `make verify`, CI
  `go-native`) unchanged in command and meaning.
* Make the extra suite impossible to confuse with the default suite: an
  explicit Go build tag plus a documented command.
* Fail closed if someone runs the tagged suite under Linux Go.
* Stay inside the test/docs/script surface; do not change hook runtime
  behavior, CI always-on jobs, or coverage thresholds to make the suite pass.
* No live network or LLM during the suite.

## Considered Options

* **A** — Opt-in Go tests behind `-tags wincompat`, plus small native
  drivers that refuse to run (or re-exec Windows `go.exe`) when `GOOS` is
  not `windows`.
* **B** — Put the Windows assertions into the default `go test ./...` suite
  so every Windows CI run and every local `make test` on Windows executes
  them.
* **C** — A PowerShell-only integration script with no Go tests.
* **D** — A second always-on CI job on `windows-2025` that runs the extra
  suite on every push.

## Decision Outcome

### The decisions

1. **D1** — Add an opt-in test suite compiled only with
   `-tags wincompat`. Untagged `go test ./...` must not compile those tests.
2. **D2** — The canonical command is
   `go test -tags wincompat -count=1 ./...` under native `GOOS=windows`.
   Convenience drivers (`scripts/test-windows.sh`,
   `scripts/test-windows.ps1`, `make test-windows`) must: (a) run that
   command when already on Windows; (b) re-exec native `go.exe` via
   `cmd.exe` when invoked from WSL on this host; (c) fail with the canonical
   command if neither is possible. Restrict `scripts/verify-scripts.sh` to
   `*.sh` so a PowerShell driver does not break `make verify`.
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
6. **D6** — An integration test must build `prepare-commit-msg.exe`, install
   it as a Git `prepare-commit-msg` hook in an isolated repo (isolated
   `APPDATA`/`USERPROFILE` so the owner's real keys are never read), and
   prove `git commit` exits 0 without contacting a network LLM.
7. **D7** — Self-update tests under the tag must apply a fixture release to
   a Windows `.exe` target on this OS and accept the mcplib Windows replace
   (including a `.old` leftover if the installer creates one).
8. **D8** — Document the on-demand command in the README developer section
   and `docs/cicd-operations.md`. Do **not** add the suite to CI `go-native`
   or `make verify`. A CI `workflow_dispatch` extra job is deferred.

### Consequences

* Good, because Windows-only contracts become executable assertions on the
  maintainer host without slowing Linux CI or changing the 80% coverage gate.
* Good, because a Linux `go test -tags wincompat ./...` fails closed instead
  of silently running zero extra tests.
* Good, because WSL and native Windows on this machine share one documented
  path to the same native suite.
* Neutral, because statement coverage of `cmd_windows.go` stays 100% in the
  default suite; the new value is assertions, not coverage points.
* Neutral, because GitHub `windows-2025` will not run the extra suite until
  a later decision.
* Bad, because host-sensitive tests (long paths, exclusive locks, Git for
  Windows) can fail on a different Windows machine. The suite is therefore
  on-demand and documented as host-run, not as a portable CI clone.
* Bad, because a PowerShell driver requires a one-line `verify-scripts.sh`
  glob change. That change is in scope and must keep `bash -n` on every
  existing `*.sh`.

### Confirmation

```text
# Default gate unchanged (native Windows).
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test ./... -count=1"
# Expected: same five packages ok; no wincompat tests in the run.

# On-demand suite (native Windows).
cmd.exe /C "cd /d C:\Users\macsm\gitrepos\prepare-commit-msg && go test -tags wincompat -count=1 ./..."
# Expected: extra Windows tests run; required cases do not skip on this host.

# Fail-closed off Windows (WSL Go, if present) OR WSL driver re-execs native go.exe.
# From WSL in this repo:
./scripts/test-windows.sh
# Expected: re-executes Windows go.exe and the tagged suite passes.

# PowerShell:
powershell.exe -NoProfile -File scripts\test-windows.ps1
# Expected: same tagged suite.

# make verify still includes workflow-lint after the *.sh filter.
# make verify must not be required to run the wincompat tests.
```

Required (non-skippable on this host) assertions: D3 flags, D3 Git-for-Windows
GatherInfo, D4 sharing-violation retry, D4 long path, D5 AppData path, D6 hook
`.exe` soft-fail commit, D7 Windows apply. A skip is allowed only for a named
absent capability on some other machine, with the skip text naming that
capability.

## Pros and Cons of the Options

### A — Opt-in `-tags wincompat` plus native drivers (chosen)

* Good, because it matches the request: on-demand, this host, extra depth.
* Good, because default CI and `make verify` stay the fleet contract from
  0003/0004.
* Good, because same-package tests can see unexported `newGitCmdContext`.
* Neutral, because developers must remember a second command.
* Bad, because tagged tests do not count toward `make coverage` — accepted,
  because they are not the coverage gate.

### B — Always-on Windows tests in the default suite

* Good, because Windows CI would execute the new assertions automatically.
* Bad, because that is not on-demand and couples this host's long-path /
  exclusive-lock / Git-for-Windows assumptions to `windows-2025`.
* Bad, because `make test` would grow slower and more environment-sensitive
  for every Windows contributor.

### C — PowerShell-only script, no Go tests

* Good, because it matches a pwsh daily driver.
* Bad, because it cannot assert unexported `SysProcAttr` without duplicating
  production code, and `go test` would not see the suite.
* Bad, because `verify-scripts.sh` currently cannot even `bash -n` a `.ps1`.

### D — Always-on extra CI job

* Good, because regressions would be caught without a maintainer remembering
  to run the suite.
* Bad, because the request is on-demand on this host, and GitHub Windows
  runners are not this host (AV, long-path policy, Git system config, GUI
  Git clients differ).

## More Information

### Evidence index

| Claim | Source |
| --- | --- |
| Native tests pass on windows/amd64 | `cmd.exe /C "go test ./... -count=1"` 2026-09-14, five `ok` lines |
| `newGitCmdContext` 100% statements, no flag assert | `go tool cover -func` on native coverage; `rg SysProcAttr` only in `cmd_windows.go` |
| Atomic retry loop lightly covered | `WriteFileAtomic` 77.8%, `ReplaceFileAtomic` 79.2% native cover |
| CI Windows job is untagged `go test ./...` | `.github/workflows/ci.yml` `go-native` / `Run native tests` |
| No Makefile Windows test target | `Makefile` `.PHONY` / `test:` / absence of `test-windows` |
| `verify-scripts.sh` bash-n's every file | `scripts/verify-scripts.sh` `find ... -type f` then `bash -n` |
| Git for Windows 2.53.0, autocrlf system true / global false | `git --version`; `git config --system/--global --list` |
| Long paths enabled | `reg query ... LongPathsEnabled` = `0x1`; system `core.longpaths=true` |
| This host already installs `prepare-commit-msg.exe` as a hook | `dir C:\Users\macsm\.global-git-hooks` |
| WSL has no `go`; native `go.exe` is 1.26.6 windows/amd64 | WSL `command -v go` failed; native `go version` |
| Race detector works here | native `go test -race -count=1 ./internal/git` ok |
| 0003 already noted cross-compile ≠ execute Windows | `docs/0003-MADR-layer-and-harden-ci-cd-quality-gates.md` Context |

### Related records

* [0003-MADR-layer-and-harden-ci-cd-quality-gates.md](0003-MADR-layer-and-harden-ci-cd-quality-gates.md) —
  native matrix and portable test fixes; does not define a Windows-only suite.
* [0004-MADR-align-cicd-with-magic-cli-remote.md](0004-MADR-align-cicd-with-magic-cli-remote.md) —
  `go-native` Windows job remains `go test ./...`.
* [0002-MADR-self-update-cli-and-github-releases-integration.md](decisions/0002-MADR-self-update-cli-and-github-releases-integration.md)
  — Windows `.old` replace is implemented in mcplib, consumed here.

### Open questions for the plan

* Exact skip policy text for machines without long paths (this host must not
  skip).
* Whether the WSL re-exec hard-codes `cmd.exe /C go` or searches
  `C:\Users\macsm\sdk\go1.26.5\bin\go.exe` when `go` is missing from the
  Windows PATH as seen by `cmd.exe` (on this host `where go` already finds
  it).
* Whether hook integration builds with `-race` or a plain `go test` binary;
  race is optional and must not be required for D6.
