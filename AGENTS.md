# AGENTS.md

Instructions for AI coding agents working in this repository. All agents read
this file. A repository-local `CLAUDE.md` / `.claude/rules/` / `.grok/rules/` /
`.opencode/rules.md` / `.codex/rules/` / `.kilo/agent/` wins only where it is
more specific than this file.

`prepare-commit-msg` is a Go CLI and Git `prepare-commit-msg` hook that
turns staged diffs into Conventional Commit messages through configured local
or remote LLM providers. `docs/architecture.md` describes the current system,
`docs/README.md` is the documentation and records entry point, and
`docs/guides/cicd-operations.md` gives the validation and release sequence.

## Dependencies

No dependency may be added without a MADR in this repository that names it.
`go.mod` and `go.sum` change in the same commit as the first import and are
cleaned in the commit that removes the last. The dependency MADR owns the exact
verification commands and version constraints.

## MADR and PLAN before mutating work

**Whenever the user asks for an MADR and a plan, load the
`madr-and-plan-writing` skill first** and follow it for authoring, naming and
review. This applies both to writing a fresh pair and to amending an existing
one.

The name is exact — it is the `name:` field of the skill. A mistyped call
returns `Unknown skill`, and an agent that proceeds without the skill writes
something shaped like a MADR while missing the required headings and the
directory rule below. Verify against the filesystem rather than memory:

```bash
ls -d ~/.claude/skills/*madr* ~/.grok/skills/*madr* ~/.codex/skills/*madr* ~/.agents/skills/*madr* 2>/dev/null
grep '^name:' ~/.claude/skills/*madr*/SKILL.md ~/.grok/skills/*madr*/SKILL.md ~/.codex/skills/*madr*/SKILL.md ~/.agents/skills/*madr*/SKILL.md 2>/dev/null
```

**Read-only investigation is allowed with no pair.** Reading, searching,
`git log` / `git show` / `git diff`, and existing tests or diagnostics that
do not write the tree do not need a MADR.

**Mutating work is not.** Before the first write, name the
`docs/decisions/NNNN-MADR-*` / `docs/decisions/NNNN-PLAN-*` pair being
executed, or stop and write one.

Mutating means: creating, editing, or deleting files; staging or committing
(except the bootstrap exception below); dependency or lockfile changes;
CI / config / hook changes; builds or installers that write the tree,
`$HOME`, or a live service; generating committed artifacts.

Order:

1. Investigate (read-only).
2. Write or amend the MADR (`status: proposed` unless the owner already
   decided). Present it. Do not implement.
3. Write or amend the PLAN. Present it.
4. Mutate **only after** the owner explicitly approves execution
   (`proceed`, `execute the plan`, `do phase N`). Stay inside that PLAN.
5. Anything discovered mid-execution that is out of scope waits: amend the
   pair, re-approve, then continue. Completing a phase is not permission to
   invent the next unwritten one.

Follow-up vs greenfield:

- **Same topic** (debug, leftover phase, bug found in that plan's live
  run): amend that number. Add a PLAN phase or an amendment in the MADR. Do
  not silently rewrite historical rationale.
- **Greenfield**: next unused `NNNN`, new MADR, new PLAN, same slug. No
  mutation until that PLAN is approved.

Bootstrap exception: authoring `docs/decisions/NNNN-MADR-*`,
`docs/decisions/NNNN-PLAN-*`, this file, and the per-agent pointers
(`.claude/rules/`, `.grok/rules/`, `.opencode/rules.md`, `.codex/rules/`,
`.kilo/agent/`) does not require a *prior* pair. Putting source, tests, CI,
or product config in that same commit is a violation.

`git push` and tags still need an explicit ask in the same turn.

When the user asks where a document belongs, load `documentation-writing`.

## Records

```text
docs/decisions/NNNN-MADR-short-slug.md
docs/decisions/NNNN-PLAN-short-slug.md
docs/reports/NNNN-REPORT-short-slug.md
docs/reports/NNNN-GATES-short-slug.md
```

- `NNNN` is a zero-padded 4-digit number, one sequence across
  `docs/decisions/` and `docs/reports/`. A MADR and its PLAN share the same
  number and the same slug.
- **Next number** comes from `python3 scripts/check_records.py --next`.
  Never reuse a number, never renumber an existing record, never leave a gap
  deliberately.
