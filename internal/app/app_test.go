package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"GameTrainerHub/internal/config"
	"GameTrainerHub/internal/logger"
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

	// 单元测试中静默事件推送，避免 Wails 检查非法 context
	app.EventEmitter = func(c context.Context, eventName string, data any) {}

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

	// 8. 详情获取与空值安全性验证
	_, emptyErr := app.GetTrainerDetail("")
	if emptyErr == nil {
		t.Error("expected error for empty detailURL, got nil")
	}

	// 创建已入库的修改器与版本
	mockTrainer := &model.Trainer{
		GameID:     game.ID,
		SourceID:   "fling",
		ExternalID: "cp2077-mock",
		TitleEN:    "Cyberpunk 2077 Trainer",
		DetailURL:  "https://flingtrainer.com/trainer/mock-cp2077/",
	}
	mockVer := &model.TrainerVersion{
		VersionName:  "v2.12 Plus 30",
		DownloadURL:  "https://flingtrainer.com/downloads/mock",
		FeaturesText: "Num 1 - Infinite Health",
	}
	_ = app.trainerRepo.UpsertTrainerWithVersions(ctx, mockTrainer, []*model.TrainerVersion{mockVer})

	// 验证降级查询本地库并能正确组装版本数据
	detail, err := app.GetTrainerDetail("https://flingtrainer.com/trainer/mock-cp2077/")
	if err != nil {
		t.Fatalf("expected to fetch mock trainer from DB, got: %v", err)
	}
	if detail == nil || len(detail.Versions) == 0 {
		t.Fatalf("expected detail with versions, got: %+v", detail)
	}
	if detail.Versions[0].VersionName != "v2.12 Plus 30" {
		t.Errorf("expected version 'v2.12 Plus 30', got: %s", detail.Versions[0].VersionName)
	}
}

func TestBlackMythWukongSearchDetailDownloadFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_bmw_test_*")
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

	// 单元测试中静默事件推送，避免 Wails 检查非法 context
	app.EventEmitter = func(c context.Context, eventName string, data any) {}

	ctx := context.Background()
	app.Startup(ctx)

	// 1. 搜索「黑神话：悟空」
	results, err := app.Search(model.SearchFilter{Query: "黑神话：悟空", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results for '黑神话：悟空'")
	}

	var bmwResult *model.SearchResult
	for _, r := range results {
		if strings.Contains(strings.ToLower(r.GameNameEN), "wukong") || strings.Contains(r.GameNameCN, "悟空") {
			bmwResult = r
			break
		}
	}
	if bmwResult == nil {
		t.Fatalf("expected to find Black Myth: Wukong in results, got %v", results)
	}
	t.Logf("Found search result: Title=%s, Slug=%s, DetailURL=%s", bmwResult.TrainerTitle, bmwResult.Slug, bmwResult.DetailURL)

	// 2. 获取游戏版本信息 (GetTrainerDetail)
	detail, err := app.GetTrainerDetail(bmwResult.DetailURL)
	if err != nil {
		t.Fatalf("GetTrainerDetail failed: %v", err)
	}
	if detail == nil {
		t.Fatal("expected detail not to be nil")
	}
	if len(detail.Versions) == 0 {
		t.Fatal("expected at least 1 version for Black Myth: Wukong")
	}

	firstVer := detail.Versions[0]
	t.Logf("Fetched version: ID=%d, Name=%s, FeaturesLen=%d", firstVer.ID, firstVer.VersionName, len(firstVer.FeaturesText))
	if firstVer.ID <= 0 {
		t.Errorf("expected version to have valid database ID (> 0), got %d", firstVer.ID)
	}
	if firstVer.DownloadURL == "" {
		t.Errorf("expected version download URL not to be empty")
	}
	if firstVer.FeaturesText == "" {
		t.Errorf("expected version features text to be populated")
	}

	// 3. 点击下载该版本修改器 (StartDownload)
	task, err := app.StartDownload(firstVer.ID)
	if err != nil {
		t.Fatalf("StartDownload failed: %v", err)
	}
	if task == nil || task.ID <= 0 {
		t.Fatalf("expected valid download task created, got: %+v", task)
	}
	t.Logf("Download task created successfully: TaskID=%d, Status=%s, File=%s", task.ID, task.Status, task.FileName)
}

func TestOtherGamesRegressionFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_reg_test_*")
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

	app.EventEmitter = func(c context.Context, eventName string, data any) {}
	ctx := context.Background()
	app.Startup(ctx)

	testCases := []struct {
		name    string
		query   string
		matchEN string
	}{
		{"Cyberpunk 2077", "赛博朋克2077", "cyberpunk"},
		{"Elden Ring", "老头环", "elden ring"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := app.Search(model.SearchFilter{Query: tc.query, Page: 1, PageSize: 5})
			if err != nil {
				t.Fatalf("[%s] Search failed: %v", tc.name, err)
			}
			if len(results) == 0 {
				t.Fatalf("[%s] expected results for '%s'", tc.name, tc.query)
			}

			target := results[0]
			if !strings.Contains(strings.ToLower(target.GameNameEN), tc.matchEN) {
				t.Errorf("[%s] expected game name to contain '%s', got '%s'", tc.name, tc.matchEN, target.GameNameEN)
			}

			// 获取版本详情
			detail, err := app.GetTrainerDetail(target.DetailURL)
			if err != nil {
				t.Fatalf("[%s] GetTrainerDetail failed: %v", tc.name, err)
			}
			if detail == nil || len(detail.Versions) == 0 {
				t.Fatalf("[%s] expected detail with versions", tc.name)
			}
			t.Logf("[%s] Versions count: %d, Latest: %s", tc.name, len(detail.Versions), detail.Versions[0].VersionName)

			// 创建下载任务
			task, err := app.StartDownload(detail.Versions[0].ID)
			if err != nil {
				t.Fatalf("[%s] StartDownload failed: %v", tc.name, err)
			}
			if task.ID <= 0 {
				t.Errorf("[%s] expected valid task id", tc.name)
			}
		})
	}
}

func TestWindowAndDirectorySettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_settings_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.AppConfig{
		BaseDir:                tempDir,
		DataDir:                filepath.Join(tempDir, "data"),
		DownloadDir:            filepath.Join(tempDir, "downloads"),
		LogDir:                 filepath.Join(tempDir, "logs"),
		MaxConcurrentDownloads: 3,
		RememberWindowSize:     true,
		WindowWidth:            1400,
		WindowHeight:           900,
		RememberWindowPosition: true,
		WindowX:                100,
		WindowY:                100,
		CloseToTray:            true,
		Language:               "zh-CN",
	}

	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}
	defer app.Shutdown(context.Background())

	// 1. 测试写权限校验
	writableDir := filepath.Join(tempDir, "new_downloads")
	if err := ValidateDirectoryWritable(writableDir); err != nil {
		t.Fatalf("ValidateDirectoryWritable failed on valid dir: %v", err)
	}

	// 2. 测试配置更新及并发数动态调整
	updatedCfg := *app.GetConfig()
	updatedCfg.MaxConcurrentDownloads = 5
	updatedCfg.CloseToTray = true
	updatedCfg.Language = "en-US"
	updatedCfg.RememberWindowSize = false

	if err := app.SaveConfig(&updatedCfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	saved := app.GetConfig()
	if saved.MaxConcurrentDownloads != 5 {
		t.Errorf("expected MaxConcurrentDownloads=5, got %d", saved.MaxConcurrentDownloads)
	}
	if saved.Language != "en-US" {
		t.Errorf("expected Language=en-US, got %s", saved.Language)
	}
	if saved.CloseToTray != true {
		t.Errorf("expected CloseToTray=true, got %v", saved.CloseToTray)
	}
	if saved.RememberWindowSize != false {
		t.Errorf("expected RememberWindowSize=false, got %v", saved.RememberWindowSize)
	}

	// 3. 测试屏幕有效性校验逻辑
	onScreen := isPositionOnScreen(100, 100, 1280, 820)
	t.Logf("isPositionOnScreen(100, 100, 1280, 820) = %v", onScreen)

	// 4. 测试 OnBeforeClose 托盘隐藏逻辑
	prevent := app.OnBeforeClose(context.Background())
	if !prevent {
		t.Errorf("expected OnBeforeClose to prevent close when CloseToTray=true, got %v", prevent)
	}

	// 5. 测试 Quit 后不拦截
	app.Quit()
	if !app.IsQuitting() {
		t.Errorf("expected isQuitting to be true")
	}
	preventAfterQuit := app.OnBeforeClose(context.Background())
	if preventAfterQuit {
		t.Errorf("expected OnBeforeClose to NOT prevent close when Quitting=true, got %v", preventAfterQuit)
	}
}

