// +build windows

package goupd

import (
	"errors"
	"os"
	"syscall"
)

func RestartProgram() error {
	pattr := &syscall.ProcAttr{
		Env:   append(os.Environ(), "GOUPD_DELAY=1"),
		Files: []uintptr{uintptr(syscall.Stdin), uintptr(syscall.Stdout), uintptr(syscall.Stderr)},
	}

	_, _, err := syscall.StartProcess(self_exe, os.Args, pattr)
	if err != nil {
		return err
	}

	os.Exit(0)
	return errors.New("program should have stopped, this message should never appear")
}
