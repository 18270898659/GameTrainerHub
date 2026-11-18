package model

import "time"

// 下载任务状态常量
const (
	DownloadStatusPending     = "pending"
	DownloadStatusDownloading = "downloading"
	DownloadStatusPaused      = "paused"
	DownloadStatusCompleted   = "completed"
	DownloadStatusFailed      = "failed"
)

// 同步状态常量
const (
	SyncStatusRunning = "running"
	SyncStatusSuccess = "success"
	SyncStatusFailed  = "failed"
)

// Source 资源网站源信息
type Source struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	BaseURL     string    `json:"base_url"`
	IsEnabled   bool      `json:"is_enabled"`
	RateLimitMs int       `json:"rate_limit_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

// Game 游戏基础信息
type Game struct {
	ID          int64       `json:"id"`
	Slug        string      `json:"slug"`
	NameEN      string      `json:"name_en"`
	NameCN      string      `json:"name_cn"`
	CoverURL    string      `json:"cover_url"`
	SteamAppID  int         `json:"steam_app_id"`
	InstallPath string      `json:"install_path"`
	Aliases     []string    `json:"aliases,omitempty"`
	IsFavorite  bool        `json:"is_favorite"`
	Trainers    []*Trainer  `json:"trainers,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// GameAlias 游戏别名与多语言表项
type GameAlias struct {
	ID             int64     `json:"id"`
	GameID         int64     `json:"game_id"`
	Alias          string    `json:"alias"`
	Lang           string    `json:"lang"` // 'zh', 'en', 'pinyin', 'initials'
	CreatedAt      time.Time `json:"created_at"`
}

// Trainer 修改器资源
type Trainer struct {
	ID            int64             `json:"id"`
	GameID        int64             `json:"game_id"`
	SourceID      string            `json:"source_id"`
	ExternalID    string            `json:"external_id"`
	TitleEN       string            `json:"title_en"`
	TitleCN       string            `json:"title_cn"`
	DetailURL     string            `json:"detail_url"`
	Author        string            `json:"author"`
	LatestVersion string            `json:"latest_version"`
	PublishedAt   time.Time         `json:"published_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	IsDeleted     bool              `json:"is_deleted"`
	Versions      []*TrainerVersion `json:"versions,omitempty"`
}

// TrainerVersion 修改器具体版本
type TrainerVersion struct {
	ID                int64     `json:"id"`
	TrainerID         int64     `json:"trainer_id"`
	VersionName       string    `json:"version_name"`
	TargetGameVersion string    `json:"target_game_version"`
	FeaturesText      string    `json:"features_text"`
	DownloadURL       string    `json:"download_url"`
	FileName          string    `json:"file_name"`
	FileSize          int64     `json:"file_size"`
	SHA256            string    `json:"sha256"`
	PublishedAt       time.Time `json:"published_at"`
	IsDownloaded      bool      `json:"is_downloaded"`
	LocalPath         string    `json:"local_path,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID               int64     `json:"id"`
	VersionID        int64     `json:"version_id"`
	GameName         string    `json:"game_name"`
	VersionName      string    `json:"version_name"`
	FileName         string    `json:"file_name"`
	SavePath         string    `json:"save_path"`
	TotalBytes       int64     `json:"total_bytes"`
	DownloadedBytes  int64     `json:"downloaded_bytes"`
	Status           string    `json:"status"` // pending, downloading, paused, completed, failed
	SpeedBytesPerSec int64     `json:"speed_bytes_per_sec"`
	Progress         float64   `json:"progress"`
	ErrorMsg         string    `json:"error_msg"`
	Retries          int       `json:"retries"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DownloadHistory 下载历史
type DownloadHistory struct {
	ID          int64     `json:"id"`
	TaskID      int64     `json:"task_id"`
	TrainerID   int64     `json:"trainer_id"`
	VersionID   int64     `json:"version_id"`
	GameName    string    `json:"game_name"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	SHA256      string    `json:"sha256"`
	SavePath    string    `json:"save_path"`
	CompletedAt time.Time `json:"completed_at"`
}

// LocalTrainer 本地已下载/导入的修改器
type LocalTrainer struct {
	ID          int64     `json:"id"`
	GameID      int64     `json:"game_id"`
	VersionID   *int64    `json:"version_id,omitempty"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	SHA256      string    `json:"sha256"`
	SourceName  string    `json:"source_name"`
	IsCustom    bool      `json:"is_custom"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	RunCount    int       `json:"run_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// Favorite 游戏收藏
type Favorite struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"game_id"`
	CreatedAt time.Time `json:"created_at"`
}

// SyncRecord 资源同步日志
type SyncRecord struct {
	ID           int64      `json:"id"`
	SourceID     string     `json:"source_id"`
	SyncType     string     `json:"sync_type"` // 'full', 'incremental', 'catalog'
	Status       string     `json:"status"`    // 'running', 'success', 'failed'
	ItemsAdded   int        `json:"items_added"`
	ItemsUpdated int        `json:"items_updated"`
	ErrorMsg     string     `json:"error_msg"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

// AppSetting 客户端设置键值对
type AppSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchFilter 搜索过滤参数
type SearchFilter struct {
	Query     string `json:"query"`
	Source    string `json:"source"`     // 'all' 或具体 source_id
	LocalOnly bool   `json:"local_only"` // 仅搜本地
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

// SearchResult 统一搜索结果
type SearchResult struct {
	GameID       int64     `json:"game_id"`
	Slug         string    `json:"slug"`
	GameNameEN   string    `json:"game_name_en"`
	GameNameCN   string    `json:"game_name_cn"`
	CoverURL     string    `json:"cover_url"`
	SourceID     string    `json:"source_id"`
	TrainerTitle string    `json:"trainer_title"`
	DetailURL    string    `json:"detail_url"`
	VersionName  string    `json:"version_name"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsDownloaded bool      `json:"is_downloaded"`
	IsOnlineOnly bool      `json:"is_online_only"`
	Score        float64   `json:"score"`
}

