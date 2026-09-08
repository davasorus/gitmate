//go:build windows

package ghapi

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// hideGHWindow stops the `gh` subprocess from flashing a console window in GUI
// (windowsgui) builds — same fix as the git subprocesses in internal/gitops.
func hideGHWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
