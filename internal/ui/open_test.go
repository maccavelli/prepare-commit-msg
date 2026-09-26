package ui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOpenBrowserDefault_CommandResult(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("test command fixture covers the Unix browser launchers")
	}

	commandName := "xdg-open"
	if runtime.GOOS == "darwin" {
		commandName = "open"
	}

	dir := t.TempDir()
	commandPath := filepath.Join(dir, commandName)
	if err := os.WriteFile(commandPath, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatalf("write command fixture: %v", err)
	}
	t.Setenv("PATH", dir)

	if err := openBrowserDefault("https://example.invalid/authorize"); err != nil {
		t.Fatalf("openBrowserDefault() error = %v", err)
	}
}

func TestOpenBrowserDefault_StartFailure(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("test command fixture covers the Unix browser launchers")
	}
	t.Setenv("PATH", t.TempDir())
	err := openBrowserDefault("https://example.invalid/authorize")
	if err == nil || !strings.Contains(err.Error(), "open browser") {
		t.Fatalf("openBrowserDefault() error = %v, want open-browser error", err)
	}
}

func TestOpenBrowserDefault_RejectsUnsafeURL(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	err := openBrowserDefault("javascript:alert(1)\n")
	if err == nil || !strings.Contains(err.Error(), "invalid browser URL") {
		t.Fatalf("openBrowserDefault() error = %v, want invalid-browser-URL error", err)
	}
}
