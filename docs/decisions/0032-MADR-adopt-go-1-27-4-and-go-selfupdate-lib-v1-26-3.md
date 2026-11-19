---
status: accepted
date: 2026-11-18
decision-makers: Project Owner
consulted: go-selfupdate-lib (its CHANGELOG for v1.26.1, v1.26.2 and v1.26.3, the compare v1.26.0...v1.26.3 by commit and file, its migration guide sections for the three releases, its 0043-MADR and 0045-MADR, and its source at v1.26.3); go-llmprovider-sdk (its go.mod at v1.29.0 and v1.31.0); a scratch copy of this repository at 2715f2d moved to both; all 2026-11-18
informed: users of the installed hook, who receive the change through `update` only when a release carries it; this record cuts none
---
# Move to Go 1.27.4 and go-selfupdate-lib v1.26.3

## Context and Problem Statement

The owner (2026-11-18), asked how the consumer bumps should proceed, chose
"One record pair per repository, approved now through CI; magic-cli-remote
after its lint phases (Recommended)". As put to the owner: this repository
moves from `go 1.27.3` to Go 1.27.4 and from go-selfupdate-lib `v1.26.0` to
`v1.26.3`; the move is committed and pushed to the default branch on a green
gate; the installed hook binary is not touched; no release is cut.

Each claim is marked:

* **[read]**: from this repository, or from a library's changelog, source or
  git objects at the version named;
* **[probed]**: run in a scratch copy on 2026-11-18, never in this tree;
* **[derived]**: reasoned from the two, and not run.

### Why Go 1.27.4

