package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings holds the application settings
type Settings struct {
	LauncherPath string `json:"launcherPath"`
	Language     string `json:"language"`
	StreamerMode bool   `json:"streamerMode"`
}

// Paths holds all the important file paths for the application
type Paths struct {
	DataDir            string
	AccountsFile       string
	SettingsFile       string
	KeyFile            string
	TempFolder         string
	LauncherSettingsPath string
}

var appPaths *Paths

// GetPaths returns the application paths, initializing them if necessary
func GetPaths() *Paths {
	if appPaths == nil {
		appData := os.Getenv("APPDATA")
		dataDir := filepath.Join(appData, "TarkovAccountSwitcher")

		appPaths = &Paths{
			DataDir:            dataDir,
			AccountsFile:       filepath.Join(dataDir, "accounts.json"),
			SettingsFile:       filepath.Join(dataDir, "settings.json"),
			KeyFile:            filepath.Join(dataDir, ".key"),
			TempFolder:         filepath.Join(dataDir, "temp"),
			LauncherSettingsPath: filepath.Join(appData, "Battlestate Games", "BsgLauncher", "settings"),
		}
	}
	return appPaths
}

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir() error {
	paths := GetPaths()
	if err := os.MkdirAll(paths.DataDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(paths.TempFolder, 0755); err != nil {
		return err
	}
	return nil
}

// GetSettings loads settings from file or returns defaults
func GetSettings() *Settings {
	paths := GetPaths()

	settings := &Settings{
		LauncherPath: `C:\Battlestate Games\BsgLauncher\BsgLauncher.exe`,
		Language:     "",
	}

	data, err := os.ReadFile(paths.SettingsFile)
	if err != nil {
		return settings
	}

	json.Unmarshal(data, settings)
	return settings
}

// SaveSettings saves settings to file
func SaveSettings(settings *Settings) error {
	paths := GetPaths()

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(paths.SettingsFile, data, 0644)
}

// SetLanguage sets and saves the language setting
func SetLanguage(language string) error {
	settings := GetSettings()
	settings.Language = language
	return SaveSettings(settings)
}

// SetLauncherPath sets and saves the launcher path setting
func SetLauncherPath(launcherPath string) error {
	settings := GetSettings()
	settings.LauncherPath = launcherPath
	return SaveSettings(settings)
}

// SetStreamerMode sets and saves the streamer mode setting
func SetStreamerMode(enabled bool) error {
	settings := GetSettings()
	settings.StreamerMode = enabled
	return SaveSettings(settings)
}

// IsStreamerMode returns whether streamer mode is enabled
func IsStreamerMode() bool {
	return GetSettings().StreamerMode
}

// MaskEmail masks an email for streamer mode (e.g. "test@email.com" -> "t***@e***.com")
func MaskEmail(email string) string {
	if !IsStreamerMode() {
		return email
	}

	// Find @ position
	atIdx := -1
	for i, c := range email {
		if c == '@' {
			atIdx = i
			break
		}
	}

	if atIdx <= 0 {
		return "****"
	}

	// Mask local part (before @)
	local := email[:atIdx]
	domain := email[atIdx+1:]

	maskedLocal := string(local[0]) + "***"

	// Mask domain (after @)
	dotIdx := -1
	for i, c := range domain {
		if c == '.' {
			dotIdx = i
			break
		}
	}

	var maskedDomain string
	if dotIdx > 0 {
		maskedDomain = string(domain[0]) + "***" + domain[dotIdx:]
	} else {
		maskedDomain = "***"
	}

	return maskedLocal + "@" + maskedDomain
}

// GetSystemLanguage returns the system language (de or en)
func GetSystemLanguage() string {
	// Try to get system locale from environment
	for _, env := range []string{"LANG", "LC_ALL", "LC_MESSAGES"} {
		if locale := os.Getenv(env); locale != "" {
			if len(locale) >= 2 && locale[:2] == "de" {
				return "de"
			}
		}
	}

	// Default to English
	return "en"
}
