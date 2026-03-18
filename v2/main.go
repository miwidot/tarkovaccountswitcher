package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/config"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Ensure data directory exists
	if err := config.EnsureDataDir(); err != nil {
		panic(err)
	}

	// Initialize encryption key
	if _, err := accounts.GetOrCreateKey(); err != nil {
		panic(err)
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "Tarkov Account Switcher",
		Width:             800,
		Height:            600,
		MinWidth:          600,
		MinHeight:         400,
		DisableResize:     false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: true,
		BackgroundColour:  &options.RGBA{R: 26, G: 26, B: 26, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnDomReady: app.domReady,
		OnShutdown: app.shutdown,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "TarkovAccountSwitcher-v2-a3f8e921",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				runtime.WindowUnminimise(app.ctx)
				runtime.WindowShow(app.ctx)
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		panic(err)
	}
}
