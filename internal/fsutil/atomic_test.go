package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")

	if err := WriteFileAtomic(path, []byte(`{"a":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"a":1}` {
		t.Fatalf("got %q", data)
	}

	// Replace existing
	if err := WriteFileAtomic(path, []byte(`{"a":2}`), 0o600); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != `{"a":2}` {
		t.Fatalf("replace got %q", data)
	}
}

func TestReplaceFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	_ = os.WriteFile(path, []byte("old"), 0o644)

	err := ReplaceFileAtomic(path, func(f *os.File) error {
		_, err := f.WriteString("new content")
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "new content" {
		t.Fatalf("got %q", data)
	}
}
