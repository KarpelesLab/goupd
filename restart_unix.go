// +build linux darwin freebsd

package goupd

import (
	"os"
	"syscall"
)

func RestartProgram() error {
	return syscall.Exec(self_exe, os.Args, append(os.Environ(), "GOUPD_DELAY=1"))
}
