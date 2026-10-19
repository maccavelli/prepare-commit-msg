---
status: accepted
date: 2026-10-19
decision-makers: Project Owner
consulted: 0015-PLAN-adopt-go-llmprovider-sdk-v1-15-0.md (Phase 5's live retry check, Deviation 5); go-llmprovider-sdk 0035-MADR-maintainer-tools-in-python.md; a probe of `make -n tools` on a scratch clone at 75a5b1e, 2026-10-19
informed: maintainers and agents working in this repository
---
# Keep maintainer tools in `tools/`, starting with a fake provider that answers 429

## Context and Problem Statement

Every claim below is marked **[read]** (from this repository at `75a5b1e`,
or the record named), **[probed]** (run on a scratch clone, never in this
tree) or **[derived]**.

0015-PLAN's Phase 5 checked the retry line live: the built hook, configured
against a local server that answers the first generation with HTTP 429, must
write one `retrying in …` line and then the server's message **[read]**.
That server was a script in an agent's session scratchpad:

* **It was written twice.** The first version answered `/chat/completions`
  in the Responses shape, so every model failed with "the answer has no
  content" after the retry (0015-PLAN's Phase 5 record,
  `0015-PLAN-adopt-go-llmprovider-sdk-v1-15-0.md:480`) **[read]**. Nothing
  had tested the fixture itself.
* **It is gone.** The scratchpad was cleared after the release's live
  checks. The next release that touches retries, the provider path or the
  wizard's endpoints needs it again **[derived]**.
* **This repository has no place for it.** `scripts/` holds what `make` and
  CI run: the hook installer, the GitHub audit, `check_records.py`,
  `go-precheck.py`, and `verify-scripts.sh` **[read]**. There is no
  `tools/` directory.

go-llmprovider-sdk made the same choice for the same reason: its
`0035-MADR-maintainer-tools-in-python.md` keeps the helpers a maintainer or
an agent runs, and CI does not, in `tools/`, each tested **[read]**.

Two facts about this repository bear on the name:

* **`make tools` already exists.** It installs the pinned linters under
  `.tools/bin` (`Makefile:82-83`) **[read]**.
* **It survives a `tools/` directory only because it is phony.** On a
  scratch clone with an empty `tools/` directory, `make -n tools` prints
  `./scripts/bootstrap-tools.sh`; with `tools` taken out of `.PHONY`
  (`Makefile:27-30`), it prints "`tools' is up to date" and runs nothing
  **[probed]**.

`scripts/verify-scripts.sh` parses every file under `scripts/` and
`.githooks/`: Python when its first line names Python, else `bash -n`
(`scripts/verify-scripts.sh:8-16`) **[read]**. It does not look in a
`tools/` directory. `.gitignore` already ignores `__pycache__/` and
`*.py[cod]` **[read]**.

## Decision Drivers

* A helper a release needs, and that went wrong once untested, is kept and
  tested.
* What CI runs stays in `scripts/`; what a person or an agent runs by hand
  goes elsewhere, as in go-llmprovider-sdk.
* Python 3 on the standard library, as `scripts/check_records.py` is. No
  host, account or machine path; anything outside the repository is an
  argument.
* The fake binds loopback only, and holds no real credential.

## Considered Options

* **A.** A `tools/` directory, with `tools/fakeprovider.py`, its test, and
  a README.
* **B.** `scripts/fakeprovider.py`, beside what CI runs.
* **C.** A Go fake under `internal/`, started by a test.
* **D.** Leave it in the scratchpad.

## Decision Outcome

Chosen option: "A", because the fake is a maintainer's tool for a live
check of the built binary, not a CI gate or a unit-test helper, and a
tested copy in the repository is what the 0015 run lacked.

### D1. `tools/`

* `tools/` holds the tools a maintainer or an agent runs and CI does not.
  `tools/README.md` lists each: what it does, and how it is run.
