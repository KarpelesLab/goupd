package goupd

import "sync"

var (
	busyMutex sync.Mutex
	busyState uint32
	busyCond  = sync.NewCond(&busyMutex)
)

func busyLock() {
	busyMutex.Lock()
	for {
		if busyState <= 0 {
			return
		}
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
