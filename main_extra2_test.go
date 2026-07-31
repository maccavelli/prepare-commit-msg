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
	t.Setenv("XDG_CONFIG_HOME", "/dev/null/forbidden")

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
