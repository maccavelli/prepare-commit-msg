---
status: complete
date: 2026-10-19
associated-madr: "0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md"
---
# Implement replacing the repository Git hooks with `make preflight` and a cross-host test run

Associated MADR: [0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md](0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md)

## Goal

* **D1:** no clone carries the repository hook layer, and the repository no
  longer ships it; the global hooks still run.
* **D2:** `make preflight` runs `check-records`, `markdownlint` and
  `verify`, and CI's `go` job runs the same three.
* **D3:** `make pre-add-check FILES=…` exists; `make verify-staged` is its
  alias.
* **D4 (with A1):** `make preflight-hosts` runs CI's jobs for `HEAD` on the
  Linux host, WSL Ubuntu 24.04, the Windows host and macOS, through
  `tools/hosts.py`.
* **D5, D6:** AGENTS.md, the guide, the architecture note and the README
  describe the push procedure; 0003-MADR and 0013-MADR carry
  superseded-in-part banners.
* **Open item:** the Python script the owner saw fail is reproduced, or
  recorded as not reproduced, before anything is deleted.

## Facts this plan rests on

Each was established on 2026-10-19 at `bdb82d6`, read-only or in a scratch
clone. Hosts are named here by role only.

| # | Fact | How |
| :--- | :--- | :--- |
| F1 | Only this clone has a local `core.hooksPath` (the managed directory). The Windows host's clone has none; no clone was found within five levels of `~` on the Linux host or in WSL Ubuntu 24.04. | `git config --local --get core.hooksPath` on each clone found |
| F2 | `make verify` exits 0 in a scratch clone on macOS in 43 s; `tools-test` runs 9 tests, OK; `0 issues.`; `No vulnerabilities found.`; coverage 83.7%. | scratch clone, `make verify` |
| F3 | `scripts/check_records.py --check-all` exits 0; the disclosure guard exits 0 over the (empty) outgoing range. | run in this tree |
| F4 | A ~1.9 MB program on stdin reaches `python3 -` intact over each transport: `ssh -o BatchMode=yes -o ConnectTimeout=5 -T <dest> python3 -` (Linux; Windows, which lands in Git for Windows' bash), `ssh … <windows> wsl -d Ubuntu-24.04 -- bash -lc 'python3 -'`, and local. | SHA-256 compared on each side (MADR A1) |
| F5 | Every host: Python 3.14, Go 1.27.2, `git`, GNU Make, `npx`, `tar`; Linux hosts: `gcc`, `CGO_ENABLED=1`. macOS has GNU Make 3.81. | toolchain probe |
| F6 | A bundle of `HEAD --tags` is 1,349,091 bytes; its clone is detached at `HEAD`, has 24 tags, is clean, and `git describe --tags --always --dirty` works. | scratch clone |
| F7 | The managed `pre-push` runs the global `pre-push`, then `.githooks/pre-push` by absolute path under `set -euo pipefail`. | `.git/prepare-commit-msg-hooks/pre-push` |
| F8 | The fleet agent commit gate runs `scripts/go-precheck.sh` if it is executable, else `make pre-add-check` if the `Makefile` has a line starting `pre-add-check:`. This repository has neither. | `~/.agents/hooks/lib/precommit-checks.sh:42-53` |
| F9 | `make verify` is `Makefile:114`; the hook targets are `Makefile:128-135` and in `.PHONY` at `Makefile:27-30`; `scripts/verify-scripts.sh:14` scans `.githooks`. | read |
| F10 | CI's `go` job runs `make verify` (`ci.yml:49`) on `ubuntu-24.04`; `go-native` runs `go test ./...` (`ci.yml:81`) on `ubuntu-24.04`, `macos-15`, `windows-2025`. Neither runs `check-records` or `markdownlint`. | read |
| F11 | `scripts/go-precheck.py` with nothing staged exits 0: "all modules verified", both pinned modules resolved from GitHub. | run in this tree |
| F12 | WSL Ubuntu 24.04 runs OpenSSH, socket-activated (`ssh.socket`, port 2222). From macOS, a direct connection to the Windows host's port 2222 timed out (no `portproxy` rule); a jump through the Windows host to `localhost:2222` reached it, and this machine's key was refused (`publickey`). The other three distros have no listener of their own. | `ss -ltnp` as root in each distro; `ssh -J`, with a scratch `known_hosts` |

## Scope

### In scope

* New: `tools/hosts.py`, `tools/hostrun.py`, `tools/hosts.example.json`.
* Changed: `Makefile`, `.github/workflows/ci.yml` (the `go` job only),
  `scripts/verify-scripts.sh`, `tools/test_tools.py`, `tools/README.md`,
  `AGENTS.md`, `README.md`, `docs/guides/cicd-operations.md`,
  `docs/architecture.md`, `docs/README.md`, and the banners of
  `0003-MADR-layer-and-harden-ci-cd-quality-gates.md` and
  `0013-MADR-align-repository-with-workspace-scaffolding.md`.
* Added by Deviation 1: `tools/fakeprovider.py`, and a dated note in
  `0016-PLAN-keep-maintainer-tools-starting-with-a-fake-provider.md`.
* Deleted: `.githooks/pre-push`, `scripts/install-hooks.sh`,
  `scripts/uninstall-hooks.sh`, `scripts/test-hooks.sh`.
* This clone's `.git/config` (`make hooks-uninstall`, Phase 4).

### Out of scope

* The global hooks directory, the disclosure guard and the dotfiles
  repository. If Phase 1 shows the guard is the failing script, its fix
  goes there, under that repository's own record.
* Any Go source, `go.mod`, `go.sum`; `go-native`, `build` and `release` in
  `ci.yml`; the release spec.
* The three WSL distros that are not CI runners (MADR D4, the owner's
  choice). A maintainer may add them to the local config; no step depends
  on them.
* The host config itself: it lives outside the repository and is never
  committed.
* WSL Ubuntu 24.04's own SSH server (F12): WSL is entered through the
  Windows host's connection (MADR A1, the owner's choice). No step
  authorizes a key there or pins its host key.

## Rules for every phase

