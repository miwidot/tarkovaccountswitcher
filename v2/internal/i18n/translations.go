package i18n

import (
	"strings"
	"sync"
)

var (
	currentLanguage = "en"
	languageMutex   sync.RWMutex

	// LanguageChangedCallback is called when the language changes
	LanguageChangedCallback func()
)

// Translation keys
const (
	// Tabs
	TabAccounts = "tabAccounts"
	TabAdd      = "tabAdd"
	TabSettings = "tabSettings"

	// Add Account Tab
	AddAccountTitle      = "addAccountTitle"
	LabelAccountName     = "labelAccountName"
	PlaceholderAccName   = "placeholderAccountName"
	LabelEmail           = "labelEmail"
	PlaceholderEmail     = "placeholderEmail"
	AddAccountHelp       = "addAccountHelp"
	BtnAddAccount        = "btnAddAccount"

	// Account List
	EmptyStateTitle    = "emptyStateTitle"
	EmptyStateSubtitle = "emptyStateSubtitle"
	StatusAutoLogin    = "statusAutoLogin"
	StatusLoginReq     = "statusLoginRequired"
	BtnSwitch          = "btnSwitch"
	BtnDelete          = "btnDelete"
	ConfirmDelete      = "confirmDelete"

	// Settings Tab
	SettingsTitle        = "settingsTitle"
	LabelLanguage        = "labelLanguage"
	LabelLauncherPath    = "labelLauncherPath"
	PlaceholderLauncher  = "placeholderLauncherPath"
	BtnBrowse            = "btnBrowse"
	BtnSave              = "btnSave"
	LabelStreamerMode    = "labelStreamerMode"
	StreamerModeHelp     = "streamerModeHelp"

	// Theme
	LabelTheme = "labelTheme"

	// Switch Result Messages (used by backend)
	SwitchAutoLogin  = "switchAutoLogin"
	SwitchManualLogin = "switchManualLogin"

	// Tray Menu
	TrayOpen = "trayOpen"
	TrayQuit = "trayQuit"

	// Status Messages
	StatusFillFields     = "statusFillFields"
	StatusAccountAdded   = "statusAccountAdded"
	StatusAccountDeleted = "statusAccountDeleted"
	StatusDeleteError    = "statusDeleteError"
	StatusLauncherRestart = "statusLauncherRestarting"
	StatusAutoLoginActive = "statusAutoLoginActive"
	StatusManualLogin    = "statusManualLogin"
	StatusError          = "statusError"
	StatusPathSaved      = "statusPathSaved"
	StatusSaveError      = "statusSaveError"
	StatusEnterPath      = "statusEnterPath"
	StatusLanguageSaved  = "statusLanguageSaved"

	// Update Notifications
	UpdateAvailableStable = "updateAvailableStable"
	UpdateAvailableBeta   = "updateAvailableBeta"
)

