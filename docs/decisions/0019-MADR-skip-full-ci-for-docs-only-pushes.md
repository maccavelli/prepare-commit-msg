---
status: accepted
date: 2026-10-20
decision-makers: Project Owner
consulted: "GitHub's workflow syntax reference (path filters), read 2026-10-20; the Actions timing and jobs APIs for run 39304468655; go-selfupdate-lib docs/decisions/0028-MADR-skip-full-ci-for-docs-only-pushes.md; this repository at 87f0e64"
informed: maintainers and agents working in this repository
---
# Skip the full CI for a change to Markdown that no code reads, and check it with a one-job docs workflow

## Context and Problem Statement

Every claim below is marked **[read]** (from this repository at `87f0e64`,
or the record or page named), **[probed]** (run read-only in this
investigation) or **[reported]** (the owner's words, not reproduced).

The owner asked, on 2026-10-20, for this repository to stop running its
full CI for pushes and pull requests that change only Markdown, the same
way go-selfupdate-lib has just done **[reported]**.

**What runs today** (`.github/workflows/ci.yml`) **[read]**:

* **Triggers** (`ci.yml:3-9`): `push` to `main` and to `v*` tags,
  `pull_request`, and `workflow_dispatch`. There is no path filter.
* **`go`** (`ci.yml:29-55`), `ubuntu-24.04`: `make check-records`
  (`ci.yml:48-49`), `make markdownlint` (`ci.yml:51-52`), then `make
  verify` (`ci.yml:54-55`).
* **`go-native`** (`ci.yml:60-87`): `go test ./...` on `ubuntu-24.04`,
  `macos-15` and `windows-2025`.
* **`build`** (`ci.yml:94-107`) calls go-selfupdate-lib's reusable build
  workflow. On a tag it builds the release; on any other ref it rehearses
  it. **`release`** (`ci.yml:112-126`) publishes, on a tag only.

So the only steps that check Markdown are `go`'s first two, and both are
`make` targets: `check-records` is `python3 scripts/check_records.py
--check-all` (`Makefile:136-137`), the naming, pairing and relative-link
check, and `markdownlint` is `npx --yes markdownlint-cli2@0.23.2`
(`Makefile:139-140`) **[read]**.

**What a run costs** **[probed]**. The run for `87f0e64`, a commit that
changes two Markdown files (`docs/README.md` and a PLAN), is
`39304468655`. It has 11 jobs: `go`, three `go-native`, the build and five
identity legs, and the publish job, skipped. It took 280 s of wall time,
in one attempt. The timing API reports 0 billable ms for every job: 5
Ubuntu, 3 Windows and 2 macOS job runs. The repository is public
(`visibility: public`), so its CI bills no minutes. The macOS identity leg
started 58 s after the other four, which had started within 2 s of each
other.

**How often it would apply** **[probed]**. Of the 33 commits on `main`
since 2026-10-10, 18 change Markdown only. Four of those change the root
`README.md`, which, as the next paragraph shows, must keep the full run.
So 14 of the 33 would have skipped it.

**Code reads one tracked Markdown file: the root `README.md`** **[read]**:

* `readme_test.go:26` reads `README.md`, and `TestReadmeCuratedModels`
  fails unless its provider table lists exactly the SDK's curated catalog
  in order (0009-MADR D4, cited at `readme_test.go:24`). It runs in the
  `go` job (`make verify` runs `make test` and `make coverage`) and in all
  three `go-native` jobs. A change to `README.md` alone can therefore fail
  CI. That is the point of the test: an SDK upgrade is followed by a
  README edit, as 0018's PLAN does with its "Changes with" section.
* No `//go:embed` names a Markdown file. The one embed is
  `selfupdate-release.json` (`update.go:18`). The commit-message prompt is
  built in Go (`main.go:502`), not read from a template file.
* The other `.md` names in code are not reads of tracked files.
  `internal/git/git_test.go:112-113` writes a `README.md` into a temporary
  repository and stages it. The remaining matches are comments
  (`migration_test.go:162`, `releasespec_test.go:52`,
  `internal/ui/setup_test.go:861`) and record citations in comments.
* `testdata/` holds no Markdown. The release spec names no Markdown asset,
  and the pinned publish workflow writes its notes with `--generate-notes`
  (go-selfupdate-lib `publish-selfupdate-release.yml` at `436a7f5`, line
  202). `opencode.json` names `.opencode/rules.md` for an agent, not for a
  build or a test.

**The tracked Markdown files** **[probed]**: 46. They are all under
`docs/` (the records, `docs/README.md`, `docs/architecture.md`,
`docs/guides/cicd-operations.md`, a GATES report), plus `AGENTS.md`,
`README.md`, `tools/README.md`, and one agent pointer each under
`.claude/`, `.codex/`, `.grok/`, `.kilo/` and `.opencode/`. Outside those
paths, the only one is `README.md`.

**GitHub's path filters** (workflow syntax reference,
`on.<push|pull_request>.<paths|paths-ignore>` and "Git diff comparisons")
**[read]**:

