package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
	"tarkov-account-switcher/internal/launcher"
)

// Account represents a saved Tarkov account.
// LauncherSession holds decrypted plaintext only transiently in memory;
// it is re-encrypted into EncryptedSession on save and never written plain.
type Account struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	LauncherSession  json.RawMessage `json:"launcherSession,omitempty"`  // legacy plaintext (migrated on load)
	EncryptedSession string          `json:"encryptedSession,omitempty"` // AES-256-CBC encrypted session
	SessionCaptured  string          `json:"sessionCaptured,omitempty"`
}

// HasSession reports whether the account has a stored session (encrypted or legacy in-memory).
func (a *Account) HasSession() bool {
	return len(a.LauncherSession) > 0 || a.EncryptedSession != ""
}

// SwitchResult holds the result of a switch operation
type SwitchResult struct {
	Success     bool
	AccountName string
	Email       string
	HasSession  bool
	Message     string
	Error       string
}

// accountsMu serializes all read-modify-write cycles on accounts.json so the
// session watcher and UI operations can't lose each other's updates.
var accountsMu sync.Mutex

// loadRaw reads accounts.json without migrating or decrypting anything.
// Caller must hold accountsMu.
func loadRaw() ([]Account, error) {
	data, err := os.ReadFile(config.GetPaths().AccountsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Account{}, nil
		}
		return nil, err
	}
	var accounts []Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

// loadMigrated loads accounts and migrates any legacy plaintext sessions into
// EncryptedSession. Persists the migration if anything changed.
// Sessions remain encrypted in the returned slice — callers needing plaintext
// must call Decrypt explicitly.
// Caller must hold accountsMu.
func loadMigrated() ([]Account, error) {
	accounts, err := loadRaw()
	if err != nil {
		return nil, err
	}
	migrated := false
	for i := range accounts {
		if len(accounts[i].LauncherSession) == 0 {
			continue
		}
		if accounts[i].EncryptedSession == "" {
			enc, err := Encrypt(string(accounts[i].LauncherSession))
			if err != nil {
				continue // leave legacy in place; retry next load
			}
			accounts[i].EncryptedSession = enc
		}
		accounts[i].LauncherSession = nil
		migrated = true
	}
	if migrated {
		if err := saveLocked(accounts); err != nil {
			return nil, err
		}
	}
	return accounts, nil
}

// saveLocked persists accounts to disk, encrypting any plaintext sessions on the way.
// Caller must hold accountsMu.
func saveLocked(accounts []Account) error {
	disk := make([]Account, len(accounts))
	for i, acc := range accounts {
		disk[i] = acc
		if len(acc.LauncherSession) > 0 {
			enc, err := Encrypt(string(acc.LauncherSession))
			if err != nil {
				return fmt.Errorf("encrypt session for %s: %w", acc.ID, err)
			}
			disk[i].EncryptedSession = enc
		}
		disk[i].LauncherSession = nil
	}
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.GetPaths().AccountsFile, data, 0644)
}

// mutate runs the given function inside a locked load-modify-save cycle.
// Return nil from mutate to save the accounts, or a non-nil error to abort.
func mutate(fn func([]Account) ([]Account, error)) error {
	accountsMu.Lock()
	defer accountsMu.Unlock()
	accounts, err := loadMigrated()
	if err != nil {
		return err
	}
	updated, err := fn(accounts)
	if err != nil {
		return err
	}
	return saveLocked(updated)
}

// ListAccounts returns all accounts for display. Sessions stay encrypted.
func ListAccounts() ([]Account, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()
	return loadMigrated()
}

// GetAccountByID returns the account with its session decrypted into LauncherSession.
func GetAccountByID(id string) (*Account, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()

	accounts, err := loadMigrated()
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		if accounts[i].ID != id {
			continue
		}
		if accounts[i].EncryptedSession != "" {
			if dec, err := Decrypt(accounts[i].EncryptedSession); err == nil {
				accounts[i].LauncherSession = json.RawMessage(dec)
			}
		}
		return &accounts[i], nil
	}
	return nil, nil
}

