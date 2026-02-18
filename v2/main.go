package main

import (
	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/launcher"
	"tarkov-account-switcher/internal/singleinstance"
	"tarkov-account-switcher/ui"
)

func main() {
	// Single instance check - disabled for now
	_ = singleinstance.Lock("TarkovAccountSwitcher_v2")

	// Ensure data directory exists
	if err := config.EnsureDataDir(); err != nil {
		panic(err)
	}

	// Initialize encryption key
	if _, err := accounts.GetOrCreateKey(); err != nil {
		panic(err)
	}

	// Set up session captured callback
	accounts.SessionCapturedCallback = func(accountID string) {
		ui.RefreshAccountsTab()
	}

	// Set up launcher started callback - minimize to tray
	launcher.OnLauncherStarted = func() {
		ui.HideWindow()
	}

	// Create app and window
	ui.CreateApp()
	window := ui.CreateMainWindow()

	// Run system tray in goroutine
	go ui.SetupTray(ui.OnTrayReady, ui.OnTrayExit)

	// Show window and run
	window.ShowAndRun()
}
