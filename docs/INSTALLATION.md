# 安装与升级指南 | Installation & Upgrade Guide

## 系统要求

* **操作系统**：Windows 10 1809 (Build 17763) 或更高版本 / Windows 11 (64-bit)
* **运行依赖**：Microsoft Edge WebView2 Runtime（Windows 11 默认自带）

## 全新安装

1. 前往 [GitHub Releases](https://github.com/18270898659/GameTrainerHub/releases) 下载最新版本的 `GameTrainerHub-Setup-x64.exe`（安装版）或 `GameTrainerHub-Portable-x64.zip`（便携版）。
2. 建议核对下载文件的 SHA-256 与 `SHA256SUMS.txt` 是否一致：
   ```powershell
   Get-FileHash -Algorithm SHA256 .\GameTrainerHub-Setup-x64.exe
   ```
3. 双击运行 `GameTrainerHub-Setup-x64.exe`，按照向导选择安装目录（默认 `C:\Program Files\GameTrainerHub`）并完成安装。

## 覆盖升级

1. 直接运行新版安装包 `GameTrainerHub-Setup-x64.exe`。
2. 点击「安装」时，若检测到 `GameTrainerHub.exe` 正在运行，安装向导将弹出“程序正在运行”确认框；点击「是」将自动优雅关闭旧进程并继续升级。
3. 升级不会覆盖或清空 `%APPDATA%\GameTrainerHub` 下的 SQLite 数据库、下载历史与配置文件。

## 卸载说明

* 通过 Windows「设置 -> 应用」或安装目录下的 `Uninstall.exe` 即可卸载。
* 默认情况下，卸载程序会保留 `%APPDATA%\GameTrainerHub` 中的用户数据与已下载的修改器文件。
