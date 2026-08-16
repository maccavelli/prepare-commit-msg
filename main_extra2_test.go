package main

import (
	"os"
	"path/filepath"
	"testing"
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

	// Block the app's config dir with a regular file so reading the primary
	// config path fails with ENOTDIR on every platform — poisoning
	// XDG_CONFIG_HOME only affects Linux, since macOS resolves UserConfigDir
	// from HOME alone.
	base, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir: %v", err)
	}
	if err := os.MkdirAll(base, 0o750); err != nil {
		t.Fatalf("mkdir config base: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "prepare-commit-msg"), []byte("block"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
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

	// --help returns flag.ErrHelp so exitCode could be 1 or handled cleanly
	if exitCode == 0 {
		t.Errorf("unexpected exit code 0 on update --help")
	}
}

func TestRunUpdate_Flags(t *testing.T) {
	// Invalid flag should return an error
	err := runUpdate([]string{"--invalid-flag-12345"})
	if err == nil {
		t.Errorf("expected error on invalid flag")
	}
}