* **Branch.** On `main`, as 0016-PLAN did: the agent stages, the owner
  commits and pushes. The agent never commits or pushes to `main`.
* **Fail first.** Each new gate and test is run against a deliberately
  broken copy, in a scratch clone or temporary directory, never in this
  tree, and the failure is quoted in the Execution Record.
* **Gate, before staging:** `make preflight` (from Phase 2 on; in Phase 1,
  nothing is staged); `make tools-test` when `tools/` changed; the
  identifier scan of the staged diff for the local account name, the host
  aliases and the org domain, recorded by what was scanned for, never by
  the match. Each exit status is captured before any filter; long output
  goes to a scratch file.
* **Assert each edit landed** by reading the result, not the editor's
  report.
* **Deviations.** Stop, record a dated entry here (and in the MADR when a
  decision or a fact changes), offer resolutions that leave the tree
  correct, and wait for the owner's choice.

## Implementation Steps

### Phase 0: records

1. 0017-MADR is `accepted`, with amendment A1 `proposed`; this PLAN is
   `proposed`; both are indexed in `docs/README.md`. *(On approval, A1 is
   `accepted` and this PLAN `in-progress`.)*
2. `make check-records` and `make markdownlint` exit 0.
3. The owner reviews; approval of this PLAN approves A1. The owner commits.

### Phase 1: reproduce the failing Python script (MADR, More Information)

No file in the tree changes in this phase.

1. **Ask the owner** for what they saw: the host, the command (`git push`,
   or a `make` target), and the last lines of the output. If they have it,
   go straight to step 4 with it.
2. **Run the push chain's Python on each target,** in a scratch clone made
   from a bundle of `HEAD --tags` (F6), with a scratchpad Python script that
   runs argument lists over the F4 transports and captures each exit
   status before any filter:
   * `make tools-test`, `make workflow-lint` (whose `verify-scripts.sh`
     parses every Python script) and `make verify`, on the Linux host, WSL
     Ubuntu 24.04, the Windows host (in Git for Windows' bash; the
     `Makefile` selects Git's bash when `OS` is `Windows_NT`,
     `Makefile:5-8`) and macOS;
   * the disclosure guard over a non-empty range, `origin/main~5` to
     `HEAD`, with the ref line on stdin as AGENTS.md's command does, through
     the global `pre-push` on each target that has one: on macOS it runs
     `python3 github-disclosure.py`; on Windows, `pre-push.windows` is
     installed as `pre-push` and runs `python`. Whether the Linux host and
     WSL have a global `pre-push` is checked first, not assumed.
3. **Record each target's result,** as `<target>: <step> exit=N (Ns)`,
   with the failing step's last lines, redacted.
4. **Decide, with the owner:**
   * the failure is in the disclosure guard: record it, and the owner takes
     it to the dotfiles repository; this plan continues, because D5 still
     runs the guard;
   * it is in a `make verify` step: a dated deviation, with a fix proposed,
     before Phase 2;
   * nothing fails anywhere: record "not reproduced on any target" with the
     results, and continue. The removal in Phase 4 takes the hook off the
     push path either way.

### Phase 1a: fix the fake provider's race *(Deviation 1, 2026-10-19)*

1. **`tools/fakeprovider.py`:** `record` logs and returns whether the
   request was the last of `max_requests`; `Handler.answer` calls it
   *before* `send_response`, then writes and flushes the body, then starts
   the shutdown thread if `record` said so.
2. **`tools/test_tools.py`:** `setUp` registers `self.tmp.cleanup` with
   `addCleanup` first, so it runs after every server cleanup, and
   `tearDown` goes. A new test,
   `test_log_line_is_written_before_the_answer_arrives`, wraps the
   server's `record` with a 0.3 s delay, makes one request, and reads the
   log at once.
3. **Fail first:** the new test against the committed `fakeprovider.py`, on
   a scratch copy, fails.
4. **Stability:** the phase's `tools/` on scratch copies, `make
   tools-test`'s command 30 times each on macOS, the Linux host and the
   Windows host: 0 failures.
5. **0016-PLAN:** a dated note in its Execution Record: the flake, found
   by this PLAN's Phase 1, fixed under Deviation 1.
6. Gate: `make tools-test`, `make verify`, `make check-records`, `make
   markdownlint`, the identifier scan. The owner commits.

### Phase 2: `make preflight`, `make pre-add-check`, and CI (D2, D3)

1. **`Makefile`, `preflight`,** after `verify` (`Makefile:114`):

   ```make
   preflight: ## Runs every CI gate that runs off GitHub: records, Markdown, verify (0017-MADR)
   	@echo "==> check-records"; $(MAKE) --no-print-directory check-records
   	@echo "==> markdownlint"; $(MAKE) --no-print-directory markdownlint
   	@echo "==> verify"; $(MAKE) --no-print-directory verify
   	@echo "preflight passed"
   ```

   Each recipe line is its own shell under `-eu -o pipefail`, so a failing
   sub-make ends the target there, and `preflight passed` prints only after
   all three.
2. **`Makefile`, `pre-add-check`,** replacing `verify-staged`
   (`Makefile:116-117`):

   ```make
   FILES ?=
   pre-add-check: tools ## Checks the staged Go snapshot, or FILES=… (scripts/go-precheck.py)
   	$(PYTHON) scripts/go-precheck.py $(FILES)

   verify-staged: pre-add-check ## Alias of pre-add-check; earlier records cite it
   ```

   `check-records` and `markdownlint` gain `##` help text. `.PHONY` gains
   `preflight` and `pre-add-check`.
3. **`.github/workflows/ci.yml`, the `go` job,** two steps before "Run
   repository quality contract" (`ci.yml:48-49`):

   ```yaml
         - name: Check records and documentation links
           run: make check-records

         - name: Lint Markdown
           run: make markdownlint
   ```

   No `setup-node` or `setup-python` step is added: the step relies on the
   runner image's `python3` and `npx`, which Phase 5 confirms from the run's
   log. If either is missing there, that is a deviation, not a silent added
   action.
