---
status: in-progress
date: 2026-10-25
associated-madr: "0022-MADR-align-docs-ci-and-disclosure-gates.md"
---
# Implement aligned docs-only CI and disclosure gates

Associated MADR: [0022-MADR-align-docs-ci-and-disclosure-gates.md](0022-MADR-align-docs-ci-and-disclosure-gates.md)

## Goal

A documentation-only push or pull request gets a Docs run with the common Markdown, record/link and portable disclosure gates plus this repository's Markdown-reading checks. Other changes retain full CI. The private pre-push guard remains part of publication.

## Scope

Affected files: `.github/workflows/ci.yml`, `.github/workflows/docs.yml` (new), `Makefile`, `scripts/check_ci_filters.py` (new), `scripts/check_disclosure.py` (new), their tests, `scripts/verify-scripts.sh`, `scripts/configure-github.sh`, `AGENTS.md`, `docs/architecture.md`, `docs/guides/cicd-operations.md`, `docs/README.md`.

Repository-specific work: Implement the prior decision's narrow docs set, excluding root `README.md` and every Markdown file that a test or tool reads unless its exact reader check runs in Docs. Add `docs.yml` with `make check-records`, `make markdownlint` and the portable tracked-text disclosure target; add that disclosure target to the full Go job. Extend the prior filter checker design to compare push/PR lists, docs inclusion, and tracked Markdown readers; run it through `make workflow-lint`. Keep `workflow_dispatch` and tags unfiltered. Update the operations guide, `AGENTS.md` and `docs/architecture.md`.

Boundary: The earlier PLAN's Phase 1 scope is expanded here for common disclosure and parity; this new PLAN supersedes its unexecuted phase only after owner approval. Do not make `README.md` docs-only while `readme_test.go` reads it.

## Implementation Steps

### Phase 1: inventory and shared checks

1. Re-read workflow triggers, job commands, Markdownlint globs, tracked Markdown files and every code or script reader of tracked Markdown. Record the docs-safe path set in this PLAN before editing filters. Include hidden agent-rule trees and test fixtures. Review required-status settings; if a filtered job would stay pending, amend this pair with a concrete migration before changing filters.
2. Add missing record/link and Markdownlint targets, reusing the current commands where they exist. Run identical pinned documentation commands in full CI and Docs. Keep Go lint, tests, vulnerability, builds and releases in full CI.
3. Implement a standard-library portable disclosure check over tracked text. It rejects real macOS, Linux and Windows user-home paths, accepts documented placeholders, prints only file, line and category, and fails on unreadable tracked input. Run the same target in full CI and Docs. Retain the private deny list and global pre-push guard locally.
4. Add fixture tests for new gates. Prove each fails on deliberately broken input in a scratch copy and passes on clean input. Record commands, statuses and redacted diagnostics. Commit the approved phase after checks pass.

### Phase 2: workflow routing and parity

1. Keep push and pull-request CI ignore sets identical and limited to audited docs-safe paths. Make Docs paths cover every ignored path and the Docs workflow itself. Mixed changes run full CI. Preserve tags, schedules, dispatch, release jobs and permissions.
2. Add or extend a Python workflow contract check. It validates actual YAML structure, compares event filters, confirms ignored tracked paths are in Docs, refuses known Markdown inputs to code tests unless their reader runs in Docs, and compares common docs-gate commands and pins. Integrate it in the workflow/tool test target.
3. Run actionlint and the contract check. In independent scratch copies, plant a mismatched path, a Markdown reader wrongly ignored, a missing Docs gate and a differing pin; require each to fail. Commit the approved phase after checks pass.

### Phase 3: documentation and observed rollout

1. Explain routing, docs-safe paths, common checks, local readers and the private pre-push guard in AGENTS.md and the architecture or CI guide. Verify this pair's exact filenames and current statuses in the records index; use the repository index generator if present.
2. Run `make check-records markdownlint workflow-lint disclosure-check`; `make preflight`; actionlint; scratch plants for MD018, broken link, real home path, root README incorrectly ignored, a Markdown reader omitted from Docs, and mismatched filters. Capture status before filtering output. Check this pair and edited documents for real-machine identifiers. Read each edited file in full and verify links.
3. After explicit authorization to push in that same turn, observe a docs-only push, a non-doc or mixed push and the next tag or scheduled run. Record run IDs, conclusions and which workflow started. Mark this PLAN complete only after the criteria below hold; otherwise record the outstanding observation. Commit at the end of each approved phase under repository rules.

## Verification

