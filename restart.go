package goupd

import "os"

func restartProgram() error {
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
	os.Exit(0)
	return nil
}
