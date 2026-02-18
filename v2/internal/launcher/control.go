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
	appData := os.Getenv("APPDATA")

	// All cache directories to clear
	cacheDirs := []string{
		// Temp folder caches
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkov"),
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkovArena"),
		// Launcher CefCache (browser cache - stores background images, etc.)
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "Cache"),
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "Code Cache"),
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "GPUCache"),
		// Session Storage (might store user preferences)
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "Session Storage"),
		// Local Storage (might store background setting)
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "Local Storage"),
	}

	// Files to delete (not directories)
	cacheFiles := []string{
		// Launcher cache log files
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "000003.log"),
	}

	// Remove directories
	for _, dir := range cacheDirs {
		if _, err := os.Stat(dir); err == nil {
			os.RemoveAll(dir)
		}
	}

	// Remove files
	for _, file := range cacheFiles {
		os.Remove(file)
	}

	// Don't delete: appData + "Battlestate Games/BsgLauncher/settings" - we need this for auth!
	_ = appData // silence unused warning
}
