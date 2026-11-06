---
status: accepted
date: 2026-11-06
decision-makers: Project Owner
consulted: the repository's rulesets, rule suites, environments, Actions policy and repository metadata (GitHub API, read 2026-11-06); `scripts/configure-github.sh`; the Makefile; `main.go`; the current docs (`README.md`, `docs/README.md`, `docs/architecture.md`, `docs/guides/cicd-operations.md`, `tools/README.md`, `AGENTS.md`); 0003-PLAN and 0012-PLAN, which configured the bypass
informed: the release maintainer
---
# Remove every bypass actor, and make the docs describe GitHub as the source of truth

*Revised 2026-11-06, before approval. The owner added: GitHub is now the
source of truth (the repository no longer exists on the private host it was
exported from), and the docs are to reflect that and the repository's
current facts. D6 to D8 and Q1 and Q2 are new.*

*Revised again 2026-11-06, before approval. The owner answered: Q1 "stop
and delete"; Q2 "apache-2.0 license approved". D9 records the license.*

*Accepted 2026-11-06. The owner chose option A: "remove the creation rule,
it is not needed since this is the source repo. proceed."*

## Context and Problem Statement

### The bypass

The `v1.10.3` tag push reported "Bypassed rule violations for
refs/tags/v1.10.3: Cannot create ref due to creations being restricted"
([0024-PLAN-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md](0024-PLAN-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md),
Phase 5). The owner's instruction (2026-11-06): the bypass "should not be
active or configured".

What is configured **[read, GitHub API, 2026-11-06]**:

| id | name | target | rules | bypass actors |
| :--- | :--- | :--- | :--- | :--- |
| 24842254 | `prepare-commit-msg-main` | branch `~DEFAULT_BRANCH` | `deletion`, `non_fast_forward` | `RepositoryRole` 5 (Admin), `always` |
| 24842346 | `prepare-commit-msg-release-tags` | tag `refs/tags/v*` | `creation`, `deletion`, `non_fast_forward` | `RepositoryRole` 5 (Admin), `always` |

Both are `active`, and `current_user_can_bypass` is `always` on both.

Where the bypass comes from **[read]**:

* `scripts/configure-github.sh` writes both bodies, `bypass_actors`
  included (`:156-191`), and `--apply` upserts them by name. A change made
  only on GitHub would be reverted by the next `--apply`.
* 0003-PLAN chose to restrict `v*` creation "to administrators and the
  GitHub Actions integration". 0012-PLAN Deviation D3 dropped the Actions
  entry, because a personal-account repository cannot name it (HTTP 422),
  which left administrators alone.
* `docs/guides/cicd-operations.md:118-122` still describes the Actions
  integration as a bypass actor. That is stale since 0012-PLAN D3.
* The same list gives the reason for the `main` bypass: "repository
  administrators able to publish mirror snapshots".

How the bypass is used **[read, rule suites for the last month]**:

* Every release tag push (`v1.8.0`, `v1.9.0`, `v1.10.0`, `v1.10.1`,
  `v1.10.2`, `v1.10.3`) has the result `bypass`. Each failed `creation`
  and passed `deletion` and `non_fast_forward`.
* All 24 pushes to `refs/heads/main` have the result `pass`. The `main`
  bypass has not been used.
* The repository has one collaborator, the owner, with the `admin` role.
  No workflow creates, moves or deletes a tag. The publish job
  (`ci.yml:128-129`) holds `contents: write`, and go-selfupdate-lib's
  publish workflow runs `gh release create --verify-tag`, which requires the
  tag to exist already.

The constraint that shapes the decision: a rule with **no** bypass actor
binds everyone. `creation` with an empty bypass list refuses every `v*`
tag, so removing the bypass alone stops every release. The documented
failed-release recovery (`cicd-operations.md:199-211`) deletes the remote
tag, and that also works only through the bypass.

### A second bypass: the `release` environment

`scripts/configure-github.sh` also creates a `release` environment with
one branch policy, `main` **[read]**. GitHub reports it with
`can_admins_bypass: true`. Nothing uses it:

* `ci.yml` names no environment;
* the pinned publish workflow (`74ab5dd3`, v1.24.0) names one only in its
  `sign` job, `selfupdate-signing`, which runs only when `sign` is true, and
  this repository does not set `sign`;
