package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

