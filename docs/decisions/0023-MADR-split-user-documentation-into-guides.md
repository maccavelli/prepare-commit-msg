---
status: proposed
date: 2026-10-25
---
# Split the user documentation into `docs/guides/`, indexed by the READMEs

## Context and Problem Statement

Assessed 2026-10-25 at `cfd03e6`. The documentation tree already follows the
fleet scaffold ([0013-MADR-align-repository-with-workspace-scaffolding.md](0013-MADR-align-repository-with-workspace-scaffolding.md)):
`docs/` holds `README.md`, `architecture.md`, `decisions/`, `reports/` and one
guide, `guides/cicd-operations.md`. A separate, unrelated pair —
[0022-MADR-align-docs-ci-and-disclosure-gates.md](0022-MADR-align-docs-ci-and-disclosure-gates.md),
proposed and unexecuted — sits uncommitted in this working tree; this decision
does not touch or execute it, and its index rows are preserved.

The root `README.md` (831 lines) mixes three audiences in one file:

* the product overview ("What it is", "Why This Exists");
* everything an end user of the hook needs — installation, configuration,
  providers and models, hook behavior, self-update, and seven historical
  "Changes with …" release-note sections;
* maintainer documentation — the Developer Experience targets and the
  maintainer release and attestation procedures.

There are no user guides: a reader who wants "how do I configure Ollama" or
"what does the hook do to my `git commit`" must wade through the whole README.
The repo-standard shape for this documentation is the one in the sibling
repository gobble-cli (`docs/README.md`, "guides/" list and "I want to…"
matrix): thin product READMEs carry the ToC and the "I want to…" matrices and
link task-shaped guides, which live in `docs/guides/`.

Two Go tests enforce documentation facts at their current locations:

* [readme_test.go](../../readme_test.go) parses the provider table in the
  root `README.md` and compares it with the SDK's curated catalog;
* [documentation_test.go](../../documentation_test.go) requires the root
  `README.md` to contain the two exact installer commands and the string
  `Release v1.10.1 moves to Go 1.27.3`, and requires
  `docs/guides/cicd-operations.md` to cite both pinned release-workflow
  references.

Both pass at this tree (verified 2026-10-25), so the current content is in
sync with the code; a split must keep the facts enforced wherever the content
lands.

## Decision Drivers

* A reader reaches a task in five minutes from a README's "I want to…" matrix.
* Reference depth lives in guides; the product READMEs stay short and stay the
  entry points, with ToC and matrix linking to the guides.
* End-user documentation and maintainer documentation separate cleanly.
* The fleet convention matches gobble-cli's established layout, which the
  owner asked this split to follow.
* Fact-pinning tests keep pinning the facts, wherever the content moves.
* The uncommitted 0022 pair is preserved, and nothing here stages or executes it.

## Considered Options

* Keep the monolithic root README and refresh it in place.
* Split the user documentation into task guides under `docs/guides/`, with ToC
  and "I want to…" matrices in the READMEs, and amend the two Go doc-reader
  tests to follow the moved content.
* Put user guides in a separate top-level `user-guides/` tree.

## Decision Outcome

Chosen option: "Split the user documentation into task guides under
`docs/guides/`, with ToC and "I want to…" matrices in the READMEs, and amend
the two Go doc-reader tests to follow the moved content", because it matches
the fleet convention the owner named, separates end-user from maintainer
documentation, and keeps every fact pinned by a test at its new location.
This does not change CI routing or reproduce 0022's scope.

### Consequences

* Good, because task-shaped guides grow without bloating the product README.
* Good, because maintainer content (developer targets, release and
  attestation procedures) sits with the maintainer runbook instead of beside
  installation instructions.
* Neutral, because a user who preferred one long page now makes one hop
  through the matrix.
* Bad, because README and guide links are outside `check_records.py`'s
  numbered-record coverage; a throwaway link resolver must be written and
  proven to fail on planted bad input, per the documentation rules.
* Bad, because two Go tests change in the same work, and temporarily more
  Markdown becomes test input that must stay in sync.

### Confirmation

The paired [0023-PLAN-split-user-documentation-into-guides.md](0023-PLAN-split-user-documentation-into-guides.md)
amends both doc-reader tests, proves the amended tests and the link resolver
fail on planted breaks, and runs `make check-records`, `make markdownlint`,
the documentation tests and `make preflight` before each commit.

## Pros and Cons of the Options

### Keep the monolithic root README and refresh it in place

* Good, because no tests or links change.
* Bad, because the README keeps growing by release, as its seven "Changes
  with …" sections show, and the maintainer content stays buried at the end.

### Split the user documentation into task guides, with matrices in the READMEs

* Good, because the READMEs become the five-minute entry points and the
  collection is where gobble-cli's is.
* Bad, because the doc-reader tests must be amended to follow the content.

### Put user guides in a separate top-level `user-guides/` tree

* Good, because user documentation is then namespaced away from records.
* Bad, because it violates the fixed docs-tree layout of the fleet scaffold
  and 0013-MADR; `docs/guides/` is the sanctioned home.

## More Information

This decision is an increment over the reconciliation of
[0020-MADR-reconcile-current-product-and-maintainer-documentation.md](0020-MADR-reconcile-current-product-and-maintainer-documentation.md)
and covers the same root README those sections shaped. The pattern to follow
is the one in the gobble-cli repository (`gobble-cli/docs/README.md`, guides
listed with one-line descriptions plus the "I want to…" matrix for users); a
relative link cannot reach that repository, so it is cited by repository and
filename. Facts were verified against the code on 2026-10-25: `go.mod` pins
(go-llmprovider-sdk `v1.22.2`, go-selfupdate-lib `v1.22.0`, Go 1.27.3), the
Makefile targets against the root README's table, the CLI flags and defaults
against the binary's help and `internal/config`, the provider endpoints and
credential variables against the SDK's provider registry, and the workflow
pins against `.github/workflows/ci.yml` and the operations guide.
