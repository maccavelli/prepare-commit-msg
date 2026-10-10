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

The root command dispatches CLI subcommands or hook execution. Hook execution
collects staged changes through `internal/git`, loads user settings through
`internal/config`, and calls providers through `go-llmprovider-sdk`.
`internal/fsutil` owns safe file writes. The update command delegates release
selection, integrity checks, replacement, and build identity to
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

The hook installer preserves previously effective host hooks. The repository
adds a pre-push wrapper that runs `make verify`; it does not add a repository
pre-commit build hook. [The CI workflow](../.github/workflows/ci.yml) runs the
same complete verification contract and native platform jobs.

`selfupdate-release.json` defines release assets and build identity. The CI
workflow calls pinned reusable build and publish workflows from
`go-selfupdate-lib`.

## Tree

```text
README.md                  product entry and documentation link
AGENTS.md                  shared agent workflow
Makefile                   local and CI command surface
cmd/                       CLI command implementations
internal/                  Git, config, UI, filesystem, and hook behavior
scripts/                   verification, hook, and repository utilities
.githooks/pre-push         full local verification before push
.github/workflows/ci.yml   CI, native tests, and release orchestration
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
