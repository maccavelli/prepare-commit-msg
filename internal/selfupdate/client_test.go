package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClient_FetchLatestRelease(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("unexpected User-Agent: %s", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"tag_name": "v4.4.0",
			"name": "Release v4.4.0",
			"html_url": "https://github.com/owner/repo/releases/tag/v4.4.0",
			"assets": [
				{
					"name": "prepare-commit-msg-linux-amd64",
					"size": 12345,
					"browser_download_url": "https://example.com/bin"
				},
				{
					"name": "SHA256SUMS",
					"size": 256,
					"browser_download_url": "https://example.com/sums"
				}
			]
		}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "test-agent", "test-token", server.Client())
	rel, err := client.FetchLatestRelease(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("FetchLatestRelease failed: %v", err)
	}

	if rel.TagName != "v4.4.0" {
		t.Errorf("rel.TagName = %q, want v4.4.0", rel.TagName)
	}
	if len(rel.Assets) != 2 {
		t.Errorf("len(rel.Assets) = %d, want 2", len(rel.Assets))
	}
	if asset := rel.FindAsset("SHA256SUMS"); asset == nil {
		t.Errorf("FindAsset(SHA256SUMS) returned nil")
	}
}

func TestClient_FetchReleaseByTag(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases/tags/v4.3.0", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"tag_name": "v4.3.0",
			"name": "Release v4.3.0"
		}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "test-agent", "", server.Client())
	rel, err := client.FetchReleaseByTag(context.Background(), "owner", "repo", "4.3.0")
	if err != nil {
		t.Fatalf("FetchReleaseByTag failed: %v", err)
	}

	if rel.TagName != "v4.3.0" {
		t.Errorf("rel.TagName = %q, want v4.3.0", rel.TagName)
	}
}

func TestParseChecksums(t *testing.T) {
	manifest := `
# Checksum manifest
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  prepare-commit-msg-linux-amd64
ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad *prepare-commit-msg-windows-amd64.exe
`
	sums, err := ParseChecksums(strings.NewReader(manifest))
	if err != nil {
		t.Fatalf("ParseChecksums failed: %v", err)
	}

	if sums["prepare-commit-msg-linux-amd64"] != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("unexpected hash for linux binary: %s", sums["prepare-commit-msg-linux-amd64"])
	}
	if sums["prepare-commit-msg-windows-amd64.exe"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Errorf("unexpected hash for windows binary: %s", sums["prepare-commit-msg-windows-amd64.exe"])
	}
}

func TestClient_DownloadAndVerifyAsset(t *testing.T) {
	content := []byte("binary-payload-content-12345")
	hasher := sha256.New()
	hasher.Write(content)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/download/bin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(content)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "test-agent", "", server.Client())
	tmpDir := t.TempDir()

	asset := &ReleaseAsset{
		Name:               "prepare-commit-msg-linux-amd64",
		BrowserDownloadURL: server.URL + "/download/bin",
	}

	// 1. Successful download and verification
	tmpFile, size, err := client.DownloadAndVerifyAsset(context.Background(), asset, expectedHash, tmpDir)
	if err != nil {
		t.Fatalf("DownloadAndVerifyAsset failed: %v", err)
	}
	if size != int64(len(content)) {
		t.Errorf("downloaded size = %d, want %d", size, len(content))
	}
	gotBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(gotBytes) != string(content) {
		t.Errorf("gotBytes = %q, want %q", string(gotBytes), string(content))
	}
	_ = os.Remove(tmpFile)

	// 2. Corrupted / mismatched hash should fail and clean up temp file
	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"
	_, _, err = client.DownloadAndVerifyAsset(context.Background(), asset, wrongHash, tmpDir)
	if err == nil {
		t.Fatalf("expected error on hash mismatch, got nil")
	}

	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "."+asset.Name+".tmp-") {
			t.Errorf("temp file %s was not cleaned up on failure", filepath.Join(tmpDir, e.Name()))
		}
	}
}

func TestNewClientDefaultsAndEnvironmentToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "environment-token")
	t.Setenv("GH_TOKEN", "ignored-token")

	client := NewClient("", "", "", nil)
	if client.BaseURL != DefaultGitHubAPI {
		t.Fatalf("BaseURL = %q, want %q", client.BaseURL, DefaultGitHubAPI)
	}
	if client.Token != "environment-token" {
		t.Fatalf("Token = %q, want environment-token", client.Token)
	}
	if client.HTTPClient == nil || client.HTTPClient.Timeout != DefaultHTTPTimeout {
		t.Fatalf("HTTPClient timeout = %v, want %v", client.HTTPClient.Timeout, DefaultHTTPTimeout)
	}

	req, err := client.newRequest(context.Background(), http.MethodGet, "https://example.com")
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}
	if got := req.Header.Get("User-Agent"); got != "prepare-commit-msg-selfupdate" {
		t.Fatalf("User-Agent = %q, want default", got)
	}
}

func TestClientHTTPAndManifestErrors(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/repos/owner/repo/releases/tags/v9.9.9", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	})
	mux.HandleFunc("/checksums", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusBadGateway)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(server.URL, "test-agent", "", server.Client())

	if _, err := client.FetchLatestRelease(context.Background(), "owner", "repo"); err == nil {
		t.Fatal("FetchLatestRelease() error = nil, want HTTP error")
	}
	if _, err := client.FetchReleaseByTag(context.Background(), "owner", "repo", "9.9.9"); err == nil {
		t.Fatal("FetchReleaseByTag() error = nil, want HTTP error")
	}
	if _, err := client.FetchChecksums(context.Background(), &Release{TagName: "v9.9.9"}); err == nil {
		t.Fatal("FetchChecksums() error = nil, want missing manifest")
	}
	if _, err := client.FetchChecksums(context.Background(), &Release{
		TagName: "v9.9.9",
		Assets: []ReleaseAsset{{
			Name:               ChecksumsFileName,
			BrowserDownloadURL: server.URL + "/checksums",
		}},
	}); err == nil {
		t.Fatal("FetchChecksums() error = nil, want HTTP error")
	}
	if _, err := ParseChecksums(strings.NewReader("not-a-manifest\n")); err == nil {
		t.Fatal("ParseChecksums() error = nil, want invalid manifest error")
	}
}

func TestDownloadAndVerifyAssetRejectsMissingChecksum(t *testing.T) {
	client := NewClient("", "test-agent", "", nil)
	_, _, err := client.DownloadAndVerifyAsset(context.Background(), &ReleaseAsset{
		Name: "prepare-commit-msg-linux-amd64",
	}, "", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "cannot be empty") {
		t.Fatalf("DownloadAndVerifyAsset() error = %v, want empty checksum error", err)
	}
}
