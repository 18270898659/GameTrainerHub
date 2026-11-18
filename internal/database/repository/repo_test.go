package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/model"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	tempDir, err := os.MkdirTemp("", "gth_repo_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}
	return db, cleanup
}

func TestGameRepositoryCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewGameRepository(db)
	ctx := context.Background()

	// 1. Create
	g := &model.Game{
		Slug:        "cyberpunk-2077",
		NameEN:      "Cyberpunk 2077",
		NameCN:      "赛博朋克2077",
		CoverURL:    "https://example.com/cover.jpg",
		SteamAppID:  1091500,
		InstallPath: "C:\\Games\\Cyberpunk 2077",
	}
	if err := repo.Create(ctx, g); err != nil {
		t.Fatalf("failed to create game: %v", err)
	}
	if g.ID == 0 {
		t.Fatal("expected non-zero game ID")
	}

	// 2. GetByID
	fetched, err := repo.GetByID(ctx, g.ID)
	if err != nil {
		t.Fatalf("failed to get game by ID: %v", err)
	}
	if fetched.NameEN != "Cyberpunk 2077" || fetched.NameCN != "赛博朋克2077" {
		t.Errorf("unexpected game names: %s / %s", fetched.NameEN, fetched.NameCN)
	}

	// 3. GetBySlug
	bySlug, err := repo.GetBySlug(ctx, "cyberpunk-2077")
	if err != nil {
		t.Fatalf("failed to get game by slug: %v", err)
	}
	if bySlug.ID != g.ID {
		t.Errorf("expected game ID %d, got %d", g.ID, bySlug.ID)
	}

	// 4. Update
	g.NameCN = "赛博朋克 2077 (更新)"
	if err := repo.Update(ctx, g); err != nil {
		t.Fatalf("failed to update game: %v", err)
	}
	updated, _ := repo.GetByID(ctx, g.ID)
	if updated.NameCN != "赛博朋克 2077 (更新)" {
		t.Errorf("expected updated name, got '%s'", updated.NameCN)
	}

	// 5. Add Alias
	if err := repo.AddAlias(ctx, g.ID, "2077", "zh"); err != nil {
		t.Fatalf("failed to add alias: %v", err)
	}
	aliases, err := repo.GetAliases(ctx, g.ID)
	if err != nil || len(aliases) != 1 || aliases[0].Alias != "2077" {
		t.Fatalf("expected 1 alias '2077', got %v", aliases)
	}

	// 6. Favorites
	if err := repo.SetFavorite(ctx, g.ID, true); err != nil {
		t.Fatalf("failed to set favorite: %v", err)
	}
	favs, err := repo.ListFavorites(ctx)
	if err != nil || len(favs) != 1 {
		t.Fatalf("expected 1 favorite, got %d", len(favs))
	}
	if err := repo.SetFavorite(ctx, g.ID, false); err != nil {
		t.Fatalf("failed to unset favorite: %v", err)
	}
	favsAfter, _ := repo.ListFavorites(ctx)
	if len(favsAfter) != 0 {
		t.Fatalf("expected 0 favorites, got %d", len(favsAfter))
	}

	// 7. UpsertBySlug
	g2 := &model.Game{
		Slug:   "cyberpunk-2077",
		NameEN: "Cyberpunk 2077 Definitive",
	}
	upserted, err := repo.UpsertBySlug(ctx, g2)
	if err != nil {
		t.Fatalf("failed to upsert game: %v", err)
	}
	if upserted.ID != g.ID || upserted.NameEN != "Cyberpunk 2077 Definitive" {
		t.Errorf("unexpected upsert result: ID=%d, Name=%s", upserted.ID, upserted.NameEN)
	}
}

func TestTrainerRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	gameRepo := NewGameRepository(db)
	trainerRepo := NewTrainerRepository(db)
	ctx := context.Background()

	// 创建测试游戏
	game := &model.Game{Slug: "elden-ring", NameEN: "ELDEN RING", NameCN: "艾尔登法环"}
	_ = gameRepo.Create(ctx, game)

	// 1. 创建修改器
	trainer := &model.Trainer{
		GameID:        game.ID,
		SourceID:      "fling",
		ExternalID:    "elden-ring-trainer-1234",
		TitleEN:       "ELDEN RING Trainer",
		TitleCN:       "艾尔登法环 修改器",
		DetailURL:     "https://flingtrainer.com/trainer/elden-ring-trainer/",
		Author:        "FLiNG",
		LatestVersion: "v1.12 Plus 34",
		PublishedAt:   time.Now(),
	}
	if err := trainerRepo.CreateTrainer(ctx, trainer); err != nil {
		t.Fatalf("failed to create trainer: %v", err)
	}

	// 2. 创建版本
	v := &model.TrainerVersion{
		TrainerID:         trainer.ID,
		VersionName:       "v1.12 Plus 34",
		TargetGameVersion: "v1.12",
		FeaturesText:      "Num 1 - Infinite Health\nNum 2 - Infinite FP",
		DownloadURL:       "https://flingtrainer.com/downloads/sample_token",
		FileName:          "Elden.Ring.v1.12.Plus.34.Trainer-FLiNG.zip",
		FileSize:          1024567,
		SHA256:            "abc123def456",
		PublishedAt:       time.Now(),
	}
	if err := trainerRepo.CreateVersion(ctx, v); err != nil {
		t.Fatalf("failed to create version: %v", err)
	}

	// 3. 验证查询
	fetchedTrainer, err := trainerRepo.GetTrainerByExternalID(ctx, "fling", "elden-ring-trainer-1234")
	if err != nil {
		t.Fatalf("failed to get trainer: %v", err)
	}
	if len(fetchedTrainer.Versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(fetchedTrainer.Versions))
	}
	if fetchedTrainer.Versions[0].VersionName != "v1.12 Plus 34" {
		t.Errorf("unexpected version name: %s", fetchedTrainer.Versions[0].VersionName)
	}

	// 4. 验证事务级 UpsertTrainerWithVersions
	v2 := &model.TrainerVersion{
		VersionName:       "v1.13 Plus 35",
		TargetGameVersion: "v1.13",
		DownloadURL:       "https://flingtrainer.com/downloads/sample_token_2",
		FileName:          "Elden.Ring.v1.13.Plus.35.Trainer-FLiNG.zip",
	}
	trainer.LatestVersion = "v1.13 Plus 35"
	err = trainerRepo.UpsertTrainerWithVersions(ctx, trainer, []*model.TrainerVersion{v, v2})
	if err != nil {
		t.Fatalf("failed to upsert trainer with versions: %v", err)
	}

	versionsAfter, _ := trainerRepo.ListVersionsByTrainerID(ctx, trainer.ID)
	if len(versionsAfter) != 2 {
		t.Fatalf("expected 2 versions after upsert, got %d", len(versionsAfter))
	}
}

func TestDownloadAndLocalTrainerRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	gameRepo := NewGameRepository(db)
	trainerRepo := NewTrainerRepository(db)
	dlRepo := NewDownloadRepository(db)
	localRepo := NewLocalTrainerRepository(db)
	ctx := context.Background()

	game := &model.Game{Slug: "rdr2", NameEN: "Red Dead Redemption 2"}
	_ = gameRepo.Create(ctx, game)
	trainer := &model.Trainer{GameID: game.ID, SourceID: "fling", ExternalID: "rdr2-trainer", TitleEN: "RDR2 Trainer"}
	_ = trainerRepo.CreateTrainer(ctx, trainer)
	ver := &model.TrainerVersion{TrainerID: trainer.ID, VersionName: "v1.0 Plus 20", DownloadURL: "https://fling.com/dl", FileName: "rdr2.zip"}
	_ = trainerRepo.CreateVersion(ctx, ver)

	// 1. 创建下载任务
	task := &model.DownloadTask{
		VersionID:       ver.ID,
		FileName:        "rdr2.zip",
		SavePath:        "C:\\downloads\\rdr2.zip",
		TotalBytes:      5000000,
		DownloadedBytes: 1000000,
		Status:          model.DownloadStatusDownloading,
	}
	if err := dlRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("failed to create download task: %v", err)
	}

	// 2. 更新下载任务
	task.DownloadedBytes = 5000000
	task.Status = model.DownloadStatusCompleted
	if err := dlRepo.UpdateTask(ctx, task); err != nil {
		t.Fatalf("failed to update download task: %v", err)
	}

	fetchedTask, err := dlRepo.GetTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if fetchedTask.Progress < 99.9 {
		t.Errorf("expected 100%% progress, got %f", fetchedTask.Progress)
	}

	// 3. 记录下载历史
	hist := &model.DownloadHistory{
		TaskID:    task.ID,
		TrainerID: trainer.ID,
		VersionID: ver.ID,
		GameName:  "Red Dead Redemption 2",
		FileName:  "rdr2.zip",
		FileSize:  5000000,
		SHA256:    "hash123",
		SavePath:  task.SavePath,
	}
	if err := dlRepo.AddHistory(ctx, hist); err != nil {
		t.Fatalf("failed to add history: %v", err)
	}
	historyList, _ := dlRepo.ListHistory(ctx, 10)
	if len(historyList) != 1 {
		t.Fatalf("expected 1 history item, got %d", len(historyList))
	}

	// 4. 登记本地修改器
	local := &model.LocalTrainer{
		GameID:     game.ID,
		VersionID:  &ver.ID,
		FilePath:   "C:\\trainers\\rdr2_trainer.exe",
		FileSize:   4500000,
		SHA256:     "sha_local_123",
		SourceName: "FLiNG",
	}
	if err := localRepo.Create(ctx, local); err != nil {
		t.Fatalf("failed to create local trainer: %v", err)
	}

	// 5. 启动计数
	if err := localRepo.RecordRun(ctx, local.ID); err != nil {
		t.Fatalf("failed to record run: %v", err)
	}
	fetchedLocal, err := localRepo.GetByFilePath(ctx, local.FilePath)
	if err != nil {
		t.Fatalf("failed to get local trainer: %v", err)
	}
	if fetchedLocal.RunCount != 1 || fetchedLocal.LastRunAt == nil {
		t.Errorf("expected run_count=1 and valid last_run_at, got %d", fetchedLocal.RunCount)
	}
}

func TestSyncAndSettingRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	syncRepo := NewSyncRepository(db)
	settingRepo := NewSettingRepository(db)
	ctx := context.Background()

	// 1. 同步记录
	rec := &model.SyncRecord{
		SourceID:     "fling",
		SyncType:     "incremental",
		Status:       model.SyncStatusRunning,
		ItemsAdded:   10,
		ItemsUpdated: 5,
	}
	if err := syncRepo.CreateRecord(ctx, rec); err != nil {
		t.Fatalf("failed to create sync record: %v", err)
	}

	rec.Status = model.SyncStatusSuccess
	rec.ItemsAdded = 15
	if err := syncRepo.UpdateRecord(ctx, rec); err != nil {
		t.Fatalf("failed to update sync record: %v", err)
	}

	latest, err := syncRepo.GetLatestRecord(ctx, "fling")
	if err != nil || latest == nil {
		t.Fatalf("failed to get latest sync record: %v", err)
	}
	if latest.ItemsAdded != 15 {
		t.Errorf("expected items_added=15, got %d", latest.ItemsAdded)
	}

	// 2. 设置项
	if err := settingRepo.Set(ctx, "theme", "light"); err != nil {
		t.Fatalf("failed to set setting: %v", err)
	}
	val, err := settingRepo.Get(ctx, "theme")
	if err != nil || val != "light" {
		t.Fatalf("expected 'light', got '%s'", val)
	}

	allSettings, err := settingRepo.GetAll(ctx)
	if err != nil || allSettings["theme"] != "light" {
		t.Fatalf("failed to get all settings: %v", err)
	}
}

