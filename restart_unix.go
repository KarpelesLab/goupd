// +build linux darwin freebsd

package goupd

import (
	"os"
	"syscall"
)

func RestartProgram() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	return syscall.Exec(exe, os.Args, append(os.Environ(), "GOUPD_DELAY=1"))
}
