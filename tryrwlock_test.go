package trylock

import (
	"reflect"
	"sync"
	"testing"
	"unsafe"
)

func TestRWMutexLayout(t *testing.T) {
	sf := reflect.TypeOf((*RWMutex)(nil)).Elem().FieldByIndex([]int{0, 0})
	if sf.Name != "w" {
		t.Fatal("sync.RWMutex first field should have name w")
	}
	if sf.Offset != uintptr(0) {
		t.Fatal("sync.RWMutex w field should have zero offset")
	}
	if sf.Type != reflect.TypeOf(sync.Mutex{}) {
		t.Fatal("sync.RWMutex w field type should be sync.Mutex")
	}

	sf = reflect.TypeOf((*RWMutex)(nil)).Elem().FieldByIndex([]int{0, 3})
	if sf.Name != "readerCount" {
		t.Fatal("sync.RWMutex third field should have name readerCount")
	}
	if off := unsafe.Offsetof(rwMutexForOffset.readerCount); sf.Offset != off {
		t.Fatalf("sync.RWMutex readerCount field should have %d offset", off)
	}
	if sf.Type != reflect.TypeOf(int32(0)) {
		t.Fatal("sync.RWMutex readerCount field type should be int32")
	}

	sf = reflect.TypeOf((*RWMutex)(nil)).Elem().FieldByIndex([]int{0, 4})
	if sf.Name != "readerWait" {
		t.Fatal("sync.RWMutex fourth field should have name readerWait")
	}
	if off := unsafe.Offsetof(rwMutexForOffset.readerWait); sf.Offset != off {
		t.Fatalf("sync.RWMutex readerWait field should have %d offset", off)
	}
	if sf.Type != reflect.TypeOf(int32(0)) {
		t.Fatal("sync.RWMutex readerWait field type should be int32")
	}
}

func TestRWMutex_TryLock(t *testing.T) {
	var rw RWMutex
	if !rw.TryLock() {
		t.Fatal("rw mutex must be unlocked")
	}
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}

	rw.Unlock()
	if !rw.TryLock() {
		t.Fatal("rw mutex must be unlocked")
	}
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}

	rw.Unlock()
	rw.Lock()
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}
	rw.Unlock()
}

func TestRWMutex_TryLockPointer(t *testing.T) {
	rw := &Mutex{}
	if !rw.TryLock() {
		t.Fatal("rw mutex must be unlocked")
	}
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}

	rw.Unlock()
	if !rw.TryLock() {
		t.Fatal("rw mutex must be unlocked")
	}
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}

	rw.Unlock()
	rw.Lock()
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}
	if rw.TryLock() {
		t.Fatal("rw mutex must be locked")
	}
	rw.Unlock()
}

func TestRWMutex_RaceLock(t *testing.T) {
	var rw RWMutex
	var x int
	for i := 0; i < 1024; i++ {
		if i%2 == 0 {
			go func() {
				if rw.TryLock() {
					x++
					rw.Unlock()
				}
			}()
			continue
		}
		go func() {
			rw.Lock()
			x++
			rw.Unlock()
		}()
	}
}

func TestRWMutex_TryRLock(t *testing.T) {
	var rw RWMutex
	if !rw.TryRLock() {
		t.Fatal("rw mutex must be unlocked")
	}
	if !rw.TryRLock() {
		t.Fatal("rw mutex must be lockable for multiple readers")
	}
	if !rw.TryRLock() {
		t.Fatal("rw mutex must be lockable for multiple readers")
	}
	rw.RUnlock()
	rw.RUnlock()
	rw.RUnlock()

	rw.Lock()
	if rw.TryRLock() {
		t.Fatal("rw mutex must be locked")
	}
	if rw.TryRLock() {
		t.Fatal("rw mutex must be locked")
	}
	rw.Unlock()
}

func TestRWMutex_TryRLockPointer(t *testing.T) {
	rw := &RWMutex{}
	if !rw.TryRLock() {
		t.Fatal("rw mutex must be unlocked")
	}
	if !rw.TryRLock() {
		t.Fatal("rw mutex must be lockable for multiple readers")
	}
	if !rw.TryRLock() {
		t.Fatal("rw mutex must be lockable for multiple readers")
	}
	rw.RUnlock()
	rw.RUnlock()
	rw.RUnlock()

	rw.Lock()
	if rw.TryRLock() {
		t.Fatal("rw mutex must be locked")
	}
	if rw.TryRLock() {
		t.Fatal("rw mutex must be locked")
	}
	rw.Unlock()
}

func TestRWMutex_RaceRLock(t *testing.T) {
	var rw RWMutex
	var x int
	dummy := func(x int) {}
	for i := 0; i < 1024; i++ {
		if i%2 == 0 {
			go func() {
				if rw.TryRLock() {
					dummy(x)
					rw.RUnlock()
				}
			}()
			continue
		}
		go func() {
			rw.Lock()
			x++
			rw.Unlock()
		}()
	}
}
