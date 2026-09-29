---
status: proposed
date: 2026-09-28
associated-madr: "0007-MADR-dependency-and-docs-link-refresh.md"
decision-makers: Project Owner
---
# Implement the Dependency and Documentation Link Refresh

Associated MADR:
[0007-MADR-dependency-and-docs-link-refresh.md](0007-MADR-dependency-and-docs-link-refresh.md)
(`status: proposed`).

This plan executes the MADR's options D-a, L-a and I-a, and nothing else. If
a fact contradicts the MADR or this plan, **stop and prompt**. Add a dated
entry to §9, amend the MADR when a decision or an asserted fact changes, and
only then continue.

## Goal

* The module's dependencies are the latest that `go get -u ./...` selects.
* Every Markdown link in the repository resolves, with no machine path left.
* `docs/README.md` lists every record, with each status its record states.

## Scope

* **Changed:**
  * `go.mod` and `go.sum`;
  * the Markdown under `docs/` that the MADR's link table names;
  * `docs/README.md`.
* **Unchanged:**
  * Go source, tests, CI and the Makefile;
  * the records' own statuses;
  * the root `README.md` text.
* **No release.** The dependency update ships with the next tag, which is not
  part of this plan.

## Tools

Two scratch scripts, kept outside the repository:
* **The link checker** lists every link in the tracked `.md` files, outside
  code fences, that is either:
  * a `file://` link or an absolute path;
  * a relative link whose file does not exist.
  It does not fetch `http(s)` links. It exits 1 when it finds anything.
* **The rewriter** applies §R2's table (step L) and §R3's entries (step I).
  Every rule asserts its match count, and a link no rule covers stops the run.

## 0. Preconditions

* `main` at `2bd6dab` or a descendant that touches none of the scoped files,
  level with `origin/main`, with a clean tree apart from this pair.
* Hooks as in the 0006 PLAN §0.1: commits go through the global message hook,
  and pre-push runs `make verify`.
* Each phase that changes files ends with one `git commit --no-edit`.
* Every push needs the owner's explicit ask in that turn.

## Phase R0 — Start

1. Set the MADR to `status: accepted`, and this plan to `status: in-progress`.
2. Commit the two records only (the bootstrap exception).

## Phase R1 — Dependencies (MADR D-a)

1. Run `go get -u ./... && go mod tidy`.
2. `git status` must show only `go.mod` and `go.sum`, and the upgrades must be
   exactly the MADR's table. Anything else is a deviation.
3. Run `make verify`. It must pass with no failing test and the coverage
   minimum met.
4. Commit.

## Phase R2 — Links (MADR L-a)

1. Run the rewriter, step L. It rewrites link **targets** only; the visible
   text of each link is unchanged.

   | Link target | Rewritten to |
   | --- | --- |
   | this repository's checkout, a file that exists | a path relative to the linking file |
   | this repository's checkout, `README.md#L60-L61` (0001 MADR) | `prepare-commit-msg/blob/9f55d6836b94b8347847d78f02a524e019c8ea1c/README.md#L60-L61` |
   | this repository's checkout, `internal/selfupdate/*` | `prepare-commit-msg/blob/79cdba965289449c2993d731d16de10f6a78ab85/<path>` |
   | this repository's checkout, the 0001 MADR's old filename | the relative path to `0001-MADR-gemini-provider-and-model-catalog-modernization.md` |
   | module cache, mcplib `v0.2.0` | `mcplib/blob/d8cce03a5007dd9f5e88f1630f97a094c25eab77/<path>` (`tree/` for directories) |
   | the author's mcplib checkout | `mcplib/blob/main/<path>` |
   | `magic-cli-remote` (redacted absolute path) | `magic-cli-remote/blob/master/<path>` |
   | `../.github/workflows/release.yml` and `quality.yml` | `prepare-commit-msg/blob/7b91a1ae38be006ffd0779aefe159bba1d597b59/<path>`, keeping `#L78` |

   Every URL is under `https://github.com/maccavelli/`.
