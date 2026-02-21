package ui

import (
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

// Undocumented uxtheme.dll dark mode APIs (loaded by ordinal).
// Used by Notepad++, Firefox, etc. for native Win32 dark mode.

var (
	dmKernel32    = syscall.NewLazyDLL("kernel32.dll")
	dmGetProcAddr = dmKernel32.NewProc("GetProcAddress")
	dmUser32      = syscall.NewLazyDLL("user32.dll")
	dmGetWindow   = dmUser32.NewProc("GetWindow")

	uxthemeHandle syscall.Handle

	pSetPreferredAppMode              uintptr
	pAllowDarkModeForWindow           uintptr
	pFlushMenuThemes                  uintptr
	pRefreshImmersiveColorPolicyState uintptr

	darkAPIsOK bool
)

func init() {
	h, err := syscall.LoadLibrary("uxtheme.dll")
	if err != nil {
		return
	}
	uxthemeHandle = h

	pSetPreferredAppMode = uxthemeOrdinal(h, 135)
	pAllowDarkModeForWindow = uxthemeOrdinal(h, 133)
	pFlushMenuThemes = uxthemeOrdinal(h, 136)
	pRefreshImmersiveColorPolicyState = uxthemeOrdinal(h, 104)

	darkAPIsOK = pSetPreferredAppMode != 0 && pAllowDarkModeForWindow != 0
}

func uxthemeOrdinal(module syscall.Handle, ordinal uint16) uintptr {
	ret, _, _ := dmGetProcAddr.Call(uintptr(module), uintptr(ordinal))
	return ret
}

const (
	appModeForceDark  = 2
	appModeForceLight = 3
)

// enableSystemDarkMode tells Windows this app wants dark/light controls.
// Must be called before window creation for best effect, and again on theme switch.
func enableSystemDarkMode(dark bool) {
	if !darkAPIsOK {
		return
	}
	mode := uintptr(appModeForceLight)
	if dark {
		mode = uintptr(appModeForceDark)
	}
	syscall.Syscall(pSetPreferredAppMode, 1, mode, 0, 0)
	if pRefreshImmersiveColorPolicyState != 0 {
		syscall.Syscall(pRefreshImmersiveColorPolicyState, 0, 0, 0, 0)
	}
	if pFlushMenuThemes != 0 {
		syscall.Syscall(pFlushMenuThemes, 0, 0, 0, 0)
	}
}

// allowDarkModeForHWND enables/disables dark mode on a specific window handle.
func allowDarkModeForHWND(hwnd win.HWND, allow bool) {
	if pAllowDarkModeForWindow == 0 {
		return
	}
	var val uintptr
	if allow {
		val = 1
	}
	syscall.Syscall(pAllowDarkModeForWindow, 2, uintptr(hwnd), val, 0)
}

const wmThemeChanged = 0x031A

// setControlDarkMode applies dark mode + SetWindowTheme to a control.
func setControlDarkMode(hwnd win.HWND, dark bool, themeName string) {
	if !darkAPIsOK {
		return
	}
	allowDarkModeForHWND(hwnd, dark)
	if themeName != "" {
		tn, _ := syscall.UTF16PtrFromString(themeName)
		win.SetWindowTheme(hwnd, tn, nil)
	} else {
		win.SetWindowTheme(hwnd, nil, nil)
	}
	win.SendMessage(hwnd, wmThemeChanged, 0, 0)
}

// comboBoxInfo mirrors the Win32 COMBOBOXINFO struct.
type comboBoxInfo struct {
	cbSize      uint32
	rcItem      win.RECT
	rcButton    win.RECT
	stateButton uint32
	hwndCombo   win.HWND
	hwndItem    win.HWND
	hwndList    win.HWND
}

const cbGetComboBoxInfo = 0x0164

// setComboBoxDarkMode applies dark mode to a ComboBox and its dropdown.
func setComboBoxDarkMode(cb *walk.ComboBox, dark bool) {
	hwnd := cb.Handle()
	if dark {
		setControlDarkMode(hwnd, true, "DarkMode_CFD")
	} else {
		setControlDarkMode(hwnd, false, "")
	}

	var info comboBoxInfo
	info.cbSize = uint32(unsafe.Sizeof(info))
	win.SendMessage(hwnd, cbGetComboBoxInfo, 0, uintptr(unsafe.Pointer(&info)))

	if info.hwndList != 0 {
		if dark {
			setControlDarkMode(info.hwndList, true, "DarkMode_Explorer")
		} else {
			setControlDarkMode(info.hwndList, false, "")
		}
	}
	if info.hwndItem != 0 {
		if dark {
			setControlDarkMode(info.hwndItem, true, "DarkMode_CFD")
		} else {
			setControlDarkMode(info.hwndItem, false, "")
		}
	}
}

// setTabWidgetDarkMode applies dark mode to the TabWidget's SysTabControl32.
func setTabWidgetDarkMode(tw *walk.TabWidget, dark bool) {
	if !darkAPIsOK || tw == nil {
		return
	}
	// SysTabControl32 is the first child of the TabWidget
	const gwChild = 5
	child, _, _ := dmGetWindow.Call(uintptr(tw.Handle()), gwChild)
	if child != 0 {
		if dark {
			setControlDarkMode(win.HWND(child), true, "DarkMode_Explorer")
		} else {
			setControlDarkMode(win.HWND(child), false, "")
		}
	}
}
