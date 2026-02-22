//go:build windows

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// ============================================================
// Minimal Win32 system tray implementation.
// Runs its own message loop on a dedicated OS thread so it
// never conflicts with Wails' main thread.
// ============================================================

// Win32 constants
const (
	wmUser    = 0x0400
	wmTray    = wmUser + 69 // custom message for tray events
	wmCommand = 0x0111
	wmDestroy = 0x0002

	wmLButtonUp    = 0x0202
	wmRButtonUp    = 0x0205
	wmLButtonDblClk = 0x0203

	nimAdd    = 0x00000000
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	imageIcon      = 1
	lrLoadFromFile = 0x0010
	lrDefaultSize  = 0x0040

	mfString    = 0x0000
	mfSeparator = 0x0800

	tpmLeftAlign   = 0x0000
	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100

	idmOpen = 1001
	idmQuit = 1002
)

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
}

type wndClassExW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  uintptr
	LpszClassName *uint16
	HIconSm       uintptr
}

type point struct {
	X, Y int32
}

type msg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

var (
	user32dll   = syscall.NewLazyDLL("user32.dll")
	shell32dll  = syscall.NewLazyDLL("shell32.dll")
	kernel32dll = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW    = user32dll.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32dll.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32dll.NewProc("DefWindowProcW")
	procGetMessageW         = user32dll.NewProc("GetMessageW")
	procTranslateMessage    = user32dll.NewProc("TranslateMessage")
	procDispatchMessageW    = user32dll.NewProc("DispatchMessageW")
	procPostQuitMessage     = user32dll.NewProc("PostQuitMessage")
	procPostMessageW        = user32dll.NewProc("PostMessageW")
	procCreatePopupMenu     = user32dll.NewProc("CreatePopupMenu")
	procAppendMenuW         = user32dll.NewProc("AppendMenuW")
	procTrackPopupMenu      = user32dll.NewProc("TrackPopupMenu")
	procDestroyMenu         = user32dll.NewProc("DestroyMenu")
	procGetCursorPos        = user32dll.NewProc("GetCursorPos")
	procSetForegroundWindow = user32dll.NewProc("SetForegroundWindow")
	procLoadImageW          = user32dll.NewProc("LoadImageW")
	procSendMessageW        = user32dll.NewProc("SendMessageW")
	procFindWindowW         = user32dll.NewProc("FindWindowW")
	procEnumWindows         = user32dll.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32dll.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible     = user32dll.NewProc("IsWindowVisible")
	procGetClassNameW       = user32dll.NewProc("GetClassNameW")
	procGetModuleHandleW    = kernel32dll.NewProc("GetModuleHandleW")
	procGetCurrentProcessId = kernel32dll.NewProc("GetCurrentProcessId")
	procShellNotifyIconW    = shell32dll.NewProc("Shell_NotifyIconW")
)

// Tray holds system tray state
type Tray struct {
	hwnd    uintptr
	nid     notifyIconData
	onShow  func()
	onQuit  func()
	mu      sync.Mutex
	running bool
}

var globalTray Tray

// trayWndProc handles messages for the hidden tray window
func trayWndProc(hwnd uintptr, umsg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch umsg {
	case wmTray:
		switch lParam {
		case wmLButtonUp, wmLButtonDblClk:
			if globalTray.onShow != nil {
				go globalTray.onShow()
			}
		case wmRButtonUp:
			showTrayContextMenu(hwnd)
		}
		return 0

	case wmCommand:
		cmdID := int(wParam & 0xFFFF)
		switch cmdID {
		case idmOpen:
			if globalTray.onShow != nil {
				go globalTray.onShow()
			}
		case idmQuit:
			if globalTray.onQuit != nil {
				go globalTray.onQuit()
			}
		}
		return 0

	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(umsg), wParam, lParam)
	return ret
}

func showTrayContextMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}

	openText, _ := syscall.UTF16PtrFromString("Open / \u00d6ffnen")
	quitText, _ := syscall.UTF16PtrFromString("Quit / Beenden")

	procAppendMenuW.Call(hMenu, mfString, idmOpen, uintptr(unsafe.Pointer(openText)))
	procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)
	procAppendMenuW.Call(hMenu, mfString, idmQuit, uintptr(unsafe.Pointer(quitText)))

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// Required so the menu dismisses when clicking elsewhere
	procSetForegroundWindow.Call(hwnd)

	procTrackPopupMenu.Call(
		hMenu,
		tpmLeftAlign|tpmRightButton,
		uintptr(pt.X), uintptr(pt.Y),
		0, hwnd, 0,
	)

	procDestroyMenu.Call(hMenu)

	// Fix: send WM_NULL so the menu doesn't reappear
	procPostMessageW.Call(hwnd, 0, 0, 0)
}

