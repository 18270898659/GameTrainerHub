package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"GameTrainerHub/internal/config"
	"GameTrainerHub/internal/model"
)

func TestAppIPC(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_app_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.AppConfig{
		BaseDir:                tempDir,
		DataDir:                filepath.Join(tempDir, "data"),
		DownloadDir:            filepath.Join(tempDir, "downloads"),
		LogDir:                 filepath.Join(tempDir, "logs"),
		MaxConcurrentDownloads: 2,
		Theme:                  "dark",
		Language:               "zh-CN",
		LogLevel:               "debug",
	}

	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}
	defer app.Shutdown(context.Background())

	ctx := context.Background()
	app.Startup(ctx)

	// 1. 测试配置读写
	curCfg := app.GetConfig()
	if curCfg.Theme != "dark" {
		t.Errorf("expected theme dark, got %s", curCfg.Theme)
	}

	curCfg.Theme = "light"
	if err := app.SaveConfig(curCfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	if app.GetConfig().Theme != "light" {
		t.Errorf("expected theme light after save, got %s", app.GetConfig().Theme)
	}

	// 2. 插入测试游戏以验证相关 IPC 方法
	game := &model.Game{Slug: "cyberpunk-2077", NameEN: "Cyberpunk 2077", NameCN: "赛博朋克2077"}
	_ = app.gameRepo.Create(ctx, game)
	_ = app.ftsRepo.IndexGame(ctx, game.ID, game.NameEN, game.NameCN, "2077", "saibopengke", "sbpk")

	// 3. 搜索测试
	results, err := app.Search(model.SearchFilter{Query: "赛博朋克", LocalOnly: true})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 || results[0].Slug != "cyberpunk-2077" {
		t.Fatalf("expected search result 'cyberpunk-2077', got %v", results)
	}

	// 4. 导入测试修改器
	mockExe := filepath.Join(tempDir, "CP2077_Trainer.exe")
	_ = os.WriteFile(mockExe, []byte("MZ_BINARY"), 0755)

	imported, err := app.ImportCustomTrainer(game.ID, mockExe)
	if err != nil {
		t.Fatalf("ImportCustomTrainer failed: %v", err)
	}
	if imported.ID == 0 {
		t.Fatal("expected non-zero imported ID")
	}

	// 5. 本地修改器列表
	locals, err := app.GetLocalTrainers()
	if err != nil || len(locals) != 1 {
		t.Fatalf("expected 1 local trainer, got %d, err: %v", len(locals), err)
	}

	// 6. 安全卡片
	card, err := app.GetSecurityCard(imported.ID)
	if err != nil {
		t.Fatalf("GetSecurityCard failed: %v", err)
	}
	if card.FileName != "CP2077_Trainer.exe" {
		t.Errorf("expected 'CP2077_Trainer.exe', got '%s'", card.FileName)
	}

	// 7. 收藏测试
	fav, err := app.ToggleFavorite(game.ID)
	if err != nil || !fav {
		t.Fatalf("expected favorite=true, got %v, err: %v", fav, err)
	}
	favList, err := app.ListFavorites()
	if err != nil || len(favList) != 1 {
		t.Fatalf("expected 1 favorite in list, got %d", len(favList))
	}
}