* the environment has no deployments.

### GitHub is the source of truth

The owner (2026-11-06): the repository has moved. **GitHub is now the
source of truth.** The repository no longer exists on the private host it
was exported from, so there is no mirror and no export. 0003-MADR set the
public repository up as a one-way mirror of that private source. The
current docs and settings still carry that model, and the sweep below found
other statements that no longer match the repository.

What does not match **[read, 2026-11-06; each checked against the code or
the API]**:

| Where | It says | The fact |
| :--- | :--- | :--- |
| GitHub repository description | "AI commit-message git hook (published export)" | nothing is exported |
| `cicd-operations.md:118-119` | the `main` bypass lets administrators "publish mirror snapshots" | there is no mirror |
| `cicd-operations.md:223-225` | "a one-way mirror … the authoritative private source before the next export" | GitHub is the source |
| `cicd-operations.md:56-60` | "run `make hooks-uninstall`" before pulling | the Makefile has no such target |
| `cicd-operations.md:115` | "GitHub-authored and repository-local actions only" | the live policy also allows go-selfupdate-lib's build and publish workflows by pattern |
| `cicd-operations.md:117` | "a `release` environment restricted to the `main` branch" | true, but no job uses it (above) |
| `cicd-operations.md:142-145` | reviewers may be added, preserving "the `main` deployment branch policy" | that policy is the `release` environment's |
| `docs/architecture.md` | nothing on where the source lives; `make verify-staged` "remains an explicit staged-snapshot diagnostic"; `make verify` covers "workflow syntax" | `verify-staged` is an alias of `pre-add-check`; `make verify` also checks shell scripts (`bash -n`, through `workflow-lint`) and runs `tools-test` |
| `README.md:815` | `make verify` runs "mod-check, fmt-check, lint, vet, test, coverage, vuln, workflow-lint, build-all" | it also runs `tools` and `tools-test` |
| `README.md:793-801` | the CLI reference's descriptions | they differ from `printUsage` (`main.go:47-55`): "show version", "show this help", "print the build identity alone" |
| `README.md:869-870` | read 0013-MADR for "the workspace and local-hook policy" | 0017-MADR replaced 0013's hook policy |
| `README.md:874-876` | "## License — Part of the split Go repository fleet." | no `LICENSE` file exists, and GitHub detects no license |
| `docs/README.md:61` | a blank line inside the records table | it ends the table; rows 0022 onward render as text |
| `0019-PLAN` front matter | `status: in-progress` | the index records it superseded by 0022-PLAN (2026-10-25) |

The remaining current docs match the code **[read]**: `AGENTS.md`,
`tools/README.md`, the per-agent pointers, the README's provider, flag,
configuration, hook and update sections, and the guide's verification,
release and rollback steps. The decision records are history and keep what
they say. That includes 0003-MADR's mirror model.

## Decision Drivers

* No actor, the owner included, is exempt from any ruleset or environment
  rule. A push the rules would refuse is refused, not reported as bypassed.
* Releases keep working: the owner pushes an annotated `v*` tag, and the
  tag run publishes it.
* `main` cannot be deleted or force-pushed, and a published `v*` tag cannot
  be deleted or moved.
* `scripts/configure-github.sh` stays the source of truth for settings. A
  re-apply converges on the state this MADR chooses.
* The current docs state current facts only, each checkable against the
  code or the API, and name GitHub as the source of truth.

## Considered Options

* A: empty both bypass lists; in the tag ruleset, replace `creation` with
  `update`; recover a failed release by burning its tag
* B: empty both bypass lists, keep `creation`, and disable the tag ruleset
  around each release
* C: empty both bypass lists and change nothing else
* D: remove the bypass on GitHub only, and leave the script as it is

The docs sweep (D6 to D8) follows from the owner's instruction and is
common to every option.

## Decision Outcome

Chosen option: "A", because it is the only option in which no actor has a
bypass and releases still work. The owner chose it (2026-11-06): the
`creation` rule "is not needed since this is the source repo".

* **D1, no bypass actors.** Both rulesets send `bypass_actors: []`, in
  `scripts/configure-github.sh` and live.
