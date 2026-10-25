---
status: "proposed"
date: 2026-10-25
associated-madr: "0023-MADR-split-user-documentation-into-guides.md"
---
# Implement the split of user documentation into `docs/guides/`

Associated MADR: [0023-MADR-split-user-documentation-into-guides.md](0023-MADR-split-user-documentation-into-guides.md)

## Goal

The root `README.md` is a short product entry page whose ToC and "I want to…"
matrix link task-shaped guides. End-user documentation lives in
`docs/guides/` as guides; maintainer release content sits with the CI/CD
runbook; both READMEs carry ToC and "I want to…" sections and link the
guides. Every fact the two Go doc-reader tests pin today is still pinned, at
the content's new location.

## Scope

Affected files: `docs/guides/installation.md`, `docs/guides/configuration.md`,
`docs/guides/providers-and-models.md`, `docs/guides/hook-behavior.md`,
`docs/guides/self-update.md`, `docs/guides/release-notes.md`,
`docs/guides/developing.md` (all new), `README.md`, `docs/README.md`,
`docs/architecture.md`, `docs/guides/cicd-operations.md`, `readme_test.go`
and `documentation_test.go`.

Out of scope: `.github/workflows/`, the `Makefile`, product code, the
uncommitted and untracked
[0022-MADR-align-docs-ci-and-disclosure-gates.md](0022-MADR-align-docs-ci-and-disclosure-gates.md)
pair and its `docs/README.md` index rows, plus any execution of that pair.
Those 0022 files stay unstaged throughout; do not delete, amend or commit
them.

Boundary: no release-note sentence may drift in meaning when it moves. The
string `Release v1.10.1 moves to Go 1.27.3` must exist in
`docs/guides/release-notes.md` after Phase 3, and the two exact installer
commands in `docs/guides/installation.md`, because
[documentation_test.go](../../documentation_test.go) and
[readme_test.go](../../readme_test.go) are amended in that phase to read
those paths.

## Implementation Steps

### Phase 1: write the user guides

1. Create `docs/guides/installation.md` from root README "Quick Start &
   Installation" (installer, pre-built binaries, build from source, global
   and per-repository hook installation, Windows). Keep the two installer
   commands verbatim.
2. Create `docs/guides/configuration.md` from "Configuration" (Interactive
   Wizard, Non-Interactive CLI Flags, Configuration File, ChatGPT plan terms)
   and "Environment Variables Reference" except the credential table (that
   goes to the providers guide), keeping the precedence list and both env-var
   tables' provider rows consistent with `internal/config` and the SDK.
3. Create `docs/guides/providers-and-models.md` from "Supported Providers &
   Models", plus the per-provider credential environment variables. Keep the
   provider table's column layout so the amended regex in `readme_test.go`
   keeps matching it.
4. Create `docs/guides/hook-behavior.md` from "Daily Workflow & Hook
   Behavior" and "How It Works & Design" (flow, trigger matrix, stderr and
   retry line formats).
5. Create `docs/guides/self-update.md` from "Self-Update" (flags table, exit
   statuses, backup and replacement behavior, security note).
6. Create `docs/guides/release-notes.md` from the seven "Changes with …" and
   "Changes from the mcplib Releases" sections, in the order the README had
   them; keep every outcome sentence as it is.