4. **Fail first,** each in a scratch clone, quoting the output:
   * a record with a broken relative link: `make preflight` exits non-zero
     after `==> check-records`, with no `==> markdownlint` line;
   * a guide with a Markdown violation: it stops after `==> markdownlint`,
     with no `==> verify` line;
   * a staged, unformatted Go file: it stops inside `verify` at
     `fmt-check`, and `preflight passed` is absent;
   * `ci.yml` with an unknown step key: `make workflow-lint` (actionlint)
     fails.
5. **`make -n pre-add-check FILES=main.go`** prints the `tools` bootstrap
   line (`./scripts/bootstrap-tools.sh`), then
   `python3 scripts/go-precheck.py main.go`; `make pre-add-check` exits 0
   with nothing staged, as the script alone does (F11); and F8's
   `grep -qE '^pre-add-check:' Makefile` exits 0.
6. **`make preflight`** exits 0 on the phase's tree and ends with
   `preflight passed`. Gate; the owner commits.

### Phase 3: `make preflight-hosts` and `tools/hosts.py` (D4, A1)

1. **`tools/hostrun.py`,** the program each host runs, standard library
   only, Python 3.12 or later. It is never imported on the host; its source
   is sent. `main(params)` with `params = {"label", "bundle_b64",
   "command"}`:
   * write the decoded bundle into `tempfile.mkdtemp(prefix="pcm-hosts-")`;
   * `git clone -q <bundle> <tmp>/work` (F6);
   * print `hostrun: <label>: python <version>, work in <path>`, flushed;
   * run `command` with `cwd=work`, `stdin=DEVNULL`, stderr merged into
     stdout, each line flushed;
   * remove the directory in `finally`, with an `onexc` that makes
     read-only files writable (git objects on Windows);
   * exit with the command's status, or 125 with `hostrun: setup failed:
     <reason>` if the clone fails.

   It ends with the fixed footer `if __name__ == "__main__":` /
   `sys.exit(main(json.load(sys.stdin)))`, which `hosts.py` replaces with a
   call carrying the parameters, after checking that the footer is there.
2. **`tools/hosts.py`,** with `#!/usr/bin/env python3`, a docstring citing
   0017-MADR D4 and A1, and:
   * **Config:** `--config PATH`, else
     `$XDG_CONFIG_HOME/prepare-commit-msg/hosts.json`, else
     `~/.config/prepare-commit-msg/hosts.json`. Shape: `{"hosts":
     [{"label", "kind", "ssh"?, "wsl"?, "command"?}], "identifiers": [...]}`.
     `label` matches `[A-Za-z0-9._-]+` and is unique; `kind` is `linux`,
     `windows` or `local`; `ssh` is required unless `local`; `wsl` (a
     distro name) only with `linux`; `command`, if given, is a list of
     strings. Any violation, or no file: exit 2, naming the path and the
     field.
   * **Default commands:** `linux` → `["make", "verify"]`; `windows` and
     `local` → `["go", "test", "./..."]` (MADR D4's table).
   * **Transports** (F4), as argument lists:
     * `local`: `[sys.executable, "-"]`;
     * `ssh`: `["ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=5",
       "-T", <ssh>, "python3", "-"]`;
     * `ssh` + `wsl`: the same prefix, then `["wsl", "-d", <wsl>, "--",
       "bash", "-lc", "'python3 -'"]`.

     Never `shell=True`. The `ssh` prefix is a module constant that the
     tests replace.
   * **What is tested:** `git status --porcelain` must be empty, or exit 2
     with "commit first: preflight-hosts tests HEAD". Then `git bundle
     create <tmp> HEAD --tags`, base64.
   * **Run:** every host in parallel (`ThreadPoolExecutor`). Each host's
     program goes on stdin, and its output, line by line, is redacted and
     written to `<run dir>/<label>.log`, where the run directory is
     `tempfile.mkdtemp(prefix="pcm-hosts-")` on this machine.
   * **Redaction:** drop SSH's own lines (post-quantum notices,
     "Pseudo-terminal"); replace `/home/<x>`, `/Users/<x>` and
     `<drive>:\Users\<x>` with `<user>`, and each configured identifier
     (escaped, case-insensitive) with `<redacted>`.
   * **Output:** one line per host, `<label>: exit=N (Ns)`, in config
     order; then, for each failing host, its last 20 lines; then
     `preflight-hosts: <passed>/<total> hosts passed; logs in <run dir>`.
     An unreachable host is `exit=255`.
   * **Flags:** `--only LABEL` (repeatable), `--config PATH`.
   * **Exit:** 0 when every host exits 0; 1 when any exits non-zero or is
     unreachable; 2 for a config or tree error (`tools/README.md`'s codes).
3. **`tools/hosts.example.json`:** the four default hosts with
   placeholders only (`<linux-ssh-destination>` for the `linux` host;
   `<windows-ssh-destination>` for the `windows` host and for the `linux`
   host with `"wsl": "Ubuntu-24.04"`; and a `local` host), and
   `"identifiers": ["<account-name>", "<host-name>", "<org-domain>"]`.
   JSON has no comments, so the extra-distro example goes in the README
   (step 6).
4. **`Makefile`:**

   ```make
   HOSTS_FLAGS ?=
   preflight-hosts: ## Runs CI's jobs for HEAD on the test hosts (tools/hosts.py; 0017-MADR)
   	$(PYTHON) tools/hosts.py $(HOSTS_FLAGS)
   ```

   It goes in `.PHONY`. It is not a prerequisite of `preflight` (MADR D4).
5. **`tools/test_tools.py`, a `HostsTest` class,** each test against a
   throwaway repository in a `TemporaryDirectory`, with one commit and the
   tag `v0.0.1`:
   * config: no file, a bad `kind`, a duplicate label, `wsl` on a
     `windows` host, and a non-list `command` each exit 2 and name the
     problem;
   * the three transports' argument lists, exactly, including
     `BatchMode=yes`, `ConnectTimeout=5` and `-T`;
   * a dirty tree exits 2, and no transport is started (a recording fake
     proves it);
   * redaction of each home form, an identifier, and SSH noise;
   * **local, end to end:** a `local` host whose command checks, in the
     clone, that `git describe --tags` prints `v0.0.1` and that the
     committed file exists, exits 0, and the tool exits 0;
   * a `local` host whose command exits 3: the line `exit=3`, the failing
     tail printed, and tool exit 1;
   * **fake `ssh`:** the prefix replaced by `[sys.executable, <fake>]`,
     where the fake records its argv to a file and then runs the rest
     through `python3 -` locally: the recorded argv is the ssh form, and
     the host exits 0; a fake that exits 255 without reading stdin gives
     `exit=255` and tool exit 1;
   * **cleanup:** the work directory the run printed no longer exists.
6. **`tools/README.md`:** a row for `hosts.py` (with `hostrun.py`, which is
   sent, not run); a section, "Cross-host run", covering the config path,
   the shape, the default commands, `HOSTS_FLAGS`, how to add an extra
   distro, and that the config is never committed.
7. **`scripts/verify-scripts.sh`** already parses every `tools/*.py`
   (`scripts/verify-scripts.sh:19-23`); `make workflow-lint` confirms that
   both new files parse.
8. **Fail first,** each on a scratch copy of `tools/`, quoting the failing
   test:
   * `BatchMode=yes` dropped: the transport test fails;
   * the dirty-tree check removed: the refusal test fails;
   * the `/Users/` redaction removed: the redaction test fails;
   * the `finally` removal skipped: the cleanup test fails;
   * the bundle made without `--tags`: the `describe` test fails.
9. **Live,** with the owner's config at the default path, outside the
   repository:
   * `make preflight-hosts` on a clean `HEAD`: four `exit=0` lines; each
     host's time recorded;
   * in a scratch clone, a committed `zz_windows_only_test.go` under
     `//go:build windows` that calls `t.Fatal`; `python3 tools/hosts.py`
     run there: the Windows host non-zero, the other three `exit=0`, tool
     exit 1;
   * a dirty scratch clone: exit 2; `--config /nonexistent`: exit 2;
   * the redacted logs and the printed output name no account, host alias
     or home path (the identifier scan run over them).
10. Gate (`make preflight`, so `tools-test` runs inside `verify`); the
    owner commits.

### Phase 4: remove the hook layer (D1), then the docs (D5, D6)

1. **Uninstall first,** on the only clone with the layer (F1): `make
   hooks-uninstall`. Assert:
   * `git config --local --get core.hooksPath` exits 1;
   * `git rev-parse --path-format=absolute --git-path hooks` names the
     global hooks directory;
   * `.git/prepare-commit-msg-hooks` does not exist.

   If F1 has changed since it was probed (a new clone with the layer), run
   the same on that clone first.
2. **Delete:** `git rm -r .githooks scripts/install-hooks.sh
   scripts/uninstall-hooks.sh scripts/test-hooks.sh`.
3. **`Makefile`:** remove `hooks-install`, `hooks-uninstall` and
   `hooks-test` (`Makefile:128-135`) and their `.PHONY` entries.
4. **`scripts/verify-scripts.sh:14`:** scan `scripts` only.
5. **`README.md`:** in "Build & Test Targets" (`README.md:712-724`), remove
   the two hook rows; add `make preflight`, `make preflight-hosts` and
   `make pre-add-check`.
6. **`docs/guides/cicd-operations.md`:** "Routine verification"
   (`:8-40`) becomes `make preflight`, then `make preflight-hosts` with the
   config and the rule for when to run it, and `make pre-add-check`. A new
   subsection, "A clone that still has the old hooks", says to run `make
   hooks-uninstall` before pulling this change; after pulling, run `git
   config --local --unset core.hooksPath` (or, if
   `.git/prepare-commit-msg-hooks/.previous-local-present` reads `yes`,
   restore the value in `.previous-local-value`), then remove
   `.git/prepare-commit-msg-hooks`. The rollback paragraph (`:235`) drops
   "use `make hooks-uninstall` for hooks and". The opening line also cites
   0017-MADR.
7. **`docs/architecture.md`:** `:41-44` describes `make preflight`, `make
   preflight-hosts` and the absence of repository hooks; the tree drops
   `.githooks/pre-push` (`:61`), and `scripts/` (`:59`) becomes
   "verification and repository utilities".
8. **`AGENTS.md`:** "Pre-add checks" names `make pre-add-check`; a new
   section, "Before a push", after it, gives D5's order (`make preflight`
   always; `make preflight-hosts` for the OS-sensitive list; the disclosure
   guard command from "Identifiers"), and says that an unreachable host is
   reported, never skipped silently.
