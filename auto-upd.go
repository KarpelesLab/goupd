package goupd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/KarpelesLab/emitter"
)

func AutoUpdate(allowTest bool) {
	slog.Info(fmt.Sprintf("[goupd] Running project %s[%s] version %s built %s", PROJECT_NAME, CHANNEL, GIT_TAG, DATE_TAG), "event", "goupd:init:autoupdate", "goupd.project", PROJECT_NAME)

	delay := os.Getenv("GOUPD_DELAY")
	if delay != "" {
		i, err := strconv.Atoi(delay)
		if err == nil {
			slog.Debug("[goupd] Just upgraded, delaying program start...", "event", "goupd:init:delay")
			time.Sleep(time.Duration(i) * time.Second)
		}
		os.Unsetenv("GOUPD_DELAY")
	}

	if allowTest {
		go autoUpdaterThread(true)
		return
	}

	if MODE != "PROD" {
		slog.Debug("[goupd] Auto-updater disabled since not in production mode", "event", "goupd:init:devmod_disabled")
		return
	}

	// install SIGHUP handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			<-c
			RunAutoUpdateCheck()
		}
	}()

	if os.Getenv("GOUPD_NOW") != "" {
		// do an immediate run
		go autoUpdaterThread(true)
	} else {
		go autoUpdaterThread(false)
	}
}

func autoUpdaterThread(immediateInitialRun bool) {
	if !immediateInitialRun {
		time.Sleep(time.Minute)
	}
	if RunAutoUpdateCheck() {
		return
	}
	trig := emitter.Global.On("check_update")
	defer emitter.Global.Off("check_update", trig)
	tick := time.NewTicker(time.Hour)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
		case <-trig:
		}
		// stop auto-updater loop if there was an update
		if RunAutoUpdateCheck() {
			return
		}
	}
}
