package accounts

import (
	"sync"
	"time"

	"tarkov-account-switcher/internal/launcher"
)

var (
	monitorOnce sync.Once

	// SessionCapturedCallback is called when a session is captured
	SessionCapturedCallback func(accountID string)
)

// StartSessionMonitor keeps saved sessions in sync with the launcher for as
// long as the app runs. The launcher rewrites its settings file when it logs
// in or refreshes an expired token at startup, and BSG invalidates the
// previous refresh token at that moment: a session saved before is dead from
// then on, and restoring it on the next switch would force a fresh login.
// Every change is stored for the account matching the launcher's login.
// Captures pause while another launcher installation (Steam) is running,
// since it shares the settings file.
func StartSessionMonitor() {
	monitorOnce.Do(func() {
		go func() {
			var lastAT, lastRT string
			for range time.Tick(2 * time.Second) {
				if !launcher.IsOnlyConfiguredLauncherRunning() {
					continue
				}
				launcherSettings, login, err := readLauncherAuth()
				if err != nil || launcherSettings == nil {
					continue
				}
				at, _ := launcherSettings["at"].(string)
				rt, _ := launcherSettings["rt"].(string)
				if at == lastAT && rt == lastRT {
					continue
				}
				lastAT, lastRT = at, rt
				if id := storeLauncherSession(launcherSettings, login); id != "" && SessionCapturedCallback != nil {
					SessionCapturedCallback(id)
				}
			}
		}()
	})
}
