package ui

import (
	"os/exec"
	"strings"

	"github.com/lxn/walk"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
	"tarkov-account-switcher/internal/updater"
)

var settingsStatusLbl *walk.Label

func buildSettingsTab(page *walk.TabPage) {
	rebuildSettingsTab(page)
}

func rebuildSettingsTab(page *walk.TabPage) {
	clearTabPage(page)
	if page.Layout() == nil {
		page.SetLayout(walk.NewVBoxLayout())
	}
	settingsStatusLbl = nil

	settings := config.GetSettings()

	// Title
	title, _ := walk.NewLabel(page)
	title.SetAlignment(walk.AlignHCenterVNear)
	title.SetText(i18n.T(i18n.SettingsTitle))
	title.SetFont(fontTitle)

	// Separator
	newSeparator(page)

	// --- Language ---
	langLbl, _ := walk.NewLabel(page)
	langLbl.SetText(i18n.T(i18n.LabelLanguage))
	langLbl.SetFont(fontBold)

	langCombo, _ := walk.NewComboBox(page)
	langCombo.SetFont(fontNormal)
	langItems := []string{"English", "Deutsch"}
	langCombo.SetModel(langItems)
	if i18n.GetLanguage() == "de" {
		langCombo.SetCurrentIndex(1)
	} else {
		langCombo.SetCurrentIndex(0)
	}
	langCombo.CurrentIndexChanged().Attach(func() {
		lang := "en"
		if langCombo.CurrentIndex() == 1 {
			lang = "de"
		}
		config.SetLanguage(lang)
		i18n.SetLanguage(lang)
		// RefreshUI will be called by i18n.LanguageChangedCallback
	})

	// Separator
	newSeparator(page)

	// --- Theme ---
	themeLbl, _ := walk.NewLabel(page)
	themeLbl.SetText(i18n.T(i18n.LabelTheme))
	themeLbl.SetFont(fontBold)

	themeCombo, _ := walk.NewComboBox(page)
	themeCombo.SetFont(fontNormal)
	themeCombo.SetModel(GetThemeNames())
	themeCombo.SetCurrentIndex(GetThemeIndex(settings.Theme))
	themeCombo.CurrentIndexChanged().Attach(func() {
		id := GetThemeIDByIndex(themeCombo.CurrentIndex())
		config.SetTheme(id)
		SetThemeByID(id)
		applyWindowTheme(mainWindow)
		RefreshUI()
	})

	// Separator
	newSeparator(page)

	// --- Launcher Path ---
	pathLbl, _ := walk.NewLabel(page)
	pathLbl.SetText(i18n.T(i18n.LabelLauncherPath))
	pathLbl.SetFont(fontBold)

	// Path row: LineEdit + Browse button
	pathComp, _ := walk.NewComposite(page)
	pathComp.SetLayout(walk.NewHBoxLayout())

	pathEdit, _ := walk.NewLineEdit(pathComp)
	pathEdit.SetText(settings.LauncherPath)
	pathEdit.SetCueBanner(i18n.T(i18n.PlaceholderLauncher))
	pathEdit.SetFont(fontNormal)

	browseBtn, _ := walk.NewPushButton(pathComp)
	browseBtn.SetText(i18n.T(i18n.BtnBrowse))
	browseBtn.SetFont(fontNormal)
	browseBtn.Clicked().Attach(func() {
		dlg := new(walk.FileDialog)
		dlg.Title = i18n.T(i18n.BtnBrowse)
		dlg.Filter = "Executable (*.exe)|*.exe"
		if ok, _ := dlg.ShowOpen(mainWindow); ok {
			pathEdit.SetText(dlg.FilePath)
		}
	})

	// Save button row
	saveComp, _ := walk.NewComposite(page)
	saveComp.SetLayout(walk.NewHBoxLayout())

	saveBtn, _ := walk.NewPushButton(saveComp)
	saveBtn.SetText(i18n.T(i18n.BtnSave))
	saveBtn.SetFont(fontBold)
	saveBtn.Clicked().Attach(func() {
		path := strings.TrimSpace(pathEdit.Text())
		if path == "" {
			if settingsStatusLbl != nil {
				settingsStatusLbl.SetText("⚠ " + i18n.T(i18n.StatusEnterPath))
			}
			return
		}
		config.SetLauncherPath(path)
		if settingsStatusLbl != nil {
			settingsStatusLbl.SetText("✓ " + i18n.T(i18n.StatusPathSaved))
		}
	})

	walk.NewHSpacer(saveComp)

	// Separator
	newSeparator(page)

	// --- Streamer Mode ---
	streamerCheck, _ := walk.NewCheckBox(page)
	streamerCheck.SetText(i18n.T(i18n.LabelStreamerMode))
	streamerCheck.SetFont(fontNormal)
	streamerCheck.SetChecked(settings.StreamerMode)
	streamerCheck.CheckedChanged().Attach(func() {
		config.SetStreamerMode(streamerCheck.Checked())
		// Refresh accounts tab to show/hide emails
		if tabWidget != nil {
			pages := tabWidget.Pages()
			if pages.Len() > 0 {
				rebuildAccountsTab(pages.At(0))
			}
		}
		if settingsStatusLbl != nil {
			mode := "OFF"
			if streamerCheck.Checked() {
				mode = "ON"
			}
			settingsStatusLbl.SetText("✓ Streamer Mode " + mode)
		}
	})

	helpLbl, _ := walk.NewLabel(page)
	helpLbl.SetText(i18n.T(i18n.StreamerModeHelp))
	helpLbl.SetFont(fontSmall)

	// Separator
	newSeparator(page)

	// Status label
	statusLbl, _ := walk.NewLabel(page)
	statusLbl.SetAlignment(walk.AlignHCenterVNear)
	statusLbl.SetFont(fontNormal)
	settingsStatusLbl = statusLbl

	// Spacer
	walk.NewVSpacer(page)

	// Version
	versionLbl, _ := walk.NewLabel(page)
	versionLbl.SetAlignment(walk.AlignHCenterVNear)
	versionLbl.SetText("Tarkov Account Switcher " + updater.CurrentVersion)
	versionLbl.SetFont(fontSmall)

	// Link
	linkComp, _ := walk.NewComposite(page)
	linkComp.SetLayout(walk.NewHBoxLayout())

	walk.NewHSpacer(linkComp)

	linkLbl, _ := walk.NewLinkLabel(linkComp)
	linkLbl.SetText(`Powered by <a href="https://tarkov-stammtisch.de">Tarkov-Stammtisch.de</a>`)
	linkLbl.SetFont(fontSmall)
	linkLbl.LinkActivated().Attach(func(link *walk.LinkLabelLink) {
		exec.Command("rundll32", "url.dll,FileProtocolHandler", link.URL()).Start()
	})

	walk.NewHSpacer(linkComp)

	applyThemeToPage(page)
}
