// Package fsutil provides small cross-platform filesystem helpers.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteFileAtomic writes data to filename via a same-directory temp file and
// rename. On Windows the rename is not atomic and may fail under AV locks, so
// a short retry loop is used (same pattern as other fleet modules).
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(filename)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()

	// Clean up tmp file in case of early return (best-effort).
	defer func() {
		_ = tmp.Close()        //nolint:errcheck // best-effort temp cleanup
		_ = os.Remove(tmpName) //nolint:errcheck // best-effort temp cleanup
	}()

	// Unix modes are ignored on Windows; ignore Chmod errors.
	_ = tmp.Chmod(perm) //nolint:errcheck // platform-dependent
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	const maxRetries = 3
	var renameErr error
	for i := 0; i <= maxRetries; i++ {
		renameErr = os.Rename(tmpName, filename)
		if renameErr == nil {
			// Success: prevent deferred Remove from deleting the final file.
			tmpName = ""
			return nil
		}
		if i < maxRetries {
			time.Sleep(time.Duration(100<<i) * time.Millisecond)
		}
	}
	return fmt.Errorf("rename temp -> %s: %w", filename, renameErr)
}

// ReplaceFileAtomic writes content from writeFn into a temp file in the same
// directory as dest, then renames over dest with Windows-friendly retries.
// writeFn receives an open *os.File positioned at the start of the temp file.
func ReplaceFileAtomic(dest string, writeFn func(f *os.File) error) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, "COMMIT_MSG_*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()

	defer func() {
		_ = tmp.Close()        //nolint:errcheck // best-effort temp cleanup
		_ = os.Remove(tmpName) //nolint:errcheck // best-effort temp cleanup
	}()

	if err := writeFn(tmp); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	const maxRetries = 3
	var renameErr error
	for i := 0; i <= maxRetries; i++ {
		renameErr = os.Rename(tmpName, dest)
		if renameErr == nil {
			tmpName = ""
			return nil
		}
		if i < maxRetries {
			time.Sleep(time.Duration(100<<i) * time.Millisecond)
		}
	}
	return fmt.Errorf("rename temp -> %s: %w", dest, renameErr)
}