var translations = map[string]map[string]string{
	"de": {
		// Tabs
		TabAccounts: "📋 Accounts",
		TabAdd:      "➕ Hinzufügen",
		TabSettings: "⚙️ Einstellungen",

		// Add Account Tab
		AddAccountTitle:    "Neuen Account hinzufügen",
		LabelAccountName:   "Account Name (z.B. \"Main\", \"Alt\")",
		PlaceholderAccName: "Main Account",
		LabelEmail:         "Email",
		PlaceholderEmail:   "your@email.com",
		AddAccountHelp:     "Nach dem Hinzufügen startet der Launcher automatisch.\nLogge dich ein - die Session wird automatisch gespeichert! ✅",
		BtnAddAccount:      "Account hinzufügen & Launcher starten",

		// Account List
		EmptyStateTitle:    "Noch keine Accounts gespeichert",
		EmptyStateSubtitle: "Füge oben deinen ersten Account hinzu",
		StatusAutoLogin:    "Auto-Login aktiv",
		StatusLoginReq:     "Login erforderlich - wird automatisch gespeichert",
		BtnSwitch:          "Wechseln",
		BtnDelete:          "Löschen",
		ConfirmDelete:      "Account wirklich löschen?",

		// Settings Tab
		SettingsTitle:       "Einstellungen",
		LabelLanguage:       "Sprache / Language",
		LabelLauncherPath:   "BSG Launcher Pfad",
		PlaceholderLauncher: `C:\Battlestate Games\BsgLauncher\BsgLauncher.exe`,
		BtnBrowse:           "Durchsuchen...",
		BtnSave:             "Speichern",
		LabelStreamerMode:   "Streamer Modus",
		StreamerModeHelp:    "Versteckt Email-Adressen mit ****",
		LabelTheme:          "Design / Theme",

		// Switch Result Messages
		SwitchAutoLogin:  "Launcher gestartet - Auto-Login aktiv!",
		SwitchManualLogin: "Bitte einloggen - Session wird automatisch gespeichert!",

		// Tray Menu
		TrayOpen: "Öffnen",
		TrayQuit: "Beenden",

		// Status Messages
		StatusFillFields:     "Bitte fülle alle Felder aus",
		StatusAccountAdded:   "✅ Account hinzugefügt!\n\nLauncher startet jetzt...\nBitte einloggen - Session wird automatisch gespeichert!",
		StatusAccountDeleted: "Account gelöscht",
		StatusDeleteError:    "Fehler beim Löschen",
		StatusLauncherRestart: "Launcher wird neu gestartet...",
		StatusAutoLoginActive: "🚀 AUTO-LOGIN AKTIV!\n\nAccount: {name}\nLauncher startet automatisch eingeloggt!",
		StatusManualLogin:    "⚠️ MANUELLES LOGIN\n\nAccount: {name}\nEmail: {email}\n\nBitte einloggen - Session wird automatisch gespeichert!",
		StatusError:          "Fehler: {error}",
		StatusPathSaved:      "Launcher Pfad gespeichert!",
		StatusSaveError:      "Fehler beim Speichern",
		StatusEnterPath:      "Bitte gib einen Pfad ein",
		StatusLanguageSaved:  "Sprache gespeichert!",

		// Update Notifications
		UpdateAvailableStable: "Update verfügbar: {version} — {url}",
		UpdateAvailableBeta:   "Neue Beta verfügbar: {version} — {url}",
	},
	"en": {
		// Tabs
		TabAccounts: "📋 Accounts",
		TabAdd:      "➕ Add",
		TabSettings: "⚙️ Settings",

		// Add Account Tab
		AddAccountTitle:    "Add New Account",
		LabelAccountName:   "Account Name (e.g. \"Main\", \"Alt\")",
		PlaceholderAccName: "Main Account",
		LabelEmail:         "Email",
		PlaceholderEmail:   "your@email.com",
		AddAccountHelp:     "After adding, the launcher will start automatically.\nLog in - the session will be saved automatically! ✅",
		BtnAddAccount:      "Add Account & Start Launcher",

		// Account List
		EmptyStateTitle:    "No accounts saved yet",
		EmptyStateSubtitle: "Add your first account above",
		StatusAutoLogin:    "Auto-login active",
		StatusLoginReq:     "Login required - will be saved automatically",
		BtnSwitch:          "Switch",
		BtnDelete:          "Delete",
		ConfirmDelete:      "Really delete account?",

		// Settings Tab
		SettingsTitle:       "Settings",
		LabelLanguage:       "Language / Sprache",
		LabelLauncherPath:   "BSG Launcher Path",
		PlaceholderLauncher: `C:\Battlestate Games\BsgLauncher\BsgLauncher.exe`,
		BtnBrowse:           "Browse...",
		BtnSave:             "Save",
		LabelStreamerMode:   "Streamer Mode",
		StreamerModeHelp:    "Hides email addresses with ****",
		LabelTheme:          "Theme / Design",

		// Switch Result Messages
		SwitchAutoLogin:  "Launcher started - Auto-login active!",
		SwitchManualLogin: "Please login - session will be saved automatically!",

		// Tray Menu
		TrayOpen: "Open",
		TrayQuit: "Quit",

		// Status Messages
		StatusFillFields:     "Please fill all fields",
		StatusAccountAdded:   "✅ Account added!\n\nLauncher starting...\nPlease login - session will be saved automatically!",
		StatusAccountDeleted: "Account deleted",
		StatusDeleteError:    "Error deleting",
		StatusLauncherRestart: "Restarting launcher...",
		StatusAutoLoginActive: "🚀 AUTO-LOGIN ACTIVE!\n\nAccount: {name}\nLauncher starting automatically logged in!",
		StatusManualLogin:    "⚠️ MANUAL LOGIN\n\nAccount: {name}\nEmail: {email}\n\nPlease login - session will be saved automatically!",
		StatusError:          "Error: {error}",
		StatusPathSaved:      "Launcher path saved!",
		StatusSaveError:      "Error saving",
		StatusEnterPath:      "Please enter a path",
		StatusLanguageSaved:  "Language saved!",

		// Update Notifications
		UpdateAvailableStable: "Update available: {version} — {url}",
		UpdateAvailableBeta:   "Beta available: {version} — {url}",
	},
}

// SetLanguage sets the current language
func SetLanguage(lang string) {
	languageMutex.Lock()
	if lang == "de" || lang == "en" {
		currentLanguage = lang
	}
	languageMutex.Unlock()

	if LanguageChangedCallback != nil {
		LanguageChangedCallback()
	}
}

// GetLanguage returns the current language
func GetLanguage() string {
	languageMutex.RLock()
	defer languageMutex.RUnlock()
	return currentLanguage
}

// T returns the translation for the given key
func T(key string) string {
	languageMutex.RLock()
	defer languageMutex.RUnlock()

	if trans, ok := translations[currentLanguage]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}

	// Fallback to English
	if trans, ok := translations["en"]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}

	return key
}

// TF returns the translation with placeholders replaced
func TF(key string, replacements map[string]string) string {
	text := T(key)
	for placeholder, value := range replacements {
		text = strings.ReplaceAll(text, "{"+placeholder+"}", value)
	}
	return text
}
