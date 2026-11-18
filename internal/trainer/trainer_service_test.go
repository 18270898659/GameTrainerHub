package trainer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
)

func setupTrainerTest(t *testing.T) (*TrainerService, *database.DB, func()) {
	tempDir, err := os.MkdirTemp("", "gth_trainer_svc_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	gameRepo := repository.NewGameRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	localRepo := repository.NewLocalTrainerRepository(db)

	svc := NewTrainerService(gameRepo, trainerRepo, localRepo)

	cleanup := func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}
	return svc, db, cleanup
}

func TestTrainerService(t *testing.T) {
	svc, db, cleanup := setupTrainerTest(t)
	defer cleanup()

	ctx := context.Background()
	gameRepo := repository.NewGameRepository(db)

	game := &model.Game{Slug: "black-myth-wukong", NameEN: "Black Myth: Wukong", NameCN: "黑神话：悟空"}
	_ = gameRepo.Create(ctx, game)

	tempDir, _ := os.MkdirTemp("", "gth_exe_test_*")
	defer os.RemoveAll(tempDir)

	mockExe := filepath.Join(tempDir, "BlackMythWukong_Trainer.exe")
	_ = os.WriteFile(mockExe, []byte("MOCK_EXE_BYTES"), 0755)

	// 1. 导入自定义修改器
	imported, err := svc.ImportCustomTrainer(ctx, game.ID, mockExe)
	if err != nil {
		t.Fatalf("ImportCustomTrainer failed: %v", err)
	}
	if !imported.IsCustom || imported.SourceName != "Custom" {
		t.Errorf("expected is_custom=true and source=Custom, got %+v", imported)
	}

	// 2. 查询列表
	list, err := svc.ListLocalTrainers(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 local trainer, got %d, err: %v", len(list), err)
	}
	if list[0].Game == nil || list[0].Game.NameCN != "黑神话：悟空" {
		t.Errorf("expected game name '黑神话：悟空', got %+v", list[0].Game)
	}

	// 3. 扫描文件夹
	scanned, err := svc.ScanFolderForTrainers(tempDir)
	if err != nil || len(scanned) != 1 {
		t.Fatalf("expected 1 scanned exe file, got %d, err: %v", len(scanned), err)
	}

	// 4. 切换收藏
	fav, err := svc.ToggleFavorite(ctx, game.ID)
	if err != nil || !fav {
		t.Fatalf("expected favorite to be true, got %v, err: %v", fav, err)
	}
	favAfter, _ := svc.ToggleFavorite(ctx, game.ID)
	if favAfter {
		t.Fatalf("expected favorite to be toggled to false, got %v", favAfter)
	}

	// 5. 删除本地修改器记录与文件
	if err := svc.DeleteLocalTrainer(ctx, imported.ID, true); err != nil {
		t.Fatalf("DeleteLocalTrainer failed: %v", err)
	}
	if _, err := os.Stat(mockExe); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted from disk")
	}
}

