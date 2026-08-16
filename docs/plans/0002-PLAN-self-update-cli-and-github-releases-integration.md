---
status: "ready_for_review"
date: 2026-08-16
associated-madr: "0002-MADR-self-update-cli-and-github-releases-integration.md"
owner: "Maintainers of prepare-commit-msg"
target-milestone: "v4.4.0"
---

# Plan: Self-Update CLI Subcommand and GitHub Releases Integration

## Executive Summary & Goal

This implementation plan translates the architectural decisions formulated in [**ADR-0002: Self-Update CLI Subcommand and GitHub Releases Integration**](file:///home/mac/gitrepos/prepare-commit-msg/docs/decisions/0002-MADR-self-update-cli-and-github-releases-integration.md) into concrete, phased, and deterministic engineering tasks.

The objective is to implement a robust, cross-platform, zero-dependency self-updater into [`prepare-commit-msg`](file:///home/mac/gitrepos/prepare-commit-msg/README.md) exposed via the `update` CLI subcommand. The updater queries the official GitHub repository releases, streams the correct platform binary asset, verifies its SHA-256 hash against the published `SHA256SUMS` manifest, and atomically replaces the active executable in-place.

---

## Prerequisites & Dependencies

* **Go Toolchain:** Go $\ge$ 1.24 (as defined in [`go.mod`](file:///home/mac/gitrepos/prepare-commit-msg/go.mod)).
* **Target Architecture Matrix:**
  * Linux x86_64 (`linux-amd64` $\rightarrow$ `prepare-commit-msg-linux-amd64`)
  * macOS ARM64 (`darwin-arm64` $\rightarrow$ `prepare-commit-msg-darwin-arm64`)
  * Windows x86_64 (`windows-amd64` $\rightarrow$ `prepare-commit-msg-windows-amd64.exe`)
* **Upstream Release Assets:** GitHub Releases under repository `maccavelli/prepare-commit-msg` publishing platform binaries and `SHA256SUMS`.
* **Zero External Dependencies:** Built entirely with Go standard library packages (`crypto/sha256`, `net/http`, `os`, `path/filepath`, `runtime`, `encoding/json`, `io`, `time`).

---

## Architecture & Technical Design Summary

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          main.go (CLI Dispatch)                             │
│                      switch args[0] -> "update"                             │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ calls
┌──────────────────────────────────────▼──────────────────────────────────────┐
│                        internal/selfupdate/updater.go                       │
│  Orchestrates resolution, release fetch, checksum validation, and swap      │
└──────────────┬───────────────────────┬───────────────────────────────┬──────┘
               │                       │                               │
┌──────────────▼────────────┐ ┌────────▼───────────┐ ┌─────────────────▼──────┐
│         client.go         │ │     semver.go      │ │  apply_unix.go /       │
│ GitHub API, asset stream, │ │ Parsing & semantic │ │  apply_windows.go      │
│ SHA256SUMS parser         │ │ version comparison │ │ Atomic file swap logic │
└───────────────────────────┘ └────────────────────┘ └────────────────────────┘
```

### File Hierarchy & Package Responsibilities

| File Path | Responsibility |
| :--- | :--- |
| [`internal/selfupdate/semver.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/semver.go) | Custom, zero-alloc SemVer parser, validator, and comparison functions (`Compare`, `IsNewer`, `CleanVersion`). |
| [`internal/selfupdate/semver_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/semver_test.go) | Table-driven unit tests for all SemVer permutations, prefixes (`v`), prerelease identifiers, and edge cases. |
| [`internal/selfupdate/client.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/client.go) | GitHub Releases REST client (`FetchLatestRelease`, `FetchReleaseByTag`, `FetchChecksums`, `DownloadAsset`). Supports rate-limit token detection (`GITHUB_TOKEN` / `GH_TOKEN`). |
| [`internal/selfupdate/client_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/client_test.go) | Unit tests mocking GitHub Releases API and `SHA256SUMS` with `net/http/httptest`. |
| [`internal/selfupdate/apply_unix.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/apply_unix.go) | POSIX-compliant atomic replacement (`//go:build !windows`): same-directory temporary file staging, `0755` permissions, `os.Rename`. |
| [`internal/selfupdate/apply_windows.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/apply_windows.go) | Windows replacement (`//go:build windows`): `.old` rollover renaming to circumvent active file locking, with automatic rollback. |
| [`internal/selfupdate/updater.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/updater.go) | High-level orchestrator (`Options`, `UpdateResult`, `Run`, `PlatformAsset`), user feedback reporting, and permission error interception. |
| [`internal/selfupdate/updater_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/updater_test.go) | End-to-end integration tests simulating binary downloads, valid/invalid checksums, dry-run checks, and elevated permission handling. |
| [`main.go`](file:///home/mac/gitrepos/prepare-commit-msg/main.go) | Subcommand routing for `update`, flag definitions (`--check`, `--force`, `--version`, `--yes`), usage documentation. |
| [`main_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/main_test.go) | CLI entrypoint tests validating flag parsing and command routing. |
| [`README.md`](file:///home/mac/gitrepos/prepare-commit-msg/README.md) | User documentation detailing `prepare-commit-msg update` usage and options. |

---

## Phased Execution Plan

### Phase 1: Core SemVer & GitHub Client Implementation

#### Objective
Implement semantic version parsing and the GitHub Releases API client with streaming SHA-256 verification and checksum manifest parsing.

#### Detailed Steps
1. **Implement `internal/selfupdate/semver.go`:**
   * Define `type Version struct { Major, Minor, Patch int; Prerelease string; Raw string }`.
   * Implement `ParseSemver(v string) (Version, error)`.
   * Implement `Compare(v1, v2 string) int` (returns `-1` if $v_1 < v_2$, `0` if $v_1 == v_2$, `1` if $v_1 > v_2$).
   * Implement `IsNewer(latest, current string) bool`.
   * Handle edge cases: leading `v` prefixes, unversioned dirty development strings (`dev`, `dirty`, `unknown`).
2. **Implement `internal/selfupdate/semver_test.go`:**
   * Test versions: `v4.3.2` vs `v4.4.0`, `1.0.0` vs `1.0.1`, `v2.0.0-rc.1` vs `v2.0.0`, `4.3.2` vs `v4.3.2` (equal).
3. **Implement `internal/selfupdate/client.go`:**
   * Define GitHub API structs: `Release`, `ReleaseAsset`.
   * Implement `NewClient(opts ...ClientOption) *Client`.
   * Check environment for `GITHUB_TOKEN` or `GH_TOKEN` and append `Authorization: Bearer <token>`.
   * Implement `FetchLatestRelease(ctx context.Context, repo string) (*Release, error)`.
   * Implement `FetchReleaseByTag(ctx context.Context, repo, tag string) (*Release, error)`.
   * Implement `FetchChecksums(ctx context.Context, checksumAssetURL string) (map[string]string, error)`.
   * Implement `DownloadAndVerifyAsset(ctx context.Context, assetURL, expectedSHA256, targetDir string) (tempFilePath string, err error)`.
4. **Implement `internal/selfupdate/client_test.go`:**
   * Mock GitHub API release endpoints using `httptest.NewServer`.
   * Test successful checksum parsing from multi-line `SHA256SUMS`.
   * Test SHA-256 verification failure (corrupted payload) ensuring temp file cleanup.

---

### Phase 2: Cross-Platform Atomic Binary Replacement

#### Objective
Implement operating system-specific binary replacement primitives that handle running process locks, same-filesystem staging, and permission errors.

#### Detailed Steps
1. **Implement `internal/selfupdate/apply_unix.go` (`//go:build !windows`):**
   * Resolve symlinks on `os.Executable()` via `filepath.EvalSymlinks`.
   * Stage temporary file in `filepath.Dir(targetPath)` (e.g. `.prepare-commit-msg.tmp-<random>`).
   * Apply executable permission `0755` using `os.Chmod`.
   * Replace binary atomically with `os.Rename(tempFile, targetPath)`.
   * Intercept `os.ErrPermission` / `EACCES` and return formatted error advising `sudo prepare-commit-msg update`.
2. **Implement `internal/selfupdate/apply_windows.go` (`//go:build windows`):**
   * Resolve executable path.
   * Rename `targetPath` $\rightarrow$ `targetPath + ".old"`.
   * Rename `tempFile` $\rightarrow$ `targetPath`.
   * Attempt best-effort deletion of `targetPath + ".old"`.
   * If step 2 fails, trigger rollback: rename `targetPath + ".old"` back to `targetPath`.
3. **Implement `internal/selfupdate/apply_test.go`:**
   * Verify file replacement on dummy binaries in temporary test directories.
   * Verify permission retention (`0755`) after swap.

---

### Phase 3: High-Level Updater Orchestrator

#### Objective
Combine client, SemVer, and atomic replacer into an intuitive, configurable orchestrator with progress reporting.

#### Detailed Steps
1. **Implement `internal/selfupdate/updater.go`:**
   * Define configuration options:
     ```go
     type Options struct {
         RepoOwner      string        // Default: "maccavelli"
         RepoName       string        // Default: "prepare-commit-msg"
         CurrentVersion string        // Current build version (e.g. main.Version)
         TargetVersion  string        // Optional: specific tag requested by user
         CheckOnly      bool          // If true, inspect without downloading/applying
         Force          bool          // If true, overwrite even if version is identical
         Timeout        time.Duration // Default: 60s
         Output         io.Writer     // Output writer (os.Stdout / buffer)
     }
     ```
   * Implement `PlatformAsset() (string, error)` mapping `runtime.GOOS` + `runtime.GOARCH` to `prepare-commit-msg-<os>-<arch>[.exe]`.
   * Implement `Run(ctx context.Context, opts Options) (*UpdateResult, error)`.
   * Provide user feedback:
     * Check phase: `Checking for updates from https://github.com/...`
     * Status phase: `New version available: v4.4.0 (current: v4.3.2)` or `prepare-commit-msg is already up to date (v4.3.2)`.
     * Download phase: `Downloading prepare-commit-msg-linux-amd64...`
     * Verification phase: `[✓] Verified SHA-256 checksum (...)`
     * Swap phase: `[✓] Successfully updated prepare-commit-msg to v4.4.0!`
2. **Implement `internal/selfupdate/updater_test.go`:**
   * Test `Run` with `CheckOnly: true` (dry run).
   * Test `Run` when already up to date (no-op).
   * Test `Run` with `Force: true` against current version.
   * Test `Run` with specific `TargetVersion`.
   * Test `Run` on unsupported architecture error handling.

---

### Phase 4: CLI Integration in `main.go` & Flag Routing

#### Objective
Expose the `update` subcommand in `main.go`, parse CLI flags, wire up signals/context timeouts, and update help documentation.

#### Detailed Steps
1. **Update `main.go`:**
   * Add case `"update":` in `switch args[0]`.
   * Implement `runUpdate(args []string) error`:
     ```go
     fs := flag.NewFlagSet("update", flag.ContinueOnError)
     check := fs.Bool("check", false, "check for updates without applying")
     force := fs.Bool("force", false, "reinstall current version or force overwrite")
     targetVersion := fs.String("version", "", "target specific version tag (e.g. v4.4.0)")
     yes := fs.Bool("yes", false, "non-interactive update")
     fs.BoolVar(yes, "y", false, "non-interactive update (shorthand)")
     ```
   * Pass `main.Version` and standard stdout/stderr streams.
2. **Update `printUsage()` in `main.go`:**
   * Add: `prepare-commit-msg update [flags] - check for and apply updates from GitHub`.
   * Add flag descriptions for `--check`, `--force`, `--version`, `--yes`.
3. **Update `main_test.go`:**
   * Test CLI invocation of `update --help`.
   * Test CLI invocation of `update --check` with mock version.

---

### Phase 5: Verification, Documentation, and Quality Gates

#### Objective
Ensure full test coverage, static analysis compliance, and documentation synchronization.

#### Detailed Steps
1. **Update `README.md`:**
   * Add "Self-Update" section explaining `prepare-commit-msg update`, `--check`, `--force`, and permissions handling.
2. **Execute Full Suite Quality Gates:**
   * Format code: `gofmt -s -w .`
   * Run vet: `go vet ./...`
   * Run linter: `golangci-lint run -c .golangci.yml ./...`
   * Run race detector: `go test -v -race ./...`
3. **Build Binary & Validate:**
   * Build local binary: `make build`
   * Test binary CLI: `./dist/prepare-commit-msg-linux-amd64 update --help`

---

## Verification & Testing Strategy

### 1. Automated Unit & Integration Tests
* **SemVer Logic (`semver_test.go`):** 100% statement coverage across standard, prefixed, invalid, and pre-release version strings.
* **HTTP & GitHub Mocking (`client_test.go`):** Mock 200 OK releases, 404 Not Found tags, 403 Rate Limit responses, and malformed JSON.
* **Integrity Validation (`client_test.go`):** Assert that a bit-flipped payload causes a checksum mismatch and immediately deletes the staged temporary file without touching the target binary.
* **Atomic Swap Mechanics (`apply_test.go`):** Verify file contents and permissions before and after swap on test binaries.
* **CLI Flag Parsing (`main_test.go`):** Verify flags are parsed and dispatched properly.

### 2. Manual CLI Verification
* Run `prepare-commit-msg update --help` to confirm formatting.
* Run `prepare-commit-msg update --check` on a built binary to confirm remote discovery.

---

## Rollback & Mitigation Procedures

* **Download/Checksum Failure:** If downloading or checksum validation fails at any point prior to file renaming, the temporary file is deleted immediately, leaving the original executable completely untouched.
* **Windows Rename Failure:** If Windows secondary rename fails, the updater executes an automated rollback restoring `app.exe.old` back to `app.exe`.
* **Explicit Version Pinning:** If a newer release contains a regression, users can downgrade at any time by running:
  ```bash
  prepare-commit-msg update --version v4.3.2 --force
  ```

---

## Granular Task Checklist

- [ ] **Phase 1: SemVer & GitHub Client**
  - [ ] Create [`internal/selfupdate/semver.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/semver.go)
  - [ ] Create [`internal/selfupdate/semver_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/semver_test.go)
  - [ ] Create [`internal/selfupdate/client.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/client.go)
  - [ ] Create [`internal/selfupdate/client_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/client_test.go)
- [ ] **Phase 2: Cross-Platform Atomic Binary Replacement**
  - [ ] Create [`internal/selfupdate/apply_unix.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/apply_unix.go)
  - [ ] Create [`internal/selfupdate/apply_windows.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/apply_windows.go)
  - [ ] Create [`internal/selfupdate/apply_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/apply_test.go)
- [ ] **Phase 3: High-Level Updater Orchestrator**
  - [ ] Create [`internal/selfupdate/updater.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/updater.go)
  - [ ] Create [`internal/selfupdate/updater_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/internal/selfupdate/updater_test.go)
- [ ] **Phase 4: CLI Command & Main Integration**
  - [ ] Update [`main.go`](file:///home/mac/gitrepos/prepare-commit-msg/main.go) with `update` subcommand and flag parsing
  - [ ] Update [`main_test.go`](file:///home/mac/gitrepos/prepare-commit-msg/main_test.go) with `update` command test cases
- [ ] **Phase 5: Documentation & Quality Verification**
  - [ ] Update [`README.md`](file:///home/mac/gitrepos/prepare-commit-msg/README.md) with `prepare-commit-msg update` documentation
  - [ ] Run `gofmt -s -w .`
  - [ ] Run `go vet ./...`
  - [ ] Run `golangci-lint run -c .golangci.yml ./...`
  - [ ] Run `go test -v -race ./...`
  - [ ] Build cross-platform binaries with `make build-all`
