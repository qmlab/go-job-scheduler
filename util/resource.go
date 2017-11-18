package util

import (
	"runtime"
	"syscall"
	"unsafe"
)

// pinToCPU pin the pid to CPU index,
// usually, pid is a outside/standalone process(no write in Go)
func PinToCPU(pid int, cpu uint) error {
	const __NR_sched_setaffinity = 203
	var mask [1024 / 64]uint8
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	mask[cpu/64] |= 1 << (cpu % 64)
	_, _, errno := syscall.RawSyscall(__NR_sched_setaffinity, uintptr(pid), uintptr(len(mask)*8), uintptr(unsafe.Pointer(&mask)))
	if errno != 0 {
		return errno
	}
	return nil
}
