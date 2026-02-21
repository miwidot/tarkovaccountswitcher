package singleinstance

import (
	"syscall"
	"unsafe"
)

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	createMutexW   = kernel32.NewProc("CreateMutexW")
	createEventW   = kernel32.NewProc("CreateEventW")
	openEventW     = kernel32.NewProc("OpenEventW")
	setEvent       = kernel32.NewProc("SetEvent")
	waitForSingle  = kernel32.NewProc("WaitForSingleObject")

	mutexHandle uintptr
	eventHandle uintptr
)

const eventName = "TarkovAccountSwitcher_v2_Show"

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

// CreateShowEvent creates a named event that other instances can signal.
func CreateShowEvent() {
	evName, err := syscall.UTF16PtrFromString(eventName)
	if err != nil {
		return
	}
	// Auto-reset event, initially non-signaled
	h, _, _ := createEventW.Call(0, 0, 0, uintptr(unsafe.Pointer(evName)))
	eventHandle = h
}

// WaitForShowSignal blocks until the event is signaled, then calls callback.
// Runs in a loop so it works for repeated signals.
func WaitForShowSignal(callback func()) {
	if eventHandle == 0 {
		return
	}
	for {
		// INFINITE = 0xFFFFFFFF
		ret, _, _ := waitForSingle.Call(eventHandle, 0xFFFFFFFF)
		if ret == 0 { // WAIT_OBJECT_0
			callback()
		}
	}
}

// SignalExisting signals the existing instance to show its window.
func SignalExisting() {
	evName, err := syscall.UTF16PtrFromString(eventName)
	if err != nil {
		return
	}
	// EVENT_MODIFY_STATE = 0x0002
	h, _, _ := openEventW.Call(0x0002, 0, uintptr(unsafe.Pointer(evName)))
	if h != 0 {
		setEvent.Call(h)
		syscall.CloseHandle(syscall.Handle(h))
	}
}

// Unlock releases the lock
func Unlock() {
	// Don't close - let Windows clean up on exit
}