GO-2026-6859 ("HTTP/1 server crash due to MaxBytesReader header race in
`net/http`") is fixed in `net/http@go1.27.4`. `govulncheck ./...` on a scratch
copy of `2715f2d` under Go 1.27.3, with `GOPROXY=off` for modules, exits 3 and
reports it, with 38 example traces, among them `update.go:58`
(`updateHTTPClient` calls `selfupdate.DefaultHTTPClient`) and
`internal/ui/setup.go:274` (`wizard.ConfigureLLM`) **[probed]**. The
advisory is in the standard library, so only the toolchain a binary is built
with removes it **[read]**.

### Does the code need Go 1.27?

The `go` line cannot fall to Go 1.26. `go-llmprovider-sdk v1.29.0` declares
`go 1.27.3` in its own `go.mod` **[read]**, and this module, set to
`go 1.26.11` on a scratch copy, is refused with "updates to go.mod needed"
**[probed]**. The line is the floor and the toolchain at once: there is no
`toolchain` line, and `go mod tidy` drops one equal to the `go` line, so none
is added **[read]**. `go-selfupdate-lib v1.26.3`'s own floor, `go 1.26.11`,
is below this module's **[read]**.

### The release this move crosses

`go-selfupdate-lib v1.26.3` is an annotated tag (`10316bca…`) on
`1dffdc9dee24a7534513dc51c9039408be37c09b`; `git ls-remote --tags` peels it to
that commit **[probed]**. `v1.26.0...v1.26.3` holds 17 commits and changes 125
files, 54 of them non-test, non-doc Go, workflow or script files **[read]**.
No exported API changes: its migration guide says `make apicheck` reports each
of the three compatible with the one before. What a program can see:

* **`v1.26.1`: an error names the package once.** `selfupdate: <product>:
  selfupdate: <reason>` becomes `selfupdate: <product>: <reason>`, for every
  error the module wraps behind a literal prefix, `ErrCheckDeferred` among
  them. The `--json` result's `error` field changes with it. `errors.Is` and
  `errors.As` match as before.
* **`v1.26.2`: a managed install's failure names its phrase once.** This
  concerns `ReplaceBeforeStop` and the `service` backends, which this
  application does not use.
* **`v1.26.3`: Go 1.27.4 and a `go 1.26.11` floor,** for GO-2026-6859, and the
  publish workflow builds its release tools with Go 1.27.4 for a caller that
  pins `v1.26.3`.
* `.github/workflows/build-selfupdate-release.yml` and
  `publish-selfupdate-release.yml` are unchanged between `v1.26.0` and
  `v1.26.3`; only the library's own `ci.yml` and one script differ
  (`git diff --stat v1.26.0 v1.26.3 -- .github scripts`) **[read]**.

### What reaches this app

The packages this repository imports are `buildinfo`, `selfupdate`,
`selfupdate/cli`, `selfupdate/releasespec` and `selfupdate/selfupdatetest`
**[read]**.

* **No code change is needed.** The application matches no library message
  text: no `.go` file and no golden file under `testdata/migration` holds a
  doubled `selfupdate:` prefix or a managed-install phrase **[read]**.
* **The one visible change is message text** on a failed `update`: a
  refusal prints `update failed: selfupdate: prepare-commit-msg: <reason>`
  without the second `selfupdate:` that `v1.26.0` printed after the product,
  the line 0031-MADR Amendment A1 recorded **[derived]**.
* **The scratch move:** in a scratch copy of `2715f2d` with `GOPROXY=off`,
  `go get github.com/maccavelli/go-selfupdate-lib@v1.26.3` and `go mod tidy`
  change one line of `go.mod` and two of `go.sum`, download nothing, and
  move no other module. The three `golang.org/x` modules stay at the versions
  `v1.26.3` requires, which are the ones already in `go.mod`
  **[probed]**.

## Decision Drivers

* **A known advisory is closed** by the toolchain the module names.
* **Pins are current** unless a record names a reason to hold one.
* **The workflow pins name the version `go.mod` requires**
  (0031-MADR D2; `0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md` D2, A1).
* **The Go version is stated in every place that is live.** A pin that
  `scripts/bootstrap-tools.sh` checks against `go env GOVERSION` fails the
  gate when it lags.
* **The change is a pin move.** The installed hook is not touched, and no
  release is cut.

## Considered Options

* **A. Move the Go line, the library requirement, both workflow pins and every
  live statement of the Go version; change no code; cut no release.**
* **B. As A, but move the modules only** and leave both workflow pins at
  `v1.26.0`'s commit.
* **C. As A, and also move go-llmprovider-sdk** to its newest tag.
* **D. Hold until a release is wanted,** and move the pins with it.

## Decision Outcome

Chosen option: "A", because it closes GO-2026-6859 now, makes the library
requirement and its workflow pins agree, needs no code, and leaves the
release decision to the owner. B leaves `ci.yml`'s comment false and the
publish workflow's release tools on Go 1.27.3. C is another library's
decision and crosses a new module (`go-llmprovider-sdk v1.31.0` requires
`go-llmprovider-sdk/authcontract v0.1.0`) **[read]**. D leaves the
advisory open on `main` for no stated reason.

### D1. The Go line moves

* `go.mod` reads `go 1.27.4`, with no `toolchain` line.
* `scripts/bootstrap-tools.sh` names `go1.27.4`, so `make tools` accepts the
  toolchain and rebuilds the three tools with it.
* `Makefile`'s `MOD_VERSION` reads `1.27.4`.
* `README.md`'s "Building from source needs Go" sentence reads `1.27.4`.
* History stays: earlier records, and the README sections that describe the
  releases built with Go 1.27.2 and 1.27.3, are not rewritten.

### D2. The library requirement moves

* `go.mod` requires `github.com/maccavelli/go-selfupdate-lib v1.26.3`,
  after `go mod tidy`.
* `go.mod` changes on that line and the `go` line alone, and `go.sum` on that
  module's two lines alone. `go mod tidy -diff` is empty and `go mod verify`
  passes.

### D3. Both workflow pins move

* Both `uses:` lines in `.github/workflows/ci.yml` pin
  `1dffdc9dee24a7534513dc51c9039408be37c09b # v1.26.3`.
* The comment above the build job names `v1.26.3` and cites this record and
  0031-MADR D2.
* `docs/guides/cicd-operations.md` quotes the new lines, in the same commit
  (`documentation_test.go` reads them).
* The `release` job's `with:` and permissions are unchanged.

### D4. No code change, no release, no installed binary

* No `.go` file changes.
* No README section "Changes with …" is added, because the README says what
  a user of a release can see and no release carries this move yet. The
  next release's record adds it.
* No tag is made. Nothing is installed: `make install`, an installer script
  and `go install` of this module are not run, and no hooks directory is
  written.

### Consequences

* Good, because GO-2026-6859 no longer reaches the module's `govulncheck`
  scan, and a binary built from `main` carries the fixed `net/http`.
* Good, because the workflow pins, the module requirement and the comment
  that says they agree are all true.
* Neutral, because a failed `update` prints one `selfupdate:` fewer, which
  this application's tests do not match.
* Bad, because a developer with Go 1.27.3 as the only toolchain must move to
  1.27.4: `make tools` refuses a different `go env GOVERSION`.
* Bad, because `go-llmprovider-sdk` stays at `v1.29.0` while `v1.31.0` is
  tagged; that is a separate record.

### Confirmation

* `go.mod` changes on the two lines D1 and D2 name, `go.sum` on D2's two
  lines. `go mod tidy -diff` is empty, and `go mod verify` passes.
* `1dffdc9dee24a7534513dc51c9039408be37c09b` occurs twice in `ci.yml` and
  twice in the operations guide, and `7c677842dface4b14b65b2e8b46aa0a7543bed13`
  in neither.
* `govulncheck ./...` on the changed tree under Go 1.27.4 reports no
  vulnerability.
* No `.go` file changes.
* `make check-records`, `make markdownlint`, `make disclosure-check` and
  `make verify` pass.

## Pros and Cons of the Options

### A. Move everything, no code, no release

* Good, because it is the smallest change that makes every live statement
  agree.
* Bad, because the pin move reaches users only with a later release.

### B. Move the modules only

* Good, because CI's workflows do not change.
* Bad, because `ci.yml`'s comment says the pins name the version `go.mod`
  requires, and would be false.

### C. Also move go-llmprovider-sdk

* Good, because it makes the SDK current.
* Bad, because `v1.31.0` adds a module to `go.sum` and carries its own
  changes, which a Go pin move should not decide.

### D. Hold until a release

* Good, because nothing lands without a release.
* Bad, because `main` builds with a toolchain that has a reachable advisory.

## More Information

* The last move of go-selfupdate-lib:
  [0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md](0031-MADR-adopt-go-llmprovider-sdk-v1-29-0-and-go-selfupdate-lib-v1-26-0.md).
  Its Go-version statements and its `v1.26.0` workflow pins are history; this
  record moves them.
* The last move of the Go version:
  [0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md](0018-MADR-adopt-go-llmprovider-sdk-v1-18-0-and-go-selfupdate-lib-v1-17-0.md).
* The workflow-pin rule:
  [0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md](0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md).
* go-selfupdate-lib's records, cited by repository and filename:
  `0043-MADR-name-the-package-once-in-wrapped-errors.md` and
  `0045-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md`.
* The implementation:
  [0032-PLAN-adopt-go-1-27-4-and-go-selfupdate-lib-v1-26-3.md](0032-PLAN-adopt-go-1-27-4-and-go-selfupdate-lib-v1-26-3.md).
