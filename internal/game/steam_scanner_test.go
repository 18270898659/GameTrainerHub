package game

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
)

const sampleVDF = `
"libraryfolders"
{
	"0"
	{
		"path"		"C:\\Program Files (x86)\\Steam"
		"label"		""
	}
	"1"
	{
		"path"		"D:\\SteamLibrary"
		"label"		"Games"
	}
}
`

const sampleACF = `
"AppState"
{
	"appid"		"1091500"
	"name"		"Cyberpunk 2077"
	"installdir"		"Cyberpunk 2077"
	"StateFlags"		"4"
}
`

func TestSteamScannerVDFAndACF(t *testing.T) {
	scanner := NewSteamScanner()

	// 1. 解析 libraryfolders.vdf
	folders := scanner.ParseLibraryFoldersVDF(sampleVDF, "C:\\DefaultSteam")
	if len(folders) != 2 {
		t.Fatalf("expected 2 folders, got %d", len(folders))
	}
	if !filepath.IsAbs(folders[0]) || !filepath.IsAbs(folders[1]) {
		t.Errorf("expected absolute paths, got %v", folders)
	}

	// 2. 解析 appmanifest ACF
	game := scanner.ParseAppManifestACF(sampleACF, "D:\\SteamLibrary")
	if game == nil {
		t.Fatal("expected non-nil discovered game")
	}
	if game.AppID != 1091500 {
		t.Errorf("expected AppID 1091500, got %d", game.AppID)
	}
	if game.Name != "Cyberpunk 2077" {
		t.Errorf("expected 'Cyberpunk 2077', got '%s'", game.Name)
	}
	if !filepath.IsAbs(game.InstallPath) {
		t.Errorf("expected absolute install path, got '%s'", game.InstallPath)
	}
}

func TestSteamScannerSync(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_steam_sync_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	gameRepo := repository.NewGameRepository(db)

	// 入库游戏
	game := &model.Game{Slug: "cyberpunk-2077", NameEN: "Cyberpunk 2077"}
	_ = gameRepo.Create(ctx, game)

	// 构造本地 mock Steam 目录
	steamDir := filepath.Join(tempDir, "Steam")
	steamapps := filepath.Join(steamDir, "steamapps")
	_ = os.MkdirAll(steamapps, 0755)
	_ = os.WriteFile(filepath.Join(steamDir, "steam.exe"), []byte("exe"), 0755)
	_ = os.WriteFile(filepath.Join(steamapps, "appmanifest_1091500.acf"), []byte(sampleACF), 0644)

	scanner := NewSteamScanner(steamDir)
	matched, err := scanner.SyncWithGameRepo(ctx, gameRepo)
	if err != nil {
		t.Fatalf("SyncWithGameRepo failed: %v", err)
	}
	if matched != 1 {
		t.Fatalf("expected 1 matched game, got %d", matched)
	}

	// 验证库中游戏已更新 AppID 与路径
	updated, _ := gameRepo.GetByID(ctx, game.ID)
	if updated.SteamAppID != 1091500 {
		t.Errorf("expected SteamAppID 1091500, got %d", updated.SteamAppID)
	}
}

