---
status: complete
date: 2026-10-19
associated-madr: "0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md"
---
# Implement keeping maintainer tools in `tools/`, starting with a fake provider that answers 429

Associated MADR: [0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md](0016-MADR-keep-maintainer-tools-starting-with-a-fake-provider.md)

## Goal

* **D1:** `tools/` exists, with its README and rules, and `make tools-test`,
  which `make verify` runs.
* **D2:** `tools/fakeprovider.py` serves the 429-then-message API.
* **D3:** its test passes, and fails on the 0015 run's fault.
* **D4:** `docs/architecture.md` and `AGENTS.md` describe `tools/`.

## Scope

### In scope

* New: `tools/README.md`, `tools/fakeprovider.py`, `tools/test_tools.py`.
* Changed: `Makefile`, `scripts/verify-scripts.sh`, `docs/architecture.md`,
  `AGENTS.md`, `docs/README.md`.

### Out of scope

* Any Go file. CI's workflow file: it already runs `make verify`
  (`.github/workflows/ci.yml:49`).
* Other tools: each later one needs its own record.

## Rules for every phase

* **Branch.** On `main`. The agent stages; the owner commits and pushes.
* **Fail first.** Each new test is run on a scratch clone with its fix
  removed, and the failure is recorded here. The tree is never dirtied for
  it.
* **Gate, before staging:**
  * `make tools-test`;
  * `make verify` (which then runs it, and `verify-scripts.sh`);
  * `make check-records`, `make markdownlint`;
  * the identifier scan of the staged files.

  Each exit status is captured before any filter.
* **Deviations.** Stop, record a dated entry here (and in the MADR when a
  decision or a fact changes), and wait for approval.

## Implementation Steps

### Phase 0: records

1. 0016-MADR is `accepted`; this PLAN is indexed. `make check-records`.

### Phase 1: `tools/` and the fake provider (D1–D4)

1. **`tools/fakeprovider.py`,** with `#!/usr/bin/env python3`, a module
   docstring that is its definition, and:
   * `serve(fail=1, status=429, retry_after=1, message=..., log=None,
     max_requests=None)` returning a running `ThreadingHTTPServer` on
     `127.0.0.1` and port 0;
   * the three endpoints of MADR D2, each body in its API's shape; a
     generation counter under a lock;
   * one log line per request, `METHOD PATH -> STATUS`, and nothing else;
   * `main(argv)`: `--fail`, `--status`, `--retry-after`, `--message`,
     `--port-file`, `--log`, `--max-requests`; it writes the port, serves,
     and exits 0 on an interrupt or at `--max-requests`, 2 on a bad
     argument.
2. **`tools/test_tools.py`** (`unittest`, the standard library only), with
   the server in-process and requests by `urllib.request`:
   * `/v1/models`: one model;
   * `/v1/chat/completions` twice, with `fail=1`: a 429 with `Retry-After:
     1`, then `choices[0].message.content` equal to the message;
   * `/v1/responses`, after a failure: `output[0].content[0].text` equal to
     the message, `type` `output_text`;
   * the log: one line per request, in the form, and no header: a request
     sent with an `Authorization` header whose value is a made-up key, which
     must not appear in the log;
   * `max_requests`: the server stops after it;
   * `main` with a bad argument: exit 2.
3. **`tools/README.md`:** what `tools/` is (MADR D1); the rules; the table,
   one row; that `make tools` is the linters and `tools/` the scripts, and
   that `tools` stays in `.PHONY` (MADR, Context).
4. **`Makefile`:** `tools-test: ## Checks the maintainer tools in tools/`,
   running `$(PYTHON) -B -m unittest discover -s tools -p 'test_*.py'`; in
   `.PHONY` (`Makefile:27-30`); and `verify` (`Makefile:114`) gains
   `tools-test`.
5. **`scripts/verify-scripts.sh`:** also parses each `tools/*.py` as Python.
   Its loop runs `bash -n` on any file whose first line does not name
   Python (`scripts/verify-scripts.sh:8-16`), so `tools/` is listed by
   `find tools -name '*.py' -not -path '*/__pycache__/*'`, never as a
   whole directory: `tools/README.md` and a test run's `__pycache__` files
   would fail `bash -n`.
6. **`docs/architecture.md`:** the tree (`docs/architecture.md:59`) gains
   `tools/   maintainer tools, not run by the hook: tools/README.md`.
7. **`AGENTS.md`:** a section "Tools", after "Pre-add checks"
   (`AGENTS.md:104`): what `tools/` holds; Python on the standard library;
   `make tools-test`; and that a release's live retry check runs the hook
   against `tools/fakeprovider.py` (0015-PLAN's Phase 5 step 3, with
   Deviation 5).
