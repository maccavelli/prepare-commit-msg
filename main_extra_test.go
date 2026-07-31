package main

import (
	"os"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	// Should just not panic.
	printUsage()
}

func TestMultiFlag(t *testing.T) {
	var mf multiFlag
	err := mf.Set("a, b, c ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mf.String() != "a,b,c" {
		t.Errorf("expected a,b,c got %q", mf.String())
	}
}

func TestSoftFail(t *testing.T) {
	var exitCode int
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
		panic("osExit")
	}

	defer func() {
		_ = recover()
		if exitCode != 0 {
			t.Errorf("expected softFail to exit 0, got %d", exitCode)
		}
	}()

	softFail("test %s", "message")
}

func TestMain_NoArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"prepare-commit-msg"} // length 1 -> args[1:] length 0

	var exitCode = -1
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
		panic("osExit called")
	}

	defer func() {
		_ = recover()
		if exitCode != 1 {
			t.Errorf("expected main with no args to exit 1, got %d", exitCode)
		}
	}()

	main()
}

func TestMain_Help(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"prepare-commit-msg", "--help"}

	var exitCode = -1
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
	}

	main()

	if exitCode != -1 {
		t.Errorf("expected main with --help to not exit, got %d", exitCode)
	}
}

func TestMain_Version(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"prepare-commit-msg", "version"}

	var exitCode = -1
	oldExit := osExit
	defer func() { osExit = oldExit }()
	osExit = func(code int) {
		exitCode = code
	}

	main()

	if exitCode != -1 {
		t.Errorf("expected main with version to not exit, got %d", exitCode)
	}
}
