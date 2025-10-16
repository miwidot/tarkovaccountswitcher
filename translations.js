const translations = {
  de: {
    // Tabs
    tabAccounts: '📋 Accounts',
    tabAdd: '➕ Hinzufügen',
    tabSettings: '⚙️ Einstellungen',

    // Add Account Tab
    addAccountTitle: 'Neuen Account hinzufügen',
    labelAccountName: 'Account Name (z.B. "Main", "Alt")',
    placeholderAccountName: 'Main Account',
    labelEmail: 'Email',
    placeholderEmail: 'your@email.com',
    addAccountHelp: 'Nach dem Hinzufügen startet der Launcher automatisch.<br>Logge dich ein - die Session wird automatisch gespeichert! ✅',
    btnAddAccount: 'Account hinzufügen & Launcher starten',

    // Account List
    emptyStateTitle: 'Noch keine Accounts gespeichert',
    emptyStateSubtitle: 'Füge oben deinen ersten Account hinzu',
    statusAutoLogin: 'Auto-Login aktiv',
    statusLoginRequired: 'Login erforderlich - wird automatisch gespeichert',
    btnSwitch: 'Wechseln',
    btnDelete: 'Löschen',
    confirmDelete: 'Account wirklich löschen?',

    // Settings Tab
    settingsTitle: 'Einstellungen',
    labelLanguage: 'Sprache / Language',
    labelLauncherPath: 'BSG Launcher Pfad',
    placeholderLauncherPath: 'C:\\Battlestate Games\\BsgLauncher\\BsgLauncher.exe',
    btnBrowse: 'Durchsuchen...',
    btnSave: 'Speichern',

    // Status Messages
    statusFillFields: 'Bitte fülle alle Felder aus',
    statusAccountAdded: '✅ Account hinzugefügt!\n\nLauncher startet jetzt...\nBitte einloggen - Session wird automatisch gespeichert!',
    statusAccountDeleted: 'Account gelöscht',
    statusDeleteError: 'Fehler beim Löschen',
    statusLauncherRestarting: 'Launcher wird neu gestartet...',
    statusAutoLoginActive: '🚀 AUTO-LOGIN AKTIV!\n\nAccount: {name}\nLauncher startet automatisch eingeloggt!',
    statusManualLogin: '⚠️ MANUELLES LOGIN\n\nAccount: {name}\nEmail: {email}\n\nBitte einloggen - Session wird automatisch gespeichert!',
    statusError: 'Fehler: {error}',
    statusPathSaved: 'Launcher Pfad gespeichert!',
    statusSaveError: 'Fehler beim Speichern',
    statusEnterPath: 'Bitte gib einen Pfad ein',
    statusLanguageSaved: 'Sprache gespeichert!',
  },

  en: {
    // Tabs
    tabAccounts: '📋 Accounts',
    tabAdd: '➕ Add',
    tabSettings: '⚙️ Settings',

    // Add Account Tab
    addAccountTitle: 'Add New Account',
    labelAccountName: 'Account Name (e.g. "Main", "Alt")',
    placeholderAccountName: 'Main Account',
    labelEmail: 'Email',
    placeholderEmail: 'your@email.com',
    addAccountHelp: 'After adding, the launcher will start automatically.<br>Log in - the session will be saved automatically! ✅',
    btnAddAccount: 'Add Account & Start Launcher',

    // Account List
    emptyStateTitle: 'No accounts saved yet',
    emptyStateSubtitle: 'Add your first account above',
    statusAutoLogin: 'Auto-login active',
    statusLoginRequired: 'Login required - will be saved automatically',
    btnSwitch: 'Switch',
    btnDelete: 'Delete',
    confirmDelete: 'Really delete account?',

    // Settings Tab
    settingsTitle: 'Settings',
    labelLanguage: 'Language / Sprache',
    labelLauncherPath: 'BSG Launcher Path',
    placeholderLauncherPath: 'C:\\Battlestate Games\\BsgLauncher\\BsgLauncher.exe',
    btnBrowse: 'Browse...',
    btnSave: 'Save',

    // Status Messages
    statusFillFields: 'Please fill all fields',
    statusAccountAdded: '✅ Account added!\n\nLauncher starting...\nPlease login - session will be saved automatically!',
    statusAccountDeleted: 'Account deleted',
    statusDeleteError: 'Error deleting',
    statusLauncherRestarting: 'Restarting launcher...',
    statusAutoLoginActive: '🚀 AUTO-LOGIN ACTIVE!\n\nAccount: {name}\nLauncher starting automatically logged in!',
    statusManualLogin: '⚠️ MANUAL LOGIN\n\nAccount: {name}\nEmail: {email}\n\nPlease login - session will be saved automatically!',
    statusError: 'Error: {error}',
    statusPathSaved: 'Launcher path saved!',
    statusSaveError: 'Error saving',
    statusEnterPath: 'Please enter a path',
    statusLanguageSaved: 'Language saved!',
  }
};

module.exports = translations;
