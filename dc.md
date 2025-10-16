# 🎮 Tarkov Account Switcher — v1.3.1 Beta (Windows)

## 📢 Ankündigung

Guten Morgen zusammen!

Viele von euch haben mehrere EFT-Accounts und müssen bei jedem Wechsel den Launcher neu authentifizieren. Deshalb haben wir einen **Account Switcher** entwickelt, der das automatisiert — getestet mit mehreren Accounts, läuft stabil!

---

## ✨ Was macht das Tool?

- **Automatische Session-Verwaltung**: Speichert eure Login-Sessions verschlüsselt
- **Ein-Klick Account-Wechsel**: Launcher wird automatisch neu gestartet mit dem gewählten Account
- **Keine Passwörter gespeichert**: Nur Email + Session-Tokens (verschlüsselt mit AES-256)
- **Auto-Login**: Nach dem ersten Login werdet ihr automatisch eingeloggt bei zukünftigen Wechseln
- **System Tray Integration**: Läuft im Hintergrund, minimiert sich automatisch

---

## 🔧 Features

✅ **Vollautomatisch**: Session wird beim Login automatisch erkannt und gespeichert
✅ **Verschlüsselte Speicherung**: Alle Daten lokal in `%APPDATA%\TarkovAccountSwitcher`
✅ **Multi-Language**: Deutsch & English (erkennt automatisch System-Sprache)
✅ **System Tray**: Minimiert sich nach Launcher-Start
✅ **Single Instance**: Nur eine App-Instanz kann laufen
✅ **Custom Launcher-Pfad**: Falls euer Launcher woanders installiert ist

---

## 🔒 Sicherheit & Technik

### Was das Tool macht:
- ✅ Liest Session-Tokens aus BSG Launcher Settings (`%APPDATA%\Battlestate Games\BsgLauncher\settings`)
- ✅ Speichert diese verschlüsselt (AES-256) lokal in `%APPDATA%\TarkovAccountSwitcher\accounts.json`
- ✅ Beim Wechsel: Launcher-Prozess killen → Session-Daten ersetzen → Launcher neu starten
- ✅ **Kein Passwort wird gespeichert** - nur Email + Session-Tokens

### Was das Tool NICHT macht:
- ❌ Keine Modifikation von Spiel-Dateien
- ❌ Keine Injection/Patching
- ❌ Keine Cloud-Synchronisation
- ❌ Keine Netzwerk-Manipulation

### Datenschutz:
- 🔐 Alle Daten bleiben **lokal auf eurem PC**
- 🔐 Verschlüsselung mit AES-256-CBC
- 🔐 Einzigartiger Encryption Key pro Installation
- 🔐 Keine Telemetrie, keine Analytics

---

## ⚠️ Ban-Risiko / TOS

**Wichtig - lest das bitte:**

- Das Tool verändert **keine Spiel-Dateien** und führt **keine Code-Injection** aus
- Es arbeitet nur mit den Launcher-Session-Daten (ähnlich wie TcNo Account Switcher)
- **Nach aktuellem Wissen**: Minimales Risiko
- **ABER**: Ich gebe **keine Garantie**. Nutzt das Tool **auf eigenes Risiko**!
- Falls BSG in Zukunft ihre TOS ändert, kann sich die Einschätzung ändern

**Empfehlungen:**
- ✅ Aktiviert 2FA bei eurem BSG Account
- ✅ Macht vor dem ersten Einsatz ein Backup eurer wichtigen Dateien
- ✅ Gebt niemals eure Zugangsdaten an Dritte weiter
- ✅ Nutzt verschiedene Passwörter für verschiedene Accounts

---

## 📥 Installation & Nutzung

### Download:
📦 **[Tarkov-Account-Switcher-v1.3.1.zip](LINK_HIER)** (~108 MB)

### Installation:
1. ZIP-Datei entpacken (wo ihr wollt, z.B. `C:\Program Files\TarkovAccountSwitcher\`)
2. `Tarkov Account Switcher.exe` starten
3. Fertig! Die App läuft im System Tray

### Ersten Account hinzufügen:
1. **Tab "Hinzufügen"** öffnen
2. **Account Name** + **Email** eingeben (z.B. "Main", "main@email.com")
3. **"Account hinzufügen & Launcher starten"** klicken
4. Launcher startet automatisch
5. **Im Launcher normal einloggen**
6. Session wird **automatisch erkannt und gespeichert** ✅
7. Account zeigt jetzt **grünes Häkchen** ✅

### Account wechseln:
1. **Tab "Accounts"** öffnen
2. Account auswählen
3. **"Wechseln"** klicken
4. Launcher startet automatisch **bereits eingeloggt**! 🚀

### Launcher-Pfad ändern (optional):
Falls euer Launcher woanders installiert ist:
1. **Tab "Einstellungen"** öffnen
2. Pfad eingeben oder **"Durchsuchen"** klicken
3. **"Pfad speichern"**

### Sprache ändern (optional):
Die App erkennt automatisch eure System-Sprache (Deutsch/English). Falls ihr manuell wechseln wollt:
1. **Tab "Einstellungen"** öffnen
2. **Sprache auswählen** (Deutsch / English)
3. UI aktualisiert sich sofort ✅

---

## 🐛 Beta-Hinweise

Dies ist eine **Beta-Version (v1.3.1)**.

**Bekannte Limitierungen:**
- Nur Windows 10/11 unterstützt
- Launcher muss BSG Launcher sein (kein Steam-Version Support)
- Session-Tokens können ablaufen (einfach neu einloggen - wird automatisch gespeichert)

**Feedback & Bugs:**
Falls ihr Probleme habt oder Feedback geben wollt, schreibt mir bitte hier im Channel!

---

## 📋 Changelog

### v1.3.1 (Aktuell)
- 🌍 **Multi-Language Support**: Deutsch & English mit automatischer System-Sprache Erkennung
- 🐛 **Session Token Fix**: Token werden jetzt korrekt gelöscht beim Account-Wechsel (verhindert falsche Session-Speicherung)
- 🐛 **Path Merge Fix**: System-spezifische Pfade werden beim Session-Restore nicht überschrieben
- ✅ Verbesserte Session-Verwaltung für stabileren Account-Wechsel

### v1.3.0
- ✅ ASAR-Packaging für sauberere Dateistruktur
- ✅ Session-Watcher Optimierungen

### v1.2.0
- ✅ Vollautomatische Session-Erkennung
- ✅ Kein Passwort-Speicherung mehr (nur Session-Tokens)
- ✅ System Tray Integration
- ✅ Single Instance Lock
- ✅ Tab-basierte UI (Accounts / Hinzufügen / Einstellungen)
- ✅ Launcher-Kill beim Account-Hinzufügen (verhindert alte Sessions)
- ✅ Email-Validierung beim Session-Capture
- ✅ Custom Icon Support

---

## 🙏 Danke fürs Testen!

Viel Spaß beim Switchen! 🎯

*Bei Fragen einfach melden!*
