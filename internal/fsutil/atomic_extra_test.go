package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic_Errors(t *testing.T) {
	// 1. mkdir failure
	err := WriteFileAtomic("/proc/sys/fs/file-max/forbidden", []byte("bad"), 0o644)
	if err == nil {
		t.Error("expected mkdir error")
	}

	// 2. create temp failure (directory exists but cannot host regular files)
	err = WriteFileAtomic("/proc/test.txt", []byte("bad"), 0o644)
	if err == nil {
		t.Error("expected create temp error")
	}
}

func TestReplaceFileAtomic_Errors(t *testing.T) {
	// 1. mkdir failure
	err := ReplaceFileAtomic("/proc/sys/fs/file-max/forbidden", func(f *os.File) error {
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
