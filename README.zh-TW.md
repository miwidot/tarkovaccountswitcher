# Tarkov Account Switcher（塔科夫多帳號切換）

**English** · [README.md](README.md)

![Version](https://img.shields.io/badge/version-2.0.8-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

適用於 **Escape from Tarkov** 的多帳號切換工具，具自動連線／工作階段管理與加密儲存。

## 功能

- **自動工作階段管理** — 在本機加密儲存登入／工作階段
- **一鍵切換帳號** — 會自動重新啟動啟動器並切換到所選帳號
- **不儲存密碼** — 僅保留電郵與連線／工作階段權杖（AES-256 加密）
- **自動登入** — 首次登入後可自動沿用工作階段
- **更新通知** — 啟動時檢查 GitHub Releases，並以橫幅顯示下載連結
- **多語系** — 英文、德文、繁體中文、簡體中文與日文，並可依系統語言自動偵測
- **實況／隱私模式** — 可在介面上遮罩電郵地址
- **系統匣整合** — 背景執行，啟動器啟動時可自動最小化視窗
- **單一實例** — 同一時間只能執行一份程式
- **主題** — 五種介面主題（含取材自 Escape from Tarkov 風格的版型）
- **依帳號的啟動器選項** — 每個帳號可分別設定 EFT 或 Arena 等相關啟動器偏好

## 下載

[最新 Release](https://github.com/miwidot/tarkovaccountswitcher/releases/latest)

## 快速上手

### 安裝

1. 從最新 Release 下載 `Tarkov Account Switcher.exe`
2. 直接執行 exe，無須另行安裝
3. 完成；程式會在系統匣中運作

### 新增第一個帳號

1. 開啟 **「Add」**（新增）分頁
2. 輸入 **帳號顯示名稱** 與 **電郵**（例如：`Main`、`main@email.com`）
3. 按 **「Add Account & Start Launcher」**（新增帳號並啟動啟動器）
4. 啟動器會自動開啟
5. **在啟動器內照常登入**
6. 連線會 **自動偵測並儲存**
7. 該帳號側會顯示綠色勾選

### 切換帳號

1. 開啟 **「Accounts」**（帳號）分頁
2. 在目標帳號旁按 **「Switch」**
3. 啟動器會自動重啟，且 **已處於登入狀態**

## 安全與 BSG 說明

### 本工具會做什麼

- 自 BSG 啟動器設定檔讀取連線／工作階段權杖（`%APPDATA%\Battlestate Games\BsgLauncher\settings`）
- 以 AES-256-CBC 加密後寫入本機 `%APPDATA%\TarkovAccountSwitcher\accounts.json`
- 切換時：終止啟動器行程、覆寫啟動器設定中的工作階段資料後再重新啟動啟動器
- **不儲存密碼** — 僅有電郵地址與工作階段權杖

### 本工具不會做什麼

- **不修改遊戲檔案** — 未讀寫或修補 EFT、Arena 等遊戲本體檔案
- **不做程式碼／記憶體注入** — 無 DLL 注入、記憶體操作或掛勾
- **不與防外掛互動** — 不碰 BattlEye 或其他防外掛元件
- **不干預網路** — 流量攔截、代理、MITM 皆不涉及
- **不與 BSG 伺服器通訊** — 程式不直接向 BSG 伺服器發出請求
- **不提供雲端同步** — 資料僅保存在你的電腦上

運作範圍僅限 **BSG 啟動器本機設定檔**，在帳號之間替換連線／工作階段權杖；概念上類似手動複製並貼上同一份設定檔。

### 隱私

- 資料 **皆保留於本機**
- AES-256-CBC 加密；每組安裝各自一把金鑰
- 無遙測、無第三方分析；（除 GitHub 更新檢查外）無對外連線取得使用資料之行為

## 免責聲明

**本工具不改動任何遊戲檔案，亦不進行程式／記憶體注入**；僅讀寫 BSG 啟動器本機設定檔以管理工作階段權杖。

- **現階段理解**：整體風險相對偏低 — 性質接近其他「帳號切換／工作階段切換」工具（例如 TcNo Account Switcher）
- **不作保證**：請 **自負風險** 使用。若 BSG 更新服務條款或使用政策，前述判斷可能改變

**建議：**

- 為 BSG 帳號啟用兩步驟驗證（2FA）
- 不同帳號使用不同密碼
- 切勿將憑證交給第三方

## 技術堆疊

- **Go** — Windows 原生後端程式
- **Wails v2** — 嵌入 **WebView2** 的桌面殼（介面為內嵌的 HTML/CSS/JS；除 Windows 內建的 WebView2 執行環境外，不需另行安裝瀏覽器）
- **Vanilla 前端** — 單頁介面位於 `v2/frontend/dist/` — 發行版本不使用 Electron；亦無需 npm 打包流程即可建出執行檔
- **AES-256-CBC** — 透過 Go 標準函式庫加密
- **Windows API** — 行程管理、原生系統匣（`Shell_NotifyIconW`）

## 從原始碼建置

實際維護中的程式碼在 **`v2/`** 目錄（儲藏庫根目錄的舊 Walk／Electron 專線僅作歷史保留）。

**先決條件**

- Go **1.23+**（數字以 `v2/go.mod` 為準； toolchain 可能鎖定較新修補版）
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **MinGW-w64 GCC** 已加入 `PATH`（Windows 上 Wails／CGO 需要）

**指令**（在 **`v2/`** 底下執行）：

```bash
cd v2
go mod tidy
wails build -platform windows/amd64
```

開發並需熱重載：`wails dev`。發行輸出在 `v2/build/bin/Tarkov Account Switcher.exe`。正式發行前，請將 PE／應用程式版本與更新器常數同步：`go run sync_version.go`（說明見 `v2/sync_version.go` 內註解）。

版面、主題與套件對照等開發說明請見：**`v2/README.md`**。

## 授權

本專案依 **MIT License** 授權 — 詳見 [LICENSE](LICENSE)。

---

**獻給塔科夫社群，用心維護**
