package launcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"tarkov-account-switcher/internal/config"
)

// KillLauncher kills the BSG Launcher process
func KillLauncher() error {
	// Kill BsgLauncher process - ignore errors as process might not be running
	cmd := exec.Command("taskkill", "/F", "/IM", "BsgLauncher.exe", "/T")
	cmd.Run() // Ignore error

	// Wait a bit to ensure process is killed
	time.Sleep(1500 * time.Millisecond)

	return nil
}

// StartLauncher starts the BSG Launcher
func StartLauncher() error {
	settings := config.GetSettings()
	launcherPath := settings.LauncherPath

	if _, err := os.Stat(launcherPath); os.IsNotExist(err) {
		return err
	}

	cmd := exec.Command(launcherPath)
	return cmd.Start()
}

// OnLauncherStarted is called after launcher starts - set by UI to minimize window
var OnLauncherStarted func()

// ClearGameCache clears EFT/Arena cache files to force fresh data from server
// This does NOT delete our saved account sessions - only game cache
func ClearGameCache() {
	// Get paths
	tempDir := os.TempDir()
	localAppData := os.Getenv("LOCALAPPDATA")

	// Get launcher installation directory from settings
	settings := config.GetSettings()
	launcherDir := filepath.Dir(settings.LauncherPath)

	// Cache directories to clear (NOT the entire CefCache - that breaks background loading)
	cacheDirs := []string{
		// Temp folder caches
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkov"),
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkovArena"),

		// Only clear the image/resource cache inside CefCache, not the whole folder
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "Cache"),

		// Launcher installation cache
		filepath.Join(launcherDir, "Temp"),
	}

	// Remove directories
	for _, dir := range cacheDirs {
		if _, err := os.Stat(dir); err == nil {
			os.RemoveAll(dir)
		}
	}
}
