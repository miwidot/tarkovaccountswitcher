package launcher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"tarkov-account-switcher/internal/config"
)

// UpdateLauncherAccount updates the launcher settings with a new account email and clears session
func UpdateLauncherAccount(email string) error {
	paths := config.GetPaths()
	launcherDataPath := filepath.Dir(paths.LauncherSettingsPath)

	// Read existing settings or create new
	var settings map[string]any

	data, err := os.ReadFile(paths.LauncherSettingsPath)
	if err == nil {
		json.Unmarshal(data, &settings)
	}

	if settings == nil {
		settings = make(map[string]any)
	}

	// Update login email and settings
	settings["login"] = email
	settings["saveLogin"] = true
	settings["keepLoggedIn"] = true
	settings["tempFolder"] = paths.TempFolder

	// CRITICAL: Delete session tokens to force fresh login
	delete(settings, "at")
	delete(settings, "rt")
	delete(settings, "atet")

	// Ensure directory exists
	if err := os.MkdirAll(launcherDataPath, 0755); err != nil {
		return err
	}

	// Write settings
	settingsData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(paths.LauncherSettingsPath, settingsData, 0644); err != nil {
		return err
	}

	clearLauncherSessionFiles(launcherDataPath)

	return nil
}

// clearLauncherSessionFiles removes separate launcher auth caches that may
// override the credentials restored from the settings file.
func clearLauncherSessionFiles(launcherDataPath string) {
	sessionFiles := []string{
		"user.json",
		"session",
		"token",
		".session",
		"auth",
		"auth.json",
		"login",
		"login.json",
	}

	for _, file := range sessionFiles {
		filePath := filepath.Join(launcherDataPath, file)
		os.Remove(filePath) // Ignore errors
	}
}

// RestoreLauncherSession restores a saved launcher session
func RestoreLauncherSession(sessionData json.RawMessage) error {
	paths := config.GetPaths()

	// Parse saved session
	var savedSession map[string]any
	if err := json.Unmarshal(sessionData, &savedSession); err != nil {
		return err
	}

	// Read existing launcher settings to preserve game state
	var existingSettings map[string]any

	data, err := os.ReadFile(paths.LauncherSettingsPath)
	if err == nil {
		json.Unmarshal(data, &existingSettings)
	}

	if existingSettings == nil {
		existingSettings = make(map[string]any)
		// Copy all from saved session as base
		for k, v := range savedSession {
			existingSettings[k] = v
		}
	}

	clearLauncherSessionFiles(filepath.Dir(paths.LauncherSettingsPath))

	// Clear known session fields even if an older saved session omitted one,
	// so credentials from the currently active Steam account cannot leak in.
	authFields := []string{"login", "at", "rt", "atet", "keepLoggedIn", "saveLogin", "selectedGame", "sysInfCheck"}
	for _, field := range authFields {
		delete(existingSettings, field)
	}

	// Launcher settings may gain auth fields in future versions. Remove unknown
	// identity/credential fields while preserving game and UI preferences.
	for field := range existingSettings {
		if isLauncherAuthSetting(field) {
			delete(existingSettings, field)
		}
	}

	for _, field := range authFields {
		if val, ok := savedSession[field]; ok {
			existingSettings[field] = val
		}
	}

	// ALWAYS use our own temp folder
	existingSettings["tempFolder"] = paths.TempFolder

	// Write settings
	settingsData, err := json.MarshalIndent(existingSettings, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(paths.LauncherSettingsPath, settingsData, 0644); err != nil {
		return err
	}

	// Restore ingame background (EnvironmentUiType) to Game.ini
	if envType, ok := savedSession["environmentUiType"].(string); ok && envType != "" {
		WriteEnvironmentUiType(envType)
	}

	return nil
}

func isLauncherAuthSetting(field string) bool {
	key := strings.ToLower(field)
	switch key {
	case "at", "rt", "atet", "sysinfcheck":
		return true
	}
	for _, marker := range []string{"auth", "token", "session", "credential", "identity", "steam", "password", "provider", "login", "account"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return key == "user" || strings.HasPrefix(key, "user_") || strings.HasSuffix(key, "_user")
}

// GetGameSettingsPath returns the path to the EFT Game.ini file
func GetGameSettingsPath() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "Battlestate Games", "Escape from Tarkov", "Settings", "Game.ini")
}

// ReadEnvironmentUiType reads the current EnvironmentUiType (ingame background) from Game.ini
func ReadEnvironmentUiType() string {
	data, err := os.ReadFile(GetGameSettingsPath())
	if err != nil {
		return ""
	}

	var gameSettings map[string]any
	if err := json.Unmarshal(data, &gameSettings); err != nil {
		return ""
	}

	if envType, ok := gameSettings["EnvironmentUiType"].(string); ok {
		return envType
	}
	return ""
}

// WriteEnvironmentUiType writes the EnvironmentUiType (ingame background) to Game.ini
func WriteEnvironmentUiType(envType string) error {
	if envType == "" {
		return nil // Nothing to restore
	}

	gameSettingsPath := GetGameSettingsPath()

	// Read existing settings
	data, err := os.ReadFile(gameSettingsPath)
	if err != nil {
		return err // Game.ini must exist
	}

	var gameSettings map[string]any
	if err := json.Unmarshal(data, &gameSettings); err != nil {
		return err
	}

	// Update EnvironmentUiType
	gameSettings["EnvironmentUiType"] = envType

	// Write back
	settingsData, err := json.MarshalIndent(gameSettings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(gameSettingsPath, settingsData, 0644)
}
