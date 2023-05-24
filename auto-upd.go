package goupd

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func AutoUpdate(allowTest bool) {
	log.Printf("[goupd] Running project %s[%s] version %s built %s", PROJECT_NAME, CHANNEL, GIT_TAG, DATE_TAG)

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

	for {
		time.Sleep(time.Hour)

		// stop auto-updater loop if there was an update
		if RunAutoUpdateCheck() {
			return
		}
	}
}
