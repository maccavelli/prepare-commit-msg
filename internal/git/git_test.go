package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsCommitMsgEmpty_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_MSG")

	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsCommitMsgEmpty(path) {
		t.Error("expected empty file to return true")
	}
}

func TestIsCommitMsgEmpty_CommentsOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_MSG")

	content := "# Please enter the commit message\n# Lines starting with '#' are ignored\n\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsCommitMsgEmpty(path) {
		t.Error("expected comments-only file to return true")
	}
}

func TestIsCommitMsgEmpty_CRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_MSG")
	content := "# comment\r\n\r\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsCommitMsgEmpty(path) {
		t.Error("expected CRLF comments-only to be empty")
	}
	if err := os.WriteFile(path, []byte("feat: x\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if IsCommitMsgEmpty(path) {
		t.Error("expected CRLF message to be non-empty")
	}
}

func TestIsCommitMsgEmpty_WithMessage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_MSG")

	content := "feat: add feature\n# comment\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if IsCommitMsgEmpty(path) {
		t.Error("expected file with message to return false")
	}
}

func TestIsCommitMsgEmpty_NonExistent(t *testing.T) {
	if !IsCommitMsgEmpty("/nonexistent/path/COMMIT_MSG") {
		t.Error("expected nonexistent file to return true")
	}
}

func TestTruncateUTF8(t *testing.T) {
	s := "hello\n世界\nmore"
	// Multi-byte runes should not be split mid-character.
	out := TruncateUTF8(s, 10)
	if !strings.Contains(out, "[diff truncated]") {
		t.Fatalf("expected truncation marker: %q", out)
	}
	// Full string under limit unchanged.
	if TruncateUTF8("abc", 100) != "abc" {
		t.Error("short string should be unchanged")
	}
}

func TestGatherInfo(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Isolate from any parent git repositories (e.g. if t.TempDir() is inside a workspace)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))

	// Not a git repo → error (no longer silent empty).
	_, err := GatherInfo(ctx, 32000)
	if err == nil {
		t.Error("expected error for non-git repo")
	}

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Fatal(err)
	}
	_ = exec.Command("git", "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "config", "user.name", "Test User").Run()

	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0o644)
	_ = exec.Command("git", "add", "README.md").Run()
	_ = exec.Command("git", "commit", "-m", "Initial commit").Run()

	_ = os.WriteFile(filepath.Join(dir, "app.go"), []byte("package main\n\nfunc main() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("timeout: 30s\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "test.json"), []byte("{}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "main.tf"), []byte("resource {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "Jenkinsfile"), []byte("pipeline {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "other.txt"), []byte("text\n"), 0o644)

	_ = exec.Command("git", "add", ".").Run()

	info, err := GatherInfo(ctx, 32000)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(info.Stats, "JSON: 1") || !strings.Contains(info.Stats, "Terraform: 1") ||
		!strings.Contains(info.Stats, "CI/CD: 1") || !strings.Contains(info.Stats, "Other: 1") {
		t.Errorf("stats missing some categories: %s", info.Stats)
	}

	infoTruncated, err := GatherInfo(ctx, 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(infoTruncated.Diff, "[diff truncated]") {
		t.Error("expected truncated diff string")
	}

	// Empty staged after reset
	_ = exec.Command("git", "reset", "HEAD").Run()
	empty, err := GatherInfo(ctx, 32000)
	if err != nil {
		t.Errorf("empty staged should not error: %v", err)
	}
	if len(empty.Files) != 0 {
		t.Errorf("expected no files, got %v", empty.Files)
	}
}

func TestGatherInfo_MissingGit(t *testing.T) {
	oldFn := lookPath
	defer func() { lookPath = oldFn }()

	lookPath = func(file string) (string, error) {
		return "", os.ErrNotExist
	}

	_, err := GatherInfo(context.Background(), 32000)
	if err == nil {
		t.Error("expected error when git is missing, got nil")
	}
}