2. The same step replaces the 0001 PLAN's five plain-text machine paths with
   `<mcplib checkout>` and `<prepare-commit-msg checkout>`.
3. **Checks:**
   * the link checker reports 0 problems;
   * every new GitHub URL resolves through the contents API;
   * a `git grep` for `file:` URLs and for the three rewritten absolute
     prefixes finds nothing. The pattern lives in the scratch tool, because
     writing it here would put those prefixes back in a committed file;
   * `git diff --stat` shows 8 files, 74 insertions and 74 deletions.
4. Commit.

## Phase R3 — Index (MADR I-a)

1. Run the rewriter, step I, on `docs/README.md`:
   * **0004 MADR:** status "Accepted" becomes "Proposed", as its front matter
     says.
   * **0002 PLAN:** status "Ready for User Review" becomes "Completed", as its
     front matter says.
   * **Added:** the 0005 MADR and PLAN, both "Proposed".
   * **Added:** the 0007 MADR ("Accepted") and PLAN ("In Progress").
2. The link checker must report 0 problems.
3. Commit.

## Phase R4 — Push (owner's ask)

1. Push `main`. The pre-push `make verify` must pass.
2. CI must pass on every job: `Go (test; build on tag)`, and Native Tests on
   Linux, Windows and macOS.

## Phase R5 — Records

1. Record each phase in §10, with its output.
2. Set this plan to `status: complete`, and its index line to "Completed",
   once §7 holds.
3. Commit, and push on the owner's ask.

## 7. Acceptance criteria

* `go.mod` has the MADR's upgrades, and `make verify` passes.
* The link checker reports 0 problems. It was seen to fail: 73 problems on
  `HEAD`, and 1 on a copy with an injected broken link.
* Every new GitHub URL resolves.
* No machine path is left in tracked Markdown.
* `docs/README.md` lists 0001–0007, with each status its record states.
* CI passes after R4.

## 8. Rollout and rollback

* Nothing ships until the next release.
* **Rollback:**
  * R1: revert its commit. The documentation phases do not depend on it.
  * R2 and R3: revert their commits. They change documentation only.

## 9. Deviation log

None yet.

## 10. Execution record

Not started.

## Appendix A — Proof record (2026-09-28)

Run on scratch archives of `2bd6dab`, never in the repository.

**R1.** `go get -u ./... && go mod tidy` gave exactly the MADR's table.
`make verify` passed:

```text
ok  	github.com/maccavelli/prepare-commit-msg	0.898s	coverage: 83.1% of statements
ok  	github.com/maccavelli/prepare-commit-msg/internal/config	0.862s	coverage: 80.1% of statements
ok  	github.com/maccavelli/prepare-commit-msg/internal/fsutil	1.218s	coverage: 78.4% of statements
ok  	github.com/maccavelli/prepare-commit-msg/internal/git	1.595s	coverage: 96.5% of statements
ok  	github.com/maccavelli/prepare-commit-msg/internal/ui	1.157s	coverage: 70.5% of statements
total coverage: 80.3% (minimum 80.0%)
No vulnerabilities found.
```

**R2 and R3.**

```text
1. checker on HEAD: exit=1 :: 73 problem link(s)
2. step L: exit=0 :: links: 8 files rewritten
    8 files changed, 74 insertions(+), 74 deletions(-)
   checker after L: exit=0 :: 0 problem link(s)
2. step I: exit=0 :: index: rewritten
    1 file changed, 18 insertions(+), 2 deletions(-)
   checker after I: exit=0 :: 0 problem link(s)
3. 22 new GitHub URLs: all OK through the contents API
4. machine paths left: none
5. checker on a copy with one injected broken link: exit=1 ::
   docs/README.md:65: MISSING decisions/0099-MADR-does-not-exist.md
   1 problem link(s)
```

The rewriter's diff is not reproduced here. Its removed lines carry the
machine paths this plan removes.
