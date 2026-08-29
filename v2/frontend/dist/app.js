// ===================================================================
// Tarkov Account Switcher — Frontend Application Logic
// Vanilla JS. Communicates with Go via Wails runtime bindings.
// Supports 5 themes: eft, killa, dark, light, cappuccino
// ===================================================================

// Translation cache — loaded from Go on startup and language change
let T = {};

const STATUS_MESSAGE_DURATION_MS = 3000;

let statusClearTimers = {};

function scheduleStatusClear(elementId) {
    if (statusClearTimers[elementId]) {
        clearTimeout(statusClearTimers[elementId]);
    }
    statusClearTimers[elementId] = setTimeout(() => {
        delete statusClearTimers[elementId];
        const el = document.getElementById(elementId);
        if (el) {
            el.textContent = '';
            el.className = 'status-message';
        }
    }, STATUS_MESSAGE_DURATION_MS);
}

function cancelStatusClear(elementId) {
    if (statusClearTimers[elementId]) {
        clearTimeout(statusClearTimers[elementId]);
        delete statusClearTimers[elementId];
    }
}

function scheduleSettingsStatusClear() {
    scheduleStatusClear('settings-status');
}

function cancelSettingsStatusClear() {
    cancelStatusClear('settings-status');
}

function scheduleAccountsStatusClear() {
    scheduleStatusClear('accounts-status');
}

function scheduleAddStatusClear() {
    scheduleStatusClear('add-status');
}

// ======================== INITIALIZATION ========================

document.addEventListener('DOMContentLoaded', async () => {
    // Wait briefly for Wails runtime to be injected
    await waitForWails();

    await loadTranslations();
    await initTheme();
    applyTranslations();
    setupTabNavigation();
    setupEventListeners();
    setupWailsEvents();
    await loadAccountsTab();
    await loadSettingsValues();
});

function waitForWails() {
    return new Promise((resolve) => {
        if (window.go && window.go.main && window.go.main.App) {
            resolve();
            return;
        }
        // Poll until Wails runtime is ready
        const interval = setInterval(() => {
            if (window.go && window.go.main && window.go.main.App) {
                clearInterval(interval);
                resolve();
            }
        }, 50);
        // Safety timeout after 5 seconds
        setTimeout(() => {
            clearInterval(interval);
            resolve();
        }, 5000);
    });
}

// ======================== THEME ========================

async function initTheme() {
    try {
        const settings = await window.go.main.App.GetSettings();
        const theme = settings.theme || 'eft';
        applyTheme(theme);
    } catch (e) {
        applyTheme('eft');
    }
}

function applyTheme(themeId) {
    document.documentElement.setAttribute('data-theme', themeId);
    document.body.setAttribute('data-theme', themeId);
}

async function onThemeChange() {
    const themeId = document.getElementById('settings-theme-select').value;
    const statusEl = document.getElementById('settings-status');

    try {
        await window.go.main.App.SetTheme(themeId);
        applyTheme(themeId);

        statusEl.textContent = '\u2713 ' + t('statusThemeSaved');
        statusEl.className = 'status-message success';
        scheduleSettingsStatusClear();
    } catch (e) {
        console.error('Theme change failed:', e);
        statusEl.textContent = tf('statusError', { error: String(e) });
        statusEl.className = 'status-message error';
    }
}

// ======================== TRANSLATIONS ========================

async function loadTranslations() {
    try {
        T = await window.go.main.App.GetAllTranslations();
    } catch (e) {
        console.error('Failed to load translations:', e);
        T = {};
    }
}

function t(key) {
    return T[key] || key;
}

function tf(key, replacements) {
    let text = t(key);
    for (const [placeholder, value] of Object.entries(replacements)) {
        text = text.replaceAll('{' + placeholder + '}', value);
    }
    return text;
}

