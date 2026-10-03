---
status: proposed
date: 2026-09-28
decision-makers: Project Owner
consulted: none
informed: none
---
# Refresh Dependencies to Latest and Repair Every Documentation Link

## Context and Problem Statement

After the mcplib `v1.6.0` promotion
([0006-MADR-mcplib-1-6-canary.md](0006-MADR-mcplib-1-6-canary.md)), the owner
asked on 2026-09-28 to "update prepare-commit-msg go.mod, update docs, fix
links". This record takes those as three changes:
* bring the module's dependencies up to date;
* bring the documentation index (`docs/README.md`) up to date;
* fix every broken or non-portable link in the tracked Markdown.

Other documentation content, such as the root `README.md` feature text, is out
of scope.

### What was measured

On 2026-09-28, at `HEAD` `2bd6dab`:

**Dependencies.** `go.mod` requires `github.com/maccavelli/mcplib v1.6.0`
directly. Every other requirement is indirect, arriving through mcplib.
`go list -m -u all` lists newer versions for eleven modules. On a scratch
archive of `HEAD`, `go get -u ./... && go mod tidy` changes only `go.mod` and
`go.sum`:

| Module | From | To |
| --- | --- | --- |
| `github.com/modelcontextprotocol/go-sdk` | `v1.6.1` | `v1.8.0` |
| `github.com/segmentio/asm` | `v1.1.3` | `v1.2.1` |
| `golang.org/x/mod` | `v0.40.0` | `v0.41.0` |
| `golang.org/x/oauth2` | `v0.36.0` | `v0.37.0` |
| `golang.org/x/sys` | `v0.47.0` | `v0.48.0` |
| `golang.org/x/term` | `v0.43.0` | `v0.46.0` |
| `golang.org/x/sync` | (not listed) | `v0.23.0` |
| `golang.org/x/time` | (not listed) | `v0.16.0` |

`make verify` passed on that copy: no failing test, total coverage 80.3%
(minimum 80.0%), and govulncheck found no vulnerabilities.

The other five modules `go list -m -u` names are in the module graph but not
in the build. `go get -u ./...` leaves them alone.

`go list -deps ./...` shows that the binary links `go-sdk` (`mcp`, `auth`,
`jsonrpc`, `oauthex`) through mcplib. mcplib `v1.6.0` itself requires `go-sdk`
`v1.6.1`. After the update, mcplib's packages in this binary would run on a
`go-sdk` version mcplib's own tests have not run on.

**Links.** A link check over every tracked `.md` file, outside code fences,
found 73 problems in 8 files:

| Kind | Count | Where |
| --- | --- | --- |
| `file://` links into the author's checkout of this repository | 57 | `docs/README.md`, the 0001 and 0002 MADRs and PLANs |
| `file://` links into the Go module cache, at mcplib `v0.2.0` | 5 | the 0001 MADR and PLAN |
| `file://` links into the author's mcplib checkout | 3 | the 0001 PLAN |
| an absolute path (already redacted to `<user>`) into `magic-cli-remote` | 1 | the 0004 MADR |
| relative links to `release.yml` and `quality.yml`, deleted by `44b95af` | 7 | the 0003 and 0004 MADRs, the 0004 PLAN |

The `file://` links name real-machine paths, which this repository's rules
keep out of committed files, and they are broken for every other reader. The
0001 PLAN also has five plain-text machine paths in a note and in its
verification commands.

The targets were traced. Of the 57 links into this repository's checkout:
* 32 name a file that still exists.
* 24 name the `internal/selfupdate/` files deleted by `9f887d3`. Their last
  version is at `79cdba9`.
* 1 names the 0001 MADR by an old filename, from before it took the `MADR`
  infix.
* One of the 32, the 0001 MADR's `README.md#L60-L61` anchor, quotes
  `gemini-2.0-flash` lines.
  Those lines are at `9f55d68`, the parent of the commit that added the MADR,
  and not at `HEAD`.
* mcplib `v0.2.0` is commit `d8cce03`, according to the module proxy's
  `Origin`. That tag is not on GitHub, but the commit is.
* `release.yml` line 78 at `7b91a1a` is the `attest-build-provenance` pin
  that the 0004 MADR cites.

**Index.** `docs/README.md` omits the 0005 pair. Two index statuses disagree
with their record's own front matter:
* **0004 MADR:** the index says "Accepted"; the record says `proposed`.
* **0002 PLAN:** the index says "Ready for User Review"; the record says
  `completed`.

## Decision Drivers

