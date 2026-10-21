---
status: in-progress
date: 2026-10-20
associated-madr: "0019-MADR-skip-full-ci-for-docs-only-pushes.md"
---
# Implement skipping the full CI for a change to Markdown that no code reads, with a docs workflow

Associated MADR: [0019-MADR-skip-full-ci-for-docs-only-pushes.md](0019-MADR-skip-full-ci-for-docs-only-pushes.md)
(option E with D3, as the owner chose).

## Goal

* **D1:** a push or pull request that changes only Markdown under
  `docs/`, `AGENTS.md`, `tools/README.md` or the five agent-pointer
  directories does not start `ci.yml`. A change to `README.md`, to any
  other file, a `workflow_dispatch` and every tag start it as before.
* **D2:** such a change starts `docs.yml`, which runs `make
  check-records` and `make markdownlint` on one Linux runner.
* **D3:** `scripts/check_ci_filters.py` fails when the two filters
  disagree, or when an ignored path is one that code reads.
* **D4:** AGENTS.md, the architecture note, the operations guide and the
  `configure-github.sh` comment describe the split.

## Facts this plan rests on

Each was established on 2026-10-20 at `87f0e64`, read-only (MADR,
Context).

| # | Fact | How |
| :--- | :--- | :--- |
| F1 | `ci.yml`'s triggers are `ci.yml:3-9`; the `go` job's Markdown steps are `ci.yml:48-52`, named "Check records and documentation links" (`make check-records`) and "Lint Markdown" (`make markdownlint`); the checkout pin is `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.0` (`ci.yml:35`, `:76`). | read |
| F2 | `make check-records` is `python3 scripts/check_records.py --check-all` (`Makefile:136-137`); `make markdownlint` is `npx --yes markdownlint-cli2@0.23.2` (`Makefile:139-140`). Neither recipe calls Go. The `Makefile`'s only parse-time command is `git describe --tags --always --dirty` (`Makefile:18`). | read |
| F3 | `readme_test.go:26` reads `README.md`. No other tracked Markdown file is read by a Go file, a `//go:embed` or `tools/*.py`; `testdata/` has no Markdown. | `git grep` over `*.go`, `*.py`, `*.sh`, `*.json`, `*.yml`, `Makefile` |
| F4 | 46 tracked `.md` files; every one except `README.md` is under `docs/`, is `AGENTS.md` or `tools/README.md`, or is under `.claude/`, `.codex/`, `.grok/`, `.kilo/` or `.opencode/`. | `git ls-files '*.md'` |
| F5 | Run `39304468655` (`87f0e64`): 11 jobs, 280 s, one attempt, 0 billable ms (5 Ubuntu, 3 Windows, 2 macOS job runs). The repository is public. | `gh api …/actions/runs/39304468655/timing`, `…/jobs` |
| F6 | Two active rulesets, neither with a required status check; no branch protection (404). | `gh api …/rulesets`, `…/rulesets/<id>`, `…/branches/main/protection` |
| F7 | `scripts/verify-scripts.sh` runs `ast.parse` on every file under `scripts/` whose first line names Python, then `actionlint` on `.github/workflows/*.yml` (`verify-scripts.sh:8-15`, `:24`); `make workflow-lint` runs it, and `make verify` runs that. | read |
| F8 | `.markdownlint-cli2.jsonc`'s globs exclude every MADR and PLAN file, so `make markdownlint` does not lint records. | read |
| F9 | 18 of the 33 commits since 2026-10-10 change Markdown only; 4 of them change `README.md`. | `git log --name-only`, classified by a scratch script |

## Scope

### In scope

| Phase | What | Files |
| :--- | :--- | :--- |
| 0 | this pair, and its `docs/README.md` rows | the two records, `docs/README.md` |
| 1 | the filters, the docs workflow, the filter check, and the script comment | `.github/workflows/ci.yml`, `.github/workflows/docs.yml` (new), `scripts/check_ci_filters.py` (new), `scripts/verify-scripts.sh`, `scripts/configure-github.sh` |
| 2 | the documents | `AGENTS.md`, `docs/architecture.md`, `docs/guides/cicd-operations.md` |
| 3 | the live checks, and closing | this PLAN, the MADR's status, `docs/README.md` |

### Out of scope

* Any job or step of `ci.yml`, and the reusable build and publish
  workflows it calls.
* `README.md`. No phase changes it, so that Phase 2's push is a
  docs-only test (Phase 3).
* The rulesets and repository settings: no required status check exists
  (F6), and none is added.
