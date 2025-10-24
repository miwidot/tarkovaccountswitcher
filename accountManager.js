const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const { exec } = require('child_process');
const util = require('util');

const execPromise = util.promisify(exec);

class AccountManager {
  constructor() {
    this.dataPath = path.join(process.env.APPDATA, 'TarkovAccountSwitcher');
    this.accountsFile = path.join(this.dataPath, 'accounts.json');
    this.settingsFile = path.join(this.dataPath, 'settings.json');
    this.keyFile = path.join(this.dataPath, '.key');
    this.tempFolder = path.join(this.dataPath, 'temp');  // Our own temp folder
    this.launcherSettingsPath = path.join(process.env.APPDATA, 'Battlestate Games', 'BsgLauncher', 'settings');

    this.ensureDataDir();
    this.ensureTempDir();
    this.encryptionKey = this.getOrCreateKey();

    this.sessionWatcher = null;
    this.watchingAccountId = null;
    this.mainWindow = null; // Will be set by main.js
  }

  setMainWindow(window) {
    this.mainWindow = window;
  }

  ensureDataDir() {
    if (!fs.existsSync(this.dataPath)) {
      fs.mkdirSync(this.dataPath, { recursive: true });
    }
  }

  ensureTempDir() {
    if (!fs.existsSync(this.tempFolder)) {
      fs.mkdirSync(this.tempFolder, { recursive: true });
    }
  }

  getOrCreateKey() {
    if (fs.existsSync(this.keyFile)) {
      return fs.readFileSync(this.keyFile);
    }
    const key = crypto.randomBytes(32);
    fs.writeFileSync(this.keyFile, key);
    return key;
  }

