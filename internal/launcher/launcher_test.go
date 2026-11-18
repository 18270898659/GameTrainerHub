package launcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
)

func TestLauncherSecurityCardAndLaunch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_launcher_test_*")
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
	localRepo := repository.NewLocalTrainerRepository(db)

	game := &model.Game{Slug: "elden-ring", NameEN: "ELDEN RING", NameCN: "艾尔登法环"}
	_ = gameRepo.Create(ctx, game)

	// 创建带特殊字符和单引号的文件路径测试
	specialDir := filepath.Join(tempDir, "Assassin's Creed (2024)")
	_ = os.MkdirAll(specialDir, 0755)
	mockExe := filepath.Join(specialDir, "EldenRing's Trainer.exe")
	_ = os.WriteFile(mockExe, []byte("MZ_MOCK_TRAINER_BINARY"), 0755)

	lt := &model.LocalTrainer{
		GameID:     game.ID,
		FilePath:   mockExe,
		FileSize:   22,
		SHA256:     "mock_hash",
		SourceName: "FLiNG",
	}
	_ = localRepo.Create(ctx, lt)

	launcher := NewLauncher(localRepo, gameRepo)

	// 1. 测试安全卡片获取
	card, err := launcher.GetSecurityCard(ctx, lt.ID)
	if err != nil {
		t.Fatalf("GetSecurityCard failed: %v", err)
	}

	if card.GameName != "艾尔登法环" {
		t.Errorf("expected '艾尔登法环', got '%s'", card.GameName)
	}
	if card.FileName != "EldenRing's Trainer.exe" {
		t.Errorf("expected \"EldenRing's Trainer.exe\", got '%s'", card.FileName)
	}
	if !strings.Contains(card.SecurityNotice, "安全警示") {
		t.Errorf("expected security notice to contain '安全警示', got '%s'", card.SecurityNotice)
	}

	// 2. 验证非 .exe 拒绝启动
	notExe := filepath.Join(tempDir, "malicious.bat")
	_ = os.WriteFile(notExe, []byte("echo hi"), 0755)
	ltNotExe := &model.LocalTrainer{GameID: game.ID, FilePath: notExe}
	_ = localRepo.Create(ctx, ltNotExe)

	err = launcher.LaunchExecutable(ctx, ltNotExe.ID, false)
	if err == nil {
		t.Fatal("expected error launching non-exe file, got nil")
	}

	// 3. 验证启动记录自增机制
	if err := localRepo.RecordRun(ctx, lt.ID); err != nil {
		t.Fatalf("RecordRun failed: %v", err)
	}
	updatedLT, _ := localRepo.GetByFilePath(ctx, mockExe)
	if updatedLT.RunCount != 1 || updatedLT.LastRunAt == nil {
		t.Errorf("expected run count = 1 and valid last run at, got %d", updatedLT.RunCount)
	}

	// 4. 验证不存在的文件获取安全卡片失败
	ltGhost := &model.LocalTrainer{GameID: game.ID, FilePath: filepath.Join(tempDir, "ghost.exe")}
	_ = localRepo.Create(ctx, ltGhost)
	_, ghostErr := launcher.GetSecurityCard(ctx, ltGhost.ID)
	if ghostErr == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}
