# Plan: Align CI/CD Workflow with `magic-cli-remote`

## Overview
Replaces the fragmented and broken CI/CD workflows (`quality.yml`, `release.yml`) with a single, unified `.github/workflows/ci.yml` modeled on `magic-cli-remote`. This restores automated release publishing whenever a `v*` tag is pushed.

---

## Phase 1: Author Unified `ci.yml` Workflow

### 1. Update [`.github/workflows/ci.yml`](../.github/workflows/ci.yml)
Define the complete pipeline in a single workflow:
* **Triggers**:
  - `push` on `branches: [main]`
  - `push` on `tags: ['v*']`
  - `pull_request`
  - `workflow_dispatch`
* **Concurrency**:
  - Group: `ci-${{ github.workflow }}-${{ github.ref }}`
  - `cancel-in-progress: ${{ github.ref_type != 'tag' }}`
* **Job 1: `go` (Test; Build on Tag)**:
  - Runs on `ubuntu-24.04` (or `ubuntu-latest`).
  - Checks out code with `fetch-depth: 0`.
  - Sets up Go from `go.mod`.
  - Runs repository quality verification (`make verify`).
  - If `github.ref_type == 'tag'`:
    - Derives `BASE="${GITHUB_REF_NAME}"`.
    - Compiles the 6 cross-platform release binaries (`make build-all VERSION="$BASE"`).
    - Verifies checksums and writes `dist/SHA256SUMS`.
    - Uploads `dist/` as workflow artifact.
* **Job 2: `go-native` (Native Tests)**:
  - Runs matrix tests natively on Linux (`ubuntu-24.04`), macOS (`macos-15`), and Windows (`windows-2025`).
  - Runs `go test ./...` natively.
* **Job 3: `release` (Publish GitHub Release)**:
  - `if: github.ref_type == 'tag'`
  - `needs: [go, go-native]`
  - `permissions: contents: write`
  - Downloads `dist/` artifact.
  - Verifies checksum manifest (`sha256sum -c SHA256SUMS`).
  - Uses `gh release create "$TAG" --generate-notes --title "$TAG"` and `gh release upload "$TAG" dist/* --clobber` to publish the release directly.

---

## Phase 2: Remove Redundant Workflow Files & Update Docs

### 1. Remove Legacy Files
* Remove [`.github/workflows/quality.yml`](../.github/workflows/quality.yml).
* Remove [`.github/workflows/release.yml`](../.github/workflows/release.yml).

### 2. Update Documentation
* Update [`docs/README.md`](../docs/README.md) to index MADR-0004 and PLAN-0004.
* Update [`docs/cicd-operations.md`](../docs/cicd-operations.md) to document the simplified tag-and-push release model.

---

## Verification and Acceptance Criteria

1. **Local Verification**:
   - `make verify` passes locally (gofmt, mod-check, vet, lint, race tests, coverage >= 80%, vuln check, script/workflow linter).
2. **Action Validation**:
   - `make workflow-lint` passes with no workflow syntax errors.
3. **End-to-End Release Test**:
   - When a release tag `v1.1.1` is pushed:
     - CI runs tests across Linux, macOS, and Windows.
     - Cross-platform binaries are compiled.
     - Release is created and published on GitHub Releases under `v1.1.1` with all 6 binaries and `SHA256SUMS`.
