# Documentation & Architectural Decisions

This directory contains design documents, architectural decision records (MADR), and technical implementation plans for [`prepare-commit-msg`](../README.md).

## Architectural Decision Records (MADR)

* [0001-MADR: Modernizing Gemini Provider, Model Catalog Integration, and Configure UX](decisions/0001-MADR-gemini-provider-and-model-catalog-modernization.md)
  * **Status:** Proposed / Under Review
  * **Topic:** Gemini model catalog accuracy, fast-model curation for commit messages, official `google.golang.org/genai` Go SDK evaluation, and ergonomic CLI configure wizard design.

* [0002-MADR: Self-Update CLI Subcommand and GitHub Releases Integration](decisions/0002-MADR-self-update-cli-and-github-releases-integration.md)
  * **Status:** Proposed / Under Review
  * **Topic:** Native `update` CLI subcommand, GitHub Releases API integration, SHA-256 integrity verification, and cross-platform in-place atomic binary replacement.

* [0003-MADR: Layer and Harden CI/CD Quality Gates](0003-MADR-layer-and-harden-ci-cd-quality-gates.md)
  * **Status:** Accepted
  * **Topic:** Repository-owned verification gates, multi-platform native testing, and local hook architecture.

* [0004-MADR: Align CI/CD Workflow with magic-cli-remote](0004-MADR-align-cicd-with-magic-cli-remote.md)
  * **Status:** Proposed
  * **Topic:** Unified CI/CD workflow, tag-driven release builds and direct GitHub release publication.

* [0005-MADR: Windows compatibility suite as a Python pre-commit hook, skipped on Unix](0005-MADR-windows-on-demand-compat-tests.md)
  * **Status:** Proposed
  * **Topic:** On-demand Windows compatibility tests, run as a pre-commit hook on Windows and skipped elsewhere.

* [0006-MADR: Canary mcplib v1.6.0-rc1 in prepare-commit-msg, Reading Vendor CLI Logins Through](decisions/0006-MADR-mcplib-1-6-canary.md)
  * **Status:** Accepted
  * **Topic:** First consumer of mcplib v1.6.0: CLI logins read through, hermetic model listing, the release-workflow pin, and the canary that promoted mcplib v1.6.0.

* [0007-MADR: Refresh Dependencies to Latest and Repair Every Documentation Link](decisions/0007-MADR-dependency-and-docs-link-refresh.md)
  * **Status:** Accepted
  * **Topic:** Portable links in place of machine paths, and an index that mirrors each record's status. Its dependency refresh is superseded by 0008.

* [0008-MADR: Drop mcplib for go-llmprovider-sdk v1.0.0 and go-selfupdate-lib v1.5.0](decisions/0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md)
  * **Status:** Accepted
  * **Topic:** Providers and the wizard from go-llmprovider-sdk, self-update and build stamps from go-selfupdate-lib, the supply-chain gate, and what a user notices.

* [0009-MADR: Move to go-llmprovider-sdk v1.2.1, and Keep Refusals, Entitlement Errors and `configure --yes` Working](decisions/0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md)
  * **Status:** Accepted
  * **Topic:** The SDK bump from v1.0.0, what it changes for the hook, refusals and "not permitted" moving to the next model, the curated fallback for `configure --yes`, and a README table checked against the SDK.

* [0010-MADR: Move to go-selfupdate-lib v1.9.0, and Keep Publishing on the Proven Workflow](decisions/0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md)
  * **Status:** Accepted
  * **Topic:** The library bump from v1.5.0, what `update` does differently, the publish workflow kept at v1.5.0's commit until v1.9.0's has published a release, and a test for `update --json`'s schema version.

* [0011-MADR: Move to go-llmprovider-sdk v1.3.2 and go-selfupdate-lib v1.10.1, and Publish Through the Newest Workflow with a Live Publish on Record](decisions/0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md)
  * **Status:** Accepted
  * **Topic:** Both library bumps, what the SDK's v1.3 changes for the hook (a refused Gemini key stops the run), `update`'s kept backups and refused overwrites, the publish pin moved to v1.10.0's workflow, the newest with a live publish, and a test that sends Gemini's real refusal through the SDK.