// startTray creates the system tray icon on a dedicated OS thread
func startTray(iconData []byte, tooltip string, onShow func(), onQuit func()) {
	globalTray.onShow = onShow
	globalTray.onQuit = onQuit

	go func() {
		// Lock THIS goroutine to its OS thread.
		// Only affects this goroutine — does NOT touch the Wails main thread.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hInstance, _, _ := procGetModuleHandleW.Call(0)

		className, _ := syscall.UTF16PtrFromString("TarkovSwitcherTray")
		wc := wndClassExW{
			LpfnWndProc:   syscall.NewCallback(trayWndProc),
			HInstance:     hInstance,
			LpszClassName: className,
		}
		wc.CbSize = uint32(unsafe.Sizeof(wc))
		procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

		hwnd, _, _ := procCreateWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			0,
			0, 0, 0, 0, 0,
			0, 0,
			hInstance,
			0,
		)
		globalTray.hwnd = hwnd

		// Write icon to temp file so LoadImage can read it
		var hIcon uintptr
		if len(iconData) > 0 {
			tmpDir := filepath.Join(os.TempDir(), "TarkovAccountSwitcher")
			os.MkdirAll(tmpDir, 0755)
			iconPath := filepath.Join(tmpDir, "tray.ico")
			if err := os.WriteFile(iconPath, iconData, 0644); err == nil {
				iconPathW, _ := syscall.UTF16PtrFromString(iconPath)
				hIcon, _, _ = procLoadImageW.Call(
					0,
					uintptr(unsafe.Pointer(iconPathW)),
					imageIcon,
					0, 0,
					lrLoadFromFile|lrDefaultSize,
				)
			}
		}

		// Create the notify icon
		tipW, _ := syscall.UTF16FromString(tooltip)
		nid := notifyIconData{
			HWnd:             hwnd,
			UID:              1,
			UFlags:           nifMessage | nifIcon | nifTip,
			UCallbackMessage: wmTray,
			HIcon:            hIcon,
		}
		nid.CbSize = uint32(unsafe.Sizeof(nid))
		copy(nid.SzTip[:], tipW)

		procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))

		globalTray.mu.Lock()
		globalTray.nid = nid
		globalTray.running = true
		globalTray.mu.Unlock()

		// Message loop — runs until WM_QUIT
		var m msg
		for {
			ret, _, _ := procGetMessageW.Call(
				uintptr(unsafe.Pointer(&m)), 0, 0, 0,
			)
			if ret == 0 || ret == ^uintptr(0) {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
		}
	}()
}

// stopTray removes the tray icon and stops the message loop
func stopTray() {
	globalTray.mu.Lock()
	defer globalTray.mu.Unlock()

	if !globalTray.running {
		return
	}

	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&globalTray.nid)))
	if globalTray.hwnd != 0 {
		procPostMessageW.Call(globalTray.hwnd, wmDestroy, 0, 0)
	}
	globalTray.running = false
}

// ============================================================
// Window icon setter — finds the Wails window and sets its icon
// via WM_SETICON, bypassing the unreliable resource system.
// ============================================================

const (
	wmSetIcon  = 0x0080
	iconSmall  = 0
	iconBig    = 1
)

// setWindowIcon finds the main application window and sets its icon.
// Must be called after the Wails window is created (e.g., in domReady).
func setWindowIcon(iconData []byte) {
	if len(iconData) == 0 {
		return
	}

	go func() {
		// Write icon to temp file
		tmpDir := filepath.Join(os.TempDir(), "TarkovAccountSwitcher")
		os.MkdirAll(tmpDir, 0755)
		iconPath := filepath.Join(tmpDir, "window.ico")
		if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
			return
		}

		iconPathW, _ := syscall.UTF16PtrFromString(iconPath)

		// Load big icon (32x32 for taskbar/alt-tab)
		hIconBig, _, _ := procLoadImageW.Call(
			0,
			uintptr(unsafe.Pointer(iconPathW)),
			imageIcon,
			32, 32,
			lrLoadFromFile,
		)

		// Load small icon (16x16 for title bar/task manager)
		hIconSmall, _, _ := procLoadImageW.Call(
			0,
			uintptr(unsafe.Pointer(iconPathW)),
			imageIcon,
			16, 16,
			lrLoadFromFile,
		)

		if hIconBig == 0 && hIconSmall == 0 {
			return
		}

		// Find the Wails window: enumerate all windows belonging to our process
		ourPid, _, _ := procGetCurrentProcessId.Call()

		enumCallback := syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
			var pid uint32
			procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

			if uintptr(pid) != ourPid {
				return 1 // continue
			}

			// Check if visible
			visible, _, _ := procIsWindowVisible.Call(hwnd)
			if visible == 0 {
				return 1 // continue
			}

			// Check class name — Wails uses a webview class
			var className [256]uint16
			procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&className[0])), 256)

			// Set the icon on this window
			if hIconBig != 0 {
				procSendMessageW.Call(hwnd, wmSetIcon, iconBig, hIconBig)
			}
			if hIconSmall != 0 {
				procSendMessageW.Call(hwnd, wmSetIcon, iconSmall, hIconSmall)
			}

			return 1 // continue to set on all our windows
		})

		procEnumWindows.Call(enumCallback, 0)
	}()
}
