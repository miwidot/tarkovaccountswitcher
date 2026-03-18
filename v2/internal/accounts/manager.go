package accounts

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/i18n"
	"tarkov-account-switcher/internal/launcher"
)

// Account represents a saved Tarkov account
type Account struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	LauncherSession  json.RawMessage `json:"launcherSession,omitempty"`  // legacy plaintext (migrated on load)
	EncryptedSession string          `json:"encryptedSession,omitempty"` // AES-256-CBC encrypted session
	SessionCaptured  string          `json:"sessionCaptured,omitempty"`
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

// GetAccounts loads all accounts from file, decrypting sessions
func GetAccounts() ([]Account, error) {
	paths := config.GetPaths()

	data, err := os.ReadFile(paths.AccountsFile)
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

	needsMigration := false
	for i := range accounts {
		// Migrate legacy plaintext sessions to encrypted
		if len(accounts[i].LauncherSession) > 0 && accounts[i].EncryptedSession == "" {
			encrypted, err := Encrypt(string(accounts[i].LauncherSession))
			if err == nil {
				accounts[i].EncryptedSession = encrypted
				accounts[i].LauncherSession = nil
				needsMigration = true
			}
		}
		// Decrypt encrypted session into LauncherSession for in-memory use
		if accounts[i].EncryptedSession != "" && len(accounts[i].LauncherSession) == 0 {
			decrypted, err := Decrypt(accounts[i].EncryptedSession)
			if err == nil {
				accounts[i].LauncherSession = json.RawMessage(decrypted)
			}
		}
	}

	if needsMigration {
		saveAccounts(accounts)
	}

	return accounts, nil
}

// saveAccounts saves all accounts to file with encrypted sessions
func saveAccounts(accounts []Account) error {
	paths := config.GetPaths()

	// Create a copy for disk — encrypt sessions, clear plaintext
	diskAccounts := make([]Account, len(accounts))
	for i, acc := range accounts {
		diskAccounts[i] = acc
		if len(acc.LauncherSession) > 0 {
			encrypted, err := Encrypt(string(acc.LauncherSession))
			if err == nil {
				diskAccounts[i].EncryptedSession = encrypted
			}
		}
		diskAccounts[i].LauncherSession = nil // never write plaintext to disk
	}

	data, err := json.MarshalIndent(diskAccounts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(paths.AccountsFile, data, 0644)
}

// AddAccount adds a new account and starts the login process
func AddAccount(name, email string) (string, error) {
	accounts, err := GetAccounts()
	if err != nil {
		return "", err
	}

	newAccount := Account{
		ID:              strconv.FormatInt(time.Now().UnixMilli(), 10),
		Name:            name,
		Email:           email,
		LauncherSession: nil,
	}

	accounts = append(accounts, newAccount)
	if err := saveAccounts(accounts); err != nil {
		return "", err
	}

	// Kill launcher and clear session
	launcher.KillLauncher()
	if err := launcher.UpdateLauncherAccount(email); err != nil {
		return "", err
	}

	// Start launcher fresh
	if err := launcher.StartLauncher(); err != nil {
		return "", err
	}

	// Notify UI to minimize
	if launcher.OnLauncherStarted != nil {
		launcher.OnLauncherStarted()
	}

	// Start session watcher after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		StartWatcher(newAccount.ID, email)
	}()

	return newAccount.ID, nil
}

// DeleteAccount removes an account by ID
func DeleteAccount(id string) error {
	accounts, err := GetAccounts()
	if err != nil {
		return err
	}

	filtered := make([]Account, 0, len(accounts))
	for _, acc := range accounts {
		if acc.ID != id {
			filtered = append(filtered, acc)
		}
	}

	return saveAccounts(filtered)
}

// GetAccountByID finds an account by its ID
func GetAccountByID(id string) (*Account, error) {
	accounts, err := GetAccounts()
	if err != nil {
		return nil, err
	}

	for _, acc := range accounts {
		if acc.ID == id {
			return &acc, nil
		}
	}

	return nil, nil
}

// UpdateAccountSession updates an account's session data
func UpdateAccountSession(id string, session json.RawMessage) error {
	accounts, err := GetAccounts()
	if err != nil {
		return err
	}

	for i := range accounts {
		if accounts[i].ID == id {
			accounts[i].LauncherSession = session
			accounts[i].SessionCaptured = time.Now().Format(time.RFC3339)
			break
		}
	}

	return saveAccounts(accounts)
}

// SwitchAccount switches to the specified account
func SwitchAccount(id string) *SwitchResult {
	// First, save current account session to capture refreshed tokens
	SaveCurrentAccountSession()

	// Kill launcher
	launcher.KillLauncher()

	// Clear game cache to force fresh data from server (backgrounds, icons, etc.)
	launcher.ClearGameCache()

	// Get account info
	account, err := GetAccountByID(id)
	if err != nil || account == nil {
		return &SwitchResult{
			Success: false,
			Error:   "Account not found",
		}
	}

	// Check if we have a saved session
	if account.LauncherSession != nil && len(account.LauncherSession) > 0 {
		// Restore session
		if err := launcher.RestoreLauncherSession(account.LauncherSession); err != nil {
			return &SwitchResult{
				Success: false,
				Error:   err.Error(),
			}
		}

		if err := launcher.StartLauncher(); err != nil {
			return &SwitchResult{
				Success: false,
				Error:   err.Error(),
			}
		}

		// Notify UI to minimize
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

	// No session saved - clear session and start fresh
	if err := launcher.UpdateLauncherAccount(account.Email); err != nil {
		return &SwitchResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	if err := launcher.StartLauncher(); err != nil {
		return &SwitchResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Notify UI to minimize
	if launcher.OnLauncherStarted != nil {
		launcher.OnLauncherStarted()
	}

	// Start session watcher after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		StartWatcher(id, account.Email)
	}()

	return &SwitchResult{
		Success:     true,
		AccountName: account.Name,
		Email:       account.Email,
		HasSession:  false,
		Message:     i18n.T(i18n.SwitchManualLogin),
	}
}

// SaveCurrentAccountSession saves the current launcher session if it matches a known account
func SaveCurrentAccountSession() {
	paths := config.GetPaths()

	data, err := os.ReadFile(paths.LauncherSettingsPath)
	if err != nil {
		return
	}

	var launcherSettings map[string]interface{}
	if err := json.Unmarshal(data, &launcherSettings); err != nil {
		return
	}

	// Check if there's a logged in user with valid tokens
	login, _ := launcherSettings["login"].(string)
	at, _ := launcherSettings["at"].(string)
	rt, _ := launcherSettings["rt"].(string)

	if login == "" || at == "" || rt == "" {
		return
	}

	// Find which of our accounts matches this email
	accounts, err := GetAccounts()
	if err != nil {
		return
	}

	for i := range accounts {
		if accounts[i].Email == login {
			sessionData, _ := json.Marshal(BuildAuthSession(launcherSettings))
			accounts[i].LauncherSession = sessionData
			accounts[i].SessionCaptured = time.Now().Format(time.RFC3339)
			saveAccounts(accounts)
			return
		}
	}
}

// HasSession checks if an account has a saved session
func (a *Account) HasSession() bool {
	return (a.LauncherSession != nil && len(a.LauncherSession) > 0) || a.EncryptedSession != ""
}

// BuildAuthSession creates the session map from launcher settings.
// Single source of truth for which fields to capture.
func BuildAuthSession(launcherSettings map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
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
