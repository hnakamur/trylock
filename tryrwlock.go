package trylock

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const rwmutexMaxReaders = 1 << 30

// RWMutex is simple sync.RWMutex + ability to try to Lock.
type RWMutex struct {
	in sync.RWMutex
}

type rwMutexTypeForOffset struct {
	w           sync.Mutex // held if there are pending writers
	writerSem   uint32     // semaphore for writers to wait for completing readers
	readerSem   uint32     // semaphore for readers to wait for completing writers
	readerCount int32      // number of pending readers
	readerWait  int32      // number of departing readers
}

var rwMutexForOffset rwMutexTypeForOffset

// Lock locks rw for writing.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (rw *RWMutex) Lock() {
	rw.in.Lock()
}

// Unlock unlocks rw for writing.
// It is a run-time error if rw is not locked on entry to Unlock.
//
// A locked RWMutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (rw *RWMutex) Unlock() {
	rw.in.Unlock()
}

// Lock locks rw for reading.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (rw *RWMutex) RLock() {
	rw.in.RLock()
}

// Unlock unlocks rw for reading.
// It is a run-time error if rw is not locked on entry to RUnlock.
//
// A locked RWMutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (rw *RWMutex) RUnlock() {
	rw.in.RUnlock()
}

// RTryLock tries to lock rw for reading. It returns true in
// case of success, false otherwise.
func (rw *RWMutex) TryRLock() bool {
	p := (*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&rw.in)) + unsafe.Offsetof(rwMutexForOffset.readerCount)))
	if r := atomic.AddInt32(p, 1); r < 0 {
		atomic.AddInt32(p, -1)
		return false
	}
	return true
}

// TryLock tries to lock rw for writing. It returns true in
// case of success, false otherwise.
func (rw *RWMutex) TryLock() bool {
	pW := (*int32)(unsafe.Pointer(&rw.in))
	pReaderCount := (*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&rw.in)) + unsafe.Offsetof(rwMutexForOffset.readerCount)))
	pReaderWait := (*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&rw.in)) + unsafe.Offsetof(rwMutexForOffset.readerWait)))

	if !atomic.CompareAndSwapInt32(pW, 0, mutexLocked) {
		return false
	}

	r := atomic.AddInt32(pReaderCount, -rwmutexMaxReaders) + rwmutexMaxReaders
	if r != 0 && atomic.AddInt32(pReaderWait, r) != 0 {
		atomic.AddInt32(pReaderWait, -r)
		atomic.AddInt32(pReaderCount, rwmutexMaxReaders)
		atomic.StoreInt32(pW, 0)
		return false
	}
	return true
}
