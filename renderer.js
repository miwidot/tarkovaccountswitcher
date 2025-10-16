const { ipcRenderer } = require('electron');
const translations = require('./translations');

let currentLanguage = 'en';
let t = translations.en;

// Load accounts and settings on startup
window.addEventListener('DOMContentLoaded', async () => {
  await loadSettings();
  await loadAccounts();
});

// Listen for session-captured events from main process
ipcRenderer.on('session-captured', (event, accountId) => {
  console.log('✅ Session captured for account:', accountId);
  // Reload accounts list to show updated status
  loadAccounts();
});

// Apply translations to UI
function applyTranslations() {
  // Update all elements with data-i18n attribute
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (t[key]) {
      el.textContent = t[key];
    }
  });

  // Update all elements with data-i18n-placeholder attribute
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    const key = el.getAttribute('data-i18n-placeholder');
    if (t[key]) {
      el.placeholder = t[key];
    }
  });

  // Update all elements with data-i18n-html attribute (for HTML content)
  document.querySelectorAll('[data-i18n-html]').forEach(el => {
    const key = el.getAttribute('data-i18n-html');
    if (t[key]) {
      el.innerHTML = t[key];
    }
  });
}

async function changeLanguage() {
  const language = document.getElementById('language').value;
  await ipcRenderer.invoke('set-language', language);

  // Update current language and translations
  currentLanguage = language;
  t = translations[language];

  // Apply translations to entire UI
  applyTranslations();

  // Reload accounts to update translated text
  await loadAccounts();

  showStatus(t.statusLanguageSaved, 'success');
}

// Tab switching function
function switchTab(tabName) {
  // Remove active class from all tabs and tab contents
  document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
  document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

  // Add active class to selected tab and content
  const tabs = document.querySelectorAll('.tab');
  if (tabName === 'accounts') tabs[0].classList.add('active');
  else if (tabName === 'add') tabs[1].classList.add('active');
  else if (tabName === 'settings') tabs[2].classList.add('active');

  document.getElementById(tabName + '-tab').classList.add('active');
}

async function loadAccounts() {
  const accounts = await ipcRenderer.invoke('get-accounts');
  const accountsList = document.getElementById('accountsList');

  if (accounts.length === 0) {
    accountsList.innerHTML = `
      <div class="empty-state">
        <p>${t.emptyStateTitle}</p>
        <p style="font-size: 12px; margin-top: 10px;">${t.emptyStateSubtitle}</p>
      </div>
    `;
    return;
  }

  accountsList.innerHTML = accounts.map(account => `
    <div class="account-item">
      <div class="account-info">
        <div class="account-name">${escapeHtml(account.name)} ${account.launcherSession ? '✅' : '⚠️'}</div>
        <div class="account-email">${escapeHtml(account.email)}</div>
        <div style="font-size: 11px; color: #666; margin-top: 3px;">
          ${account.launcherSession ? t.statusAutoLogin : t.statusLoginRequired}
        </div>
      </div>
      <div class="account-actions">
        <button class="btn-switch" onclick="switchAccount('${account.id}')">
          ${t.btnSwitch}
        </button>
        <button class="btn-danger" onclick="deleteAccount('${account.id}')">
          ${t.btnDelete}
        </button>
      </div>
    </div>
  `).join('');
}

async function addAccount() {
  const name = document.getElementById('accountName').value.trim();
  const email = document.getElementById('accountEmail').value.trim();

  if (!name || !email) {
    showStatus(t.statusFillFields, 'error');
    return;
  }

  const account = { name, email };
  const result = await ipcRenderer.invoke('add-account', account);

  if (result.success) {
    showStatus(t.statusAccountAdded, 'success');
    document.getElementById('accountName').value = '';
    document.getElementById('accountEmail').value = '';
    loadAccounts();

    // Switch back to accounts tab
    switchTab('accounts');
  } else {
    showStatus(t.statusError.replace('{error}', result.error), 'error');
  }
}

async function deleteAccount(id) {
  if (!confirm(t.confirmDelete)) {
    return;
  }

  const result = await ipcRenderer.invoke('delete-account', id);

  if (result.success) {
    showStatus(t.statusAccountDeleted, 'success');
    loadAccounts();
  } else {
    showStatus(t.statusDeleteError, 'error');
  }
}

async function switchAccount(id) {
  showStatus(t.statusLauncherRestarting, 'success');

  const result = await ipcRenderer.invoke('switch-account', { id });

  if (result.success) {
    if (result.hasSession) {
      showStatus(t.statusAutoLoginActive.replace('{name}', result.accountName), 'success');
    } else {
      showStatus(t.statusManualLogin.replace('{name}', result.accountName).replace('{email}', result.email), 'success');
    }
  } else {
    showStatus(t.statusError.replace('{error}', result.error), 'error');
  }
}

function showStatus(message, type) {
  const statusEl = document.getElementById('statusMessage');
  statusEl.innerHTML = message.replace(/\n/g, '<br>');
  statusEl.className = `status-message status-${type}`;
  statusEl.style.display = 'block';

  // Scroll to top to show status
  window.scrollTo({ top: 0, behavior: 'smooth' });

  setTimeout(() => {
    statusEl.style.display = 'none';
  }, 6000);
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

async function loadSettings() {
  const settings = await ipcRenderer.invoke('get-settings');
  document.getElementById('launcherPath').value = settings.launcherPath;

  // Set language
  if (settings.language) {
    currentLanguage = settings.language;
    t = translations[currentLanguage];
    document.getElementById('language').value = currentLanguage;

    // Apply translations to UI
    applyTranslations();
  }
}

async function saveLauncherPath() {
  const launcherPath = document.getElementById('launcherPath').value.trim();

  if (!launcherPath) {
    showStatus(t.statusEnterPath, 'error');
    return;
  }

  const result = await ipcRenderer.invoke('set-launcher-path', launcherPath);

  if (result.success) {
    showStatus(t.statusPathSaved, 'success');
  } else {
    showStatus(t.statusSaveError, 'error');
  }
}

async function browseLauncherPath() {
  const result = await ipcRenderer.invoke('browse-launcher-path');

  if (result.filePath) {
    document.getElementById('launcherPath').value = result.filePath;
  }
}
