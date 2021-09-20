// +build linux darwin freebsd

package goupd

import (
	"log"
	"os"
	"path/filepath"
	"syscall"
)

var unix_exe string

func init() {
	// query os.Executable() at init time since golang used to do it, but then stopped doing it
	// see: https://go-review.googlesource.com/c/go/+/311790
	exe, err := os.Executable()
	if err == nil {
		unix_exe = exe
		return
	}

	exe, err = filepath.Abs(os.Args[0])
	if err == nil {
		unix_exe = exe
		log.Printf("[goupd] Unable to locate executable with the good method, using %s instead", unix_exe)
		return
	}
	unix_exe = os.Args[0]
	log.Printf("[goupd] Unable to locate executable with ether the good method or the bad method, using %s instead", unix_exe)
}

func RestartProgram() error {
	return syscall.Exec(unix_exe, os.Args, append(os.Environ(), "GOUPD_DELAY=1"))
}
