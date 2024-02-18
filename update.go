package goupd

import (
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
)

var autoUpdateLock sync.Mutex

// SignalVersion is called when seeing another peer running the same software
// to notify of its version. This will check if the peer is updated compared
// to us, and call RunAutoUpdateCheck() if necessary
func SignalVersion(git, build string) {
	if git == "" {
		return
	}
	if git == GIT_TAG {
		return
	}

	// compare build
	if build <= DATE_TAG {
		return // we are more recent (or equal)
	}

	// perform check
	go RunAutoUpdateCheck()
}

// SignalVersionChannel performs the same as SignalVersion but will also check channel
func SignalVersionChannel(git, build, channel string) {
	if channel != CHANNEL {
		return
	}

	SignalVersion(git, build)
}

// RunAutoUpdateCheck will perform the update check, update the executable and
// return false if no update was performed. In case of update the program
// should restart and RunAutoUpdateCheck() should not return, but if it does,
// it'll return true.
func RunAutoUpdateCheck() bool {
	autoUpdateLock.Lock()
	defer autoUpdateLock.Unlock()

	// get latest version
	if PROJECT_NAME == "unconfigured" {
		slog.Error("[goupd] Auto-updater failed to run, project not properly configured", "event", "goupd:badcfg")
		return false
	}

	version, err := GetLatest(PROJECT_NAME, CHANNEL)
	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] Auto-updater failed: %s", err), "event", "goupd:update_fail", "goupd.project", PROJECT_NAME)
		return false
	}

	if version.IsCurrent() {
		// no update
		return false
	}

	slog.Info(fmt.Sprintf("[goupd] New version found %s/%s (current: %s/%s) - downloading...", version.DateTag, version.GitTag, DATE_TAG, GIT_TAG), "event", "goupd:update_found", "goupd.project", PROJECT_NAME)

	// install
	err = version.Install()
	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] Auto-updater failed: %s", err), "event", "goupd:update_fail", "goupd.project", PROJECT_NAME)
		return false
	}

	slog.Info("[goupd] Program upgraded, restarting", "event", "goupd:restart_trigger", "goupd.project", PROJECT_NAME)
	Restart()
	return true
}

// SwitchChannel will update the current running daemon to run on the given channel. It will
// return false if the running instance is already the latest version on that channel
func SwitchChannel(channel string) bool {
	if channel == CHANNEL {
		return false
	}

	autoUpdateLock.Lock()
	defer autoUpdateLock.Unlock()

	// get latest version on that channel
	if PROJECT_NAME == "unconfigured" {
		slog.Error("[goupd] Auto-updater failed to run, project not properly configured", "event", "goupd:switch_channel:badcfg")
		return false
	}

	version, err := GetLatest(PROJECT_NAME, channel)
	if err != nil {
		// maybe channel does not exist?
		return false
	}

	if err = version.CheckArch(runtime.GOOS, runtime.GOARCH); err != nil {
		// TODO report errors
		return false
	}

	slog.Info(fmt.Sprintf("[goupd] Switching to channel %s version %s/%s (current: %s/%s) - downloading...", channel, version.DateTag, version.GitTag, DATE_TAG, GIT_TAG), "event", "goupd:switch_channel:running", "goupd.project", PROJECT_NAME)
	err = version.Install()

	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] Auto-updater failed: %s", err), "event", "goupd:switch_channel:fail", "goupd.project", PROJECT_NAME)
		return false
	}

	slog.Info(fmt.Sprintf("[goupd] Program upgraded, restarting"), "event", "goupd:switch_channel:restart", "goupd.project", PROJECT_NAME)
	Restart()
	return true
}

func Fetch(projectName, curTag, os, arch, channel string, cb func(dateTag, gitTag string, r io.Reader) error) error {
	version, err := GetLatest(projectName, channel)
	if err != nil {
		return err
	}
	if version.GitTag == curTag {
		return nil
	}
	if err = version.CheckArch(os, arch); err != nil {
		return err
	}

	// download actual update
	r, err := version.Download(os, arch)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer r.Close()

	return cb(version.DateTag, version.GitTag, r)
}
