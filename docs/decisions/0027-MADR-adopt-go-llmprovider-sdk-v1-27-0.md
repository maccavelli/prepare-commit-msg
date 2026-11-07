---
status: accepted
date: 2026-11-06
decision-makers: Project Owner
consulted: go-llmprovider-sdk (its CHANGELOG for v1.27.0, the compare v1.26.1...v1.27.0 by commit and file, its 0070-MADR, and its source at v1.27.0); go-selfupdate-lib's tags; a scratch clone of this repository at e56c281 moved to v1.27.0; a probe of the SDK's catalog at v1.26.1 and v1.27.0 against a fake Grok listing; all 2026-11-06
informed: users of the installed hook, who receive the change through `update` once it is released
---
# Move to go-llmprovider-sdk v1.27.0, a catalog-only release, with no other change to the code

*Revised 2026-11-06, the day it was drafted. The owner answered Q1 to Q3:
release `v1.10.4`; a new "Changes with" section; and release before 0026's
first pruning (Q3 option C, not the recommended A). D3 and D4 hold as
written. D5 is rewritten to the owner's answer.*

## Context and Problem Statement

The owner (2026-11-06): go-llmprovider-sdk "has a new tagged release we
need to pin".

This repository pins its two libraries in four places **[read]**:

| Pin | Where | Pinned | Newest |
| :--- | :--- | :--- | :--- |
| go-llmprovider-sdk | `go.mod:6` | `v1.26.1` | `v1.27.0`, tagged 2026-11-06 |
| go-selfupdate-lib | `go.mod:7` | `v1.24.0` | `v1.24.0` |
| go-selfupdate-lib build workflow | `.github/workflows/ci.yml` | `74ab5dd3…` (`v1.24.0`) | unchanged |
| go-selfupdate-lib publish workflow | `.github/workflows/ci.yml` | `74ab5dd3…` (`v1.24.0`) | unchanged |

So only the first row moves. The workflow pins, the operations guide's
quoted `uses:` lines and `TestCurrentDocumentation` are not touched
(`0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md`
D2).

What has to be decided is whether the module moves, what the README tells
a user, whether the move is released, and how it is ordered with
`0026-MADR-keep-five-stable-releases.md`, whose first pruning is not yet
run.

Each claim is marked:

* **[read]**: from this repository, or from the SDK's changelog, source or
  git objects at the version named;
* **[probed]**: run in the session scratchpad on 2026-11-06, never in this
  tree;
* **[derived]**: reasoned from the two, and not run.

### The release

* `v1.27.0` is an annotated tag on `52ab75fa31131e2aee2b313a0a8cf923afbad26d`.
  The SDK's CI passed on the tag. `proxy.golang.org` and
  `go list -m -versions` list it **[read]**.
* `v1.26.1...v1.27.0` holds 8 commits and changes 23 files. The Go changes
  are all in `llmprovider/catalog/`. `wizard/`, `llmprovider/providers/`,
  the root `llmprovider` package and `go.mod` are unchanged **[read]**.
* Its `go.mod` still requires `go 1.27.3`, `golang.org/x/term v0.43.0` and
  `golang.org/x/sys v0.47.0` **[read]**. This repository already requires
  newer `x/term` and `x/sys` through go-selfupdate-lib, so no indirect
  requirement moves **[probed]**.
* The SDK repository has immutable releases off, and its GitHub release for
  `v1.27.0` carries no assets **[read]**. A Go module does not depend on
  that: `go.sum` holds the module's hash, and the checksum database holds
  the same one. `go mod verify` reports "all modules verified"
  **[probed]**.

### What v1.27.0 changes

From its CHANGELOG and its
`0070-MADR-place-grok-4-20-by-date-and-split-model-labels.md` **[read]**:

* **Fixed, ranking.** `catalog.Rank` read `grok-4.20-*` as generation
  4.20, above `grok-4.7`, though xAI released it before `grok-4.3`. A
  family's two-digit minor is now read as an integer or a decimal. The
  family's dates decide, else its siblings, else the integer reading. So
  `grok-4.20` is 4.2, and `gemini-3.10` still follows `gemini-3.9`. Gemini,
  Claude and Grok listings are ranked this way; under the integer reading,
  every score is as before.
* **Changed, labels.** Eight label notes:
  * `grok-4.6`'s no longer calls it the Grok CLI's default, which is
    grok-build's compiled-in fallback;
  * `grok-4.5`'s "previous flagship" was false;
  * the notes that ranked a model in time no longer do;
  * `claude-opus-4-8`'s names its generation.

  A label is still the name padded to 23 columns, then the note in
  brackets, starred on the provider's first static model.