9. **Banners:** `0003-MADR-…`'s "Partially superseded" note gains
   0017-MADR for its pre-push hook; `0013-MADR-…` gains a note at D8 that
   0017-MADR retires `.githooks/pre-push` and the installer. No rationale
   is rewritten.
10. **Assert:**
    * `git ls-files .githooks scripts/install-hooks.sh
      scripts/uninstall-hooks.sh scripts/test-hooks.sh` prints nothing;
    * `make -n hooks-install` exits non-zero with "No rule to make target";
    * a search for `hooks-install`, `hooks-uninstall`, `hooks-test`,
      `install-hooks`, `test-hooks` and `.githooks` in `Makefile`,
      `scripts/`, `tools/`, `.github/`, `README.md`, `AGENTS.md`,
      `docs/architecture.md` and `docs/guides/` finds only the recovery
      subsection. Records keep their history.
11. Gate. The owner's commit of this phase is the first without the layer:
    its message comes from the global `prepare-commit-msg` hook (MADR,
    Confirmation).

### Phase 5: push, and the first CI run

1. **Before the push** (D5), the agent runs `make preflight`, then `make
   preflight-hosts` (the change touches the `Makefile`, `ci.yml` and
   `tools/`), then AGENTS.md's disclosure-guard command over the outgoing
   range. All three exit 0, or the push waits.
