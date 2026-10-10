---
status: proposed
date: 2026-10-10
associated-madr: "0013-MADR-align-repository-with-workspace-scaffolding.md"
---

# Implement the fleet workspace scaffold and documentation-tree normalization

Associated MADR:
[0013-MADR-align-repository-with-workspace-scaffolding.md](0013-MADR-align-repository-with-workspace-scaffolding.md)

## Goal

Bring this existing Go product into complete alignment with the fleet
workspace scaffold while preserving its product behavior, product gates,
record numbers, historical rationale, and unrelated work.

Done means:

* all five supported coding harnesses enter through repository-local pointers
  and share `AGENTS.md` as the normative workflow;
* every canonical byte-copy file matches its source template;
* the repository has working records, Markdown, identifier, and LF gates that
  have each been seen to fail on the prescribed scratch plant;
* all numbered records are in `docs/decisions/` or `docs/reports/` without
  renumbering;
* the operations runbook is in `docs/guides/`;
* the root README, docs index, architecture document, guide, records, and gate
  report form a navigable, link-clean tree;
* the existing Go and repository verification contract still passes;
* the execution record contains command exit statuses, plant failures,
  template comparisons, deviations, and anything deliberately left undone.

This PLAN is proposed for review. It authorizes no implementation, staging,
commit, push, tag, installer, build, or live-service mutation until the owner
explicitly approves execution or an individual phase.

## Research Baseline

### Repository facts

The plan is based on the repository at `3a1874f`, which commits the proposed
MADR. At authoring time this PLAN is untracked, and
`0012-PLAN-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md`
has a separate unstaged execution-log update. That update is outside this
PLAN and must be preserved. Before execution, re-run the baseline probes
rather than assuming this state is still current.

* The repository is an existing Go product with a `Makefile`, CI, local hook
  wrappers, release tooling, and a hand-maintained docs index.
* The active hooks path is a repository wrapper. Its recorded previous path
  matches the global hooks setting, and its `pre-commit`, `pre-push`, and
  `prepare-commit-msg` wrappers are present. This supports
  `git commit --no-edit`.
* The current product checks are `make verify` for the complete contract and
  `make verify-staged` for the staged Go/module snapshot.
* The documentation has one subject tree. There is no adjacent stack with its
  own records.
* The repository-wide sequence uses `0001` through `0013`; after this pair is
  present, the next unused number is `0014`.
* A complete scratch copy with the canonical records checker passes existing
  relative links and reports eight placement warnings.
* Moving the eight misplaced records and the operations runbook in a scratch
  copy, without repairing links, produces exactly 23 broken relative links.
  The deterministic repair set is recorded in Phase 2.

### Canonical template facts

The canonical source is
`~/.agents/skills/workspace-scaffolding/templates/`. These SHA-256 values pin
the templates assessed when this PLAN was written:

| Template | SHA-256 |
| :--- | :--- |
| `.claude.gitignore` | `ced771be89a8b13b6517f27cdcf0c8aadc2e16863cf7be34100eba45572ed1e5` |
| `pointers/claude.md` | `12c6efea28ca74d16937f37b258b15d6a27a9e1603f8efcb80ad80434e15a0e6` |
| `pointers/grok.md` | `8c795bcc0258e1f793098252adb6c144a090ff5fa3ab9036735fea996c4d1f26` |
| `pointers/opencode.md` | `da47769ae5321e3fa9f19a1fb363dd419e43801244a754a2183e0dee8993cc33` |
| `opencode.json` | `40a650f6e9160e3f1f04c2adbfb51cf3bcb25285e178b9bb5c1f36fd082a012c` |
| `pointers/codex.md` | `9d324af31baf398466912122bcd22b0067a375832310adbd268c1118dbc4d946` |
| `pointers/kilo.md` | `cc7a0de8b9d238682574cd8a6f361eadc78ad921a1a15553f1e178aaa54085cd` |
| `.kilo.gitignore` | `1e8b8ec1ffd844689e6e1a3bd1b06488f9a806843bd6010ae2f3bfdfd7cc03f5` |
| `.markdownlint-cli2.jsonc` | `3c7bf774fdad72e57d230a6c7954ce0384453c076b66cf40496099277a25bbc9` |
| `scripts/check_records.py` | `2f04a1d2fb5bc25616b5743573ed851aaf8f914196ef1c63d8c31cad22e9d928` |
| `AGENTS.md` | `83281fe22aeccab90e246d9943d3218038b5c7e7f52cde3175800cd29f7544a2` |
| `docs/README.md` | `2d3a1be9d22f1016dacc99257a7c3998b2eff2ae6492dc937169f3a3a0917f4a` |
| `docs/architecture.md` | `718f962a8762cc0f1c45a1d7da12f362dbb2876d511f53fcd012cc8ce844f737` |
| `Makefile.snippet` | `9347ab66e306c96ba7e29c07a0a34da32c604c907372289092deb1f9656890e5` |
| `.gitattributes` | `d9442b00a1c4e7a1282ee6e77947cc409503385ba3eac95229e2a43e290816d6` |
| `gitignore-fragment` | `ada0482c733c373d9f5ee5b6675593dbc7422e773faefa8b130d0ec5f4b0f32a` |

