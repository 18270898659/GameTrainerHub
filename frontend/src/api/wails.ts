import type {
  SearchResult,
  SearchFilter,
  Trainer,
  TrainerVersion,
  DownloadTask,
  DownloadHistory,
  LocalTrainerDetail,
  LocalTrainer,
  SecurityCardInfo,
  AppConfig,
  Game,
  IntegrityResult,
  DiagnosticReport,
  LogEntry,
  UpdateCheckResult
} from '../types'
import { DEFAULT_SEED_LATEST } from '../utils/seedData'

// 声明全局 window 上的 Wails 桥接对象
declare global {
  interface Window {
    go?: {
      app?: {
        App?: {
          Search: (filter: SearchFilter) => Promise<SearchResult[]>
          GetSearchSuggestions: (query: string) => Promise<SearchResult[]>
          GetTrainerDetail: (detailURL: string) => Promise<any>
          GetGameDetail: (gameID: number) => Promise<any>
          GetLatestTrainers: (page: number) => Promise<SearchResult[]>
          GetGameCount: () => Promise<number>
          GetLocalTrainers: () => Promise<LocalTrainerDetail[]>
          ImportCustomTrainer: (gameID: number, filePath: string) => Promise<LocalTrainer>
          DeleteLocalTrainer: (id: number, deleteFile: boolean) => Promise<void>
          ToggleFavorite: (gameID: number) => Promise<boolean>
          ToggleFavoriteBySlug: (slug: string, nameEN: string, nameCN: string, coverURL: string) => Promise<boolean>
          ListFavorites: () => Promise<Game[]>
          StartDownload: (versionID: number) => Promise<DownloadTask>
          RedownloadLocalTrainer: (localTrainerID: number) => Promise<DownloadTask>
          PauseDownload: (taskID: number) => Promise<void>
          ResumeDownload: (taskID: number) => Promise<void>
          CancelDownload: (taskID: number) => Promise<void>
          ListDownloadTasks: () => Promise<DownloadTask[]>
          ListDownloadHistory: (limit: number) => Promise<DownloadHistory[]>
          DeleteDownloadHistory: (id: number, deletePhysicalFile: boolean) => Promise<void>
          ClearDownloadHistory: () => Promise<void>
          GetSecurityCard: (localTrainerID: number) => Promise<SecurityCardInfo>
          LaunchTrainer: (localTrainerID: number, runAsAdmin: boolean) => Promise<void>
          OpenContainingFolder: (filePath: string) => Promise<void>
          ScanSteamGames: () => Promise<number>
          TriggerSync: () => Promise<any>
          GetConfig: () => Promise<AppConfig>
          SaveConfig: (cfg: AppConfig) => Promise<void>
          SelectDownloadDirectory: () => Promise<string>

          PauseAllDownloads: () => Promise<void>
          ResumeAllDownloads: () => Promise<void>
          GetSearchHistory: (limit: number) => Promise<string[]>
          AddSearchHistory: (keyword: string) => Promise<void>
          RemoveSearchHistory: (keyword: string) => Promise<void>
          ClearSearchHistory: () => Promise<void>
          RecordGameVisit: (gameID: number) => Promise<void>
          GetRecentGames: (limit: number) => Promise<Game[]>
          GetRecentTrainers: (limit: number) => Promise<LocalTrainerDetail[]>
          VerifyTrainerIntegrity: (localTrainerID: number) => Promise<IntegrityResult>
          RelocateTrainer: (localTrainerID: number, newFilePath: string) => Promise<void>
          IsTrainerRunning: (localTrainerID: number) => Promise<boolean>
          StopTrainer: (localTrainerID: number) => Promise<void>
          SelectTrainerExecutable: () => Promise<string>
          CheckHotkeyConflict: (hotkeyStr: string) => Promise<boolean>
          ApplyHotkeys: (toggleWindow: string, openDownloads: string, enabled: boolean) => Promise<void>
          ResetHotkeys: () => Promise<Record<string, string>>
          SuspendHotkeys: (suspend: boolean) => Promise<void>
          GetDefaultHotkeys: () => Promise<Record<string, string>>
          CheckForUpdates: () => Promise<UpdateCheckResult>
          DiagnoseEnvironment: () => Promise<DiagnosticReport>
          ExportDiagnostics: () => Promise<string>
          GetRecentLogs: (limit: number, category: string) => Promise<LogEntry[]>
          ExportConfig: () => Promise<string>
          ImportConfig: (jsonStr: string) => Promise<void>
          ResetConfig: () => Promise<void>
        }
      }
    }
    runtime?: {
      EventsOn: (eventName: string, callback: (data: any) => void) => () => void
    }
  }
}

