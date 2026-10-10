---
status: proposed
date: 2026-10-10
decision-makers: Project Owner
informed: Repository maintainers and coding agents
---

# Adopt the fleet workspace scaffold and normalize the documentation tree

## Context and Problem Statement

This repository is an existing Go product with repository-owned build, test,
release, and staged-snapshot gates. It does not yet have the in-repository
agent workspace defined by the fleet `workspace-scaffolding` standard.

[0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md](0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md)
recorded the missing workspace and split documentation tree, then explicitly
left both for a separate decision. This MADR is that decision. It does not
change the product, its dependencies, its CI workflows, or its release policy.

The repository is classified as **existing product, no workspace**:

* Product source, tests, a `Makefile`, CI, local hook wrappers, and twelve
  existing decision numbers are present.
* `AGENTS.md`, every required harness pointer, the records checker, and the
  Markdown configuration are absent.
* The repository's purpose is the `prepare-commit-msg` hook and CLI. No
  adjacent stack has accumulated its own records, so the fixed documentation
  layout requires one root `docs/` tree.

### Assessment findings

The read-only assessment compared the working tree with the canonical
workspace templates and found the following gaps.

| Area | Evidence in the repository | Alignment required |
| :--- | :--- | :--- |
| Agent instructions | `AGENTS.md` is absent. | Generate it from the canonical template with repository-specific scope, dependency, and pre-add sections. |
| Harness coverage | The Claude, Grok, OpenCode, Codex, and Kilo pointer files are absent. Their two harness-local ignore files and root `opencode.json` are also absent. | Install all five pointers and the supporting files byte-for-byte from the canonical templates. |
| Record tooling | `scripts/check_records.py` and `.markdownlint-cli2.jsonc` are absent. | Copy both canonical files byte-for-byte. |
| Make entry points | The `Makefile` has product targets, including `verify` and `verify-staged`, but no portable `SHELL`, `check-records`, or `markdownlint` targets. | Add the canonical portable shell setup and the two documentation targets without inventing or replacing Go gates. |
| Record placement | Six records, the `0003` through `0005` MADR/PLAN pairs, sit directly in `docs/`. The `0001` and `0002` plans sit in `docs/plans/`. | Move all eight files into `docs/decisions/` without renumbering, and repair links at their new depths. |
| Guide placement | `docs/cicd-operations.md` is a task-oriented operations guide at the `docs/` root. | Move it to `docs/guides/cicd-operations.md` and repair inbound and outbound links. This real guide creates `docs/guides/`. |
| Current architecture | `docs/architecture.md` is absent. | Add a present-tense description of the CLI, its packages, hooks, configuration, quality gates, and release boundary. Historical rationale stays in records. |
| Documentation entry points | The root `README.md` does not link to `docs/README.md`. The docs index is a hand-kept list and has no architecture link or “I want to…” matrix. | Add the root link, retain the hand-kept status information, convert the record index to the fixed table shape, and add task-oriented navigation. |
| Reports | `docs/reports/` is absent. | Create it with the implementation's real `0013-GATES-workspace-scaffolding.md` verification record; do not create an empty directory. |
| Line endings | `.gitattributes` contains only the byte-exact migration-fixture exception. `git ls-files --eol` reports no repository text attribute for tracked text files. | Add the canonical LF and binary rules, preserving `testdata/migration/** -text` as the named repository delta. |
| Ignore rules | `.gitignore` has product entries but lacks the scaffold fragment's Python, scratch, cache, temporary-file, and worktree patterns. | Merge the missing fragment entries; preserve every existing product entry. |
| Commit-message policy | The active repository hook path is a repository wrapper. Its recorded previous path matches the global hooks setting, and all three wrapper files are present. | Retain the standard `git commit --no-edit` policy; the wrapper satisfies the requirement to chain to the global hook directory. |

A complete scratch copy with the canonical records checker produced:

* `--next`: exit 0 and `0013`;
* `--check-all`: exit 0 with no broken relative links;
* eight placement warnings for the records outside `docs/decisions/`.

The placement warnings are structural findings even though the current checker
deliberately reports them as warnings. Moving the six root records one level
deeper will change their relative paths, so the move and link repair must be
one planned operation. The two files under `docs/plans/` remain at the same
depth when moved, but their links still require verification from the new
location.

## Decision Drivers

* Give every supported coding harness the same mutation gate and record
  workflow from a fresh clone.
