package ui

import (
	"sync"
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
	dmTrackMouse  = dmUser32.NewProc("TrackMouseEvent")
	dmFillRect    = dmUser32.NewProc("FillRect")

	dmGdi32            = syscall.NewLazyDLL("gdi32.dll")
	dmCreateSolidBrush = dmGdi32.NewProc("CreateSolidBrush")
	dmDeleteObject     = dmGdi32.NewProc("DeleteObject")
	dmSetBkMode        = dmGdi32.NewProc("SetBkMode")
	dmSetTextColor     = dmGdi32.NewProc("SetTextColor")
	dmSelectObject     = dmGdi32.NewProc("SelectObject")

	uxthemeHandle syscall.Handle

	pSetPreferredAppMode              uintptr
	pAllowDarkModeForWindow           uintptr
	pFlushMenuThemes                  uintptr
	pRefreshImmersiveColorPolicyState uintptr

	darkAPIsOK      bool
	darkBtnCallback uintptr
	darkBtnOldProcs sync.Map // win.HWND -> uintptr
	darkBtnHover    sync.Map // win.HWND -> bool
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

	darkBtnCallback = syscall.NewCallback(darkBtnWndProc)
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

// ---------------------------------------------------------------------------
// Owner-draw dark buttons — subclass each PushButton's WndProc to handle
// WM_PAINT with custom dark drawing (same approach as Notepad++).
// ---------------------------------------------------------------------------

type trackMouseEventInfo struct {
	cbSize      uint32
	dwFlags     uint32
	hwndTrack   uintptr
	dwHoverTime uint32
}

const (
	tmeLeave      = 0x00000002
	wmMouseLeave  = 0x02A3
	wmNCDestroy   = 0x0082
	bstPushed     = 0x0004
	transparentBk = 1
)

// subclassDarkButton replaces the button's WndProc with our dark painter.
func subclassDarkButton(btn *walk.PushButton) {
	hwnd := btn.Handle()
	if _, exists := darkBtnOldProcs.Load(hwnd); exists {
		return
	}
	oldProc := win.SetWindowLongPtr(hwnd, win.GWLP_WNDPROC, darkBtnCallback)
	darkBtnOldProcs.Store(hwnd, oldProc)
	win.InvalidateRect(hwnd, nil, true)
}

func darkBtnWndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	oldProcVal, ok := darkBtnOldProcs.Load(hwnd)
	if !ok {
		return win.DefWindowProc(hwnd, msg, wParam, lParam)
	}
	oldProc := oldProcVal.(uintptr)

	switch msg {
	case win.WM_PAINT:
		paintDarkButton(hwnd)
		return 0

	case win.WM_ERASEBKGND:
		return 1

	case win.WM_MOUSEMOVE:
		if _, loaded := darkBtnHover.LoadOrStore(hwnd, true); !loaded {
			// Start tracking for mouse leave
			var tme trackMouseEventInfo
			tme.cbSize = uint32(unsafe.Sizeof(tme))
			tme.dwFlags = tmeLeave
			tme.hwndTrack = uintptr(hwnd)
			dmTrackMouse.Call(uintptr(unsafe.Pointer(&tme)))
			win.InvalidateRect(hwnd, nil, false)
		}
		return win.CallWindowProc(oldProc, hwnd, msg, wParam, lParam)

	case wmMouseLeave:
		darkBtnHover.Delete(hwnd)
		win.InvalidateRect(hwnd, nil, false)
		return win.CallWindowProc(oldProc, hwnd, msg, wParam, lParam)

	case wmNCDestroy:
		darkBtnOldProcs.Delete(hwnd)
		darkBtnHover.Delete(hwnd)
		return win.CallWindowProc(oldProc, hwnd, msg, wParam, lParam)

	default:
		return win.CallWindowProc(oldProc, hwnd, msg, wParam, lParam)
	}
}

func paintDarkButton(hwnd win.HWND) {
	t := CurrentTheme()

	var ps win.PAINTSTRUCT
	hdc := win.BeginPaint(hwnd, &ps)
	if hdc == 0 {
		return
	}
	defer win.EndPaint(hwnd, &ps)

	var rc win.RECT
	win.GetClientRect(hwnd, &rc)

	// Determine state
	state := win.SendMessage(hwnd, win.BM_GETSTATE, 0, 0)
	_, hovering := darkBtnHover.Load(hwnd)

	var bgColor, textColor, borderColor walk.Color
	if state&bstPushed != 0 {
		bgColor = t.AccentColor
		textColor = walk.RGB(255, 255, 255)
		borderColor = t.AccentColor
	} else if hovering {
		bgColor = t.CardBg
		textColor = t.TextPrimary
		borderColor = t.AccentColor
	} else {
		bgColor = t.SurfaceBg
		textColor = t.TextPrimary
		borderColor = t.TextSecondary
	}

	// Draw border (fill full rect with border color, then fill inner with bg)
	borderBrush, _, _ := dmCreateSolidBrush.Call(uintptr(borderColor))
	dmFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&rc)), borderBrush)
	dmDeleteObject.Call(borderBrush)

	innerRc := rc
	innerRc.Left++
	innerRc.Top++
	innerRc.Right--
	innerRc.Bottom--
	bgBrush, _, _ := dmCreateSolidBrush.Call(uintptr(bgColor))
	dmFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&innerRc)), bgBrush)
	dmDeleteObject.Call(bgBrush)

	// Set text mode
	dmSetBkMode.Call(uintptr(hdc), transparentBk)
	dmSetTextColor.Call(uintptr(hdc), uintptr(textColor))

	// Select button font
	hFont := win.SendMessage(hwnd, win.WM_GETFONT, 0, 0)
	if hFont != 0 {
		oldFont, _, _ := dmSelectObject.Call(uintptr(hdc), hFont)
		defer dmSelectObject.Call(uintptr(hdc), oldFont)
	}

	// Get button text and draw centered
	textLen := win.SendMessage(hwnd, win.WM_GETTEXTLENGTH, 0, 0)
	if textLen > 0 {
		buf := make([]uint16, textLen+1)
		win.SendMessage(hwnd, win.WM_GETTEXT, uintptr(textLen+1), uintptr(unsafe.Pointer(&buf[0])))
		win.DrawTextEx(hdc, &buf[0], int32(textLen), &innerRc,
			win.DT_CENTER|win.DT_VCENTER|win.DT_SINGLELINE, nil)
	}
}