async function applyTranslations() {
    // Tab buttons
    setText('tab-btn-accounts', t('tabAccounts'));
    setText('tab-btn-add', t('tabAdd'));
    setText('tab-btn-settings', t('tabSettings'));

    // Add tab
    setText('add-title', t('addAccountTitle'));
    setText('add-name-label', t('labelAccountName'));
    setPlaceholder('add-name-input', t('placeholderAccountName'));
    setText('add-email-label', t('labelEmail'));
    setPlaceholder('add-email-input', t('placeholderEmail'));
    setText('add-help-text', t('addAccountHelp'));
    setText('add-submit-btn', t('btnAddAccount'));

    // Empty state
    setText('empty-title', t('emptyStateTitle'));
    setText('empty-subtitle', t('emptyStateSubtitle'));

    // Settings tab
    setText('settings-title', t('settingsTitle'));
    setText('settings-lang-label', t('labelLanguage'));
    setText('settings-theme-label', t('labelTheme'));
    setText('settings-path-label', t('labelLauncherPath'));
    setPlaceholder('settings-path-input', t('placeholderLauncherPath'));
    setText('settings-browse-btn', t('btnBrowse'));
    setText('settings-save-btn', t('btnSave'));
    setText('settings-autostart-label', t('labelAutoStart'));
    setText('settings-autostart-help', t('autoStartHelp'));
    setText('settings-streamer-label', t('labelStreamerMode'));
    setText('settings-streamer-help', t('streamerModeHelp'));
    setText('settings-quit-btn', t('btnQuit'));
    setText('footer-powered-prefix', t('footerPoweredBy'));

    applyThemeSelectLabels();

    document.title = t('htmlPageTitle');

    // Version
    try {
        const version = await window.go.main.App.GetVersion();
        setText('version-text', tf('aboutVersion', { version }));
    } catch (e) { /* ignore */ }
}

function applyThemeSelectLabels() {
    const sel = document.getElementById('settings-theme-select');
    if (!sel) return;
    const map = {
        eft: 'themeNameEft',
        killa: 'themeNameKilla',
        dark: 'themeNameDark',
        light: 'themeNameLight',
        cappuccino: 'themeNameCappuccino'
    };
    for (let i = 0; i < sel.options.length; i++) {
        const opt = sel.options[i];
        const key = map[opt.value];
        if (key) opt.textContent = t(key);
    }
}

function setText(id, text) {
    const el = document.getElementById(id);
    if (el) el.textContent = text;
}

function setPlaceholder(id, text) {
    const el = document.getElementById(id);
    if (el) el.placeholder = text;
}

// ======================== TAB NAVIGATION ========================

function setupTabNavigation() {
    const tabBtns = document.querySelectorAll('.tab-btn');
    const panels = document.querySelectorAll('.tab-panel');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const tabName = btn.dataset.tab;

            tabBtns.forEach(b => b.classList.remove('active'));
            panels.forEach(p => p.classList.remove('active'));

            btn.classList.add('active');
            document.getElementById('panel-' + tabName).classList.add('active');
        });
    });
}

function selectTab(tabName) {
    const btn = document.querySelector('.tab-btn[data-tab="' + tabName + '"]');
    if (btn) btn.click();
}

// ======================== ACCOUNTS TAB ========================

async function loadAccountsTab() {
    const listEl = document.getElementById('accounts-list');
    const emptyEl = document.getElementById('accounts-empty');
    const statusEl = document.getElementById('accounts-status');

    try {
        const accounts = await window.go.main.App.GetAccounts();

        if (!accounts || accounts.length === 0) {
            listEl.classList.add('hidden');
            listEl.innerHTML = '';
            emptyEl.classList.remove('hidden');
            return;
        }

        emptyEl.classList.add('hidden');
        listEl.classList.remove('hidden');
        listEl.innerHTML = '';

        accounts.forEach(acc => {
            listEl.appendChild(createAccountCard(acc));
        });

        ensureAccountListDragReorder(listEl);

    } catch (e) {
        if (statusEl) {
            statusEl.textContent = tf('statusError', { error: String(e) });
            statusEl.className = 'status-message error';
            scheduleAccountsStatusClear();
        }
    }
}

