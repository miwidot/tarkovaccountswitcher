package main

import (
	"os"

	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/launcher"
	"tarkov-account-switcher/internal/singleinstance"
	"tarkov-account-switcher/ui"
)

func main() {
	// Single instance check - exit if already running
	if !singleinstance.Lock("TarkovAccountSwitcher_v2") {
		os.Exit(0)
	}

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

	// Run the Walk UI (blocks until app exits)
	if err := ui.Run(); err != nil {
		panic(err)
	}
}