* Each tool: Python 3 on the standard library; a `#!/usr/bin/env python3`
  first line; the repository root found from its own file; exit codes 0
  clean, 1 a finding, 2 a tool error. It names no host, account or path of
  this machine.
* `make tools-test` runs `python3 -B -m unittest discover -s tools -p
  'test_*.py'`, and `make verify`, which CI runs, runs it (Q1).
  `scripts/verify-scripts.sh` parses `tools/` too.
* `tools` stays in `.PHONY`; `tools/README.md` says why.

### D2. `tools/fakeprovider.py`

* **It serves,** on `127.0.0.1` only, an OpenAI-compatible API:
  * `GET …/models`: one model;
  * `POST …/chat/completions`: a chat completion, in that API's shape;
  * `POST …/responses`: a Responses answer, in that API's shape.
* **It fails first.** The first `--fail N` generations (default 1) are
  answered with `--status` (default 429) and `Retry-After: --retry-after`
  (default 1). Every later one is answered with `--message`.
* **It reports** its port to `--port-file`, and one line per request to
  `--log`: `METHOD PATH -> STATUS`. It never logs a header or a body, so a
  key the hook sends is never written.
* **It stops** on an interrupt, or after `--max-requests`.
* **As a module,** `serve(...)` returns the running server, for its test.

### D3. The test

`tools/test_tools.py` starts the fake in-process and checks:

* each endpoint answers in its API's shape, parsed as the SDK's provider
  would read it;
* the first `N` generations get the status and `Retry-After`, then the
  message;
* the log holds the method, path and status, and no header.

It fails on a planted break: a chat completion in the Responses shape, the
fault of the 0015 run.

### D4. The documents

* `docs/architecture.md`'s tree names `tools/`.
* `AGENTS.md` gains a "Tools" section: what `tools/` is, and that a
  release's live retry check uses `tools/fakeprovider.py`.
* `docs/README.md` indexes this record and its plan.

### Consequences

* Good, because the live retry check is one command, against a fixture that
  is itself tested.
* Good, because later helpers have a place, and a rule.
* Neutral, because `make tools` (the linters) and `tools/` (the scripts)
  share a name; the README says which is which.
* Bad, because `make tools`'s recipe now depends on `tools` staying in
  `.PHONY`.

### Confirmation

* `make tools-test` passes, and fails on the planted Responses-shaped chat
  completion, run on a scratch clone.
* `make -n tools` still prints `./scripts/bootstrap-tools.sh` with
  `tools/` present.
* `make verify` passes: `verify-scripts.sh` parses `tools/`.
* A live run: `configure --yes --provider together --base-url` the fake,
  then the hook, writes one retry line and the fake's message.

## Pros and Cons of the Options

### A. `tools/`

* Good, because it separates what CI runs from what a person runs, as the
  SDK does.
* Bad, because of the name it shares with `make tools`.

### B. `scripts/`

* Good, because there is no new directory.
* Bad, because `scripts/` would mix CI's gates with a maintainer's fixture,
  and its README and rules are about the former.

### C. A Go fake under `internal/`

* Good, because Go tests already use `httptest` this way.
* Bad, because the live check runs the built binary from a shell, not a
  test: it needs a server that runs on its own.

### D. Leave it in the scratchpad

* Bad, because it is already gone once.

## Owner questions

* **Q1.** Should `make verify`, and so CI, run `make tools-test`? Proposed:
  yes. The test is fast and offline, and a broken fixture is found before a
  release needs it. **Answered 2026-10-19: yes.** `verify` gains
  `tools-test` (D1).

## More Information

* The live check it serves: 0015-PLAN's Phase 5 step 3, and its
  Deviation 5 (`together`, since this app gives OpenAI no endpoint).
* go-llmprovider-sdk's `0035-MADR-maintainer-tools-in-python.md`, and its
  `0047-MADR-keep-five-more-scratch-helpers-as-tools.md`, proposed the same
  day, for the SDK's own tools.