function createAccountCard(acc) {
    const card = document.createElement('div');
    card.className = 'account-card';
    card.dataset.accountId = acc.id;

    const statusClass = acc.hasSession ? 'has-session' : 'no-session';
    const statusText = acc.hasSession ? t('statusAutoLogin') : t('statusLoginRequired');

    card.innerHTML =
        '<div class="account-drag-handle" title="Drag to reorder" aria-hidden="true">' +
            '<span class="account-drag-grip"></span>' +
        '</div>' +
        '<div class="account-info">' +
            '<div class="account-name">' + escapeHtml(acc.name) + '</div>' +
            '<div class="account-email">' + escapeHtml(acc.email) + '</div>' +
            '<div class="account-status ' + statusClass + '">' +
                '<span class="status-dot"></span>' +
                '<span>' + escapeHtml(statusText) + '</span>' +
            '</div>' +
        '</div>' +
        '<div class="account-actions">' +
            '<button class="btn btn-primary" data-action="switch">' + t('btnSwitch') + '</button>' +
            '<button class="btn btn-danger" data-action="delete">' + t('btnDelete') + '</button>' +
        '</div>';

    card.querySelector('[data-action="switch"]').addEventListener('click', () => {
        onSwitchAccount(acc.id);
    });
    card.querySelector('[data-action="delete"]').addEventListener('click', () => {
        onDeleteAccount(acc.id);
    });

    return card;
}

let accountDragReorderBound = false;
let accountDragState = null;

function ensureAccountListDragReorder(listEl) {
    if (accountDragReorderBound) return;
    accountDragReorderBound = true;

    listEl.addEventListener('pointerdown', (e) => {
        if (e.button !== 0) return;
        if (e.target.closest('.account-actions') || e.target.closest('button')) return;

        const dragArea = e.target.closest('.account-drag-handle, .account-info');
        if (!dragArea) return;

        const card = dragArea.closest('.account-card');
        if (!card) return;

        e.preventDefault();
        dragArea.setPointerCapture(e.pointerId);

        accountDragState = {
            card,
            pointerId: e.pointerId,
            moved: false,
            listEl
        };
        card.classList.add('account-card-dragging');
    });

    listEl.addEventListener('pointermove', (e) => {
        if (!accountDragState || e.pointerId !== accountDragState.pointerId) return;

        accountDragState.moved = true;
        const { card, listEl: list } = accountDragState;
        const cards = [...list.querySelectorAll('.account-card')].filter(c => c !== card);
        const midY = e.clientY;

        let placed = false;
        for (const other of cards) {
            const rect = other.getBoundingClientRect();
            if (midY < rect.top + rect.height / 2) {
                list.insertBefore(card, other);
                placed = true;
                break;
            }
        }
        if (!placed) {
            list.appendChild(card);
        }
    });

    listEl.addEventListener('pointerup', finishAccountDragReorder);
    listEl.addEventListener('pointercancel', finishAccountDragReorder);
}

async function finishAccountDragReorder(e) {
    if (!accountDragState || e.pointerId !== accountDragState.pointerId) return;

    const { card, moved, listEl } = accountDragState;
    card.classList.remove('account-card-dragging');
    accountDragState = null;

    if (!moved) return;

    const orderedIDs = [...listEl.querySelectorAll('.account-card')]
        .map(c => c.dataset.accountId)
        .filter(Boolean);

    const statusEl = document.getElementById('accounts-status');
    try {
        await window.go.main.App.ReorderAccounts(orderedIDs);
        if (statusEl) {
            statusEl.textContent = '';
            statusEl.className = 'status-message';
        }
    } catch (err) {
        if (statusEl) {
            statusEl.textContent = tf('statusError', { error: String(err) });
            statusEl.className = 'status-message error';
            scheduleAccountsStatusClear();
        }
        await loadAccountsTab();
    }
}

async function onSwitchAccount(id) {
    const statusEl = document.getElementById('accounts-status');
    statusEl.textContent = '\u23F3 ' + t('statusLauncherRestarting');
    statusEl.className = 'status-message info';

    try {
        const result = await window.go.main.App.SwitchAccount(id);

        if (result.success) {
            if (result.hasSession) {
                statusEl.textContent = tf('statusAutoLoginActive', {
                    name: result.accountName
                });
                statusEl.className = 'status-message success';
            } else {
                statusEl.textContent = tf('statusManualLogin', {
                    name: result.accountName,
                    email: result.email
                });
                statusEl.className = 'status-message warning';
            }
        } else {
            statusEl.textContent = tf('statusError', { error: result.error });
            statusEl.className = 'status-message error';
        }
        scheduleAccountsStatusClear();
    } catch (e) {
        statusEl.textContent = tf('statusError', { error: String(e) });
        statusEl.className = 'status-message error';
        scheduleAccountsStatusClear();
    }
}

