package launcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"tarkov-account-switcher/internal/config"
)

const launcherExe = "BsgLauncher.exe"

// CloseLauncher asks the BSG launcher to exit cleanly. The launcher keeps
// tokens it refreshes after a raid only in memory and writes them to its
// settings file on a clean exit; BSG invalidates the old refresh token at the
// same time, so a force-killed launcher leaves a dead session behind.
// Returns true once no launcher process is left.
func CloseLauncher(timeout time.Duration) bool {
	procs := launcherProcesses()
	if len(procs) == 0 {
		return true
	}
	// The Restart Manager sends the launcher the messages Windows sends at
	// logoff, which make it exit even when it is set to hide in the tray on
	// close (WM_CLOSE only hides it then). Only the main processes take part:
	// the CEF subprocesses share the exe name but have no windows, and the
	// Restart Manager refuses the whole request if it includes them.
	var mainPIDs []uint32
	for pid, p := range procs {
		if _, isChild := procs[p.parentPID]; !isChild {
			mainPIDs = append(mainPIDs, pid)
		}
	}
	if !shutdownApps(mainPIDs) {
		// taskkill without /F posts WM_CLOSE to the launcher's windows.
		hiddenCommand("taskkill", "/IM", launcherExe).Run()
	}
	return waitForLauncherExit(timeout)
}

// KillLauncher closes the BSG Launcher gracefully and force-kills it only if
// it doesn't exit in time.
func KillLauncher() error {
	if CloseLauncher(8 * time.Second) {
		return nil
	}
	hiddenCommand("taskkill", "/F", "/IM", launcherExe, "/T").Run()
	waitForLauncherExit(3 * time.Second)
	return nil
}

func waitForLauncherExit(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if len(launcherProcesses()) == 0 {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
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
	procs := launcherProcesses()
	if len(procs) == 0 {
		return false
	}
	want := filepath.Clean(config.GetSettings().LauncherPath)
	for _, p := range procs {
		if p.path == "" || !strings.EqualFold(filepath.Clean(p.path), want) {
			return false
		}
	}
	return true
}

type launcherProcess struct {
	path      string // "" if it couldn't be queried
	parentPID uint32
}

// launcherProcesses returns every running BsgLauncher.exe keyed by PID.
// The launcher's CEF subprocesses run under the same exe name.
func launcherProcesses() map[uint32]launcherProcess {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snap)

	procs := make(map[uint32]launcherProcess)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	for err = windows.Process32First(snap, &entry); err == nil; err = windows.Process32Next(snap, &entry) {
		if strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), launcherExe) {
			procs[entry.ProcessID] = launcherProcess{
				path:      processImagePath(entry.ProcessID),
				parentPID: entry.ParentProcessID,
			}
		}
	}
	return procs
}

func processImagePath(pid uint32) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_LONG_PATH)
	n := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &n); err != nil {
		return ""
	}
	return windows.UTF16ToString(buf[:n])
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
