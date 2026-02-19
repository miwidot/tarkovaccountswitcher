package accounts

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"tarkov-account-switcher/internal/config"
	"tarkov-account-switcher/internal/launcher"
)

// Account represents a saved Tarkov account
type Account struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	LauncherSession json.RawMessage `json:"launcherSession"`
	SessionCaptured string          `json:"sessionCaptured,omitempty"`
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

// GetAccounts loads all accounts from file
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

	return accounts, nil
}

// saveAccounts saves all accounts to file
func saveAccounts(accounts []Account) error {
	paths := config.GetPaths()

	data, err := json.MarshalIndent(accounts, "", "  ")
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
			Message:     "Launcher gestartet - Auto-Login aktiv!",
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
		Message:     "Bitte einloggen - Session wird automatisch gespeichert!",
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
			// Save auth-related fields, selectedGame AND ingame background
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
			sessionData, _ := json.Marshal(authSession)
			accounts[i].LauncherSession = sessionData
			accounts[i].SessionCaptured = time.Now().Format(time.RFC3339)
			saveAccounts(accounts)
			return
		}
	}
}

// HasSession checks if an account has a saved session
func (a *Account) HasSession() bool {
	return a.LauncherSession != nil && len(a.LauncherSession) > 0
}
