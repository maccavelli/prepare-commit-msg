---
status: in-progress
date: 2026-11-07
associated-madr: "0026-MADR-keep-five-stable-releases.md"
---
# Implement keeping only the eight newest stable releases

Associated MADR: [0026-MADR-keep-five-stable-releases.md](0026-MADR-keep-five-stable-releases.md)

*Revised 2026-11-07, before approval, to follow the MADR's revision: the
window is eight, the tool's default is `--keep 8`, and the expected first
pruning is recomputed from the 13 releases and 28 tags that exist since
`v1.10.4`. Test 17 guards the default against `--keep`. The filename keeps
its slug; the MADR says why.*

## Goal

* `tools/prune_releases.py` exists, behaves as MADR D1 to D5 set out, is
  tested in `tools/test_tools.py`, and is documented.
* The repository has exactly eight releases and eight `v*` tags. If no
  release lands first, they are `v1.7.0`, `v1.8.0`, `v1.9.0`, `v1.10.0`,
  `v1.10.1`, `v1.10.2`, `v1.10.3` and `v1.10.4`.
* The operations guide makes pruning a step of every release, and the
  README says how far back `--version` reaches.

## Scope

In scope:

* `tools/prune_releases.py` (new) and its tests in `tools/test_tools.py`.
* `tools/README.md`, `README.md` (`:164`, `:481`) and
  `docs/guides/cicd-operations.md` ("Controlled release").
* `docs/README.md`: this pair's index rows. The MADR's and the PLAN's
  status and execution record.
* Live, in Phase 3: deleting the releases and tags the dry run lists.
  Today that is the releases `v1.2.0` to `v1.6.0` (5) and 20 tags: those
  five releases' tags, plus the 15 tags with no release.

Out of scope:

* `go.mod` (MADR D7) and the Go module proxy.
* The release workflows and `scripts/`. The pruning runs by hand.
* Any release made after 2026-11-07. If one lands before Phase 3, Phase 3
  step 1 handles it.
* gobble-cli and the other repositories.
* The existing tools' file modes. They are `100644`, so their documented
  `tools/<name>.py` form needs `python3` in front. The new tool follows them,
  and its docs say `python3 tools/prune_releases.py`.

Each of the following needs the owner's word in the turn it happens: a
commit, naming `main`; a push, naming `main`; and the live `--apply`.

## Implementation Steps

### Phase 0: the record

1. Set the MADR to `status: accepted`, and this PLAN to
   `status: in-progress`.
2. Add both index rows to `docs/README.md` after the 0025 rows, before the
   0027 rows:
   * MADR: "Keep only the eight newest stable releases", accepted with the
     date;
   * PLAN: "Implement keeping only the eight newest stable releases",
     in progress.
3. Run `make check-records` and `make markdownlint`. Both exit 0.
4. Commit nothing yet. Phase 0's changes go in Phase 1's commit.

### Phase 1: the tool, its tests and the docs

#### Step 1. Write `tools/prune_releases.py`

Python 3 on the standard library. File mode `100644`, as the other tools
have. Begin with `#!/usr/bin/env python3` and a docstring that cites the
MADR, states the exit codes and gives the usage. Set
`REPO = Path(__file__).resolve().parent.parent`.

Its parts, each a function or constant a test can reach:

* **`GH = ("gh",)`.** The command prefix. `main` takes it as a parameter,
  as `hosts.main` takes `ssh`, so a test passes a fake.
* **`DEFAULT_KEEP = 8`.** The window when `--keep` is not given (MADR D1,
  D5).
* **`parse_version(tag)`.** Returns `(major, minor, patch, prerelease)`
  for `vMAJOR.MINOR.PATCH[-PRERELEASE]`, with no leading zeros, and `None`
  for anything else. The sort key for stable versions is
  `(major, minor, patch)`, compared as integers.
* **`repo_from_url(url)`.** Returns `OWNER/REPO` from
  `https://github.com/OWNER/REPO`, with or without `.git`, from
  `git@github.com:OWNER/REPO.git`, and from
  `ssh://git@github.com/OWNER/REPO.git`. Any other host or shape is an
  error.
