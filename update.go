package goupd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var autoUpdateLock sync.Mutex

// BeforeRestart is called just before the program is restarted, and can be
// used to prepare for restart, such as duplicating fds before exec/etc.
var BeforeRestart func()

var RestartFunction func() error = RestartProgram

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

	updated := false

	err := Fetch(PROJECT_NAME, GIT_TAG, runtime.GOOS, runtime.GOARCH, CHANNEL, func(dateTag, gitTag string, r io.Reader) error {
		slog.Info(fmt.Sprintf("[goupd] New version found %s/%s (current: %s/%s) - downloading...", dateTag, gitTag, DATE_TAG, GIT_TAG), "event", "goupd:update_found", "goupd.project", PROJECT_NAME)
		updated = true

		return installUpdate(r)
	})

	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] Auto-updater failed: %s", err), "event", "goupd:update_fail", "goupd.project", PROJECT_NAME)
		return false
	} else if !updated {
		return false
	}

	busyLock()
	defer busyUnlock()

	slog.Info("[goupd] Program upgraded, restarting", "event", "goupd:restart_trigger", "goupd.project", PROJECT_NAME)
	if BeforeRestart != nil {
		BeforeRestart()
	}
	err = RestartFunction()
	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] restart failed: %s", err), "event", "goupd:restart_fail", "goupd.project", PROJECT_NAME)
	}
	return true
}

// SwitchChannel will update the current running daemon to run on the given channel. It will
// return false if the running instance is already the latest version on that channel
func SwitchChannel(channel string) bool {
	autoUpdateLock.Lock()
	defer autoUpdateLock.Unlock()

	// get latest version on that channel
	if PROJECT_NAME == "unconfigured" {
		slog.Error("[goupd] Auto-updater failed to run, project not properly configured", "event", "goupd:switch_channel:badcfg")
		return false
	}

	updated := false

	err := Fetch(PROJECT_NAME, GIT_TAG, runtime.GOOS, runtime.GOARCH, channel, func(dateTag, gitTag string, r io.Reader) error {
		slog.Info(fmt.Sprintf("[goupd] Switching to channel %s version %s/%s (current: %s/%s) - downloading...", channel, dateTag, gitTag, DATE_TAG, GIT_TAG), "event", "goupd:switch_channel:running", "goupd.project", PROJECT_NAME)
		updated = true

		return installUpdate(r)
	})

	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] Auto-updater failed: %s", err), "event", "goupd:switch_channel:fail", "goupd.project", PROJECT_NAME)
		return false
	} else if !updated {
		return false
	}

	busyLock()
	defer busyUnlock()

	slog.Info(fmt.Sprintf("[goupd] Program upgraded, restarting"), "event", "goupd:switch_channel:restart", "goupd.project", PROJECT_NAME)
	if BeforeRestart != nil {
		BeforeRestart()
	}
	err = RestartFunction()
	if err != nil {
		slog.Error(fmt.Sprintf("[goupd] restart failed: %s", err), "event", "goupd:switch_channel:restart_fail", "goupd.project", PROJECT_NAME)
	}
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

func installUpdate(r io.Reader) error {
	// install updated file (in io.Reader)
	exe, err := filepath.EvalSymlinks(self_exe)
	if err != nil {
		return fmt.Errorf("failed to find exe: %w", err)
	}

	// decompose executable
	dir := filepath.Dir(exe)
	name := filepath.Base(exe)

	// copy data in new file
	newPath := filepath.Join(dir, "."+name+".new")
	fp, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}
	defer fp.Close()

	_, err = io.Copy(fp, r)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	err = fp.Close()
	if err != nil {
		// delayed error because disk full?
		return fmt.Errorf("close failed: %w", err)
	}

	// move files
	oldPath := filepath.Join(dir, "."+name+".old")

	err = os.Rename(exe, oldPath)
	if err != nil {
		return fmt.Errorf("update rename failed: %w", err)
	}

	err = os.Rename(newPath, exe)
	if err != nil {
		// rename failed, revert previous rename (hopefully successful)
		os.Rename(oldPath, exe)
		return fmt.Errorf("update second rename failed: %w", err)
	}

	// attempt to remove old
	err = os.Remove(oldPath)
	if err != nil {
		// hide it since remove failed
		hideFile(oldPath)
	}

	return nil
}