const isWails = () => typeof window !== 'undefined' && !!window.go?.app?.App

/**
 * 等待 Wails 运行时注入就绪，避免组件早于 IPC 绑定加载导致返回空
 */
async function waitForWails(timeoutMs = 2000): Promise<boolean> {
  if (isWails()) return true
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    if (isWails()) return true
    await new Promise((resolve) => setTimeout(resolve, 30))
  }
  return isWails()
}

export const api = {
  async search(filter: SearchFilter): Promise<SearchResult[]> {
    await waitForWails(1000)
    if (isWails()) {
      return window.go!.app!.App!.Search(filter)
    }
    const q = (filter.query || '').trim().toLowerCase()
    if (!q) return DEFAULT_SEED_LATEST
    return DEFAULT_SEED_LATEST.filter(
      (it) => it.game_name_en.toLowerCase().includes(q) || it.game_name_cn.includes(q)
    )
  },

  async getSearchSuggestions(query: string): Promise<SearchResult[]> {
    await waitForWails(1000)
    if (isWails()) {
      return window.go!.app!.App!.GetSearchSuggestions(query)
    }
    const q = (query || '').trim().toLowerCase()
    if (!q) return DEFAULT_SEED_LATEST.slice(0, 6)
    return DEFAULT_SEED_LATEST.filter(
      (it) => it.game_name_en.toLowerCase().includes(q) || it.game_name_cn.includes(q)
    ).slice(0, 6)
  },

  async getTrainerDetail(detailURL: string): Promise<{ trainer: Trainer; versions: TrainerVersion[] }> {
    if (isWails()) {
      const res = await window.go!.app!.App!.GetTrainerDetail(detailURL)
      if (!res) {
        return {
          trainer: {
            id: 0,
            game_id: 0,
            source_id: 'fling',
            external_id: '',
            title_en: '',
            title_cn: '',
            detail_url: detailURL,
            author: 'FLiNG',
            latest_version: '',
            published_at: '',
            updated_at: '',
            is_deleted: false
          },
          versions: []
        }
      }

      // 兼容结构：支持直接返回 Trainer 对象（含 versions 数组）或旧版元组数组结构
      if (Array.isArray(res)) {
        const tr = res[0] || {}
        const vers = res[1] || tr.versions || []
        return { trainer: tr, versions: Array.isArray(vers) ? vers : [] }
      }

      return {
        trainer: res,
        versions: Array.isArray(res.versions) ? res.versions : []
      }
    }
    return {
      trainer: {
        id: 1,
        game_id: 1,
        source_id: 'fling',
        external_id: 'demo',
        title_en: 'Demo Game Trainer',
        title_cn: '演示游戏修改器',
        detail_url: detailURL,
        author: 'FLiNG',
        latest_version: 'v1.0 Plus 10',
        published_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        is_deleted: false
      },
      versions: []
    }
  },

  async getGameDetail(gameID: number): Promise<{ trainer: Trainer; versions: TrainerVersion[] }> {
    if (isWails()) {
      const res = await window.go!.app!.App!.GetGameDetail(gameID)
      if (!res) {
        throw new Error('未找到游戏详情')
      }
      if (Array.isArray(res)) {
        const tr = res[0] || {}
        const vers = res[1] || tr.versions || []
        return { trainer: tr, versions: Array.isArray(vers) ? vers : [] }
      }
      return {
        trainer: res,
        versions: Array.isArray(res.versions) ? res.versions : []
      }
    }
    throw new Error('Not running in Wails desktop')
  },

  async getLatestTrainers(page: number = 1): Promise<SearchResult[]> {
    await waitForWails(2000)
    if (isWails()) {
      try {
        const res = await window.go!.app!.App!.GetLatestTrainers(page)
        if (Array.isArray(res) && res.length > 0) {
          return res
        }
      } catch (err) {
        console.warn('Wails GetLatestTrainers failed, fallback to seed data', err)
      }
    }
    return DEFAULT_SEED_LATEST
  },

  async getGameCount(): Promise<number> {
    await waitForWails(1500)
    if (isWails()) {
      try {
        const count = await window.go!.app!.App!.GetGameCount()
        if (count && count > 0) {
          return count
        }
      } catch (err) {
        console.warn('Wails GetGameCount failed', err)
      }
    }
    return 767
  },

  async getLocalTrainers(): Promise<LocalTrainerDetail[]> {
    if (isWails()) {
      return window.go!.app!.App!.GetLocalTrainers()
    }
    return []
  },

  async importCustomTrainer(gameID: number, filePath: string): Promise<LocalTrainer> {
    if (isWails()) {
      return window.go!.app!.App!.ImportCustomTrainer(gameID, filePath)
    }
    throw new Error('Not running in Wails desktop')
  },

  async deleteLocalTrainer(id: number, deleteFile: boolean = false): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.DeleteLocalTrainer(id, deleteFile)
    }
  },

  async toggleFavorite(gameID: number): Promise<boolean> {
    if (isWails()) {
      return window.go!.app!.App!.ToggleFavorite(gameID)
    }
    return false
  },

  async toggleFavoriteBySlug(slug: string, nameEN: string, nameCN: string, coverURL: string): Promise<boolean> {
    if (isWails()) {
      return window.go!.app!.App!.ToggleFavoriteBySlug(slug, nameEN, nameCN, coverURL)
    }
    return false
  },

  async listFavorites(): Promise<Game[]> {
    if (isWails()) {
      return window.go!.app!.App!.ListFavorites()
    }
    return []
  },

  async startDownload(versionID: number): Promise<DownloadTask> {
    if (isWails()) {
      return window.go!.app!.App!.StartDownload(versionID)
    }
    throw new Error('Not running in Wails desktop')
  },

  async pauseDownload(taskID: number): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.PauseDownload(taskID)
    }
  },

  async resumeDownload(taskID: number): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.ResumeDownload(taskID)
    }
  },

  async cancelDownload(taskID: number): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.CancelDownload(taskID)
    }
  },

  async listDownloadTasks(): Promise<DownloadTask[]> {
    if (isWails()) {
      const res = await window.go!.app!.App!.ListDownloadTasks()
      return Array.isArray(res) ? res : []
    }
    return []
  },

  async deleteDownloadTask(taskID: number): Promise<void> {
    if (isWails()) {
      return (window as any).go.app.App.DeleteDownloadTask(taskID)
    }
  },

  async listDownloadHistory(limit: number = 50): Promise<DownloadHistory[]> {
    if (isWails()) {
      const res = await window.go!.app!.App!.ListDownloadHistory(limit)
      return Array.isArray(res) ? res : []
    }
    return []
  },

  async deleteDownloadHistory(id: number, deletePhysicalFile: boolean = false): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.DeleteDownloadHistory(id, deletePhysicalFile)
    }
  },

  async clearDownloadHistory(): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.ClearDownloadHistory()
    }
  },

  async getSecurityCard(localTrainerID: number): Promise<SecurityCardInfo> {
    if (isWails()) {
      return window.go!.app!.App!.GetSecurityCard(localTrainerID)
    }
    throw new Error('Not running in Wails desktop')
  },

  async launchTrainer(localTrainerID: number, runAsAdmin: boolean): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.LaunchTrainer(localTrainerID, runAsAdmin)
    }
  },

  async openContainingFolder(filePath: string): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.OpenContainingFolder(filePath)
    }
  },

  async scanSteamGames(): Promise<number> {
    if (isWails()) {
      return window.go!.app!.App!.ScanSteamGames()
    }
    return 0
  },

  async triggerSync(): Promise<any> {
    if (isWails()) {
      return window.go!.app!.App!.TriggerSync()
    }
  },

  async getConfig(): Promise<AppConfig> {
    if (isWails()) {
      return window.go!.app!.App!.GetConfig()
    }
    return {
      base_dir: '',
      data_dir: '',
      download_dir: 'C:\\Downloads\\Trainers',
      log_dir: '',
      max_concurrent_downloads: 3,
      auto_sync_interval_hours: 24,
      auto_sync_enabled: true,
      proxy_url: '',
      theme: 'dark',
      language: 'zh-CN',
      log_level: 'info',
      safe_launch_prompt: true,
      close_to_tray: false,
      remember_window_size: true,
      window_width: 1280,
      window_height: 820,
      remember_window_position: false,
      window_x: 0,
      window_y: 0
    }
  },

  async saveConfig(cfg: AppConfig): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.SaveConfig(cfg)
    }
  },

  async selectDownloadDirectory(): Promise<string> {
    if (isWails()) {
      return window.go!.app!.App!.SelectDownloadDirectory()
    }
    return ''
  },

  async pauseAllDownloads(): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.PauseAllDownloads()
    }
  },

  async resumeAllDownloads(): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.ResumeAllDownloads()
    }
  },

  async getSearchHistory(limit: number = 20): Promise<string[]> {
    if (isWails()) {
      return window.go!.app!.App!.GetSearchHistory(limit)
    }
    return []
  },

  async addSearchHistory(keyword: string): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.AddSearchHistory(keyword)
    }
  },

  async removeSearchHistory(keyword: string): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.RemoveSearchHistory(keyword)
    }
  },

  async clearSearchHistory(): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.ClearSearchHistory()
    }
  },

  async recordGameVisit(gameID: number): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.RecordGameVisit(gameID)
    }
  },

  async getRecentGames(limit: number = 10): Promise<Game[]> {
    if (isWails()) {
      return window.go!.app!.App!.GetRecentGames(limit)
    }
    return []
  },

  async getRecentTrainers(limit: number = 10): Promise<LocalTrainerDetail[]> {
    if (isWails()) {
      return window.go!.app!.App!.GetRecentTrainers(limit)
    }
    return []
  },

  async verifyTrainerIntegrity(localTrainerID: number): Promise<IntegrityResult> {
    if (isWails()) {
      return window.go!.app!.App!.VerifyTrainerIntegrity(localTrainerID)
    }
    return {
      trainer_id: localTrainerID,
      file_path: '',
      file_exists: false,
      expected_sha256: '',
      actual_sha256: '',
      file_size: 0,
      is_valid: false,
      message: 'Not in desktop'
    }
  },

  async relocateTrainer(localTrainerID: number, newFilePath: string): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.RelocateTrainer(localTrainerID, newFilePath)
    }
  },

  async isTrainerRunning(localTrainerID: number): Promise<boolean> {
    if (isWails()) {
      return window.go!.app!.App!.IsTrainerRunning(localTrainerID)
    }
    return false
  },

  async stopTrainer(localTrainerID: number): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.StopTrainer(localTrainerID)
    }
  },

  async selectTrainerExecutable(): Promise<string> {
    if (isWails()) {
      return window.go!.app!.App!.SelectTrainerExecutable()
    }
    return ''
  },

  async checkHotkeyConflict(hotkeyStr: string): Promise<boolean> {
    if (isWails()) {
      return window.go!.app!.App!.CheckHotkeyConflict(hotkeyStr)
    }
    return false
  },

  async applyHotkeys(toggleWindow: string, openDownloads: string, enabled: boolean): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.ApplyHotkeys(toggleWindow, openDownloads, enabled)
    }
  },

  async resetHotkeys(): Promise<Record<string, string>> {
    if (isWails()) {
      return window.go!.app!.App!.ResetHotkeys()
    }
    return {
      hotkey_toggle_window: 'Alt + Ctrl + Q',
      hotkey_open_downloads: 'Alt + Ctrl + J'
    }
  },

  async suspendHotkeys(suspend: boolean): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.SuspendHotkeys(suspend)
    }
  },

  async getDefaultHotkeys(): Promise<Record<string, string>> {
    if (isWails()) {
      return window.go!.app!.App!.GetDefaultHotkeys()
    }
    return {
      hotkey_toggle_window: 'Alt + Ctrl + Q',
      hotkey_open_downloads: 'Alt + Ctrl + J'
    }
  },

  async checkForUpdates(): Promise<UpdateCheckResult> {
    if (isWails()) {
      return window.go!.app!.App!.CheckForUpdates()
    }
    return {
      has_update: false,
      current_version: '1.0.0',
      latest_version: '1.0.0',
      release_notes: '最新版本'
    }
  },

  async diagnoseEnvironment(): Promise<DiagnosticReport> {
    if (isWails()) {
      return window.go!.app!.App!.DiagnoseEnvironment()
    }
    return {
      generated_at: new Date().toISOString(),
      app_version: '1.0.0',
      os: 'windows',
      arch: 'amd64',
      download_dir: '',
      download_dir_writable: true,
      available_disk_bytes: 1024 * 1024 * 1024,
      db_path: '',
      db_integrity_ok: true,
      active_downloads: 0,
      local_trainers_count: 0,
      favorites_count: 0
    }
  },

  async exportDiagnostics(): Promise<string> {
    if (isWails()) {
      return window.go!.app!.App!.ExportDiagnostics()
    }
    return ''
  },

  async getRecentLogs(limit: number = 100, category: string = ''): Promise<LogEntry[]> {
    if (isWails()) {
      return window.go!.app!.App!.GetRecentLogs(limit, category)
    }
    return []
  },

  async exportConfig(): Promise<string> {
    if (isWails()) {
      return window.go!.app!.App!.ExportConfig()
    }
    return ''
  },

  async importConfig(jsonStr: string): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.ImportConfig(jsonStr)
    }
  },

  async resetConfig(): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.ResetConfig()
    }
  },

  async redownloadLocalTrainer(localTrainerID: number): Promise<DownloadTask> {
    if (isWails()) {
      return window.go!.app!.App!.RedownloadLocalTrainer(localTrainerID)
    }
    return {
      id: Date.now(),
      version_id: 0,
      game_name: '本地修改器',
      version_name: '',
      file_name: 'trainer.zip',
      save_path: '',
      total_bytes: 0,
      downloaded_bytes: 0,
      status: 'pending',
      speed_bytes_per_sec: 0,
      progress: 0,
      error_msg: '',
      retries: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
  },

  onDownloadProgress(callback: (task: DownloadTask) => void): (() => void) | undefined {
    if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
      return window.runtime.EventsOn('download:progress', callback)
    }
    return undefined
  },

  onDownloadCompleted(callback: (task: DownloadTask) => void): (() => void) | undefined {
    if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
      return window.runtime.EventsOn('download:completed', callback)
    }
    return undefined
  },

  onDownloadFailed(callback: (task: DownloadTask) => void): (() => void) | undefined {
    if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
      return window.runtime.EventsOn('download:failed', callback)
    }
    return undefined
  },

  onNavigate(callback: (route: string) => void): (() => void) | undefined {
    if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
      return window.runtime.EventsOn('navigate:downloads', () => callback('/downloads'))
    }
    return undefined
  }
}