* [0012-MADR: Move to Go 1.27.2 and go-selfupdate-lib v1.13.0, and Build, Publish and Install through the Library's Release Pipeline](decisions/0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md)
  * **Status:** Accepted
  * **Topic:** Go 1.27.1's advisories in the gates and the shipped hook, the `go` line at 1.27.2, the library at v1.13.0 (schema 4), both workflows pinned at v1.13.0, the build workflow and an identity command, installers that leave `core.hooksPath` to the user, tool pins, and the rulesets, README and `testfile.txt` findings.

## Implementation Plans

* [0001-PLAN: Gemini Provider Modernization Implementation Plan](plans/0001-PLAN-gemini-provider-and-model-catalog-modernization.md)
  * **Status:** Ready for User Review
  * **Topic:** Step-by-step execution plan across `mcplib` and `prepare-commit-msg` with automated and manual verification strategies.

* [0002-PLAN: Self-Update CLI Subcommand and GitHub Releases Integration](plans/0002-PLAN-self-update-cli-and-github-releases-integration.md)
  * **Status:** Completed
  * **Topic:** Phased execution plan for `internal/selfupdate`, SemVer, GitHub client, cross-platform atomic binary swaps, and CLI integration.

* [0003-PLAN: Layer and Harden CI/CD Quality Gates](0003-PLAN-layer-and-harden-ci-cd-quality-gates.md)
  * **Status:** Completed
  * **Topic:** Phased execution plan for local quality contracts, multi-platform native tests, and hook management.

* [0004-PLAN: Align CI/CD Workflow with magic-cli-remote](0004-PLAN-align-cicd-with-magic-cli-remote.md)
  * **Status:** In Progress
  * **Topic:** Consolidated single CI/CD workflow, automated tag builds, and direct GitHub release publication.

* [0005-PLAN: Windows compatibility suite as a Python pre-commit hook](0005-PLAN-windows-on-demand-compat-tests.md)
  * **Status:** Proposed
  * **Topic:** The Windows hook, its Makefile target, and how it is skipped on Unix.

* [0006-PLAN: Implement the mcplib v1.6.0-rc1 Canary](decisions/0006-PLAN-mcplib-1-6-canary.md)
  * **Status:** Completed
  * **Topic:** Phases S0–S8: fix `main`, adopt the candidate, release v1.4.0, canary on macOS and Windows, promote mcplib v1.6.0.

* [0007-PLAN: Implement the Dependency and Documentation Link Refresh](decisions/0007-PLAN-dependency-and-docs-link-refresh.md)
  * **Status:** Completed
  * **Topic:** Phases R0–R5: links rewritten to relative paths and pinned GitHub URLs, and the index made to mirror front matter. R1 is superseded.

* [0008-PLAN: Implement dropping mcplib](decisions/0008-PLAN-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md)
  * **Status:** Completed
  * **Topic:** Phases 0–6: self-update, providers, the supply-chain gate, docs, the release and live check, and close-out.

* [0009-PLAN: Implement the Move to go-llmprovider-sdk v1.2.1](decisions/0009-PLAN-adopt-go-llmprovider-sdk-v1-2-1.md)
  * **Status:** Completed
  * **Topic:** Phases 0–5: the adaptations on v1.0.0, the bump, the README and its test, release v1.6.0 with the live check, and close-out.

* [0010-PLAN: Implement the Move to go-selfupdate-lib v1.9.0](decisions/0010-PLAN-adopt-go-selfupdate-lib-v1-9-0.md)
  * **Status:** Completed
  * **Topic:** Phases 0–4: the bump, the pin comment and the schema test, the README, release v1.7.0 with the live check, and close-out.

* [0011-PLAN: Implement the Move to go-llmprovider-sdk v1.3.2 and go-selfupdate-lib v1.10.1](decisions/0011-PLAN-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md)
  * **Status:** Superseded by 0012-PLAN
  * **Topic:** Phases 0–4: the bumps, the pin and the refused-key test, the README, release v1.8.0 with the live checks, and close-out.

* [0012-PLAN: Implement the Move to Go 1.27.2 and go-selfupdate-lib v1.13.0, and the Library's Build, Publish and Install Pipeline](decisions/0012-PLAN-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md)
  * **Status:** In Progress
  * **Topic:** Phases 0–6: the records and 0011-PLAN's carried release, the toolchain, library and tools, the spec, `identity` and both workflows at v1.13.0, the installers, hygiene and the rulesets, release v1.8.0 with the live checks, and close-out.