* Both workflows pass pinned Markdownlint, numbered-record/index and relative-link validation, and portable tracked-text disclosure scanning.
* Each Markdown reader runs on changes to its input, either in Docs or full CI.
* A docs-only push starts Docs but not full CI; a non-doc or mixed push starts full CI. Tags, schedules and dispatch retain existing behavior.
* Scratch plants fail their intended gates; filter or gate drift fails the parity test.
* Before an authorized push, the local pre-push guard checks outgoing commits. Its private deny list is never copied into CI.

## Rollout and Rollback

Land gates before relying on the filters. If routing misclassifies a change, revert the routing commit through a new reviewed commit so full CI runs for every push again. Keep the shared gates while repairing the audit. This PLAN does not authorize a push or tag.

## Execution Record

### Approval and supersession (2026-10-25)

The owner said "start on the 0022 plan first" at this tree, after committing
this pair and the 0023 pair. That approval starts the plan's phases; per the
boundary above, the owner's approval also supersedes
[0019-PLAN-skip-full-ci-for-docs-only-pushes.md](0019-PLAN-skip-full-ci-for-docs-only-pushes.md)'s
unexecuted Phase 1: workflows and filters execute through this plan only, and
0019's remaining document and closing steps fold into this plan's Phase 3.
0019-MADR stays accepted; its PLAN is marked superseded in the records index,
and its file is left untouched elsewhere.

The 0023 pair (split user documentation into `docs/guides/`) is approved next,
not now: its phases stay unstarted until the owner says so, and its reader
amendments to `readme_test.go` and `documentation_test.go` will interact with
this plan's docs-safe set at that time.

### Phase 1 inventory (re-read 2026-10-25 at 0754b73, read-only)

| # | Fact | How |
| :--- | :--- | :--- |
| I1 | `ci.yml` triggers are push (`main`, tags `v*`), pull request and `workflow_dispatch` (`ci.yml:3-9`); the `go` job runs `make check-records`, `make markdownlint` and `make verify` (`ci.yml:48-55`); native jobs run `go test ./...` on Linux, macOS and Windows; build and publish call the two pinned `go-selfupdate-lib` workflows. Checkout is `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.0` (`ci.yml:35`, `:76`). | read |
| I2 | The documentation commands exist and are `make check-records` → `python3 scripts/check_records.py --check-all` and `make markdownlint` → `npx --yes markdownlint-cli2@0.23.2` (`Makefile:136-140`). Step 2's "missing" targets: none. | read |
| I3 | The code's Markdown readers are `readme_test.go` (root `README.md`, `readme_test.go:26`) and `documentation_test.go` (root `README.md`, `docs/guides/cicd-operations.md`, `.github/workflows/ci.yml`, `documentation_test.go:95`). `update.go` embeds only `selfupdate-release.json`; no `testdata/` file is Markdown; `tools/*.py` name records in docstrings only. `scripts/check_records.py` is a documentation gate that runs in both workflows, not a content reader. | `git grep` over `*.go`, `tools/*.py`, `scripts/*.py` |
| I4 | 58 tracked Markdown files. Docs-safe: every one except `README.md` (I3, `readme_test.go`) and `docs/guides/cicd-operations.md` (I3, `documentation_test.go`, a Go test that runs only inside `make verify`). | `git ls-files '*.md'` |
| I5 | Required-status review: the rulesets `prepare-commit-msg-main` and `prepare-commit-msg-release-tags` are active with no required status checks, and `main` has no branch protection (HTTP 404). A filtered job cannot stay pending; no settings migration is needed. | `gh api` on rulesets and `branches/main/protection` |
| I6 | `.markdownlint-cli2.jsonc` lints all Markdown except every MADR and PLAN record. | read |
| I7 | The repository holds no tracked binary; `.tools/` is untracked and `testdata/`'s fixtures are text (`git ls-files`). | read |

### Docs-safe path set (recorded before any filter edit, per Phase 1 step 1)

A broad `docs/**.md` ignore cannot be used: it would swallow
`docs/guides/cicd-operations.md`, which I3 shows is a code-read Markdown. The
audited set enumerates safe subtrees, each matching tracked
files, none matching the two code-read paths:

```text
docs/README.md
docs/architecture.md
docs/decisions/*.md
docs/reports/*.md
AGENTS.md
tools/README.md
.claude/rules/*.md
.codex/rules/*.md
.grok/rules/*.md
.kilo/agent/*.md
.opencode/rules.md
```

Docs' `paths` covers every listed path generically with `'**.md'`, plus
`'.markdownlint-cli2.jsonc'` and `'.github/workflows/docs.yml'`. When the 0023
plan adds guides whose exact readers run only in full CI, their paths join the
code-read exclusions of this set in that pair's own amendment.