func TestPhase2IPCMethods(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_p2_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.AppConfig{
		BaseDir:                tempDir,
		DataDir:                filepath.Join(tempDir, "data"),
		DownloadDir:            filepath.Join(tempDir, "downloads"),
		LogDir:                 filepath.Join(tempDir, "logs"),
		MaxConcurrentDownloads: 3,
		Theme:                  "dark",
		Language:               "zh-CN",
		LogLevel:               "debug",
		CloseToTray:            true,
	}

	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}
	defer app.Shutdown(context.Background())

	// 1. 搜索历史 IPC
	if err := app.AddSearchHistory("黑神话：悟空"); err != nil {
		t.Fatalf("AddSearchHistory failed: %v", err)
	}
	if err := app.AddSearchHistory("艾尔登法环"); err != nil {
		t.Fatalf("AddSearchHistory failed: %v", err)
	}
	hist, err := app.GetSearchHistory(10)
	if err != nil || len(hist) != 2 {
		t.Fatalf("expected 2 history items, got %d (err: %v)", len(hist), err)
	}
	if err := app.RemoveSearchHistory("艾尔登法环"); err != nil {
		t.Fatalf("RemoveSearchHistory failed: %v", err)
	}
	histAfter, _ := app.GetSearchHistory(10)
	if len(histAfter) != 1 {
		t.Errorf("expected 1 history item after remove, got %d", len(histAfter))
	}
	_ = app.ClearSearchHistory()

	// 2. 访问历史 IPC
	game := &model.Game{Slug: "wukong", NameEN: "Black Myth: Wukong", NameCN: "黑神话：悟空"}
	_ = app.gameRepo.Create(context.Background(), game)
	if err := app.RecordGameVisit(game.ID); err != nil {
		t.Fatalf("RecordGameVisit failed: %v", err)
	}
	recentGames, err := app.GetRecentGames(5)
	if err != nil || len(recentGames) != 1 {
		t.Fatalf("expected 1 recent game, got %d (err: %v)", len(recentGames), err)
	}

	// 3. 诊断与日志 IPC
	logger.Info("Diagnostics test entry", "module", "test", "token", "super_secret_123")
	report, err := app.DiagnoseEnvironment()
	if err != nil {
		t.Fatalf("DiagnoseEnvironment failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil diagnostic report")
	}

	exportPath, err := app.ExportDiagnostics()
	if err != nil {
		t.Fatalf("ExportDiagnostics failed: %v", err)
	}
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Errorf("export file %s does not exist", exportPath)
	}

	logs, err := app.GetRecentLogs(50, "")
	if err != nil || len(logs) == 0 {
		t.Errorf("expected logs, got %d (err: %v)", len(logs), err)
	}

	// 4. 配置导出、重置与导入 IPC
	jsonExport, err := app.ExportConfig()
	if err != nil || len(jsonExport) == 0 {
		t.Fatalf("ExportConfig failed: %v", err)
	}
	if err := app.ResetConfig(); err != nil {
		t.Fatalf("ResetConfig failed: %v", err)
	}
	if err := app.ImportConfig(jsonExport); err != nil {
		t.Fatalf("ImportConfig failed: %v", err)
	}

	// 5. 检查更新 IPC
	updateRes, err := app.CheckForUpdates()
	if err != nil || updateRes == nil {
		t.Fatalf("CheckForUpdates failed: %v", err)
	}
	if updateRes.CurrentVersion != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", updateRes.CurrentVersion)
	}

	// 6. 批量下载暂停与恢复
	_ = app.PauseAllDownloads()
	_ = app.ResumeAllDownloads()
}

