# GameTrainerHub

> **Windows 桌面端 PC 游戏修改器聚合管理客户端**  
> *A Windows desktop application for discovering, downloading, launching, and managing game trainers.*

[![Windows CI](https://github.com/18270898659/GameTrainerHub/actions/workflows/ci.yml/badge.svg)](https://github.com/18270898659/GameTrainerHub/actions/workflows/ci.yml)
[![Release](https://github.com/18270898659/GameTrainerHub/actions/workflows/release.yml/badge.svg)](https://github.com/18270898659/GameTrainerHub/actions/workflows/release.yml)

---

## 1. 项目简介 | Overview

**GameTrainerHub** 是一款面向 Windows 10 / Windows 11 (x64) 平台的现代 PC 游戏修改器聚合管理桌面客户端，基于 **Go + Wails v2 + Vue 3 + TypeScript + SQLite** 构建。

以 FLiNG Trainer 为首发数据源，通过标准化的站点适配器架构实现资源发现、中英双向智能检索、HTTP Range 分块断点续传下载、本地多版本库管理、安全隔离启动与全局快捷键控制。

---

## 2. 项目定位与核心目标 | Positioning & Goals

* **轻量原生桌面体验**：基于 Wails v2 与系统 WebView2 运行时，免去 Electron 体积膨胀，兼具原生 Windows 托盘与全局快捷键响应能力。
* **本地优先与离线可用**：内置 SQLite FTS5 全文检索引擎与增量同步机制，兼顾毫秒级本地检索与在线实时补充检索。
* **安全透明的启动隔离**：拒绝静默提权或隐藏行为；启动前展示文件来源、SHA-256 哈希与权限风险提示，由用户显式确认。
* **可靠安装与无损升级**：采用 NSIS 安装包并内置原生 Win32 进程检测与三态安全失败（Fail-Safe）保护，升级安装完整保留用户 SQLite 数据与下载记录。

---

## 3. 功能特性 | Features

* **中英双向与拼音智能搜索**：支持中文查英文、英文查中文、游戏简称/别名（如“黑猴”、“老头环”）、全拼及拼音首字母匹配。
* **混合检索引擎**：本地 SQLite FTS5 全文索引与 FLiNG 站点实时抓取无缝融合并自动去重排序。
* **高可用下载管理器**：支持多任务并发调度、HTTP Range 断点续传、SHA-256 完整性校验、ZIP 自动解压与异常状态持久化恢复。
* **本地修改器库与 Steam 扫描**：自动识别本地已安装的 Steam 游戏，支持多版本保留、收藏管理、一键打开目录与下载历史回溯。
* **全局快捷键与系统托盘**：支持物理键盘实时录入组合键（支持 `Ctrl` / `Alt` / `Shift` / `Win`）、系统级热键冲突检测与原子回滚（默认 `Alt + Ctrl + Q` 呼出/隐藏，`Alt + Ctrl + J` 直达下载）。
* **安全启动弹窗**：启动外部修改器前核验路径与哈希，明确提示 Windows UAC / 杀毒误报风险与隔离策略。
* **多语言界面**：内置简体中文、繁体中文与 English 实时切换，支持深色/浅色主题。

---

## 4. 项目截图 | Screenshots

<img width="1266" height="844" alt="首页" src="https://github.com/user-attachments/assets/145391f1-b253-4855-b76d-cb550af5231f" />
<img width="1266" height="844" alt="搜索" src="https://github.com/user-attachments/assets/38cf246d-4064-4856-ab20-dda6fabf6a4d" />
<img width="1266" height="844" alt="下载管理" src="https://github.com/user-attachments/assets/107a262a-fb6d-4490-9b5b-ad10c2524bbe" />
<img width="1266" height="844" alt="本地修改器" src="https://github.com/user-attachments/assets/89376c35-27f8-4d66-8dea-0721a000b90c" />
<img width="1266" height="844" alt="设置" src="https://github.com/user-attachments/assets/e3036727-85ef-4d2f-8713-a344fb564671" />


---

## 5. 技术架构 | Architecture

```text
+-------------------------------------------------------------------+
|                   Vue 3 + TypeScript + Element Plus               |
|         (Views: Home / Search / Downloads / Library / Settings)   |
+---------------------------------+---------------------------------+
                                  | Wails v2 IPC Bridge
+---------------------------------v---------------------------------+
|                        Go Backend (internal/app)                  |
|  +-------------+  +--------------+  +-----------+  +-----------+  |
|  |   Search    |  |   Download   |  |  Trainer  |  |  Launcher |  |
|  |  & Translate|  |   Manager    |  |  Library  |  | (Isolated)|  |
|  +------+------+  +------+-------+  +-----+-----+  +-----+-----+  |
|         |                |                |                |      |
|  +------v----------------v----------------v----------------v---+  |
|  |       SQLite (modernc.org/sqlite + FTS5) & App Config       |  |
|  +-------------------------------------------------------------+  |
|  |    Windows Tray & Native Global Hotkey (user32.dll / Win32) |  |
+-------------------------------------------------------------------+
```

---

## 6. 技术栈说明 | Tech Stack

| 层级 | 技术选型 | 说明 |
| :--- | :--- | :--- |
| **桌面框架** | Wails v2 (`v2.17.1`) | Go 后端与 Windows WebView2 桥接 |
| **后端语言** | Go (`1.26.0`) | 并发下载调度、HTML 解析、Win32 API 调用 |
| **前端框架** | Vue 3 (`3.5`) + TypeScript (`5.7`) + Vite (`6.0`) | 响应式 SPA 与类型安全 |
| **UI 与状态** | Element Plus (`2.9`) + Pinia (`2.3`) + Vue Router (`4.5`) | 界面组件与状态管理 |
| **本地数据库** | SQLite (`modernc.org/sqlite` + FTS5) | 纯 Go 零 CGO SQLite 引擎 |
| **安装打包** | NSIS 3 (`Unicode` + MUI2 + Win32 `System::Call`) | Windows x64 安装包与便携版 ZIP |

---

## 7. 项目目录结构 | Project Structure

```text
GameTrainerHub/
├── .github/workflows/          # GitHub Actions CI 检查与 Release 自动发布工作流
├── build/                      # 构建资源与 Windows 安装脚本
│   ├── installer/              # NSIS 安装脚本 (installer.nsi)、辅助关闭脚本、声明文件
│   └── windows/                # Windows 应用图标、清单与版本资源
├── docs/                       # 安装指南、开发文档与常见问题
├── frontend/                   # Vue 3 + TypeScript 前端工程
│   └── src/
│       ├── api/                # Wails IPC 封装
│       ├── components/         # 详情抽屉、快捷键录入、安全提示弹窗等组件
│       ├── i18n/               # 简中 / 繁中 / 英文语言包
│       ├── stores/             # Pinia 状态管理
│       └── views/              # 首页、搜索、下载、本地库、设置视图
├── internal/                   # Go 后端核心模块及单元测试
│   ├── app/                    # Wails 应用生命周期与 IPC 暴露接口
│   ├── config/                 # 配置加载与持久化
│   ├── database/               # SQLite 连接、迁移与 FTS5 仓储
│   ├── diagnostics/            # 网络与运行环境诊断
│   ├── download/               # 多线程 Worker 与 Range 断点续传
│   ├── game/                   # Steam 本地游戏扫描
│   ├── launcher/               # 安全启动器与哈希计算
│   ├── logger/                 # 结构化日志与轮转
│   ├── model/                  # 领域数据模型
│   ├── search/                 # 混合搜索调度
│   ├── source/                 # 站点适配器抽象与 FLiNG 实现
│   ├── sync/                   # 本地种子数据与增量同步
│   ├── trainer/                # 本地修改器管理
│   ├── translate/              # 游戏名中英词典与拼音索引
│   ├── tray/                   # Windows 系统托盘与全局快捷键
│   └── update/                 # 应用版本检查
├── main.go                     # 程序主入口
├── wails.json                  # Wails 项目配置
├── build.bat                   # 快速本地构建脚本
├── package.bat                 # 一键前端构建 + Wails 打包 + 便携版 + NSIS 安装包
└── CHANGELOG.md                # 版本变更记录
```

---

## 8. 环境依赖 | Prerequisites

* **运行环境（终端用户）**：
  * Windows 10 (1809 / Build 17763 或更高版本) 或 Windows 11 (64-bit)
  * Microsoft Edge WebView2 Runtime（Windows 11 默认内置；Windows 10 若缺失，安装向导将引导下载）
* **编译环境（开发者）**：
  * Go `1.26+`
  * Node.js `20+` & npm
  * Wails CLI `v2.17.1` (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1`)
  * NSIS `3.x`（仅构建 Windows 安装包时需要）

---

## 9. 开发环境搭建 | Development Setup

```bash
# 1. 克隆仓库
git clone https://github.com/18270898659/GameTrainerHub.git
cd GameTrainerHub

# 2. 安装前端依赖
cd frontend
npm install
cd ..

# 3. 验证 Go 依赖与测试
go mod download
go vet ./...
go test -count=1 ./internal/...

# 4. 启动 Wails 热重载开发模式
wails dev
```

更多说明请参阅 [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)。

---

## 10. 本地编译方法 | Local Build

```bash
# 方式一：使用内置构建脚本（Windows）
build.bat

# 方式二：手动构建
cd frontend && npm run build && cd ..
wails build -clean -s
```

编译产物位于 `build/bin/GameTrainerHub.exe`。

---

## 11. Windows 安装包构建方法 | Installer Packaging

执行根目录下的 `package.bat`，脚本将自动完成前端构建、Wails 生产编译、便携版 ZIP 压缩、NSIS 安装包编译及 SHA-256 校验文件生成：

```bash
package.bat
```

全部产物将输出至 `bin/` 目录：
* `bin/GameTrainerHub-Setup-x64.exe`（Windows x64 安装程序）
* `bin/GameTrainerHub-Portable-x64.zip`（免安装便携版）
* `bin/GameTrainerHub.exe`（独立可执行文件）
* `bin/checksums.txt`（SHA-256 校验清单）

---

## 12. 用户使用说明 | User Guide

1. 从 [GitHub Releases](https://github.com/18270898659/GameTrainerHub/releases) 下载 `GameTrainerHub-Setup-x64.exe` 或便携版压缩包。
2. 安装后启动客户端，在搜索栏输入中文、英文或拼音简称查找对应游戏修改器。
3. 点击下载后可在「下载管理」查看实时进度，下载完成后自动归档至「本地库」。
4. 启动修改器前请仔细阅读安全提示弹窗并确认文件哈希，再继续启动。
5. 可在「设置」页面自定义下载目录、界面语言、主题以及全局快捷键。

详细安装与升级流程请参阅 [docs/INSTALLATION.md](docs/INSTALLATION.md)。

---

## 13. 配置与数据存储说明 | Configuration & Data Storage

为严格遵守 Windows 权限隔离规范，客户端运行期间不会向 `C:\Program Files\GameTrainerHub` 写入任何用户数据：

* **配置文件**：`%APPDATA%\GameTrainerHub\config.json`
* **SQLite 数据库**：`%APPDATA%\GameTrainerHub\data\trainers.db`
* **默认下载目录**：`%APPDATA%\GameTrainerHub\downloads\`（可在设置中修改）
* **运行日志**：`%APPDATA%\GameTrainerHub\logs\`

卸载程序默认保留 `%APPDATA%\GameTrainerHub` 中的用户数据，确保重装或升级不会丢失已下载的修改器与历史记录。

---

## 14. 常见问题与解决方案 | FAQ

请参阅 [docs/FAQ.md](docs/FAQ.md) 获取关于 WebView2 缺失、杀毒软件误报、覆盖安装进程占用检测及全局快捷键冲突的详细排查方案。

---

## 15. 安全与隐私说明 | Security & Privacy

* 本软件不收集、不上传任何用户个人隐私数据，所有配置、收藏与下载历史均仅存储于本地 `%APPDATA%\GameTrainerHub`。
* 下载文件均计算并校验 SHA-256 摘要，启动外部修改器前不进行静默提权。
* 发现安全漏洞请遵循 [SECURITY.md](SECURITY.md) 中的流程反馈。

---

## 16. 许可证信息 | License

本项目基于 [MIT License](LICENSE) 开源发布。您可以在遵守 MIT 协议条款的前提下自由使用、修改与分发本软件。

---

## 17. 版本发布记录 | Changelog

完整版本变更历史请参阅 [CHANGELOG.md](CHANGELOG.md)。

---

## 18. 问题反馈与贡献指南 | Contributing & Feedback

* 提交 Bug 或功能建议：[GitHub Issues](https://github.com/18270898659/GameTrainerHub/issues)
* 参与代码或文档贡献：请先阅读 [CONTRIBUTING.md](CONTRIBUTING.md)
