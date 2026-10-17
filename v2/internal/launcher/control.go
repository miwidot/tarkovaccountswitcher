package launcher

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"tarkov-account-switcher/internal/config"
)

// KillLauncher kills the BSG Launcher process and waits for it to exit
func KillLauncher() error {
	cmd := hiddenCommand("taskkill", "/F", "/IM", "BsgLauncher.exe", "/T")
	cmd.Run() // Ignore error - process might not be running

	// Poll for process exit instead of fixed sleep
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		check := hiddenCommand("tasklist", "/FI", "IMAGENAME eq BsgLauncher.exe", "/NH")
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

	if _, err := os.Stat(launcherPath); err != nil {
		return err
	}

	cmd := exec.Command(launcherPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = cmd.Process.Release()
	return nil
}

// IsOnlyConfiguredLauncherRunning reports whether all running BSG launcher
// instances are from the configured installation. Steam and official installs
// share the settings file, so session capture must pause if both are running.
func IsOnlyConfiguredLauncherRunning() bool {
	configuredPath := config.GetSettings().LauncherPath
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
		"$ErrorActionPreference='SilentlyContinue'; Get-Process -Name 'BsgLauncher' | ForEach-Object { $_.Path }").Output()
	if err != nil {
		return false
	}
	want := filepath.Clean(configuredPath)
	found := false
	for _, line := range strings.Split(string(out), "\n") {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		if !strings.EqualFold(filepath.Clean(path), want) {
			return false
		}
		found = true
	}
	return found
}

func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd
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