* The pins and comment 0018-PLAN changes in `ci.yml`. They are that
  plan's.
* Whether `markdownlint` covers the agent-pointer directories. Phase 1
  records what it does; any change to `.markdownlint-cli2.jsonc` needs its
  own record.
* A release. The change is CI and documents only.

## Rules for every phase

* **Branch.** On `main`. The agent stages; the owner commits and pushes.
  Each push is on the owner's word.
* **Order with 0018.** Phase 0's index rows, and Phase 1, wait until
  [0018-PLAN-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md](0018-PLAN-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md)'s
  records are committed, and Phase 1 waits until its `ci.yml` change is
  committed too. Line numbers in `ci.yml` are re-read then, not taken from
  F1.
* **Fail first.** Each new check is run against a deliberately broken
  input in a scratch clone outside the repository, never in this tree.
  The repository has no plant helper. The failure is quoted in the
  Execution Record.
* **Gate, before staging:** `make preflight`; `make preflight-hosts` in
  Phase 1, since it changes scripts and workflows (AGENTS.md, "Before a
  push"); the identifier scan of the staged diff for the local account
  name, the host aliases and the hostname domain, recorded by what was
  scanned for. Each exit status is captured before any filter, and long
  output goes to a scratch file.
* **Assert each edit landed** by reading the result.
* **Deviations.** Stop, record a dated entry here (and in the MADR when a
  decision or a fact changes), offer resolutions that leave the tree
  correct, and wait for the owner's choice.

## Implementation Steps

### Phase 0: records

1. This pair: the MADR `proposed`, this PLAN `proposed`. *(On approval, the
   MADR is `accepted` and this PLAN `in-progress`.)*
2. Once 0018's records are committed, `docs/README.md` gains, after the
   0018 rows:

   ```markdown
   | 0019 | MADR | [Skip the full CI for a change to Markdown that no code reads, and check it with a one-job docs workflow](decisions/0019-MADR-skip-full-ci-for-docs-only-pushes.md) | proposed |
   | 0019 | PLAN | [Implement skipping the full CI for a change to Markdown that no code reads, with a docs workflow](decisions/0019-PLAN-skip-full-ci-for-docs-only-pushes.md) | proposed |
   ```

   The Status cells follow the records' statuses at the time of the commit.
3. `make check-records` and `make markdownlint` exit 0. The owner commits.

### Phase 1: the filters, the docs workflow and the filter check (D1, D2, D3)

1. **`ci.yml`** (D1). `on.push` keeps `branches` and `tags`, and gains,
   as does `on.pull_request`:

   ```yaml
   paths-ignore: ['docs/**.md', 'AGENTS.md', 'tools/README.md', '.claude/**.md', '.codex/**.md', '.grok/**.md', '.kilo/**.md', '.opencode/**.md']
   ```

   `on.pull_request:` becomes a mapping with that one key. A comment above
   `on.push` says: a change to these Markdown paths alone runs
   `docs.yml` instead; GitHub does not evaluate path filters for a tag
   push, so every tag runs this in full; `README.md` is not listed because
   `readme_test.go` reads it; and `scripts/check_ci_filters.py` checks the
   list (this pair). Nothing else in `ci.yml` changes.
2. **`.github/workflows/docs.yml`,** new (D2):

   ```yaml
   name: Docs

   # Checks a change to Markdown that ci.yml skips, with the commands ci.yml's
   # go job runs for it: make check-records and make markdownlint. A change to
   # anything else, README.md included, runs ci.yml, and so does every tag
   # (docs/decisions/0019-MADR-skip-full-ci-for-docs-only-pushes.md).
   # scripts/check_ci_filters.py requires every pattern in ci.yml's
   # paths-ignore to appear in both lists below.
   on:
     push:
       branches: [main]
       paths: ['**.md', 'docs/**.md', 'AGENTS.md', 'tools/README.md', '.claude/**.md', '.codex/**.md', '.grok/**.md', '.kilo/**.md', '.opencode/**.md', '.markdownlint-cli2.jsonc', '.github/workflows/docs.yml']
     pull_request:
       paths: ['**.md', 'docs/**.md', 'AGENTS.md', 'tools/README.md', '.claude/**.md', '.codex/**.md', '.grok/**.md', '.kilo/**.md', '.opencode/**.md', '.markdownlint-cli2.jsonc', '.github/workflows/docs.yml']

   permissions:
     contents: read

   concurrency:
     group: docs-${{ github.ref }}
     cancel-in-progress: true

   jobs:
     docs:
       name: Docs (records and Markdown)
       runs-on: ubuntu-24.04
       timeout-minutes: 10
       steps:
         - name: Checkout
           uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.0
           with:
             persist-credentials: false

         - name: Check records and documentation links
           run: make check-records

         - name: Lint Markdown
           run: make markdownlint
   ```

   The two `run:` lines are `ci.yml`'s, unchanged (F1). The checkout pin
   is `ci.yml`'s current one; if 0018 or a later change has moved it,
   this file takes the pin `ci.yml` has then. There is no setup action
   (F2): the runner image's `make`, `bash`, `git`, `python3` and `npx` are
   what the `go` job's same steps use today.
3. **`scripts/check_ci_filters.py`,** new (D3). Python 3, standard library,
   first line `#!/usr/bin/env python3`, a docstring citing this pair.
   With no arguments it checks the repository it sits in (`git
   rev-parse --show-toplevel`). It prints one line per problem, then a
   summary (`ci filters: N patterns, M reads checked, K problems`), and
   exits 1 if there is any problem. It checks:
   1. `ci.yml` has exactly two `paths-ignore: [...]` lines, one under
      `on.push` and one under `on.pull_request`, each a one-line flow list
      of quoted strings, and the two lists are equal. A missing line, a
      block-style list or a third occurrence is a problem.
   2. `docs.yml` has a `paths: [...]` line under each of `on.push` and
      `on.pull_request`, and every `ci.yml` pattern is in both.
   3. Patterns are translated as the reference states: `**` matches any
      characters, `*` any characters except `/`, `?` one character. A
      pattern that matches no tracked file is a problem, so a typo cannot
      pass unseen.
   4. The *read set* is: each tracked `.md` path named by a string literal
      (`"…"` or `` `…` `` in `.go`; `'…'` or `"…"` in `tools/*.py`), after
      a leading `./` is dropped; each tracked file a `//go:embed` pattern
      matches, read relative to the Go file's directory; and each tracked
      file under `testdata/`. No `ci.yml` pattern may match a path in the
      read set. Today the read set's only Markdown path is `README.md`.
4. **`scripts/verify-scripts.sh`:** after the `actionlint` line,
   `"$PYTHON" "$REPO_ROOT/scripts/check_ci_filters.py"`. The `scripts/`
   loop already parses the new file (F7).
5. **`scripts/configure-github.sh:71-73`:** the comment says `ci.yml` is
   the one workflow that builds and releases, citing 0004-MADR and
   0012-MADR as now, and that `docs.yml` checks Markdown-only changes
   (this pair). The code below it does not change: it tracks `ci.yml`'s
   hardened commit, and `docs.yml` has no write permission.
6. **Seen failing.** Make a scratch clone of `HEAD` outside the
   repository with `git clone --no-local`, copy the phase's five files
   into it, and `git add` them there, since the check lists tracked files.
   Make one fresh clone per plant. In each, run the three commands, and
   record each exit status and the first problem line:

   | # | Plant | Expected to fail |
   | :--- | :--- | :--- |
   | P0 | none | nothing: all three exit 0 |
   | P1 | a relative link in `docs/guides/cicd-operations.md` pointed at a file that does not exist | `make check-records` (`broken relative link: …`) |
   | P2 | `## Routine verification` written `##Routine verification` | `make markdownlint` (MD018) |
   | P3 | both `ci.yml` lists replaced by `['**.md']` | the filter check (`README.md`, read by `readme_test.go`) |
   | P4 | `'README.md'` added to both `ci.yml` lists | the filter check |
   | P5 | `'AGENTS.md'` dropped from `ci.yml`'s `pull_request` list only | the filter check (lists differ) |
   | P6 | `'.kilo/**.md'` dropped from `docs.yml`'s `push` list | the filter check (not in `docs.yml`) |
   | P7 | a tracked `zz_plant_test.go` holding `os.ReadFile("docs/architecture.md")` | the filter check |
   | P8 | a tracked `zz_plant.go` holding `//go:embed docs/architecture.md` | the filter check |
   | P9 | both `paths-ignore` lines deleted from `ci.yml` | the filter check (no list found) |
   | P10 | `'docs/**.mdx'` added to both `ci.yml` lists and to `docs.yml` | the filter check (matches no tracked file) |

   The three commands are `make check-records`, `npx --yes
   markdownlint-cli2@0.23.2` and `python3 scripts/check_ci_filters.py`.
   Each plant must fail only its own command. Also record, without acting
   on it, whether `npx --yes markdownlint-cli2@0.23.2` fails on an MD018
   planted in `.claude/rules/madr-and-plan-skill.md` (Out of scope).
7. **Checks:** `make workflow-lint` (actionlint on both workflows, the
   parse and the new check), then the phase gate: `make preflight`, `make
   preflight-hosts`, the identifier scan. Stage the five files. The owner
   commits.

### Phase 2: the documents (D4)

Markdown only, and none of it `README.md`, so its push tests D1.

1. **`AGENTS.md`:** under "Before a push", one paragraph. A change to
   Markdown under `ci.yml`'s `paths-ignore` runs only `docs.yml` in CI. A
   change to `README.md` or any other file runs `ci.yml`, as does every
   tag. A test, embed or tool that reads a Markdown file keeps that file
   out of the list in the same commit, and `scripts/check_ci_filters.py`
   checks it. The MADR is cited in backticks, as AGENTS.md cites records.
2. **`docs/architecture.md`:**
   * "Quality and delivery boundary": after the sentence on `ci.yml`, a
     sentence on the split and the tag rule, linking the MADR;
   * "Tree": a line `.github/workflows/docs.yml  Markdown-only checks`
     after `ci.yml`'s, aligned with it.
3. **`docs/guides/cicd-operations.md`:**
   * "Routine verification": after "CI's `go` job runs the same three
     targets", which workflow a push starts, and that `make preflight`
     still runs all three gates either way;
   * "Controlled release": after step 1, a note that when `main`'s
     commits since its last "CI" run change ignored Markdown only, that
     run tested the same code, and the tag's unfiltered run is the
     release's proof.
4. **Checks:** `make check-records`, `make markdownlint`, the identifier
   scan. `make preflight-hosts` is exempt (docs only). Stage the three
   files. The owner commits.

### Phase 3: live checks and closing

1. **Phase 1's push** changes workflows and scripts, so it starts "CI";
   it also changes `docs.yml`, so it starts "Docs". Record both run IDs,
   their conclusions, "CI"'s job count and "Docs"' duration.
2. **Phase 2's push** changes ignored Markdown only. GitHub's run list for
   that commit (`gh api "repos/maccavelli/prepare-commit-msg/actions/runs?head_sha=<sha>"`)
   must hold one "Docs" run and no "CI" run. Record both facts, and the
   run's duration.
3. **Left to later commits,** each recorded here when it happens or named
   as left:
   * the next push that changes `README.md` starts "CI" (0018-PLAN's
     README phase is likely to be the first);
   * the next release tag starts a full "CI", recorded in that release's
     PLAN.
4. **Closing:** "Verification, at closing" in the Execution Record; this
   PLAN `complete`; `docs/README.md`'s row for it `complete`. The MADR
   stays `accepted`.

## Verification

* **V1 (filters):** `ci.yml` has D1's list on `push` and `pull_request`
  and nothing else changed; `docs.yml` matches Phase 1 step 2; `make
  workflow-lint` passes.
* **V2 (seen failing):** P1–P10 each fail only their own command, P0
  passes all three, and the outputs are quoted.
* **V3 (live, docs-only):** Phase 2's commit has a "Docs" run and no "CI"
  run.
* **V4 (live, code):** Phase 1's commit has a "CI" run in which every job
  passes and only the publish job is skipped, and a "Docs" run.
* **V5 (README and tags):** the next README push and the next tag each
  start "CI"; recorded when they happen, or named as left.
* **V6 (docs):** AGENTS.md, the architecture note and the guide say so;
  `make check-records` and `make markdownlint` are clean; the identifier
  scan is clean.

## Rollout and Rollback

* **Rollout:** Phases 1 and 2 on `main`, each pushed on the owner's word,
  Phase 1 first. Phase 2 is the docs-only test, so it is not folded into
  Phase 1's push.
* **Rollback:** `git revert` of Phase 1's commit restores the full CI for
  every push and removes `docs.yml` and the check. Phase 2's sentences
  then describe a split that no longer exists, so its commit is reverted
  too, or amended under this number. Nothing outside the repository
  depends on the split. No setting or ruleset changes, so there is nothing
  to restore on GitHub.

## Execution Record

### Approval (2026-10-20)

The owner approved option E with D3, the filter check ("Approve, with
D3"). The MADR became `accepted`, and this plan `in-progress`.

Phase 0's index rows wait: `docs/README.md` holds the uncommitted rows of
the 0018 pair, which another change owns. The rows go in after 0018 is
committed, and Phase 1 waits for 0018's `ci.yml` commit, as Phase 1
says.
