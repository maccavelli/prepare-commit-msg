---
status: accepted
date: 2026-11-07
decision-makers: Project Owner
consulted: the repository's releases, tags, rulesets and settings (GitHub API, read on 2026-11-06 and again on 2026-11-07); `proxy.golang.org`'s version list for this module; the attestations API for one `v1.2.0` asset; GitHub's immutable-releases documentation; go-selfupdate-lib at `74ab5dd376a4c864debb32b2d8abfc7856059ba3` (`v1.24.0`), its release lookup, channel policy and release scripts; `tools/` and its tests; the owner's answers of 2026-11-06 and 2026-11-07
informed: users of the installed hook, whose `update --version` and installer `--version` reach only the kept releases
---
# Keep only the eight newest stable releases, pruned by a hand-run maintainer tool

*Amendment A1 (2026-11-07, after the first run) corrects a fact: deleting
a release removed GitHub's release attestation for its assets, and kept
their build-provenance attestation. See "Amendments".*

*Revised 2026-11-07, before approval. The owner changed the window from
five releases to eight. The facts were read again: `v1.10.4` has been
released
([0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md](0027-MADR-adopt-go-llmprovider-sdk-v1-27-0.md)),
so there are 13 releases and 28 tags, and the expected first pruning is
recomputed. The four choices among options stand; only the count they apply
to changed. The filename keeps its slug, `keep-five-stable-releases`,
because the 0027 records cite and link this file by that name.*

*Revised 2026-11-06, before approval. The facts were re-read and hold.
Corrections: the tool is `tools/prune_releases.py`, because `tools/`'s tests
import each tool as a module, and a hyphen cannot be imported. Its `gh` is a
parameter, as `hosts.py`'s `ssh` is. Added: what `update --channel` does
with prereleases; how a prerelease is compared with a stable release; tags
and releases that are not semantic versions; how a failed `--apply` is
resumed; the attestations; and what changes if a release lands before the
first run.*

## Context and Problem Statement

The owner (2026-11-06): "how can we ensure that only 5 releases are saved
and available, not the 27 as now". Then (2026-11-07): "Change 0026 madr
spec to keep 8 releases".

Each claim is marked:

* **[read]**: from this repository, the GitHub API, the proxy, or
  go-selfupdate-lib's source at the commit named;
* **[derived]**: reasoned from those, and not run.

### What exists

Read again on 2026-11-07 **[read]**:

| Layer | Count | Detail |
| :--- | :--- | :--- |
| GitHub releases | 13 | `v1.2.0` to `v1.10.4`. None is a draft or a prerelease. Each reports `immutable: true`. `v1.2.0` to `v1.7.0` carry 7 assets each: the six binaries and `SHA256SUMS`. `v1.8.0` to `v1.10.4` carry 9, the installers included |
| Latest release | 1 | `v1.10.4` |
| `v*` tags | 28 | `v0.1.0` to `v1.10.4`. Each is an annotated tag and a semantic version. The 15 tags `v0.1.0` to `v1.1.4` have no release. No tag is outside `v*`, and no release lacks its tag |
| Go module versions | 28 | `proxy.golang.org`'s `@v/list` names exactly the 28 tags |
| Rulesets | 1 | `prepare-commit-msg-main`, on the default branch. No ruleset targets tags |
| Immutable releases | on | enabled for the repository, not enforced by the owner account |
| Collaborators | 1 | the owner |

### What can and cannot be removed

* **A release can be deleted, immutable or not.** GitHub
  ([immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases)):
  the tag "cannot be deleted while the release exists. If you delete the
  immutable release, you can delete the tag, but you cannot reuse the same
  tag name." So a release is deleted before its tag.
* **A tag can now be deleted.** The only ruleset targets `main`
  ([0025-MADR-remove-ruleset-bypass-actors.md](0025-MADR-remove-ruleset-bypass-actors.md)
  A1) **[read]**.
* **A Go module version cannot be removed.** The proxy keeps every version it
  has fetched, and the checksum database records each one for good.
  A `retract` directive in `go.mod` would hide a version from
  `go list -m -versions` and `@latest`, but `go install …@<version>` would
  still fetch it.
* **GitHub has no release-retention setting,** and go-selfupdate-lib has
  none. Its `retention-days` inputs cover workflow artifacts only
  (`build-selfupdate-release.yml`) **[read]**.

### What depends on old releases

* **`update`**, with no `--version`, asks for GitHub's latest release
  (`GitHubSource.Latest`), which is always kept **[read]**.
* **`update --version <v>`** asks for that tag's release
  (`GitHubSource.ByTag`). **The installers' `--version`** downloads from
  that tag's release. Once a release is deleted, neither can install it.
  `README.md:481`'s example names `v1.2.0` **[read]**.
* **The installers themselves** are fetched from the latest release
  (`releases/latest/download/install.sh`), so `v1.7.0` lacking them does not
  stop `--version v1.7.0`: that release has the binaries and `SHA256SUMS`
  **[read]** **[derived]**.
