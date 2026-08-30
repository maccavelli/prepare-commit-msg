---
status: proposed
date: 2026-08-30
decision-makers:
  - Maintainers of prepare-commit-msg
---

# Align CI/CD Workflow with `magic-cli-remote` Tag & Release Pattern

## Context and Problem Statement

When a maintainer pushed release tag `v1.1.1`, CI executed unit tests but did not build or publish release binaries. A manual dispatch of the release workflow subsequently failed during action setup in the `Build, Verify, and Attest` job.

Investigation revealed three distinct issues in the current CI/CD configuration:
1. **Workflow Trigger Disconnect**: [`.github/workflows/release.yml`](../.github/workflows/release.yml) was configured to run exclusively via `workflow_dispatch` (manual dispatch), completely ignoring `push: tags: ['v*']` events.
2. **Invalid Action Reference**: [`.github/workflows/release.yml`](../.github/workflows/release.yml#L78) referenced a non-existent commit SHA (`1e69f48acb82d1966a394da916b4c1698aa569d6 # v4.1.0`) for `actions/attest-build-provenance`, causing GitHub Actions runner setup to fail immediately (Run [`33329311164`](https://github.com/maccavelli/prepare-commit-msg/actions/runs/33329311164)).
3. **Over-Engineered Staging Scheme**: The workflow attempted multi-step draft creation, verification, and editing (`gh release create --draft` followed by `gh release edit --draft=false`) rather than publishing directly to GitHub Releases upon successful build.

The maintainer directed adopting the established, proven CI/CD architecture from [`magic-cli-remote`](/Users/saxsmith/gitrepos/go/magic-cli-remote/.github/workflows/ci.yml) as the baseline for `prepare-commit-msg`.

## Decision Drivers

* **Automated Tag-Driven Releases**: Pushing a `v*` tag must automatically execute quality checks, compile cross-platform binaries, and publish the GitHub release.
* **Single Workflow Simplicity**: Consolidate CI and Release into a single [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) workflow, eliminating fragmentation across `ci.yml`, `quality.yml`, and `release.yml`.
* **Reliable Concurrency**: Branch and PR runs cancel superseded runs, while tag runs are never cancelled (`cancel-in-progress: ${{ github.ref_type != 'tag' }}`).
* **Native Cross-Platform Testing**: Execute native tests across Linux, macOS, and Windows runners before publishing.
* **Direct Publication**: Use GitHub CLI (`gh release create` and `gh release upload --clobber`) directly, without intermediate draft staging.

## Decision Outcome

Adopt the `magic-cli-remote` workflow pattern:

1. **Unified Workflow File**: Consolidate all CI/CD into [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) with triggers:
   ```yaml
   on:
     push:
       branches: [main]
       tags:
         - "v*"
     pull_request:
     workflow_dispatch:
   ```
2. **Concurrency Policy**:
   ```yaml
   concurrency:
     group: ci-${{ github.workflow }}-${{ github.ref }}
     cancel-in-progress: ${{ github.ref_type != 'tag' }}
   ```
3. **Pipeline Stages**:
   * **`go` Job (Test & Tag Build)**:
     * Runs on `ubuntu-24.04` (or `ubuntu-latest`).
     * Runs repository verification (`make verify` or `fmt-check`, `mod-check`, `vet`, `lint`, race tests, coverage, `govulncheck`, `workflow-lint`).
     * On tag runs (`if: github.ref_type == 'tag'`): Compiles the 6 target binaries (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`), writes `SHA256SUMS`, and uploads the artifact bundle.
   * **`go-native` Job (Matrix Tests)**:
     * Matrix across `ubuntu-24.04` (Linux), `macos-15` (macOS), and `windows-2025` (Windows).
     * Runs native `go test ./...` on each OS.
   * **`release` Job (Publish Assets)**:
     * Condition: `if: github.ref_type == 'tag'`, `needs: [go, go-native]`.
     * Runs on `ubuntu-latest` with `permissions: contents: write`.
     * Downloads the built binaries artifact.
     * Verifies `SHA256SUMS`.
     * Creates and publishes the GitHub release:
       ```bash
       TAG="${GITHUB_REF_NAME}"
       if ! gh release view "$TAG" >/dev/null 2>&1; then
         gh release create "$TAG" --generate-notes --title "$TAG"
       fi
       gh release upload "$TAG" dist/* --clobber
       ```
4. **Cleanup**: Remove redundant files [`.github/workflows/quality.yml`](../.github/workflows/quality.yml) and [`.github/workflows/release.yml`](../.github/workflows/release.yml).

## Consequences

### Positive
* Pushing a tag like `v1.1.1` automatically triggers the complete pipeline and publishes the release assets.
* Eliminates broken and invalid action SHAs.
* Matches the fleet-standard CI/CD pattern established in `magic-cli-remote`.
* Retains multi-platform native testing on Linux, macOS, and Windows.

### Neutral
* Local Makefile and hook commands (`make verify`, `make test`, `make build-all`) remain unchanged and authoritative.
