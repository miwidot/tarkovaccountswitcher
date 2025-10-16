# 🎮 Tarkov Account Switcher

![Version](https://img.shields.io/badge/version-1.3.1-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

Multi-account switcher for **Escape from Tarkov** with automatic session management and encrypted storage.

## ✨ Features

- ✅ **Automatic Session Management**: Saves login sessions encrypted locally
- ✅ **One-Click Account Switching**: Launcher restarts automatically with selected account
- ✅ **No Password Storage**: Only email + session tokens (AES-256 encrypted)
- ✅ **Auto-Login**: Automatically logs in after first login
- ✅ **Multi-Language**: German & English with automatic system language detection
- ✅ **System Tray Integration**: Runs in background, auto-minimizes
- ✅ **Single Instance**: Only one app instance can run

## 📥 Download

**Latest Release: v1.3.1**

📦 [Tarkov-Account-Switcher-v1.3.1.zip](https://github.com/miwidot/tarkov-account-switcher/releases/latest) (~108 MB)

## 🚀 Quick Start

### Installation

1. Download and extract `Tarkov-Account-Switcher-v1.3.1.zip`
2. Run `Tarkov Account Switcher.exe`
3. Done! App runs in system tray

### Adding Your First Account

1. Open **"Add"** tab
2. Enter **Account Name** + **Email** (e.g., "Main", "main@email.com")
3. Click **"Add Account & Start Launcher"**
4. Launcher starts automatically
5. **Log in normally in the launcher**
6. Session is **automatically detected and saved** ✅
7. Account now shows **green checkmark** ✅

### Switching Accounts

1. Open **"Accounts"** tab
2. Select account
3. Click **"Switch"**
4. Launcher starts automatically **already logged in**! 🚀

## 🔒 Security & Privacy

### What the Tool Does:
- ✅ Reads session tokens from BSG Launcher Settings (`%APPDATA%\Battlestate Games\BsgLauncher\settings`)
- ✅ Stores them encrypted (AES-256) locally in `%APPDATA%\TarkovAccountSwitcher\accounts.json`
- ✅ On switch: Kill launcher → Replace session data → Restart launcher
- ✅ **No password stored** - only email + session tokens

### What the Tool Does NOT Do:
- ❌ No game file modification
- ❌ No injection/patching
- ❌ No cloud synchronization
- ❌ No network manipulation

### Privacy:
- 🔐 All data stays **local on your PC**
- 🔐 AES-256-CBC encryption
- 🔐 Unique encryption key per installation
- 🔐 No telemetry, no analytics

## ⚠️ Disclaimer

**Important - Please Read:**

- This tool does **not modify game files** and performs **no code injection**
- It only works with launcher session data (similar to TcNo Account Switcher)
- **Current assessment**: Minimal risk
- **BUT**: I give **no guarantee**. Use at **your own risk**!
- If BSG changes their TOS in the future, the assessment may change

**Recommendations:**
- ✅ Enable 2FA on your BSG account
- ✅ Backup important files before first use
- ✅ Never share credentials with third parties
- ✅ Use different passwords for different accounts

## 🛠️ Tech Stack

- **Electron** - Cross-platform desktop framework
- **Node.js** - Backend runtime
- **AES-256-CBC** - Encryption
- **Windows Process Management** - Launcher control

## 📋 Changelog

### v1.3.1 (Current)
- 🌍 **Multi-Language Support**: German & English with automatic system language detection
- 🐛 **Session Token Fix**: Tokens are now correctly deleted on account switch (prevents false session storage)
- 🐛 **Path Merge Fix**: System-specific paths are not overwritten on session restore
- ✅ Improved session management for more stable account switching

### v1.3.0
- ✅ ASAR packaging for cleaner file structure
- ✅ Session watcher optimizations

### v1.2.0
- ✅ Fully automatic session detection
- ✅ No password storage (only session tokens)
- ✅ System Tray integration
- ✅ Single Instance Lock
- ✅ Tab-based UI (Accounts / Add / Settings)

[View full changelog →](./dc.md)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Thanks to all beta testers!
- Inspired by TcNo Account Switcher

## ⚠️ Support

**Beta Version Notice:**
This is a beta version (v1.3.1). If you encounter issues or have feedback, please open an issue on GitHub!

---

**Made with ❤️ for the Tarkov community**