* Preserve the repository's existing Go, hook, CI, and release controls.
* Put every document in the fixed, navigable tree without renumbering or
  rewriting historical rationale.
* Make numbering, pairing, placement, relative links, Markdown style, and LF
  checkout behavior observable through repository-owned commands.
* Keep copied scaffold files identical to their canonical templates so a
  future refresh can detect drift with `cmp`.
* Make the change reviewable and reversible in phases while unrelated
  uncommitted work is preserved.

## Considered Options

* Install the complete workspace scaffold and normalize the existing docs tree
* Install only the agent gate and leave the documentation layout unchanged
* Keep the repository-specific conventions and defer workspace alignment

## Decision Outcome

Chosen option: "Install the complete workspace scaffold and normalize the
existing docs tree", because it establishes one enforceable workflow for all
five harnesses, makes the existing records conform to the repository-wide
layout, and retains the product's established gates as repository-specific
policy.

This outcome is proposed for owner review. Acceptance authorizes writing a
same-slug implementation plan; it does not by itself authorize executing that
plan.

### D1. Use one root documentation tree

The repository has no adjacent stack that warrants a second tree. The target
layout is:

```text
README.md
docs/
  README.md
  architecture.md
  decisions/   NNNN-MADR-*.md and NNNN-PLAN-*.md
  reports/     NNNN-REPORT-*.md and NNNN-GATES-*.md
  guides/      task-oriented, unnumbered guides
```

The existing records keep their numbers, slugs, status, rationale, and Git
history. Moves use `git mv`. `docs/cicd-operations.md` becomes the first real
guide. The implementation's verification record becomes the first real report
tree entry.

### D2. Install the canonical scaffold-owned files

The following files must match the canonical templates byte-for-byte:

* `.claude/.gitignore`;
* `.claude/rules/madr-and-plan-skill.md`;
* `.grok/rules/madr-plan-before-mutating-work.md`;
* `.opencode/rules.md`;
* `opencode.json`;
* `.codex/rules/madr-plan-before-mutating-work.md`;
* `.kilo/agent/madr-and-plan-skill.md`;
* `.kilo/.gitignore`;
* `.markdownlint-cli2.jsonc`;
* `scripts/check_records.py`.

`AGENTS.md`, `docs/README.md`, and `docs/architecture.md` are generated from
template slots and therefore carry named repository content. `.gitignore` is a
merge. `.gitattributes` keeps the migration fixture's `-text` rule as its named
delta from the template.

### D3. Preserve and name repository-specific policy in `AGENTS.md`

The scope paragraph identifies this repository as the Go CLI and Git hook that
generates commit messages from staged changes. It points to
`docs/architecture.md` for the current structure and `docs/README.md` for the
documentation entry point.

The dependency section requires a MADR before adding a dependency and keeps
lockfile or module-file changes with the first import and last removal. It
does not duplicate dependency versions from `go.mod`.

The pre-add section retains the existing product gate:

* staged Go and module changes run the repository's `make verify-staged`
  contract;
* every commit runs `make check-records` and `make markdownlint` after those
  targets are installed;
* files that fail a required check are not committed.

The commit section uses `git commit --no-edit`. No local identity, host path,
or machine-specific setting is added.

### D4. Extend the Makefile without replacing product gates

The canonical portable Bash selection and `.SHELLFLAGS` are added because the
repository already has a `Makefile`. `check-records` and `markdownlint` are
added to `.PHONY` and copied from the standard snippet.

`verify`, `verify-staged`, the repository hooks, the Go tool bootstrap, and CI
remain product-owned. Whether the documentation targets later join a broader
product aggregate requires an explicit plan step; the scaffold does not invent
a new Go gate or CI workflow.

### D5. Normalize navigation and repair links as one change

The implementation moves all misplaced records and the operations guide,
repairs links using the records checker output, and updates every known inbound
reference. It then rewrites `docs/README.md` as the required hand-maintained
records table and “I want to…” matrix, retaining status and topic information
that remains useful.

The root `README.md` gains a link to `docs/README.md` without replacing product
content. `docs/architecture.md` describes only the current system. Historical
claims and decision rationale remain in MADRs and PLAN execution records.

### D6. Add LF and ignore policy through named merges

`.gitattributes` gains the canonical `* text=auto eol=lf` and binary extension
rules. The existing `testdata/migration/** -text` exception remains because
those fixtures are byte-exact test data.