* **`update --channel rc`** lists every release, 30 to a page
  (`ListReleases`), and takes the highest one the channel admits. A stable
  release is admitted on every channel (`semverPolicy.Admits`). So a
  prerelease lower than the newest stable release is never chosen, and
  deleting it changes no channel's answer **[read]** **[derived]**.
* **The publish workflow** reads the latest release to set `--latest`
  (`scripts/release-latest-flag.sh`). It refuses a tag that already has a
  release (`refuse-existing-release.sh`). Pruning keeps the latest, so
  neither changes **[read]** **[derived]**.
* **This repository's code and tests read no published release.** The
  versions in `migration_test.go` and `releasespec_test.go` are fixtures.
  The README and `documentation_test.go` name only
  `releases/latest/download/…` **[read]**.
* **Attestations.** The attestations API returns 2 attestations for
  `v1.2.0`'s `prepare-commit-msg-linux-amd64` digest, on 2026-11-06 and
  again on 2026-11-07 **[read]**. They are kept by subject digest, not by
  release, so they should stay after the release is deleted **[derived]**.
  The first run checks it. *Corrected by A1: one of the two stayed.*

### What may change before the first run

A release may land before the first `--apply`. The rules below choose the
set at run time; only the expected lists in the Confirmation and the PLAN
assume no release lands first. If `v1.10.5` lands first, the newest eight
stable releases are `v1.8.0` to `v1.10.5`, and `v1.7.0` is deleted too
**[derived]**.

## Decision Drivers

* At most eight releases are available, and the eight newest stable ones
  are always among them.
* No workflow gets the power to delete releases or tags.
* A deletion is irreversible, so each one is seen before it happens.
* The tool follows `tools/`'s conventions (0016-MADR): Python 3 on the
  standard library, run by hand, tested in `tools/test_tools.py`, naming no
  account or host, and exit codes 0 clean, 1 a finding, 2 a tool error.

## Considered Options

The owner chose among three options on each of four questions
(2026-11-06), when the window was five. On 2026-11-07 the owner set it to
eight; the options below are stated with that count.

* **What counts.**
  * A, chosen: the eight newest stable releases; a prerelease is deleted
    once a newer stable release exists.
  * B: stable releases only; prereleases are never pruned.
  * C: the eight newest releases of any kind.
* **The 15 tags without a release.**
  * A, chosen: delete all 15 now.
  * B: keep them all.
  * C: delete the `v0.x` tags, and keep `v1.0.0` to `v1.1.4`.
* **Go retraction.**
  * A, chosen: none.
  * B: retract everything older than the kept eight, at each release.
  * C: retract `v0.1.0` to `v1.1.4` once.
* **The mechanism.**
  * A, chosen: a hand-run tool.
  * B: a CI step after publish.
  * C: a one-off cleanup by hand, with no tool.

## Decision Outcome

Chosen options: "A" on all four, with a window of eight. Each was the
agent's recommendation, and the owner chose it; the owner set the count. A
hand-run tool keeps the policy without handing a workflow delete rights. Go
retraction would add a step to every release for a hiding that
`go install …@<version>` ignores.

* **D1, the window.** The eight newest stable releases are kept, ordered by
  semantic version, so `v1.10.0` is above `v1.9.0`. A stable release is not
  a draft, not a prerelease, and has a tag of the form
  `vMAJOR.MINOR.PATCH`. Older stable releases are deleted, with their tags.
* **D2, prereleases.** A prerelease is deleted, with its tag, when its
  version is lower than the newest stable release's. A prerelease is lower
  than a stable release exactly when its `MAJOR.MINOR.PATCH` is at or below
  the stable one's, so `v1.11.0-rc.1` is lower than `v1.11.0`. A newer
  prerelease is kept. Drafts are never deleted, only reported.
* **D3, tags without a release.** A `v*` tag with no release is deleted
  when its version is lower than the oldest kept release's. Today that is
  the 15 tags from `v0.1.0` to `v1.1.4`. A newer one is kept and reported.
  It may belong to a release still running, or to a failed one, which the
  guide's recovery handles.
* **D4, what is never touched.** A tag that is not `v*`, a `v*` tag that is
  not a semantic version, and a release whose tag is either are reported,
  never deleted, and never counted in the window.
* **D5, the tool.** `tools/prune_releases.py`:
  * takes the repository from `git remote get-url origin` (an HTTPS or SSH
    GitHub URL), or `--repo OWNER/REPO`, and `--keep N` (default 8, at
    least 1). It reads the releases, the latest release and the tag refs
    with `gh api`;
  * by default prints the plan and deletes nothing: each release and tag it
    keeps, deletes and reports;
  * with `--apply`, deletes each release, then that release's tag, then
    each tag without a release. Then it reads back the releases, the latest
    release and the tags, and checks they match the plan;
  * refuses, deleting nothing, if fewer than `--keep` stable releases exist,
    if the plan would delete the latest release, or if any listing fails;
  * stops at the first failed deletion, prints what it deleted, and exits 2.
    A rerun finishes the work: a release already deleted is no longer
    listed, and its tag is then a tag without a release below the window;
  * exits 0 when there is nothing to prune or `--apply` succeeded, 1 when a
    dry run has something to prune, and 2 on a refusal or any other tool
    error.