Recompute these hashes before execution. A changed byte-copy template, a
changed workflow rule, or a changed target path is a plan deviation: inspect
the change, amend the MADR or PLAN as needed, present the amendment, and wait
for approval. Do not silently execute a newer scaffold than the one reviewed.

## Scope

### In scope

* Accepting and indexing this MADR/PLAN pair after owner approval.
* Merging the canonical LF and ignore policies into `.gitattributes` and
  `.gitignore`.
* Adding the canonical records checker and Markdown configuration.
* Adding the canonical portable `SHELL`, `check-records`, and `markdownlint`
  Makefile content without altering the Go gates.
* Moving the `0001` through `0005` records that are outside
  `docs/decisions/`, with all required link repairs.
* Moving `docs/cicd-operations.md` to
  `docs/guides/cicd-operations.md` and repairing live references.
* Adding `docs/architecture.md` and
  `docs/reports/0013-GATES-workspace-scaffolding.md`.
* Reworking `docs/README.md` into the required records table and “I want
  to…” matrix while retaining useful status and topic information.
* Adding a root README documentation entry without replacing product prose.
* Updating the live record path in the comment in
  `scripts/configure-github.sh`; no shell behavior changes.
* Adding `AGENTS.md`, the five harness pointers, two harness-local ignore
  files, and `opencode.json`.
* Running and recording the prescribed scratch plants and final gates.
* Phase-end commits with `git commit --no-edit` after required checks pass,
  when execution of that phase has been approved.

### Out of scope

* Go source, tests, module requirements, dependency versions, linter versions,
  release specifications, CI workflow behavior, GitHub settings, tags,
  releases, and live-service changes.
* Changing repository hook behavior or reinstalling hooks.
* Changing the canonical user skill or any host-level identity, hook, MCP, or
  application configuration.
* Adding `.editorconfig`, Dependabot, `SECURITY.md`, `CODEOWNERS`,
  `CONTRIBUTING.md`, `CHANGELOG.md`, a new CI workflow, application-specific
  Claude settings, or extra Kilo configuration.
* Renumbering records, rewriting accepted decisions, or modernizing old record
  prose beyond path repairs required by the moves.
* Pushes and tags. They require a separate explicit ask in the same turn.

## Rules for Every Phase

1. Capture `git status --short --branch` before editing. Preserve every path
   not named by the active phase. If a named path has changed since this PLAN
   was reviewed, stop and present the overlap before writing.
2. Re-read the whole file after every in-place edit and assert the intended
   replacement landed. Do not infer success from a patch command alone.
3. Keep canonical byte-copy files byte-identical. Use exact, asserted slot
   replacement for generated files. Do not hand-edit a copied pointer or gate.
4. Redirect long check output to a scratch file, capture the command's exit
   status before filtering or summarizing it, and preserve the relevant output
   in the PLAN or GATES record.
5. Run new-gate failure plants only in a scratch copy or scratch repository.
   Never plant a failure in the working tree.
6. Do not weaken a rule, exclude a failing document, skip a test, discard an
   error, or change scope to make a check pass.
7. If an unexpected failure or new required file is outside this scope, stop.
   Record the evidence and resolution in a dated PLAN deviation; amend the
   MADR if a decision or asserted fact changes; obtain approval before
   continuing.
