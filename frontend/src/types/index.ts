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
  is_favorite?: boolean
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
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'failed' | 'retrying'
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
  file_exists?: boolean
  is_running?: boolean
  target_game_version?: string
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
  requires_admin: boolean
  security_notice: string
}

export interface AppConfig {
  config_version?: number
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
  remember_window_size: boolean
  window_width: number
  window_height: number
  remember_window_position: boolean
  window_x: number
  window_y: number
  notify_on_download_complete?: boolean
  notify_on_download_failed?: boolean
  notify_on_trainer_update?: boolean
  global_hotkeys_enabled?: boolean
  hotkey_toggle_window?: string
  hotkey_open_downloads?: string
  auto_backup_config?: boolean
}

export interface IntegrityResult {
  trainer_id: number
  file_path: string
  file_exists: boolean
  expected_sha256: string
  actual_sha256: string
  file_size: number
  is_valid: boolean
  message: string
}

export interface NetworkDiagnosis {
  target_host: string
  dns_resolved: boolean
  ip_addresses: string
  http_status: number
  latency_ms: number
  is_reachable: boolean
  error_message?: string
}

export interface DiagnosticReport {
  generated_at: string
  app_version: string
  os: string
  arch: string
  download_dir: string
  download_dir_writable: boolean
  available_disk_bytes: number
  db_path: string
  db_integrity_ok: boolean
  network?: NetworkDiagnosis
  active_downloads: number
  local_trainers_count: number
  favorites_count: number
}

export interface LogEntry {
  timestamp: string
  level: string
  category: string
  message: string
  details?: string
}

export interface UpdateCheckResult {
  has_update: boolean
  current_version: string
  latest_version: string
  release_notes: string
  download_url?: string
  package_sha256?: string
  published_at?: string
}

export interface VisitedGame {
  game_id: number
  game_name_cn: string
  game_name_en: string
  cover_url: string
  visited_at: string
}

