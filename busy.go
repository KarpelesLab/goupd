package goupd

import "sync/atomic"

var busyCnt uint64

func BusyAdd(n uint64) {
	atomic.AddUint64(&busyCnt, n)
}

func BusySub(n uint64) {
	atomic.AddUint64(&busyCnt, ^uint64(n-1))
}

func Busy() {
	atomic.AddUint64(&busyCnt, 1)
}

func Unbusy() {
	atomic.AddUint64(&busyCnt, ^uint64(0))
}

func BusyState() uint64 {
	return atomic.LoadUint64(&busyCnt)
}
