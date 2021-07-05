package goupd

import (
	"sync"
	"testing"
	"time"
)

type atomicString struct {
	v string
	l sync.Mutex
}

func (a *atomicString) Append(v string) {
	a.l.Lock()
	defer a.l.Unlock()
	a.v += v
}

func (a *atomicString) String() string {
	return a.v
}

func TestBusy(t *testing.T) {
	as := &atomicString{}

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			as.Append("0")
			time.Sleep(100 * time.Millisecond)
			as.Append("1")
			time.Sleep(100 * time.Millisecond)
			Busy()
			as.Append("2")
			time.Sleep(100 * time.Millisecond)
			as.Append("3")
			Unbusy()
		}()
	}

	time.Sleep(50 * time.Millisecond)
	busyLock()
	as.Append("!")
	time.Sleep(200 * time.Millisecond)
	as.Append("R")
	busyUnlock()
	time.Sleep(10 * time.Millisecond)
	busyLock()
	as.Append("L")
	busyUnlock()

	wg.Wait()

	if as.String() != "000!111R222333L" {
		t.Errorf("failed test, was expecting 000!111R222333L but got %s", as)
	}

	//log.Printf("as = %s", as)
	// 000!111R222333L
}
