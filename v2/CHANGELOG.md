# Changelog - Tarkov Account Switcher v2

## v2.0.3 (2026-02-22)

- Fix: v2.0.2 release exe was compiled with v2.0.1 version string, causing false update notification

---

## v2.0.2 (2026-02-22)

### UI Framework Migration
- **Migrated from lxn/walk to Wails v2** (WebView2-based frontend)
- Fully themed HTML/CSS/JS frontend replaces native Windows controls
- All buttons, dropdowns, tabs, and cards are now fully styled (no more native Windows look)
- Binary size: ~13 MB

### 5 Premium Themes
- **Escape from Tarkov** — Authentic in-game menu style with khaki/gold accents, Oswald military font, noise grain overlay, scanline effect, vignette, background image, chamfered card corners, gold glow effects, diamond dividers
- **Killa Girl** — Anime Killa aesthetic with neon pink/magenta accents, 3-stripe diagonal overlay (Killa's signature), purple vignette, neon glow on buttons/cards, background image, industrial dark palette
- **Dark** — Clean dark theme with blue accents
- **Light** — Clean light theme with blue accents
- **Cappuccino** — Warm brown tones

### System Tray
- Custom Win32 API tray implementation (`Shell_NotifyIconW`) on dedicated goroutine
- Right-click context menu with Open/Quit (language-aware)
- Double-click to show window

### Known Limitations Resolved (from Walk)
- ~~Buttons and tab headers remain native Windows style~~ — Now fully styled via CSS
- ~~ComboBox dropdowns may appear light in dark mode~~ — Now styled select elements

---

## v2.0.1 (2026-02-22)

### Theme System (Walk)
- Added 3 themes: Light, Dark, Cappuccino
- Auto-detect Windows dark mode via registry (`AppsUseLightTheme`)
- DWM dark titlebar via `DwmSetWindowAttribute`
- Theme selection in Settings tab

### Dark Mode Fixes (Walk)
- Undocumented uxtheme.dll APIs (ordinals 132/133/135/136) for system dark mode
- `SetPreferredAppMode`, `AllowDarkModeForWindow` for per-control dark mode
- `SetWindowTheme("DarkMode_Explorer")` for tabs, `"DarkMode_CFD"` for ComboBox/LineEdit
- Owner-draw button painting via WndProc subclassing (custom GDI dark buttons)

### Single Instance Show
- Re-launching the exe now shows the existing window from tray instead of silently exiting
- Implemented via Named Windows Events (`CreateEventW`/`SetEvent`)

---

## v2.0.0 (2026-02-19)

### Complete Rewrite
- **Rewritten from Electron to Go** with lxn/walk native Windows GUI
- Binary size reduced from ~120MB to ~20MB
- Native Windows application with embedded icon

### Features
- Account management (Add, Delete, Switch)
- Session capture and auto-login (2s polling, 5min timeout)
- AES-256-CBC encryption (random IV, PKCS7 padding)
- Streamer Mode (mask email addresses)
- System tray integration with minimize on launcher start
- Multi-language support (German/English with system detection)
- Single instance lock (Windows Mutex)
- Update notifications via GitHub Releases API
- Per-account `selectedGame` (EFT/Arena) save/restore
- Per-account `EnvironmentUiType` (ingame background) save/restore
- Cache clearing on account switch

### Technical
- Go with lxn/walk GUI framework
- Windows resource embedding (icon, manifest, version info)
- Dark theme with green accent (#00b894)
