//go:build windows

package gitops

import (
	"os/exec"
	"syscall"
)

// hideWindow prevents each spawned git process from flashing a console window.
// In a GUI (windowsgui) build the parent has no console, so without this every
// git invocation creates its own terminal window — the source of the cascading
// terminal popups. CREATE_NO_WINDOW suppresses that.
const createNoWindow = 0x08000000

func hideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
