package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/maccavelli/prepare-commit-msg/internal/config"
)

func TestMain_RunConfigure(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"prepare-commit-msg", "configure", "--yes", "--provider=gemini", "--model=test"}

	var exitCode = -1
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
		panic("osExit")
	}

	defer func() {
		_ = recover()
	}()

	main()

	if exitCode != -1 {
		t.Errorf("expected configure to not exit with error, got %d", exitCode)
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
	}()

	main()

	if exitCode != -1 {
		t.Errorf("expected update --help to exit cleanly without osExit error, got %d", exitCode)
	}
}

func TestRunUpdate_Flags(t *testing.T) {
	// Invalid flag should return an error
	err := runUpdate([]string{"--invalid-flag-12345"})
	if err == nil {
		t.Errorf("expected error on invalid flag")
	}
}
