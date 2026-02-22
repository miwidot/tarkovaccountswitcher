# Changelog - Tarkov Account Switcher v2

## v2.0.0 (2026-02-22)

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
- Replaces energye/systray library (which caused thread conflicts with Wails)
- Right-click context menu with Open/Quit (language-aware: Oeffnen/Beenden)
- Double-click to show window
- Proper cleanup on app exit

### Application Icon
- Icon properly embedded in PE resources via `wails build`
- Shows correctly in Windows Explorer, taskbar, and Task Manager
- System tray uses embedded icon via `//go:embed`

### All v1 Features Preserved
- Account management (Add, Delete, Switch)
- Automatic session capture (2s polling, 5min timeout)
- One-click account switching with launcher restart
- AES-256-CBC encryption (random IV, PKCS7 padding)
- Update notifications via GitHub Releases API
- Streamer Mode (email masking)
- Multi-language (German/English with system detection)
- Single instance lock (Wails built-in + Windows Mutex)
- Per-account `selectedGame` (EFT/Arena) save/restore
- Per-account `EnvironmentUiType` (ingame background) save/restore
- Cache clearing on account switch

### Cache Clearing on Switch
Clears the following to ensure fresh data per account:
- `%TEMP%\Battlestate Games\EscapeFromTarkov\`
- `%TEMP%\Battlestate Games\EscapeFromTarkovArena\`
- `%LOCALAPPDATA%\Battlestate Games\BsgLauncher\CefCache\` (Cache, Local Storage, Session Storage)

### Technical
- Wails v2.11.0 with WebView2 frontend
- `HideWindowOnClose: true` for minimize-to-tray behavior
- Assets embedded via `//go:embed all:frontend/dist`
- Frontend: vanilla HTML/CSS/JS (no npm, no bundler, no framework)
- CSS custom properties for theme system (`[data-theme="..."]`)
- Window drag via `--wails-draggable: drag` on header
- Events from Go to JS: `session-captured`, `update-available`

### Known Limitations Resolved (from v1)
- ~~Buttons and tab headers remain native Windows style~~ — Now fully styled via CSS
- ~~ComboBox dropdowns may appear light in dark mode~~ — Now styled select elements
