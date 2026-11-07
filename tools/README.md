# Maintainer tools

The tools a maintainer, or an agent, runs by hand while working in this
repository, as
[0016-MADR](../docs/decisions/0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md)
decides. What `make` and CI run as gates is in [`scripts/`](../scripts/).

Each tool is Python 3 on the standard library, starts with
`#!/usr/bin/env python3`, and finds the repository from its own file, so it
runs from any working directory. None names a host, an account or a machine
path; anything outside the repository is an argument.

Exit codes, for every tool: **0** clean, **1** a finding, **2** a tool error.

| Tool | What it does | Usage |
| :--- | :--- | :--- |
| `fakeprovider.py` | a local OpenAI-compatible API on `127.0.0.1` that answers the first generations with an error and `Retry-After`, then a commit message; it logs each request's method, path and status, and never a header | `tools/fakeprovider.py [--fail N] [--status CODE] [--retry-after S] [--message TEXT] [--port-file FILE] [--log FILE] [--max-requests N]` |
| `hosts.py` | runs CI's jobs for `HEAD` on the test hosts, in parallel, and reports one line per host ([0017-MADR](../docs/decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md) D4) | `make preflight-hosts [HOSTS_FLAGS='--only LABEL --config PATH']` |
| `hostrun.py` | what each host runs: a clone of `HEAD` from a bundle, the command, then cleanup. `hosts.py` sends its source to the host's `python3 -`; it is not run by hand | — |
| `prune_releases.py` | keeps the newest stable GitHub releases and deletes older releases and tags, after a dry run ([0026-MADR](../docs/decisions/0026-MADR-keep-five-stable-releases.md)) | `python3 tools/prune_releases.py [--repo OWNER/REPO] [--keep N] [--apply]` |

## The cross-host run

`make preflight-hosts` tests the commit, not the working tree: it refuses a
dirty tree. It bundles `HEAD` and the tags, and sends each host one program
on stdin over `ssh -o BatchMode=yes -o ConnectTimeout=5 -T`. A WSL distro is
entered through its Windows host's connection, with `wsl -d <distro> --
bash -lc 'python3 -'`. Each host needs Git, Python 3.12 or later as
`python3`, and the toolchain its command needs.

The hosts are named only in a config file that is never committed:
`--config PATH`, else `$XDG_CONFIG_HOME/prepare-commit-msg/hosts.json`, else
`~/.config/prepare-commit-msg/hosts.json`. Copy
[`hosts.example.json`](hosts.example.json) there and fill in the SSH
destinations. Each host has a `label`, a `kind` and, unless it is `local`,
an `ssh` destination; a `linux` host may add `wsl`, a distro name. The
`command` defaults to the kind's CI job:

| `kind` | Default `command` | CI job |
| :--- | :--- | :--- |
| `linux` | `make verify` | `go`, and `go-native` on Linux |
| `windows` | `go test ./...` | `go-native` on Windows |
| `local` | `go test ./...` | `go-native` on this machine's OS |

To also test a WSL distro that CI does not run, add it with its own label
and `go test ./...`:

```json
{"label": "wsl-almalinux-10", "kind": "linux", "ssh": "<windows-ssh-destination>",
 "wsl": "AlmaLinux-10", "command": ["go", "test", "./..."]}
```

`identifiers` lists strings, such as an account or host name, that are
replaced with `<redacted>` in everything the tool prints or logs; home paths
become `<user>` either way. Each host's full output is in
`<log dir>/<label>.log`, and the last line names the log directory.

## The live retry check

A release that changes retries, the provider path or the wizard's
endpoints checks the hook's retry line against the fake, on a scratch build
with a scratch `HOME` (0015-PLAN, Phase 5 step 3). OpenAI takes no endpoint
in this app, so the check configures `together`:

```bash
tools/fakeprovider.py --fail 1 --port-file /tmp/pcm-fake.port --log /tmp/pcm-fake.log &
TOGETHER_API_KEY=made-up-key ./prepare-commit-msg configure --yes --provider together \
  --base-url "http://127.0.0.1:$(cat /tmp/pcm-fake.port)/v1" --model zai-org/GLM-5.3-Flash
```

Then run the hook in a scratch repository with a staged file. Its stderr
holds one `retrying in` line, and the message is the fake's.

## Pruning releases

`prune_releases.py` keeps the eight newest stable releases, or `--keep N`.
A stable release is not a draft, not a prerelease, and is tagged
`vMAJOR.MINOR.PATCH`. It deletes:

- each older stable release, then its tag;
- each prerelease at or below the newest stable release, then its tag;
- each `v*` tag with no release below the oldest kept release.

It reports, and never deletes, drafts, newer prereleases, newer tags with no
release, and any tag or release whose tag is not a semantic `v*` version.
None of these counts toward the window.

A run prints the plan and deletes nothing; it exits 1 when there is
something to delete and 0 when there is not. `--apply` deletes, each release
before its tag, as immutable releases require. It then reads the releases
and tags back and checks them against the plan. It refuses, deleting
nothing, when fewer stable releases exist than `--keep`, when the plan would
delete the latest release, or when a listing fails. It stops at the first
failed deletion and lists what it deleted; running `--apply` again finishes
the work.

The repository is this clone's `origin` unless `--repo` names one. `gh` must
be logged in as an account that can delete releases and tags. A deleted
release cannot be restored, and its tag name cannot be used again.

## Tests

`make tools-test` runs [`test_tools.py`](test_tools.py), and `make verify`,
which CI runs, runs it. Each test fails on a planted break and passes clean.

## `make tools` is something else

`make tools` installs the pinned linters under `.tools/bin`; it is not this
directory. It still runs with `tools/` present only because `tools` is in the
Makefile's `.PHONY`: keep it there, or Make takes the directory for the
target and reports "`tools' is up to date".
