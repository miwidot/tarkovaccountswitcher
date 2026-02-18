package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
)

//go:embed icon.png
var appIconData []byte

var (
	App        fyne.App
	MainWindow fyne.Window
	mainTabs   *container.AppTabs
)

// CreateApp creates the Fyne application
func CreateApp() fyne.App {
	App = app.NewWithID("com.tarkov.accountswitcher")
	App.Settings().SetTheme(&ModernDarkTheme{})
	App.SetIcon(fyne.NewStaticResource("icon.png", appIconData))
	return App
}

// CreateMainWindow creates the main application window
func CreateMainWindow() fyne.Window {
	MainWindow = App.NewWindow("Tarkov Account Switcher v2")
	MainWindow.Resize(fyne.NewSize(800, 600))
	MainWindow.CenterOnScreen()
	MainWindow.SetIcon(fyne.NewStaticResource("icon.png", appIconData))

	// Load settings and set language
	settings := config.GetSettings()
	if settings.Language != "" {
		i18n.SetLanguage(settings.Language)
	} else {
		i18n.SetLanguage(config.GetSystemLanguage())
	}

	// Create tabs
	mainTabs = container.NewAppTabs(
		container.NewTabItem("📋 Accounts", CreateAccountsTab()),
		container.NewTabItem("➕ Add", CreateAddTab()),
		container.NewTabItem("⚙️ Settings", CreateSettingsTab()),
	)
	mainTabs.SetTabLocation(container.TabLocationTop)

	MainWindow.SetContent(mainTabs)

	// Handle close to minimize to tray
	MainWindow.SetCloseIntercept(func() {
		MainWindow.Hide()
	})

	// Setup language change callback
	i18n.LanguageChangedCallback = func() {
		RefreshUI()
	}

	return MainWindow
}

// RefreshUI refreshes the UI after language change
func RefreshUI() {
	if mainTabs == nil {
		return
	}

	// Update tab titles
	mainTabs.Items[0].Text = "📋 " + i18n.T(i18n.TabAccounts)[2:]
	mainTabs.Items[1].Text = "➕ " + i18n.T(i18n.TabAdd)[2:]
	mainTabs.Items[2].Text = "⚙️ " + i18n.T(i18n.TabSettings)[2:]

	// Recreate tab content
	mainTabs.Items[0].Content = CreateAccountsTab()
	mainTabs.Items[1].Content = CreateAddTab()
	mainTabs.Items[2].Content = CreateSettingsTab()

	mainTabs.Refresh()
}

// RefreshAccountsTab refreshes only the accounts tab
func RefreshAccountsTab() {
	if mainTabs == nil || len(mainTabs.Items) < 1 {
		return
	}

	mainTabs.Items[0].Content = CreateAccountsTab()
	mainTabs.Refresh()
}

// ShowWindow shows the main window
func ShowWindow() {
	if MainWindow != nil {
		MainWindow.Show()
		MainWindow.RequestFocus()
	}
}

// HideWindow hides the main window
func HideWindow() {
	if MainWindow != nil {
		MainWindow.Hide()
	}
}

// SelectTab selects a tab by index
func SelectTab(index int) {
	if mainTabs != nil && index >= 0 && index < len(mainTabs.Items) {
		mainTabs.SelectIndex(index)
	}
}
