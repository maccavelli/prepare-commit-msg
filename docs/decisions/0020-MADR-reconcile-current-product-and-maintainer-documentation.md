---
status: accepted
date: 2026-10-23
decision-makers: Project Owner
consulted: repository working tree at 81b9d4c; published v1.10.1 release installers; README, architecture note, documentation index, CI/CD operations guide, Makefile, CI workflow, Go module metadata, and focused source and test review
informed: users of the installed hook and repository maintainers
---
# Reconcile the current product and maintainer documentation with the released implementation

## Context and Problem Statement

Recent changes upgraded the Go toolchain and both libraries, replaced
repository hooks with Make targets, added cross-host preflight support, and
published `v1.10.1`. The documentation must distinguish the current product
behavior from historical MADR and PLAN evidence, which remains immutable as a
record of what was known and decided at the time.

This review compared the current working tree at `81b9d4c` with the public
release and recorded the following evidence.

* The implementation is a Go CLI and `prepare-commit-msg` hook. It gathers a
  staged diff, honors the configured byte limit, asks the selected provider
  and up to three fallbacks, soft-fails without blocking the commit, and
  atomically writes the generated message while preserving Git's existing
  comments (`main.go`, `internal/git/`, and `internal/fsutil/`). `configure`,
  `update`, `version`, and `identity` are separate commands.
* `go.mod` names Go `1.27.3`, `go-llmprovider-sdk v1.21.0`, and
  `go-selfupdate-lib v1.17.2`. The Makefile and CI workflow implement the
  documented quality and release boundary. `make check-records` passed on
  2026-10-23.
* The published Unix installer fetched by the README's `latest/download`
  URL identifies itself as the `v1.10.1` installer. Its documented `--dir`
  option selects the target directory, verifies every downloaded file against
  `SHA256SUMS`, and invokes `prepare-commit-msg identity` before retaining
  the binary. An end-to-end run with an isolated temporary target directory
  exited zero and installed a binary whose identity was
  `v1.10.1 (release) 52ad5494ccfe`.
* The PowerShell installer at the matching published URL was inspected but
  not executed on this macOS host. Its parameter block accepts the README's
  `-InstallDir` and `-NoPathUpdate` arguments, and its source performs the
  corresponding checksum and identity checks. Windows execution remains
  required before asserting a cross-platform live-install result.

The review also found current-documentation drift:

* At the start of this review, `docs/README.md` had no rows for the accepted
  `0019-MADR-skip-full-ci-for-docs-only-pushes.md` or its implementation
  plan, contrary to the repository rule that it indexes every record.
* `docs/guides/cicd-operations.md` names reusable build and publish workflow
  commit `5e199c831b5691ea687943e3c3fd495d50c739ed`, while
  `.github/workflows/ci.yml` pins both to
  `036289ecc701d84c7749f4950a57a154c1af2cf4` (`v1.17.2`).
* The README's section headed “Changes with Go 1.27.3,
  go-llmprovider-sdk v1.21.0 and go-selfupdate-lib v1.17.2” calls the change
  “The next release,” although the `v1.10.1` release was published and its
  verification is recorded in
  `0018-PLAN-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md`.

## Decision Drivers

* Current user and maintainer instructions must be demonstrably true.
* Historical decision records must retain their evidence and chronology.
* Installation instructions need an executable, release-backed verification.
* Documentation changes should be reviewable and testable without changing
  product behavior.

## Considered Options

* Correct only the explicitly stale statements.
* Reconcile all current-facing documentation and add durable verification for
  the release install commands.
* Rewrite historical decision records to make every old statement current.

## Decision Outcome

Chosen option: "Reconcile all current-facing documentation and add durable verification for the release install commands," because it corrects the identified drift, protects the highest-risk user path, and preserves historical records as evidence rather than recasting them as current documentation.

### Consequences

* Good, because the root README, docs index, architecture note, and operations
  guide will have clear, distinct current responsibilities.
* Good, because a future release or installer-template change can fail a
  focused check before inaccurate installation guidance is merged.
* Neutral, because a Windows-capable environment is still needed for a live
  PowerShell installation test.
* Bad, because current documentation must be reviewed whenever release
  workflow pins, installer interfaces, or user-visible behavior change.

### Confirmation

The follow-on implementation must:

* update the root README's released-version wording and retain the verified
  Unix and PowerShell one-line commands;
* reconcile `docs/README.md`, `docs/architecture.md`, and
  `docs/guides/cicd-operations.md` with the current source and CI workflow;
* add or extend focused tests that detect drift in the installer commands and
  current release-workflow facts, including a deliberately broken fixture
  that demonstrates each new assertion fails; and
* run `make check-records`, Markdown linting, the focused tests, and a
  non-destructive Unix installer smoke test in an isolated directory. Run the
  PowerShell smoke test on Windows before claiming it has passed.

## Pros and Cons of the Options

### Correct only the explicitly stale statements

* Good, because the immediate wording errors are small.
* Bad, because it does not establish ownership or checks for closely related
  documentation and installation claims.

### Reconcile all current-facing documentation and add durable verification for the release install commands

* Good, because it covers the public entry point and the maintainer runbook
  in one evidence-led update.
* Good, because testable facts are less likely to drift after later releases.
* Bad, because it requires a scoped implementation plan and a Windows check
  for complete installer coverage.

### Rewrite historical decision records to make every old statement current

* Good, because a superficial search would show fewer old values.
* Bad, because it destroys the records' historical context and conflicts with
  the repository rule not to silently rewrite rationale.

## More Information

This is a review record only. No user-facing product documentation was
corrected in this change. The required `docs/README.md` index rows for this
MADR and the previously omitted 0019 pair are included as record-maintenance
changes. A later approved PLAN should define the exact documentation edits,
test fixtures, and Windows-host verification.

*Accepted 2026-10-23 by the Project Owner. The associated implementation plan
is `0020-PLAN-reconcile-current-product-and-maintainer-documentation.md`.
