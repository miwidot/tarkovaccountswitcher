package main

import (
	"context"
	_ "embed"
	"fmt"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"tarkov-account-switcher/internal/accounts"
	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
	"tarkov-account-switcher/internal/launcher"
	"tarkov-account-switcher/internal/updater"
)

//go:embed assets/icon.ico
var trayIconData []byte

// App struct — all public methods are bound to the frontend
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load settings and set language. Accept canonical tags directly; map known
	// legacy short codes (e.g. "de"/"en") so an existing user's choice survives;
	// otherwise fall back to the system UI language.
	settings := config.GetSettings()
	lang := settings.Language
	if !i18n.IsSupportedLocale(lang) {
		if mapped := i18n.NormalizeLegacyLocale(lang); mapped != "" {
			lang = mapped
		}
	}
	if i18n.IsSupportedLocale(lang) {
		i18n.SetLanguage(lang)
		// Persist the normalized tag so the migration only happens once.
		if lang != settings.Language {
			if err := config.SetLanguage(lang); err != nil {
				fmt.Printf("i18n: failed to persist migrated locale: %v\n", err)
			}
		}
	} else {
		i18n.SetLanguage(config.GetSystemLanguage())
		if settings.Language != "" {
			if err := config.SetLanguage(i18n.GetLanguage()); err != nil {
				fmt.Printf("i18n: failed to persist fallback locale: %v\n", err)
			}
		}
	}

	// Session captured callback -> emit event to frontend
	accounts.SessionCapturedCallback = func(accountID string) {
		wailsRuntime.EventsEmit(a.ctx, "session-captured", accountID)
	}

	// Launcher started callback -> hide window
	launcher.OnLauncherStarted = func() {
		wailsRuntime.WindowHide(a.ctx)
	}
}

func (a *App) domReady(ctx context.Context) {
	// Set window icon (Win32 API — reliable regardless of build method)
	setWindowIcon(trayIconData)

	// Start system tray
	a.setupSystemTray()

	// Background update check
	updater.CheckAsync(func(result updater.Result) {
		wailsRuntime.EventsEmit(a.ctx, "update-available", map[string]any{
			"stable": result.StableUpdate,
			"beta":   result.BetaUpdate,
		})
	})
}

func (a *App) shutdown(ctx context.Context) {
	stopTray()
}

// ==================== SYSTEM TRAY ====================

func (a *App) setupSystemTray() {
	tooltip := i18n.TF(i18n.TrayTooltip, map[string]string{"version": updater.CurrentVersion})

	onShow := func() {
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
		setWindowIcon(trayIconData)
	}

	onQuit := func() {
		stopTray()
		wailsRuntime.Quit(a.ctx)
	}

	startTray(trayIconData, tooltip, onShow, onQuit)
}

// ==================== ACCOUNTS ====================

// AccountDTO is sent to the frontend (no raw session data exposed)
type AccountDTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	HasSession      bool   `json:"hasSession"`
	SessionCaptured string `json:"sessionCaptured"`
}

func toDTO(acc accounts.Account) AccountDTO {
	return AccountDTO{
		ID:              acc.ID,
		Name:            acc.Name,
		Email:           config.MaskEmail(acc.Email),
		HasSession:      acc.HasSession(),
		SessionCaptured: acc.SessionCaptured,
	}
}

// GetAccounts returns all accounts as DTOs
func (a *App) GetAccounts() ([]AccountDTO, error) {
	accs, err := accounts.ListAccounts()
	if err != nil {
		return nil, err
	}
	a.resizeWindowToFit(len(accs))
	dtos := make([]AccountDTO, len(accs))
	for i, acc := range accs {
		dtos[i] = toDTO(acc)
	}
	return dtos, nil
}

// resizeWindowToFit sizes the window to comfortably fit the given number of
// account cards. Called on every accounts list fetch; the window grows or
// shrinks with the list. Uses a native Win32 SetWindowPos helper so the
// user's manual width stays intact (Wails' WindowSetSize is not DPI-stable
// on Windows when the width is round-tripped back in).
func (a *App) resizeWindowToFit(count int) {
	const (
		chrome  = 180 // title bar + tab bar + panel padding + update banner room
		perCard = 100 // card body + 6px gap — generous to avoid any overflow
		safety  = 40
		floor   = 380
	)
	needed := floor
	if count > 0 {
		needed = chrome + perCard*count + safety
		if needed < floor {
			needed = floor
		}
	}
	if screens, err := wailsRuntime.ScreenGetAll(a.ctx); err == nil {
		for _, s := range screens {
			if !s.IsCurrent {
				continue
			}
			if max := s.Size.Height - 80; needed > max {
				needed = max
			}
			break
		}
	}
	setWindowHeight(needed)
}

// SwitchResultDTO is the result of a switch operation
type SwitchResultDTO struct {
	Success     bool   `json:"success"`
	AccountName string `json:"accountName"`
	Email       string `json:"email"`
	HasSession  bool   `json:"hasSession"`
	Message     string `json:"message"`
	Error       string `json:"error"`
}

// SwitchAccount switches to the given account
func (a *App) SwitchAccount(id string) SwitchResultDTO {
	result := accounts.SwitchAccount(id)
	return SwitchResultDTO{
		Success:     result.Success,
		AccountName: result.AccountName,
		Email:       config.MaskEmail(result.Email),
		HasSession:  result.HasSession,
		Message:     result.Message,
		Error:       result.Error,
	}
}

// AddAccount adds a new account and starts the login flow
func (a *App) AddAccount(name, email string) error {
	_, err := accounts.AddAccount(name, email)
	return err
}

