// +build windows

package goupd

import (
	"errors"
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

	pattr := &syscall.ProcAttr{
		Env:   append(os.Environ(), "GOUPD_DELAY=1"),
		Files: []uintptr{uintptr(syscall.Stdin), uintptr(syscall.Stdout), uintptr(syscall.Stderr)},
	}

	_, _, err = syscall.StartProcess(exe, os.Args, pattr)
	if err != nil {
		return err
	}

	os.Exit(0)
	return errors.New("program should have stopped, this message should never appear")
}

func SetRestartCallback(cb func() error) {
	restartCallback = cb
}