8. Before each phase commit, inspect `git diff` and `git diff --cached`, stage
   only the phase paths, run that phase's pre-add checks, and confirm the hook
   path is still the global directory or its chaining repository wrapper.
9. Commit with `git commit --no-edit`. Never pass `-m`, `-F`, `-c`, `-C`, or
   `--amend`; never bypass hooks. Record the resulting commit in this PLAN.
10. Do not push or tag.

## Implementation Steps

### Phase 0: accept and commit the reviewed records

**Entry condition:** the owner explicitly accepts
`0013-MADR-align-repository-with-workspace-scaffolding.md` and approves this
PLAN or Phase 0.

**Files:**

* `docs/decisions/0013-MADR-align-repository-with-workspace-scaffolding.md`
* `docs/decisions/0013-PLAN-align-repository-with-workspace-scaffolding.md`
* `docs/README.md`

1. Re-run the repository and template baseline probes. Confirm:
   * the only decision number newly claimed is `0013`;
   * `3a1874f` or its descendant contains the proposed MADR unchanged except
     for separately reviewed amendments;
   * the MADR and PLAN filenames and slugs are identical apart from the kind;
   * this PLAN links to the MADR and the MADR names this same-slug PLAN;
   * the current hook wrapper still chains to the configured global hooks
     directory;
   * no named phase path contains an unreviewed concurrent edit.
2. Change the MADR status from `proposed` to `accepted`, with the actual
   approval date. Change this PLAN from `proposed` to `in-progress`, with the
   same date. Replace the MADR's inline PLAN filename in `More Information`
   with a relative Markdown link to this PLAN, then confirm the pair links
   both ways.
3. Add the `0013` MADR and PLAN to the current hand-maintained sections of
   `docs/README.md`. Use their full filenames and actual statuses. Do not
   restructure the index yet; Phase 2 performs that atomic migration.
4. Build a complete scratch copy excluding `.git`, ignored build outputs, and
   tool caches. Install the canonical records checker in that scratch copy and
   run:

   ```bash
   python3 scripts/check_records.py --next
   python3 scripts/check_records.py --check-all
   ```

   Require `0014` from `--next`, exit 0 from `--check-all`, no broken links,
   and only the eight already-known placement warnings.
5. Read all three edited files in full. Confirm metadata, links, statuses, and
   the final newline.
6. Stage only the three files, inspect the staged diff, run the scratch record
   check once more against the staged snapshot, and commit with
   `git commit --no-edit`.
7. Add an execution entry to this phase before staging. It must contain the
   baseline commit, check exit statuses, warning count, and `commit pending`.
   Record the resulting commit at the start of Phase 1 and in the final
   handoff; do not amend Phase 0 merely to insert its own commit ID.

**Phase acceptance:** the accepted MADR, in-progress PLAN, and temporary index
entries are committed together; the working tree contains no uncommitted
Phase 0 changes.

### Phase 1: merge LF and ignore policy

**Files:** `.gitattributes`, `.gitignore`, this PLAN.

1. Replace the leading portion of `.gitattributes` with the canonical template
   bytes. Append the existing repository-specific fixture block unchanged:

   ```gitattributes
   # Byte-exact fixtures: migration_test.go compares the update command's output
   # with these byte for byte, so no checkout may convert their line endings
   # (docs/decisions/0008-PLAN-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md, D4).
   testdata/migration/** -text
   ```

   This appended block is the only allowed delta from the canonical template.
2. Append the canonical `.gitignore` fragment without deleting, sorting, or
   deduplicating existing product entries. The final file must contain these
   previously missing patterns exactly once:

   ```text
   __pycache__/
   *.py[cod]
   coverage_*.out
   *.log
   *.tmp
   tmp/
   temp/
   .cache/
   .worktrees/
   ```

3. Prove the LF rule in an isolated scratch Git repository with
   `core.attributesFile` isolated from host configuration and
   `core.autocrlf=true`:
   * an LF fixture checked out without the repository `.gitattributes` must
     become CRLF;
   * the same indexed bytes checked out with the repository
     `.gitattributes` must remain LF;
   * a path under `testdata/migration/` must report `-text`.
