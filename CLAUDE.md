# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Tarkov Account Switcher is a Windows-only Electron desktop application for managing multiple Escape from Tarkov accounts. It automates session token management to enable one-click account switching without storing passwords.

## Build Commands

```bash
npm start        # Run in development mode
npm run build    # Build release package (outputs to release/)
```

No test or lint scripts are configured.

## Architecture

This is a classic Electron IPC architecture with three main modules:

```
┌─────────────────────────────────────────┐
│      MAIN PROCESS (main.js)             │
│  - Window & tray management             │
│  - Single instance lock                 │
│  - IPC handlers for all account ops     │
└────────────┬────────────────────────────┘
             │ IPC Bridge
┌────────────┴────────────────────────────┐
│    RENDERER PROCESS (renderer.js)       │
│  - UI tab system (Accounts/Add/Settings)│
│  - Form handling & status messages      │
│  - Translation management               │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│  ACCOUNT MANAGER (accountManager.js)    │
│  - Encryption (AES-256-CBC)             │
│  - Session capture via polling          │
│  - Launcher process control             │
│  - Data persistence to %APPDATA%        │
└─────────────────────────────────────────┘
```

**Key files:**
- `main.js` - Entry point, window/tray creation, IPC handlers
- `renderer.js` - Frontend UI logic
- `accountManager.js` - Core business logic (encryption, session management, process control)
- `translations.js` - i18n strings (German/English)
- `index.html` - UI layout

**Data storage location:** `%APPDATA%\TarkovAccountSwitcher\`
- `accounts.json` - Encrypted accounts + sessions
- `settings.json` - App settings
- `.key` - AES-256 encryption key

**Launcher settings monitored:** `%APPDATA%\Battlestate Games\BsgLauncher\settings`

## Important Implementation Details

- **nodeIntegration: true** and **contextIsolation: false** - Renderer has direct Node.js access
- **Polling-based session detection** - 2-second interval polling (not file watchers), 5-minute timeout
- **Windows-specific** - Uses `taskkill` and `exec()` for process control
- **No code signing** - Windows SmartScreen warnings expected on install