* Every link resolves, for any reader, on any machine.
* No real-machine path in a committed file.
* A historical record keeps what it meant. A link to something since deleted
  or changed points at the version the record described.
* The dependency update is verified by the repository's own gate.
* The index reports what each record says of itself. It does not decide a
  record's status on the record's behalf.

## Considered Options

* **Dependencies:**
  * D-a: `go get -u ./...`, everything in the build.
  * D-b: only the `golang.org/x/*` modules.
  * D-c: keep mcplib's versions and change nothing.
* **Links:**
  * L-a: relative links where the target exists; GitHub permalinks at the
    described commit where it does not; GitHub URLs for other repositories.
  * L-b: relative links only; drop the links whose target is gone.
  * L-c: GitHub `main` URLs for everything.
* **Index statuses:**
  * I-a: mirror each record's front matter where it has one; otherwise keep
    the index text.
  * I-b: keep the index text as it is.
  * I-c: change the records' statuses to match the index.

## Decision Outcome

Chosen options: **D-a**, **L-a** and **I-a**.
* **D-a**, because the owner asked for the update, and the full gate passes
  on it.
* **L-a**, because it is the only option that both fixes every link and keeps
  each historical record's meaning.
* **I-a**, because a record's status is its own claim to make.

### Consequences

* Good, because every Markdown link in the repository resolves, and the
  checker finds 0 problems.
* Good, because no committed Markdown carries a machine path.
* Good, because the dependencies are current. The next release ships them.
* Neutral, because this plan cuts no release. The update reaches users with
  the next tag.
* Bad, because `go-sdk` `v1.8.0` runs under mcplib code that mcplib's tests
  have not run on. This repository's gate covers only its own behaviour. The
  mitigation is the same bump in mcplib, which is a follow-up there.
* Bad, because the 0004 MADR's index line goes from "Accepted" to "Proposed".
  That is what the record says, although its work (the unified `ci.yml`)
  shipped. Changing the record is for the owner, outside this record.
* Neutral, because the link checker is a scratch tool, not a gate. A new
  `file://` link could still be committed later.

### Confirmation

* The dependency change is exactly the table above, and `make verify`
  passes.
* The link checker reports 0 problems. It reported 73 on `HEAD`, and 1 on a
  copy with one injected broken link.
* Each new GitHub URL resolves through the GitHub contents API.
* `git grep` finds no machine path in tracked files.
* CI passes.

## Pros and Cons of the Options

### D-a: `go get -u ./...`

* Good, because it does what was asked, completely.
* Good, because `make verify` passes on it (measured).
* Bad, because it moves `go-sdk` ahead of mcplib's tested version.

### D-b: `golang.org/x/*` only

* Good, because those modules have the lowest risk.
* Bad, because it leaves `go-sdk` and `segmentio/asm` behind with no stated
  reason, which is only a partial update.

### D-c: no change

* Good, because nothing moves.
* Bad, because it does not do what was asked.

### L-a: relative links, permalinks to the described version, GitHub URLs across repositories

* Good, because every link resolves (measured: 0 problems, and all 22 new
  URLs resolve).
* Good, because the deleted and changed targets keep their meaning.
* Neutral, because the permalinks are long.

### L-b: relative links only

* Bad, because 40 links would be dropped: the 24 deleted `internal/selfupdate/`
  files, the 7 deleted workflows, and the 9 into other repositories. That
  loses evidence the records cite.

### L-c: GitHub `main` for everything

* Bad, because it breaks every deleted target, and moves the line anchors.

### I-a: mirror front matter

* Good, because the index stops contradicting the records.
* Bad, because 0004 visibly reads "Proposed" until the owner updates it.

### I-b: keep the index text

* Bad, because the index keeps contradicting two records.

### I-c: change the records

* Bad, because it decides record statuses without the owner.

## More Information

* **Plan:**
  [0007-PLAN-dependency-and-docs-link-refresh.md](0007-PLAN-dependency-and-docs-link-refresh.md).
* **Follow-ups outside this record:**
  * bump `go-sdk` in mcplib, so its own tests cover `v1.8.0`;
  * set the 0004 MADR's status, which the owner decides;
  * optionally add the link check to `make verify`.

## Amendment 2026-10-03: the dependency half is superseded

[0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md](0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md)
(D7) removes mcplib and, with it, the MCP go-sdk and the other indirect
modules this record would refresh. Its dependency decision (D-a) is
superseded and is not carried out. The link (L-a) and index (I-a) decisions
stand as proposed. This record stays `proposed` for them.
