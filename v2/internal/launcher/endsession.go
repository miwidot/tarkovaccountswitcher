package launcher

import (
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procSendMessageTimeoutW      = user32.NewProc("SendMessageTimeoutW")

	// EnumWindows has no context pointer worth using from Go; the callback
	// fills these under enumMu.
	enumMu              sync.Mutex
	enumWindowsCallback = syscall.NewCallback(collectWindow)
	enumPIDs            map[uint32]bool
	enumHwnds           []uintptr
)

const (
	wmQueryEndSession  = 0x0011
	wmEndSession       = 0x0016
	endSessionCloseApp = 0x1
	smtoAbortIfHung    = 0x0002
)

func collectWindow(hwnd, _ uintptr) uintptr {
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if enumPIDs[pid] {
		enumHwnds = append(enumHwnds, hwnd)
	}
	return 1
}

// endSession sends the given processes' top-level windows (hidden ones
// included) the messages Windows sends when it asks an app to close for an
// update (WM_QUERYENDSESSION/WM_ENDSESSION with ENDSESSION_CLOSEAPP). Apps
// exit on these even when they only hide to the tray on WM_CLOSE. Returns
// false if no window was found.
func endSession(pids []uint32) bool {
	enumMu.Lock()
	enumPIDs = make(map[uint32]bool, len(pids))
	for _, pid := range pids {
		enumPIDs[pid] = true
	}
	enumHwnds = nil
	procEnumWindows.Call(enumWindowsCallback, 0)
	hwnds := enumHwnds
	enumPIDs, enumHwnds = nil, nil
	enumMu.Unlock()
	if len(hwnds) == 0 {
		return false
	}

	var result uintptr
	for _, hwnd := range hwnds {
		procSendMessageTimeoutW.Call(hwnd, wmQueryEndSession, 0, endSessionCloseApp,
			smtoAbortIfHung, 5000, uintptr(unsafe.Pointer(&result)))
	}
	for _, hwnd := range hwnds {
		procSendMessageTimeoutW.Call(hwnd, wmEndSession, 1, endSessionCloseApp,
			smtoAbortIfHung, 5000, uintptr(unsafe.Pointer(&result)))
	}
	return true
}