* **`plan(releases, tags, latest, keep)`.** Pure. `releases` is a list of
  `Release(id, tag, draft, prerelease)`, `tags` a list of tag names,
  `latest` the latest release's tag, and `keep` the window. It returns a
  `Plan`:
  * `keep`: the kept releases, newest first;
  * `delete_releases`: each release to delete, with its tag;
  * `delete_tags`: the tags with no release to delete;
  * `report`: one line for each draft, each kept prerelease, each kept tag
    without a release, and each tag or release D4 leaves alone;
  * `refusal`: a reason, or `None`.

  The rules, in order:
  1. Releases whose tag `parse_version` rejects, or that are not `v*`, go
     to `report` (D4).
  2. Drafts go to `report` (D2).
  3. Stable releases are sorted. Fewer than `keep` is a refusal: "only N
     stable releases; --keep is K". The top `keep` are kept, and the rest
     go to `delete_releases` (D1).
  4. Each prerelease whose `(major, minor, patch)` is at or below the
     newest stable's goes to `delete_releases`. Each other one is kept and
     reported (D2).
  5. Each `v*` tag with no release that is a version: below the oldest kept
     stable release's version, it goes to `delete_tags`; otherwise it is
     reported. A tag that is not a version is reported (D3, D4).
  6. If `latest` is in `delete_releases`, that is a refusal: "the plan
     would delete the latest release, vX.Y.Z" (D5).

  A refusal leaves `delete_releases` and `delete_tags` empty.
* **`read_state(gh, repo)`.** It makes three calls, and any non-zero exit
  or JSON it cannot read raises `ToolError` with `gh`'s stderr:
  * `gh api --paginate --slurp repos/REPO/releases?per_page=100`;
  * `gh api repos/REPO/releases/latest`;
  * `gh api --paginate --slurp repos/REPO/git/matching-refs/tags/`, with
    `refs/tags/` cut from each `ref`.

  `--slurp` gives one JSON array of pages. It is in the installed `gh`
  (2.102.0).
* **`apply(gh, repo, plan, out)`.** For each release in `delete_releases`,
  oldest first:
  * `gh api --method DELETE repos/REPO/releases/ID`;
  * then `gh api --method DELETE repos/REPO/git/refs/tags/TAG`.

  Then the same tag call for each entry in `delete_tags`. It prints each
  deletion as it succeeds. At the first failure it prints the failure and
  the deletions so far, and raises `ToolError` (D5).
* **`read_back(gh, repo, plan)`.** Calls `read_state` again, and checks:
  every kept release is listed; no deleted release or tag is listed; the
  latest release is the one it was. Each mismatch is printed, and any one
  raises `ToolError`.
* **`main(argv=None, gh=GH, out=sys.stdout, err=sys.stderr)`.** It parses
  `--repo`, `--keep` (an integer of at least 1, default `DEFAULT_KEEP`) and
  `--apply`. An argument error exits 2. Without `--repo`, it runs
  `git -C REPO remote get-url origin`. Then it prints the plan, for
  example:

  ```text
  repository OWNER/REPO, keeping the 8 newest stable releases
  keep    release v1.10.4 (latest)
  delete  release v1.6.0 and its tag
  delete  tag v1.1.4 (no release)
  report  draft v2.0.0: drafts are never deleted
  5 releases and 20 tags to delete
  ```

  The exit codes:
  * a refusal, a `ToolError` or a bad argument: 2;
  * a dry run with something to delete: 1, ending with "dry run: nothing
    deleted; run again with --apply";
  * a dry run with nothing to delete: 0, with "nothing to prune";
  * `--apply`: `apply`, then `read_back`, then 0 with "pruned: N releases
    and M tags".

  With `--apply` and nothing to delete, it exits 0 and makes no DELETE
  call.

#### Step 2. Add the tests to `tools/test_tools.py`

Add `import prune_releases` beside the other imports, and a class
`PruneReleasesTest`.

**The fake `gh`** is a constant `FAKE_GH`, written to a temp file, like
`FAKE_SSH`. The test passes `gh=(sys.executable, script, state.json,
calls.jsonl)`. The fake:

* appends its arguments to `calls.jsonl`, one JSON array a line;
* reads `state.json`, which holds `releases`, `tags` and `latest`, and
  optionally `fail` (paths that exit 1) and `sticky` (paths whose DELETE
  succeeds and changes nothing);
* serves the releases as one page wrapped in a list, for `--slurp`; the
  latest as an object, or exits 1 with "Not Found" when there is none; and
  the tags as `[[{"ref": "refs/tags/TAG"}, …]]`;
