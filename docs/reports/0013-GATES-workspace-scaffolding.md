---
status: in-progress
date: 2026-10-10
associated-plan: "0013-PLAN-align-repository-with-workspace-scaffolding.md"
---

# Workspace-scaffolding verification

Associated PLAN:
[0013-PLAN-align-repository-with-workspace-scaffolding.md](../decisions/0013-PLAN-align-repository-with-workspace-scaffolding.md)

## Baseline

Execution started from Phase 0 commit `7f883c2`. Phase 1 committed LF and
ignore policy as `013e650`. The current sequence ends at 0013 and the expected
next number is 0014.

## Template hashes

All sixteen reviewed SHA-256 values in the PLAN matched the canonical
workspace templates before execution. Revalidation at final close-out is
pending.

## Copied-file comparisons

| Target | Result |
| :--- | :--- |
| `.markdownlint-cli2.jsonc` | `cmp` exit 0 |
| `scripts/check_records.py` | `cmp` exit 0 |
| Eight Phase 3 harness support files | pending |

## Moves and links

The unrepaired scratch move exited 1 with exactly 25 broken relative links.
Nine production moves place all numbered records under `docs/decisions/` and
the runbook under `docs/guides/`. The records table contains all 27 records;
there are no numbered records at the docs root. Record and unnumbered-document
link checks exit 0 with no placement warning.

The reviewed residual old-path allowlist consists of earlier records that
describe the former tree, rejected 0005 work, and commands that actually ran.
The old runbook text in the 0004 PLAN is a historical link label whose target
is repaired. Live script and guide references use the new paths.

## Scratch plants

| Gate | Observation |
| :--- | :--- |
| LF checkout | Without repository attributes, `core.autocrlf=true` produced CRLF; with the rule it produced LF; migration fixtures reported `text: unset`. |
| Legacy installer dependency | Current installer without a repository pre-commit exited 1 with `repository hook is not executable:`. |
| Retired repository pre-commit | Restoring `install_wrapper pre-commit` in a scratch copy made the revised hook test exit 1 with `repository hook is not executable:`. |
| Carried-forward hook comparison | Removing the generic wrapper comparison in a scratch installer made the suite exit 1 with `installer overwrote a modified carried-forward hook`. |
| No-host pre-commit assertion | Planting a managed pre-commit in the scratch no-host case made the suite exit 1 with `installer created a managed pre-commit without a host pre-commit`. |
| Records links | A missing link in a scratch root README made `--check-all` exit 1 with `broken relative link:`. |
| Markdown MD004 | pending |
| Markdown MD013 in README | pending |
| Markdown MD013 in AGENTS.md | pending |
| Record exclusion | pending |
| Identifier scan | pending |

## Actual-tree gates

| Gate | Result |
| :--- | :--- |
| Phase 1 attributes | 81 tracked paths inspected; zero noncanonical results |
| `bash -n scripts/install-hooks.sh scripts/test-hooks.sh` | exit 0 |
| `shellcheck scripts/install-hooks.sh scripts/test-hooks.sh scripts/configure-github.sh` | exit 0 |
| `make hooks-test` | exit 0; fresh install, idempotence, legacy migration, conditional host pre-commit, refusal, pre-push input, and uninstall cases passed |
| This clone's managed pre-commit | Absent, because the saved host hooks directory has no executable pre-commit; approved conditional behavior |
| Records gates | `--next` exit 0 with `0014`; `--check-all` and `make check-records` exit 0 with no output beyond the Make recipe |
| Markdown gates | Installed v0.23.2 binary and `make markdownlint` exit 0; five files linted, zero issues |
| `make verify` | Authorized run exit 0; modules verified, zero lint issues, vet and race tests passed, 81.7% coverage, no vulnerabilities, workflow checks and six cross-builds passed |

## Identifier scan

Pending Phase 3, when all generated and copied files exist.

## Product checks

The complete `make verify` contract passed on the final Phase 2 tree. Its
initial sandboxed run could not access the external Go build cache; the
authorized run exited 0. The repository pre-push hook and CI continue to
invoke `make verify`; `make verify-staged` remains available on demand.

## Remaining work

Commit the passing Phase 2 tree, install and plant-test the Phase 3 agent
files, then close out this record.
