// +build windows

package goupd

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

var win_exe string

func init() {
	exe, err := os.Executable()
	if err == nil {
		win_exe = exe
		return
	}

	exe, err := filepath.Abs(os.Args[0])
	if err == nil {
		win_exe = exe
		return
	}

	win_exe = os.Args[0] // ???
}

func RestartProgram() error {
	pattr := &syscall.ProcAttr{
		Env:   append(os.Environ(), "GOUPD_DELAY=1"),
		Files: []uintptr{uintptr(syscall.Stdin), uintptr(syscall.Stdout), uintptr(syscall.Stderr)},
	}

	_, _, err = syscall.StartProcess(win_exe, os.Args, pattr)
	if err != nil {
		return err
	}

	os.Exit(0)
	return errors.New("program should have stopped, this message should never appear")
}