func TestDefectFixesSuite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_defect_test_*")
	if err != nil {
		t.Fatal(err)
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
	}

	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}
	defer app.Shutdown(context.Background())

	ctx := context.Background()

	// 1. 多游戏收藏测试 (Multi-game favorites)
	fav1, err := app.ToggleFavoriteBySlug("black-myth-wukong", "Black Myth: Wukong", "黑神话：悟空", "http://example.com/bmw.jpg")
	if err != nil || !fav1 {
		t.Fatalf("expected fav1 true, got %v (err: %v)", fav1, err)
	}

	fav2, err := app.ToggleFavoriteBySlug("elden-ring", "Elden Ring", "艾尔登法环", "http://example.com/er.jpg")
	if err != nil || !fav2 {
		t.Fatalf("expected fav2 true, got %v (err: %v)", fav2, err)
	}

	fav3, err := app.ToggleFavoriteBySlug("cyberpunk-2077", "Cyberpunk 2077", "赛博朋克2077", "http://example.com/cp.jpg")
	if err != nil || !fav3 {
		t.Fatalf("expected fav3 true, got %v (err: %v)", fav3, err)
	}

	favList, err := app.ListFavorites()
	if err != nil {
		t.Fatalf("ListFavorites failed: %v", err)
	}
	if len(favList) != 3 {
		t.Fatalf("expected 3 favorites, got %d", len(favList))
	}

	// 取消其中一个收藏 (Elden Ring)
	g2, _ := app.gameRepo.GetBySlug(ctx, "elden-ring")
	if g2 == nil {
		t.Fatal("expected elden-ring game exists")
	}
	favAfter, err := app.ToggleFavorite(g2.ID)
	if err != nil || favAfter {
		t.Fatalf("expected fav false after toggle, got %v", favAfter)
	}
	favList2, _ := app.ListFavorites()
	if len(favList2) != 2 {
		t.Fatalf("expected 2 favorites remaining, got %d", len(favList2))
	}

	// 2. 根据游戏 ID 查询详情 (GetGameDetail)
	g1, _ := app.gameRepo.GetBySlug(ctx, "black-myth-wukong")
	if g1 == nil {
		t.Fatal("expected black-myth-wukong game exists")
	}
	// 先关联一个修改器到 g1
	tr := &model.Trainer{
		GameID:      g1.ID,
		SourceID:    "fling",
		ExternalID:  "bmw-tr",
		TitleEN:     "Black Myth Wukong Trainer",
		TitleCN:     "黑神话：悟空 修改器",
		DetailURL:   "http://example.com/bmw",
		Author:      "FLiNG",
		PublishedAt: time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = app.trainerRepo.CreateTrainer(ctx, tr)
	ver := &model.TrainerVersion{
		TrainerID:   tr.ID,
		VersionName: "v1.0.8.14815",
		DownloadURL: "http://example.com/dl",
		FileName:    "bmw.zip",
		PublishedAt: time.Now(),
	}
	_ = app.trainerRepo.CreateVersion(ctx, ver)

	gameDetail, err := app.GetGameDetail(g1.ID)
	if err != nil || gameDetail == nil {
		t.Fatalf("GetGameDetail failed: %v", err)
	}
	if gameDetail.GameID != g1.ID {
		t.Errorf("expected GameID %d, got %d", g1.ID, gameDetail.GameID)
	}
	if len(gameDetail.Versions) == 0 {
		t.Errorf("expected versions in game detail")
	}

	// 3. 最近使用修改器过滤 (只展示真正启动过的修改器)
	recentBefore, err := app.GetRecentTrainers(10)
	if err != nil {
		t.Fatalf("GetRecentTrainers failed: %v", err)
	}
	if len(recentBefore) != 0 {
		t.Fatalf("expected 0 recent ran trainers, got %d", len(recentBefore))
	}

	// 导入一个修改器并模拟启动 (RecordRun)
	exePath := filepath.Join(tempDir, "test_trainer.exe")
	_ = os.WriteFile(exePath, []byte("fake exe content"), 0755)
	lt, err := app.ImportCustomTrainer(g1.ID, exePath)
	if err != nil || lt == nil {
		t.Fatalf("ImportCustomTrainer failed: %v", err)
	}
	// 未启动前，依然应为 0
	recentStillEmpty, _ := app.GetRecentTrainers(10)
	if len(recentStillEmpty) != 0 {
		t.Fatalf("expected 0 before run, got %d", len(recentStillEmpty))
	}
	// 启动一次
	_ = app.localRepo.RecordRun(ctx, lt.ID)
	recentAfterRun, _ := app.GetRecentTrainers(10)
	if len(recentAfterRun) != 1 {
		t.Fatalf("expected 1 recent trainer after run, got %d", len(recentAfterRun))
	}
	if recentAfterRun[0].LocalTrainer.ID != lt.ID {
		t.Errorf("expected lt ID %d, got %d", lt.ID, recentAfterRun[0].LocalTrainer.ID)
	}

	// 4. 下载历史与活动任务分离及删除历史
	histRecord := &model.DownloadHistory{
		TaskID:      101,
		TrainerID:   tr.ID,
		VersionID:   ver.ID,
		GameName:    "黑神话：悟空",
		TrainerName: "Black Myth Wukong Trainer",
		VersionName: "v1.0.8",
		FileName:    "bmw.zip",
		FileSize:    1024,
		SavePath:    exePath,
	}
	if err := app.dlRepo.AddHistory(ctx, histRecord); err != nil {
		t.Fatalf("AddHistory failed: %v", err)
	}
	histList, err := app.ListDownloadHistory(10)
	if err != nil || len(histList) == 0 {
		t.Fatalf("ListDownloadHistory failed: %v", err)
	}
	if !histList[0].FileExists {
		t.Errorf("expected file_exists true since exePath exists")
	}

	// 删除单条历史（默认不删除物理文件）
	if err := app.DeleteDownloadHistory(histList[0].ID, false); err != nil {
		t.Fatalf("DeleteDownloadHistory failed: %v", err)
	}
	histAfterDel, _ := app.ListDownloadHistory(10)
	if len(histAfterDel) != 0 {
		t.Errorf("expected 0 history after delete, got %d", len(histAfterDel))
	}
	// 验证物理文件依然保留
	if _, statErr := os.Stat(exePath); statErr != nil {
		t.Errorf("expected physical file to still exist after history deletion, got error: %v", statErr)
	}
}

