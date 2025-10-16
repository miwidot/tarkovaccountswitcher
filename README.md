# 🎮 Tarkov Account Switcher

Multi-account switcher for **Escape from Tarkov** with automatic session management.

![Version](https://img.shields.io/badge/version-1.3.1-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

---

## ✨ Features

- **🔄 Automatic Session Management**: Saves login sessions encrypted locally
- **⚡ One-Click Account Switching**: Launcher restarts automatically with selected account
- **🔒 Secure**: Only email + session tokens stored (AES-256 encrypted), no passwords
- **🚀 Auto-Login**: After first login, future logins are automatic
- **🌍 Multi-Language**: German & English (auto-detects system language)
- **🖥️ System Tray Integration**: Runs in background, auto-minimizes after launcher starts
- **🎯 Single Instance Lock**: Only one app instance can run
- **⚙️ Custom Launcher Path**: Configure if your launcher is installed elsewhere

---

## 📥 Installation

### Requirements
- Windows 10/11
- BSG Launcher (not Steam version)
- Node.js (for development only)

### Download
📦 **[Latest Release (v1.3.1)](https://github.com/miwidot/tarkovaccountswitcher/releases/latest)** (~108 MB)

### Setup
1. Extract ZIP file to desired location (e.g., `C:\Program Files\TarkovAccountSwitcher\`)
2. Run `Tarkov Account Switcher.exe`
3. Done! App runs in system tray

---

## 🚀 Usage

### Adding First Account
1. Open **"Add" / "Hinzufügen"** tab
2. Enter **Account Name** + **Email** (e.g., "Main", "main@email.com")
3. Click **"Add Account & Start Launcher"**
4. Launcher starts automatically
5. **Login normally in the launcher**
6. Session is **automatically detected and saved** ✅
7. Account now shows **green checkmark** ✅

### Switching Accounts
1. Open **"Accounts"** tab
2. Select account
3. Click **"Switch" / "Wechseln"**
4. Launcher restarts **already logged in**! 🚀

### Changing Launcher Path (Optional)
If your launcher is installed elsewhere:
1. Open **"Settings" / "Einstellungen"** tab
2. Enter path or click **"Browse" / "Durchsuchen"**
3. Click **"Save" / "Speichern"**

### Changing Language (Optional)
App auto-detects system language (German/English). To change manually:
1. Open **"Settings" / "Einstellungen"** tab
2. Select **Language** (Deutsch / English)
3. UI updates immediately ✅

---

## 🔒 Security & Technical Details

### What the tool does:
- ✅ Reads session tokens from BSG Launcher settings (`%APPDATA%\Battlestate Games\BsgLauncher\settings`)
- ✅ Stores them encrypted (AES-256) locally in `%APPDATA%\TarkovAccountSwitcher\accounts.json`
- ✅ On switch: Kill launcher process → Replace session data → Restart launcher
- ✅ **No passwords stored** - only email + session tokens

### What the tool does NOT do:
- ❌ No modification of game files
- ❌ No injection/patching
- ❌ No cloud synchronization
- ❌ No network manipulation

### Privacy:
- 🔐 All data stays **local on your PC**
- 🔐 AES-256-CBC encryption
- 🔐 Unique encryption key per installation
- 🔐 No telemetry, no analytics

---

## ⚠️ Ban Risk / TOS

**Important - please read:**

- This tool does **not modify game files** and performs **no code injection**
- It only works with launcher session data (similar to TcNo Account Switcher)
- **Current assessment**: Minimal risk
- **BUT**: I give **no guarantees**. Use at **your own risk**!
- If BSG changes their TOS in the future, the assessment may change

**Recommendations:**
- ✅ Enable 2FA on your BSG account
- ✅ Backup important files before first use
- ✅ Never share your credentials with third parties
- ✅ Use different passwords for different accounts

---

## 🛠️ Development

### Tech Stack
- **Electron**: Desktop app framework
- **Node.js**: Backend runtime
- **JavaScript**: Main language

### Project Structure
```
tarkovaccountswitcher/
├── main.js              # Electron main process
├── renderer.js          # Frontend logic
├── accountManager.js    # Session management
├── translations.js      # Multi-language support
├── index.html           # UI
├── package.json         # Dependencies
└── icon.png            # App icon
```

### Building from Source
```bash
# Install dependencies
npm install

# Run in development
npm start

# Build for Windows
npx electron-packager . "Tarkov Account Switcher" \
  --platform=win32 --arch=x64 \
  --icon=icon.ico \
  --asar \
  --out=build
```

---

## 📋 Changelog

### v1.3.1 (Current)
- 🌍 **Multi-Language Support**: German & English with auto-detection
- 🐛 **Session Token Fix**: Tokens properly deleted on account switch (prevents false session storage)
- 🐛 **Path Merge Fix**: System-specific paths not overwritten during session restore
- ✅ Improved session management for more stable account switching

### v1.3.0
- ✅ ASAR packaging for cleaner file structure
- ✅ Session watcher optimizations

### v1.2.0
- ✅ Fully automatic session detection
- ✅ No password storage (only session tokens)
- ✅ System tray integration
- ✅ Single instance lock
- ✅ Tab-based UI (Accounts / Add / Settings)
- ✅ Launcher kill on account add (prevents old sessions)
- ✅ Email validation during session capture
- ✅ Custom icon support

---

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🙏 Credits

Created by [@miwidot](https://github.com/miwidot)

---

## ⚠️ Disclaimer

This is an unofficial tool and is not affiliated with, endorsed by, or connected to Battlestate Games Limited or Escape from Tarkov. Use at your own risk.

---

## 🐛 Issues & Feedback

Found a bug or have a feature request? Please [open an issue](https://github.com/miwidot/tarkovaccountswitcher/issues)!

---

**Happy Switching! 🎯**
