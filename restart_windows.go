// +build windows

package goupd

import (
	"errors"
	"os"
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

	pattr := &ProcAttr{
		Env: append(os.Environ(), "GOUPD_DELAY=1"),
	}

	_, _, err = StartProcess(exe, os.Args, pattr)
	if err != nil {
		return err
	}

	os.Exit(0)
	return errors.New("program should have stopped, this message should never appear")
}

func SetRestartCallback(cb func() error) {
	restartCallback = cb
}
