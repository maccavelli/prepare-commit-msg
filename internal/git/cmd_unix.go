//go:build !windows

package git

import (
	"context"
	"os/exec"
)

func newGitCmdContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	//nolint:gosec // G204: git binary name and args are constructed by trusted callers.
	return exec.CommandContext(ctx, name, args...)
}
