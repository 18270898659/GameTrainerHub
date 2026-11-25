# Changelog

All notable changes to **GameTrainerHub** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.1] - 2026-11-25

### Added
- **MIT License**: Added standard [LICENSE](LICENSE) file and updated repository license metadata.

### Fixed
- **GitHub Actions CI & Release Workflows**: Ordered frontend build (`npm ci && npm run build`) before `go vet ./...` so `//go:embed all:frontend/dist` resolves cleanly on fresh CI runners.

---

## [1.0.0] - 2026-11-25

### Added
- **Desktop Client Core**: Built Windows x64 desktop application with Go 1.26, Wails v2.17.1, Vue 3, TypeScript, and Element Plus.
- **Hybrid Search & Translation**: SQLite FTS5 full-text search merged with live FLiNG Trainer adapter scraping, supporting Chinese-English bidirectional lookup, common aliases, full pinyin, and initials matching.
- **Download Manager**: Concurrent download queue with HTTP Range resumable downloads, SHA-256 verification, automatic archive extraction, and persistent task recovery.
- **Local Library & Steam Integration**: Local Steam library scanning, multi-version trainer storage, favorites, and download history tracking.
- **Security Launch Guard**: Pre-launch security modal displaying executable source path, SHA-256 hash, and privilege/isolation warnings without silent elevation.
- **Global Hotkeys & System Tray**: Physical keyboard combination capture (`HotkeyInput`), system-level hotkey conflict detection, atomic rollback on failure, and native Windows tray integration.
- **Internationalization & Themes**: Simplified Chinese (`zh-CN`), Traditional Chinese (`zh-TW`), and English (`en-US`) UI with light/dark theme support.
- **Windows Installer & Portable Package**: NSIS x64 installer (`GameTrainerHub-Setup-x64.exe`) and portable ZIP (`GameTrainerHub-Portable-x64.zip`).

### Fixed
- **Windows 10 / Windows 11 Installer Process Detection**: Replaced fragile external PowerShell `-File` invocation with native Win32 API process enumeration (`CreateToolhelp32Snapshot`, `QueryFullProcessImageNameW`), canonical path normalization (`GetFullPathNameW`, `GetLongPathNameW`), a strict 3-state Fail-Safe model (`NOT_RUNNING`, `RUNNING`, `DETECTION_ERROR`), and pre-overwrite double verification in `SecCore`.
- **Global Hotkey Thread Affinity**: Dispatched Win32 `RegisterHotKey` / `UnregisterHotKey` calls via `SendMessageW` to the tray window owner thread, resolving cross-thread registration failures.
- **Static Analysis Safety**: Eliminated `uintptr` to `unsafe.Pointer` conversion in `internal/tray/tray_windows.go` so `go vet ./...` passes cleanly.
