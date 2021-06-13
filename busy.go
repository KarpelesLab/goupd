package goupd

import "sync"

var busyMutex sync.RWMutex

func busyLock() {
	busyMutex.Lock()
}

func busyUnlock() {
	busyMutex.Unlock()
}

func Busy() {
	busyMutex.RLock()
}

func Unbusy() {
	busyMutex.RUnlock()
}
