package goupd

import "os"

var shutdownCallback func() error = nil
var restartCallback func() error = nil

func restartProgram() error {
	if restartCallback != nil {
		return restartCallback()
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	procAttr.Env = append(os.Environ(), "GOUPD_DELAY=5")

	_, err = os.StartProcess(exe, os.Args, procAttr)
	if err != nil {
		return err
	}

	if shutdownCallback != nil {
		return shutdownCallback()
	} else {
		os.Exit(0)
	}

	return nil
}

func SetShutdownCallback(cb func() error) {
	shutdownCallback = cb
}

func SetRestartCallback(cb func() error) {
	restartCallback = cb
}
