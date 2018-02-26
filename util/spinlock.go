package util

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type SpinLock interface {
	sync.Locker
	TryLock() bool
}

type SpinMutex uint32

func (s *SpinMutex) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(s), 0, 1) {
		runtime.Gosched()
	}
}

func (s *SpinMutex) Unlock() {
	atomic.StoreUint32((*uint32)(s), 0)
}

func (s *SpinMutex) TryLock() bool {
	return atomic.CompareAndSwapUint32((*uint32)(s), 0, 1)
}
