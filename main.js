const { app, BrowserWindow, ipcMain, dialog, Tray, Menu, nativeImage } = require('electron');
const path = require('path');
const AccountManager = require('./accountManager');
const { version } = require('./package.json');

// Set process title for Task Manager
process.title = `Tarkov Account Switcher v${version}`;

let mainWindow;
let tray = null;
const accountManager = new AccountManager();

// Single Instance Lock - only allow one instance of the app
const gotTheLock = app.requestSingleInstanceLock();

if (!gotTheLock) {
  // Another instance is already running, quit this one
  app.quit();
} else {
  // If someone tries to run a second instance, focus the existing window
  app.on('second-instance', (event, commandLine, workingDirectory) => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      if (!mainWindow.isVisible()) mainWindow.show();
      mainWindow.focus();
    }
  });
}

function createTray() {
  // Load custom icon
  const iconPath = path.join(__dirname, 'icon.png');
  const icon = nativeImage.createFromPath(iconPath).resize({ width: 16, height: 16 });
  tray = new Tray(icon);

  const contextMenu = Menu.buildFromTemplate([
    {
      label: 'Öffnen',
      click: () => {
        mainWindow.show();
        mainWindow.focus();
      }
    },
    {
      label: 'Beenden',
      click: () => {
        app.quit();
      }
    }
  ]);

  tray.setToolTip(`Tarkov Account Switcher v${version}`);
  tray.setContextMenu(contextMenu);

  // Single click to show window
  tray.on('click', () => {
    if (mainWindow.isVisible()) {
      mainWindow.hide();
    } else {
      mainWindow.show();
      mainWindow.focus();
    }
  });
}

function createWindow() {
  const iconPath = path.join(__dirname, 'icon.png');

  mainWindow = new BrowserWindow({
    width: 850,
    height: 800,
    minWidth: 800,
    minHeight: 700,
    title: `Tarkov Account Switcher v${version}`,
    icon: iconPath,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false
    },
    autoHideMenuBar: true,
    resizable: true
  });

  mainWindow.loadFile('index.html');

  // Give AccountManager access to mainWindow for sending events
  accountManager.setMainWindow(mainWindow);

  // Set process name for Task Manager
  if (process.platform === 'win32') {
    app.setAppUserModelId(`Tarkov Account Switcher v${version}`);
  }

  // Prevent window from closing, just hide it
  mainWindow.on('close', (event) => {
    if (!app.isQuitting) {
      event.preventDefault();
      mainWindow.hide();
    }
  });

  // Create tray icon
  createTray();
}

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  // Don't quit on window close - we're using tray
  // App will quit via tray menu or app.quit()
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});

app.on('before-quit', () => {
  app.isQuitting = true;

  // Destroy tray icon to prevent it from staying in system tray
  if (tray) {
    tray.destroy();
    tray = null;
  }
});

// IPC Handlers
ipcMain.handle('get-accounts', async () => {
  return accountManager.getAccounts();
});

ipcMain.handle('add-account', async (event, account) => {
  return await accountManager.addAccount(account);
});

ipcMain.handle('delete-account', async (event, id) => {
  return accountManager.deleteAccount(id);
});

ipcMain.handle('switch-account', async (event, account) => {
  const result = await accountManager.switchAccount(account);

  // Hide window to tray after starting launcher
  if (result.success) {
    setTimeout(() => {
      mainWindow.hide();
    }, 2000); // Wait 2 seconds so user can see the status message
  }

  return result;
});

ipcMain.handle('get-settings', async () => {
  const settings = accountManager.getSettings();

  // Auto-detect system language on first run
  if (!settings.language) {
    const systemLocale = app.getLocale(); // e.g. "de", "en-US", "en-GB"
    settings.language = systemLocale.startsWith('de') ? 'de' : 'en';
    accountManager.setLanguage(settings.language);
  }

  return settings;
});

ipcMain.handle('set-language', async (event, language) => {
  return accountManager.setLanguage(language);
});

ipcMain.handle('set-launcher-path', async (event, launcherPath) => {
  return accountManager.setLauncherPath(launcherPath);
});

ipcMain.handle('browse-launcher-path', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openFile'],
    filters: [
      { name: 'Executable', extensions: ['exe'] }
    ]
  });

  if (!result.canceled && result.filePaths.length > 0) {
    return { filePath: result.filePaths[0] };
  }

  return { filePath: null };
});

ipcMain.handle('capture-session', async (event, accountId) => {
  return accountManager.captureLauncherSession(accountId);
});
