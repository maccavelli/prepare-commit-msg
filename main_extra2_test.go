package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

func TestMain_RunConfigure(t *testing.T) {
	testConfigRoot := t.TempDir()
	t.Setenv("HOME", testConfigRoot)
	t.Setenv("USERPROFILE", testConfigRoot)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(testConfigRoot, ".config"))
	t.Setenv("APPDATA", filepath.Join(testConfigRoot, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(testConfigRoot, "AppData", "Local"))

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{
		"prepare-commit-msg", "configure", "--yes", "--provider=gemini",
		"--model=test", "--api-key=test-key",
	}

	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		t.Fatalf("configure unexpectedly called osExit(%d)", code)
	}

	main()

	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	rel, err := filepath.Rel(testConfigRoot, configPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		t.Fatalf("config path %q is outside isolated root", configPath)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("isolated config was not written: %v", err)
	}
	conf, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	pc := conf.Providers["gemini"]
	if conf.ActiveProvider != "gemini" || pc.Model != "test" || pc.APIKey != "test-key" {
		t.Fatalf("isolated config = provider %q, model %q, key match %t", conf.ActiveProvider, pc.Model, pc.APIKey == "test-key")
	}
}

func TestMain_RunHook_MissingFile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"prepare-commit-msg", "/non/existent/file"}

	var exitCode = -1
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
		panic("osExit")
	}

	defer func() {
		_ = recover()
		if exitCode != 1 {
			t.Errorf("expected missing file to exit 1, got %d", exitCode)
		}
	}()

	main()
}

func TestRunHook_SkipSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	os.WriteFile(path, []byte(""), 0o644)

	// "message" source triggers shouldSkipSource
	runHook([]string{path, "message"})
}

func TestRunHook_NotEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	os.WriteFile(path, []byte("feat: manual edit"), 0o644)

	// file is not empty, returns early
	runHook([]string{path})
}

func TestRunHook_ConfigError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("APPDATA", filepath.Join(tmp, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "AppData", "Local"))

	// Malformed primary configuration deterministically exercises the load
	// failure on every supported filesystem and operating system.
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		t.Fatalf("mkdir config directory: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	os.WriteFile(path, []byte(""), 0o644)

	var exitCode = -1
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
		panic("osExit")
	}

	defer func() {
		_ = recover()
		if exitCode != 0 {
			t.Errorf("expected softFail to exit 0 on config error, got %d", exitCode)
		}
	}()

	runHook([]string{path})
}

func TestMain_RunUpdate_Help(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"prepare-commit-msg", "update", "--help"}

	var exitCode = -1
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
		panic("osExit")
	}

	defer func() {
		_ = recover()
		if exitCode != 0 {
			t.Errorf("expected update --help to exit 0, got %d", exitCode)
		}
	}()

	main()
}

func TestRunUpdate_Flags(t *testing.T) {
	_, err := runUpdate(context.Background(), []string{"--invalid-flag-12345"})
	if err == nil {
		t.Errorf("expected error on invalid flag")
	}
}
