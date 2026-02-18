package ui

import (
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
)

func CreateSettingsTab() fyne.CanvasObject {
	settings := config.GetSettings()

	// Title
	title := widget.NewLabelWithStyle(
		i18n.T(i18n.SettingsTitle),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Status
	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Language
	langSelect := widget.NewSelect([]string{"English", "Deutsch"}, func(val string) {
		lang := "en"
		if val == "Deutsch" {
			lang = "de"
		}
		config.SetLanguage(lang)
		i18n.SetLanguage(lang)
		statusLabel.SetText("✓ " + i18n.T(i18n.StatusLanguageSaved))
	})
	if i18n.GetLanguage() == "de" {
		langSelect.SetSelected("Deutsch")
	} else {
		langSelect.SetSelected("English")
	}

	// Launcher path
	pathEntry := widget.NewEntry()
	pathEntry.SetText(settings.LauncherPath)

	browseBtn := widget.NewButton(i18n.T(i18n.BtnBrowse), func() {
		fd := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err == nil && r != nil {
				pathEntry.SetText(r.URI().Path())
				r.Close()
			}
		}, MainWindow)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
		fd.Show()
	})

	saveBtn := widget.NewButton(i18n.T(i18n.BtnSave), func() {
		path := strings.TrimSpace(pathEntry.Text)
		if path == "" {
			statusLabel.SetText("⚠️ " + i18n.T(i18n.StatusEnterPath))
			return
		}
		config.SetLauncherPath(path)
		statusLabel.SetText("✓ " + i18n.T(i18n.StatusPathSaved))
	})
	saveBtn.Importance = widget.HighImportance

	pathRow := container.NewBorder(nil, nil, nil, browseBtn, pathEntry)

	// Streamer Mode
	streamerCheck := widget.NewCheck(i18n.T(i18n.LabelStreamerMode), func(checked bool) {
		config.SetStreamerMode(checked)
		RefreshAccountsTab()
		statusLabel.SetText("✓ Streamer Mode " + map[bool]string{true: "ON", false: "OFF"}[checked])
	})
	streamerCheck.Checked = settings.StreamerMode

	streamerHelp := widget.NewLabel(i18n.T(i18n.StreamerModeHelp))
	streamerHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Layout
	form := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewLabel(i18n.T(i18n.LabelLanguage)),
		langSelect,
		widget.NewSeparator(),
		widget.NewLabel(i18n.T(i18n.LabelLauncherPath)),
		pathRow,
		container.NewHBox(saveBtn),
		widget.NewSeparator(),
		streamerCheck,
		streamerHelp,
		widget.NewSeparator(),
		statusLabel,
		widget.NewLabel(""),
		widget.NewLabelWithStyle("Tarkov Account Switcher v2.0.0", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		container.NewCenter(createPoweredByLink()),
	)

	return container.NewPadded(container.NewPadded(form))
}

func createPoweredByLink() fyne.CanvasObject {
	link, _ := url.Parse("https://tarkov-stammtisch.de")
	hyperlink := widget.NewHyperlink("Powered by Tarkov-Stammtisch.de", link)
	return hyperlink
}
