package goupd

import (
	"log"
	"os"
	"strconv"
	"time"
)

func AutoUpdate(allowTest bool) {
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

	if allowTest {
		go autoUpdaterThread(true)
		return
	}

	if MODE != "PROD" {
		log.Println("[goupd] Auto-updater disabled since not in production mode")
		return
	}

	go autoUpdaterThread(false)
}

func autoUpdaterThread(initialRun bool) {
	if initialRun {
		if RunAutoUpdateCheck() {
			return
		}
	}

	for {
		time.Sleep(time.Hour)

		// stop auto-updater loop if there was an update
		if RunAutoUpdateCheck() {
			return
		}
	}
}
