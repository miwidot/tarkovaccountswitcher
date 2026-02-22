# Tarkov Account Switcher v2 (Go + Wails)

Complete rewrite with **Wails v2** (WebView2) frontend and **Escape from Tarkov** in-game menu themed UI.

**Result:** ~13 MB native Windows app with 5 premium themes.

## Prerequisites

### 1. Go 1.21+
Download from: https://go.dev/dl/

### 2. Wails CLI v2
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 3. MinGW-w64 (GCC for Windows)
Wails requires CGO which needs a C compiler.

Download MinGW-w64:
- https://github.com/niXman/mingw-builds-binaries/releases
- Choose: `x86_64-13.2.0-release-posix-seh-ucrt-rt_v11-rev1.7z` (or similar)

Installation:
1. Extract to `C:\mingw64`
2. Add `C:\mingw64\bin` to your PATH environment variable
3. Verify: Open new terminal, run `gcc --version`

## Build

### Production Build
```bash
wails build -platform windows/amd64
```
Output: `build/bin/Tarkov Account Switcher.exe`

### Development Mode (Hot Reload)
```bash
wails dev
```

### Resolve Dependencies
```bash
go mod tidy
```

## Architecture

```
Go Backend (main.go + app.go)
│  Wails entry point, window management, system tray
│  Method bindings exposed to frontend via Wails IPC
│
├── Wails IPC (auto-generated bindings)
│
Frontend (frontend/dist/)
│  Single-page 3-tab app (vanilla HTML/CSS/JS)
│  5 themes via CSS custom properties
│  Oswald font (military/stencil aesthetic)
│
Internal Packages (internal/)
   accounts/  — CRUD, AES-256-CBC encryption, session watcher
   config/    — Settings, paths, email masking
   launcher/  — Process control (kill/start), settings read/write
   i18n/      — German/English translations
   updater/   — GitHub release checker
   singleinstance/ — Windows mutex
```

## Project Structure

```
v2/
├── main.go                       # Wails entry point, window config, asset embedding
├── app.go                        # Go binding layer (all methods callable from JS)
├── tray_windows.go               # Native Win32 system tray + window icon setter
├── frontend/
│   └── dist/
│       ├── index.html            # UI layout (3 tabs: Accounts, Add, Settings)
│       ├── style.css             # Theme system (5 themes) + premium overrides
│       ├── app.js                # Frontend logic (vanilla JS, Wails bindings)
│       ├── fonts/
│       │   └── oswald-regular.woff2
│       └── images/
│           ├── bg-eft.jpg        # EFT theme background image
│           └── bg-killa.jpg      # Killa Girl theme background image
├── internal/
│   ├── accounts/
│   │   ├── manager.go            # Account CRUD, session management
│   │   ├── encryption.go         # AES-256-CBC encryption (random IV, PKCS7)
│   │   └── watcher.go            # Session polling (2s interval, 5min timeout)
│   ├── launcher/
│   │   ├── control.go            # Kill/Start BSG Launcher, cache clearing
│   │   └── settings.go           # Launcher settings read/write, Game.ini
│   ├── config/
│   │   └── settings.go           # App settings, paths, email masking
│   ├── singleinstance/
│   │   └── mutex.go              # Windows Mutex for single instance
│   ├── i18n/
│   │   └── translations.go       # DE/EN translations (50+ keys)
│   └── updater/
│       └── updater.go            # GitHub API release checker
├── assets/
│   ├── icon.ico                  # App icon (embedded via //go:embed)
│   └── icon.png                  # Source icon (1024x1024)
├── build/
│   ├── appicon.png               # Wails build icon (PE resource embedding)
│   └── windows/
│       ├── icon.ico              # Windows icon resource
│       ├── info.json             # Version info for PE resource
│       └── wails.exe.manifest    # Windows manifest
├── winres/
│   ├── winres.json               # Windows resource config
│   ├── icon.ico                  # Icon for resource embedding
│   └── TarkovAccountSwitcher.manifest
├── wails.json                    # Wails project config
├── go.mod
└── go.sum
```

## Features

- **Account Management** — Add, delete, switch accounts with one click
- **Session Capture** — Automatic 2-second polling with 5-minute timeout
- **AES-256-CBC Encryption** — Random IV, PKCS7 padding, unique key per install
- **System Tray** — Native Win32 tray (custom implementation, no library conflicts)
- **Single Instance Lock** — Wails built-in + Windows Mutex fallback
- **Multi-Language** — German/English with system language detection
- **Streamer Mode** — Email masking (`t***@e***.com`)
- **Update Checker** — GitHub API release polling
- **Launcher Control** — taskkill/start for BSG Launcher
- **Cache Clearing** — Temp, CefCache, Arena cache cleared on switch
- **Per-Account Settings** — `selectedGame` (EFT/Arena) + `EnvironmentUiType` (ingame background)

## Themes

5 themes via CSS custom properties (`[data-theme="..."]`):

| Theme | Style | Accent | Special Effects |
|-------|-------|--------|-----------------|
| **Escape from Tarkov** | Dark military | Khaki/gold `#c8aa6e` | Noise grain, scanlines, vignette, background image, chamfered corners, gold glow |
| **Killa Girl** | Dark industrial | Neon pink `#e84393` | 3-stripe overlay, neon glow, purple vignette, background image, chamfered corners |
| **Dark** | Clean dark | Blue `#60a0ff` | Standard flat design |
| **Light** | Clean light | Blue `#0078d4` | Standard flat design |
| **Cappuccino** | Warm brown | Brown `#8c501e` | Standard flat design |

### Adding a New Theme

1. Add CSS custom properties block in `style.css` (copy from existing theme)
2. Add premium overrides section if theme has special effects
3. Add `<option>` to theme dropdown in `index.html`
4. Optional: Add background image to `frontend/dist/images/`

## Data Location

`%APPDATA%\TarkovAccountSwitcher\`
- `accounts.json` — Encrypted accounts + sessions
- `settings.json` — App settings (language, theme, launcher path, streamer mode)
- `.key` — AES-256 encryption key (mode 0600)

## Key Implementation Details

- **Wails v2** with `HideWindowOnClose: true` (close = minimize to tray)
- **Frontend is vanilla HTML/CSS/JS** — no npm, no bundler, no framework
- **Assets embedded** via `//go:embed all:frontend/dist` into the Go binary
- **System tray** uses custom Win32 API (`Shell_NotifyIconW`) on separate goroutine
- **Go methods** exposed to JS via Wails auto-binding (`window.go.main.App.*`)
- **Events** from Go to JS via `runtime.EventsEmit` (session-captured, update-available)
- **Windows-specific** — Uses `taskkill` for process control, Windows Mutex
- **No code signing** — Windows SmartScreen warnings expected on first run