async function onDeleteAccount(id) {
    try {
        const confirmed = await window.go.main.App.ConfirmDelete();
        if (!confirmed) return;

        await window.go.main.App.DeleteAccount(id);

        const statusEl = document.getElementById('accounts-status');
        statusEl.textContent = '\u2713 ' + t('statusAccountDeleted');
        statusEl.className = 'status-message success';
        scheduleAccountsStatusClear();

        await loadAccountsTab();
    } catch (e) {
        const statusEl = document.getElementById('accounts-status');
        if (statusEl) {
            statusEl.textContent = tf('statusDeleteFailed', { error: String(e) });
            statusEl.className = 'status-message error';
            scheduleAccountsStatusClear();
        }
        console.error('Delete failed:', e);
    }
}

// ======================== ADD ACCOUNT ========================

function setupEventListeners() {
    document.getElementById('add-submit-btn').addEventListener('click', onAddAccount);
    document.getElementById('settings-lang-select').addEventListener('change', onLanguageChange);
    document.getElementById('settings-theme-select').addEventListener('change', onThemeChange);
    document.getElementById('settings-browse-btn').addEventListener('click', onBrowsePath);
    document.getElementById('settings-save-btn').addEventListener('click', onSavePath);
    document.getElementById('settings-autostart-check').addEventListener('change', onAutoStartToggle);
    document.getElementById('settings-streamer-check').addEventListener('change', onStreamerToggle);
    document.getElementById('settings-quit-btn').addEventListener('click', onQuitApp);

    // Allow Enter key in add form
    document.getElementById('add-email-input').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') onAddAccount();
    });
    document.getElementById('add-name-input').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            document.getElementById('add-email-input').focus();
        }
    });
}

async function onAddAccount() {
    const nameInput = document.getElementById('add-name-input');
    const emailInput = document.getElementById('add-email-input');
    const statusEl = document.getElementById('add-status');

    const name = nameInput.value.trim();
    const email = emailInput.value.trim();

    if (!name || !email) {
        statusEl.textContent = '\u26A0\uFE0F ' + t('statusFillFields');
        statusEl.className = 'status-message warning';
        scheduleAddStatusClear();
        return;
    }

    statusEl.textContent = '\u23F3 ' + t('statusLauncherRestarting');
    statusEl.className = 'status-message info';

    try {
        await window.go.main.App.AddAccount(name, email);

        statusEl.textContent = '';
        statusEl.className = 'status-message';

        nameInput.value = '';
        emailInput.value = '';

        await loadAccountsTab();

        const accountsStatusEl = document.getElementById('accounts-status');
        accountsStatusEl.textContent = t('statusAccountAdded');
        accountsStatusEl.className = 'status-message success';
        scheduleAccountsStatusClear();

        selectTab('accounts');
    } catch (e) {
        statusEl.textContent = tf('statusError', { error: String(e) });
        statusEl.className = 'status-message error';
        scheduleAddStatusClear();
    }
}

// ======================== SETTINGS ========================

async function loadSettingsValues() {
    try {
        const settings = await window.go.main.App.GetSettings();

        document.getElementById('settings-lang-select').value = settings.language;
        document.getElementById('settings-path-input').value = settings.launcherPath;
        document.getElementById('settings-autostart-check').checked = settings.autoStart;
        document.getElementById('settings-streamer-check').checked = settings.streamerMode;

        // Set theme dropdown
        const themeSelect = document.getElementById('settings-theme-select');
        if (settings.theme) {
            themeSelect.value = settings.theme;
        } else {
            themeSelect.value = 'eft';
        }
    } catch (e) {
        console.error('Failed to load settings:', e);
    }
}

async function onLanguageChange() {
    const lang = document.getElementById('settings-lang-select').value;
    const statusEl = document.getElementById('settings-status');

    try {
        await window.go.main.App.SetLanguage(lang);
        await loadTranslations();
        applyTranslations();
        await loadAccountsTab();

        statusEl.textContent = '\u2713 ' + t('statusLanguageSaved');
        statusEl.className = 'status-message success';
        scheduleSettingsStatusClear();
    } catch (e) {
        console.error('Language change failed:', e);
        statusEl.textContent = tf('statusError', { error: String(e) });
        statusEl.className = 'status-message error';
    }
}

