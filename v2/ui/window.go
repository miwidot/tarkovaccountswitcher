package ui

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
	"tarkov-account-switcher/internal/updater"
)

//go:embed icon.ico
var iconData []byte

var (
	mainWindow *walk.MainWindow
	tabWidget  *walk.TabWidget
	appIcon    *walk.Icon
	iconPath   string

	// Fonts
	fontNormal *walk.Font
	fontBold   *walk.Font
	fontSmall  *walk.Font
	fontTitle  *walk.Font
)

func initFonts() {
	fontNormal, _ = walk.NewFont("Segoe UI", 10, 0)
	fontBold, _ = walk.NewFont("Segoe UI", 10, walk.FontBold)
	fontSmall, _ = walk.NewFont("Segoe UI", 9, 0)
	fontTitle, _ = walk.NewFont("Segoe UI", 13, walk.FontBold)
}

// Run creates and runs the main Walk application. Blocks until quit.
func Run() error {
	// Write embedded icon to temp file for Walk
	tmpDir := filepath.Join(os.TempDir(), "TarkovAccountSwitcher")
	os.MkdirAll(tmpDir, 0755)
	iconPath = filepath.Join(tmpDir, "icon.ico")
	os.WriteFile(iconPath, iconData, 0644)

	var err error
	appIcon, err = walk.NewIconFromFile(iconPath)
	if err != nil {
		appIcon = nil
	}

	// Load settings and set language
	settings := config.GetSettings()
	if settings.Language != "" {
		i18n.SetLanguage(settings.Language)
	} else {
		i18n.SetLanguage(config.GetSystemLanguage())
	}

	// Init fonts and theme
	initFonts()
	initTheme()

	// Setup language change callback
	i18n.LanguageChangedCallback = func() {
		RunOnUI(func() {
			RefreshUI()
		})
	}

	var accountsPage, addPage, settingsPage *walk.TabPage

	err = MainWindow{
		AssignTo: &mainWindow,
		Title:    "Tarkov Account Switcher v2",
		MinSize:  Size{Width: 600, Height: 400},
		Size:     Size{Width: 800, Height: 600},
		Font:     Font{Family: "Segoe UI", PointSize: 10},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			TabWidget{
				AssignTo: &tabWidget,
				Pages: []TabPage{
					{
						AssignTo: &accountsPage,
						Title:    i18n.T(i18n.TabAccounts),
						Layout:   VBox{Margins: Margins{Left: 10, Top: 8, Right: 10, Bottom: 8}},
					},
					{
						AssignTo: &addPage,
						Title:    i18n.T(i18n.TabAdd),
						Layout:   VBox{Margins: Margins{Left: 20, Top: 10, Right: 20, Bottom: 10}},
					},
					{
						AssignTo: &settingsPage,
						Title:    i18n.T(i18n.TabSettings),
						Layout:   VBox{Margins: Margins{Left: 20, Top: 10, Right: 20, Bottom: 10}},
					},
				},
			},
		},
	}.Create()
	if err != nil {
		return err
	}

	// Set icon on window
	if appIcon != nil {
		mainWindow.SetIcon(appIcon)
	}

	// Apply dark/light theme
	applyWindowTheme(mainWindow)

	// Close intercept → minimize to tray
	mainWindow.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		*canceled = true
		mainWindow.SetVisible(false)
	})

	// Build tab contents
	buildAccountsTab(accountsPage)
	buildAddTab(addPage)
	buildSettingsTab(settingsPage)

	// Setup tray
	if err := setupTray(); err != nil {
		return err
	}

	// Background update check
	updater.CheckAsync(func(result updater.Result) {
		RunOnUI(func() {
			var lines []string
			if result.StableUpdate != nil {
				lines = append(lines, i18n.TF(i18n.UpdateAvailableStable, map[string]string{
					"version": result.StableUpdate.Version,
					"url":     `<a href="` + result.StableUpdate.ReleaseURL + `">Download</a>`,
				}))
			}
			if result.BetaUpdate != nil {
				lines = append(lines, i18n.TF(i18n.UpdateAvailableBeta, map[string]string{
					"version": result.BetaUpdate.Version,
					"url":     `<a href="` + result.BetaUpdate.ReleaseURL + `">Download</a>`,
				}))
			}
			if len(lines) > 0 {
				ShowUpdateBanner(strings.Join(lines, "\n"))
			}
		})
	})

	// Center on screen
	screenW := int(win.GetSystemMetrics(win.SM_CXSCREEN))
	screenH := int(win.GetSystemMetrics(win.SM_CYSCREEN))
	mainWindow.SetX((screenW - mainWindow.Width()) / 2)
	mainWindow.SetY((screenH - mainWindow.Height()) / 2)

	mainWindow.Run()
	return nil
}

// RunOnUI executes fn on the UI thread. Safe to call from any goroutine.
func RunOnUI(fn func()) {
	if mainWindow != nil {
		mainWindow.Synchronize(fn)
	}
}

// RefreshUI refreshes all tabs after a language change.
func RefreshUI() {
	if tabWidget == nil {
		return
	}

	pages := tabWidget.Pages()
	if pages.Len() < 3 {
		return
	}

	pages.At(0).SetTitle(i18n.T(i18n.TabAccounts))
	pages.At(1).SetTitle(i18n.T(i18n.TabAdd))
	pages.At(2).SetTitle(i18n.T(i18n.TabSettings))

	rebuildAccountsTab(pages.At(0))
	rebuildAddTab(pages.At(1))
	rebuildSettingsTab(pages.At(2))
}

// RefreshAccountsTab refreshes only the accounts tab.
func RefreshAccountsTab() {
	RunOnUI(func() {
		if tabWidget == nil {
			return
		}
		pages := tabWidget.Pages()
		if pages.Len() < 1 {
			return
		}
		rebuildAccountsTab(pages.At(0))
	})
}

// ShowWindow shows and activates the main window.
func ShowWindow() {
	RunOnUI(func() {
		if mainWindow != nil {
			mainWindow.SetVisible(true)
			mainWindow.Activate()
		}
	})
}

// HideWindow hides the main window.
func HideWindow() {
	RunOnUI(func() {
		if mainWindow != nil {
			mainWindow.SetVisible(false)
		}
	})
}

// SelectTab selects a tab by index.
func SelectTab(index int) {
	RunOnUI(func() {
		if tabWidget != nil {
			tabWidget.SetCurrentIndex(index)
		}
	})
}

// newSeparator adds vertical spacing between sections.
func newSeparator(parent walk.Container) {
	comp, _ := walk.NewComposite(parent)
	vl := walk.NewVBoxLayout()
	vl.SetMargins(walk.Margins{VNear: 3, VFar: 3})
	comp.SetLayout(vl)
}

// clearTabPage removes all children from a tab page.
func clearTabPage(page *walk.TabPage) {
	children := page.Children()
	for children.Len() > 0 {
		child := children.At(0)
		child.SetParent(nil)
		child.Dispose()
	}
}