* on a release DELETE, removes the release, writes the state, and exits 0;
* on a tag DELETE, exits 1 with "422 … immutable release" while a release
  with that tag exists, as GitHub does. Otherwise it removes the tag,
  writes the state, and exits 0.

**The fixture `LIVE`** is the repository's state on 2026-11-07: the
thirteen releases `v1.2.0` to `v1.10.4`, stable and not drafts, ids 1 to
13; the 28 tags `v0.1.0` to `v1.10.4`; latest `v1.10.4`.

Tests 1 to 6 call `plan(…, keep=8)`. Tests 7 to 14 and 16 go through
`main` with no `--keep`, so they run under `DEFAULT_KEEP`. Test 17 passes
`--keep`. Each test names what it guards:

| # | Test | Asserts |
| :--- | :--- | :--- |
| 1 | `test_versions_sort_as_numbers` | `plan(LIVE)` keeps `v1.10.4`, `v1.10.3`, `v1.10.2`, `v1.10.1`, `v1.10.0`, `v1.9.0`, `v1.8.0`, `v1.7.0`, in that order |
| 2 | `test_live_state_gives_the_expected_plan` | `LIVE` deletes the releases `v1.2.0` to `v1.6.0` and the 15 tags `v0.1.0` to `v1.1.4`, and reports nothing: 5 releases and 20 tags |
| 3 | `test_prerelease_below_newest_stable_is_deleted` | `v1.10.4-rc.1` and `v1.10.0-rc.2` are deleted; `v1.11.0-rc.1` is kept and reported |
| 4 | `test_draft_is_reported_never_deleted` | a draft `v1.2.5`, below the window, is in `report`, not in `delete_releases` |
| 5 | `test_tag_without_release_is_pruned_below_the_window_only` | `v1.6.5`, with no release, is deleted; `v1.10.5` is kept and reported; `v1.7.0`, the oldest kept, has its release and is kept. *Deviation 1 adds: `v1.8.5`, with no release and inside the window, is kept and reported* |
| 6 | `test_other_tags_and_releases_are_left_alone` | the tags `nightly`, `v1`, `v1.2` and `v01.2.3`, and a release tagged `release-1`, are reported, not deleted and not counted |
| 7 | `test_apply_deletes_each_release_before_its_tag` | through `main(["--repo", "o/r", "--apply"])` on `LIVE`: exit 0. In `calls.jsonl`, each release DELETE comes before its tag's, and the 15 tag-only DELETEs come last |
| 8 | `test_refuses_with_fewer_stable_releases_than_keep` | seven stable releases and no `--keep`, dry run and `--apply`: exit 2, "only 7 stable releases; --keep is 8" on stderr, no DELETE call |
| 9 | `test_refuses_to_delete_the_latest_release` | `LIVE` with latest `v1.2.0`, under `--apply`: exit 2, no DELETE call |
| 10 | `test_refuses_when_a_listing_fails` | each of the three listing paths in `fail` in turn, under `--apply`: exit 2, no DELETE call |
| 11 | `test_dry_run_deletes_nothing` | `LIVE`, dry run: exit 1 with "nothing deleted", no DELETE call. Eight stable releases and eight tags: exit 0 with "nothing to prune" |
| 12 | `test_apply_then_dry_run_is_clean` | `--apply` on `LIVE`: exit 0. The fake's state has the eight releases and eight tags. A second dry run exits 0 |
| 13 | `test_read_back_catches_a_deletion_that_did_not_hold` | one release DELETE path in `sticky`: `--apply` exits 2 and names that release |
| 14 | `test_failed_deletion_stops_and_a_rerun_finishes` | one tag DELETE path in `fail`: exit 2. No call follows the failed one. The output lists what was deleted. Remove the `fail`, rerun `--apply`: exit 0, eight releases and eight tags left |
| 15 | `test_repository_from_origin_url` | the four URL forms give `OWNER/REPO`; `https://gitlab.example/o/r` and `not-a-url` raise |
| 16 | `test_bad_arguments_exit_two` | `--keep 0`, `--keep x` and `--no-such-flag`: exit 2, with no `gh` call |
| 17 | `test_keep_overrides_the_default` | `LIVE`, dry run with `--keep 5`: exit 1, the releases `v1.2.0` to `v1.9.0` and 23 tags to delete. With `--keep 14`: exit 2, "only 13 stable releases; --keep is 14" |