async function onBrowsePath() {
    try {
        const path = await window.go.main.App.BrowseLauncherPath();
        if (path) {
            document.getElementById('settings-path-input').value = path;
        }
    } catch (e) {
        console.error('Browse failed:', e);
    }
}

async function onSavePath() {
    const path = document.getElementById('settings-path-input').value.trim();
    const statusEl = document.getElementById('settings-status');

    if (!path) {
        statusEl.textContent = '\u26A0\uFE0F ' + t('statusEnterPath');
        statusEl.className = 'status-message warning';
        scheduleSettingsStatusClear();
        return;
    }

    try {
        await window.go.main.App.SetLauncherPath(path);
        statusEl.textContent = '\u2713 ' + t('statusPathSaved');
        statusEl.className = 'status-message success';
        scheduleSettingsStatusClear();
    } catch (e) {
        cancelSettingsStatusClear();
        statusEl.textContent = tf('statusError', { error: String(e) });
        statusEl.className = 'status-message error';
    }
}

async function onAutoStartToggle() {
    const checked = document.getElementById('settings-autostart-check').checked;
    const statusEl = document.getElementById('settings-status');

    try {
        await window.go.main.App.SetAutoStart(checked);
        statusEl.textContent = '\u2713 ' + (checked ? t('statusAutoStartOn') : t('statusAutoStartOff'));
        statusEl.className = 'status-message success';
        scheduleSettingsStatusClear();
    } catch (e) {
        console.error('Autostart toggle failed:', e);
        document.getElementById('settings-autostart-check').checked = !checked;
        statusEl.textContent = tf('statusError', { error: String(e) });
        statusEl.className = 'status-message error';
    }
}

async function onStreamerToggle() {
    const checked = document.getElementById('settings-streamer-check').checked;
    const statusEl = document.getElementById('settings-status');

    try {
        await window.go.main.App.SetStreamerMode(checked);
        statusEl.textContent = '\u2713 ' + (checked ? t('statusStreamerOn') : t('statusStreamerOff'));
        statusEl.className = 'status-message success';
        scheduleSettingsStatusClear();

        // Refresh accounts to apply email masking
        await loadAccountsTab();
    } catch (e) {
        console.error('Streamer mode toggle failed:', e);
        document.getElementById('settings-streamer-check').checked = !checked;
        statusEl.textContent = tf('statusError', { error: String(e) });
        statusEl.className = 'status-message error';
    }
}

// ======================== QUIT ========================

async function onQuitApp() {
    try {
        await window.go.main.App.QuitApp();
    } catch (e) {
        console.error('Quit failed:', e);
    }
}

// ======================== WAILS EVENTS ========================

function setupWailsEvents() {
    if (!window.runtime) {
        console.warn('Wails runtime not available for events');
        return;
    }

    // Session captured by Go backend -> refresh account list
    window.runtime.EventsOn('session-captured', async () => {
        await loadAccountsTab();
    });

    // Update available -> show banner (only if version is actually newer)
    window.runtime.EventsOn('update-available', async (data) => {
        let currentVersion = '';
        try {
            currentVersion = await window.go.main.App.GetVersion();
        } catch (e) { /* ignore */ }

        const banner = document.getElementById('update-banner');
        let lines = [];

        if (data && data.stable) {
            const v = data.stable.Version || data.stable.version || '';
            if (v !== currentVersion) {
                const href = data.stable.ReleaseURL || data.stable.releaseURL || '#';
                const dl = escapeHtml(t('linkDownload'));
                lines.push(tf('updateAvailableStable', {
                    version: v,
                    url: '<a href="' + href + '" target="_blank" rel="noopener noreferrer">' + dl + '</a>'
                }));
            }
        }
        if (data && data.beta) {
            const v = data.beta.Version || data.beta.version || '';
            if (v !== currentVersion) {
                const href = data.beta.ReleaseURL || data.beta.releaseURL || '#';
                const dl = escapeHtml(t('linkDownload'));
                lines.push(tf('updateAvailableBeta', {
                    version: v,
                    url: '<a href="' + href + '" target="_blank" rel="noopener noreferrer">' + dl + '</a>'
                }));
            }
        }

        if (lines.length > 0) {
            banner.innerHTML = lines.join('<br>');
            banner.classList.remove('hidden');
        }
    });
}

// ======================== UTILITIES ========================

function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