`.gitignore` keeps all product entries and gains the missing canonical patterns:

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

### D7. Prove each new gate before relying on it

The implementation must validate the final tree with:

```bash
python3 scripts/check_records.py --next
python3 scripts/check_records.py --check-all
npx --yes markdownlint-cli2@0.23.2
make check-records
make markdownlint
git ls-files --eol
```

Every new gate must first fail on the canonical planted input in a scratch
copy. This includes a broken relative link, Markdown list style, Markdown line
length in `README.md` and `AGENTS.md`, the record exclusion, the identifier
scan, and the isolated line-ending checkout experiment. The working tree must
not be dirtied for these experiments.

The final verification record must name each plant, its observed failure, each
command's exit status, the copied-file `cmp` results, and the identifier scan.

### Consequences

* Good, because a fresh clone gives all five supported harnesses the same
  mutation gate and record-writing entry point.
* Good, because records, guides, architecture, and observations have one
  predictable location and one repository-wide sequence.
* Good, because existing Go and release controls remain intact and visible in
  the repository-specific `AGENTS.md` slots.
* Good, because copied-file comparisons and scratch plants distinguish a real
  gate from an unexercised configuration file.
* Bad, because eight historical records and one guide move, requiring careful
  link repair and a larger documentation diff.
* Bad, because Markdown and line-ending rules may expose pre-existing cleanup
  work that must be recorded as a plan deviation rather than bypassed.
* Neutral, because the scaffold adds Node-based Markdown lint invocation while
  keeping Node artifacts outside the repository.

### Confirmation

The decision is implemented only when all of the following are true:

* every scaffold-owned byte-copy target is present and `cmp`-identical to its
  template;
* every generated or merged file has only the named repository delta;
* all five harness pointers direct agents to `AGENTS.md` and load
  `madr-and-plan-writing` by its exact name;
* all numbered records are under `docs/decisions/` or `docs/reports/`, every
  link resolves, and the next unused number is `0014`;
* the root README, docs index, architecture document, operations guide, and
  `0013-GATES-workspace-scaffolding.md` are mutually reachable;
* the records, Markdown, line-ending, and identifier gates pass on the real
  tree and have produced the expected failures on scratch plants;
* the repository's existing product checks still pass in the phases that touch
  their inputs;
* the associated PLAN records the actual output, deviations, and anything
  intentionally left undone.

## Pros and Cons of the Options

### Install the complete workspace scaffold and normalize the existing docs tree

* Good, because it meets the whole standard instead of leaving structural
  warnings and harness-specific behavior behind.
* Good, because future refreshes can compare canonical files mechanically.
* Good, because it preserves the product gates and records their use in the
  shared agent policy.
* Bad, because moving historical files makes link repair and review more
  demanding than installing pointers alone.

### Install only the agent gate and leave the documentation layout unchanged

* Good, because it would reduce the initial move and link-repair work.
* Neutral, because agents could learn the mutation rule from `AGENTS.md`.
* Bad, because the canonical records checker would continue warning about
  eight misplaced records.
* Bad, because there would still be no architecture document, guide library,
  reports location, or task-oriented docs navigation.
* Bad, because the repository would claim workspace alignment while omitting
  standards that make the gate auditable.

### Keep the repository-specific conventions and defer workspace alignment

* Good, because it requires no immediate migration.
* Neutral, because the existing Go and release gates continue to protect the
  product.
* Bad, because fresh clones still lack shared agent instructions and four of
  the five harnesses have no repository-local entry point.
* Bad, because numbering, placement, Markdown, links, and LF behavior remain
  conventions without the workspace-owned checks.
* Bad, because the split documentation tree becomes harder to normalize as
  more records and guides accumulate.

## More Information

The assessment was read-only until this MADR was created. Existing source,
tests, CI, hooks, product configuration, records, and unrelated uncommitted
work were not changed.

If accepted, the next authoring step is
`0013-PLAN-align-repository-with-workspace-scaffolding.md`. The plan must split
the work into reviewable phases, include the documentation moves and their link
repairs in the same phase, and preserve unrelated work already present in the
working tree.

The workspace standard does not authorize `.editorconfig`, Dependabot,
`SECURITY.md`, `CODEOWNERS`, `CONTRIBUTING.md`, `CHANGELOG.md`, a new CI
workflow, application-specific Claude settings, extra Kilo configuration, or
local Git identity. None is part of this decision.