4. Run `git check-attr text eol` on representative Go, shell, Markdown, JSON,
   and fixture paths. Run `git ls-files --eol` and inspect every non-LF or
   non-text result. The only expected non-text tracked paths are explicit
   fixture or binary rules.
5. Compare the canonical `.gitattributes` prefix byte-for-byte and compare the
   merged `.gitignore` entries with the fragment. Record the commands, exit
   statuses, CRLF/LF observation, and named delta in this PLAN.
6. Stage only `.gitattributes`, `.gitignore`, and this PLAN; inspect the staged
   diff; commit with `git commit --no-edit`.

**Phase acceptance:** tracked text has the repository LF attribute, the
byte-exact migration fixtures remain exempt, every canonical ignore entry is
present, and no product ignore entry was removed.

### Phase 2: install docs gates and normalize the documentation tree atomically

This phase combines gate installation with all changes required for those
gates to pass. Do not commit a state in which the new records or Markdown gate
fails.

**New files:**

* `.markdownlint-cli2.jsonc`
* `scripts/check_records.py`
* `docs/architecture.md`
* `docs/reports/0013-GATES-workspace-scaffolding.md`

**Modified files:**

* `Makefile`
* `README.md`
* `docs/README.md`
* `scripts/configure-github.sh` (comment path only)
* records whose live relative links change because of a move
* this PLAN

**Moves:**

| From | To |
| :--- | :--- |
| `docs/0003-MADR-layer-and-harden-ci-cd-quality-gates.md` | `docs/decisions/0003-MADR-layer-and-harden-ci-cd-quality-gates.md` |
| `docs/0003-PLAN-layer-and-harden-ci-cd-quality-gates.md` | `docs/decisions/0003-PLAN-layer-and-harden-ci-cd-quality-gates.md` |
| `docs/0004-MADR-align-cicd-with-magic-cli-remote.md` | `docs/decisions/0004-MADR-align-cicd-with-magic-cli-remote.md` |
| `docs/0004-PLAN-align-cicd-with-magic-cli-remote.md` | `docs/decisions/0004-PLAN-align-cicd-with-magic-cli-remote.md` |
| `docs/0005-MADR-windows-on-demand-compat-tests.md` | `docs/decisions/0005-MADR-windows-on-demand-compat-tests.md` |
| `docs/0005-PLAN-windows-on-demand-compat-tests.md` | `docs/decisions/0005-PLAN-windows-on-demand-compat-tests.md` |
| `docs/plans/0001-PLAN-gemini-provider-and-model-catalog-modernization.md` | `docs/decisions/0001-PLAN-gemini-provider-and-model-catalog-modernization.md` |
| `docs/plans/0002-PLAN-self-update-cli-and-github-releases-integration.md` | `docs/decisions/0002-PLAN-self-update-cli-and-github-releases-integration.md` |
| `docs/cicd-operations.md` | `docs/guides/cicd-operations.md` |

1. Install `.markdownlint-cli2.jsonc` and `scripts/check_records.py` from the
   pinned canonical templates. Confirm each with `cmp` immediately.
2. Integrate the canonical Makefile snippet:
   * place the portable Windows/Unix `SHELL` selection and `.SHELLFLAGS`
     before the first existing variable assignment;
   * add `check-records` and `markdownlint` to the existing `.PHONY` list;
   * place the two canonical recipes after `verify-staged` and before the
     interactive/build convenience targets;
   * do not change `verify`, `verify-staged`, a Go command, a tool pin, or an
     existing target dependency.
3. Create `docs/guides/` by moving the real operations runbook. Create
   `docs/reports/` by writing the real GATES record in step 8. Do not create an
   empty documentation directory or `.gitkeep`.
4. Move each MADR and its PLAN together with `git mv`. Remove the now-empty
   `docs/plans/` directory. Do not renumber, retitle, or change metadata merely
   to modernize an older record.
