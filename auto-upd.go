package goupd

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

func AutoUpdate() {
	log.Printf("[goupd] Running project %s version %s built %s", PROJECT_NAME, GIT_TAG, DATE_TAG)

	delay := os.Getenv("GOUPD_DELAY")
	if delay != "" {
		i, err := strconv.Atoi(delay)
		if err == nil {
			log.Printf("[goupd] Just upgraded, delaying program start...")
			time.Sleep(time.Duration(i) * time.Second)
		}
		os.Unsetenv("GOUPD_DELAY")
	}

	if MODE != "PROD" {
		log.Println("[goupd] Auto-updater disabled since not in production mode")
		return
	}

	go autoUpdaterThread()
}

func autoUpdaterThread() {
	for {
		time.Sleep(time.Hour)

		// stop auto-updater loop if there was an update
		if runAutoUpdateCheck() {
			return
		}
	}
}

func runAutoUpdateCheck() bool {
	// get latest version
	if PROJECT_NAME == "unconfigured" {
		log.Println("[goupd] Auto-updater failed to run, project not properly configured")
		return false
	}
	resp, err := http.Get("https://dist-go.tristandev.net/" + PROJECT_NAME + "/LATEST")
	if err != nil {
		log.Printf("[goupd] Auto-updater failed to run: %s", err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[goupd] Auto-updater failed to read latest version: %s", err)
		return false
	}

	updInfo := strings.SplitN(strings.TrimSpace(string(body)), " ", 3)
	if len(updInfo) != 3 {
		log.Printf("[goupd] Auto-updater failed to parse update data (%s)", body)
		return false
	}

	if updInfo[1] == GIT_TAG {
		log.Printf("[goupd] Current version is up to date (%s)", GIT_TAG)
		return false
	}

	log.Printf("[goupd] New version found %s/%s (current: %s/%s) - downloading...", updInfo[0], updInfo[1], DATE_TAG, GIT_TAG)

	updPrefix := updInfo[2]

	// check if compatible version is available
	resp, err = http.Get("https://dist-go.tristandev.net/" + PROJECT_NAME + "/" + updPrefix + ".arch")
	if err != nil {
		log.Printf("[goupd] Auto-updater failed to get arch info: %s", err)
		return false
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[goupd] Auto-updater failed to read arch info: %s", err)
		return false
	}

	found := false
	myself := runtime.GOOS + "_" + runtime.GOARCH

	for _, arch := range strings.Split(strings.TrimSpace(string(body)), " ") {
		if arch == myself {
			found = true
			break
		}
	}

	if !found {
		log.Printf("[goupd] Auto-updater unable to run, no version available for %s", myself)
		return false
	}

	// download actual update
	resp, err = http.Get("https://dist-go.tristandev.net/" + PROJECT_NAME + "/" + updPrefix + ".tar.xz")
	if err != nil {
		log.Printf("[goupd] Auto-updater failed to get update: %s", err)
		return false
	}
	defer resp.Body.Close()

	tarStream, err := xz.NewReader(resp.Body)
	if err != nil {
		log.Printf("[goupd] Auto-updater failed to decompress update: %s", err)
		return false
	}

	tarInfo := tar.NewReader(tarStream)

	for {
		header, err := tarInfo.Next()
		if err == io.EOF {
			log.Println("[goupd] Auto-updater failed to find appropriate version")
			return false
		}
		if err != nil {
			log.Printf("[goupd] Auto-updater failed to read update: %s", err)
			return false
		}

		if strings.HasSuffix(header.Name, myself) {
			log.Printf("[goupd] FOUND file %s", header.Name)
			err = installUpdate(tarInfo)
			if err != nil {
				log.Printf("[goupd] Auto-updater failed to install update: %s", err)
			} else {
				log.Printf("[goupd] Program upgraded, restarting")
				restartProgram()
				return true
			}
			return false
		}

		log.Printf("[goupd] Skipping file %s", header.Name)
	}
}

func installUpdate(r io.Reader) error {
	// install updated file (in io.Reader)
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	// decompose executable
	dir := filepath.Dir(exe)
	name := filepath.Base(exe)

	// copy data in new file
	newPath := filepath.Join(dir, "."+name+".new")
	fp, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = io.Copy(fp, r)
	if err != nil {
		return err
	}
	fp.Close()

	// move files
	oldPath := filepath.Join(dir, "."+name+".old")

	err = os.Rename(exe, oldPath)
	if err != nil {
		return err
	}

	err = os.Rename(newPath, exe)
	if err != nil {
		// rename failed, revert previous rename (hopefully successful)
		os.Rename(oldPath, exe)
		return err
	}

	// attempt to remove old
	err = os.Remove(oldPath)
	if err != nil {
		// hide it since remove failed
		hideFile(oldPath)
	}

	return nil
}