* **D2, the tag rules.** `prepare-commit-msg-release-tags` keeps `deletion`
  and `non_fast_forward`, drops `creation`, and adds `update`.
  * Without a bypass actor, `creation` would refuse every release tag.
  * `update` refuses moving an existing `v*` tag. Today only the owner's
    bypass makes that possible, and `non_fast_forward` does not refuse a
    move to a descendant commit.
  * `prepare-commit-msg-main` keeps its rules, `deletion` and
    `non_fast_forward`.
* **D3, failed-release recovery.** Once pushed, a `v*` tag is permanent.
  If its run fails before publishing, the fix lands on `main` and the next
  patch version is tagged. The failed tag stays, unreleased.
  `cicd-operations.md`'s "Delete the remote and local failed tag" step goes.
  * Deleting that tag would need a ruleset edit, which is a bypass by
    another name.
* **D4, the guide's settings list.** `cicd-operations.md`'s converged state
  states D1, D2, and the selected-actions patterns as the live policy has
  them. It loses the stale Actions-integration and mirror-snapshot wording,
  the `release` environment's line, and the required-reviewers paragraph
  about that environment's branch policy (Q1).
* **D5, applying.** The settings change is applied with
  `scripts/configure-github.sh --apply`, after the script change is on
  `origin/main` (the script refuses otherwise). It is read back with
  `scripts/audit-github.sh` and the API.
* **D6, GitHub as the source of truth, in the docs.** `docs/architecture.md`
  says that this GitHub repository is the source of truth: work lands on its
  `main`, and releases are tagged there. `cicd-operations.md:223-225` says
  the same, with accepted dependency updates landing on `main`. No current
  doc names the old private host or describes a mirror or an export.
* **D7, the other current-fact corrections.** Each row of the table above
  is corrected to the fact in its last column. Where a statement describes
  the code, the code wins, and the code itself does not change: for example,
  the README's CLI reference becomes `printUsage`'s text. The `hooks-uninstall`
  paragraph is kept for an old clone, without the step that names the removed
  target. 0019-PLAN's front matter becomes `superseded`, matching the index.
* **D8, the repository description.** The GitHub description drops
  "(published export)" and becomes "AI commit-message git hook". It is a
  live setting, changed with `gh repo edit`, which the owner authorizes when
  it is run.