* "Path filters are not evaluated for pushes of tags." A release tag keeps
  the full run.
* `paths` and `paths-ignore` cannot be used together for one event. "If
  you want to both include and exclude path patterns for a single event,
  use the `paths` filter prefixed with the `!` character." The page shows
  no `!` pattern inside `paths-ignore`, so `paths-ignore: ['**.md',
  '!README.md']` relies on behaviour the reference does not state.
* In `paths`, order matters: a negative pattern after a positive match
  excludes the path, and a positive pattern after a negative match
  includes it again.
* Branch and path filters must both be satisfied.
* Pushes use a two-dot diff and pull requests a three-dot diff. A push of
  more than 1,000 commits, or a diff that times out, always runs. With
  more than 3,000 changed files, a filter match outside the first 3,000
  does not start the workflow.
* A workflow skipped by a path filter leaves its checks "Pending", which
  blocks a pull request that requires them.
* The reference's cheat sheet was not read verbatim in this
  investigation. In particular, whether `*` and `**` match a path under a
  dot directory such as `.github/` or `.claude/` is **not established**.

**Repository settings** **[probed]**:

* Two rulesets, both active. `prepare-commit-msg-main` has `deletion` and
  `non_fast_forward`. `prepare-commit-msg-release-tags` has `creation`,
  `deletion` and `non_fast_forward`. Neither requires a status check.
* `branches/main/protection` returns 404 ("Branch not protected").
* `scripts/configure-github.sh` writes the same rulesets
  (`configure-github.sh:153-188`), with `sha_pinning_required: true` and
  only GitHub-owned actions plus go-selfupdate-lib's two reusable workflows
  allowed (`configure-github.sh:103-119`) **[read]**. Its comment at
  `configure-github.sh:71-73` says "ci.yml is the one workflow".
* `docs/guides/cicd-operations.md` says the GitHub repository is a one-way
  mirror, and a public pull request is not merged
  (`cicd-operations.md`, "Dependency and action-pin updates") **[read]**.

**The release procedure** (`cicd-operations.md`, "Controlled release")
asks for `main` to pass `make verify` locally before the tag. It does not
ask for a green CI run on `main` **[read]**.

**The sibling's decision** **[read]**: go-selfupdate-lib
`docs/decisions/0028-MADR-skip-full-ci-for-docs-only-pushes.md`, accepted
and implemented on 2026-10-20. It added `paths-ignore: ['**.md']` to its
`ci.yml`'s `push` and `pull_request` triggers, and a one-job `docs.yml`,
filtered with `paths:`, that runs its CI's Markdown checks with the same
commands (its option B). There, no code reads a tracked Markdown file. Its
CI was 18 jobs and 15–20 minutes; its first docs-only push ran one "Docs"
job in 12 s (run `39482188668`) and no CI. Its CI also billed 0 minutes,
and its owner went ahead for turnaround, less exposure to the macOS runner
queue, and fewer runs.

