//go:build windows

package git

import (
	"context"
	"os/exec"
	"syscall"
)

// CREATE_NO_WINDOW prevents a console window flash when the hook is invoked
// from GUI Git clients (GitHub Desktop, VS, Sourcetree).
const createNoWindow = 0x08000000

func newGitCmdContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
	return cmd
}
