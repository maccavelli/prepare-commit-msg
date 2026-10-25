# Architecture

How this repository is put together, as it is now. This file carries no
history and no rationale. The decision records under
[decisions/](decisions/) explain why the system has this shape, and
[README.md](README.md) points to the records and guides people use most.

## What it is

`prepare-commit-msg` is a Go CLI and Git `prepare-commit-msg` hook. During a
normal commit it reads the staged diff, asks a configured local or remote LLM
provider for a Conventional Commit message, and writes that message into Git's
commit-message file. The same binary exposes `configure`, `update`,
`version`, and `identity` commands.

## Runtime flow

`main.go` dispatches CLI subcommands or hook execution. The hook skips Git
sources with an already supplied or generated message, collects staged changes
through `internal/git`, loads user settings through `internal/config`, and
calls providers through `go-llmprovider-sdk`. It tries the configured primary
model followed by at most three fallbacks, retries eligible transient failures,
and soft-fails so Git leaves the commit editor available when analysis or
generation cannot complete.
The list of providers, whether each needs a key, and each one's default
endpoint come from the SDK's provider registry (`providers.Default()`),
from which `internal/config` builds its provider list. A provider's saved
`base_url`, when present, replaces that default endpoint. Configuration is
stored in the platform configuration directory; it supports API keys, saved
OAuth sessions kept separately from `config.json`, and vendor-CLI credentials
read from their existing auth files. `internal/fsutil` owns safe file writes.
The update command delegates release selection, integrity checks, replacement,
and build identity to
`go-selfupdate-lib`.

Interactive configuration lives in `internal/ui`. The root
[README](../README.md) explains installation and user behavior. The
[CI/CD operations guide](guides/cicd-operations.md) gives maintainers the
verification, repository-settings, and release procedures.

## Quality and delivery boundary

The `Makefile` is the shared local and CI command surface. `make verify`
checks modules, formatting, lint, vet, race-enabled tests, coverage,
vulnerabilities, workflow syntax, and all release cross-builds.
`make verify-staged` remains an explicit staged-snapshot diagnostic.

The repository installs no Git hooks. Before a push, `make preflight` runs
`check-records`, `markdownlint` and `verify`, and `make preflight-hosts` runs
CI's jobs for `HEAD` on the test hosts through `tools/hosts.py`
([0017-MADR](decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md)).
[The CI workflow](../.github/workflows/ci.yml) runs those three targets, the
`disclosure-check` step beside them, and the native platform jobs. A change
to Markdown that no code reads starts [the Docs workflow](../.github/workflows/docs.yml)
instead, with the same records, Markdown and disclosure gates, while every
tag and every other change starts the full workflow
([0022-MADR](decisions/0022-MADR-align-docs-ci-and-disclosure-gates.md)).

`selfupdate-release.json` defines release assets and build identity. The CI
workflow calls pinned reusable build and publish workflows from
`go-selfupdate-lib`.

## Tree

```text
README.md                  product entry and documentation link
AGENTS.md                  shared agent workflow
Makefile                   local and CI command surface
main.go                    CLI dispatch, hook execution, and the commit-message write
update.go                  update command and release-asset selection
internal/                  git, config, UI, and atomic file writes
scripts/                   verification and repository utilities
tools/                     maintainer tools, run by hand: tools/README.md
.github/workflows/ci.yml   CI, native tests, and release orchestration
.github/workflows/docs.yml  Markdown-only checks
selfupdate-release.json    release asset and identity specification
docs/
  README.md                documentation index and task matrix
  architecture.md          this current-system description
  decisions/               MADR/PLAN pairs
  reports/                 numbered verification evidence
  guides/                  task-oriented operations guides
```

## What is not here

Provider implementations and provider-specific model discovery live in
`go-llmprovider-sdk`. Self-update and reusable release-workflow internals live
in `go-selfupdate-lib`. Host identity, global Git-hook configuration,
credentials, and live GitHub repository settings remain outside the repository.
