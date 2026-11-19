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
  Game
} from '../types'

// 声明全局 window 上的 Wails 桥接对象
declare global {
  interface Window {
    go?: {
      app?: {
        App?: {
          Search: (filter: SearchFilter) => Promise<SearchResult[]>
          GetTrainerDetail: (detailURL: string) => Promise<any>
          GetLatestTrainers: (page: number) => Promise<SearchResult[]>
          GetLocalTrainers: () => Promise<LocalTrainerDetail[]>
          ImportCustomTrainer: (gameID: number, filePath: string) => Promise<LocalTrainer>
          DeleteLocalTrainer: (id: number, deleteFile: boolean) => Promise<void>
          ToggleFavorite: (gameID: number) => Promise<boolean>
          ListFavorites: () => Promise<Game[]>
          StartDownload: (versionID: number) => Promise<DownloadTask>
          PauseDownload: (taskID: number) => Promise<void>
          ResumeDownload: (taskID: number) => Promise<void>
          CancelDownload: (taskID: number) => Promise<void>
          ListDownloadTasks: () => Promise<DownloadTask[]>
          ListDownloadHistory: (limit: number) => Promise<DownloadHistory[]>
          GetSecurityCard: (localTrainerID: number) => Promise<SecurityCardInfo>
          LaunchTrainer: (localTrainerID: number, runAsAdmin: boolean) => Promise<void>
          OpenContainingFolder: (filePath: string) => Promise<void>
          ScanSteamGames: () => Promise<number>
          TriggerSync: () => Promise<any>
          GetConfig: () => Promise<AppConfig>
          SaveConfig: (cfg: AppConfig) => Promise<void>
        }
      }
    }
    runtime?: {
      EventsOn: (eventName: string, callback: (data: any) => void) => () => void
    }
  }
}

const isWails = () => typeof window !== 'undefined' && !!window.go?.app?.App

export const api = {
  async search(filter: SearchFilter): Promise<SearchResult[]> {
    if (isWails()) {
      return window.go!.app!.App!.Search(filter)
    }
    return []
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

  async getLatestTrainers(page: number = 1): Promise<SearchResult[]> {
    if (isWails()) {
      return window.go!.app!.App!.GetLatestTrainers(page)
    }
    return []
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
      return window.go!.app!.App!.ListDownloadTasks()
    }
    return []
  },

  async listDownloadHistory(limit: number = 50): Promise<DownloadHistory[]> {
    if (isWails()) {
      return window.go!.app!.App!.ListDownloadHistory(limit)
    }
    return []
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
      close_to_tray: false
    }
  },

  async saveConfig(cfg: AppConfig): Promise<void> {
    if (isWails()) {
      return window.go!.app!.App!.SaveConfig(cfg)
    }
  },

  onDownloadProgress(callback: (task: DownloadTask) => void): (() => void) | undefined {
    if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
      return window.runtime.EventsOn('download:progress', callback)
    }
    return undefined
  }
}

