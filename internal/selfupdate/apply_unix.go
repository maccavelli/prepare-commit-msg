//go:build !windows

package selfupdate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrPermissionDenied indicates that replacing the binary failed due to filesystem permissions.
var ErrPermissionDenied = errors.New("permission denied")

// ReplaceExecutable replaces targetPath with newBinaryPath on POSIX systems.
func ReplaceExecutable(newBinaryPath, targetPath string) error {
	// If targetPath is a symlink, resolve to the real target file.
	if realPath, err := filepath.EvalSymlinks(targetPath); err == nil {
		targetPath = realPath
	}

	// Ensure the new binary has executable permissions (rwxr-xr-x).
	if err := os.Chmod(newBinaryPath, 0o755); err != nil { //nolint:gosec // binary must have executable permissions (0755)
		if errors.Is(err, os.ErrPermission) || os.IsPermission(err) {
			return fmt.Errorf("%w: cannot set permissions on %s", ErrPermissionDenied, newBinaryPath)
		}
		return fmt.Errorf("chmod %s: %w", newBinaryPath, err)
	}

	// Atomic inode rename
	if err := os.Rename(newBinaryPath, targetPath); err != nil {
		if errors.Is(err, os.ErrPermission) || os.IsPermission(err) {
			return fmt.Errorf("%w: cannot write or replace %s", ErrPermissionDenied, targetPath)
		}
		return fmt.Errorf("rename %s -> %s: %w", newBinaryPath, targetPath, err)
	}

	return nil
}
