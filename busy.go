package goupd

import (
	"sync"
	"time"
)

var (
	busyMutex sync.Mutex
	busyState uint32
	busyCond  = sync.NewCond(&busyMutex)
)

func busyLock() {
	var s sync.Once
	timeout := time.Now().Add(1 * time.Hour)

	busyMutex.Lock()
	for {
		if busyState <= 0 {
			return
		}

		if time.Until(timeout) <= 0 {
			// let's consider this is good
			return
		}

		s.Do(func() {
			go func() {
				time.Sleep(time.Until(timeout))
				// force wake up once we reach timeout
				busyCond.Broadcast()
			}()
		})
		// other tasks are running, wait for busyState to decrease
		busyCond.Wait()
	}
}

func busyUnlock() {
	busyMutex.Unlock()
}

func Busy() {
	busyMutex.Lock()
	defer busyMutex.Unlock()
	busyState += 1
}

func Unbusy() {
	busyMutex.Lock()
	defer busyMutex.Unlock()
	busyState -= 1

	if busyState <= 0 {
		busyCond.Broadcast()
	}
}