**Which raises the question:** how should a change to Markdown alone skip
the full CI here, when one Markdown file is an input to a test?

## Decision Drivers

* A change to a file that code reads keeps the full run. Today that is
  `README.md`.
* A Markdown-only change keeps the checks CI runs on Markdown:
  `check-records` and `markdownlint`.
* Every tag keeps the full run.
* A filter that matches less than intended must cost an extra run, never a
  skipped one.
* No third-party action, and nothing that hides a failure: a skipped job
  must not read as a pass.
* A later change cannot quietly disable CI for code: a new Markdown read
  under an ignored path, or a broadened pattern, fails a check.

## Considered Options

* **A.** `paths-ignore: ['**.md']` on `ci.yml`'s `push` and `pull_request`
  triggers, and nothing else.
* **B.** As A, plus a one-job `docs.yml` (go-selfupdate-lib's choice).
* **C.** One workflow: a first job decides from `git diff --name-only`
  whether code changed, and every heavy job carries an `if:` on its
  output; a docs job always runs.
* **D.** `[skip ci]` in the commit message of a docs-only commit.
* **E.** `paths-ignore` naming only the Markdown no code reads (every
  tracked `.md` except `README.md`), a one-job `docs.yml`, and a check that
  keeps the two filters and the code's Markdown reads consistent.
* **F.** As E, but `ci.yml` uses `paths: ['**', '!**.md', 'README.md']`
  instead of `paths-ignore`.

## Decision Outcome

Chosen option: "E", with D3 (the owner, 2026-10-20: "Approve, with D3"),
because it is B with the one correction
this repository's facts require. A change to `README.md` still runs
`TestReadmeCuratedModels` on four runners, and every other Markdown-only
change runs one Linux job. Every pattern in `paths-ignore` names Markdown
under a fixed path, so a pattern that matches less than intended starts
the full run rather than skipping it.

### D1. `ci.yml`'s filter

`on.push` keeps `branches: [main]` and `tags: ["v*"]`, and gains, as does
`on.pull_request`:

```yaml
paths-ignore: ['docs/**.md', 'AGENTS.md', 'tools/README.md', '.claude/**.md', '.codex/**.md', '.grok/**.md', '.kilo/**.md', '.opencode/**.md']
```

* `README.md` is not listed. A change to it, alone or with others, runs
  the full CI.
* A Markdown file at a path not listed, such as a future
  `internal/x/README.md` or a `testdata/` file, also runs the full CI. That
  is the safe direction: a new kind of Markdown file must be added to the
  list on purpose.
* A comment above the filter cites this record, says that a tag push is
  never filtered, and says that `README.md` is excluded because
  `readme_test.go` reads it.
* `workflow_dispatch`, the jobs and their steps do not change.

### D2. `.github/workflows/docs.yml`

A new workflow, named `Docs`:

* **Triggers:** `push` to `main` and `pull_request`, each with `paths:`
  set to `**.md`, D1's eight patterns, `.markdownlint-cli2.jsonc` and
  `.github/workflows/docs.yml`. D1's patterns are listed again, even
  though `**.md` may already cover them, because whether `**` matches
  under a dot directory is not established. Listing them means no path
  that `ci.yml` ignores can be missed by both workflows.
* **Settings:** `permissions: contents: read`, and the concurrency group
  `docs-${{ github.ref }}` with `cancel-in-progress: true`.
* **One job,** `ubuntu-24.04`, `timeout-minutes: 10`. Its checkout is
  pinned to the SHA `ci.yml` uses (`3d3c42e5…`, `v7.0.0`), with
  `persist-credentials: false`. It runs `ci.yml`'s two Markdown steps with
  the same names and commands: `make check-records`, then `make
  markdownlint`.
* No setup action. Both targets need only `make`, `bash`, `git`,
  `python3` and `npx` from the runner image, as the `go` job's same two
  steps already do; neither recipe calls Go.
* A header comment says what the workflow is for and cites this record.

A push that changes code and Markdown runs both workflows, so the two
checks run twice.

### D3. A check that keeps the filters honest

`scripts/check_ci_filters.py`, Python 3 on the standard library, run by
`scripts/verify-scripts.sh` and so by `make workflow-lint`, `make verify`,
`make preflight` and CI's `go` job. It fails unless:

1. `ci.yml`'s `push` and `pull_request` triggers carry the same
   `paths-ignore` list, written on one line in flow style. A list it cannot
   find is a failure, not a pass.
2. Every pattern in that list also appears in `docs.yml`'s `paths` for
   `push` and for `pull_request`.
3. No pattern in that list matches a tracked file that code reads. That
   is a tracked Markdown path named by a string literal in a tracked `.go`
   file or in `tools/*.py`, a `//go:embed` pattern, or a tracked file under
   `testdata/`. Patterns are matched as GitHub states them: `*` matches
   any characters except `/`, and `**` matches any characters.
4. Every pattern in that list matches at least one tracked file, so a
   typo cannot pass unseen.

Today it passes, since `README.md` is the only such path and no pattern
matches it. It fails if someone writes `'**.md'` back, adds `README.md` to
the list, or adds a test that reads `docs/architecture.md`.

It cannot see a path built at run time, such as `filepath.Join("docs",
name)`. D4's AGENTS.md rule covers that case.

### D4. Documents

* `AGENTS.md`, "Before a push": a Markdown-only change
  under D1's paths runs `docs.yml` in CI, not `ci.yml`. A test, embed or
  tool that reads a Markdown file keeps that file out of `ci.yml`'s
  `paths-ignore` in the same commit.
* `docs/architecture.md`, "Quality and delivery boundary": the split, and
  that a tag always runs the full CI. Its "Tree" lists `docs.yml` beside
  `ci.yml`.
* `docs/guides/cicd-operations.md`: "Routine verification" says which
  workflow a push starts, and "Controlled release" says the tag's full run
  is the code proof when `main`'s last commit was docs-only.
* `scripts/configure-github.sh:71-73`: the comment says `ci.yml` is the
  one workflow that builds and releases, and `docs.yml` checks Markdown.
  Only the comment changes.

### Consequences

* Good, because 14 of the last 33 commits would have run one Linux job
  instead of 10 jobs on three operating systems, with no macOS runner to
  wait for.
* Good, because a change to `README.md` still runs the test that reads
  it, on all four runners.
* Good, because a code push, a `workflow_dispatch` and every tag run as
  before.
* Good, because D3 turns the rule "nothing CI ignores is read by code"
  into a check that fails, where B leaves it to memory.
* Neutral, because the saving is time and runner jobs, not money: the
  repository is public and its CI bills 0 minutes. A run here is 280 s,
  not the sibling's 15–20 minutes, so the gain is smaller than there.
* Neutral, because a push that changes code and Markdown runs both
  workflows.
* Neutral, because before a tag, `main`'s last full CI run may be for an
  earlier commit when the commits since change Markdown only. The code is
  the same, and the tag's own run is unfiltered.
* Bad, because the ignore list names paths, so a new directory of
  ignorable Markdown runs the full CI until the list names it.
* Bad, because a required status check, if a ruleset ever adds one, must
  not name a `ci.yml` job: a docs-only pull request would leave it
  "Pending". The record that adds one must say so.
* Bad, because two workflows and a new script are three more things for
  actionlint, `verify-scripts.sh` and a reviewer to cover.

### Confirmation

* `make workflow-lint` passes: actionlint on both workflows, `ast.parse`
  on the new script, and the new script itself.
* On a scratch clone outside the repository, `make check-records` fails on
  a planted broken link and `make markdownlint` on a planted MD018 error,
  and both pass on the clean clone. `scripts/check_ci_filters.py` fails on
  each planted break D3 names, and on a `paths-ignore` list that differs
  between `push` and `pull_request`, and passes on the clean clone.
* After the change lands, a push that changes only files under D1's paths
  shows a "Docs" run and no "CI" run, and a push that changes a
  non-Markdown file shows "CI". The PLAN records both run IDs.
* The next push that changes `README.md` shows "CI", and the next release
  tag shows a full "CI" run. Each is recorded when it happens.

## Pros and Cons of the Options

### A. `paths-ignore: ['**.md']` only

* Good, because it is one line per trigger.
* Bad, because a README-only change skips `TestReadmeCuratedModels`, the
  one test that reads Markdown.
* Bad, because a docs-only change then gets no `check-records` or
  `markdownlint`, and a broken link reaches `main` unchecked.

### B. `paths-ignore: ['**.md']` and a docs workflow

* Good, because the docs keep their checks, and it matches
  go-selfupdate-lib's workflow exactly.
* Bad, because here, unlike there, code reads a Markdown file. A
  README-only change would skip the test that exists to catch a README
  that has drifted from the SDK, and the next code push would fail on a
  README already on `main`.

### C. One workflow, gated jobs

* Good, because one file holds it all.
* Bad, because every job gains an `if:`, and a wrong condition can skip
  CI for code with nothing to show it; a skipped job reads as neither pass
  nor fail.
* Bad, because the change-detection job runs on every push, and `build`
  and `release` are reusable-workflow calls whose skip has to be threaded
  through `needs`.

### D. `[skip ci]`

* Bad, because the global `prepare-commit-msg` hook, which this repository
  builds, writes the commit messages. A person would have to add the token
  by hand each time.
* Bad, because it skips the docs checks too, and a token left on a code
  commit skips its CI.

### E. Ignore only the Markdown no code reads, a docs workflow, and a filter check

* Good, because it keeps B's gain for every Markdown file except the one
  the tests read.
* Good, because a pattern that is too narrow fails toward a full run.
* Good, because D3 makes "no ignored path is read by code" a check that
  runs in CI and in `make preflight`.
* Bad, because the list names directories, and has to grow when a new
  kind of Markdown file appears.
* Bad, because D3 is a new script to own, and its parsing assumes the
  one-line flow style it checks for.

### F. As E, with `paths` and negation in `ci.yml`

* Good, because `'**'`, `'!**.md'`, `'README.md'` reads as "everything
  except Markdown, but README.md counts", and the reference documents that
  order rule.
* Bad, because if `**` does not match a path under a dot directory (not
  established), a change to `.github/workflows/ci.yml` or `.golangci.yml`
  alone would skip CI. That fails in the unsafe direction.
* Bad, because a wrong positive pattern skips CI for code, where E's
  wrong pattern only adds a run.

## More Information

* The implementation is
  [0019-PLAN-skip-full-ci-for-docs-only-pushes.md](0019-PLAN-skip-full-ci-for-docs-only-pushes.md).
* The CI being changed is described in `docs/architecture.md`, "Quality
  and delivery boundary", and `docs/guides/cicd-operations.md`.
* The README test's rule is 0009-MADR D4
  ([0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md](0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md)).
* Sequencing: `.github/workflows/ci.yml` is also changed by
  [0018-PLAN-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md](0018-PLAN-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md)
  (its `uses:` pins and the comment above `build`, `ci.yml:99-105` and
  `:120`). That pair is uncommitted at the time of writing. This record's
  PLAN lands its workflow change after 0018's `ci.yml` commit, and its
  line numbers are re-read then.
* Open for the owner: whether D3's check is wanted, or whether D1's
  comment and D4's AGENTS.md rule are enough. The recommendation is to
  keep D3, for the last decision driver.
