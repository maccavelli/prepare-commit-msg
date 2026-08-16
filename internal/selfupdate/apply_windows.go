//go:build windows

package selfupdate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrPermissionDenied indicates that replacing the binary failed due to filesystem permissions.
var ErrPermissionDenied = errors.New("permission denied")

// ReplaceExecutable replaces targetPath with newBinaryPath on Windows systems.
// Windows locks running executables against overwriting/truncating, but permits
// renaming active binaries in NTFS.
func ReplaceExecutable(newBinaryPath, targetPath string) error {
	if realPath, err := filepath.EvalSymlinks(targetPath); err == nil {
		targetPath = realPath
	}

	oldPath := targetPath + ".old"
	_ = os.Remove(oldPath) // Remove any stale leftover .old file

	// Step 1: Rename currently running binary to .old
	if err := os.Rename(targetPath, oldPath); err != nil {
		if errors.Is(err, os.ErrPermission) || os.IsPermission(err) {
			return fmt.Errorf("%w: cannot rename running binary %s", ErrPermissionDenied, targetPath)
		}
		return fmt.Errorf("rename %s -> %s: %w", targetPath, oldPath, err)
	}

	// Step 2: Move new binary into targetPath
	var moveErr error
	for retry := 0; retry < 5; retry++ {
		moveErr = os.Rename(newBinaryPath, targetPath)
		if moveErr == nil {
			break
		}
		time.Sleep(time.Duration(50<<retry) * time.Millisecond)
	}

	if moveErr != nil {
		// Rollback: restore old binary
		_ = os.Rename(oldPath, targetPath)
		if errors.Is(moveErr, os.ErrPermission) || os.IsPermission(moveErr) {
			return fmt.Errorf("%w: cannot move new binary to %s", ErrPermissionDenied, targetPath)
		}
		return fmt.Errorf("move new binary %s -> %s: %w", newBinaryPath, targetPath, moveErr)
	}

	// Step 3: Best-effort removal of .old file (may fail if still open by current process)
	_ = os.Remove(oldPath)

	return nil
}
