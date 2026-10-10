# Documentation

This tree holds three kinds of document:

- reports that measured something;
- the decisions behind the work;
- guides you follow to do a thing.

- **[architecture.md](architecture.md)** — how the pieces fit together, as they are now.
- **[guides/](guides/)** — task-shaped operational walkthroughs.
- **[decisions/](decisions/)** — numbered MADR/PLAN pairs.
- **[reports/](reports/)** — numbered observations that decide nothing.

## Records

| Number | Kind | Record | Status |
| :--- | :--- | :--- | :--- |
| 0001 | MADR | [Gemini provider and model catalog modernization](decisions/0001-MADR-gemini-provider-and-model-catalog-modernization.md) | proposed |
| 0001 | PLAN | [Implement Gemini provider modernization](decisions/0001-PLAN-gemini-provider-and-model-catalog-modernization.md) | ready for review |
| 0002 | MADR | [Self-update CLI and GitHub Releases](decisions/0002-MADR-self-update-cli-and-github-releases-integration.md) | proposed |
| 0002 | PLAN | [Implement self-update CLI and GitHub Releases](decisions/0002-PLAN-self-update-cli-and-github-releases-integration.md) | completed |
| 0003 | MADR | [Layer and harden CI/CD quality gates](decisions/0003-MADR-layer-and-harden-ci-cd-quality-gates.md) | accepted |
| 0003 | PLAN | [Implement layered CI/CD quality gates](decisions/0003-PLAN-layer-and-harden-ci-cd-quality-gates.md) | completed |
| 0004 | MADR | [Align CI/CD with magic-cli-remote](decisions/0004-MADR-align-cicd-with-magic-cli-remote.md) | proposed |
| 0004 | PLAN | [Implement CI/CD alignment](decisions/0004-PLAN-align-cicd-with-magic-cli-remote.md) | in progress |
| 0005 | MADR | [Windows compatibility pre-commit suite](decisions/0005-MADR-windows-on-demand-compat-tests.md) | rejected |
| 0005 | PLAN | [Implement Windows compatibility pre-commit suite](decisions/0005-PLAN-windows-on-demand-compat-tests.md) | rejected |
| 0006 | MADR | [Canary mcplib v1.6.0-rc1](decisions/0006-MADR-mcplib-1-6-canary.md) | accepted |
| 0006 | PLAN | [Implement the mcplib canary](decisions/0006-PLAN-mcplib-1-6-canary.md) | complete |
| 0007 | MADR | [Refresh dependencies and repair documentation links](decisions/0007-MADR-dependency-and-docs-link-refresh.md) | accepted |
| 0007 | PLAN | [Implement dependency and link refresh](decisions/0007-PLAN-dependency-and-docs-link-refresh.md) | complete |
| 0008 | MADR | [Adopt go-llmprovider-sdk and go-selfupdate-lib](decisions/0008-MADR-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md) | accepted |
| 0008 | PLAN | [Implement the provider and self-update library adoption](decisions/0008-PLAN-adopt-go-llmprovider-sdk-and-go-selfupdate-lib.md) | complete |
| 0009 | MADR | [Adopt go-llmprovider-sdk v1.2.1](decisions/0009-MADR-adopt-go-llmprovider-sdk-v1-2-1.md) | accepted |
| 0009 | PLAN | [Implement go-llmprovider-sdk v1.2.1](decisions/0009-PLAN-adopt-go-llmprovider-sdk-v1-2-1.md) | complete |
| 0010 | MADR | [Adopt go-selfupdate-lib v1.9.0](decisions/0010-MADR-adopt-go-selfupdate-lib-v1-9-0.md) | accepted |
| 0010 | PLAN | [Implement go-selfupdate-lib v1.9.0](decisions/0010-PLAN-adopt-go-selfupdate-lib-v1-9-0.md) | complete |
| 0011 | MADR | [Adopt provider SDK v1.3.2 and self-update library v1.10.1](decisions/0011-MADR-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md) | accepted |
| 0011 | PLAN | [Implement provider SDK v1.3.2 and self-update library v1.10.1](decisions/0011-PLAN-adopt-go-llmprovider-sdk-v1-3-2-and-go-selfupdate-lib-v1-10-1.md) | superseded |
| 0012 | MADR | [Adopt Go 1.27.2 and go-selfupdate-lib v1.13.0](decisions/0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md) | accepted |
| 0012 | PLAN | [Implement Go 1.27.2 and go-selfupdate-lib v1.13.0](decisions/0012-PLAN-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md) | complete |
| 0013 | MADR | [Adopt the fleet workspace scaffold](decisions/0013-MADR-align-repository-with-workspace-scaffolding.md) | accepted |
| 0013 | PLAN | [Implement the fleet workspace scaffold](decisions/0013-PLAN-align-repository-with-workspace-scaffolding.md) | in progress |
| 0013 | GATES | [Workspace-scaffolding verification](reports/0013-GATES-workspace-scaffolding.md) | in progress |

## I want to…

| I want to… | Start here |
| :--- | :--- |
| understand how this fits together | [architecture.md](architecture.md) |
| install and configure the commit-message hook | [root README](../README.md#quick-start--installation) |
| run developer checks | [CI/CD operations guide](guides/cicd-operations.md#routine-verification) |
| perform or recover a release | [CI/CD operations guide](guides/cicd-operations.md#controlled-release) |
| inspect repository settings procedures | [CI/CD operations guide](guides/cicd-operations.md#repository-settings-audit) |
| know why the workspace and hook policy changed | [0013 MADR](decisions/0013-MADR-align-repository-with-workspace-scaffolding.md) |
| follow repository workflow rules | [AGENTS.md](../AGENTS.md) |
