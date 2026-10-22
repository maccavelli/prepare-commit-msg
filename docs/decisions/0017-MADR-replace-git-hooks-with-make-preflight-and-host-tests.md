---
status: accepted
date: 2026-10-19
decision-makers: Project Owner
consulted: <cli-b> and magic-cli-remote `make preflight`; go-tui-lib 0018-MADR-python-verification-tools.md (its `remote` runner); the owner's dev-host fleet notes
informed: maintainers and agents working in this repository
---
# Replace the repository Git hooks with `make preflight` and a cross-host test run

## Context and Problem Statement

Every claim below is marked **[read]** (from this repository at `bdb82d6`,
or the record or sibling repository named), **[probed]** (run read-only in
this investigation) or **[reported]** (the owner's words, not reproduced).

The owner, on 2026-10-19:

> i do not like the pre-commit/push git hooks in place for this. one python
> script fails. it is running too much before the push can proceed. i want
> this repo to behave from a cicd perspective similar to how the other newer
> repos work, as far as testing across the test hosts, make pre-flight, etc.

**What runs on a push today.** `make hooks-install` sets this clone's local
`core.hooksPath` to a managed directory under `.git/`
(`scripts/install-hooks.sh`) **[read]**. Its `pre-push` runs the previous
global `pre-push` (the GitHub disclosure guard, `github-disclosure.py`), then
`.githooks/pre-push`, which is `exec make -C "$REPO_ROOT" verify` **[read]**.
Every other executable in the global hooks directory is carried forward as a
bash passthrough, including one named `github-disclosure.py` that is a bash
script **[read]**. The repository pre-commit hook is already gone
(0013-MADR-align-repository-with-workspace-scaffolding.md D8) **[read]**.

**`make verify` is the whole CI contract, serially** (`Makefile:114`)
**[read]**: `tools`, `mod-check`, `fmt-check`, `lint`, `vet`,
`test` (`go test -race -v ./...`), `coverage` (a second full test run),
`vuln` (needs the vulnerability database over the network), `workflow-lint`,
`tools-test`, and `build-all` (six cross-compiled binaries). Every `git push`,
including a docs-only push, waits on all of it.

**What CI runs** (`.github/workflows/ci.yml`) **[read]**:

| Job | Runner | Command |
| --- | --- | --- |
| `go` | ubuntu-24.04 | `make verify` (`ci.yml:49`) |
| `go-native` | ubuntu-24.04, macos-15, windows-2025 | `go test ./...` (`ci.yml:81`) |
| `build` | reusable go-selfupdate-lib workflow | the release rehearsal |

So the pre-push hook reproduces the `go` job on the developer's OS (macOS
here), not on Linux where CI runs it, and reproduces none of `go-native`'s
Linux or Windows runs. A Windows-only failure still reaches CI first.