// AddAccount adds a new account and kicks off the login flow.
func AddAccount(name, email string) (string, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	newAccount := Account{
		ID:    strconv.FormatInt(time.Now().UnixMilli(), 10),
		Name:  name,
		Email: email,
	}

	if err := mutate(func(accs []Account) ([]Account, error) {
		for _, acc := range accs {
			if strings.EqualFold(strings.TrimSpace(acc.Email), email) {
				return nil, errors.New(i18n.T(i18n.ErrDuplicateEmail))
			}
		}
		return append(accs, newAccount), nil
	}); err != nil {
		return "", err
	}

	// Launcher I/O happens outside the accounts lock.
	closeLauncherSavingSession()
	if err := launcher.UpdateLauncherAccount(email); err != nil {
		return "", err
	}
	if err := launcher.StartLauncher(); err != nil {
		return "", err
	}
	if launcher.OnLauncherStarted != nil {
		launcher.OnLauncherStarted()
	}

	return newAccount.ID, nil
}

// DeleteAccount removes an account by ID.
func DeleteAccount(id string) error {
	return mutate(func(accs []Account) ([]Account, error) {
		filtered := accs[:0]
		for _, acc := range accs {
			if acc.ID != id {
				filtered = append(filtered, acc)
			}
		}
		return filtered, nil
	})
}

// SaveSessionBeforeShutdown runs while Windows waits to log off or shut down.
// Left alone, the launcher starts closing on its own but is killed before it
// writes tokens refreshed since its last save, so the next start needs a
// fresh login. Closing it here, while shutdown is still blocked, gives it time
// to save; the result is then stored for the matching account.
func SaveSessionBeforeShutdown() {
	if !launcher.IsOnlyConfiguredLauncherRunning() {
		return
	}
	// No force-kill fallback: Windows terminates the launcher anyway.
	if launcher.CloseLauncher(8 * time.Second) {
		captureLauncherSession()
	}
}

// closeLauncherSavingSession closes the launcher so it writes its current
// tokens, then stores them for the matching account. Must run before anything
// rewrites the launcher settings file. Skips the capture if another launcher
// installation (Steam) was running, since it shares the settings file.
func closeLauncherSavingSession() {
	onlyOurs := launcher.IsOnlyConfiguredLauncherRunning()
	launcher.KillLauncher()
	if onlyOurs {
		captureLauncherSession()
	}
}

// captureLauncherSession stores the session from the launcher settings file
// for the account whose email matches its login. Returns that account's ID if
// its saved session changed, "" otherwise.
func captureLauncherSession() string {
	launcherSettings, login, err := readLauncherAuth()
	if err != nil || launcherSettings == nil {
		return ""
	}
	return storeLauncherSession(launcherSettings, login)
}

var errSessionUnchanged = errors.New("session unchanged")

func storeLauncherSession(launcherSettings map[string]any, login string) string {
	sessionData, err := json.Marshal(BuildAuthSession(launcherSettings))
	if err != nil {
		return ""
	}
	at, _ := launcherSettings["at"].(string)
	rt, _ := launcherSettings["rt"].(string)
	var capturedID string
	err = mutate(func(accs []Account) ([]Account, error) {
		for i := range accs {
			if !sameEmail(accs[i].Email, login) {
				continue
			}
			if savedAT, savedRT := savedTokens(accs[i]); savedAT == at && savedRT == rt {
				return nil, errSessionUnchanged
			}
			accs[i].LauncherSession = sessionData
			accs[i].SessionCaptured = time.Now().Format(time.RFC3339)
			capturedID = accs[i].ID
			return accs, nil
		}
		return nil, errSessionUnchanged
	})
	if err != nil {
		return ""
	}
	return capturedID
}

// savedTokens returns the access and refresh token of an account's saved session.
func savedTokens(acc Account) (at, rt string) {
	if acc.EncryptedSession == "" {
		return "", ""
	}
	dec, err := Decrypt(acc.EncryptedSession)
	if err != nil {
		return "", ""
	}
	var tokens struct {
		AT string `json:"at"`
		RT string `json:"rt"`
	}
	_ = json.Unmarshal([]byte(dec), &tokens)
	return tokens.AT, tokens.RT
}

