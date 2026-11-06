---
status: in-progress
date: 2026-11-06
associated-madr: "0025-MADR-remove-ruleset-bypass-actors.md"
---
# Implement removing every bypass actor and making the docs describe GitHub as the source of truth

Associated MADR: [0025-MADR-remove-ruleset-bypass-actors.md](0025-MADR-remove-ruleset-bypass-actors.md)

*Revised 2026-11-06, before approval, for the owner's addition: the docs
are to reflect GitHub as the source of truth and the repository's current
facts (MADR D6 to D8).*

*Revised again 2026-11-06, before approval, for the owner's answers: Q1(a),
the `release` environment stops being created and is deleted; Q2(b), the
Apache License 2.0 (MADR D9).*

*Approved 2026-11-06 ("remove the creation rule … proceed"). Phases 0 to 2
run locally. Committing, pushing and Phase 3's live changes each wait for
the owner's word.*

## Goal

* Neither ruleset has a bypass actor, in `scripts/configure-github.sh` or
  on GitHub, and the tag ruleset is `deletion`, `non_fast_forward` and
  `update`.
* The `release` environment is neither created by the script nor present
  on GitHub.
* Every current doc states the repository's current facts and names this
  GitHub repository as the source of truth.
* The repository description no longer says "export".
* `LICENSE` holds the Apache License 2.0, and GitHub detects it.

## Scope

In scope:

* `scripts/configure-github.sh`:
  * the two ruleset bodies, and the comment above the tag body;
  * the `release` environment's read, plan entries and apply (`:129-153`,
    `:202`, `:215-216`, `:262-284`), which go (Q1).
* `LICENSE`, new (D9).
* `docs/guides/cicd-operations.md`: `:56-60`, `:112-122`, `:142-145`,
  `:199-213` and `:223-225` (MADR D3, D4, D6, D7).
* `docs/architecture.md`: the source-of-truth statement, `make verify` and
  `make verify-staged` (D6, D7).
* `README.md`: `:815`, `:793-801`, `:869-870` and `:874-876` (D7, D9).
* `docs/README.md`: the blank line at `:61`, and index rows for this pair.
* `docs/decisions/0019-PLAN-skip-full-ci-for-docs-only-pushes.md`: front
  matter `status` only.
* Live: rulesets 24842254 and 24842346, the `release` environment (Q1), and
  the repository description (D8). GitHub's license detection follows from
  the pushed `LICENSE`.

Out of scope:

* Any other record's text. Records are history, 0003-MADR's mirror model
  included.
* Code behaviour. Where a doc and the code disagree, the doc changes.
* `scripts/audit-github.sh`. It already reports `bypass_actors`, and
  reports the environment as `absent` when it is gone.
* The other settings the script applies: the Actions policy, the token
  permissions, and Dependabot.
* A `NOTICE` file, per-file license headers, and `LICENSE` in the release
  assets (MADR D9).
* Commits, pushes, tags and live changes. Each needs the owner's word in
  the turn it happens, naming `main` for a commit or push.

## Implementation Steps

### Phase 0: the record

1. Add index rows for this pair to `docs/README.md`.
2. Run `make check-records` and `make markdownlint`.

### Phase 1: the settings script

1. Edit `scripts/configure-github.sh` with a scratch Python script that
   asserts each anchor occurs exactly once before replacing it:
   * both ruleset bodies get `bypass_actors: []`;
   * the tag body gets `{type: "update"}` in place of `{type: "creation"}`;
   * the comment above the tag body is rewritten to say that there is no
     bypass actor, that the rules bind the owner, and why `creation` is
     absent. It cites
     `docs/decisions/0025-MADR-remove-ruleset-bypass-actors.md` D1 and D2;
   * the environment read, its two plan entries and its apply block are
     removed, so the script neither creates nor reconciles it (Q1).
2. Run `bash -n scripts/configure-github.sh`, then the dry run,
   `scripts/configure-github.sh > <scratch>/plan.json`, with its status
   captured. Check the plan with a scratch Python script:
   * both ruleset bodies have `bypass_actors == []`;
   * the tag rules are exactly `deletion`, `non_fast_forward` and `update`;
   * the main rules are unchanged;
   * no entry names `environments/release`.
3. Prove the check fails: run it against a scratch copy of the plan with an
   Admin bypass entry put back, and against one with the environment entry
   put back. Each must fail.

### Phase 2: the docs