2. The owner pushes.
3. **The CI run on `main`:** `gh run list --branch main --limit 1`, then
   `gh run view <id> --log` to a scratch file. Assert:
   * `go`: the "Check records and documentation links" and "Lint Markdown"
     steps succeeded, and the log shows `npx` and `python3` resolved from
     the image (Phase 2, step 3);
   * `go-native`: Linux, macOS and Windows succeeded;
   * `build`: the rehearsal succeeded.
4. Every Verification row is met: this PLAN goes to `complete` and A1 to
   `accepted`, in a records commit.

## Verification

| # | Criterion | MADR | Phase | Check |
| :--- | :--- | :--- | :--- | :--- |
| V1 | The failing script is reproduced and routed, or recorded as not reproduced on any target | More Information | 1 | the Phase 1 record |
| V2 | `make preflight` passes clean and stops at the named gate on each plant | D2 | 2 | the fail-first record; `make preflight` |
| V3 | `make pre-add-check` exists, `verify-staged` aliases it, and the fleet gate finds it | D3 | 2 | `make -n`; F8's `grep` |
| V4 | CI's `go` job runs `check-records` and `markdownlint` and passes | D2 | 5 | the run's log |
| V5 | `tools/hosts.py`'s tests pass and each failed on its plant | D4 | 3 | `make tools-test`; the fail-first record |
| V6 | `make preflight-hosts` passes four hosts on a clean `HEAD`, flags only the Windows host on a Windows-only failure, and refuses a dirty tree and a missing config | D4, A1 | 3 | the live record |
| V7 | No clone has the layer; the repository does not ship it; commits still get generated messages | D1 | 4 | Phase 4 steps 1, 10, 11 |
| V8 | The docs give D5's procedure; 0003 and 0013 carry banners | D5, D6 | 4 | read; `make check-records`; `make markdownlint` |
| V9 | Nothing committed names a host, account or real path; the guard passes | Drivers | every | the identifier scan; the guard |
| V10 | The first CI run is green on `go`, the three `go-native` runners and `build` | Confirmation | 5 | `gh run view` |

## Rollout and Rollback

* **Nothing reaches users:** no Go file changes, and no release is cut.
* **Order protects clones:** Phase 4 uninstalls before it deletes. A clone
  that pulls Phase 4 with the layer still installed fails every push (F7)
  until the guide's recovery is run.
* **Rollback, per phase:** revert that phase's commit.
  * Phase 4 reverted: the files return; a clone that wants the hook again
    runs `make hooks-install`.
  * Phase 3 reverted: `make preflight-hosts` is gone; the config outside the
    repository is inert.
  * Phase 2 reverted: CI returns to `make verify` alone.

## Execution Record

Each phase records what was done, the commands' exit statuses and quoted
output, each fail-first result, the hosts' times, and any deviation, dated.

### Phase 0: records (2026-10-19)

* The owner accepted the MADR ("accepted") and approved this PLAN and A1
  ("approved to proceed"). A1 is `accepted`; this PLAN is `in-progress`;
  `docs/README.md` indexes both records.
* The owner committed the records as drafted as `d12b24d`; the status
  changes above are staged on top of it. `make check-records` exit 0;
  `make markdownlint` exit 0; the identifier scan of the staged diff (the
  local account names, the host aliases and domains, the Windows host's
  address, home paths) found nothing.
* Before approval, the owner corrected one claim: WSL Ubuntu 24.04 does
  run an SSH server. It was probed (F12), A1's reason was struck and
  annotated, and the owner chose `wsl -d` through the Windows host's
  connection over that server.

### Phase 1: reproduce the failing Python script (2026-10-19)

No file in the tree changed. Step 1's question was put to the owner; the
reproduction (step 2) ran without waiting for the answer, and found the
script.

* **Step 2,** on a clone of `d12b24d` from a `HEAD --tags` bundle on each
  target, over the F4 transports. `exit` is each step's own status; every
  target's transport exited 0 and its clone was removed afterwards:

  | Target | `tools-test` | `workflow-lint` | `verify` | disclosure guard, `HEAD~5..HEAD` |
  | :--- | :--- | :--- | :--- | :--- |
  | Linux host | **exit=2** (4 s) | exit=0 (143 s) | exit=0 (260 s) | no global `pre-push` on the host |
  | WSL Ubuntu 24.04 | exit=0 (4 s) | exit=0 (12 s) | **exit=2** (33 s), at `tools-test` | no global hooks path |
  | Windows host | **exit=2** (5 s) | exit=0 (14 s) | exit=0 (56 s) | exit=0 (1 s) |
  | macOS | exit=0 (4 s) | exit=0 (11 s) | exit=0 (24 s) | exit=0 (1 s) |

  The Linux host's long `workflow-lint` and `verify` include the first
  `make tools` bootstrap in a fresh clone.
* **The script is `tools/test_tools.py`,** run by `make tools-test` and so
  by `make verify`, the pre-push hook and CI's `go` job. Each failure is
  the same test:

  ```text
  FAIL: test_log_holds_method_path_status_and_no_header (test_tools.FakeProviderTest...)
  AssertionError: Lists differ: ['GET[29 chars]1/chat/completions -> 429'] != ['GET[29 chars]1/chat/completions -> 429', 'POST /v1/chat/completions -> 200']
  ```

  On Windows it also raised, in `tearDown`:
  `OSError: [WinError 145] The directory is not empty: '...\Temp\tmp…'`.
* **Cause** (`tools/fakeprovider.py:106-115`, `tools/test_tools.py:29-40`):
  1. `Handler.answer` writes the response body, then calls `record`, which
     writes the log line. The client has its whole answer
     (`Content-Length`) and returns, so the test can read the log before
     the server thread writes the last line.
  2. `tearDown` removes the temporary directory before the cleanups
     registered with `addCleanup` stop the server, because unittest runs
     `tearDown` first. A late log write then lands in a directory being
     removed: on Windows, `WinError 145`.
