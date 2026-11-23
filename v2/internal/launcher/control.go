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
	// Only the main processes are asked to close: the CEF subprocesses share
	// the exe name and exit with them.
	var mainPIDs []uint32
	for pid, parentPID := range procs {
		if _, isChild := procs[parentPID]; !isChild {
			mainPIDs = append(mainPIDs, pid)
		}
	}
	if !endSession(mainPIDs) {
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
		if !IsLauncherRunning() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// CheckLauncherPath reports whether the configured launcher executable exists.
func CheckLauncherPath() error {
	_, err := os.Stat(config.GetSettings().LauncherPath)
	return err
}

// StartLauncher starts the BSG Launcher
func StartLauncher() error {
	if err := CheckLauncherPath(); err != nil {
		return err
	}

	cmd := exec.Command(config.GetSettings().LauncherPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	_ = cmd.Process.Release()
	return nil
}

// IsLauncherRunning reports whether any BSG launcher process is running.
func IsLauncherRunning() bool {
	return len(launcherProcesses()) > 0
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
	for pid := range procs {
		path := processImagePath(pid)
		if path == "" || !strings.EqualFold(filepath.Clean(path), want) {
			return false
		}
	}
	return true
}

// launcherProcesses returns the parent PID of every running BsgLauncher.exe
// keyed by PID. The launcher's CEF subprocesses run under the same exe name.
func launcherProcesses() map[uint32]uint32 {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(snap)

	procs := make(map[uint32]uint32)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	for err = windows.Process32First(snap, &entry); err == nil; err = windows.Process32Next(snap, &entry) {
		if strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), launcherExe) {
			procs[entry.ProcessID] = entry.ParentProcessID
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

// ClearGameCache clears EFT/Arena and launcher cache files, so nothing cached
// for the previous account carries over to the next one.
// This does NOT delete our saved account sessions - only caches.
func ClearGameCache() {
	tempDir := os.TempDir()
	localAppData := os.Getenv("LOCALAPPDATA")
	paths := config.GetPaths()

	settings := config.GetSettings()
	launcherDir := filepath.Dir(settings.LauncherPath)

	cacheDirs := []string{
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkov"),
		filepath.Join(tempDir, "Battlestate Games", "EscapeFromTarkovArena"),
		filepath.Join(localAppData, "Battlestate Games", "BsgLauncher", "CefCache", "Cache"),
		filepath.Join(launcherDir, "Temp"),
		// The launcher's temp folder: its settings point here (tempFolder).
		paths.TempFolder,
	}

	for _, dir := range cacheDirs {
		os.RemoveAll(dir)
	}
	os.MkdirAll(paths.TempFolder, 0755)
}
