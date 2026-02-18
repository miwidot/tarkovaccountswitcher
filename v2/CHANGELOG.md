# Changelog - Tarkov Account Switcher v2 (Go Rewrite)

## v2.0.0 (2026-02-19)

### Complete Rewrite
- **Rewritten from Electron to Go** with Fyne GUI framework
- Binary size reduced from ~120MB to ~20MB
- Native Windows application with embedded icon

### Features
- Account management (Add, Delete, Switch)
- Session capture and auto-login
- Streamer Mode (mask email addresses)
- System tray integration with minimize on launcher start
- Multi-language support (German/English)
- Dark theme with green accent (#00b894)

### Account Switching Improvements
- Save and restore `selectedGame` (EFT/Arena) per account
- Clear game cache on account switch to force fresh server data
- Preserve auth tokens while clearing UI cache

### Cache Clearing on Switch
Clears the following to ensure fresh data per account:
- `%TEMP%\Battlestate Games\EscapeFromTarkov\`
- `%TEMP%\Battlestate Games\EscapeFromTarkovArena\`
- `%LOCALAPPDATA%\Battlestate Games\BsgLauncher\CefCache\` (Cache, Local Storage, Session Storage)

### Technical
- Single instance lock (Windows Mutex)
- Proper app exit handling (no zombie processes)
- Windows resource embedding (icon, manifest, version info)