* **Proven** on scratch copies of `tools/` only, each in a temporary
  directory (`make tools-test`'s command, run N times):

  | Variant | macOS | Linux host | Windows host |
  | :--- | :--- | :--- | :--- |
  | as committed, 30 runs | 4/30 failed | 7/30 failed | 9/30 failed |
  | as committed, `record` delayed 0.3 s, 5 runs | 5/5 failed | 5/5 failed | 5/5 failed |
  | candidate fix, 30 runs | 0/30 failed | 0/30 failed | 0/30 failed |
  | candidate fix, `record` delayed 0.3 s, 5 runs | 0/5 failed | 0/5 failed | 0/5 failed |

  The candidate fix logs before the response is sent (the server stops
  only after it is sent), and registers the temporary directory's removal
  as the first cleanup, so it runs last.
* **CI** passed both runs since the tests landed (`5981bc9`, `bdb82d6`),
  consistent with a 13–30% failure rate.
* **The disclosure guard is not the failing script.** It exited 0 over five
  real commits on macOS and on Windows.
* **Step 4:** the failure is in a `make verify` step, so it is a deviation:
  Deviation 1, below, on the owner's choice.

### Deviation 1 (2026-10-19): `tools/test_tools.py` races its fake provider

* **Found:** Phase 1's record above. The script the owner saw fail is the
  fake provider's test, which fails 13–30% of runs on every OS.
* **Decision (owner, "Fix under 0017"):** the fix goes under this PLAN as
  Phase 1a, before Phase 2, with a test that fails every time on the old
  code. 0016-PLAN, whose deliverable it is, gets a dated note pointing
  here, and is not reopened.
* **Added to scope:** `tools/fakeprovider.py`, `tools/test_tools.py`, and
  `0016-PLAN-keep-maintainer-tools-starting-with-a-fake-provider.md` (the
  note only).
* **Not changed:** what `fakeprovider.py` serves, logs, or how it stops; the
  log line is written earlier, and the server stops at the same point,
  after the last answer is sent.

### Phase 1a: fix the fake provider's race (2026-10-19)

* Phase 0's status changes were committed by the owner as `60f1de1`.
* **Steps 1–2:** `record` returns whether the request is the last;
  `answer` calls it before `send_response`, flushes the body, then starts
  the shutdown; `setUp` registers the temporary directory's removal first,
  and `tearDown` is gone; the new test
  `test_log_line_is_written_before_the_answer_arrives` was added. The diff
  is the candidate fix that Phase 1 proved, plus the new test and a comment
  citing this deviation.
* **Step 3, fail first:** the new test, with `HEAD`'s `fakeprovider.py`, on
  a scratch copy: exit 1.

  ```text
  ERROR: test_log_line_is_written_before_the_answer_arrives (test_tools.FakeProviderTest...)
  FileNotFoundError: [Errno 2] No such file or directory: '.../T/tmp…/requests.log'
  FAILED (errors=1)
  ```

* **Step 4, stability:** this phase's `tools/`, 30 runs each: macOS 0/30
  failed, the Linux host 0/30, the Windows host 0/30; each run `Ran 10
  tests`.
* **Step 5:** 0016-PLAN's Execution Record gained "Later finding
  (2026-10-19)", pointing here.
* **Step 6, the gate:** `make tools-test` exit 0 (`Ran 10 tests`, OK);
  `make verify` exit 0 (`0 issues.`, coverage 83.7%, `No vulnerabilities
  found.`, the tools' 10 tests); `make check-records` 0; `make
  markdownlint` 0; the identifier scan of the staged diff (the local
  account names, the host aliases and domains, the Windows host's address,
  home and temporary-directory paths): no hit; `py_compile` of both files:
  0. Staged for the owner's commit.

### Phase 2: `make preflight`, `make pre-add-check`, and CI (2026-10-19)

* Phase 1a was committed by the owner as `ce43f97`.
* **Steps 1–3,** as written: `preflight`, `pre-add-check` (with `FILES ?=`)
  and `verify-staged` as its alias in the `Makefile`, both new targets in
  `.PHONY`, and `##` help text on `check-records` and `markdownlint`; in
  `ci.yml`'s `go` job, "Check records and documentation links" and "Lint
  Markdown" before "Run repository quality contract". The help text says
  `FILES=...`, in ASCII; the `Makefile` carries no non-ASCII byte. `make
  help` lists all five targets.
* **Step 4, fail first,** each in its own scratch clone of `ce43f97`
  carrying this phase's `Makefile` and `ci.yml`, removed afterwards:

  | Plant | Command | Exit | Banners printed | The failure |
  | :--- | :--- | :--- | :--- | :--- |
  | none (control) | `make preflight` | 0 (25 s) | all three, then `preflight passed` | — |
  | a record link to `0017-MADR-no-such-record.md` | `make preflight` | 2 | `==> check-records` only | `broken relative link: docs/decisions/0017-PLAN-…:8: 0017-MADR-no-such-record.md` |
  | `#No space after the hash` in the guide | `make preflight` | 2 | `check-records`, `markdownlint` | `MD018/no-missing-space-atx` at `docs/guides/cicd-operations.md:247:1` |
  | a staged, unformatted `zz_unformatted.go` | `make preflight` | 2 | all three; no `preflight passed` | `gofmt found unformatted files: zz_unformatted.go`; `[fmt-check] Error 1` |
  | `no-such-key: true` on the "Lint Markdown" step | `make workflow-lint` | 2 | — | actionlint: `ci.yml:53:9: unexpected key "no-such-key" for step` |

* **Step 5:** `make -n pre-add-check FILES=main.go` printed
  `./scripts/bootstrap-tools.sh` then `python3 scripts/go-precheck.py
  main.go`; with nothing staged, `make pre-add-check` exit 0 (both pinned
  modules resolved from GitHub); `make -n verify-staged` ends with
  `python3 scripts/go-precheck.py`; `grep -qE '^pre-add-check:' Makefile`
  exit 0, so the fleet gate (F8) now finds it.
* **Step 6:** `make preflight` in the tree: exit 0; `==> check-records`,
  `==> markdownlint`, `==> verify`, `0 issues.`, coverage 83.7%, `No
  vulnerabilities found.`, the tools' 10 tests, `preflight passed`.
  Staged for the owner's commit, after the identifier scan below.
* **Not yet confirmed:** that the CI image's `python3` and `npx` run the
  two new steps. That is Phase 5 step 3, from the first run's log.
* **Asked after staging:** why CI needs `npx`. `make markdownlint` is
  `npx --yes markdownlint-cli2@0.23.2`; acp-go-sdk2, go-selfupdate-lib and
  go-tui-lib run the same command in CI on the runner image's Node, with no
  `setup-node`, while go-llmprovider-sdk runs only its records check. The
  owner kept the step as staged and committed the phase as `92611a4`.

### Phase 3: `make preflight-hosts` and `tools/hosts.py` (2026-10-19)

* **Steps 1–4, 6:** `tools/hostrun.py`, `tools/hosts.py`,
  `tools/hosts.example.json` (placeholders only), the `preflight-hosts`
  target with `HOSTS_FLAGS ?=` (in `.PHONY`, not a prerequisite of
  `preflight`), and `tools/README.md`'s two rows and its section, "The
  cross-host run". As written, plus:
  * `hostrun.py` sets its stdout to UTF-8, because a Windows pipe defaults
    to the ANSI code page, and exits 127 when the command cannot be
    started;
  * `hosts.py` rejects unknown config fields, an `ssh` or `wsl` on a
    `local` host, empty identifiers, and an unknown `--only` label, each
    with exit 2.
* **Step 5:** 15 tests in `HostsTest`: the five config errors, the
  default commands, the three exact transports, the dirty-tree refusal
  (no transport started), the redaction, a local host end to end
  (`describe=v0.0.1`, the committed file present), a failing host's tail
  and exit 1, the fake `ssh` (its recorded argv `["dest", "python3",
  "-"]`), an unreachable host (`exit=255`, exit 1), `--only`, and the
  clone's removal. The throwaway repository sets `core.hooksPath` to an
  empty path, so the global `prepare-commit-msg` hook never runs in a
  test. `make tools-test`: `Ran 25 tests`, OK; no `pcm-hosts-*` directory
  left in the temporary directory.
* **Step 7:** `make workflow-lint` exit 0; `py_compile` of both files: 0.
* **Step 8, fail first,** each on a scratch copy of `tools/`, removed
  afterwards; each named test also passed on the unplanted copy:

  | Plant | Test | Failure |
  | :--- | :--- | :--- |
  | `BatchMode=yes` dropped | `test_transports_are_exact_argument_lists` | `Lists differ: ['ssh', '-o', 'ConnectTimeout=5', '-T', 'dest', …` |
  | the dirty-tree check removed | `test_dirty_tree_exits_two_and_starts_no_transport` | `AssertionError: 1 != 2` |
  | the `/Users/` redaction removed | `test_redaction` | `'/Users/alice/x …' != '/Users/<user>/x …'` |
  | `hostrun`'s `finally` removal skipped | `test_clone_is_removed_afterwards` | `… was left behind`; the leaked directory was then removed |
  | the bundle made without `--tags` | `test_local_host_runs_on_a_clone_of_head_with_its_tags` | `1 != 0 : m: exit=9` |

* **Step 9, live.** The owner's config was written outside the repository,
  at `~/.config/prepare-commit-msg/hosts.json`: the four default hosts, and
  the account names, host aliases, host names and domain collected from the
  hosts as identifiers (never printed). The tree had this phase
  uncommitted, and the tool refuses a dirty tree, so each run was in a
  scratch clone of `92611a4` with this phase's files committed: the same
  content as this phase's commit.

  | Run | Exit | Hosts |
  | :--- | :--- | :--- |
  | `make preflight-hosts`, clean `HEAD` | 0 (48 s) | linux 0 (36 s), wsl-ubuntu-24.04 0 (47 s), windows 0 (3 s), macos 0 (4 s): `4/4 hosts passed` |
  | a committed `zz_windows_only_test.go` (`//go:build windows`, `t.Fatal`) | 1 (41 s) | linux 0, wsl-ubuntu-24.04 0, **windows 1**, macos 0: `3/4`; the Windows tail holds `planted: fails on Windows only` |
  | a dirty tree | 2 | `commit first: preflight-hosts tests HEAD, and the tree has changes` |
  | `--config /nonexistent/hosts.json` | 2 | `no config at /nonexistent/hosts.json: …` |

  The short Windows and macOS times are Go's test cache: the clone is new,
  but the packages and their inputs are unchanged. The Windows-only run
  shows the tests do run there when an input changes.
* **The scan of the live output and logs:** no unredacted home path. One
  identifier hit in each full run, traced: a three-letter account name
  inside the label `macos`, on the summary line. The label is the owner's
  config, printed as given; no host's log held an identifier.
* **The gate:** `make preflight` in the tree: exit 0 (`0 issues.`, coverage
  83.7%, `No vulnerabilities found.`, `Ran 25 tests`, `preflight passed`);
  `make check-records` 0; `make markdownlint` 0. The identifier scan of the
  staged diff's added lines, with the config's identifiers and home and
  temporary paths: six hits, each traced and none an identifier. Five are
  the three-letter account name inside "macOS", "macos" and "machine"; one
  is the made-up user `carol` in `test_redaction`'s fixtures. Staged for the
  owner's commit.

### Deviation 2 (2026-10-19): a record linked to a deleted script

* **Found:** at Phase 4 step 11, `make preflight` exit 2 at
  `check-records`: `broken relative link:
  docs/decisions/0005-MADR-windows-on-demand-compat-tests.md:76:
  ../../scripts/install-hooks.sh`, the only broken link. Step 2 deleted the
  script; step 10 said records keep their history and did not foresee a
  record linking to a deleted file.
* **Decision (owner, "Pinned GitHub link"):** the link points at the file
  as of `bdb82d6`, which is on `origin/main` and holds the file, with an
  inline note that 0017-MADR deleted it. The record's text is otherwise
  unchanged.
* **Added to scope:** `docs/decisions/0005-MADR-windows-on-demand-compat-tests.md`
  (that one link).

### Phase 4: remove the hook layer, then the docs (2026-10-19)

* Phase 3 was committed by the owner as `1724d22`. F1 was re-checked first:
  this clone had the managed `core.hooksPath`; the Windows host's clone has
  none.
* **Step 1:** `make hooks-uninstall` exit 0, "restored the previous local
  core.hooksPath state". Then `git config --local --get core.hooksPath`
  exit 1; `git rev-parse --path-format=absolute --git-path hooks` prints
  the global hooks directory; `.git/prepare-commit-msg-hooks` is gone.
* **Steps 2–4:** `git rm -r` of `.githooks`, `scripts/install-hooks.sh`,
  `scripts/uninstall-hooks.sh` and `scripts/test-hooks.sh`; the three
  targets and their `.PHONY` entries out of the `Makefile`;
  `scripts/verify-scripts.sh` finds files under `scripts` only.
* **Steps 5–8:**
  * `README.md`: the two hook rows replaced by `make preflight`, `make
    preflight-hosts` and `make pre-add-check`.
  * `docs/guides/cicd-operations.md`: "Routine verification" rewritten for
    `preflight`, `preflight-hosts` (when, and that it tests `HEAD`) and
    `pre-add-check`, plus "A clone that still has the old hooks". Its
    recovery is in separate blocks for the `no` and `yes` cases, so that
    copying it whole cannot run both. The rollback line has no hook step,
    and the opening cites 0017-MADR.
  * `docs/architecture.md`: the hooks paragraph and the tree.
  * `AGENTS.md`: `make pre-add-check` under "Pre-add checks", and a new
    "Before a push".
  * `docs/README.md`: two "I want to…" rows added, two extended.
* **Step 9:** a dated "Partially superseded" note in 0003-MADR's banner and
  at 0013-MADR's D8. No rationale was rewritten.
* **Step 10:** `git ls-files` of the four paths: nothing; `make -n
  hooks-install` exit 2, "No rule to make target `hooks-install'"; the
  search for the hook names and `.githooks` in the `Makefile`, `scripts/`,
  `tools/`, `.github/`, `README.md`, `AGENTS.md`, `docs/architecture.md` and
  `docs/guides/` finds only the recovery subsection
  (`docs/guides/cicd-operations.md:48-51`).
* **Step 11:** the first `make preflight` failed at `check-records`
  (Deviation 2). After that fix: exit 0 (`0 issues.`, coverage 83.7%, `No
  vulnerabilities found.`, `Ran 25 tests`, `preflight passed`). The pinned
  URL answers HTTP 200. The global hooks directory's `prepare-commit-msg`
  is executable; the owner's commit of this phase is the first to use it
  without the layer, and its message confirms V7.
* **The gate:** `make check-records` 0; `make markdownlint` 0. The
  identifier scan of the staged diff: two hits, both traced and neither an
  identifier: "machine", and the repository's GitHub owner in Deviation 2's
  URL, which `go.mod` and `ci.yml` already carry. Staged for the owner's
  commit.

### Phase 5: push, and the first CI run (2026-10-19)

* Phase 4 was committed by the owner as `fc399d7`. Its message,
  `chore(ci): replace Git hooks with Make preflight checks`, was written by
  the global `prepare-commit-msg` hook, with no local `core.hooksPath`
  (V7).
* **Step 1,** on the clean tree at `fc399d7`:
  * `make preflight`: exit 0 (`0 issues.`, coverage 83.7%, `No
    vulnerabilities found.`, `Ran 25 tests`, `preflight passed`);
  * `make preflight-hosts`, the first run on the tree itself: exit 0;
    `linux: exit=0 (38s)`, `wsl-ubuntu-24.04: exit=0 (31s)`, `windows:
    exit=0 (3s)`, `macos: exit=0 (3s)`, `4/4 hosts passed`;
  * the disclosure guard over `bdb82d6..fc399d7`: exit 0, no output.
* **Step 2:** the owner pushed the six commits, `d12b24d` to `fc399d7`.
* **Step 3,** CI run `39303119999` on `fc399d7`: `success`.
  * `Go (quality contract)`: "Check records and documentation links",
    "Lint Markdown" and "Run repository quality contract" all `success`.
    The log shows `python3 scripts/check_records.py --check-all`, then
    `npx --yes markdownlint-cli2@0.23.2` with `markdownlint-cli2 v0.23.2
    (markdownlint v0.41.1)`, `Linting: 8 files` and `Summary: 0 issues`,
    from the runner image's `python3` and Node. No setup step was added.
    `make verify`'s log holds `0 issues.`, coverage 83.7%, `No
    vulnerabilities found.`, and `Ran 25 tests`, so `HostsTest` runs on
    the CI runner too.
  * `Native Tests (Linux)`, `(macOS)`, `(Windows)`: `success`.
  * `Build release / build` and its five identity checks: `success`;
    `Publish GitHub Release`: skipped, as on any untagged push.
* **Step 4:** every Verification row is met, below; this PLAN is
  `complete`. MADR A1 was already `accepted` at Phase 0.

### Verification, as met

| # | Met | Evidence |
| :--- | :--- | :--- |
| V1 | yes | Phase 1: `tools/test_tools.py`'s log race, 13–30% of runs; fixed in Phase 1a (Deviation 1) |
| V2 | yes | Phase 2's fail-first table; `make preflight` exit 0 at each phase |
| V3 | yes | Phase 2 step 5 |
| V4 | yes | Phase 5 step 3: both steps `success`, from the image's `python3` and `npx` |
| V5 | yes | Phase 3 steps 5 and 8: 15 tests, five plants each failing its test |
| V6 | yes | Phase 3 step 9 and Phase 5 step 1 |
| V7 | yes | Phase 4 steps 1, 10, 11; `fc399d7`'s generated message |
| V8 | yes | Phase 4 steps 5–9; `make check-records`, `make markdownlint` 0 |
| V9 | yes | the identifier scan at each phase, every hit traced; the guard at Phase 5 step 1 |
| V10 | yes | CI run `39303119999` |

**Not done, and why:**

* No CI change beyond the `go` job's two steps, as scoped.
* The `available-dev-hosts` skill was not updated with F12 (WSL Ubuntu
  24.04's SSH server). It is outside this repository; it was offered to the
  owner and not asked for.
