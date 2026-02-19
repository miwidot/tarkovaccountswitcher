package ui

import (
	"os"

	"github.com/lxn/walk"

	"tarkov-account-switcher/internal/updater"
)

var notifyIcon *walk.NotifyIcon

// setupTray creates the system tray icon with context menu.
func setupTray() error {
	var err error
	notifyIcon, err = walk.NewNotifyIcon(mainWindow)
	if err != nil {
		return err
	}

	// Set icon
	if appIcon != nil {
		notifyIcon.SetIcon(appIcon)
	}
	notifyIcon.SetToolTip("Tarkov Account Switcher " + updater.CurrentVersion)
	notifyIcon.SetVisible(true)

	// Double-click opens window
	notifyIcon.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			ShowWindow()
		}
	})

	// Context menu
	menu := notifyIcon.ContextMenu()

	openAction := walk.NewAction()
	openAction.SetText("Open / Öffnen")
	openAction.Triggered().Attach(func() {
		ShowWindow()
	})
	menu.Actions().Add(openAction)

	menu.Actions().Add(walk.NewSeparatorAction())

	quitAction := walk.NewAction()
	quitAction.SetText("Quit / Beenden")
	quitAction.Triggered().Attach(func() {
		QuitApp()
	})
	menu.Actions().Add(quitAction)

	return nil
}

// QuitApp cleans up and exits the application.
func QuitApp() {
	if notifyIcon != nil {
		notifyIcon.Dispose()
	}
	if mainWindow != nil {
		mainWindow.Close()
	}
	os.Exit(0)
}