8. **Fail first,** each on a scratch clone:
   * the chat completion answered in the Responses shape (the 0015 run's
     fault): the chat test fails;
   * the `Retry-After` header left out: the 429 test fails;
   * the log given the request's headers: the "no header" test fails;
   * `verify-scripts.sh` given the whole `tools/` directory: `make
     workflow-lint` fails on `tools/README.md`.
9. **`make -n tools`** still prints `./scripts/bootstrap-tools.sh` (MADR,
   Confirmation).
10. **A live run,** on a scratch build of the tree with a scratch `HOME`:
    the fake with `--fail 1 --port-file …`; `configure --yes --provider
    together --base-url http://127.0.0.1:<port>/v1` with a made-up
    `TOGETHER_API_KEY`; the hook on a scratch repository with one staged
    file. Its stderr holds one `retrying in` line, and the message is the
    fake's. The installed hook is not touched.

## Verification

| # | Criterion | Phase | Check |
| :--- | :--- | :--- | :--- |
| V1 | The fake answers each endpoint in its API's shape, fails first, and logs no header | 1 | `make tools-test` |
| V2 | The test fails on the 0015 fault, and each other plant | 1 | the fail-first record |
| V3 | `make verify` runs `tools-test` and parses `tools/*.py`, and CI passes | 1 | `make verify`; the CI run on `main` |
| V4 | `make tools` still runs the bootstrap | 1 | `make -n tools` |
| V5 | The hook's live retry line, against the fake | 1 | step 10's log |

## Rollout and Rollback

* Nothing reaches users: `tools/` is not built into the binary, and no
  release is needed.
* Rollback: revert the phase's commit. `make verify` then no longer runs
  `tools-test`.

## Execution Record

### Phase 0: records (2026-10-19)

* The MADR was accepted and this PLAN indexed, committed by the owner as
  `c15103a`; the owner then approved it ("approved, proceed").

### Phase 1: `tools/` and the fake provider (2026-10-19)

* **No deviation.**
* **Steps 1–7** as written:
  * `tools/fakeprovider.py`: `serve(...)` returns a running
    `FakeProvider` (a `ThreadingHTTPServer` on `127.0.0.1`, port 0); its
    generation counter and its request count are under one lock;
    `--max-requests` stops it from a thread, so a request is answered
    before the server shuts down.
  * `tools/test_tools.py`: nine tests; every request carries an
    `Authorization` header with a made-up key, which the log test requires
    absent.
  * `tools/README.md`; the `tools-test` target, in `.PHONY` and in
    `verify`; `verify-scripts.sh` parses `tools/*.py` in a loop of its own;
    the tree in `docs/architecture.md`; AGENTS.md's "Tools".
* **Step 8, fail first,** each on a fresh scratch copy of the working tree:
  * the chat completion in the Responses shape (the 0015 fault):
    `test_chat_completions_fail_first_then_answer` fails, `'response' !=
    'chat.completion'`;
  * `Retry-After` left out: the same test fails, `None != '1'`;
  * the log given the request's headers:
    `test_log_holds_method_path_status_and_no_header` fails;
  * `verify-scripts.sh` given the whole `tools/` directory: `make
    workflow-lint` exits 2, `tools/README.md: line 5: syntax error near
    unexpected token`. The trap of step 5 is real.
* **Step 9.** `make -n tools` prints `./scripts/bootstrap-tools.sh`.
* **Step 10, live,** on a scratch build of the tree with a scratch `HOME`:
  the fake with `--fail 1`; `configure --yes --provider together
  --base-url http://127.0.0.1:<port>/v1` with a made-up
  `TOGETHER_API_KEY`, exit 0; the hook on a scratch repository with one
  staged file, exit 0:

  ```text
  prepare-commit-msg: generating via together (zai-org/GLM-5.3-Flash)...
  prepare-commit-msg: together: retrying in 1.2s (attempt 1 of 3): rate limited
  ```

  The message was the fake's, `chore: add live retry check file`; the
  fake's log held `POST /v1/chat/completions -> 429`, then `-> 200`, and the
  key 0 times. The installed hook's SHA-256 was the same before and after.
* **The gate:** `make tools-test`: 9 tests, OK. `make verify`: exit 0 (0
  issues; coverage 83.7%; no vulnerabilities; the tools' tests). `make
  check-records`, `make markdownlint`: 0. The identifier scan of the staged
  files: no hit.
* **After the push.** The owner committed the phase as `5981bc9` and pushed,
  after the disclosure guard passed (exit 0). CI run `39277673444`:
  `success`, every job (the quality contract, the native tests on Windows,
  Linux and macOS, the release build and its five identity checks); the
  publish job skipped, as without a tag. The quality job's log holds
  `python3 -B -m unittest discover -s tools -p 'test_*.py'` and `Ran 9
  tests`. Two `gh run watch` sessions ended on GitHub API errors (HTTP 502)
  while the build jobs queued; the run itself never failed.
* **V1–V5 are met,** so this PLAN is `complete`.
