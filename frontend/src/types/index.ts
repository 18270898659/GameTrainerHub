export interface Game {
  id: number
  slug: string
  name_en: string
  name_cn: string
  cover_url: string
  steam_app_id: number
  install_path: string
  aliases?: string[]
  is_favorite: boolean
  created_at: string
  updated_at: string
}

export interface Trainer {
  id: number
  game_id: number
  source_id: string
  external_id: string
  title_en: string
  title_cn: string
  detail_url: string
  author: string
  latest_version: string
  published_at: string
  updated_at: string
  is_deleted: boolean
  versions?: TrainerVersion[]
}

export interface TrainerVersion {
  id: number
  trainer_id: number
  version_name: string
  target_game_version: string
  features_text: string
  download_url: string
  file_name: string
  file_size: number
  sha256: string
  published_at: string
  is_downloaded: boolean
  local_path?: string
}

export interface SearchResult {
  game_id: number
  slug: string
  game_name_en: string
  game_name_cn: string
  cover_url: string
  source_id: string
  trainer_title: string
  detail_url: string
  version_name: string
  updated_at: string
  is_downloaded: boolean
  is_online_only: boolean
  score: number
}

export interface SearchFilter {
  query: string
  source: string
  local_only: boolean
  page: number
  page_size: number
}

export interface DownloadTask {
  id: number
  version_id: number
  game_name: string
  version_name: string
  file_name: string
  save_path: string
  total_bytes: number
  downloaded_bytes: number
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed'
  speed_bytes_per_sec: number
  progress: number
  error_msg: string
  retries: number
  created_at: string
  updated_at: string
}

export interface DownloadHistory {
  id: number
  task_id: number
  trainer_id: number
  version_id: number
  game_name: string
  file_name: string
  file_size: number
  sha256: string
  save_path: string
  completed_at: string
}

export interface LocalTrainer {
  id: number
  game_id: number
  version_id?: number
  file_path: string
  file_size: number
  sha256: string
  source_name: string
  is_custom: boolean
  last_run_at?: string
  run_count: number
  created_at: string
}

export interface LocalTrainerDetail {
  local_trainer: LocalTrainer
  game?: Game
  version?: TrainerVersion
}

export interface SecurityCardInfo {
  trainer_id: number
  game_name: string
  file_name: string
  file_path: string
  file_size: number
  sha256: string
  source_name: string
  is_custom: boolean
  security_notice: string
}

export interface AppConfig {
  base_dir: string
  data_dir: string
  download_dir: string
  log_dir: string
  max_concurrent_downloads: number
  auto_sync_interval_hours: number
  auto_sync_enabled: boolean
  proxy_url: string
  theme: string
  language: string
  log_level: string
  safe_launch_prompt: boolean
  close_to_tray: boolean
}

