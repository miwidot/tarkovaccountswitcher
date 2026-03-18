package launcher

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"tarkov-account-switcher/internal/config"
)

// KillLauncher kills the BSG Launcher process and waits for it to exit
func KillLauncher() error {
	cmd := exec.Command("taskkill", "/F", "/IM", "BsgLauncher.exe", "/T")
	cmd.Run() // Ignore error - process might not be running

	// Poll for process exit instead of fixed sleep
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		check := exec.Command("tasklist", "/FI", "IMAGENAME eq BsgLauncher.exe", "/NH")
		out, _ := check.Output()
		if !bytes.Contains(out, []byte("BsgLauncher")) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

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
	tempDir := os.TempDir()
	localAppData := os.Getenv("LOCALAPPDATA")

	settings := config.GetSettings()
	launcherDir := filepath.Dir(settings.LauncherPath)

	cacheDirs := []string{
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkov"),
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkovArena"),
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "Cache"),
		filepath.Join(launcherDir, "Temp"),
	}

	for _, dir := range cacheDirs {
		os.RemoveAll(dir)
	}
}
