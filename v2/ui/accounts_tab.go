package ui

import (
	"os/exec"

	"github.com/lxn/walk"

	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
)

var (
	accountsStatusLabel *walk.Label
	updateBannerLbl     *walk.LinkLabel
	pendingUpdateText   string
)

func buildAccountsTab(page *walk.TabPage) {
	rebuildAccountsTab(page)
}

func rebuildAccountsTab(page *walk.TabPage) {
	clearTabPage(page)
	if page.Layout() == nil {
		page.SetLayout(walk.NewVBoxLayout())
	}
	accountsStatusLabel = nil
	updateBannerLbl = nil

	// Update banner (persists across rebuilds)
	if pendingUpdateText != "" {
		banner, _ := walk.NewLinkLabel(page)
		banner.SetText(pendingUpdateText)
		banner.SetFont(fontNormal)
		banner.LinkActivated().Attach(func(link *walk.LinkLabelLink) {
			exec.Command("rundll32", "url.dll,FileProtocolHandler", link.URL()).Start()
		})
		updateBannerLbl = banner
	}

	accs, err := accounts.GetAccounts()
	if err != nil {
		lbl, _ := walk.NewLabel(page)
		lbl.SetText("Error: " + err.Error())
		lbl.SetFont(fontNormal)
		return
	}

	// Status label at top
	statusLbl, _ := walk.NewLabel(page)
	statusLbl.SetAlignment(walk.AlignHCenterVNear)
	statusLbl.SetFont(fontNormal)
	accountsStatusLabel = statusLbl

	if len(accs) == 0 {
		walk.NewVSpacer(page)

		emptyTitle, _ := walk.NewLabel(page)
		emptyTitle.SetAlignment(walk.AlignHCenterVNear)
		emptyTitle.SetText(i18n.T(i18n.EmptyStateTitle))
		emptyTitle.SetFont(fontTitle)

		emptySub, _ := walk.NewLabel(page)
		emptySub.SetAlignment(walk.AlignHCenterVNear)
		emptySub.SetText(i18n.T(i18n.EmptyStateSubtitle))
		emptySub.SetFont(fontNormal)

		walk.NewVSpacer(page)
		applyThemeToPage(page)
		return
	}

	// Scrollable area
	scrollView, _ := walk.NewScrollView(page)
	vl := walk.NewVBoxLayout()
	vl.SetSpacing(8)
	vl.SetMargins(walk.Margins{HNear: 4, VNear: 4, HFar: 4, VFar: 4})
	scrollView.SetLayout(vl)

	cardBg := walk.RGB(240, 240, 240)

	for _, acc := range accs {
		createAccountCard(scrollView, acc, cardBg)
	}

	walk.NewVSpacer(scrollView)
	applyThemeToPage(page)
}

func createAccountCard(parent walk.Container, acc accounts.Account, cardBg walk.Color) {
	// Card with light gray background
	card, _ := walk.NewComposite(parent)
	hl := walk.NewHBoxLayout()
	hl.SetMargins(walk.Margins{HNear: 14, VNear: 10, HFar: 14, VFar: 10})
	hl.SetSpacing(12)
	card.SetLayout(hl)

	bg, _ := walk.NewSolidColorBrush(cardBg)
	card.SetBackground(bg)

	// Info section (left)
	infoComp, _ := walk.NewComposite(card)
	infoVL := walk.NewVBoxLayout()
	infoVL.SetMargins(walk.Margins{})
	infoVL.SetSpacing(2)
	infoComp.SetLayout(infoVL)
	infoComp.SetBackground(bg)

	nameLbl, _ := walk.NewLabel(infoComp)
	nameLbl.SetText(acc.Name)
	nameLbl.SetFont(fontBold)

	emailLbl, _ := walk.NewLabel(infoComp)
	emailLbl.SetText(config.MaskEmail(acc.Email))
	emailLbl.SetFont(fontSmall)

	statusLbl, _ := walk.NewLabel(infoComp)
	statusLbl.SetFont(fontSmall)
	if acc.HasSession() {
		statusLbl.SetText("✅ " + i18n.T(i18n.StatusAutoLogin))
	} else {
		statusLbl.SetText("⚠ " + i18n.T(i18n.StatusLoginReq))
	}

	// Spacer pushes buttons right
	walk.NewHSpacer(card)

	// Buttons (right)
	btnComp, _ := walk.NewComposite(card)
	btnVL := walk.NewVBoxLayout()
	btnVL.SetMargins(walk.Margins{})
	btnVL.SetSpacing(6)
	btnComp.SetLayout(btnVL)
	btnComp.SetBackground(bg)

	localAcc := acc

	switchBtn, _ := walk.NewPushButton(btnComp)
	switchBtn.SetText(i18n.T(i18n.BtnSwitch))
	switchBtn.SetFont(fontNormal)
	switchBtn.Clicked().Attach(func() {
		onSwitchAccount(localAcc)
	})

	deleteBtn, _ := walk.NewPushButton(btnComp)
	deleteBtn.SetText(i18n.T(i18n.BtnDelete))
	deleteBtn.SetFont(fontSmall)
	deleteBtn.Clicked().Attach(func() {
		onDeleteAccount(localAcc)
	})
}

func onSwitchAccount(acc accounts.Account) {
	if accountsStatusLabel != nil {
		accountsStatusLabel.SetText("⏳ " + i18n.T(i18n.StatusLauncherRestart))
	}

	go func() {
		result := accounts.SwitchAccount(acc.ID)
		RunOnUI(func() {
			if accountsStatusLabel == nil {
				return
			}
			if result.Success {
				if result.HasSession {
					accountsStatusLabel.SetText(i18n.TF(i18n.StatusAutoLoginActive, map[string]string{"name": result.AccountName}))
				} else {
					accountsStatusLabel.SetText(i18n.TF(i18n.StatusManualLogin, map[string]string{"name": result.AccountName, "email": result.Email}))
				}
			} else {
				accountsStatusLabel.SetText("❌ " + result.Error)
			}
		})
	}()
}

func onDeleteAccount(acc accounts.Account) {
	if mainWindow == nil {
		return
	}
	ret := walk.MsgBox(mainWindow, i18n.T(i18n.BtnDelete), i18n.T(i18n.ConfirmDelete),
		walk.MsgBoxYesNo|walk.MsgBoxIconQuestion)
	if ret == walk.DlgCmdYes {
		accounts.DeleteAccount(acc.ID)
		if accountsStatusLabel != nil {
			accountsStatusLabel.SetText("✓ " + i18n.T(i18n.StatusAccountDeleted))
		}
		if tabWidget != nil {
			pages := tabWidget.Pages()
			if pages.Len() > 0 {
				rebuildAccountsTab(pages.At(0))
			}
		}
	}
}

func SetStatus(msg string) {
	RunOnUI(func() {
		if accountsStatusLabel != nil {
			accountsStatusLabel.SetText(msg)
		}
	})
}

// ShowUpdateBanner sets the update banner text and displays it.
func ShowUpdateBanner(text string) {
	pendingUpdateText = text
	if updateBannerLbl != nil {
		updateBannerLbl.SetText(text)
	} else {
		// Banner doesn't exist yet — rebuild to show it
		if tabWidget != nil {
			pages := tabWidget.Pages()
			if pages.Len() > 0 {
				rebuildAccountsTab(pages.At(0))
			}
		}
	}
}