func TestRedownloadLocalTrainerAndGetGameDetailFlow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_redownload_test_*")
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

	app.EventEmitter = func(c context.Context, eventName string, data any) {}
	ctx := context.Background()
	app.Startup(ctx)

	// 1. 创建测试游戏与修改器版本
	game, err := app.gameRepo.GetOrCreateBySlug(ctx, "black-myth-wukong", "Black Myth: Wukong", "黑神话：悟空", "https://example.com/cover.jpg")
	if err != nil || game == nil {
		t.Fatalf("GetOrCreateBySlug failed: %v", err)
	}

	tr := &model.Trainer{
		GameID:        game.ID,
		SourceID:      "fling",
		ExternalID:    "bmw-01",
		TitleEN:       "Black Myth: Wukong Trainer",
		TitleCN:       "黑神话：悟空 修改器",
		DetailURL:     "https://flingtrainer.com/trainer/black-myth-wukong-trainer/",
		LatestVersion: "v1.0.8",
	}
	ver := &model.TrainerVersion{
		VersionName:  "v1.0.8 Plus 35",
		DownloadURL:  "https://example.com/download/bmw.zip",
		FileName:     "bmw.zip",
		FileSize:     2048,
		SHA256:       "abc123456789",
		FeaturesText: "Inf Health\nInf Stamina",
	}
	if err := app.trainerRepo.UpsertTrainerWithVersions(ctx, tr, []*model.TrainerVersion{ver}); err != nil {
		t.Fatalf("UpsertTrainerWithVersions failed: %v", err)
	}

	versions, _ := app.trainerRepo.ListVersionsByTrainerID(ctx, tr.ID)
	if len(versions) == 0 {
		t.Fatalf("expected version in db")
	}
	verID := versions[0].ID

	// 2. 测试 GetGameDetail 正常通过 GameID 查询到详情与可用版本列表
	detailTr, err := app.GetGameDetail(game.ID)
	if err != nil || detailTr == nil {
		t.Fatalf("GetGameDetail failed: %v", err)
	}
	if len(detailTr.Versions) == 0 {
		t.Fatalf("expected versions populated from GetGameDetail")
	}

	// 3. 注册本地修改器记录
	fakeExePath := filepath.Join(cfg.DownloadDir, "bmw_trainer.exe")
	_ = os.WriteFile(fakeExePath, []byte("fake binary"), 0644)
	lt := &model.LocalTrainer{
		GameID:     game.ID,
		VersionID:  &verID,
		FilePath:   fakeExePath,
		FileSize:   11,
		SHA256:     "mockhash",
		SourceName: "FLiNG",
	}
	if err := app.localRepo.Upsert(ctx, lt); err != nil {
		t.Fatalf("Upsert local trainer failed: %v", err)
	}

	// 4. 模拟本地修改器文件被删除/隔离
	_ = os.Remove(fakeExePath)

	// 5. 调用 LaunchTrainer 预期捕获友好的文件缺失异常而非系统底层 fork/exec 崩溃
	launchErr := app.LaunchTrainer(lt.ID, false)
	if launchErr == nil {
		t.Fatalf("expected error when launching missing trainer file, got nil")
	}
	if !strings.Contains(launchErr.Error(), "不存在") && !strings.Contains(launchErr.Error(), "not exist") {
		t.Errorf("expected missing file friendly error, got: %v", launchErr)
	}

	// 6. 调用 RedownloadLocalTrainer 触发重新下载
	redownloadTask, err := app.RedownloadLocalTrainer(lt.ID)
	if err != nil {
		t.Fatalf("RedownloadLocalTrainer failed: %v", err)
	}
	if redownloadTask == nil || redownloadTask.VersionID != verID {
		t.Errorf("expected task for version %d, got %v", verID, redownloadTask)
	}

	// 7. 再次确认本地修改器记录依然完整保留（禁止静默删除）
	existingLT, err := app.localRepo.GetByID(ctx, lt.ID)
	if err != nil || existingLT == nil {
		t.Fatalf("expected local trainer record retained, got nil: %v", err)
	}
}

func TestSearchSuggestionsAndHistoryPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_suggest_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.AppConfig{
		BaseDir:     tempDir,
		DataDir:     filepath.Join(tempDir, "data"),
		DownloadDir: filepath.Join(tempDir, "download"),
		LogDir:      filepath.Join(tempDir, "logs"),
		Language:    "zh-CN",
	}

	app, err := NewApp(cfg)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}
	app.EventEmitter = func(ctx context.Context, eventName string, data any) {}
	ctx := context.Background()
	app.Startup(ctx)
	defer app.Shutdown(ctx)

	// 1. 建立测试游戏与修改器数据
	g := &model.Game{
		Slug:   "elden-ring",
		NameEN: "ELDEN RING",
		NameCN: "艾尔登法环",
	}
	if err := app.gameRepo.Create(ctx, g); err != nil {
		t.Fatalf("create game failed: %v", err)
	}
	_ = app.ftsRepo.IndexGame(ctx, g.ID, g.NameEN, g.NameCN, "老头环 法环", "aierdengfahuan", "aedfh")

	// 2. 测试 GetSearchSuggestions (中文及拼音建议)
	suggs, err := app.GetSearchSuggestions("老头环")
	if err != nil {
		t.Fatalf("GetSearchSuggestions failed: %v", err)
	}
	if len(suggs) == 0 {
		t.Errorf("expected suggestions for '老头环', got empty")
	} else if suggs[0].GameNameCN != "艾尔登法环" && suggs[0].GameNameEN != "ELDEN RING" {
		t.Errorf("unexpected suggestion result: %v", suggs[0])
	}

	// 3. 测试下载历史物理删除区分机制
	testFilePath := filepath.Join(tempDir, "download", "test_file.exe")
	_ = os.WriteFile(testFilePath, []byte("fake-exe"), 0644)
	hist := &model.DownloadHistory{
		TaskID:    101,
		TrainerID: 201,
		VersionID: 301,
		GameName:  "ELDEN RING",
		FileName:  "test_file.exe",
		SavePath:  testFilePath,
	}
	if err := app.dlRepo.AddHistory(ctx, hist); err != nil {
		t.Fatalf("AddHistory failed: %v", err)
	}

	// 3.1 默认删除历史不删文件
	if err := app.DeleteDownloadHistory(hist.ID, false); err != nil {
		t.Fatalf("DeleteDownloadHistory failed: %v", err)
	}
	if _, statErr := os.Stat(testFilePath); statErr != nil {
		t.Errorf("expected physical file to still exist, got: %v", statErr)
	}

	// 3.2 显式勾选删除文件
	_ = app.dlRepo.AddHistory(ctx, hist)
	if err := app.DeleteDownloadHistory(hist.ID, true); err != nil {
		t.Fatalf("DeleteDownloadHistory with deletePhysicalFile failed: %v", err)
	}
	if _, statErr := os.Stat(testFilePath); !os.IsNotExist(statErr) {
		t.Errorf("expected physical file to be deleted, but still exists")
	}
}