### Deviation 1 (2026-10-25, resolved the same day): the generic checker meets the redaction tooling's own text

Phase 1's checker, implemented and run against the tree as written, reports
six findings that are leakage-shaped but deliberate, and none is a
documented placeholder:

* `tools/hosts.py` lines 64-66 — the `HOMES` rewriting table 0017-MADR D4
  defines, whose subject is replacing real home paths; its regex fragments
  are necessarily concrete-shaped.
* `tools/test_tools.py` lines 273-274 — `hosts.redact`'s own test literals,
  which must show a concrete input beside the redacted result to prove the
  substitution.
* `docs/decisions/0017-PLAN-replace-git-hooks-with-make-preflight-and-host-tests.md`
  line 647 — recorded historical evidence quoting the planted test's
  assertion output; a record's evidence is not rewritten.

The phase's own checks caught three implementation slips first, all fixed in
this phase before staging: a fixture test written into the wrong test class,
a checker placeholder rule that markdown markup reaching into a segment
defeated, and a manual unittest run leaving `scripts/__pycache__` for
`verify-scripts.sh`'s scripts loop to sweep — the loop now carries the
`tools/` loop's `__pycache__` exclusion. An exit status was once captured
through a pipeline, a rule break caught in review; later statuses are
captured before any filter.

Resolutions offered, none of them a silent loosening:

* **R-A: file-level exemptions with reasons in the checker.** Precise and
  durable, but silent: future edits inside the three files, including this
  deviation's own additions, skip the check.
* **R-B (recommended): accept a line that itself demonstrates redaction** —
  a line carrying a full `<…>` token is about redaction, not leakage. All
  six findings pass mechanically; ordinary lines keep full coverage; the
  residual risk, a real home path coinciding on a line that also carries a
  `<…>` token, is covered by the private pre-push guard that owns the real
  names.
* **R-C: change the tree.** Rewriting the record's evidence falsifies it and
  rewriting the tooling's test input defeats the test; not viable.
* Doing nothing: the shared disclosure gate cannot pass on this tree, so the
  phase cannot commit, and every CI run fails `disclosure-check` once the
  gate lands.

**Chosen: R-B, the owner's choice.** Implemented in
`scripts/check_disclosure.py`: a line carrying a full `<…>` token
demonstrates redaction and is accepted whole; a wrapped literal splits such a
demonstration across lines, so the flagged line is accepted when the token
sits on the line beside it too. The residual risk, a real home path within a
line of a `<…>` token, is named in the checker's docstring and covered by the
private pre-push guard, which owns the real names. Proven by fixture tests
including the adjacent-line case, the boundary one line past the token, and
this plan's P-D plants: a planted real-shaped macOS home path fails with
file, line and category; a planted line demonstrating redaction passes; and
a planted unreadable file exits 2.

### Phase 2 executed (2026-10-25)

Routing and parity landed, after the owner's standing rule of 2026-10-25
that agent tooling and multi-part evaluation are modular Python and that
native whole-file tools replace text-processing one-liners — recorded in
AGENTS.md the same day and followed from here.

* `ci.yml` gained one `paths-ignore` list under `on.push` (which keeps
  `branches` and `tags`) and one under `on.pull_request`, both the audited
  docs-safe set; `docs.yml` gained the `paths` lists `['**.md',
  '.markdownlint-cli2.jsonc', '.github/workflows/docs.yml']` on both events.
* `scripts/check_ci_filters.py` (new) implements the extended checker: strict
  flow-list shapes, exact agreement within each workflow, per-pattern tracked
  coverage, docs-coverage of ignored paths, the Markdown-readers refusal with
  its "unless Docs runs the reader" allowance, doc-gate and pin parity, and
the tags/workflow_dispatch guard. `make workflow-lint` runs it after
actionlint, and its 21 fixture tests run in `make tools-test`.
* `configure-github.sh`'s workflow-tracking comment now names `docs.yml` and
  the 0022 MADR.
* The scratch clones (one per plant, the phase's uncommitted files applied)
  each behaved as specified: the clean clone exits 0; the mismatched-path,
  README-ignored, missing-Docs-gate and differing-pin plants each exit 1 with
  the expected problem line. Fixture testing caught three real checker
  defects first, all fixed before staging: the `on:`/`jobs:` section boundary
  collapsed at the first two-space key, the flow lists anchor at event-child
  depth, and the embed scanner's regex could never match off byte zero, so
  `update.go`'s embed was silently unscanned.