1. Make every edit with a scratch Python script that asserts each anchor
   occurs exactly once before replacing it:
   * `cicd-operations.md`:
     * `:56-60`: the paragraph is kept for an old clone, but the step
       "Before pulling this change, run `make hooks-uninstall`" goes and is
       replaced by a note that the target no longer exists, so the clone is
       restored by hand as below;
     * `:115`: GitHub-owned actions, plus go-selfupdate-lib's
       `build-selfupdate-release.yml` and `publish-selfupdate-release.yml`
       by pattern, with full-SHA pinning;
     * `:117`: the `release` environment's line goes (Q1);
     * `:118-122`: D1 and D2;
     * `:142-145`: the required-reviewers paragraph goes, since it is about
       the environment's branch policy (Q1);
     * `:199-213`: D3;
     * `:223-225`: D6.
   * `docs/architecture.md`:
     * a sentence that this GitHub repository is the source of truth:
       work lands on its `main`, and releases are tagged there;
     * `make verify`'s list adds shell-script syntax and the maintainer
       tools' tests;
     * `make pre-add-check` is named as the staged-snapshot check, with
       `verify-staged` as its alias.
   * `README.md`:
     * `:815`: `make verify`'s list as the Makefile has it;
     * `:793-801`: the CLI block becomes `printUsage`'s six usage lines,
       with `prepare-commit-msg` in place of `%s`;
     * `:869-870`: cite 0013-MADR for the workspace and 0017-MADR for the
       hook policy that replaced it;
     * `:874-876`: in place of the fleet sentence, one sentence saying the
       project is licensed under the Apache License, Version 2.0, with a
       relative link to the root `LICENSE` (D9).
   * `docs/README.md`: delete the blank line at `:61`.
   * `0019-PLAN` front matter: `status: superseded`.
   * `LICENSE`: ~~fetch `gh api licenses/apache-2.0 --jq .body` again~~
     fetch `gh api licenses/apache-2.0`, take the JSON `body` field
     (Deviation 1), assert its SHA-256 is D9's, and write it byte for byte to
     `LICENSE`.