* **Added.** `catalog.SortByRank`, `catalog.LabelName` and
  `catalog.LabelNote`.
* **Unchanged.** Every `catalog.Static` list. The search's name field is
  now taken from `LabelName` rather than cut from the label at `[`. For
  every label this gives the same text **[read]** **[derived]**.

### What reaches this app

The app calls `catalog.Static`, `catalog.List`, `catalog.OptionsFromEnv`,
`catalog.WithLimit`, `catalog.MaxListed` and `catalog.WithClaudeWorkspace`.
It calls no `Rank`, `Label` or `Search` of its own. The wizard calls
`Label` for its menus (`wizard/configure.go:1001`,
`wizard/model_select.go:186`, `:269`) **[read]**.

A probe built the same program against `v1.26.1` and `v1.27.0`
**[probed]**. It served a fake xAI `/models` listing and called
`catalog.List` with `WithLimit(catalog.MaxListed)`, as `discoverModels`
does (`internal/ui/setup.go:158-184`). The listing's ids came from the
SDK's dated fixture: `grok-4.7`, `grok-4.6`, `grok-4.5`, `grok-4.3`,
`grok-build-0.1`, two `grok-4-1-fast-*` and three `grok-4.20-*`. It found:

* **The first five Grok models are unchanged.** `catalog.List` puts the
  static list first, in its order: `grok-4.6`, `grok-4.5`, `grok-4.7`,
  `grok-4.3`, `grok-build-0.1`. Only the slots after them are ranked.
  `MaxListed` is 6.
* **`grok-4.20-*`'s rank** falls from 93, above every Grok model, to 40,
  the same as `grok-4-1-fast-*` and below `grok-4.3`'s 45.
* **So the sixth menu slot depends on the listing's order.** At `v1.26.1`,
  `grok-4.20-0309-reasoning` took it whatever the order. At `v1.27.0`, it
  goes to whichever of the two families, both scoring 40, comes first in
  the listing. With `grok-build-0.1` absent, the same holds for slots five
  and six.
* **The defaults do not move.** `configure --yes` takes the first listed
  model, `grok-4.6`. Its fallbacks are the next three,
  `config.MaxFallbacks` (`internal/config/config.go:24`): `grok-4.5`,
  `grok-4.7` and `grok-4.3`. Both are the same at both versions
  **[probed]** **[derived]**. A saved configuration is not re-read through
  the catalog, so the hook's model and fallbacks do not change
  **[read]**.
* **The labels the wizard shows** differ in the eight notes, among them:

  | Model | `v1.26.1` note | `v1.27.0` note |
  | :--- | :--- | :--- |
  | `grok-4.6` | ★ Recommended: the Grok CLI's default | ★ Recommended: the Grok CLI's built-in default |
  | `grok-4.5` | previous flagship | second in the Grok CLI's built-in list |
  | `grok-4.7` | current flagship | frontier model |
  | `claude-sonnet-5-5` | latest Sonnet, high precision | high precision |
  | `claude-opus-5-5` | latest Opus, maximum capability | maximum capability |
  | `claude-opus-4-8` | maximum capability | maximum capability of Opus 4 |
  | `gemini-3.8-flash` | newest Flash, frontier coding | frontier coding |
  | `gpt-5.6-luna` | previous efficient tier | efficient tier |

  No README text or test in this repository quotes a label **[read]**.

This corrects the agent's first assessment, given in the session before
this record. It said a Grok listing holding `grok-4.20` would change the
menu's order and its default pick. The probe shows the default is the
static list's first model, at both versions. Only the ranked slots after
the static five can change.

### The move, on a scratch clone

On a clone of `e56c281` **[probed]**:

* `go get github.com/maccavelli/go-llmprovider-sdk@v1.27.0` and
  `go mod tidy` change one line of `go.mod` and two of `go.sum`, the
  module's `h1:` and `/go.mod` hashes. The `/go.mod` hash is the same as
  `v1.26.1`'s. `go mod tidy -diff` is empty.
* `make verify` exits 0:
  * mod-check, fmt-check, golangci-lint, `go vet`, and
    `go test -race ./...`, `TestReadmeCuratedModels` included;
  * total coverage 83.9%, against a minimum of 80.0%;
  * govulncheck: "No vulnerabilities found.";
  * workflow-lint: "ci filters: 11 patterns, 11 reads checked, 0
    problems";
  * the tools' 25 tests and the scripts' 34;
  * the six release cross-builds.