5. Repair the 23 links demonstrated by the scratch move:

   | File after the move | Required repair |
   | :--- | :--- |
   | `0003-MADR-layer-and-harden-ci-cd-quality-gates.md` | Change two repository-root targets from `../…` to `../../…`: the CI workflow and `Makefile`. |
   | `0004-MADR-align-cicd-with-magic-cli-remote.md` | Change both CI workflow targets from `../.github/…` to `../../.github/…`. |
   | `0004-PLAN-align-cicd-with-magic-cli-remote.md` | Change the CI target to `../../.github/…`, the docs index target to `../README.md`, and the runbook target to `../guides/cicd-operations.md`. |
   | `0005-MADR-windows-on-demand-compat-tests.md` | Change five root targets to begin `../../`: `scripts/go-precheck.py`, `.githooks/pre-commit` twice, `scripts/install-hooks.sh`, and `scripts/verify-scripts.sh`. |
   | `0006-MADR-mcplib-1-6-canary.md` | Change the `0004` link from `../0004-MADR-…` to the sibling `0004-MADR-…`. |
   | `docs/README.md` | Point the eight `0001` through `0005` misplaced-record links into `decisions/`. The final rewrite in step 9 supplies these targets. |
   | `docs/guides/cicd-operations.md` | Point `0003-MADR-…` to `../decisions/0003-MADR-…` and the CI workflow to `../../.github/workflows/ci.yml`. |

   Run the records checker after these repairs. It must report neither a broken
   link nor a placement warning.
6. Repair live plain-text references:
   * change the comment in `scripts/configure-github.sh` from
     `docs/0004-MADR-…` to `docs/decisions/0004-MADR-…` without changing
     executable shell text;
   * change the two operational runbook references in
     `0006-PLAN-mcplib-1-6-canary.md` to
     `docs/guides/cicd-operations.md`.

   Preserve old paths where an accepted MADR or completed PLAN is explicitly
   describing the repository's former state or an execution command that ran
   at that time. Record the reviewed historical allowlist in the GATES file so
   a residual-path search has an explained result rather than an unexplained
   match.
7. Write `docs/architecture.md` from the canonical template with present-tense
   repository facts:
   * **What it is:** a Go CLI and `prepare-commit-msg` hook with `configure`,
     `update`, `version`, and `identity` command paths;
   * **Runtime flow:** staged diff collection in `internal/git`, configuration
     loading in `internal/config`, provider calls through
     `go-llmprovider-sdk`, message writes through `internal/fsutil`, and update
     behavior through `go-selfupdate-lib`;
   * **Human-facing setup:** `internal/ui`, the root README, and the operations
     guide;
   * **Quality and delivery boundary:** Make targets, repository-local hook
     wrappers, the single CI workflow, `selfupdate-release.json`, and the
     external reusable release workflows;
   * **Tree:** the actual post-migration root and docs structure;
   * **What is not here:** provider implementations, self-update internals,
     host identity, global hook configuration, and live GitHub settings.

   Do not include historical rationale, future design, or a placeholder.
8. Create `docs/reports/0013-GATES-workspace-scaffolding.md` with:
   * front matter `status: in-progress`, the execution date, and
     `associated-plan: "0013-PLAN-align-repository-with-workspace-scaffolding.md"`;
   * a body link to `../decisions/0013-PLAN-align-repository-with-workspace-scaffolding.md`;
   * sections for baseline, template hashes, copied-file comparisons, move and
     link checks, scratch plants, actual-tree gates, identifier scan, product
     checks, and remaining work;
   * factual `pending` entries for Phase 3 plants, rather than empty cells or
     invented results.
9. Rewrite `docs/README.md` using the canonical structure while preserving
   repository information:
   * link `architecture.md`, `guides/cicd-operations.md`, `decisions/`, and
     `reports/`;
   * build one table sorted by number, then kind, with Number, Kind, Record,
     and Status columns;
   * index every MADR and PLAN from `0001` through `0013` plus
     `0013-GATES-workspace-scaffolding.md`;
   * take status from front matter where present; for legacy records without
     status metadata, preserve the current index's status rather than inventing
     one;
   * add “I want to…” rows for understanding architecture, installing and
     configuring the hook, running developer checks, performing a release,
     reading the operations runbook, and understanding the workspace decision;
   * defer the `AGENTS.md` navigation row until Phase 3 creates that file.
10. Add a `Documentation` entry to the root README table of contents and a
    `## Documentation` section immediately before `## License`. The section
    links to `docs/README.md`, `docs/architecture.md`, and
    `docs/guides/cicd-operations.md`, and gives a compact “I want to…” excerpt.
    Preserve all product sections and prose.
