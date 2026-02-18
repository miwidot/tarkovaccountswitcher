package ui

import (
	_ "embed"
	"os"

	"fyne.io/systray"
)

//go:embed icon.ico
var iconData []byte

var (
	quitChan chan struct{}
)

// SetupTray initializes the system tray
func SetupTray(onReady func(), onExit func()) {
	quitChan = make(chan struct{})
	systray.Run(onReady, onExit)
}

// OnTrayReady is called when the tray is ready
func OnTrayReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Tarkov Account Switcher")
	systray.SetTooltip("Tarkov Account Switcher v2.0.0")

	// Menu items
	mOpen := systray.AddMenuItem("Open / Öffnen", "Show window")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit / Beenden", "Exit application")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				ShowWindow()
			case <-mQuit.ClickedCh:
				if quitChan != nil {
					close(quitChan)
				}
				systray.Quit()
				if App != nil {
					App.Quit()
				}
				return
			}
		}
	}()
}

// OnTrayExit is called when the tray exits
func OnTrayExit() {
	// Quit the Fyne app when tray exits
	if App != nil {
		App.Quit()
	}
	// Force exit to ensure process terminates
	os.Exit(0)
}

// GetQuitChan returns the quit channel
func GetQuitChan() <-chan struct{} {
	return quitChan
}

// QuitApp signals the app to quit
func QuitApp() {
	if quitChan != nil {
		close(quitChan)
	}
	systray.Quit()
	if App != nil {
		App.Quit()
	}
}