// DeleteAccount removes an account by ID
func (a *App) DeleteAccount(id string) error {
	return accounts.DeleteAccount(id)
}

// ReorderAccounts saves accounts in the order given by orderedIDs.
func (a *App) ReorderAccounts(orderedIDs []string) error {
	return accounts.ReorderAccounts(orderedIDs)
}

// ==================== SETTINGS ====================

// SettingsDTO for frontend consumption
type SettingsDTO struct {
	LauncherPath string `json:"launcherPath"`
	Language     string `json:"language"`
	StreamerMode bool   `json:"streamerMode"`
	Theme        string `json:"theme"`
	AutoStart    bool   `json:"autoStart"`
}

// GetSettings returns current settings
func (a *App) GetSettings() SettingsDTO {
	s := config.GetSettings()
	return SettingsDTO{
		LauncherPath: s.LauncherPath,
		Language:     i18n.GetLanguage(),
		StreamerMode: s.StreamerMode,
		Theme:        s.Theme,
		AutoStart:    s.AutoStart,
	}
}

// SetLanguage changes the language
func (a *App) SetLanguage(lang string) error {
	if !i18n.IsSupportedLocale(lang) {
		return fmt.Errorf("%s", i18n.TF(i18n.ErrUnsupportedLanguage, map[string]string{"lang": lang}))
	}
	err := config.SetLanguage(lang)
	if err != nil {
		return err
	}
	i18n.SetLanguage(lang)
	return nil
}

// SetLauncherPath saves the launcher path
func (a *App) SetLauncherPath(path string) error {
	return config.SetLauncherPath(path)
}

// SetStreamerMode toggles streamer mode
func (a *App) SetStreamerMode(enabled bool) error {
	return config.SetStreamerMode(enabled)
}

// SetAutoStart enables or disables Windows autostart
func (a *App) SetAutoStart(enabled bool) error {
	if err := config.ApplyAutoStart(enabled); err != nil {
		return err
	}
	return config.SetAutoStart(enabled)
}

// SetTheme saves the theme preference
func (a *App) SetTheme(id string) error {
	return config.SetTheme(id)
}

// BrowseLauncherPath opens a native file dialog
func (a *App) BrowseLauncherPath() (string, error) {
	var path string
	var err error
	runWithDialogIconRefresh(trayIconData, func() {
		path, err = wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
			Title: i18n.T(i18n.DlgSelectLauncher),
			Filters: []wailsRuntime.FileFilter{
				{DisplayName: i18n.T(i18n.DlgFilterExe), Pattern: "*.exe"},
			},
		})
	})
	return path, err
}

// ==================== TRANSLATIONS ====================

// GetAllTranslations returns all translations for the current language
func (a *App) GetAllTranslations() map[string]string {
	keys := []string{
		i18n.TabAccounts, i18n.TabAdd, i18n.TabSettings,
		i18n.AddAccountTitle, i18n.LabelAccountName, i18n.PlaceholderAccName,
		i18n.LabelEmail, i18n.PlaceholderEmail, i18n.AddAccountHelp, i18n.BtnAddAccount,
		i18n.EmptyStateTitle, i18n.EmptyStateSubtitle,
		i18n.StatusAutoLogin, i18n.StatusLoginReq,
		i18n.BtnSwitch, i18n.BtnDelete, i18n.ConfirmDelete, i18n.BtnYes, i18n.BtnNo,
		i18n.SettingsTitle, i18n.LabelLanguage, i18n.LabelLauncherPath,
		i18n.PlaceholderLauncher, i18n.BtnBrowse, i18n.BtnSave,
		i18n.LabelAutoStart, i18n.AutoStartHelp, i18n.BtnQuit,
		i18n.LabelStreamerMode, i18n.StreamerModeHelp,
		i18n.LabelTheme,
		i18n.ThemeNameEft, i18n.ThemeNameKilla, i18n.ThemeNameDark, i18n.ThemeNameLight, i18n.ThemeNameCappuccino,
		i18n.StatusFillFields, i18n.StatusAccountAdded, i18n.StatusAccountDeleted,
		i18n.StatusLauncherRestart, i18n.StatusAutoLoginActive, i18n.StatusManualLogin,
		i18n.StatusPathSaved, i18n.StatusEnterPath, i18n.StatusLanguageSaved, i18n.StatusThemeSaved,
		i18n.StatusError, i18n.StatusDeleteFailed,
		i18n.StatusAutoStartOn, i18n.StatusAutoStartOff, i18n.StatusStreamerOn, i18n.StatusStreamerOff,
		i18n.AboutVersion, i18n.HtmlPageTitle, i18n.LinkDownload, i18n.FooterPoweredBy,
		i18n.UpdateAvailableStable, i18n.UpdateAvailableBeta,
	}
	result := make(map[string]string, len(keys))
	for _, k := range keys {
		result[k] = i18n.T(k)
	}
	return result
}

// GetCurrentLanguage returns the current language code
func (a *App) GetCurrentLanguage() string {
	return i18n.GetLanguage()
}

// ==================== MISC ====================

// QuitApp completely closes the application
func (a *App) QuitApp() {
	stopTray()
	wailsRuntime.Quit(a.ctx)
}

// GetVersion returns the current app version
func (a *App) GetVersion() string {
	return updater.CurrentVersion
}

// ConfirmDelete shows a native confirmation dialog
func (a *App) ConfirmDelete() (bool, error) {
	return showQuestionDialog(
		trayIconData,
		i18n.T(i18n.BtnDelete),
		i18n.T(i18n.ConfirmDelete),
	)
}
