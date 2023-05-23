package goupd

import (
	"compress/bzip2"
	"fmt"
	"io"
	"log"
	"net/http"
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

// RunAutoUpdateCheck will perform the update check, update the executable and
// return false if no update was performed. In case of update the program
// should restart and RunAutoUpdateCheck() should not return, but if it does,
// it'll return true.
func RunAutoUpdateCheck() bool {
	autoUpdateLock.Lock()
	defer autoUpdateLock.Unlock()

	// get latest version
	if PROJECT_NAME == "unconfigured" {
		log.Println("[goupd] Auto-updater failed to run, project not properly configured")
		return false
	}

	updated := false

	err := Fetch(PROJECT_NAME, GIT_TAG, runtime.GOOS, runtime.GOARCH, func(dateTag, gitTag string, r io.Reader) error {
		log.Printf("[goupd] New version found %s/%s (current: %s/%s) - downloading...", dateTag, gitTag, DATE_TAG, GIT_TAG)
		updated = true

		return installUpdate(r)
	})

	if err != nil {
		log.Printf("[goupd] Auto-updater failed: %s", err)
		return false
	} else if !updated {
		return false
	}

	busyLock()
	defer busyUnlock()

	log.Printf("[goupd] Program upgraded, restarting")
	if BeforeRestart != nil {
		BeforeRestart()
	}
	err = RestartFunction()
	if err != nil {
		log.Printf("[goupd] restart failed: %s", err)
	}
	return true
}

func Fetch(projectName, curTag, os, arch string, cb func(dateTag, gitTag string, r io.Reader) error) error {
	dlUrl, dateTag, gitTag, err := GetUpdate(projectName, curTag, os, arch)
	if err != nil {
		return err
	}
	if gitTag == curTag {
		return nil
	}
	if dlUrl == "" {
		return nil
	}

	// download actual update
	resp, err := http.Get(dlUrl)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	var r io.Reader
	r = resp.Body

	r = bzip2.NewReader(r)

	return cb(dateTag, gitTag, r)
}

func GetUpdate(projectName, curTag, os, arch string) (string, string, string, error) {
	latest := HOST + projectName + "/LATEST"
	if CHANNEL != "" {
		// for example LATEST-testing
		latest += "-" + CHANNEL
	}
	updInfo, err := httpGetFields(latest)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read latest version: %w", err)
	}
	if len(updInfo) != 3 {
		return "", "", "", fmt.Errorf("failed to parse update data (%v)", updInfo)
	}

	dateTag := updInfo[0]   // 20230518035112
	gitTag := updInfo[1]    // e894f37
	updPrefix := updInfo[2] // packagename-stable_20230518035112_e894f37

	target := os + "_" + arch
	dlUrl := HOST + projectName + "/" + updPrefix + "/" + projectName + "_" + target + ".bz2"

	if curTag == updInfo[1] {
		// no update needed, don't perform arch check
		return dlUrl, dateTag, gitTag, nil
	}

	// check if compatible version is available
	archs, err := httpGetFields(HOST + projectName + "/" + updPrefix + ".arch")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read arch info: %w", err)
	}

	found := false

	for _, subarch := range archs {
		if subarch == target {
			found = true
			break
		}
	}

	if !found {
		return "", "", "", fmt.Errorf("no version available for %s", target)
	}

	return dlUrl, dateTag, gitTag, nil
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