#### Step 3. Prove each test fails on a planted break

Write `prune_mutants.py` in the session scratchpad. It copies `tools/` to a
temp directory and, for each break, replaces one exact snippet in the copy,
asserting it occurs once. It runs
`python3 -B -m unittest discover -s <copy> -p 'test_tools.py'`, records the
exit status and the failing test names, and then restores the copy. It
first runs the clean copy, which passes.

| Break | Expected to fail |
| :--- | :--- |
| stable releases sorted by tag text | 1, 2 |
| `keep` off by one (keep `keep + 1`) | 1, 2, 5, 12, 14, 17 |
| `DEFAULT_KEEP = 5` | 8, 11, 12, 14 |
| `--keep` ignored, so the default always applies | 17 |
| prereleases never deleted | 3 |
| a prerelease's core compared with `<`, not `<=`, so `v1.10.4-rc.1` is kept beside `v1.10.4` | 3 |
| drafts treated as stable | 4 |
| the tag-only bound at or below the oldest kept, not below | 5. *Deviation 1: equivalent; no input can tell it apart, because a tag equal to the oldest kept release has that release and never reaches the tag-only rule* |
| the tag-only bound taken from the newest kept release, not the oldest (Deviation 1) | 5 |
| non-versions treated as `v0.0.0` | 6 |
| tags deleted before releases | 7 (the fake refuses the tag), 12 |
| each of the three refusals removed, one at a time | 8 and 17; 9; 10 |
| the dry run calls `apply` | 11 |
| `read_back` skipped | 13 |
| `apply` carries on after a failure | 14 |
| the `scp`-form URL not parsed | 15 |
| `--keep 0` accepted | 16 |

Each break must make its tests fail and leave the clean run passing. A
break may fail more tests than it names; the named ones must fail. A
break that fails nothing means a missing test: add it, then rerun. Record
each break's result in the execution record.

#### Step 4. Update the docs (MADR D8)

* **`tools/README.md`.** A table row for `prune_releases.py`, with usage
  `python3 tools/prune_releases.py [--repo OWNER/REPO] [--keep N] [--apply]`.
  A section, "Pruning releases", saying:
  * the window, eight unless `--keep` says otherwise, and what is reported,
    never deleted;
  * that the dry run is the default;
  * the refusals;
  * that a failed `--apply` is finished by running it again.
* **`README.md:164`.** "`--version vX.Y.Z` installs another of the eight
  newest stable releases; older releases are deleted", and keep the rest
  of the bullet.
* **`README.md:481`.** The row becomes
  `prepare-commit-msg update --version vX.Y.Z --yes`, "Pin or roll back to
  one of the eight newest stable releases."
* **`docs/guides/cicd-operations.md`, "Controlled release".** After the
  checksum command, a step "After the release is verified":

  ```bash
  python3 tools/prune_releases.py            # read the plan; exit 1 means work to do
  python3 tools/prune_releases.py --apply
  git fetch --prune --prune-tags origin
  ```

  Say that it deletes the releases and tags below the eight newest stable
  releases, and that a deleted release's tag name cannot be reused. Say
  that `update` and the installers' default path follow the latest release,
  which is always kept, and that other clones drop the deleted tags with
  the same fetch. Cite the MADR by full filename.

#### Step 5. Run the gates

Capture each exit status before any filter, with output in the scratchpad:

* `make tools-test`, then `make preflight` (records, Markdown and
  `make verify`, which runs `tools-test` again);
* `make disclosure-check`;
* `git diff --stat`: only the files in Scope.

`README.md` is read by `readme_test.go`, so `make preflight`'s Go tests
cover the two README edits.

#### Step 6. Commit

On the owner's word naming `main`, stage the files in Scope, then run
`git commit --no-edit`. Then run `make preflight-hosts` on the commit: the
Linux host runs `make verify`, which runs the tool's tests. Report a host
that cannot be reached.

### Phase 2: push and CI

1. Run the disclosure guard over the outgoing commits, as `AGENTS.md`
   gives it. It exits 0.
2. Push on the owner's word naming `main`.
3. Wait for `ci.yml` and `docs.yml` on the pushed commit, with
   `gh run list --commit <sha>`. `tools/prune_releases.py` and
   `tools/test_tools.py` are not under `paths-ignore`, so `ci.yml` runs in
   full. Both pass. A failure stops the phase, and is recorded as a
   deviation.

