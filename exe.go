package goupd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

var self_exe string

func init() {
	// query os.Executable() at init time since golang used to do it, but then stopped doing it
	// see: https://go-review.googlesource.com/c/go/+/311790
	exe, err := os.Executable()
	if err == nil {
		self_exe = exe
		return
	}

	exe, err = filepath.Abs(os.Args[0])
	if err == nil {
		self_exe = exe
		slog.Error(fmt.Sprintf("[goupd] Unable to locate executable with the good method, using %s instead", self_exe), "event", "goupd:init:selfexe_fallback")
		return
	}
	self_exe = os.Args[0]
	slog.Error(fmt.Sprintf("[goupd] Unable to locate executable with ether the good method or the bad method, using %s instead", self_exe), "event", "goupd:init:selfexe_fail")
}
