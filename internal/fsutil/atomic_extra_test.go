package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic_Errors(t *testing.T) {
	// 1. mkdir failure: a regular file cannot become a parent directory.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("block"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	err := WriteFileAtomic(filepath.Join(blocker, "child", "test.txt"), []byte("bad"), 0o644)
	if err == nil {
		t.Error("expected mkdir error")
	}

	// 2. create temp failure: NUL is invalid in a filename on every supported OS.
	err = WriteFileAtomic(filepath.Join(t.TempDir(), "bad\x00name"), []byte("bad"), 0o644)
	if err == nil {
		t.Error("expected create temp error")
	}
}

func TestReplaceFileAtomic_Errors(t *testing.T) {
	// 1. create temp failure: the destination's parent is a regular file.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("block"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	err := ReplaceFileAtomic(filepath.Join(blocker, "test.txt"), func(f *os.File) error {
		return nil
	})
	if err == nil {
		t.Error("expected mkdir error")
	}

	// 2. writer error
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err = ReplaceFileAtomic(path, func(f *os.File) error {
		return errors.New("writer failed")
	})
	if err == nil || err.Error() != "writer failed" {
		t.Errorf("expected writer failed error, got %v", err)
	}
}
