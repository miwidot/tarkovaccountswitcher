package accounts

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/launcher"
)

var (
	watcherMutex     sync.Mutex
	watcherRunning   bool
	watcherAccountID string
	stopChan         chan struct{}

	// SessionCapturedCallback is called when a session is captured
	SessionCapturedCallback func(accountID string)
)

// StartWatcher starts watching for session tokens
func StartWatcher(accountID, expectedEmail string) {
	watcherMutex.Lock()

	// Stop any existing watcher
	if watcherRunning && stopChan != nil {
		close(stopChan)
	}

	watcherRunning = true
	watcherAccountID = accountID
	stopChan = make(chan struct{})
	localStopChan := stopChan

	watcherMutex.Unlock()

	paths := config.GetPaths()
	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(5 * time.Minute)

	defer ticker.Stop()

	for {
		select {
		case <-localStopChan:
			return

		case <-timeout:
			watcherMutex.Lock()
			if watcherAccountID == accountID {
				watcherRunning = false
				watcherAccountID = ""
			}
			watcherMutex.Unlock()
			return

		case <-ticker.C:
			data, err := os.ReadFile(paths.LauncherSettingsPath)
			if err != nil {
				continue
			}

			var launcherSettings map[string]interface{}
			if err := json.Unmarshal(data, &launcherSettings); err != nil {
				continue
			}

			// Check if user logged in with correct email and has session tokens
			login, _ := launcherSettings["login"].(string)
			at, _ := launcherSettings["at"].(string)
			rt, _ := launcherSettings["rt"].(string)

			if login == expectedEmail && at != "" && rt != "" {
				// Session detected - capture auth fields, selectedGame AND ingame background
				authSession := map[string]interface{}{
					"login":             launcherSettings["login"],
					"at":                launcherSettings["at"],
					"rt":                launcherSettings["rt"],
					"atet":              launcherSettings["atet"],
					"keepLoggedIn":      launcherSettings["keepLoggedIn"],
					"saveLogin":         launcherSettings["saveLogin"],
					"selectedGame":      launcherSettings["selectedGame"],
					"environmentUiType": launcher.ReadEnvironmentUiType(),
				}
				sessionData, err := json.Marshal(authSession)
				if err != nil {
					continue
				}

				if err := UpdateAccountSession(accountID, sessionData); err != nil {
					continue
				}

				// Stop watcher
				watcherMutex.Lock()
				if watcherAccountID == accountID {
					watcherRunning = false
					watcherAccountID = ""
				}
				watcherMutex.Unlock()

				// Notify callback
				if SessionCapturedCallback != nil {
					SessionCapturedCallback(accountID)
				}

				return
			}
		}
	}
}

// StopWatcher stops the current session watcher
func StopWatcher() {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	if watcherRunning && stopChan != nil {
		close(stopChan)
		watcherRunning = false
		watcherAccountID = ""
		stopChan = nil
	}
}

// IsWatching returns whether the watcher is currently running
func IsWatching() bool {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()
	return watcherRunning
}

// GetWatchingAccountID returns the account ID being watched
func GetWatchingAccountID() string {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()
	return watcherAccountID
}