2. Check the result with a scratch Python script:
   * no current doc (`README.md`, `AGENTS.md`, `docs/README.md`,
     `docs/architecture.md`, `docs/guides/*.md`, `tools/README.md`) contains
     "mirror" as a description of the repository, "next export",
     "published export", "authoritative private source", "run `make
     hooks-uninstall`" or "split Go repository fleet";
   * the README's CLI block equals `printUsage`'s lines, as read from
     `main.go`;
   * the README's `make verify` list equals the Makefile's `verify`
     prerequisites;
   * `docs/README.md`'s records table has no blank line between its header
     and its last row;
   * `LICENSE`'s SHA-256 is D9's, and the README's License section names
     the Apache License 2.0 and links `LICENSE`.
3. Prove the check fails on a scratch copy of the tree with each kind of
   finding planted once: a phrase put back, a CLI line changed, a blank
   line in the table, and one changed byte in `LICENSE`. Each must fail.
4. Run `make preflight`. It runs `check-records`, `markdownlint` and
   `verify`, which includes `workflow-lint` (`bash -n` on the scripts, and
   actionlint), `readme_test.go` and `documentation_test.go`. Then run
   `make disclosure-check`.
5. Commit Phases 0 to 2 with `git commit --no-edit`, on the owner's word
   naming `main`.
6. Run `make preflight-hosts`, because the change includes a script.

### Phase 3: apply and confirm

1. Run the disclosure guard over the outgoing commits.
2. Push, on the owner's word.
3. Wait for CI on the pushed commit to pass. `README.md` and the guide are
   not under `paths-ignore`, so `ci.yml` runs in full.
4. Save a pre-apply audit with `scripts/audit-github.sh` to the scratchpad.
5. On the owner's word, apply the settings and capture each exit status:
   * `scripts/configure-github.sh --apply`;
   * `gh api --method DELETE repos/maccavelli/prepare-commit-msg/environments/release`,
     because the script no longer manages it;
   * `gh repo edit maccavelli/prepare-commit-msg --description "AI commit-message git hook"`.
6. Read back and check:
   * `GET …/rulesets/24842254` and `…/24842346`: `bypass_actors` is `[]`,
     `current_user_can_bypass` is `never`, the rules are as in Phase 1
     step 2, and the ids are unchanged;
   * `GET …/environments/release`: 404;
   * `GET …`: the description, and `license.spdx_id` `Apache-2.0`.
7. Diff the pre- and post-apply audits. Only the rulesets and the
   environment may differ.
8. Record the results in this PLAN, mark it `complete`, update the index
   row, and commit and push on the owner's word.

## Verification

* The settings check of Phase 1 step 2 and the docs check of Phase 2
  step 2, each seen to fail on broken copies.
* `make preflight`, `make disclosure-check` and `make preflight-hosts`.
* The API readback and audit diff of Phase 3 steps 6 and 7.
* At the next release, a deferred check: the tag push prints no "Bypassed
  rule violations", and its rule suite has the result `pass` with no
  `creation` evaluation. It is recorded in that release's PLAN.
* Not probed: refusal of a delete, a move, or a force-push. The MADR's
  Confirmation gives the reason.

## Rollout and Rollback

* The rollout is Phase 3 step 5.
* Rolling back:
  * the rulesets and the environment: revert the Phase 1 commit, push it,
    and run `--apply` again. That restores the Admin bypass, the `creation`
    rule and the `release` environment with its `main` branch policy;
  * the description: `gh repo edit --description` with the old text;
  * the docs: a revert commit;
  * `LICENSE`: a revert removes the file from later commits. It does not
    withdraw the license already granted on commits pushed with it: the
    license's section 2 makes the grant perpetual and irrevocable.
* Until the apply, the live settings keep both bypasses. Releases are
  unaffected in either state.
* Between the push and the apply, the script and the live settings differ.
  Nothing else reads them.

## Execution record

### Phase 0 (2026-11-06)

* The MADR's status is `accepted` and this PLAN's is `in-progress`. The
  owner chose option A: "remove the creation rule, it is not needed since
  this is the source repo. proceed."
* `docs/README.md` has index rows for both records.
* `make check-records` and `make markdownlint` exited 0.

### Phase 1 (2026-11-06)

* `scripts/configure-github.sh` was edited by a scratch script that
  asserted each of its six anchors once (14 insertions, 73 deletions):
  * both rulesets have `bypass_actors: []`;
  * the tag rules are `deletion`, `non_fast_forward` and `update`;
  * the `release` environment's read, plan entries and apply are gone;
  * new comments cite MADR D1, D2 and Q1.
* `bash -n scripts/configure-github.sh` exited 0.
* The dry run exited 0 with nothing on stderr, so `remote_ready` is true.
  Its mutations are the Actions policy, selected actions, workflow
  permissions, `automated-security-fixes`, and the two ruleset upserts. No
  environment is among them.
* The plan check exited 0 (`clean`). Against broken copies of the plan it
  exited 1 each time, naming the fault:
  * an Admin bypass entry on the tag ruleset;
  * an Admin bypass entry on the main ruleset;
  * `creation` put back;
  * `update` removed;
  * an `environments/release` entry put back.

### Deviation 1 (2026-11-06): D9's license checksum included a newline that `gh` appends

* **Found.** Before writing `LICENSE`, the agent read the API's JSON. Its
  `body` is 11,357 bytes, ends in one newline, has 201 lines, and has the
  SHA-256 `c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4`.
  The figure in D9, 202 lines and `c95bae1d…`, came from
  `gh api … --jq .body`, which adds a newline of its own.
* **Decision (agent, within the decision as written).** D9 asks for the text
  "exactly as GitHub's `GET /licenses/apache-2.0` serves it". That is the
  `body` field, without the extra blank line. D9's figure is corrected in the
  MADR, and Phase 2 takes the `body` field from the JSON.
* **Scope.** No file is added.

### Phase 2 (2026-11-06)

* One scratch script made every edit. It asserted each anchor once, built
  the README's CLI block from `printUsage` in `main.go`, and wrote
  `LICENSE` from the API's `body` after asserting its SHA-256:
  * `docs/guides/cicd-operations.md`: five edits (`hooks-uninstall`; the
    converged state; the environment and required-reviewers lines; failed-release
    recovery; the source of truth). The apply-path paragraph now names what
    the script checks, which is `ci.yml`'s commit and blob.
  * `docs/architecture.md`: three edits (the source of truth and the
    license; `make verify` and `pre-add-check`; `LICENSE` in the tree).
  * `README.md`: four edits (the CLI block; `make verify`'s list; the 0013
    and 0017 pointers; the License section).
  * `docs/README.md`: the blank line in the records table is gone.
  * `0019-PLAN`: `status: superseded`.
  * `LICENSE`: 11,357 bytes, SHA-256 `c71d239d…` (D9).
* The docs check exited 0 (`clean`). On scratch copies, each planted fault
  made it exit 1 with the matching finding:
  * the mirror sentence put back in the guide;
  * the fleet sentence in place of the License text, which gave two
    findings;
  * a changed CLI line;
  * `tools-test` dropped from the `make verify` list;
  * a blank line back in the table;
  * one flipped byte in `LICENSE`.
* `make preflight` exited 0, ending "preflight passed":
  * `check-records` and `markdownlint` passed;
  * `make verify` passed. Its tests passed, total coverage is 83.9%
    (minimum 80.0%), `verify-scripts.sh` covered the edited script, the tool
    tests ran 25 and 34 tests, OK, and all six cross-builds were made.
* `make disclosure-check`: 124 files scanned, 0 findings.
* Not yet run: step 5, the commit, which waits for the owner's word naming
  `main`, then step 6, `make preflight-hosts`, which tests the committed
  `HEAD`.
