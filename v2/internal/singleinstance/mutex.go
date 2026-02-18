package singleinstance

import (
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	createMutexW = kernel32.NewProc("CreateMutexW")

	mutexHandle uintptr
)

// Lock attempts to acquire a single instance lock
func Lock(name string) bool {
	mutexName, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return true // Allow on error
	}

	handle, _, err := createMutexW.Call(
		0,
		0,
		uintptr(unsafe.Pointer(mutexName)),
	)

	if handle == 0 {
		return true // Allow on error
	}

	// ERROR_ALREADY_EXISTS = 183
	if err == syscall.Errno(183) {
		return false
	}

	mutexHandle = handle
	return true
}

// Unlock releases the lock
func Unlock() {
	// Don't close - let Windows clean up on exit
}
