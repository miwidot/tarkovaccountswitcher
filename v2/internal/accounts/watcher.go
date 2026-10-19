package accounts

import (
	"sync"
	"time"

	"tarkov-account-switcher/internal/launcher"
)

var (
	watchMu       sync.Mutex
	watchDeadline time.Time

	// SessionCapturedCallback is called when a session is captured
	SessionCapturedCallback func(accountID string)
)

// WatchLauncherSession stores session changes the launcher writes for the
// next 5 minutes (e.g. the login after switching to an account without a
// saved session). Calling it again extends the window. Kept short on purpose:
// security software treats an app that keeps reading another program's login
// tokens as a credential stealer. Captures pause while another launcher
// installation (Steam) is running, since it shares the settings file.
func WatchLauncherSession() {
	watchMu.Lock()
	defer watchMu.Unlock()
	running := time.Now().Before(watchDeadline)
	watchDeadline = time.Now().Add(5 * time.Minute)
	if running {
		return
	}
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			watchMu.Lock()
			expired := time.Now().After(watchDeadline)
			watchMu.Unlock()
			if expired {
				return
			}
			if !launcher.IsOnlyConfiguredLauncherRunning() {
				continue
			}
			if id := captureLauncherSession(); id != "" && SessionCapturedCallback != nil {
				SessionCapturedCallback(id)
			}
		}
	}()
}