* **D6, when it runs.** By hand, once a new release is verified. Running it
  is a step in the operations guide's controlled release. It also runs once
  now, to bring the repository to eight releases.
* **D7, no Go retraction.** `go.mod` gets no `retract` directive. The proxy
  keeps serving every version by name.
* **D8, docs.**
  * `README.md:164` and `:481`: `update --version` and the installers'
    `--version` reach the eight newest stable releases only. The `v1.2.0`
    example becomes `vX.Y.Z`.
  * `docs/guides/cicd-operations.md`: the pruning step in the controlled
    release, and how a clone drops the deleted tags
    (`git fetch --prune --prune-tags origin`).
  * `tools/README.md`: the tool's row and a usage section.

### Consequences

* Good, because the releases page and the tag list show at most eight
  versions, plus any newer prerelease, draft or unreleased tag, each of
  which the tool reports.
* Good, because only the owner deletes anything, through a tool that shows
  its plan first.
* Good, because `update`, `update --channel` and the publish workflow see
  the same answers as before.
* Bad, because rolling back by version reaches only the eight kept
  releases. Today that is `v1.7.0` and later.
* Bad, because a deleted release is gone for good. Its binaries,
  `SHA256SUMS` and installers cannot be restored, and its tag name can never
  be reused.
* Bad, because the window holds only if the tool is run. It is a manual
  step at each release.
* Neutral, because every version stays fetchable from `proxy.golang.org` by
  name. Pruning changes what GitHub offers, not what Go tooling can fetch.
* Neutral, because other clones keep the deleted tags until they fetch with
  `--prune-tags`.
* Neutral, because `v1.7.0`, the oldest kept today, carries no installers.
  Installing it by version uses the latest release's installer, which
  downloads only the binary and `SHA256SUMS`.

### Confirmation

* The tool's tests in `tools/test_tools.py` cover the plan, the default
  window and `--keep`, the order of deletions, the refusals, the exit codes
  and the read-back, against a fake `gh`. Each is seen to fail on a planted
  break.
* After the first `--apply`, read back. If no release has landed first:
  * releases: exactly `v1.7.0`, `v1.8.0`, `v1.9.0`, `v1.10.0`, `v1.10.1`,
    `v1.10.2`, `v1.10.3` and `v1.10.4`, each `immutable: true`;
  * GitHub's latest release: `v1.10.4`;
  * `v*` tags: exactly those eight;
  * `proxy.golang.org`'s `@v/list`: the same 28 versions;
  * the attestations for `v1.2.0`'s `prepare-commit-msg-linux-amd64`: still
    2. *The first run found 1; see A1.*

  If a release has landed first, the lists are the eight newest stable
  releases at that time.
* After that, a dry run exits 0, with nothing to prune.

## Amendments

### A1 (2026-11-07): deleting a release removes its release attestation

#### Context

The MADR expected both attestations for `v1.2.0`'s
`prepare-commit-msg-linux-amd64` digest to stay after that release was
deleted **[derived]**, and the PLAN's Phase 3 read-back checked the count:
"If it is not, record it in the MADR as a corrected fact; it does not stop
the phase."

#### Corrected fact

After the first `--apply` (2026-11-07), the attestations API returns 1 for
that digest, not 2 **[read]**:

* **Kept:** the build-provenance attestation, predicate
  `https://slsa.dev/provenance/v1`, whose 7 subjects are that release's
  binaries and `SHA256SUMS`. The release workflow signs it.
* **Removed:** GitHub's immutable-release attestation, predicate
  `https://in-toto.io/attestation/release/v0.2`. The kept `v1.7.0` still
  has both kinds for its `prepare-commit-msg-linux-amd64` **[read]**, so
  each release carries one, and deleting the release takes it.

#### Consequences

* Neutral, because `gh attestation verify --signer-workflow …` checks the
  build provenance, which stays: a copy of a deleted release's binary
  should still verify **[derived]**; no copy was kept to run it.
* Bad, because a deleted release's release attestation is gone with it,
  like its assets. No decision changes: D1 to D8 stand.

## More Information

* [GitHub: immutable releases](https://docs.github.com/en/code-security/concepts/supply-chain-security/immutable-releases)
* [GitHub: releases API](https://docs.github.com/en/rest/releases/releases)
* [GitHub: git references API](https://docs.github.com/en/rest/git/refs)
* [Go modules reference: retract](https://go.dev/ref/mod#go-mod-file-retract)
* The tools' conventions:
  [0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md](0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md)
* Implementation: [0026-PLAN-keep-five-stable-releases.md](0026-PLAN-keep-five-stable-releases.md)
