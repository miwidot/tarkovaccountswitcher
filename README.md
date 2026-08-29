# Tarkov Account Switcher

**繁體中文** · [README.zh-TW.md](README.zh-TW.md)

![Version](https://img.shields.io/badge/version-2.0.8-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

Multi-account switcher for **Escape from Tarkov** with automatic session management and encrypted storage.

## Features

- **Automatic Session Management** — Saves login sessions encrypted locally
- **One-Click Account Switching** — Launcher restarts automatically with selected account
- **No Password Storage** — Only email + session tokens (AES-256 encrypted)
- **Auto-Login** — Automatically logs in after first login
- **Update Notifications** — Checks GitHub Releases on startup, shows banner with download link
- **Multi-Language** — English, German, Traditional Chinese, Simplified Chinese & Japanese with automatic system language detection
- **Streamer Mode** — Hides email addresses in the UI
- **System Tray Integration** — Runs in background, auto-minimizes when launcher starts
- **Single Instance** — Only one app instance can run
- **Themes** — Five UI themes including Escape from Tarkov–inspired layouts
- **Per-Account Launcher Options** — EFT or Arena and related launcher preferences per saved account

## Download

[Latest Release](https://github.com/miwidot/tarkovaccountswitcher/releases/latest)

## Quick Start

### Installation

1. Download `Tarkov Account Switcher.exe` from the latest release
2. Run the exe — no installation needed
3. Done! App runs in system tray

### Adding Your First Account

1. Open the **"Add"** tab
2. Enter **Account Name** + **Email** (e.g. "Main", "main@email.com")
3. Click **"Add Account & Start Launcher"**
4. Launcher starts automatically
5. **Log in normally in the launcher**
6. Session is **automatically detected and saved**
7. Account now shows green checkmark

### Switching Accounts

1. Open the **"Accounts"** tab
2. Click **"Switch"** on the desired account
3. Launcher restarts automatically — **already logged in**

## Security & BSG Statement

### What this tool does:

- Reads session tokens from BSG Launcher settings (`%APPDATA%\Battlestate Games\BsgLauncher\settings`)
- Stores them encrypted (AES-256-CBC) locally in `%APPDATA%\TarkovAccountSwitcher\accounts.json`
- On switch: Kills launcher process, replaces session data in launcher settings, restarts launcher
- **No passwords are stored** — only email addresses and session tokens

### What this tool does NOT do:

- **No game file modification** — No EFT or Arena files are read, written or patched
- **No code injection** — No DLL injection, no memory manipulation, no hooking
- **No anti-cheat interaction** — Does not touch BattlEye or any anti-cheat component
- **No network manipulation** — No traffic interception, no proxy, no MITM
- **No BSG server communication** — The tool never contacts BSG servers directly
- **No cloud sync** — All data stays local on your machine

The tool exclusively operates on the **BSG Launcher's local settings file** to swap session tokens between accounts. This is comparable to manually copying and pasting the settings file.

### Privacy

- All data stays **local on your PC**
- AES-256-CBC encryption with a unique key per installation
- No telemetry, no analytics, no network calls (except the GitHub update check)

## Disclaimer

**This tool does not modify any game files and performs no code injection.** It only reads and writes the BSG Launcher's local settings file to manage session tokens.

- **Current assessment**: Minimal risk — similar to other account switchers (e.g. TcNo Account Switcher)
- **No guarantee**: Use at **your own risk**. If BSG changes their Terms of Service, the assessment may change.

**Recommendations:**

- Enable 2FA on your BSG account
- Use different passwords for different accounts
- Never share credentials with third parties

## Tech Stack

- **Go** — Native Windows application backend
- **Wails v2** — Desktop shell with embedded **WebView2** (UI is bundled HTML/CSS/JS, no separate browser install beyond the Windows WebView2 runtime)
- **Vanilla frontend** — Single-page UI under `v2/frontend/dist/` — no Electron, no npm bundler for releases
- **AES-256-CBC** — Encryption via Go stdlib
- **Windows API** — Process management, native system tray (`Shell_NotifyIconW`)

## Building from Source

Active code lives under **`v2/`** (legacy Walk/Electron stacks in the repo root are historical only).

**Prerequisites**

- Go **1.23+** (see `v2/go.mod`; toolchain may pin a newer patch)
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **MinGW-w64 GCC** on `PATH` (Wails/CGO requirement on Windows)

**Commands** (run from **`v2/`**):

```bash
cd v2
go mod tidy
wails build -platform windows/amd64
```

Development with hot reload: `wails dev`. Release output: `v2/build/bin/Tarkov Account Switcher.exe`. Before a release build, sync the PE/App version from the updater constant: `go run sync_version.go` (see comments in `v2/sync_version.go`).

Developer-oriented layout, themes, and package map: **`v2/README.md`**.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

**Made with care for the Tarkov community**
