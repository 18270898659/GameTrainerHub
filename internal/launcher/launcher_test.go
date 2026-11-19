package launcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

func TestLauncherDuplicateLaunchAndStop(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_launcher_dup_*")
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

	game := &model.Game{Slug: "test-game", NameEN: "Test Game", NameCN: "测试游戏"}
	_ = gameRepo.Create(ctx, game)

	// 使用系统的 cmd.exe 复制为一个测试用修改器可执行文件
	systemCmd := filepath.Join(os.Getenv("WINDIR"), "System32", "cmd.exe")
	if _, err := os.Stat(systemCmd); err != nil {
		t.Skip("cmd.exe not found in Windows directory")
	}

	mockExe := filepath.Join(tempDir, "dummy_trainer.exe")
	cmdBytes, err := os.ReadFile(systemCmd)
	if err != nil {
		t.Fatalf("failed to read cmd.exe: %v", err)
	}
	if err := os.WriteFile(mockExe, cmdBytes, 0755); err != nil {
		t.Fatalf("failed to write dummy_trainer.exe: %v", err)
	}

	lt := &model.LocalTrainer{
		GameID:     game.ID,
		FilePath:   mockExe,
		FileSize:   int64(len(cmdBytes)),
		SourceName: "FLiNG",
	}
	_ = localRepo.Create(ctx, lt)

	launcher := NewLauncher(localRepo, gameRepo)

	// 1. 初次启动
	if err := launcher.LaunchExecutable(ctx, lt.ID, false); err != nil {
		t.Fatalf("failed to launch executable: %v", err)
	}

	// 2. 验证运行状态被追踪
	if !launcher.IsTrainerRunning(lt.ID) {
		t.Errorf("expected trainer to be marked running")
	}

	// 3. 验证重复启动被有效拦截
	dupErr := launcher.LaunchExecutable(ctx, lt.ID, false)
	if dupErr == nil {
		t.Errorf("expected duplicate launch to return error, got nil")
	} else if !strings.Contains(dupErr.Error(), "already running") {
		t.Errorf("expected error message to mention 'already running', got: %v", dupErr)
	}

	// 4. 停止修改器并验证状态清理
	if err := launcher.StopTrainer(lt.ID); err != nil {
		t.Fatalf("failed to stop trainer: %v", err)
	}
	if launcher.IsTrainerRunning(lt.ID) {
		t.Errorf("expected trainer to no longer be marked running after stop")
	}
}

func TestElevationErrorDetection(t *testing.T) {
	// 1. 测试 Windows 740 错误判定
	err740 := &os.PathError{
		Op:   "fork/exec",
		Path: `C:\Users\bigant\AppData\Roaming\Game Trainer Hub\download\Cyberpunk 2077\v1.03 Plus 10 Trainer.exe`,
		Err:  syscall.Errno(740),
	}
	if !isElevationRequiredError(err740) {
		t.Errorf("expected err740 to be recognized as elevation required")
	}

	// 2. 测试文本形式的提权错误
	textErr := fmt.Errorf("fork/exec C:\\test.exe: The requested operation requires elevation.")
	if !isElevationRequiredError(textErr) {
		t.Errorf("expected text elevation error to be recognized")
	}

	// 3. 测试普通错误不被误判为提权
	normErr := fmt.Errorf("file not found")
	if isElevationRequiredError(normErr) {
		t.Errorf("did not expect normal error to be recognized as elevation")
	}
}

func TestManifestDetection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_manifest_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 模拟带 requireAdministrator manifest 的 EXE
	adminExe := filepath.Join(tempDir, "admin_trainer.exe")
	contentWithAdmin := []byte("MZ....<requestedExecutionLevel level=\"requireAdministrator\" uiAccess=\"false\"/>....")
	if err := os.WriteFile(adminExe, contentWithAdmin, 0755); err != nil {
		t.Fatal(err)
	}
	if !HasRequireAdministratorManifest(adminExe) {
		t.Errorf("expected HasRequireAdministratorManifest to return true for admin exe")
	}

	// 模拟普通 EXE
	normalExe := filepath.Join(tempDir, "normal_tool.exe")
	contentNormal := []byte("MZ....<requestedExecutionLevel level=\"asInvoker\" uiAccess=\"false\"/>....")
	if err := os.WriteFile(normalExe, contentNormal, 0755); err != nil {
		t.Fatal(err)
	}
	if HasRequireAdministratorManifest(normalExe) {
		t.Errorf("expected HasRequireAdministratorManifest to return false for normal exe")
	}
}

func TestOpenContainingFolderRobustness(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_explorer_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	launcher := NewLauncher(nil, nil)

	// 1. 验证存在的文件打开 (含中文和空格)
	chineseSpaceDir := filepath.Join(tempDir, "测试 游戏 文件夹 (2024)")
	_ = os.MkdirAll(chineseSpaceDir, 0755)
	chineseExe := filepath.Join(chineseSpaceDir, "黑神话 悟空 修改器 v1.0.exe")
	_ = os.WriteFile(chineseExe, []byte("MZ_TEST"), 0755)

	if err := launcher.OpenContainingFolder(chineseExe); err != nil {
		t.Errorf("OpenContainingFolder failed on chinese/space file: %v", err)
	}

	// 2. 验证文件被删除但所在文件夹仍然存在时，优雅降级打开所在目录
	ghostFile := filepath.Join(chineseSpaceDir, "non_existent.exe")
	if err := launcher.OpenContainingFolder(ghostFile); err != nil {
		t.Errorf("OpenContainingFolder should gracefully open parent dir when file missing, got err: %v", err)
	}

	// 3. 验证连父目录也完全不存在时，返回明确且友好的错误提示
	impossiblePath := `Z:\definitely\non_existent_drive_99999\ghost.exe`
	err = launcher.OpenContainingFolder(impossiblePath)
	if err == nil {
		t.Errorf("expected error for non-existent path, got nil")
	} else if !strings.Contains(err.Error(), "不存在") {
		t.Errorf("expected error message to mention '不存在', got: %v", err)
	}

	// 4. 验证在 PATH 缺少系统目录时，getExplorerExePath() 依然能够稳定解析到 explorer.exe
	explorerPath := getExplorerExePath()
	if explorerPath == "" {
		t.Fatalf("expected non-empty explorer path")
	}
	if _, err := os.Stat(explorerPath); err != nil {
		t.Fatalf("resolved explorer path does not exist on disk: %s", explorerPath)
	}
}
