package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Settings holds the application settings
type Settings struct {
	LauncherPath string `json:"launcherPath"`
	Language     string `json:"language"`
	StreamerMode bool   `json:"streamerMode"`
	Theme        string `json:"theme"`
	AutoStart    bool   `json:"autoStart"`
}

// Paths holds all the important file paths for the application
type Paths struct {
	DataDir              string
	AccountsFile         string
	SettingsFile         string
	KeyFile              string
	TempFolder           string
	LauncherSettingsPath string
}

const defaultLauncherPath = `C:\Battlestate Games\BsgLauncher\BsgLauncher.exe`

var (
	appPaths  *Paths
	pathsOnce sync.Once

	settingsMu sync.Mutex
	cached     *Settings
)

// GetPaths returns the application paths, initializing them exactly once
func GetPaths() *Paths {
	pathsOnce.Do(func() {
		appData := os.Getenv("APPDATA")
		dataDir := filepath.Join(appData, "TarkovAccountSwitcher")

		appPaths = &Paths{
			DataDir:              dataDir,
			AccountsFile:         filepath.Join(dataDir, "accounts.json"),
			SettingsFile:         filepath.Join(dataDir, "settings.json"),
			KeyFile:              filepath.Join(dataDir, ".key"),
			TempFolder:           filepath.Join(dataDir, "temp"),
			LauncherSettingsPath: filepath.Join(appData, "Battlestate Games", "BsgLauncher", "settings"),
		}
	})
	return appPaths
}

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir() error {
	paths := GetPaths()
	if err := os.MkdirAll(paths.DataDir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(paths.TempFolder, 0755)
}

// WriteFileAtomic writes data to a temporary file next to path and renames it
// over path, so a crash or power loss mid-write (the app writes during
// Windows shutdown) can't leave a truncated file behind.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// loadLocked returns the cached settings, populating from disk on first call.
// Caller must hold settingsMu.
func loadLocked() *Settings {
	if cached != nil {
		return cached
	}
	s := &Settings{LauncherPath: defaultLauncherPath}
	if data, err := os.ReadFile(GetPaths().SettingsFile); err == nil {
		_ = json.Unmarshal(data, s)
	}
	cached = s
	return s
}

// writeLocked persists the cached settings to disk.
// Caller must hold settingsMu.
func writeLocked() error {
	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(GetPaths().SettingsFile, data, 0644)
}

// update mutates the cached settings under lock and persists the result.
func update(mutate func(*Settings)) error {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	mutate(loadLocked())
	return writeLocked()
}

// GetSettings returns a snapshot of the current settings.
func GetSettings() Settings {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	return *loadLocked()
}

// SetLanguage sets and saves the language setting
func SetLanguage(language string) error {
	return update(func(s *Settings) { s.Language = language })
}

// SetLauncherPath sets and saves the launcher path setting
func SetLauncherPath(launcherPath string) error {
	return update(func(s *Settings) { s.LauncherPath = launcherPath })
}

// SetStreamerMode sets and saves the streamer mode setting
func SetStreamerMode(enabled bool) error {
	return update(func(s *Settings) { s.StreamerMode = enabled })
}

// SetTheme sets and saves the theme setting
func SetTheme(id string) error {
	return update(func(s *Settings) { s.Theme = id })
}

// SetAutoStart sets and saves the autostart setting
func SetAutoStart(enabled bool) error {
	return update(func(s *Settings) { s.AutoStart = enabled })
}

// IsStreamerMode returns whether streamer mode is enabled
func IsStreamerMode() bool {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	return loadLocked().StreamerMode
}

// MaskEmail masks an email for streamer mode (e.g. "test@email.com" -> "t***@e***.com")
func MaskEmail(email string) string {
	if !IsStreamerMode() {
		return email
	}

	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return "****"
	}

	local := email[:at]
	domain := email[at+1:]
	maskedLocal := local[:1] + "***"

	dot := strings.IndexByte(domain, '.')
	if dot <= 0 {
		return maskedLocal + "@***"
	}
	return maskedLocal + "@" + domain[:1] + "***" + domain[dot:]
}
