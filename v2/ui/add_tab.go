package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/i18n"
)

func CreateAddTab() fyne.CanvasObject {
	// Title
	title := widget.NewLabelWithStyle(
		i18n.T(i18n.AddAccountTitle),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Form
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder(i18n.T(i18n.PlaceholderAccName))

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder(i18n.T(i18n.PlaceholderEmail))

	// Status
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	statusLabel.Alignment = fyne.TextAlignCenter

	// Help
	helpLabel := widget.NewLabel(i18n.T(i18n.AddAccountHelp))
	helpLabel.Wrapping = fyne.TextWrapWord
	helpLabel.Alignment = fyne.TextAlignCenter

	// Button
	addBtn := widget.NewButton(i18n.T(i18n.BtnAddAccount), func() {
		name := strings.TrimSpace(nameEntry.Text)
		email := strings.TrimSpace(emailEntry.Text)

		if name == "" || email == "" {
			statusLabel.SetText("⚠️ " + i18n.T(i18n.StatusFillFields))
			return
		}

		statusLabel.SetText("⏳ " + i18n.T(i18n.StatusLauncherRestart))

		go func() {
			_, err := accounts.AddAccount(name, email)
			if err != nil {
				statusLabel.SetText("❌ " + err.Error())
				return
			}

			statusLabel.SetText("✅ " + i18n.T(i18n.StatusAccountAdded))
			nameEntry.SetText("")
			emailEntry.SetText("")
			RefreshAccountsTab()
			SelectTab(0)
		}()
	})
	addBtn.Importance = widget.HighImportance

	// Form layout
	form := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewLabel(i18n.T(i18n.LabelAccountName)),
		nameEntry,
		widget.NewLabel(i18n.T(i18n.LabelEmail)),
		emailEntry,
		widget.NewSeparator(),
		helpLabel,
		widget.NewSeparator(),
		container.NewCenter(addBtn),
		statusLabel,
	)

	return container.NewPadded(container.NewPadded(form))
}
