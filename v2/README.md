# Tarkov Account Switcher v2 (Go)

Complete rewrite of the Electron app in Go with Fyne GUI framework.

**Target:** 100% Feature parity, ~10 MB instead of 120 MB

## Prerequisites

### 1. Go 1.21+
Download from: https://go.dev/dl/

### 2. MinGW-w64 (GCC for Windows)
Fyne requires CGO which needs a C compiler.

Download MinGW-w64:
- https://github.com/niXman/mingw-builds-binaries/releases
- Choose: `x86_64-13.2.0-release-posix-seh-ucrt-rt_v11-rev1.7z` (or similar)

Installation:
1. Extract to `C:\mingw64`
2. Add `C:\mingw64\bin` to your PATH environment variable
3. Verify: Open new terminal, run `gcc --version`

## Build

### Option 1: PowerShell
```powershell
.\build.ps1
```

### Option 2: Command Prompt
```batch
build.bat
```

### Option 3: Manual
```bash
set CGO_ENABLED=1
go build -ldflags="-H windowsgui -s -w" -o "Tarkov Account Switcher.exe"
```

Expected output size: **~8-12 MB**

## Project Structure

```
v2/
├── main.go                 # Entry point
├── internal/
│   ├── accounts/
│   │   ├── manager.go      # Account CRUD, Session Management
│   │   ├── encryption.go   # AES-256-CBC Encryption
│   │   └── watcher.go      # Session Polling (2s interval)
│   ├── launcher/
│   │   ├── control.go      # Kill/Start BSG Launcher
│   │   └── settings.go     # Launcher Settings read/write
│   ├── config/
│   │   └── settings.go     # App Settings (language, paths)
│   ├── singleinstance/
│   │   └── mutex.go        # Windows Mutex for single instance
│   └── i18n/
│       └── translations.go # DE/EN Translations
├── ui/
│   ├── window.go           # Main window
│   ├── theme.go            # Dark theme with green accent
│   ├── accounts_tab.go     # Accounts list UI
│   ├── add_tab.go          # Add account form
│   ├── settings_tab.go     # Settings UI
│   └── tray.go             # System tray
├── assets/
│   ├── icon.png
│   └── icon.ico
├── go.mod
└── go.sum
```

## Features

- **Single Instance Lock** - Only one instance can run
- **System Tray** - Minimize to tray, quick access menu
- **Dark Theme** - Green accent (#00b894)
- **i18n** - German and English
- **Session Polling** - 2 second interval, 5 minute timeout
- **AES-256-CBC Encryption** - Compatible with v1 data format
- **Launcher Control** - taskkill/start for BSG Launcher

## Data Location

Same as v1: `%APPDATA%\TarkovAccountSwitcher\`
- `accounts.json` - Encrypted accounts + sessions
- `settings.json` - App settings
- `.key` - AES-256 encryption key
