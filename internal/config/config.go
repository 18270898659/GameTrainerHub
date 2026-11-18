package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// AppConfig 客户端配置
type AppConfig struct {
	BaseDir                string `json:"base_dir"`
	DataDir                string `json:"data_dir"`
	DownloadDir            string `json:"download_dir"`
	LogDir                 string `json:"log_dir"`
	MaxConcurrentDownloads int    `json:"max_concurrent_downloads"`
	AutoSyncIntervalHours  int    `json:"auto_sync_interval_hours"`
	AutoSyncEnabled        bool   `json:"auto_sync_enabled"`
	ProxyURL               string `json:"proxy_url"`
	Theme                  string `json:"theme"`                   // "dark", "light", "system"
	Language               string `json:"language"`                // "zh-CN", "en-US"
	LogLevel               string `json:"log_level"`               // "debug", "info", "warn", "error"
	SafeLaunchPrompt       bool   `json:"safe_launch_prompt"`      // 启动前必须安全风险弹窗确认
	CloseToTray            bool   `json:"close_to_tray"`           // 关闭窗口时最小化到系统托盘

	configFilePath string      `json:"-"`
	mu             sync.RWMutex `json:"-"`
}

var (
	globalConfig *AppConfig
	configOnce   sync.Once
)

// DefaultConfig 获取默认配置
func DefaultConfig() *AppConfig {
	var baseDir string
	appData := os.Getenv("APPDATA")
	if appData != "" {
		baseDir = filepath.Join(appData, "GameTrainerHub")
	} else {
		// 备用：工作目录下的 data 目录
		baseDir = "data"
	}

	return &AppConfig{
		BaseDir:                baseDir,
		DataDir:                filepath.Join(baseDir, "data"),
		DownloadDir:            filepath.Join(baseDir, "downloads"),
		LogDir:                 filepath.Join(baseDir, "logs"),
		MaxConcurrentDownloads: 3,
		AutoSyncIntervalHours:  24,
		AutoSyncEnabled:        true,
		ProxyURL:               "",
		Theme:                  "dark",
		Language:               "zh-CN",
		LogLevel:               "info",
		SafeLaunchPrompt:       true,
		CloseToTray:            false,
		configFilePath:         filepath.Join(baseDir, "config.json"),
	}
}

// GetConfig 获取单例全局配置
func GetConfig() *AppConfig {
	configOnce.Do(func() {
		cfg, err := LoadConfig("")
		if err != nil || cfg == nil {
			globalConfig = DefaultConfig()
		} else {
			globalConfig = cfg
		}
	})
	return globalConfig
}

// LoadConfig 从指定或默认路径加载配置
func LoadConfig(customPath string) (*AppConfig, error) {
	cfg := DefaultConfig()
	targetPath := cfg.configFilePath
	if customPath != "" {
		targetPath = customPath
		cfg.configFilePath = customPath
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在则尝试初始化保存一份
			_ = cfg.EnsureDirectories()
			_ = cfg.Save()
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 补全内部字段
	cfg.configFilePath = targetPath
	if cfg.BaseDir == "" {
		cfg.BaseDir = filepath.Dir(targetPath)
	}
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(cfg.BaseDir, "data")
	}
	if cfg.DownloadDir == "" {
		cfg.DownloadDir = filepath.Join(cfg.BaseDir, "downloads")
	}
	if cfg.LogDir == "" {
		cfg.LogDir = filepath.Join(cfg.BaseDir, "logs")
	}
	if cfg.MaxConcurrentDownloads <= 0 {
		cfg.MaxConcurrentDownloads = 3
	}

	return cfg, nil
}

// EnsureDirectories 确保数据、下载与日志目录存在
func (c *AppConfig) EnsureDirectories() error {
	dirs := []string{c.BaseDir, c.DataDir, c.DownloadDir, c.LogDir}
	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

// Save 将当前配置持久化到 JSON 文件
func (c *AppConfig) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.configFilePath == "" {
		if c.BaseDir != "" {
			c.configFilePath = filepath.Join(c.BaseDir, "config.json")
		} else {
			c.configFilePath = "config.json"
		}
	}

	if err := c.EnsureDirectories(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.configFilePath, data, 0644)
}

// Update 更新配置
func (c *AppConfig) Update(fn func(cfg *AppConfig)) error {
	c.mu.Lock()
	fn(c)
	c.mu.Unlock()
	return c.Save()
}

// GetDBPath 获取 SQLite 数据库文件绝对路径
func (c *AppConfig) GetDBPath() string {
	return filepath.Join(c.DataDir, "trainers.db")
}