11. Prove unnumbered-link detection in a scratch copy before relying on it:
    insert a missing relative link into the scratch root README and require
    `check_records.py --check-all` to exit 1 with `broken relative link:`.
    Also confirm the old runbook path is dead from its new directory and its
    repaired targets resolve.
12. Run Markdown lint on the phase tree. Fix failures in `README.md`,
    `docs/README.md`, `docs/architecture.md`, and the moved guide without
    changing technical meaning. MADRs and PLANs remain excluded by the
    canonical config. Do not add an exclusion to conceal a failure.
13. Verify the phase:

    ```bash
    python3 scripts/check_records.py --next
    python3 scripts/check_records.py --check-all
    npx --yes markdownlint-cli2@0.23.2
    make check-records
    make markdownlint
    make verify
    shellcheck scripts/configure-github.sh
    ```

    Require `0014`, zero record/link errors, zero placement warnings, zero
    Markdown errors, and exit 0 from the product checks. `make verify` is run
    here because this phase changes the Makefile and a shell-script comment; do
    not repeat it later unless a subsequent change affects its inputs or this
    run fails.
14. Record complete command output or an exact bounded excerpt plus the exit
    status in the GATES file and this PLAN. Compare both canonical byte-copy
    files with `cmp`.
15. Stage only the Phase 2 files and moves. Run `make check-records`,
    `make markdownlint`, and `make verify-staged` on the staged snapshot.
    Inspect the staged rename detection and confirm no product source, test,
    module, CI, hook, or release file changed. Commit with
    `git commit --no-edit`.

**Phase acceptance:** the fixed docs tree exists; all records and the guide are
in their decided locations; all links and Markdown pass; docs tooling and Make
entry points work; the GATES record contains real Phase 1 and Phase 2 evidence;
the product contract passes; the agent files remain absent until Phase 3.

### Phase 3: install agent instructions and prove every new gate

Agent files land last so `AGENTS.md` is covered by Markdown lint from its first
commit.

**New canonical byte-copy files:**

* `.claude/.gitignore`
* `.claude/rules/madr-and-plan-skill.md`
* `.grok/rules/madr-plan-before-mutating-work.md`
* `.opencode/rules.md`
* `opencode.json`
* `.codex/rules/madr-plan-before-mutating-work.md`
* `.kilo/agent/madr-and-plan-skill.md`
* `.kilo/.gitignore`

**Generated file:** `AGENTS.md`.

1. Generate `AGENTS.md` from the pinned canonical template with an asserted
   exact replacement of each slot. Use this scope paragraph:

   > `prepare-commit-msg` is a Go CLI and Git `prepare-commit-msg` hook that
   > turns staged diffs into Conventional Commit messages through configured
   > local or remote LLM providers. `docs/architecture.md` describes the
   > current system, `docs/README.md` is the documentation and records entry
   > point, and `docs/guides/cicd-operations.md` gives the validation and
   > release sequence.

   Use this dependency policy:

   > No dependency may be added without a MADR in this repository that names
   > it. `go.mod` and `go.sum` change in the same commit as the first import and
   > are cleaned in the commit that removes the last. The dependency MADR owns
   > the exact verification commands and version constraints.

   Use these pre-add rules:

   * After staging any Go or module-file change, run `make verify-staged`.
   * Before every commit, run `make check-records` and `make markdownlint`.
   * Run any additional product check required by the active PLAN phase.
   * A file that fails a required check is not committed.

   Assert that `{{SCOPE}}`, `{{DEPENDENCIES}}`, `{{PREADD}}`, and every other
   `{{…}}` token are absent after replacement. Leave the template's identifier
   and commit sections unchanged.
2. Install all eight supporting files from the pinned canonical templates.
   Do not alter pointer prose, add harness-specific workflow copies, or create
   `CLAUDE.md`.
3. Compare each of the ten byte-copy targets from the MADR with its template
   using `cmp`: the eight files above plus `.markdownlint-cli2.jsonc` and
   `scripts/check_records.py`. Every comparison must exit 0.
4. Add the `AGENTS.md` row to the `docs/README.md` “I want to…” matrix. Confirm
   the root README, docs index, architecture document, runbook, GATES record,
   MADR, PLAN, and `AGENTS.md` are mutually reachable through relative links.
