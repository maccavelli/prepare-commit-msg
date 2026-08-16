package selfupdate

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultGitHubAPI is the default GitHub REST API root endpoint.
	DefaultGitHubAPI = "https://api.github.com"

	// DefaultHTTPTimeout is the client HTTP request timeout.
	DefaultHTTPTimeout = 60 * time.Second

	// ChecksumsFileName is the standard checksum manifest asset name.
	ChecksumsFileName = "SHA256SUMS"
)

// Release represents a GitHub Release API response.
type Release struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	Body        string         `json:"body"`
	HTMLURL     string         `json:"html_url"`
	PublishedAt time.Time      `json:"published_at"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
	Assets      []ReleaseAsset `json:"assets"`
}

// ReleaseAsset represents an individual asset uploaded to a GitHub Release.
type ReleaseAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	State              string `json:"state"`
}

// Client interacts with GitHub Releases and asset endpoints.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	UserAgent  string
	Token      string
}

// NewClient initializes a GitHub Releases client.
func NewClient(baseURL, userAgent, token string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultGitHubAPI
	}
	baseURL = strings.TrimRight(baseURL, "/")

	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultHTTPTimeout}
	}

	if token == "" {
		if t := os.Getenv("GITHUB_TOKEN"); t != "" {
			token = t
		} else if t := os.Getenv("GH_TOKEN"); t != "" {
			token = t
		}
	}

	return &Client{
		HTTPClient: httpClient,
		BaseURL:    baseURL,
		UserAgent:  userAgent,
		Token:      token,
	}
}

func (c *Client) newRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	} else {
		req.Header.Set("User-Agent", "prepare-commit-msg-selfupdate")
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return req, nil
}

// FetchLatestRelease retrieves the latest non-draft, non-prerelease release.
func (c *Client) FetchLatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.BaseURL, owner, repo)
	req, err := c.newRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("github api error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release json: %w", err)
	}

	return &rel, nil
}

// FetchReleaseByTag retrieves a specific release by its tag name.
func (c *Client) FetchReleaseByTag(ctx context.Context, owner, repo, tag string) (*Release, error) {
	tag = strings.TrimSpace(tag)
	if !strings.HasPrefix(tag, "v") && !strings.HasPrefix(tag, "V") {
		tag = "v" + tag
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", c.BaseURL, owner, repo, tag)
	req, err := c.newRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch release tag %q: %w", tag, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("github api error (status %d) for tag %q: %s", resp.StatusCode, tag, strings.TrimSpace(string(body)))
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release json for tag %q: %w", tag, err)
	}

	return &rel, nil
}

// FindAsset finds an asset by exact name in a release.
func (rel *Release) FindAsset(name string) *ReleaseAsset {
	for i := range rel.Assets {
		if rel.Assets[i].Name == name {
			return &rel.Assets[i]
		}
	}
	return nil
}

// FetchChecksums downloads the SHA256SUMS asset and returns a map of filename -> hex hash.
func (c *Client) FetchChecksums(ctx context.Context, rel *Release) (map[string]string, error) {
	asset := rel.FindAsset(ChecksumsFileName)
	if asset == nil {
		return nil, fmt.Errorf("checksum manifest %q not found in release %s", ChecksumsFileName, rel.TagName)
	}

	req, err := c.newRequest(ctx, http.MethodGet, asset.BrowserDownloadURL)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download checksums failed with status %d", resp.StatusCode)
	}

	return ParseChecksums(resp.Body)
}

// ParseChecksums parses a standard sha256sum manifest format (<hash>  <filename>).
func ParseChecksums(r io.Reader) (map[string]string, error) {
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		hash := strings.ToLower(fields[0])
		filename := fields[1]
		filename = strings.TrimPrefix(filename, "*") // strip binary mode prefix if present
		filename = filepath.Base(filename)

		if len(hash) == 64 { // SHA-256 hex length
			checksums[filename] = hash
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan checksums: %w", err)
	}

	if len(checksums) == 0 {
		return nil, fmt.Errorf("no valid sha256 checksums found in manifest")
	}

	return checksums, nil
}

// DownloadAndVerifyAsset streams an asset to a temporary file in targetDir,
// validating its SHA-256 checksum against expectedSHA256.
func (c *Client) DownloadAndVerifyAsset(ctx context.Context, asset *ReleaseAsset, expectedSHA256, targetDir string) (string, int64, error) {
	if expectedSHA256 == "" {
		return "", 0, fmt.Errorf("expected sha256 checksum cannot be empty")
	}
	expectedSHA256 = strings.ToLower(strings.TrimSpace(expectedSHA256))

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", 0, fmt.Errorf("target dir %q: %w", targetDir, err)
	}

	tmpFile, err := os.CreateTemp(targetDir, "."+asset.Name+".tmp-*")
	if err != nil {
		return "", 0, fmt.Errorf("create temporary download file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on failure
	cleanup := true
	defer func() {
		_ = tmpFile.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	req, err := c.newRequest(ctx, http.MethodGet, asset.BrowserDownloadURL)
	if err != nil {
		return "", 0, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("download asset %s: %w", asset.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("download asset %s failed with status %d", asset.Name, resp.StatusCode)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, hasher)

	written, err := io.Copy(multiWriter, resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("write asset stream: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return "", 0, fmt.Errorf("sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", 0, fmt.Errorf("close temp file: %w", err)
	}

	actualSHA256 := hex.EncodeToString(hasher.Sum(nil))
	if actualSHA256 != expectedSHA256 {
		return "", 0, fmt.Errorf("checksum mismatch for %s: expected %s, got %s",
			asset.Name, expectedSHA256, actualSHA256)
	}

	cleanup = false
	return tmpPath, written, nil
}