// readLauncherAuth reads the launcher settings file. It returns nil settings
// (and no error) when the file holds no complete login.
func readLauncherAuth() (map[string]any, string, error) {
	data, err := os.ReadFile(config.GetPaths().LauncherSettingsPath)
	if err != nil {
		return nil, "", err
	}
	var launcherSettings map[string]any
	if err := json.Unmarshal(data, &launcherSettings); err != nil {
		return nil, "", err
	}
	login, _ := launcherSettings["login"].(string)
	at, _ := launcherSettings["at"].(string)
	rt, _ := launcherSettings["rt"].(string)
	if login == "" || at == "" || rt == "" {
		return nil, "", nil
	}
	return launcherSettings, login, nil
}

func sameEmail(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// ReorderAccounts persists accounts in the given ID order.
func ReorderAccounts(orderedIDs []string) error {
	return mutate(func(accs []Account) ([]Account, error) {
		if len(orderedIDs) != len(accs) {
			return nil, fmt.Errorf("reorder: expected %d ids, got %d", len(accs), len(orderedIDs))
		}
		byID := make(map[string]Account, len(accs))
		for _, acc := range accs {
			byID[acc.ID] = acc
		}
		reordered := make([]Account, 0, len(orderedIDs))
		seen := make(map[string]bool, len(orderedIDs))
		for _, id := range orderedIDs {
			if seen[id] {
				return nil, fmt.Errorf("reorder: duplicate account id %q", id)
			}
			seen[id] = true
			acc, ok := byID[id]
			if !ok {
				return nil, fmt.Errorf("reorder: unknown account id %q", id)
			}
			reordered = append(reordered, acc)
		}
		return reordered, nil
	})
}

// UpdateAccountSession stores a freshly captured session for an account.
func UpdateAccountSession(id string, session json.RawMessage) error {
	return mutate(func(accs []Account) ([]Account, error) {
		for i := range accs {
			if accs[i].ID == id {
				accs[i].LauncherSession = session
				accs[i].SessionCaptured = time.Now().Format(time.RFC3339)
				break
			}
		}
		return accs, nil
	})
}

// SwitchAccount switches to the specified account.
func SwitchAccount(id string) *SwitchResult {
	closeLauncherSavingSession()
	launcher.ClearGameCache()

	account, err := GetAccountByID(id)
	if err != nil || account == nil {
		return &SwitchResult{Success: false, Error: i18n.T(i18n.ErrAccountNotFound)}
	}

	if len(account.LauncherSession) > 0 {
		if err := launcher.RestoreLauncherSession(account.LauncherSession); err != nil {
			return &SwitchResult{Success: false, Error: err.Error()}
		}
		if err := launcher.StartLauncher(); err != nil {
			return &SwitchResult{Success: false, Error: err.Error()}
		}
		if launcher.OnLauncherStarted != nil {
			launcher.OnLauncherStarted()
		}
		return &SwitchResult{
			Success:     true,
			AccountName: account.Name,
			Email:       account.Email,
			HasSession:  true,
			Message:     i18n.T(i18n.SwitchAutoLogin),
		}
	}

	// No saved session — clear launcher tokens and start fresh.
	if err := launcher.UpdateLauncherAccount(account.Email); err != nil {
		return &SwitchResult{Success: false, Error: err.Error()}
	}
	if err := launcher.StartLauncher(); err != nil {
		return &SwitchResult{Success: false, Error: err.Error()}
	}
	if launcher.OnLauncherStarted != nil {
		launcher.OnLauncherStarted()
	}

	return &SwitchResult{
		Success:     true,
		AccountName: account.Name,
		Email:       account.Email,
		HasSession:  false,
		Message:     i18n.T(i18n.SwitchManualLogin),
	}
}

// BuildAuthSession creates the session map from launcher settings.
// Single source of truth for which fields to capture.
func BuildAuthSession(launcherSettings map[string]any) map[string]any {
	return map[string]any{
		"login":             launcherSettings["login"],
		"at":                launcherSettings["at"],
		"rt":                launcherSettings["rt"],
		"atet":              launcherSettings["atet"],
		"sysInfCheck":       launcherSettings["sysInfCheck"],
		"keepLoggedIn":      true,
		"saveLogin":         true,
		"selectedGame":      launcherSettings["selectedGame"],
		"environmentUiType": launcher.ReadEnvironmentUiType(),
	}
}