### Phase 3: the first pruning, live

1. **Re-read the state.** In a scratchpad script, read the releases, the
   latest, the tags, the proxy's `@v/list`, the rulesets and the
   immutable-releases setting. Compare them with the MADR's "What exists"
   (read 2026-11-07).
   * If they are the same, the expected plan is Step 2's test 2: 5
     releases (`v1.2.0` to `v1.6.0`) and 20 tags.
   * If a release has landed (for example `v1.10.5`), work out the expected
     plan by MADR D1 to D4, show it to the owner with the difference, and
     go on only on their word. This is foreseen by the MADR, so it is not a
     deviation, but it is recorded.
   * Anything else that differs (a draft, a prerelease, a ruleset on tags,
     immutable releases off) stops the phase.
2. **Save the before-state** to the scratchpad, for the record:
   * the releases JSON, with each asset's name and digest;
   * the tag refs JSON, and `git ls-remote --tags origin`, which gives each
     tag's object ID;
   * the proxy's `@v/list`;
   * the attestation count for `v1.2.0`'s `prepare-commit-msg-linux-amd64`
     digest (`sha256:b18ac336…`): 2 on 2026-11-06 and on 2026-11-07.
3. **Dry run.** `python3 tools/prune_releases.py > dry-run.txt 2>&1`, with
   its status captured. Expect exit 1, and the header "keeping the 8 newest
   stable releases". A scratchpad script checks the plan lines against the
   expected plan, exactly. A mismatch stops the phase.
4. **Show the owner** the plan as printed, and wait for their word to
   apply.
5. **Apply.** `python3 tools/prune_releases.py --apply > apply.txt 2>&1`,
   with its status captured. Expect exit 0.
   * Exit 2 with a refusal: nothing was deleted. Stop and report.
   * Exit 2 after some deletions: stop, report what `apply.txt` lists as
     deleted, and record a deviation. Rerun `--apply` on the owner's word.
6. **Read back** with a scratchpad script, independent of the tool:
   * the releases are exactly the expected eight, each `immutable: true`;
   * the latest release is unchanged;
   * the `v*` tags are exactly the expected eight;
   * the rulesets and the immutable-releases setting are unchanged;
   * the proxy's `@v/list` is the 28 saved versions;
   * the `v1.2.0` attestation count is still 2. If it is not, record it in
     the MADR as a corrected fact; it does not stop the phase;
   * `python3 tools/prune_releases.py` exits 0, "nothing to prune".
7. **This clone.** Run `git fetch --prune --prune-tags origin`, then
   `git tag -l 'v*'`. It lists exactly the expected eight.

### Phase 4: the records

1. In this PLAN's execution record, put each phase's commands, statuses
   and output: the mutant table's results, the gates, the commit and CI
   run, the before-state, the dry run, the apply output and the read-back.
   Add a dated entry for every deviation.
2. Set this PLAN to `status: complete`. Set its index row to complete, and
   update the MADR's row if an amendment was made.
3. Run `make check-records`, `make markdownlint` and
   `make disclosure-check`.
4. Commit and push on the owner's word, naming `main`. These are records
   only, so `docs.yml` runs.

## Verification

* The tool's 17 tests pass, and each planted break fails the tests the
  mutant table names.
* `make preflight`, `make disclosure-check`, `make preflight-hosts`, the
  disclosure guard, and CI on the pushed commit.
* Phase 3's read-back, and a second dry run that exits 0.
* This clone's tags are the eight kept.

## Rollout and Rollback

* The rollout is Phase 3 step 5, one `--apply`, after the owner has seen
  the plan.
* The live deletions cannot be rolled back. A deleted release's assets are
  gone, and its tag name cannot be reused.
  * A tag that had no release can be pushed again from a clone that still
    has it, with the object ID saved in Phase 3 step 2.
  * A tag whose release was deleted cannot be pushed again.
* The tool and the docs roll back with a revert commit. Without the tool,
  the window is not kept, and nothing else changes.

## Execution record

The owner approved execution on 2026-11-07 ("committed, proceed"), after
committing this pair's revision to eight releases as `897480f`.

### Phase 0 (2026-11-07)

* The MADR is `status: accepted`, and this PLAN `status: in-progress`.
* `docs/README.md` has both rows, after 0025's and before 0027's.
* `make check-records`: exit 0. `make markdownlint`: exit 0.

### Phase 1 (2026-11-07)