### Interaction with the release pruning

`0026-MADR-keep-five-stable-releases.md` is proposed, and its first
`--apply` has not run. Its rules choose the five newest stable releases at
run time. If this move is released as `v1.10.4` before that run, the kept
five become `v1.10.0` to `v1.10.4`, and `v1.9.0` is deleted too. Its PLAN's
Phase 3 step 1 handles that case **[read]** **[derived]**.

## Decision Drivers

* **Pins are current** unless a record names a reason to hold one.
* **The change is a pin move.** New library features are separate
  decisions.
* **The README says what a user can see,** in the "Changes with …"
  pattern.
* **A release carries something a user would want,** and the owner tags it
  in the turn it is made.
* **No step of one record makes another record's expected results wrong
  without saying so.**

## Considered Options

The module move is common to every option, and not in question:

* **A. Move the pin, and keep the app's code as it is.** Rely on the SDK
  for the ranking and the labels, and adopt none of its additions.
* **B. Move the pin, and adopt `LabelName` and `LabelNote`.** Use them in
  the app's own output, such as `configure`'s closing summary, which prints
  the bare model id (`internal/ui/setup.go:346-348`).
* **C. Hold at `v1.26.1`.** Record the reason, and move with the SDK's
  next release.

## Decision Outcome

Chosen option: "A", because the release fixes a wrong ranking and wrong
label text in the SDK, which this app receives with no code of its own. The
additions answer needs this app does not have. B would add Go code and
tests for wording. C would leave a fix the next move must take anyway,
with no reason to hold.

D3 to D5 record the owner's answers to the questions below.

### D1. The module moves

* `go.mod` requires `github.com/maccavelli/go-llmprovider-sdk v1.27.0`,
  after `go mod tidy`.
* `go.mod` changes on that line alone. `go.sum` changes on that module's
  two lines alone. `go mod tidy -diff` is empty, and `go mod verify` passes.
* go-selfupdate-lib, both workflow pins, the operations guide and the Go
  toolchain do not change.

### D2. No other code changes

* Not adopted: `catalog.SortByRank`, `catalog.LabelName` and
  `catalog.LabelNote`.
* No test is added. The app's tests and `TestReadmeCuratedModels` already
  cover what reaches it: the static lists, `discoverModels` and the
  wizard's wiring. The ranking and label changes are the SDK's, and the
  SDK's own tests cover them (its `0070` tests) **[read]**.

### D3. The README says what changed (Q2)

* A section, "Changes with go-llmprovider-sdk v1.27.0", after the
  `v1.26.1` one, with its table-of-contents line. It says, for
  `configure`:
  * Grok's `grok-4.20` models are ranked by release, below `grok-4.3`, so
    they no longer take the menu's last place ahead of older-numbered
    models; the default model and fallbacks are as before;
  * some model descriptions in the menus are reworded, so none ranks a
    model in time. Grok 4.6 is still the recommended Grok model.
* It says the hook, `update` and the installers behave as before.
* It names the release that carries it, `v1.10.4` (D4).

### D4. The release (Q1)

* Release `v1.10.4`, a patch. It carries a ranking fix and corrected
  descriptions, with no change to the hook. The owner tags after the change
  is on `main` and CI there is green, and asks for the tag in the turn it
  is made.

### D5. The order with 0026 (Q3)

* This move is committed, pushed and released as `v1.10.4` without waiting
  for `0026-MADR-keep-five-stable-releases.md`.
* When 0026's first pruning runs, after `v1.10.4` is verified, its PLAN's
  Phase 3 step 1 finds the new release, and works out the expected plan by
  its MADR's D1 to D4. Keep `v1.10.0` to `v1.10.4`. Delete the releases
  `v1.2.0` to `v1.9.0` (8), their tags, and the 15 tags without a release:
  8 releases and 23 tags **[derived]**. That case is foreseen by 0026, so
  this record needs no amendment there. The owner sees the plan before any
  deletion.

### Consequences

* Good, because the pin is current, and `configure` stops showing Grok's
  `grok-4.20` as the newest Grok model.
* Good, because the menus no longer show descriptions the next release
  would make false, such as "previous flagship".
* Good, because nothing else moves: one module line, no code, no workflow.
* Neutral, because the hook, a saved configuration, `configure --yes`'s
  default model and fallbacks, `update` and the installers behave as
  before **[probed]** **[derived]**.
