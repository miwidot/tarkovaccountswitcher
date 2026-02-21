package ui

import (
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"tarkov-account-switcher/internal/config"
)

var (
	libdwmapi                *windows.LazyDLL
	procDwmSetWindowAttribute *windows.LazyProc
)

func init() {
	libdwmapi = windows.NewLazySystemDLL("dwmapi.dll")
	procDwmSetWindowAttribute = libdwmapi.NewProc("DwmSetWindowAttribute")
}

// Theme defines all semantic colors for the UI.
type Theme struct {
	Name string

	WindowBg  walk.Color
	SurfaceBg walk.Color
	CardBg    walk.Color

	TextPrimary   walk.Color
	TextSecondary walk.Color
	TextTitle     walk.Color

	StatusSuccess walk.Color
	StatusWarning walk.Color
	StatusError   walk.Color

	AccentColor walk.Color

	UseDarkWindowChrome bool
}

var themes = map[string]*Theme{
	"light": {
		Name:          "Light",
		WindowBg:      walk.RGB(255, 255, 255),
		SurfaceBg:     walk.RGB(249, 249, 249),
		CardBg:        walk.RGB(240, 240, 240),
		TextPrimary:   walk.RGB(20, 20, 20),
		TextSecondary: walk.RGB(100, 100, 100),
		TextTitle:     walk.RGB(0, 0, 0),
		StatusSuccess: walk.RGB(22, 128, 57),
		StatusWarning: walk.RGB(180, 130, 0),
		StatusError:   walk.RGB(200, 30, 30),
		AccentColor:   walk.RGB(0, 120, 212),
	},
	"dark": {
		Name:                "Dark",
		WindowBg:            walk.RGB(32, 32, 32),
		SurfaceBg:           walk.RGB(40, 40, 40),
		CardBg:              walk.RGB(55, 55, 55),
		TextPrimary:         walk.RGB(230, 230, 230),
		TextSecondary:       walk.RGB(160, 160, 160),
		TextTitle:           walk.RGB(255, 255, 255),
		StatusSuccess:       walk.RGB(80, 200, 120),
		StatusWarning:       walk.RGB(230, 180, 50),
		StatusError:         walk.RGB(240, 80, 80),
		AccentColor:         walk.RGB(96, 160, 255),
		UseDarkWindowChrome: true,
	},
	"cappuccino": {
		Name:          "Cappuccino",
		WindowBg:      walk.RGB(245, 235, 220),
		SurfaceBg:     walk.RGB(238, 225, 208),
		CardBg:        walk.RGB(225, 210, 190),
		TextPrimary:   walk.RGB(60, 40, 25),
		TextSecondary: walk.RGB(120, 90, 65),
		TextTitle:     walk.RGB(45, 30, 15),
		StatusSuccess: walk.RGB(50, 120, 60),
		StatusWarning: walk.RGB(170, 120, 20),
		StatusError:   walk.RGB(180, 50, 40),
		AccentColor:   walk.RGB(140, 80, 30),
	},
}

// themeOrder defines the display order in the ComboBox.
var themeOrder = []string{"light", "dark", "cappuccino"}

var currentTheme *Theme

// CurrentTheme returns the active theme.
func CurrentTheme() *Theme {
	if currentTheme == nil {
		currentTheme = themes["light"]
	}
	return currentTheme
}

// initTheme loads the theme from settings or detects Windows dark mode.
func initTheme() {
	settings := config.GetSettings()
	id := settings.Theme
	if id == "" {
		if isWindowsDarkMode() {
			id = "dark"
		} else {
			id = "light"
		}
	}
	if t, ok := themes[id]; ok {
		currentTheme = t
	} else {
		currentTheme = themes["light"]
	}
}

// isWindowsDarkMode checks the Windows registry for dark mode preference.
func isWindowsDarkMode() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false
	}
	return val == 0
}

// applyWindowTheme sets the window background color and dark titlebar.
func applyWindowTheme(mw *walk.MainWindow) {
	t := CurrentTheme()

	bg, _ := walk.NewSolidColorBrush(t.WindowBg)
	mw.SetBackground(bg)

	if t.UseDarkWindowChrome {
		setDarkTitlebar(mw.Handle(), true)
	} else {
		setDarkTitlebar(mw.Handle(), false)
	}
}

// setDarkTitlebar uses DwmSetWindowAttribute to toggle the dark titlebar.
// Attribute 20 (DWMWA_USE_IMMERSIVE_DARK_MODE) — Windows 10 20H1+.
func setDarkTitlebar(hwnd win.HWND, dark bool) {
	if libdwmapi.Load() != nil {
		return
	}

	var val int32
	if dark {
		val = 1
	}

	// DWMWA_USE_IMMERSIVE_DARK_MODE = 20
	syscall.Syscall6(procDwmSetWindowAttribute.Addr(), 4,
		uintptr(hwnd),
		uintptr(20),
		uintptr(unsafe.Pointer(&val)),
		uintptr(unsafe.Sizeof(val)),
		0, 0,
	)
}

// applyThemeToPage sets background and text colors on a page and its children.
func applyThemeToPage(page *walk.TabPage) {
	t := CurrentTheme()
	bg, _ := walk.NewSolidColorBrush(t.SurfaceBg)
	page.SetBackground(bg)
	applyThemeToChildren(page, t)
}

func applyThemeToChildren(container walk.Container, t *Theme) {
	children := container.Children()
	for i := 0; i < children.Len(); i++ {
		w := children.At(i)

		switch v := w.(type) {
		case *walk.Label:
			v.SetTextColor(t.TextPrimary)
		case *walk.TextLabel:
			v.SetTextColor(t.TextPrimary)
		case *walk.LineEdit:
			setBg, _ := walk.NewSolidColorBrush(t.CardBg)
			v.SetBackground(setBg)
			v.SetTextColor(t.TextPrimary)
		case *walk.ScrollView:
			svBg, _ := walk.NewSolidColorBrush(t.SurfaceBg)
			v.SetBackground(svBg)
			applyThemeToChildren(v, t)
		case *walk.Composite:
			compBg, _ := walk.NewSolidColorBrush(t.SurfaceBg)
			v.SetBackground(compBg)
			applyThemeToChildren(v, t)
		}
	}
}

// SetThemeByID switches the active theme by ID.
func SetThemeByID(id string) {
	if t, ok := themes[id]; ok {
		currentTheme = t
	}
}

// GetThemeNames returns display names in order for the ComboBox.
func GetThemeNames() []string {
	names := make([]string, len(themeOrder))
	for i, id := range themeOrder {
		names[i] = themes[id].Name
	}
	return names
}

// GetThemeIndex returns the ComboBox index for a theme ID (-1 if unknown).
func GetThemeIndex(id string) int {
	for i, tid := range themeOrder {
		if tid == id {
			return i
		}
	}
	return 0
}

// GetThemeIDByIndex returns the theme ID for a ComboBox index.
func GetThemeIDByIndex(idx int) string {
	if idx >= 0 && idx < len(themeOrder) {
		return themeOrder[idx]
	}
	return "light"
}
