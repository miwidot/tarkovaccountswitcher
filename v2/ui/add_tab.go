package ui

import (
	"strings"

	"github.com/lxn/walk"

	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/i18n"
)

var (
	addNameEdit  *walk.LineEdit
	addEmailEdit *walk.LineEdit
	addStatusLbl *walk.Label
)

func buildAddTab(page *walk.TabPage) {
	rebuildAddTab(page)
}

func rebuildAddTab(page *walk.TabPage) {
	clearTabPage(page)
	if page.Layout() == nil {
		page.SetLayout(walk.NewVBoxLayout())
	}
	addNameEdit = nil
	addEmailEdit = nil
	addStatusLbl = nil

	// Title
	title, _ := walk.NewLabel(page)
	title.SetAlignment(walk.AlignHCenterVNear)
	title.SetText(i18n.T(i18n.AddAccountTitle))
	title.SetFont(fontTitle)

	// Separator
	newSeparator(page)

	// Account Name label
	nameLbl, _ := walk.NewLabel(page)
	nameLbl.SetText(i18n.T(i18n.LabelAccountName))
	nameLbl.SetFont(fontBold)

	// Account Name entry
	nameEdit, _ := walk.NewLineEdit(page)
	nameEdit.SetCueBanner(i18n.T(i18n.PlaceholderAccName))
	nameEdit.SetFont(fontNormal)
	addNameEdit = nameEdit

	// Email label
	emailLbl, _ := walk.NewLabel(page)
	emailLbl.SetText(i18n.T(i18n.LabelEmail))
	emailLbl.SetFont(fontBold)

	// Email entry
	emailEdit, _ := walk.NewLineEdit(page)
	emailEdit.SetCueBanner(i18n.T(i18n.PlaceholderEmail))
	emailEdit.SetFont(fontNormal)
	addEmailEdit = emailEdit

	// Separator
	newSeparator(page)

	// Help text
	helpLbl, _ := walk.NewTextLabel(page)
	helpLbl.SetText(i18n.T(i18n.AddAccountHelp))
	helpLbl.SetTextAlignment(walk.AlignHCenterVNear)
	helpLbl.SetFont(fontSmall)

	// Separator
	newSeparator(page)

	// Button row (centered using composite)
	btnComp, _ := walk.NewComposite(page)
	btnComp.SetLayout(walk.NewHBoxLayout())

	walk.NewHSpacer(btnComp)

	addBtn, _ := walk.NewPushButton(btnComp)
	addBtn.SetText(i18n.T(i18n.BtnAddAccount))
	addBtn.SetFont(fontBold)
	addBtn.Clicked().Attach(func() {
		onAddAccount()
	})

	walk.NewHSpacer(btnComp)

	// Status label
	statusLbl, _ := walk.NewLabel(page)
	statusLbl.SetAlignment(walk.AlignHCenterVNear)
	statusLbl.SetFont(fontNormal)
	addStatusLbl = statusLbl

	// Bottom spacer
	walk.NewVSpacer(page)

	applyThemeToPage(page)
}

func onAddAccount() {
	if addNameEdit == nil || addEmailEdit == nil {
		return
	}

	name := strings.TrimSpace(addNameEdit.Text())
	email := strings.TrimSpace(addEmailEdit.Text())

	if name == "" || email == "" {
		if addStatusLbl != nil {
			addStatusLbl.SetText("⚠ " + i18n.T(i18n.StatusFillFields))
		}
		return
	}

	if addStatusLbl != nil {
		addStatusLbl.SetText("⏳ " + i18n.T(i18n.StatusLauncherRestart))
	}

	go func() {
		_, err := accounts.AddAccount(name, email)
		RunOnUI(func() {
			if err != nil {
				if addStatusLbl != nil {
					addStatusLbl.SetText("❌ " + err.Error())
				}
				return
			}

			if addStatusLbl != nil {
				addStatusLbl.SetText("✅ " + i18n.T(i18n.StatusAccountAdded))
			}
			if addNameEdit != nil {
				addNameEdit.SetText("")
			}
			if addEmailEdit != nil {
				addEmailEdit.SetText("")
			}

			// Refresh accounts tab and switch to it
			if tabWidget != nil {
				pages := tabWidget.Pages()
				if pages.Len() > 0 {
					rebuildAccountsTab(pages.At(0))
				}
				tabWidget.SetCurrentIndex(0)
			}
		})
	}()
}
