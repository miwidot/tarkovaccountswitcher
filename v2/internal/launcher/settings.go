package launcher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tarkov-account-switcher/internal/config"
)

// UpdateLauncherAccount updates the launcher settings with a new account email and clears session
func UpdateLauncherAccount(email string) error {
	paths := config.GetPaths()
	launcherDataPath := filepath.Dir(paths.LauncherSettingsPath)

	settings, err := readLauncherSettings()
	if err != nil {
		return err
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

	return os.WriteFile(paths.LauncherSettingsPath, settingsData, 0644)
}

// readLauncherSettings reads the launcher settings file. It returns nil
// without error if the file doesn't exist yet, and an error if it exists but
// can't be read or parsed: writing it back then would wipe the launcher's own
// settings (game paths, preferences).
func readLauncherSettings() (map[string]any, error) {
	data, err := os.ReadFile(config.GetPaths().LauncherSettingsPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("launcher settings file is unreadable: %w", err)
	}
	return settings, nil
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
	existingSettings, err := readLauncherSettings()
	if err != nil {
		return err
	}

	if existingSettings == nil {
		existingSettings = make(map[string]any)
		// Copy all from saved session as base
		for k, v := range savedSession {
			existingSettings[k] = v
		}
	}

	// Restore auth-related fields AND selectedGame from saved session
	// Preserve everything else (games, UI preferences) from current launcher state
	authFields := []string{"login", "at", "rt", "atet", "keepLoggedIn", "saveLogin", "selectedGame", "sysInfCheck"}
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
