package goupd

import (
	"fmt"
	"log/slog"

	"github.com/KarpelesLab/shutdown"
)

// BeforeRestart is called just before the program is restarted, and can be
// used to prepare for restart, such as duplicating fds before exec/etc.
var BeforeRestart func() = shutdown.RunDefer

// RestartFunction is the functions that actually performs the restart, and
// by default will be RestartProgram which is a OS dependent implementation
var RestartFunction func() error = RestartProgram

func Restart() {
	busyLock()
	defer busyUnlock()

	if BeforeRestart != nil {
		BeforeRestart()
	}

	err := RestartFunction()
	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] restart failed: %s", err), "event", "goupd:switch_channel:restart_fail", "goupd.project", PROJECT_NAME)
	}
}
