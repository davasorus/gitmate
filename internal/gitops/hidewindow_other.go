//go:build !windows

package gitops

import "os/exec"

// hideWindow is a no-op on non-Windows platforms (no console-window flashing).
func hideWindow(_ *exec.Cmd) {}
