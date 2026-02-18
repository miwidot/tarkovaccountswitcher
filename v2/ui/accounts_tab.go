package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
)

var globalStatusLabel *widget.Label

// CreateAccountsTab creates the accounts list tab
func CreateAccountsTab() fyne.CanvasObject {
	accs, err := accounts.GetAccounts()
	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}

	// Status label
	globalStatusLabel = widget.NewLabel("")
	globalStatusLabel.Wrapping = fyne.TextWrapWord
	globalStatusLabel.Alignment = fyne.TextAlignCenter

	if len(accs) == 0 {
		// Empty state
		emptyIcon := canvas.NewText("📭", color.White)
		emptyIcon.TextSize = 48
		emptyIcon.Alignment = fyne.TextAlignCenter

		emptyTitle := widget.NewLabelWithStyle(
			i18n.T(i18n.EmptyStateTitle),
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		)
		emptySubtitle := widget.NewLabel(i18n.T(i18n.EmptyStateSubtitle))
		emptySubtitle.Alignment = fyne.TextAlignCenter

		return container.NewVBox(
			globalStatusLabel,
			layout.NewSpacer(),
			container.NewCenter(emptyIcon),
			container.NewCenter(emptyTitle),
			container.NewCenter(emptySubtitle),
			layout.NewSpacer(),
		)
	}

	// Account cards
	var cards []fyne.CanvasObject
	for _, acc := range accs {
		cards = append(cards, createAccountCard(acc))
	}

	list := container.NewVBox(cards...)
	scroll := container.NewVScroll(list)

	return container.NewBorder(
		container.NewVBox(globalStatusLabel, widget.NewSeparator()),
		nil, nil, nil,
		scroll,
	)
}

func createAccountCard(acc accounts.Account) fyne.CanvasObject {
	// Avatar
	initial := string([]rune(acc.Name)[0])
	avatarText := canvas.NewText(initial, color.White)
	avatarText.TextSize = 20
	avatarText.TextStyle = fyne.TextStyle{Bold: true}

	avatarBg := canvas.NewCircle(ColorPrimary)
	avatar := container.NewStack(
		container.NewCenter(container.NewPadded(container.NewPadded(avatarBg))),
		container.NewCenter(avatarText),
	)
	avatar = container.NewPadded(avatar)

	// Info
	name := widget.NewLabelWithStyle(acc.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	email := widget.NewLabel(config.MaskEmail(acc.Email))
	email.TextStyle = fyne.TextStyle{Italic: true}

	var status *widget.Label
	if acc.HasSession() {
		status = widget.NewLabel("✅ " + i18n.T(i18n.StatusAutoLogin))
	} else {
		status = widget.NewLabel("⚠️ " + i18n.T(i18n.StatusLoginReq))
	}

	info := container.NewVBox(name, email, status)

	// Buttons - compact
	switchBtn := widget.NewButton("▶", func() {
		onSwitch(acc)
	})
	switchBtn.Importance = widget.HighImportance

	deleteBtn := widget.NewButton("X", func() {
		onDelete(acc)
	})
	deleteBtn.Importance = widget.DangerImportance

	buttons := container.NewVBox(switchBtn, deleteBtn)

	// Card layout
	left := container.NewHBox(avatar, info)
	content := container.NewBorder(nil, nil, left, buttons, nil)

	// Card with background
	bg := canvas.NewRectangle(ColorSurface)
	bg.CornerRadius = 8

	card := container.NewStack(bg, container.NewPadded(content))
	return container.NewPadded(card)
}

func onSwitch(acc accounts.Account) {
	if globalStatusLabel != nil {
		globalStatusLabel.SetText("⏳ " + i18n.T(i18n.StatusLauncherRestart))
	}

	go func() {
		result := accounts.SwitchAccount(acc.ID)
		if result.Success {
			if result.HasSession {
				globalStatusLabel.SetText(i18n.TF(i18n.StatusAutoLoginActive, map[string]string{"name": result.AccountName}))
			} else {
				globalStatusLabel.SetText(i18n.TF(i18n.StatusManualLogin, map[string]string{"name": result.AccountName, "email": result.Email}))
			}
		} else {
			globalStatusLabel.SetText("❌ " + result.Error)
		}
	}()
}

func onDelete(acc accounts.Account) {
	dialog.ShowConfirm("Delete", i18n.T(i18n.ConfirmDelete), func(ok bool) {
		if ok {
			accounts.DeleteAccount(acc.ID)
			globalStatusLabel.SetText("✓ " + i18n.T(i18n.StatusAccountDeleted))
			RefreshAccountsTab()
		}
	}, MainWindow)
}

func SetStatus(msg string) {
	if globalStatusLabel != nil {
		globalStatusLabel.SetText(msg)
	}
}