7. Create `docs/guides/developing.md` from "Developer Experience" ("Toolchain
   & Quality Gates" and "Build & Test Targets"); verify every cited target
   against the `Makefile` in place.
8. Move the "Maintainer Release" release-identity, checksum and attestation
   verification commands into `docs/guides/cicd-operations.md` ("Controlled
   release"), and update that guide's sentence "The identity check and the
   attestation verify command stay in the root README" to name the
   operations guide instead. Verify the workflow-references citations the
   `documentation_test.go` fixture requires remain in this guide.
9. Prove each of the seven new guides fails the planned link resolver on one
   deliberately planted bad relative link before the resolver runs for real:
   work in a scratch copy under `/tmp`, never in the tree, and record the
   planted input and the failure.
10. Commit this phase only after the owner has approved execution of this
    plan; run the pre-add checks in the plan's Verification section first.

### Phase 2: retitle the entry points

1. Rewrite the root `README.md` as the product entry page: identity and "Why
   This Exists", a ToC for the page, and the "I want to…" matrix whose rows
   link the seven guides, `docs/architecture.md`, the records and
   `AGENTS.md`; keep the License note. Drop the sections that became guides.
2. Update `docs/README.md`'s upper block to list each new guide with a
   one-line description, gobble-cli style, and expand the "I want to…"
   matrix with user-task rows (install the hook, configure a provider and
   its credentials, see supported providers and models, know what the hook
   does to a commit, update the binary, see what changed in a release,
   build and test locally). Leave the hand-maintained Records table, its
   rows and the owner's 0022 rows exactly as they are; the index stays
   hand-edited — `check_records.py --write-index` is not used here.
3. Update `docs/architecture.md`: the Tree section's README and guides
   descriptions and the pointer that the root README explains installation,
   which now points to the guides.
4. Run the link resolver for real: every link in `README.md`,
   `docs/README.md`, `docs/architecture.md`, `docs/guides/*.md` and
   `tools/README.md` must resolve from its own file, and the pre-move root
   README anchors (for example `#quick-start--installation`) must now be
   dead as confirmed by the resolver's output.
5. Commit this phase only after the pre-add checks pass.

### Phase 3: follow the doc readers, verify, close

1. Amend `readme_test.go` to parse `docs/guides/providers-and-models.md` and
   amend `documentation_test.go`'s fixture paths and assertions:
   `docs/guides/installation.md` for the installer commands,
   `docs/guides/release-notes.md` for the v1.10.1 release string, and the
   operations guide for the workflow references.
2. Prove the amended readers fail on deliberately broken scratch-tree copies
   (a wrong installer command, a missing provider row, a wrong curated
   order) and pass on the real tree, mirroring the existing
   `TestCurrentDocumentationDetectsBrokenFixtures` pattern where practical;
   record the commands and observed failures.
3. From a clean tree run: `make check-records`, `make markdownlint`,
   `go test ./...`, then `make preflight`. Capture each status before
   filtering output.
4. Read each edited file in full and check links by hand where the resolver
   has no opinion (anchor text). Run the identifiers check: no hostname,
   account name, org-internal path or real-machine absolute path is added;
   placeholders (`<user>`, `/home/<user>/...`) are required. Check the two
   new pair files and every edited document.
5. Commit the Go test amendments and any final index rows for this pair in
   the same or a final commit, updating the Records table's 0023 STATUS cells
   as this plan's phases complete.
6. Mark this PLAN complete only when three conditions hold: the amended
   readers pass on the real tree and fail on the planted copies; `make
   preflight` passes; the link resolver reports zero broken links. Send to
   the owner the summary of which phases ran, what was verified and what was
   not (nothing here observes a push; `make preflight-hosts` is not expected
   to be affected by docs and test-reader changes, and is run only if the
   review asks).

## Verification

* `make check-records` and `make markdownlint` pass on the tree.
* `go test ./...` passes; the amended `TestReadmeCuratedModels` demands the
  SDK catalog in `docs/guides/providers-and-models.md`, and
  `TestCurrentDocumentation` demands the installer commands and release
  string from their guide locations.
* The link resolver reads every link in the two READMEs, `architecture.md`,
  the seven new guides, `cicd-operations.md` and `tools/README.md`, and
  exposes planted breakage and dead pre-move anchors first.
* `make preflight` passes before each commit of a phase.
* No 0022 file is staged, amended or deleted at any point.

## Rollout and Rollback

Documentation and Go test changes only; no CI, workflow or product change, so
rollout is the phase commits themselves. Rollback is a Git revert of the
split commits, which restores README, docs readers and guides; the docs-only
nature of the change keeps `make preflight-hosts` out of the critical path.
This plan does not authorize a push, a tag or a release.