  encrypt(text) {
    const iv = crypto.randomBytes(16);
    const cipher = crypto.createCipheriv('aes-256-cbc', this.encryptionKey, iv);
    let encrypted = cipher.update(text, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    return iv.toString('hex') + ':' + encrypted;
  }

  decrypt(text) {
    const parts = text.split(':');
    const iv = Buffer.from(parts[0], 'hex');
    const encrypted = parts[1];
    const decipher = crypto.createDecipheriv('aes-256-cbc', this.encryptionKey, iv);
    let decrypted = decipher.update(encrypted, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    return decrypted;
  }

  getAccounts() {
    if (!fs.existsSync(this.accountsFile)) {
      return [];
    }
    const data = fs.readFileSync(this.accountsFile, 'utf8');
    return JSON.parse(data);
  }

  async addAccount(account) {
    const accounts = this.getAccounts();

    const newAccount = {
      id: Date.now().toString(),
      name: account.name,
      email: account.email,
      launcherSession: null  // Will be captured automatically after login
    };

    accounts.push(newAccount);
    fs.writeFileSync(this.accountsFile, JSON.stringify(accounts, null, 2));

    // Kill launcher and clear session first (like switching accounts)
    await this.killLauncher();
    await this.updateLauncherAccount(newAccount.email);

    // Start launcher fresh
    await this.startLauncher();

    // Start session watcher
    setTimeout(() => {
      console.log('Starting session watcher for new account...');
      this.startSessionWatcher(newAccount.id, account.email);
    }, 2000);

    return { success: true, id: newAccount.id };
  }

  deleteAccount(id) {
    let accounts = this.getAccounts();
    accounts = accounts.filter(acc => acc.id !== id);
    fs.writeFileSync(this.accountsFile, JSON.stringify(accounts, null, 2));
    return { success: true };
  }

  getSettings() {
    if (!fs.existsSync(this.settingsFile)) {
      return {
        launcherPath: 'C:\\Battlestate Games\\BsgLauncher\\BsgLauncher.exe',
        language: null  // Will be set by main.js based on system locale
      };
    }
    const data = fs.readFileSync(this.settingsFile, 'utf8');
    return JSON.parse(data);
  }

  setLanguage(language) {
    const settings = this.getSettings();
    settings.language = language;
    return this.saveSettings(settings);
  }

  saveSettings(settings) {
    fs.writeFileSync(this.settingsFile, JSON.stringify(settings, null, 2));
    return { success: true };
  }

  setLauncherPath(launcherPath) {
    const settings = this.getSettings();
    settings.launcherPath = launcherPath;
    return this.saveSettings(settings);
  }

  async killLauncher() {
    try {
      // Kill BsgLauncher process
      await execPromise('taskkill /F /IM BsgLauncher.exe /T');
      // Wait a bit to ensure process is killed
      await new Promise(resolve => setTimeout(resolve, 1500));
      return true;
    } catch (error) {
      // Process might not be running, that's ok
      return true;
    }
  }

  captureLauncherSession(accountId) {
    try {
      const launcherSettingsPath = path.join(process.env.APPDATA, 'Battlestate Games', 'BsgLauncher', 'settings');

      if (!fs.existsSync(launcherSettingsPath)) {
        return { success: false, error: 'Launcher settings not found' };
      }

      const launcherSettings = JSON.parse(fs.readFileSync(launcherSettingsPath, 'utf8'));

      // Check if user is logged in (has tokens)
      if (!launcherSettings.at || !launcherSettings.rt) {
        return { success: false, error: 'No active session found' };
      }

      // Update account with session data
      const accounts = this.getAccounts();
      const account = accounts.find(acc => acc.id === accountId);

      if (!account) {
        return { success: false, error: 'Account not found' };
      }

      // Check if the logged in email matches the account email
      if (launcherSettings.login !== account.email) {
        return { success: false, error: 'Wrong account logged in (expected: ' + account.email + ', got: ' + launcherSettings.login + ')' };
      }

      // Save session with our own temp folder path
      // This ensures launcher always uses our controlled temp directory
      const sessionToSave = { ...launcherSettings };
      sessionToSave.tempFolder = this.tempFolder;
      // Keep gamesRootDir if it exists, otherwise launcher will use default

      // Store the launcher settings
      account.launcherSession = sessionToSave;
      account.sessionCaptured = new Date().toISOString();

      fs.writeFileSync(this.accountsFile, JSON.stringify(accounts, null, 2));

      return { success: true, message: 'Session captured!' };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async restoreLauncherSession(launcherSession) {
    try {
      const launcherSettingsPath = path.join(process.env.APPDATA, 'Battlestate Games', 'BsgLauncher', 'settings');

      // Read existing settings ONLY to preserve gamesRootDir if it exists
      let existingGamesRootDir = null;
      if (fs.existsSync(launcherSettingsPath)) {
        try {
          const existingSettings = JSON.parse(fs.readFileSync(launcherSettingsPath, 'utf8'));
          existingGamesRootDir = existingSettings.gamesRootDir;
        } catch (err) {
          // If file is corrupt, ignore
        }
      }

      // Use ONLY our saved session, don't merge with existing settings
      // This prevents conflicts with old/expired tokens in launcher settings
      const settingsToWrite = {
        ...launcherSession,      // Our complete saved session
        gamesRootDir: existingGamesRootDir || launcherSession.gamesRootDir,  // Preserve gamesRootDir if exists
        tempFolder: this.tempFolder  // ALWAYS use our own temp folder
      };

      fs.writeFileSync(launcherSettingsPath, JSON.stringify(settingsToWrite, null, 2));

      return { success: true };
    } catch (error) {
      console.error('Error restoring session:', error);
      return { success: false, error: error.message };
    }
  }

  async updateLauncherAccount(email) {
    try {
      // Update launcher settings with new account email
      const launcherDataPath = path.join(process.env.APPDATA, 'Battlestate Games', 'BsgLauncher');
      const settingsFile = path.join(launcherDataPath, 'settings');

      // Update the email in settings file
      if (fs.existsSync(settingsFile)) {
        try {
          const settingsData = fs.readFileSync(settingsFile, 'utf8');
          const settings = JSON.parse(settingsData);

          // Update login email and enable saveLogin
          settings.login = email;
          settings.saveLogin = true;
          settings.keepLoggedIn = false;
          settings.tempFolder = this.tempFolder;  // Use our temp folder

          // CRITICAL: Delete session tokens to force fresh login
          delete settings.at;
          delete settings.rt;
          delete settings.atet;

          fs.writeFileSync(settingsFile, JSON.stringify(settings, null, 2));
        } catch (err) {
          console.error('Error updating login in settings:', err);
        }
      } else {
        // Create settings file if it doesn't exist
        const defaultSettings = {
          games: [],
          login: email,
          saveLogin: true,
          keepLoggedIn: false,
          tempFolder: this.tempFolder,  // Use our temp folder
          language: 'en',
          settingsVersion: 2
        };

        if (!fs.existsSync(launcherDataPath)) {
          fs.mkdirSync(launcherDataPath, { recursive: true });
        }

        fs.writeFileSync(settingsFile, JSON.stringify(defaultSettings, null, 2));
      }

      // Delete session files to force re-login
      const sessionFiles = [
        path.join(launcherDataPath, 'user.json'),
        path.join(launcherDataPath, 'session'),
        path.join(launcherDataPath, 'token'),
        path.join(launcherDataPath, '.session'),
        path.join(launcherDataPath, 'auth'),
        path.join(launcherDataPath, 'auth.json'),
        path.join(launcherDataPath, 'login'),
        path.join(launcherDataPath, 'login.json'),
      ];

      for (const file of sessionFiles) {
        if (fs.existsSync(file)) {
          try {
            fs.unlinkSync(file);
          } catch (err) {
            // Ignore errors, file might be locked
          }
        }
      }

      return true;
    } catch (error) {
      // Don't fail if update fails
      console.error('Account update error:', error);
      return true;
    }
  }

  async startLauncher() {
    try {
      const settings = this.getSettings();
      const launcherPath = settings.launcherPath;

      if (!fs.existsSync(launcherPath)) {
        throw new Error('Launcher not found at: ' + launcherPath);
      }

      exec(`"${launcherPath}"`);
      return { success: true };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  startSessionWatcher(accountId, expectedEmail) {
    // Stop any existing watcher
    this.stopSessionWatcher();

    this.watchingAccountId = accountId;

    // Check every 2 seconds for session tokens
    this.sessionWatcher = setInterval(() => {
      try {
        if (!fs.existsSync(this.launcherSettingsPath)) {
          return;
        }

        const launcherSettings = JSON.parse(fs.readFileSync(this.launcherSettingsPath, 'utf8'));

        // Check if user logged in with correct email and has session tokens
        if (launcherSettings.login === expectedEmail &&
            launcherSettings.at &&
            launcherSettings.rt) {

          console.log('✅ Session detected! Auto-saving...');

          // Capture the session
          const result = this.captureLauncherSession(accountId);

          if (result.success) {
            console.log('✅ Session saved successfully!');
            this.stopSessionWatcher();

            // Notify renderer to refresh account list
            if (this.mainWindow && !this.mainWindow.isDestroyed()) {
              this.mainWindow.webContents.send('session-captured', accountId);
            }
          }
        }
      } catch (error) {
        // File might not exist yet or be invalid JSON
        console.log('Waiting for valid session...');
      }
    }, 2000); // Check every 2 seconds

    // Stop watching after 5 minutes (timeout)
    setTimeout(() => {
      if (this.watchingAccountId === accountId) {
        console.log('⏱️ Session watch timeout - stopping');
        this.stopSessionWatcher();
      }
    }, 300000); // 5 minutes
  }

  stopSessionWatcher() {
    if (this.sessionWatcher) {
      clearInterval(this.sessionWatcher);
      this.sessionWatcher = null;
      this.watchingAccountId = null;
    }
  }

  async saveCurrentAccountSession() {
    try {
      const launcherSettingsPath = path.join(process.env.APPDATA, 'Battlestate Games', 'BsgLauncher', 'settings');

      if (!fs.existsSync(launcherSettingsPath)) {
        return; // No settings file, nothing to save
      }

      const launcherSettings = JSON.parse(fs.readFileSync(launcherSettingsPath, 'utf8'));

      // Check if there's a logged in user with valid tokens
      if (!launcherSettings.login || !launcherSettings.at || !launcherSettings.rt) {
        return; // No active session
      }

      // Find which of our accounts matches this email
      const accounts = this.getAccounts();
      const currentAccount = accounts.find(acc => acc.email === launcherSettings.login);

      if (!currentAccount) {
        return; // Not one of our accounts
      }

      // Save the current session (with potentially refreshed tokens)
      const sessionToSave = { ...launcherSettings };
      sessionToSave.tempFolder = this.tempFolder;

      currentAccount.launcherSession = sessionToSave;
      currentAccount.sessionCaptured = new Date().toISOString();

      fs.writeFileSync(this.accountsFile, JSON.stringify(accounts, null, 2));

      console.log(`✅ Saved current session for ${currentAccount.name} (${currentAccount.email}) before switching`);
    } catch (error) {
      console.error('Error saving current account session:', error);
      // Don't fail the switch if this fails
    }
  }

  async switchAccount(account) {
    try {
      // 0. BEFORE switching, save current account's session to capture any refreshed tokens
      await this.saveCurrentAccountSession();

      // 1. Kill launcher
      await this.killLauncher();

      // 2. Get account info
      const accounts = this.getAccounts();
      const fullAccount = accounts.find(acc => acc.id === account.id);

      if (!fullAccount) {
        return { success: false, error: 'Account not found' };
      }

      // 3. Check if we have a saved session
      if (fullAccount.launcherSession) {
        // Always restore session - if expired, user will see login screen
        await this.restoreLauncherSession(fullAccount.launcherSession);
        await this.startLauncher();

        // DO NOT start session watcher - session is already saved!
        // Watcher would just re-capture the same session and cause issues

        return {
          success: true,
          accountName: fullAccount.name,
          email: fullAccount.email,
          hasSession: true,
          message: 'Launcher gestartet - Auto-Login aktiv!'
        };
      } else {
        // No session saved - clear session and start fresh
        await this.updateLauncherAccount(fullAccount.email);
        await this.startLauncher();

        // Start session watcher
        setTimeout(() => {
          console.log('Starting session watcher for account switch...');
          this.startSessionWatcher(fullAccount.id, fullAccount.email);
        }, 2000);

        return {
          success: true,
          accountName: fullAccount.name,
          email: fullAccount.email,
          hasSession: false,
          message: 'Bitte einloggen - Session wird automatisch gespeichert!'
        };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  }
}

module.exports = AccountManager;