* **Step 1.** `tools/prune_releases.py`, mode `100644`, as Step 1 sets it
  out. Two details beyond the step: the `Plan` carries the latest release's
  tag and the listed tag names, for `read_back` and `apply`; and a release
  whose tag ref does not exist is deleted without a tag DELETE, and printed
  as "(it has no tag)". Test 13 uses that case.
* **Step 2.** `PruneReleasesTest`, 17 tests, against `FAKE_GH`. Test 13
  plants the sticky release DELETE on `v1.2.0` with its tag ref removed from
  the fixture. A sticky release with its tag still present would stop at
  the tag DELETE, which the fake refuses, before the read-back, so the test
  would not separate the read-back from the refusal.
  `python3 -B -m unittest -v test_tools.PruneReleasesTest`: 17 ok.
* **Step 3.** `prune_mutants.py`, in the session scratchpad, on a temp
  copy of `tools/`, with the whole `test_tools.py` run for each break. The
  clean copy passes before and after. Each break, and the
  `PruneReleasesTest` numbers it failed (no other class failed):

  | Break | Failed |
  | :--- | :--- |
  | stable releases sorted by tag text | 1, 2, 3, 5, 6, 7, 9, 11, 12, 13, 14, 17 |
  | `keep` off by one | 1, 2, 5, 6, 7, 11, 12, 14, 17 |
  | `DEFAULT_KEEP = 5` | 7, 8, 11, 12, 14 |
  | `--keep` ignored | 17 |
  | prereleases never deleted | 3 |
  | a prerelease's core compared with `<` | 3 |
  | drafts treated as stable | 4 |
  | the tag-only bound at or below the oldest kept | none: equivalent (Deviation 1) |
  | the tag-only bound taken from the newest kept (Deviation 1) | 5 |
  | non-versions treated as `v0.0.0` | 6 |
  | tags deleted before releases | 7, 12, 13, 14 |
  | the fewer-than-`--keep` refusal removed | 8, 17 |
  | the latest-release refusal removed | 9 |
  | the failed-listing refusal removed (a failed listing read as empty) | 10 |
  | the dry run calls `apply` | 11, 17 |
  | `read_back` skipped | 13 |
  | `apply` carries on after a failure | 14 |
  | the `scp`-form URL not parsed | 15 |
  | `--keep 0` accepted | 16 |

* **Step 4.** `tools/README.md`: the table row and "Pruning releases".
  `README.md`: the installer bullet now reads "installs another of the
  eight newest stable releases; older releases are deleted", and the CLI
  table row is `update --version vX.Y.Z --yes`, "Pin or roll back to one of
  the eight newest stable releases." `docs/guides/cicd-operations.md`:
  the pruning step after the checksum command in "Controlled release",
  citing the MADR, and linking `tools/README.md#pruning-releases`.
* **Step 5.** `make tools-test`: exit 0, 42 tool tests and 34 script
  tests. `make preflight`: exit 0, "preflight passed".
  `make disclosure-check`: exit 0, 131 files, 0 findings. That check reads
  tracked files only, so the untracked `tools/prune_releases.py` was also
  searched for home paths, the local account name and the host name: no
  hits. `git status`: only the files in Scope.

#### Deviation 1 (2026-11-07): an equivalent break, and test 5's bound

* **Found.** In Step 3's first run, the break "the tag-only bound at or
  below the oldest kept, not below" failed no test (exit 0). It cannot fail
  one: a tag equal to the oldest kept release is that release's tag, so it
  never reaches the tag-only rule. A scratchpad probe
  (`prune_bound_probe.py`) then showed that test 5 as written also passed
  with the bound taken from the *newest* kept release, which would delete
  every unreleased tag inside the window. With a tag `v1.8.5` added to
  test 5, that break fails and the clean run passes.
* **Options put to the owner.** Swap the break and extend test 5
  (recommended); extend test 5 and keep the original row, marked
  equivalent; or record the break as equivalent and change no test.
* **Chosen** by the owner: extend test 5, and keep the original row,
  marked equivalent. Test 5 also asserts that `v1.8.5`, with no release
  and inside the window, is kept and reported. The mutant table keeps its
  row, annotated, and adds "the tag-only bound taken from the newest kept
  release", which fails test 5. Test 5's row in Step 2 is annotated. No
  file outside Scope changed, and the MADR's decisions and facts stand.