5. Create independent scratch copies for the required plants. Record the exact
   command, exit status, and diagnostic in the GATES file:

   | Gate | Plant | Required observation |
   | :--- | :--- | :--- |
   | Records links | Add `[x](missing.md)` to scratch `README.md`. | `--check-all` exits 1 and prints `broken relative link:`. |
   | Markdown MD004 | Add `* item` to scratch `README.md`. | markdownlint exits nonzero with `MD004/ul-style`. |
   | Markdown MD013 | Add one paragraph of 45 `word` tokens, 224 characters including spaces, to scratch `README.md`. | markdownlint exits nonzero with `MD013` and expected limit 200. |
   | Agent-file coverage | Add the same 224-character paragraph to scratch `AGENTS.md`. | markdownlint exits nonzero with `MD013` for `AGENTS.md`. |
   | Record exclusion | Add the same paragraph to a scratch `NNNN-MADR-*.md`. | markdownlint exits 0 for that plant alone. |
   | Identifier scan | Add the local account name to scratch `AGENTS.md`. | the identifier scan reports the file without printing the identifier in the record. |
   | LF checkout | Use Phase 1's isolated `core.autocrlf=true` experiment. | without the repo rule the fixture is CRLF; with it the fixture is LF. |

   Restore nothing in the working tree because no plant is made there.
6. Scan every added or changed file for the local account name, email and Git
   host domains, hostname domain, and user-profile path prefixes. Read search
   values from the environment at run time. Record only categories and hit
   counts. Require zero hits; placeholders such as `<user>` are allowed.
7. Run the actual-tree gates, redirecting long output and capturing each exit
   status before inspection:

   ```bash
   python3 scripts/check_records.py --next
   python3 scripts/check_records.py --check-all
   npx --yes markdownlint-cli2@0.23.2
   make check-records
   make markdownlint
   git ls-files --eol
   ```

   Require `0014`; no records errors or warnings; no Markdown errors; and
   `attr/text=auto eol=lf` for tracked text except explicit `-text` fixture or
   binary rules.
8. Update the GATES record with every plant and actual-tree result. Leave its
   status `in-progress` until the final phase. Update this PLAN's Phase 3
   execution record, also leaving this PLAN `in-progress`.
9. Stage only the Phase 3 files plus the GATES record, docs index, and this
   PLAN. Run `make check-records` and `make markdownlint` on the staged phase.
   Inspect the staged diff and all `cmp` results. Commit with
   `git commit --no-edit`.

**Phase acceptance:** all agent files are present, the generated slots carry
the reviewed repository policy, every copied file matches, each new gate has a
recorded failing plant, all actual-tree gates pass, and the evidence remains
marked in progress for close-out.

### Phase 4: close out the execution record

**Files:**

* `docs/decisions/0013-PLAN-align-repository-with-workspace-scaffolding.md`
* `docs/reports/0013-GATES-workspace-scaffolding.md`
* `docs/README.md`

1. Start from the committed Phase 3 tree. Confirm `git status --short` is empty
   and inspect the Phase 0 through Phase 3 commits and file lists.
2. Confirm every MADR acceptance criterion using the recorded evidence. Do not
   repeat `make verify` if Phase 2 passed and no later phase changed its inputs;
   cite the Phase 2 result. Re-run only the current docs and records gates:

   ```bash
   python3 scripts/check_records.py --next
   python3 scripts/check_records.py --check-all
   npx --yes markdownlint-cli2@0.23.2
   make check-records
   make markdownlint
   ```
3. Change the GATES status to `complete` and replace every factual `pending`
   entry with the actual result or a named, approved deviation. Change this
   PLAN status to `complete` only if every acceptance criterion holds.
4. Update the `0013` PLAN and GATES statuses in `docs/README.md`. Confirm its
   table indexes every repository record exactly once by full filename.
5. Record in this PLAN:
   * every phase commit and file list;
   * the exact check exit statuses and relevant output;
   * every scratch plant and its failure diagnostic;
   * all ten byte-copy `cmp` results;
   * the generated and merged file deltas;
   * the identifier-scan categories and zero-hit result;
   * anything not done and why;
   * every deviation and its approval, or an explicit statement that there
     were none.