**What CI does not run.** `make check-records` and `make markdownlint` are
AGENTS.md's pre-add checks, but no CI job runs them **[read]**. `make
hooks-test` is in neither `verify` nor CI **[read]**.

**The failing Python script is not identified.** The owner reports one
fails **[reported]**. In this investigation `make tools-test` (9 tests) and
`python3 scripts/check_records.py --check-all` both exit 0, and the
disclosure guard exits 0 over the current, empty outgoing range **[probed]**.
The Python the push path runs is the disclosure guard, `tools/` tests, the
Python `ast.parse` in `scripts/verify-scripts.sh`, and Python-based linters
inside `make verify`; which one fails, and on which host, is open (More
Information).

**How the newer repositories work** **[read]**:

* **<cli-b>** and **magic-cli-remote** set no local `core.hooksPath` and
  carry no `.githooks/`. Each has `make preflight`: every CI gate in one
  target, each announced with `==> <gate>`, stopping at the first failure,
  ending `preflight passed`. Each has `make pre-add-check FILES=…`, which the
  fleet's agent commit gate calls when `scripts/go-precheck.sh` is absent.
* **go-tui-lib** runs the same gates on its other hosts through a Python
  tool, `python3 -m tools.verify remote`
  (0018-MADR-python-verification-tools.md): it copies the tree and a git
  bundle of `HEAD` and the tags to the host over `scp`, runs the gates there
  through `ssh -o BatchMode=yes` with argument lists only, and redacts home
  paths and configured identifiers from what comes back. The hosts are named
  only in a config file outside the repository; a placeholder example is
  committed.
* <cli-b>'s 0007-PLAN ran `make preflight` inside a WSL distro the same
  way, by hand.

**The fleet's test hosts** cover every CI runner's platform: a Linux
workstation, a Windows workstation (Git for Windows shell), four WSL distros
on it (one is Ubuntu 24.04, CI's Linux image), and the macOS machine this
session runs on. All are reached by SSH key with `BatchMode=yes`; none may be
named in this repository (AGENTS.md, "Identifiers") **[read]**.

**Which raises the question:** should verification stay a Git hook, and if
not, what replaces it so that a push is known to pass CI before it is made?

## Decision Drivers

* **A push is fast and never blocked by a local gate.** The owner's
  complaint, and the newer repositories' behaviour.
* **Green locally means green in CI.** Whatever replaces the hook must cover
  every CI job that can run off GitHub, on the platform CI runs it on.
* **One command per question.** "Is this tree CI-clean?" and "does it pass
  on every OS?" each get one `make` target, named as in the sibling
  repositories.
* **The global hooks are not this repository's.** The commit-message hook
  and the disclosure guard come from the owner's global `core.hooksPath` and
  keep running; AGENTS.md forbids bypassing the guard.
* **Nothing committed names a host, account or real path.**
* **Repository tools are Python 3 on the standard library, each tested**
  (0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md).
* **No existing clone is left broken.** Deleting `.githooks/pre-push` under
  an installed wrapper would fail every push from that clone.

## Considered Options

* **A. Keep the hooks as they are**
* **B. Keep a repository pre-push hook, slimmed to the fast gates**
* **C. Remove the repository hook layer; add `make preflight` and `make preflight-hosts`**
* **D. Remove the repository hook layer and rely on CI alone**

## Decision Outcome

The owner accepted this decision on 2026-10-19 ("accepted"), with the four
answers recorded under More Information.

Chosen option: "C", because it is what <cli-b>, magic-cli-remote and
go-tui-lib do, it makes `git push` return as fast as the disclosure guard
does, and — unlike today's hook — it runs CI's `go` job on Linux and its
`go-native` job on Linux, macOS and Windows before the push, which is the
owner's "ensure it will pass once the push hits the cicd".

### D1. Remove the repository hook layer, uninstall first

Each clone that ran `make hooks-install` runs `make hooks-uninstall` first,
which restores its previous `core.hooksPath` state and deletes the managed
directory. Only then are `.githooks/`, `scripts/install-hooks.sh`,
`scripts/uninstall-hooks.sh`, `scripts/test-hooks.sh` and the
`hooks-install`, `hooks-uninstall` and `hooks-test` targets deleted, and
`scripts/verify-scripts.sh` stops scanning `.githooks/`.

The order is the point: the managed `pre-push` calls
`.githooks/pre-push` by absolute path under `set -euo pipefail`, so
deleting it first fails every push from that clone. A clone found with the
managed directory after the scripts are gone is restored by hand:
`git config --local --unset core.hooksPath` (or the value in its
`.previous-local-value`), then removing `.git/prepare-commit-msg-hooks`. The
operations guide carries that recovery.

The global `prepare-commit-msg` and `pre-push` hooks are untouched; with no
local `core.hooksPath`, Git runs them directly.

### D2. `make preflight`: every CI gate that runs off GitHub, in one target

`make preflight` runs, in order, each announced as `==> <gate>` and stopping
at the first failure:

1. `make check-records` and `make markdownlint` — seconds, so first;
2. `make verify` — CI's `go` job's quality contract, unchanged.

It ends with `preflight passed`.

CI's `go` job gains two steps, `make check-records` and `make markdownlint`,
before `make verify` *(the owner's choice, 2026-10-19)*. The ubuntu-24.04
runner image carries `python3` and `npx`; the PLAN confirms both on the
first run rather than assuming them. So `make preflight` and the `go` job
run the same three targets in the same order, and neither can drift from
the other while `make verify` stays the single definition of the Go gates.

### D3. `make pre-add-check`: the fleet name for the staged check

`make pre-add-check FILES=…` runs `scripts/go-precheck.py $(FILES)`, the
same contract as <cli-b>'s and magic-cli-remote's, so the fleet's agent
commit gate finds it. `make verify-staged` stays as an alias, because
earlier records cite it.

### D4. `make preflight-hosts`: CI's jobs on the platforms CI runs them on

A new maintainer tool, `tools/hosts.py`, run by `make preflight-hosts`,
modelled on go-tui-lib's `tools.verify remote` and smaller:

1. **What it tests is `HEAD`.** It refuses a dirty tree, because CI tests
   the pushed commit. It sends a git bundle of `HEAD` and the tags (the
   `Makefile` reads `git describe`, and CI checks out with full history) and
   clones it into a fresh temporary directory on each host, which it removes
   afterwards.
2. **Each host runs its CI job's command.** The config gives each host a
   kind and a command; the defaults are:

   | Kind | Command | Mirrors |
   | --- | --- | --- |
   | `linux` (the Linux workstation, the WSL Ubuntu 24.04 distro) | `make verify` | the `go` job, and `go-native` on Linux |
   | `windows` (Git for Windows shell) | `go test ./...` | `go-native` on Windows |
   | `local` (macOS, run in place, no SSH) | `go test ./...` | `go-native` on macOS |

   The default list is CI's platforms only *(the owner's choice,
   2026-10-19)*. The other three WSL distros are not CI runners; a
   maintainer may add any of them to the local config with `go test ./...`,
   and the example config shows how, but no default run includes them.
3. **Transport.** `ssh` and `scp` with `-o BatchMode=yes -o
   ConnectTimeout=5`, argument lists only, never `shell=True`; a WSL distro
   is reached through the Windows host's connection with `wsl -d <distro> --
   bash -lc`. A host that refuses batch authentication is a failure, never a
   password prompt.
4. **Output.** Hosts run in parallel; each one's full output goes to a file
   under the system temporary directory, and the tool prints one line per
   host, `<label>: exit=N (Ns)`, then the failing hosts' last lines. Home
   paths and the config's identifiers are redacted from everything it
   prints. It exits non-zero if any host fails or cannot be reached.
5. **Config, never committed.**
   `$XDG_CONFIG_HOME/prepare-commit-msg/hosts.json` (default
   `~/.config/prepare-commit-msg/hosts.json`), or `--config PATH`, names
   each host's label, kind, SSH destination and command. With no config the
   tool says where the file goes and exits 2. `tools/hosts.example.json`
   shows the shape with placeholders only.
6. **Tests.** `tools/test_tools.py` covers it against a fake `ssh` and `scp`
   that record their argument lists, a throwaway repository, a dirty-tree
   refusal and the redaction, so `make tools-test` (and so `make verify`)
   runs them.

`make preflight-hosts` is by hand, not a gate inside `preflight`: it needs
the hosts to be up and reachable, which a laptop away from them is not.

### D5. The push procedure

Before asking the owner to push, an agent runs, in order: `make preflight`,
always; `make preflight-hosts` when the change can behave differently per OS
(Go sources, `go.mod`/`go.sum`, build tags, paths, `os/exec`, permissions,
`Makefile`, scripts, workflows) — a change to docs and records only is
exempt; and AGENTS.md's disclosure-guard command *(scoped, the owner's
choice, 2026-10-19)*. A host it could not reach is said in the handoff, never passed over
in silence. AGENTS.md, `docs/guides/cicd-operations.md`,
`docs/architecture.md`, the README's `make` table and `docs/README.md` are
updated to say so and to drop the hook instructions.

### D6. What this supersedes

* The pre-push part of
  [0003-MADR-layer-and-harden-ci-cd-quality-gates.md](0003-MADR-layer-and-harden-ci-cd-quality-gates.md),
  "Local hooks and developer workflow". Its CI, release and
  repository-settings decisions stay in force.
* [0013-MADR-align-repository-with-workspace-scaffolding.md](0013-MADR-align-repository-with-workspace-scaffolding.md)
  D8's retention of `.githooks/pre-push` and the hook installer. D8's
  retirement of the pre-commit hook stands.
* It is the "new trigger decision" 0013 D8 requires for a future
  Windows-only local suite
  ([0005-MADR-windows-on-demand-compat-tests.md](0005-MADR-windows-on-demand-compat-tests.md)):
  Windows runs by hand through `make preflight-hosts`, not from a hook.

Both earlier records get a superseded-in-part banner pointing here; their
rationale is not rewritten.

### Consequences

* Good, because `git push` runs only the global disclosure guard, and a
  docs-only push no longer waits on a full test, lint, vulnerability and
  six-target build cycle.
* Good, because CI's `go` job is reproduced on Linux and `go-native` on all
  three OSes before the push, which the hook never did.
* Good, because `check-records` and `markdownlint` failures are caught by
  preflight locally and by CI on every push, where today nothing but a
  person's memory runs them.
* Good, because the repository matches its siblings: the same target names,
  no local `core.hooksPath`, and the fleet agent gate finds
  `make pre-add-check`.
* Good, because 411 lines of hook, installer, uninstaller and their test
  leave the repository.
* Neutral, because whichever Python script fails today is not fixed by this
  record if it is the global disclosure guard, which stays; it is diagnosed
  in the PLAN either way.
* Bad, because nothing now stops a push that skipped `make preflight`; CI
  is the backstop, and the push procedure (D5) is a rule, not a mechanism.
* Bad, because `tools/hosts.py` is new code to own, and each host must keep
  the CI toolchain (Go from `go.mod`, the linters `make tools` installs,
  `python3`, `npx` where `make verify` needs it); toolchains drift and must
  be re-probed.
* Bad, because every clone with the hooks installed needs one
  `make hooks-uninstall` before it pulls the deletion, or a manual recovery.

### Confirmation

* On this clone and every other clone with the hooks installed, after
  `make hooks-uninstall`: `git config --local --get core.hooksPath` exits 1
  and `git rev-parse --path-format=absolute --git-path hooks` names the
  global hooks directory; a commit still gets a generated message.
* `git ls-files .githooks scripts/install-hooks.sh scripts/uninstall-hooks.sh
  scripts/test-hooks.sh` prints nothing, and `make -n hooks-install` fails.
* `make preflight` exits 0 on a clean tree and stops at the gate named on a
  deliberately broken copy (an unformatted Go file; a record with a broken
  link).
* `make preflight-hosts` reports `exit=0` for every configured host on a
  clean `HEAD`; on a scratch branch with a Windows-only failing test it
  reports that host non-zero and the others zero; with a dirty tree it
  refuses; with no config it exits 2.
* `make tools-test` runs the new tests, and each was seen to fail on a
  broken copy of `tools/hosts.py`.
* No committed file names a host, account or real path; the disclosure
  guard passes over the outgoing commits.
* The first CI run after the change is green on `go` (with its new
  `check-records` and `markdownlint` steps) and all three `go-native`
  runners, and `actionlint` passes on the edited workflow.

## Pros and Cons of the Options

### A. Keep the hooks as they are

* Good, because nothing changes and an unverified tree cannot be pushed from
  a clone that installed the hooks.
* Bad, because every push waits on all of `make verify`, which the owner
  rejects.
* Bad, because it verifies on the developer's OS only, so Linux- and
  Windows-only failures still reach CI first.
* Bad, because only clones that ran `make hooks-install` are gated, so it
  guarantees less than it appears to.
* Bad, because the repository keeps a hook-composition layer its siblings
  do not have.

### B. Keep a slimmed repository pre-push hook

* Good, because some gate remains automatic.
* Neutral, because pushes get faster, but by dropping exactly the gates
  (tests, vulnerabilities, cross-builds) most likely to fail CI.
* Bad, because the installer, uninstaller and their test all stay, and the
  owner asked for the hooks to go.
* Bad, because it still does nothing for other OSes.

### C. No repository hooks; `make preflight` and `make preflight-hosts`

* Good, because it is the siblings' model, with go-tui-lib's proven
  cross-host transport.
* Good, because it covers CI's jobs on CI's platforms, which no hook here
  ever did.
* Good, because push latency is the disclosure guard's alone.
* Bad, because the checks run only when someone runs them.
* Bad, because it adds a tool and its config to maintain.

### D. No repository hooks; CI alone

* Good, because it is the least code: delete the hooks and stop.
* Bad, because every failure is found after the push, on GitHub, where it
  is slowest to diagnose and, for a tag, can fail a release.
* Bad, because it ignores the owner's explicit ask for `make preflight` and
  cross-host testing.

## More Information

* **Resolved, 2026-10-19:** the failing script is `tools/test_tools.py`.
  Its fake provider logs a request after sending the answer, so the test
  can read the log too early (13–30% of runs, on every OS). It is a gate
  inside `make verify`, so it is fixed under the PLAN, as Deviation 1
  ([0017-PLAN-replace-git-hooks-with-make-preflight-and-host-tests.md](0017-PLAN-replace-git-hooks-with-make-preflight-and-host-tests.md),
  Phase 1). No decision here changes.
* **The owner's answers, 2026-10-19,** each the recommended option:
  * *Which Python script fails:* reproduce it first. The PLAN's first phase
    runs the managed `pre-push` by hand with a ref line on stdin, in a
    scratch clone, on the host where it failed, before anything is deleted.
    If it is the global disclosure guard, the fix belongs in the dotfiles
    repository, and this record's D5 still runs it. If it is a gate inside
    `make verify`, it fails `make preflight` and CI too, and is fixed under
    the PLAN as a dated deviation.
  * *CI runs `check-records` and `markdownlint`:* yes (D2).
  * *Default hosts:* CI's platforms only (D4).
  * *Push procedure:* `make preflight` always, `make preflight-hosts` when
    the change is OS-sensitive (D5).
* Sibling evidence: <cli-b> `Makefile` (`preflight`, `pre-add-check`);
  magic-cli-remote `Makefile` (`preflight`, "a green preflight means a green
  CI"); go-tui-lib `docs/decisions/0018-MADR-python-verification-tools.md`
  and `tools/verify/remote.py`.
* The implementation is
  [0017-PLAN-replace-git-hooks-with-make-preflight-and-host-tests.md](0017-PLAN-replace-git-hooks-with-make-preflight-and-host-tests.md).

## Amendments

### A1 (2026-10-19): the bundle travels on stdin; no `scp`

*Status: accepted (2026-10-19), with the PLAN* (the owner's "approved to
proceed"). Found while grounding the PLAN; it changes D4 items 1 and 3,
and nothing else.

**Found** **[probed]**, 2026-10-19, read-only on every host:

* A 1,866,976-byte Python program, sent on stdin to `python3 -`, arrived
  with a matching SHA-256 on all four targets: over `ssh -o BatchMode=yes
  -o ConnectTimeout=5 -T` to the Linux host and to the Windows host (whose
  `ssh` lands in Git for Windows' bash, `os.name` `nt`), through the
  Windows host to `wsl -d Ubuntu-24.04 -- bash -lc 'python3 -'`, and
  locally. A bundle of `HEAD --tags` at `bdb82d6` is 1,349,091 bytes.
* `git clone` of that bundle checks out `HEAD` (detached, the same commit)
  with all 24 tags, `git describe --tags --always --dirty` works, and the
  clone is clean.
* Every host has `python3` 3.14, Go 1.27.2 (what
  `scripts/bootstrap-tools.sh` requires), `git`, `make`, and on Linux
  `gcc` with `CGO_ENABLED=1`, which `go test -race` needs.
* *Corrected 2026-10-19, at the owner's word, then probed:* WSL Ubuntu
  24.04 runs its own OpenSSH server, socket-activated by systemd
  (`ssh.socket`) on port 2222. It is not reachable on that port from the
  macOS machine directly (the connection timed out; the Windows host has
  no `portproxy` rule), but it is through the Windows host as a jump host
  (`-J <windows> -p 2222 <user>@localhost`), which reaches it and is
  offered its ED25519 host key. There, this machine's key is refused
  (`Permission denied (publickey,password)`). The other three distros
  have no SSH listener of their own.

**Decided.** The tool sends one program on stdin: the remote runner's
source with the base64 bundle, the command and the label appended. There
is no `scp` and no upload path: one channel, proven on every target, in
place of two, and no remote path to choose or clean up. ~~A WSL distro has
no SSH server~~ *(false, corrected above: Ubuntu 24.04 has one)*.

WSL is reached through the Windows host's own connection, `wsl -d <distro>
-- bash -lc 'python3 -'`, and not through the distro's SSH server *(the
owner's choice, 2026-10-19)*, because:

* it works today, proven, with no change to any host or to `~/.ssh`;
  reaching the distro's server would need this machine's key added to its
  `authorized_keys` and its host key pinned;
* it is the fleet's rule for nested environments: one SSH connection per
  physical machine, with WSL entered through it;
* it works for any distro, where only Ubuntu 24.04 has a server.

The cost is a `wsl` field on a `linux` host, and one more argument list
in the tool. The runner, which needs Python 3.12 or
later, writes the bundle to a temporary directory, clones it, runs the
command, and removes the directory. The `local` kind runs the same program
through `python3 -`, so the tests exercise the code the hosts run. D4's
other items stand.
