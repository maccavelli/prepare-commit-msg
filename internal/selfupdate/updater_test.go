package selfupdate

import (
	"bytes"
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

func TestPlatformAssetFor(t *testing.T) {
	tests := []struct {
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{"linux", "amd64", "prepare-commit-msg-linux-amd64", false},
		{"linux", "arm64", "prepare-commit-msg-linux-arm64", false},
		{"darwin", "arm64", "prepare-commit-msg-darwin-arm64", false},
		{"darwin", "amd64", "prepare-commit-msg-darwin-amd64", false},
		{"windows", "amd64", "prepare-commit-msg-windows-amd64.exe", false},
		{"windows", "arm64", "prepare-commit-msg-windows-arm64.exe", false},
		{"freebsd", "amd64", "", true},
		{"linux", "386", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"_"+tt.goarch, func(t *testing.T) {
			got, err := PlatformAssetFor(tt.goos, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PlatformAssetFor(%q, %q) error = %v, wantErr = %v", tt.goos, tt.goarch, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("PlatformAssetFor(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}

func TestRun_Workflow(t *testing.T) {
	assetName, err := PlatformAsset()
	if err != nil {
		t.Skipf("Platform asset not supported on current platform: %v", err)
	}

	binaryPayload := []byte("new-binary-v4.4.0-payload")
	hasher := sha256.New()
	hasher.Write(binaryPayload)
	checksumHex := hex.EncodeToString(hasher.Sum(nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"tag_name": "v4.4.0",
			"name": "Release v4.4.0",
			"html_url": "https://github.com/owner/repo/releases/tag/v4.4.0",
			"body": "Release notes for v4.4.0",
			"assets": [
				{
					"name": "%s",
					"size": %d,
					"browser_download_url": "%s/download/%s"
				},
				{
					"name": "SHA256SUMS",
					"size": 100,
					"browser_download_url": "%s/download/SHA256SUMS"
				}
			]
		}`, assetName, len(binaryPayload), "http://"+r.Host, assetName, "http://"+r.Host)
	})

	mux.HandleFunc("/download/"+assetName, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(binaryPayload)
	})

	mux.HandleFunc("/download/SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s  %s\n", checksumHex, assetName)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "test-agent", "", server.Client())
	tmpDir := t.TempDir()

	currentExec := filepath.Join(tmpDir, "prepare-commit-msg")
	if err := os.WriteFile(currentExec, []byte("old-binary-v4.3.2"), 0o755); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 1. Dry run check (--check)
	var checkOut bytes.Buffer
	res, err := Run(context.Background(), Options{
		RepoOwner:      "owner",
		RepoName:       "repo",
		CurrentVersion: "v4.3.2",
		CheckOnly:      true,
		ExecutablePath: currentExec,
		Client:         client,
		Output:         &checkOut,
	})
	if err != nil {
		t.Fatalf("Run check failed: %v", err)
	}
	if res.Updated {
		t.Errorf("res.Updated should be false on check-only")
	}
	if !strings.Contains(checkOut.String(), "An update is available") {
		t.Errorf("expected check output to mention update, got: %s", checkOut.String())
	}

	// 2. Already up to date
	var upToDateOut bytes.Buffer
	res, err = Run(context.Background(), Options{
		RepoOwner:      "owner",
		RepoName:       "repo",
		CurrentVersion: "v4.4.0",
		ExecutablePath: currentExec,
		Client:         client,
		Output:         &upToDateOut,
	})
	if err != nil {
		t.Fatalf("Run up-to-date failed: %v", err)
	}
	if res.Updated {
		t.Errorf("res.Updated should be false when already up to date")
	}
	if !strings.Contains(upToDateOut.String(), "already up to date") {
		t.Errorf("expected output to mention already up to date, got: %s", upToDateOut.String())
	}

	// 3. Successful update
	var updateOut bytes.Buffer
	res, err = Run(context.Background(), Options{
		RepoOwner:      "owner",
		RepoName:       "repo",
		CurrentVersion: "v4.3.2",
		ExecutablePath: currentExec,
		Client:         client,
		Output:         &updateOut,
	})
	if err != nil {
		t.Fatalf("Run update failed: %v", err)
	}
	if !res.Updated {
		t.Errorf("res.Updated should be true after successful update")
	}

	// Verify the executable on disk was replaced
	newContent, err := os.ReadFile(currentExec)
	if err != nil {
		t.Fatalf("ReadFile after update failed: %v", err)
	}
	if string(newContent) != string(binaryPayload) {
		t.Errorf("executable content = %q, want %q", string(newContent), string(binaryPayload))
	}
}
