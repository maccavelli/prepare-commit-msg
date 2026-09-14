package ui

import (
	"fmt"
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

	for _, test := range []struct {
		name     string
		exitCode int
		wantErr  bool
	}{
		{name: "success", exitCode: 0},
		{name: "failure", exitCode: 9, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			commandPath := filepath.Join(dir, commandName)
			script := fmt.Sprintf("#!/bin/sh\nexit %d\n", test.exitCode)
			if err := os.WriteFile(commandPath, []byte(script), 0o700); err != nil {
				t.Fatalf("write command fixture: %v", err)
			}
			t.Setenv("PATH", dir)

			err := openBrowserDefault("https://example.invalid/authorize")
			if test.wantErr && (err == nil || !strings.Contains(err.Error(), "open browser")) {
				t.Fatalf("openBrowserDefault() error = %v, want open-browser error", err)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("openBrowserDefault() error = %v", err)
			}
		})
	}
}

func TestOpenBrowserDefault_RejectsUnsafeURL(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	err := openBrowserDefault("javascript:alert(1)\n")
	if err == nil || !strings.Contains(err.Error(), "invalid browser URL") {
		t.Fatalf("openBrowserDefault() error = %v, want invalid-browser-URL error", err)
	}
}
