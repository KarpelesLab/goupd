package goupd

import (
	"os"
	"syscall"
)

var restartCallback func() error = nil

func restartProgram() error {
	if restartCallback != nil {
		return restartCallback()
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	return syscall.Exec(exe, os.Args, append(os.Environ(), "GOUPD_DELAY=5"))
}

func SetRestartCallback(cb func() error) {
	restartCallback = cb
}
