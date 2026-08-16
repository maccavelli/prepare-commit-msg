# Documentation & Architectural Decisions

This directory contains design documents, architectural decision records (MADR), and technical implementation plans for [`prepare-commit-msg`](file:///home/mac/gitrepos/prepare-commit-msg/README.md).

## Architectural Decision Records (MADR)

* [0001-MADR: Modernizing Gemini Provider, Model Catalog Integration, and Configure UX](file:///home/mac/gitrepos/prepare-commit-msg/docs/decisions/0001-MADR-gemini-provider-and-model-catalog-modernization.md)
  * **Status:** Proposed / Under Review
  * **Topic:** Gemini model catalog accuracy, fast-model curation for commit messages, official `google.golang.org/genai` Go SDK evaluation, and ergonomic CLI configure wizard design.

* [0002-MADR: Self-Update CLI Subcommand and GitHub Releases Integration](file:///home/mac/gitrepos/prepare-commit-msg/docs/decisions/0002-MADR-self-update-cli-and-github-releases-integration.md)
  * **Status:** Proposed / Under Review
  * **Topic:** Native `update` CLI subcommand, GitHub Releases API integration, SHA-256 integrity verification, and cross-platform in-place atomic binary replacement.

## Implementation Plans

* [0001-PLAN: Gemini Provider Modernization Implementation Plan](file:///home/mac/gitrepos/prepare-commit-msg/docs/plans/0001-PLAN-gemini-provider-and-model-catalog-modernization.md)
  * **Status:** Ready for User Review
  * **Topic:** Step-by-step execution plan across `mcplib` and `prepare-commit-msg` with automated and manual verification strategies.

* [0002-PLAN: Self-Update CLI Subcommand and GitHub Releases Integration](file:///home/mac/gitrepos/prepare-commit-msg/docs/plans/0002-PLAN-self-update-cli-and-github-releases-integration.md)
  * **Status:** Ready for User Review
  * **Topic:** Phased execution plan for `internal/selfupdate`, SemVer, GitHub client, cross-platform atomic binary swaps, and CLI integration.
