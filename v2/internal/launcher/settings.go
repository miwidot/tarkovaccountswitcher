package launcher

import (
	"encoding/json"
	"os"
	"path/filepath"

	"tarkov-account-switcher/internal/config"
)

// UpdateLauncherAccount updates the launcher settings with a new account email and clears session
func UpdateLauncherAccount(email string) error {
	paths := config.GetPaths()
	launcherDataPath := filepath.Dir(paths.LauncherSettingsPath)

	// Read existing settings or create new
	var settings map[string]interface{}

	data, err := os.ReadFile(paths.LauncherSettingsPath)
	if err == nil {
		json.Unmarshal(data, &settings)
	}

	if settings == nil {
		settings = make(map[string]interface{})
	}

	// Update login email and settings
	settings["login"] = email
	settings["saveLogin"] = true
	settings["keepLoggedIn"] = false
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

	// Delete session files to force re-login
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

	return nil
}

// RestoreLauncherSession restores a saved launcher session
func RestoreLauncherSession(sessionData json.RawMessage) error {
	paths := config.GetPaths()

	// Parse saved session
	var savedSession map[string]interface{}
	if err := json.Unmarshal(sessionData, &savedSession); err != nil {
		return err
	}

	// Read existing launcher settings to preserve game state
	var existingSettings map[string]interface{}

	data, err := os.ReadFile(paths.LauncherSettingsPath)
	if err == nil {
		json.Unmarshal(data, &existingSettings)
	}

	if existingSettings == nil {
		existingSettings = make(map[string]interface{})
		// Copy all from saved session as base
		for k, v := range savedSession {
			existingSettings[k] = v
		}
	}

	// Restore auth-related fields AND selectedGame from saved session
	// Preserve everything else (games, UI preferences) from current launcher state
	authFields := []string{"login", "at", "rt", "atet", "keepLoggedIn", "saveLogin", "selectedGame"}
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

	return os.WriteFile(paths.LauncherSettingsPath, settingsData, 0644)
}

// ReadLauncherSettings reads the current launcher settings
func ReadLauncherSettings() (map[string]interface{}, error) {
	paths := config.GetPaths()

	data, err := os.ReadFile(paths.LauncherSettingsPath)
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return settings, nil
}
