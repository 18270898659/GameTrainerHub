package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_config_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// 1. 加载不存在的配置应生成默认配置并保存
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.MaxConcurrentDownloads != 3 {
		t.Errorf("expected MaxConcurrentDownloads=3, got %d", cfg.MaxConcurrentDownloads)
	}
	if cfg.Theme != "dark" {
		t.Errorf("expected Theme='dark', got '%s'", cfg.Theme)
	}
	if !cfg.CloseToTray {
		t.Errorf("expected CloseToTray=true by default, got %v", cfg.CloseToTray)
	}

	// 测试历史配置文件未包含 close_to_tray 时应默认兼容开启
	histConfigPath := filepath.Join(tempDir, "hist_config.json")
	_ = os.WriteFile(histConfigPath, []byte(`{"theme": "dark", "max_concurrent_downloads": 3}`), 0644)
	histCfg, err := LoadConfig(histConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig for hist config failed: %v", err)
	}
	if !histCfg.CloseToTray {
		t.Errorf("expected historical config without close_to_tray to default to true, got %v", histCfg.CloseToTray)
	}

	// 2. 检查生成的文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file was not created at %s", configPath)
	}

	// 3. 修改并保存
	err = cfg.Update(func(c *AppConfig) {
		c.MaxConcurrentDownloads = 5
		c.ProxyURL = "http://127.0.0.1:1080"
		c.Language = "en-US"
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 4. 重新加载并验证
	cfgReloaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig reload failed: %v", err)
	}
	if cfgReloaded.MaxConcurrentDownloads != 5 {
		t.Errorf("expected MaxConcurrentDownloads=5, got %d", cfgReloaded.MaxConcurrentDownloads)
	}
	if cfgReloaded.ProxyURL != "http://127.0.0.1:1080" {
		t.Errorf("expected ProxyURL='http://127.0.0.1:1080', got '%s'", cfgReloaded.ProxyURL)
	}
	if cfgReloaded.Language != "en-US" {
		t.Errorf("expected Language='en-US', got '%s'", cfgReloaded.Language)
	}
}

func TestEnsureDirectories(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_dirs_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &AppConfig{
		BaseDir:     tempDir,
		DataDir:     filepath.Join(tempDir, "sub_data"),
		DownloadDir: filepath.Join(tempDir, "sub_dl"),
		LogDir:      filepath.Join(tempDir, "sub_logs"),
	}

	if err := cfg.EnsureDirectories(); err != nil {
		t.Fatalf("EnsureDirectories failed: %v", err)
	}

	for _, d := range []string{cfg.DataDir, cfg.DownloadDir, cfg.LogDir} {
		if fi, err := os.Stat(d); os.IsNotExist(err) || !fi.IsDir() {
			t.Errorf("directory %s was not created", d)
		}
	}
}

func TestConfigBackupResetExportImport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// 1. 验证保存后生成 backup 文件
	err = cfg.Update(func(c *AppConfig) {
		c.MaxConcurrentDownloads = 6
		c.AutoBackupConfig = true
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	backupPath := filepath.Join(tempDir, "config.backup.json")
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("expected backup file %s to exist", backupPath)
	}

	// 2. 导出 JSON
	jsonStr, err := cfg.ExportConfigJSON()
	if err != nil {
		t.Fatalf("ExportConfigJSON failed: %v", err)
	}
	if len(jsonStr) == 0 {
		t.Fatal("expected non-empty JSON string")
	}

	// 3. 恢复默认设置
	err = cfg.ResetToDefault()
	if err != nil {
		t.Fatalf("ResetToDefault failed: %v", err)
	}
	if cfg.MaxConcurrentDownloads != 3 {
		t.Errorf("expected reset to 3, got %d", cfg.MaxConcurrentDownloads)
	}

	// 4. 从导出的 JSON 重新导入
	err = cfg.ImportConfigJSON(jsonStr)
	if err != nil {
		t.Fatalf("ImportConfigJSON failed: %v", err)
	}
	if cfg.MaxConcurrentDownloads != 6 {
		t.Errorf("expected imported 6, got %d", cfg.MaxConcurrentDownloads)
	}
}

