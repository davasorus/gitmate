//go:build !windows

package ghapi

import "os/exec"

func hideGHWindow(_ *exec.Cmd) {}
