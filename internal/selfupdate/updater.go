package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// DefaultRepoOwner is the GitHub repository owner.
	DefaultRepoOwner = "maccavelli"

	// DefaultRepoName is the GitHub repository name.
	DefaultRepoName = "prepare-commit-msg"
)

// PlatformAsset returns the expected binary release asset name for the current GOOS and GOARCH.
func PlatformAsset() (string, error) {
	return PlatformAssetFor(runtime.GOOS, runtime.GOARCH)
}

// PlatformAssetFor returns the expected binary asset name for the given OS and architecture.
func PlatformAssetFor(goos, goarch string) (string, error) {
	switch {
	case goos == "linux" && goarch == "amd64":
		return "prepare-commit-msg-linux-amd64", nil
	case goos == "darwin" && goarch == "arm64":
		return "prepare-commit-msg-darwin-arm64", nil
	case goos == "windows" && goarch == "amd64":
		return "prepare-commit-msg-windows-amd64.exe", nil
	default:
		return "", fmt.Errorf("no prebuilt release binary available for %s/%s. Build from source via 'make build' or 'go install'", goos, goarch)
	}
}

// Options configure the self-update execution.
type Options struct {
	RepoOwner      string
	RepoName       string
	CurrentVersion string
	TargetVersion  string
	CheckOnly      bool
	Force          bool
	Timeout        time.Duration
	ExecutablePath string
	Client         *Client
	Output         io.Writer
}

// UpdateResult captures the outcome of an update check or installation.
type UpdateResult struct {
	CurrentVersion string
	LatestVersion  string
	Updated        bool
	ReleaseURL     string
	ReleaseNotes   string
}

// Run executes the self-update check and upgrade workflow.
func Run(ctx context.Context, opts Options) (*UpdateResult, error) {
	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	owner := opts.RepoOwner
	if owner == "" {
		owner = DefaultRepoOwner
	}
	repo := opts.RepoName
	if repo == "" {
		repo = DefaultRepoName
	}

	currVer := opts.CurrentVersion
	if currVer == "" {
		currVer = "unknown"
	}

	execPath := opts.ExecutablePath
	if execPath == "" {
		p, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("determine executable path: %w", err)
		}
		execPath = p
	}

	// Resolve symlinks to operate on the canonical binary
	if realPath, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = realPath
	}

	assetName, err := PlatformAsset()
	if err != nil {
		return nil, err
	}

	client := opts.Client
	if client == nil {
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = DefaultHTTPTimeout
		}
		client = NewClient(DefaultGitHubAPI, "prepare-commit-msg/"+currVer, "", nil)
	}

	fmt.Fprintf(out, "Checking for updates from https://github.com/%s/%s...\n", owner, repo)

	var rel *Release
	if opts.TargetVersion != "" {
		rel, err = client.FetchReleaseByTag(ctx, owner, repo, opts.TargetVersion)
	} else {
		rel, err = client.FetchLatestRelease(ctx, owner, repo)
	}
	if err != nil {
		return nil, fmt.Errorf("fetch release information: %w", err)
	}

	latestVer := rel.TagName
	result := &UpdateResult{
		CurrentVersion: currVer,
		LatestVersion:  latestVer,
		ReleaseURL:     rel.HTMLURL,
		ReleaseNotes:   rel.Body,
	}

	isNewer := IsNewer(latestVer, currVer)
	isSame := Compare(latestVer, currVer) == 0

	// Check if already up to date
	if !opts.Force && opts.TargetVersion == "" && !isNewer {
		fmt.Fprintf(out, "prepare-commit-msg is already up to date (%s).\n", currVer)
		return result, nil
	}

	if !opts.Force && opts.TargetVersion != "" && isSame {
		fmt.Fprintf(out, "prepare-commit-msg is already at requested version (%s).\n", latestVer)
		return result, nil
	}

	// Dry run mode (--check)
	if opts.CheckOnly {
		if isNewer {
			fmt.Fprintf(out, "An update is available:\n")
			fmt.Fprintf(out, "  Current version: %s\n", currVer)
			fmt.Fprintf(out, "  Latest version:  %s\n", latestVer)
			fmt.Fprintf(out, "  Release page:    %s\n", rel.HTMLURL)
			fmt.Fprintf(out, "Run 'prepare-commit-msg update' to apply the update.\n")
		} else {
			fmt.Fprintf(out, "prepare-commit-msg is up to date (%s).\n", currVer)
		}
		return result, nil
	}

	// Find the platform binary asset
	binaryAsset := rel.FindAsset(assetName)
	if binaryAsset == nil {
		return nil, fmt.Errorf("release %s does not contain asset %s for this platform", latestVer, assetName)
	}

	// Fetch checksums manifest
	fmt.Fprintf(out, "New version available: %s (current: %s)\n", latestVer, currVer)
	fmt.Fprintf(out, "Downloading %s from release %s...\n", assetName, latestVer)

	checksums, err := client.FetchChecksums(ctx, rel)
	if err != nil {
		return nil, fmt.Errorf("fetch checksums manifest: %w", err)
	}

	expectedHash, ok := checksums[assetName]
	if !ok {
		return nil, fmt.Errorf("asset %s not listed in release %s SHA256SUMS manifest", assetName, latestVer)
	}

	// Stage temp file in the same directory as the target executable
	targetDir := filepath.Dir(execPath)
	tempFile, size, err := client.DownloadAndVerifyAsset(ctx, binaryAsset, expectedHash, targetDir)
	if err != nil {
		return nil, fmt.Errorf("download and verify binary: %w", err)
	}

	fmt.Fprintf(out, "[✓] Downloaded binary (%s)\n", formatBytes(size))
	fmt.Fprintf(out, "[✓] Verified SHA-256 checksum (%s)\n", expectedHash)

	// In-place atomic replacement
	if err := ReplaceExecutable(tempFile, execPath); err != nil {
		if errors.Is(err, ErrPermissionDenied) {
			fmt.Fprintf(out, "\nprepare-commit-msg: error: permission denied while updating %s\n", execPath)
			fmt.Fprintf(out, "prepare-commit-msg: please rerun the command with elevated permissions:\n")
			fmt.Fprintf(out, "    sudo prepare-commit-msg update\n\n")
			return nil, fmt.Errorf("permission denied updating %s", execPath)
		}
		return nil, fmt.Errorf("replace executable %s: %w", execPath, err)
	}

	result.Updated = true
	fmt.Fprintf(out, "[✓] Successfully updated prepare-commit-msg to %s!\n", latestVer)
	return result, nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