* **D9, the license (Q2, the owner's choice).** The repository is licensed
  under the Apache License, Version 2.0.
  * `LICENSE` at the root holds the license text exactly as GitHub's
    `GET /licenses/apache-2.0` serves it in its `body` field: 201 lines,
    11,357 bytes, SHA-256
    `c71d239df91726fc519c6eb72d318ec65820627232b2f796219e87dcf35d0ab4`.
    *(Corrected 2026-11-06, during execution: the figure first recorded here,
    202 lines and `c95bae1d…`, was of that text plus the newline
    `gh api --jq` appends. 0025-PLAN, Deviation 1.)*
    The appendix, with its `[yyyy] [name of copyright owner]` template,
    stays as part of that text.
  * The README's "License" section says the project is licensed under the
    Apache License 2.0, and links `LICENSE`.
  * No `NOTICE` file, no per-file headers, and no change to the release
    assets. The license requires none of these.
  * The dependencies allow it **[read]**: go-llmprovider-sdk and
    go-selfupdate-lib are Apache-2.0 (GitHub's license detection), and
    `golang.org/x/mod`, `sys`, `term` and `tools` carry the BSD-style
    licence text.
  * This is recorded here, not in a MADR of its own, because the owner
    decided it in answer to Q2, and the README change it needs is already
    in this decision's PLAN.

This decision supersedes, for the bypass lists and the tag `creation` rule
only, 0003-PLAN's tag-ruleset step and 0012-PLAN's Deviation D3. It also
records that the private-source and public-mirror model 0003-MADR assumed
has ended. Those records are left as written, as history.

### Questions for the owner, answered 2026-11-06

* **Q1, the unused `release` environment.** Its admin bypass is a bypass
  too.
  * **(a), chosen ("stop and delete"):** stop creating it in
    `configure-github.sh` and delete it live. Nothing uses it, and it can be
    recreated with a protection that binds everyone when a job needs it.
  * (b), not chosen: keep it, and set `can_admins_bypass: false`.
  * (c), not chosen: leave it, and document it as unused.
* **Q2, the README's "License" section.** No license is published, which
  makes this a decision, not a correction.
  * (a), not chosen: state "No license file is published".
  * **(b), chosen ("apache-2.0 license approved"):** the Apache License 2.0,
    as D9.

### Consequences

* Good, because no push is reported as a bypass. GitHub evaluates every
  rule with nobody exempt.
* Good, because `main` and published `v*` tags are now protected from the
  owner's own mistakes: no force-push, no deletion, no moved tag.
* Good, because the configuration script, the live settings and the docs
  describe one state again, and that state names GitHub as the source of
  truth.
* Bad, because any actor with write access can create a `v*` tag. Today
  that means the owner, and a token with `contents: write`: the publish
  job's `GITHUB_TOKEN`.
  * Neutral, because a tag pushed with a `GITHUB_TOKEN` starts no workflow
    run, so it cannot start a release.
* Bad, because a mistaken or failed tag cannot be removed. Its version
  number is burned, and `proxy.golang.org` may list it as a module version.
  * Neutral, because the proxy keeps a version it has already fetched even
    after its tag is deleted, so deletion never fully withdrew one.
* Good, because the public code has a license, which GitHub detects and
  shows.
  * Neutral, because releases up to `v1.10.3` were published without a
    `LICENSE` file. This change does not alter their assets.
* Bad, because the sweep is a snapshot. A later change can make a doc stale
  again, and only `readme_test.go` and `documentation_test.go` guard parts
  of two docs.
* Neutral, because undoing the settings change is a script change plus
  `--apply`.

### Confirmation

* `GET /repos/…/rulesets/<id>` for both rulesets shows `bypass_actors: []`
  and `current_user_can_bypass: never`. The tag ruleset's rules are
  `deletion`, `non_fast_forward` and `update`.
* `GET /repos/…/environments/release` returns 404 (Q1).
* `scripts/audit-github.sh` reports the same.
* `GET /repos/…` shows the new description, and `license.spdx_id`
  `Apache-2.0`. `LICENSE`'s SHA-256 matches D9.
* A scratch Python check over the current docs finds none of: "mirror",
  "export" in the source-of-truth sense, "authoritative private source",
  "hooks-uninstall" given as a step, "split Go repository fleet". The README's
  CLI block equals `printUsage`'s lines, and `docs/README.md`'s records table
  has no blank line. The check is seen to fail on a scratch copy with one
  phrase put back.
* The next release tag push has the rule-suite result `pass` with no
  `creation` evaluation, and git prints no "Bypassed rule violations".
* Not probed live: refusal of a tag deletion or move, and of a force-push
  to `main`. A probe against a real ref would delete or rewrite it if the
  rule did not hold, and a probe tag matching `v*` could never be removed
  afterwards. Those refusals rest on GitHub's documented rule semantics.

## Pros and Cons of the Options

### A: empty bypass lists, `creation` replaced by `update`, failed tags burned

* Good, because it meets every decision driver.
* Bad, because tag creation is no longer restricted beyond write access,
  and a failed tag is permanent.

### B: empty bypass lists, `creation` kept, the tag ruleset disabled around each release

* Good, because creation stays restricted between releases.
* Bad, because each release becomes a manual ruleset edit, which is a
  bypass by another name and opens a window in which all three tag rules
  are off.
* Bad, because a release left half-done leaves the ruleset disabled.

### C: empty bypass lists and nothing else

* Bad, because `creation` then refuses every `v*` tag, so no release can
  be published.

### D: live change only

* Bad, because the next `scripts/configure-github.sh --apply` restores the
  bypass, and the script and guide go on describing it.

## More Information

* [GitHub: available rules for rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets)
* [GitHub: repository ruleset endpoints](https://docs.github.com/en/rest/repos/rules)
* [GitHub: deployment environments](https://docs.github.com/en/rest/deployments/environments)
* Implementation: [0025-PLAN-remove-ruleset-bypass-actors.md](0025-PLAN-remove-ruleset-bypass-actors.md)