* Neutral, because at `v1.27.0` the sixth Grok menu slot follows the
  listing's order between two families of equal rank. That is the SDK's
  choice, not this app's.
* Bad, because a release for this change uses a version number. Released
  before 0026's first pruning (D5), it moves that pruning's window on by
  one, so `v1.9.0` is deleted too, and rolling back by version reaches
  `v1.10.0` and later.

### Confirmation

* `go.mod` and `go.sum` change only on D1's lines. `go mod tidy -diff` is
  empty, and `go mod verify` passes.
* `grep` finds `74ab5dd376a4c864debb32b2d8abfc7856059ba3` twice in
  `ci.yml` and twice in the operations guide, as before.
* `make preflight`, `make disclosure-check` and `make preflight-hosts` pass
  on the committed `HEAD`. CI is green on `main`, its build rehearsal
  included.
* The README section and its contents line exist, and `make check-records`
  and `make markdownlint` pass.
* If D4 releases: the tag run publishes the six binaries, `SHA256SUMS` and
  both installers as an immutable release with no "Bypassed rule
  violations" message, the check deferred by
  `0025-PLAN-remove-ruleset-bypass-actors.md`. `update` on an installed
  `v1.10.3` moves to it, and `version` names it.

## Pros and Cons of the Options

### A. Move the pin, and keep the app's code as it is

* Good, because the change is one module line and the README.
* Good, because the fix and the wording reach the wizard through the SDK.
* Neutral, because the app's own summary keeps printing bare model ids, as
  it does today.

### B. Move the pin, and adopt `LabelName` and `LabelNote`

* Good, because `configure`'s summary could print "Grok 4.6" rather than
  `grok-4.6`.
* Bad, because it adds Go code, tests and a visible output change for a
  wording gain the owner has not asked for.
* Bad, because it ties the app's output to the SDK's curated names, which
  change between releases, as this one shows.

### C. Hold at `v1.26.1`

* Good, because nothing changes now.
* Bad, because `configure` keeps ranking `grok-4.20` above every Grok model
  and keeps notes the SDK found false.
* Bad, because the next move must take this release anyway, with no reason
  recorded to hold it.

## Owner questions

* **Q1 (D4). Is the move released?**
  * A, recommended: release `v1.10.4`, a patch, once the change is on
    `main` and CI is green.
  * B: leave it unreleased, and ship it with the next change that needs a
    release. The README section says "the next release".
  * C: release it as `v1.11.0`, a minor, to mark the SDK's minor version.
* **Q2 (D3). What does the README say?**
  * A, recommended: a "Changes with go-llmprovider-sdk v1.27.0" section, in
    the pattern of the others, with its contents line.
  * B: one line added to the "Changes with go-llmprovider-sdk v1.26.1 and
    go-selfupdate-lib v1.24.0" section. That section describes a release
    already made, so this changes its history.
  * C: nothing in the README. This record holds the change, and users see
    the new descriptions in `configure`.
* **Q3 (D5). How is this ordered with 0026's first pruning?**
  * A, recommended: commit and push the move whenever approved, and tag
    only after 0026's first `--apply` is read back.
  * B: run all of 0026 first, its records included, then start this move.
  * C: release this first, and let 0026's PLAN Phase 3 step 1 recompute its
    expected lists, so `v1.9.0` is deleted too.

**Answered 2026-11-06.** Q1: "Release v1.10.4", option A, as recommended;
D4 records it. Q2: "New 'Changes with' section", option A, as recommended;
D3 records it. Q3: "Release 0027 first", option C, where A was
recommended; D5 records it. The chosen option, A (the move with no other
code change), was not questioned and stands.

## More Information

* The SDK's record, cited by repository and filename: go-llmprovider-sdk
  `0070-MADR-place-grok-4-20-by-date-and-split-model-labels.md`.
* The last move:
  [0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md](0024-MADR-adopt-go-llmprovider-sdk-v1-26-1-and-go-selfupdate-lib-v1-24-0.md).
* The release pruning:
  [0026-MADR-keep-five-stable-releases.md](0026-MADR-keep-five-stable-releases.md).
* The deferred release check:
  [0025-PLAN-remove-ruleset-bypass-actors.md](0025-PLAN-remove-ruleset-bypass-actors.md).
* The implementation:
  [0027-PLAN-adopt-go-llmprovider-sdk-v1-27-0.md](0027-PLAN-adopt-go-llmprovider-sdk-v1-27-0.md).