- `make check-records` validates naming, pairing, placement, and every
  relative link in the records and the unnumbered docs. If Make is absent,
  run `python3 scripts/check_records.py --check-all`.
- Cite records by full filename, never by number alone. Cite another
  repository's record by repository and filename; a relative link cannot reach
  it.
- `docs/README.md` indexes every record by hand. Update it in the same change.

## Pre-add checks

- Before every commit, run `make check-records` and `make markdownlint`.
- With Go files or `go.mod`/`go.sum` staged, run `make pre-add-check`
  (`FILES=...` to name files); the fleet agent commit gate runs it too.
- Run any additional product check required by the active PLAN phase.
- A file that fails a required check is not committed.

## Commands and tooling

Standing rule, from the owner:

- Do not script or write tooling in bash, and do not drive multi-part work
  with chained one-liners. Multi-part commands, and any work whose
  evaluation feeds a decision, run as Python.
- Repository tooling is idiomatic, modular Python on the standard library,
  with tests beside it, like `scripts/` and `tools/`.
- Prefer the native file tools — a whole-file read, a content search, a
  pattern listing — over `grep` / `sed` / `cut` / `awk` one-liners: filtered
  slices can truncate, and an edit must land on what the file actually says.
- When a decision depends on file content, read the file whole and evaluate
  it against a scratch temp directory; never conclude from an excerpt.
- A command whose exit status feeds a decision is captured before any filter,
  and long output goes to a scratch file.

## Before a push

A change to Markdown under `ci.yml`'s `paths-ignore` runs only `docs.yml` in
CI: the records, Markdown and disclosure gates, never the Go contract. A
change to `README.md`, to the operations guide, to any other file, and every
tag run `ci.yml` in full, because GitHub does not evaluate path filters for
a tag push
(`docs/decisions/0022-MADR-align-docs-ci-and-disclosure-gates.md`). A test,
`//go:embed` or tool that reads a Markdown file keeps that file out of the
ignore list in the same commit, and `scripts/check_ci_filters.py` checks it
in `make workflow-lint`; the local readers stay `readme_test.go` (the root
README) and `documentation_test.go` (the operations guide). Markdown itself
is screened for home paths by `make disclosure-check`, in both workflows;
private identity patterns stay with the owner's pre-push guard below.

The repository installs no Git hooks
(`docs/decisions/0017-MADR-replace-git-hooks-with-make-preflight-and-host-tests.md`).
Before asking the owner to push, in this order:

1. `make preflight`, always. It runs CI's `go` job's targets locally.
2. `make preflight-hosts`, when the change can behave differently per OS: Go
   sources, `go.mod`/`go.sum`, build tags, paths, `os/exec`, permissions, the
   `Makefile`, scripts or workflows. Changes to docs and records only are
   exempt. It tests the committed `HEAD`. A host that cannot be reached is
   reported in the handoff, never skipped silently.
3. The disclosure-guard command under "Identifiers".

## Tools

`tools/` holds the tools a maintainer or an agent runs by hand
([`tools/README.md`](tools/README.md); 0016-MADR); `scripts/` holds what
`make` and CI run. Each tool is Python 3 on the standard library, with a test
in `tools/test_tools.py`; `make tools-test` runs them, and `make verify` runs
it. A release's live retry check runs the hook against
`tools/fakeprovider.py` (0015-PLAN's Phase 5 step 3, with Deviation 5).

## Identifiers

Nothing committed carries a hostname, account name, org-internal path, or a
real-machine absolute path. Use placeholders (`<user>`, `/home/<user>/...`).

- When recording an identifier scan in a record or a commit, describe what was
  scanned for ("the local account name", "the hostname domain"); never quote
  it.
- Pushes to GitHub pass through a global pre-push disclosure guard. Before
  asking the owner to push, run the guard itself over the outgoing commits:

  ```bash
  echo "refs/heads/main $(git rev-parse HEAD) refs/heads/main $(git rev-parse origin/main)" |
    python3 ~/.global-git-hooks/github-disclosure.py pre-push origin "$(git remote get-url origin)"
  ```

- Never bypass it with `--no-verify`.

## Commits

`git commit --no-edit`. Never pass `-m` / `--message` / `-F`. The global
`prepare-commit-msg` hook writes the message from the staged diff.
