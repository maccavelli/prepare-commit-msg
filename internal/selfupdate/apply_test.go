package selfupdate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReplaceExecutable(t *testing.T) {
	tmpDir := t.TempDir()

	targetPath := filepath.Join(tmpDir, "my-app")
	if err := os.WriteFile(targetPath, []byte("version-1.0"), 0o755); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	newBinaryPath := filepath.Join(tmpDir, ".my-app.tmp-12345")
	if err := os.WriteFile(newBinaryPath, []byte("version-2.0"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := ReplaceExecutable(newBinaryPath, targetPath); err != nil {
		t.Fatalf("ReplaceExecutable failed: %v", err)
	}

	// Verify target contents updated
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("ReadFile targetPath failed: %v", err)
	}
	if string(content) != "version-2.0" {
		t.Errorf("targetPath content = %q, want version-2.0", string(content))
	}

	// Verify temporary file was moved/removed
	if _, err := os.Stat(newBinaryPath); !os.IsNotExist(err) {
		t.Errorf("newBinaryPath %s still exists", newBinaryPath)
	}

	// Verify file permissions (on Unix, should be executable 0755)
	info, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("Stat targetPath failed: %v", err)
	}
	// Windows does not implement Unix execute permission bits.
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		t.Errorf("targetPath mode %v is not executable", info.Mode())
	}
}