6. Read all three close-out files in full. Run `make check-records` and
   `make markdownlint`; stage only those files; inspect the staged diff; and
   commit with `git commit --no-edit`.
7. After the commit, confirm a clean status and run the two Make documentation
   gates against `HEAD`. Record the commit in the handoff; do not amend the
   close-out commit merely to insert its own hash.

**Phase acceptance:** the PLAN and GATES record are complete, the docs index
agrees with them, the working tree is clean, and the committed tree passes both
repository-owned documentation gates.

## Verification

### Required success matrix

| Concern | Command or observation | Required result |
| :--- | :--- | :--- |
| Sequence | `python3 scripts/check_records.py --next` | Exit 0; `0014`. |
| Record structure and links | `python3 scripts/check_records.py --check-all` | Exit 0; no errors or warnings. |
| Make wrapper | `make check-records` | Exit 0 with the same result. |
| Markdown | `npx --yes markdownlint-cli2@0.23.2` | Exit 0; zero errors. |
| Markdown Make wrapper | `make markdownlint` | Exit 0 with the same result. |
| Canonical files | Ten `cmp` calls | Ten exit-0 results. |
| AGENTS slots | Exact-token scan | No `{{…}}` token; reviewed scope, dependency, and pre-add prose present. |
| Record placement | Repository file inventory | All MADRs/PLANs under `docs/decisions/`; GATES under `docs/reports/`; no `docs/plans/` and no record at the `docs/` root. |
| Guide placement | File and link inventory | Runbook under `docs/guides/`; live links use its new path. |
| LF policy | Isolated checkout plant and `git ls-files --eol` | CRLF without the rule, LF with it; tracked text attributed LF; fixtures explicitly `-text`. |
| Identifier hygiene | Environment-derived scan | Zero real-machine identifier hits in added or changed files. |
| Product contract | `make verify` in Phase 2 | Exit 0 after the Makefile and shell-comment changes. |
| Staged contract | `make verify-staged` in Phase 2 | Exit 0 on the staged phase. |
| Working tree | `git status --short` after Phase 4 | Empty. |

### Required failure plants

The GATES record is incomplete unless it contains observed failures for:

* a missing relative link;
* Markdown MD004;
* Markdown MD013 in the root README;
* Markdown MD013 in `AGENTS.md`;
* an identifier in scratch `AGENTS.md`;
* CRLF checkout without the repository LF rule.

It must also contain the successful control showing that the same long
paragraph in a MADR is excluded from Markdown lint.

### Observable documentation outcome

A reader arriving at the root README can reach the docs index. From the docs
index the reader can:

* understand the current system through `docs/architecture.md`;
* follow development and release operations through
  `docs/guides/cicd-operations.md`;
* find every decision, plan, and gate record in the records table;
* reach `AGENTS.md` for repository workflow rules;
* reach the `0013` MADR to understand why the scaffold exists.

## Rollout and Rollback

This is a repository-only rollout. It changes no binary, dependency, CI job,
remote setting, hook installation, tag, release, or live service.

Each phase lands as its own commit after its checks pass. If a committed phase
must be undone, use a new `git revert` commit for that exact phase, then repair
the MADR/PLAN status and docs index. Do not reset, rewrite, or amend published
history.

Rollback order is the reverse of rollout:

1. Revert Phase 4 status-only close-out.
2. Revert Phase 3 agent files and pointer installation.
3. Revert Phase 2 docs tooling and moves as one unit so MADR/PLAN pairs and
   their links return together.
4. Revert Phase 1 LF and ignore policy.
5. Revert Phase 0 only if the decision itself is withdrawn; mark the MADR
   rejected or superseded rather than deleting accepted history.

Before a commit exists, reverse only files created or moved by the active
phase, and only after confirming they do not contain concurrent user changes.
Never use `git reset --hard`, `git clean`, broad checkout/restore commands, or
stash deletion. A rollback that would remove user work requires a new explicit
instruction.

No push or tag is part of rollout. If the owner later asks to push, run the
repository's disclosure guard over the outgoing commits first and report its
result.

## Execution Record

No phase has been approved or executed. Populate this section during execution
with dated phase entries, command exit statuses, relevant output, commit IDs,
deviations, and work deliberately left undone.
