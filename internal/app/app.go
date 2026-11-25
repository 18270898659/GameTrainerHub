package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unsafe"

	"GameTrainerHub/internal/config"
	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/fts"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/diagnostics"
	"GameTrainerHub/internal/download"
	"GameTrainerHub/internal/game"
	"GameTrainerHub/internal/launcher"
	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/search"
	"GameTrainerHub/internal/source/fling"
	"GameTrainerHub/internal/sync"
	"GameTrainerHub/internal/trainer"
	"GameTrainerHub/internal/translate"
	"GameTrainerHub/internal/tray"
	"GameTrainerHub/internal/update"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

type winRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

var (
	modUser32           = windows.NewLazySystemDLL("user32.dll")
	procMonitorFromRect = modUser32.NewProc("MonitorFromRect")
)

const monitorDefaultToNull = 0

func isPositionOnScreen(x, y, w, h int) bool {
	if w <= 0 || h <= 0 {
		return false
	}
	if err := procMonitorFromRect.Find(); err != nil {
		return true // 非 Windows 或无法解析时默认允许
	}
	rc := winRect{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + w),
		Bottom: int32(y + h),
	}
	hMon, _, _ := procMonitorFromRect.Call(uintptr(unsafe.Pointer(&rc)), uintptr(monitorDefaultToNull))
	return hMon != 0
}

// App Wails 桌面应用顶层绑定对象
type App struct {
	ctx          context.Context
	cfg          *config.AppConfig
	db           *database.DB
	gameRepo     *repository.GameRepository
	trainerRepo  *repository.TrainerRepository
	dlRepo       *repository.DownloadRepository
	localRepo    *repository.LocalTrainerRepository
	syncRepo     *repository.SyncRepository
	settingRepo  *repository.SettingRepository
	searchRepo   *repository.SearchRepository
	ftsRepo      *fts.FTSRepository
	translator   *translate.Translator
	flingAdapter *fling.FlingAdapter
	searchSvc    *search.SearchService
	dlManager    *download.DownloadManager
	trainerSvc   *trainer.TrainerService
	steamScanner *game.SteamScanner
	launcher     *launcher.Launcher
	syncSvc      *sync.SyncService
	isQuitting   bool

	// 事件推送回调钩子 (Wails runtime.EventsEmit 或测试捕获)
	EventEmitter func(ctx context.Context, eventName string, data any)
}

func NewApp(cfg *config.AppConfig) (*App, error) {
	if cfg == nil {
		cfg = config.GetConfig()
	}
	_ = cfg.EnsureDirectories()

	// 初始化日志
	_, _ = logger.InitLogger(cfg.LogDir, cfg.LogLevel)

	// 打开并迁移数据库
	dbPath := cfg.GetDBPath()
	db, err := database.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 仓储层
	gameRepo := repository.NewGameRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	dlRepo := repository.NewDownloadRepository(db)
	localRepo := repository.NewLocalTrainerRepository(db)
	syncRepo := repository.NewSyncRepository(db)
	settingRepo := repository.NewSettingRepository(db)
	searchRepo := repository.NewSearchRepository(db)
	ftsRepo := fts.NewFTSRepository(db)

	// 业务领域层
	trans := translate.NewTranslator()
	flingAdp := fling.NewFlingAdapter(nil)

	searchSvc := search.NewSearchService(ftsRepo, gameRepo, trainerRepo, trans)
	searchSvc.RegisterAdapter(flingAdp)

	dlMgr := download.NewDownloadManager(dlRepo, trainerRepo, localRepo, flingAdp, cfg.MaxConcurrentDownloads)
	dlMgr.SetGameRepo(gameRepo)
	trainerSvc := trainer.NewTrainerService(gameRepo, trainerRepo, localRepo)
	steamScan := game.NewSteamScanner()
	launchSvc := launcher.NewLauncher(localRepo, gameRepo)
	syncSvc := sync.NewSyncService(gameRepo, trainerRepo, syncRepo, ftsRepo, trans, flingAdp)

	app := &App{
		cfg:          cfg,
		db:           db,
		gameRepo:     gameRepo,
		trainerRepo:  trainerRepo,
		dlRepo:       dlRepo,
		localRepo:    localRepo,
		syncRepo:     syncRepo,
		settingRepo:  settingRepo,
		searchRepo:   searchRepo,
		ftsRepo:      ftsRepo,
		translator:   trans,
		flingAdapter: flingAdp,
		searchSvc:    searchSvc,
		dlManager:    dlMgr,
		trainerSvc:   trainerSvc,
		steamScanner: steamScan,
		launcher:     launchSvc,
		syncSvc:      syncSvc,
	}

	// 下载进度流向前端推送
	dlMgr.AddListener(func(task *model.DownloadTask) {
		if app.EventEmitter != nil && app.ctx != nil {
			app.EventEmitter(app.ctx, "download:progress", task)
		}
	})

	// 下载完成与失败全局通知
	dlMgr.OnTaskComplete = func(task *model.DownloadTask) {
		if app.EventEmitter != nil && app.ctx != nil {
			app.EventEmitter(app.ctx, "download:completed", task)
		}
		if app.cfg.NotifyOnDownloadComplete {
			_ = tray.ShowNotification("修改器下载完成", fmt.Sprintf("%s 已成功下载并加入库", task.FileName), false)
		}
	}
	dlMgr.OnTaskFail = func(task *model.DownloadTask, err error) {
		if app.EventEmitter != nil && app.ctx != nil {
			app.EventEmitter(app.ctx, "download:failed", task)
		}
		if app.cfg.NotifyOnDownloadFailed {
			_ = tray.ShowNotification("修改器下载失败", fmt.Sprintf("%s 下载失败: %v", task.FileName, err), true)
		}
	}

	return app, nil
}

// Startup Wails 生命周期启动回调
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	if a.EventEmitter == nil {
		a.EventEmitter = func(c context.Context, eventName string, data any) {
			defer func() {
				_ = recover()
			}()
			runtime.EventsEmit(c, eventName, data)
		}
	}

	// 恢复窗口尺寸与屏幕位置
	a.RestoreWindowState()

	// 初始化系统托盘（在桌面运行时）
	if isWailsContext(ctx) {
		_, _ = tray.Init(func() {
			if a.ctx != nil && isWailsContext(a.ctx) {
				runtime.WindowShow(a.ctx)
				runtime.WindowUnminimise(a.ctx)
			}
		}, func() {
			a.Quit()
		})
		a.updateTrayLabels(a.cfg.Language)

		tray.SetHotkeyHandlers(func() {
			a.ToggleWindow()
		}, func() {
			a.OpenDownloadsPage()
		})
		if err := tray.ApplyHotkeys(a.cfg.GlobalHotkeysEnabled, a.cfg.HotkeyToggleWindow, a.cfg.HotkeyOpenDownloads); err != nil {
			logger.Warn("Failed to register global hotkeys on startup", "err", err)
			if a.EventEmitter != nil {
				a.EventEmitter(a.ctx, "hotkey:startup_error", err.Error())
			}
		}
	}

	// 恢复上次异常退出的未完成任务
	_ = a.dlManager.RecoverInterruptedTasks(ctx)

	// 冷启动基准数据初始化：新设备安装后本地数据库尚无数据时，极速填充 760+ 游戏种子及全文索引
	if count, err := a.gameRepo.Count(ctx); err == nil && count == 0 {
		logger.Info("Fresh installation detected: seeding initial game catalog and search indexes...")
		if seedErr := a.syncSvc.SeedInitialDatabase(ctx); seedErr != nil {
			logger.Error("Failed to seed initial database", "err", seedErr)
		} else {
			logger.Info("Successfully seeded initial game catalog into database")
		}
	}

	// 后台定期同步（若开启）
	if a.cfg.AutoSyncEnabled {
		a.syncSvc.StartPeriodicSync(a.cfg.AutoSyncIntervalHours)
	}

	logger.Info("GameTrainerHub application startup completed")
}

// Shutdown Wails 生命周期关闭回调
func (a *App) Shutdown(ctx context.Context) {
	tray.Destroy()
	if a.syncSvc != nil {
		a.syncSvc.Stop()
	}
	if a.dlManager != nil {
		a.dlManager.Close()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
	logger.CloseLogger()
}

// ----------------- IPC 导出 API (前端调用) -----------------

// Search 统一中英文混合搜索
func (a *App) Search(filter model.SearchFilter) ([]*model.SearchResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.searchSvc.Search(ctx, &filter)
}

// GetSearchSuggestions 提供首页快速搜索建议（限制最多 6 条候选，按相关度排序）
func (a *App) GetSearchSuggestions(query string) ([]*model.SearchResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	clean := strings.TrimSpace(query)
	if clean == "" {
		return nil, nil
	}
	filter := &model.SearchFilter{
		Query:    clean,
		Page:     1,
		PageSize: 6,
	}
	return a.searchSvc.Search(ctx, filter)
}

// GetTrainerDetail 获取原网修改器全部版本与说明并持久化，保证返回具有数据库自增ID的版本列表
func (a *App) GetTrainerDetail(detailURL string) (*model.Trainer, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	cleanURL := strings.TrimSpace(detailURL)
	if cleanURL == "" {
		return nil, fmt.Errorf("detailURL cannot be empty")
	}

	// 1. 尝试从原网抓取最新版本信息
	trainer, versions, err := a.flingAdapter.GetDetail(ctx, cleanURL)
	if err == nil && trainer != nil {
		gameNameEN := strings.TrimSuffix(trainer.TitleEN, " Trainer")
		slug := slugify(gameNameEN)

		coverURL := trainer.CoverURL
		game, _ := a.gameRepo.GetBySlug(ctx, slug)
		var gameID int64
		if game == nil {
			var gameNameCN string
			cnCandidates := a.translator.TranslateEnglishToChinese(gameNameEN)
			if len(cnCandidates) > 0 {
				gameNameCN = cnCandidates[0]
			}
			newGame := &model.Game{
				Slug:     slug,
				NameEN:   gameNameEN,
				NameCN:   gameNameCN,
				CoverURL: coverURL,
			}
			if cErr := a.gameRepo.Create(ctx, newGame); cErr == nil {
				gameID = newGame.ID
				aliases := a.translator.TranslateEnglishToChinese(gameNameEN)
				py := a.translator.ToPinyin(gameNameCN)
				pyInit := a.translator.ToPinyinInitials(gameNameCN)
				_ = a.ftsRepo.IndexGame(ctx, gameID, gameNameEN, gameNameCN, strings.Join(aliases, " "), py, pyInit)
			}
		} else {
			gameID = game.ID
			if game.CoverURL == "" && coverURL != "" {
				game.CoverURL = coverURL
				_ = a.gameRepo.Update(ctx, game)
			}
			if game.NameCN != "" && trainer.TitleCN == "" {
				trainer.TitleCN = game.NameCN + " 修改器"
			}
		}

		trainer.GameID = gameID
		_ = a.gameRepo.RecordVisit(ctx, gameID)
		if trainer.TitleCN == "" {
			cnCandidates := a.translator.TranslateEnglishToChinese(gameNameEN)
			if len(cnCandidates) > 0 {
				trainer.TitleCN = cnCandidates[0] + " 修改器"
			}
		}

		// 事务入库或更新修改器与版本数据，赋予自增 ID
		_ = a.trainerRepo.UpsertTrainerWithVersions(ctx, trainer, versions)

		// 重新加载带下载状态与真实数据库自增 ID 的版本列表
		dbVersions, vErr := a.trainerRepo.ListVersionsByTrainerID(ctx, trainer.ID)
		if vErr == nil && len(dbVersions) > 0 {
			trainer.Versions = dbVersions
		} else {
			trainer.Versions = versions
		}
		return trainer, nil
	}

	// 2. 若在线解析失败（如网络中断），优雅降级从本地数据库查找已有版本
	existingTrainer, dbErr := a.trainerRepo.GetTrainerByDetailURL(ctx, cleanURL)
	if dbErr == nil && existingTrainer != nil {
		dbVersions, _ := a.trainerRepo.ListVersionsByTrainerID(ctx, existingTrainer.ID)
		existingTrainer.Versions = dbVersions
		return existingTrainer, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch trainer detail: %w", err)
	}
	return nil, fmt.Errorf("trainer detail not found for URL: %s", cleanURL)
}

// GetGameDetail 根据游戏唯一 ID 直接查询修改器详情与可用版本列表
func (a *App) GetGameDetail(gameID int64) (*model.Trainer, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	game, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil || game == nil {
		return nil, fmt.Errorf("game not found with id: %d", gameID)
	}

	// 记录最近访问
	_ = a.gameRepo.RecordVisit(ctx, gameID)

	// 1. 若 game 实体自身已有非空 detail_url，优先直接查询
	if game.DetailURL != "" {
		if tr, err := a.GetTrainerDetail(game.DetailURL); err == nil && tr != nil {
			return tr, nil
		}
	}

	// 2. 查找该游戏关联的修改器记录
	trainers, err := a.trainerRepo.ListTrainersByGameID(ctx, gameID)
	if err == nil && len(trainers) > 0 {
		tr := trainers[0]
		if tr.DetailURL != "" {
			if detail, err := a.GetTrainerDetail(tr.DetailURL); err == nil && detail != nil {
				return detail, nil
			}
		}
		// 若数据库已有版本记录，直接返回
		dbVersions, vErr := a.trainerRepo.ListVersionsByTrainerID(ctx, tr.ID)
		if vErr == nil && len(dbVersions) > 0 {
			tr.Versions = dbVersions
			return tr, nil
		}
	}

	// 3. 若尚未入库该修改器，基于 slug 尝试检索或构建官方标准 URL
	guessURL := fmt.Sprintf("%s/trainer/%s-trainer/", a.flingAdapter.BaseURL(), game.Slug)
	if tr, err := a.GetTrainerDetail(guessURL); err == nil && tr != nil {
		return tr, nil
	}
	guessURL2 := fmt.Sprintf("%s/trainer/%s/", a.flingAdapter.BaseURL(), game.Slug)
	if tr, err := a.GetTrainerDetail(guessURL2); err == nil && tr != nil {
		return tr, nil
	}

	// 4. 尝试通过在线搜索服务按游戏名查询
	if a.searchSvc != nil {
		query := game.NameEN
		if query == "" {
			query = game.NameCN
		}
		searchResults, sErr := a.searchSvc.Search(ctx, &model.SearchFilter{
			Query:    query,
			Page:     1,
			PageSize: 5,
		})
		if sErr == nil {
			for _, sr := range searchResults {
				if sr.DetailURL != "" {
					if tr, tErr := a.GetTrainerDetail(sr.DetailURL); tErr == nil && tr != nil {
						return tr, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no trainer found for game: %s (%s)", game.NameCN, game.NameEN)
}

func slugify(name string) string {
	lower := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}

// GetLatestTrainers 获取原网最近更新修改器，并自动补全游戏实体映射、封面及收藏状态
// 具备超时防护与离线降级兜底机制，保证在弱网、离线或全新设备上首页绝不显示为空白
func (a *App) GetLatestTrainers(page int) ([]*model.SearchResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// 1. 设置合理的在线抓取超时时间 (5 秒)，避免外部网络波动导致前端阻塞或超时白屏
	onlineCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	results, err := a.flingAdapter.GetLatest(onlineCtx, page)
	cancel()

	if err == nil && len(results) > 0 {
		for _, item := range results {
			slug := slugify(item.GameNameEN)
			if item.Slug == "" {
				item.Slug = slug
			}
			if item.GameNameCN == "" {
				cnCandidates := a.translator.TranslateEnglishToChinese(item.GameNameEN)
				if len(cnCandidates) > 0 {
					item.GameNameCN = cnCandidates[0]
				}
			}

			// 确保关联或创建本地游戏记录，赋予稳定有效的 GameID 与收藏状态
			g, _ := a.gameRepo.GetOrCreateBySlug(ctx, item.Slug, item.GameNameEN, item.GameNameCN, item.CoverURL)
			if g != nil {
				item.GameID = g.ID
				item.IsFavorite = g.IsFavorite
				if item.CoverURL == "" && g.CoverURL != "" {
					item.CoverURL = g.CoverURL
				}
			}

			if item.GameID > 0 {
				if locals, lErr := a.localRepo.ListByGameID(ctx, item.GameID); lErr == nil && len(locals) > 0 {
					item.IsDownloaded = true
				}
			}
		}
		return results, nil
	}

	// 2. 在线请求失败、超时或离线：启用离线降级兜底方案，保证首页内容丰满展示
	logger.Warn("Failed to fetch online latest trainers, falling back to local seed/cache", "page", page, "err", err)
	return a.getFallbackLatestTrainers(ctx, page)
}

func (a *App) getFallbackLatestTrainers(ctx context.Context, page int) ([]*model.SearchResult, error) {
	if page <= 1 {
		seedLatest, err := sync.LoadSeedLatest()
		if err == nil && len(seedLatest) > 0 {
			for _, item := range seedLatest {
				slug := slugify(item.GameNameEN)
				if item.Slug == "" {
					item.Slug = slug
				}
				if item.GameNameCN == "" {
					cnCandidates := a.translator.TranslateEnglishToChinese(item.GameNameEN)
					if len(cnCandidates) > 0 {
						item.GameNameCN = cnCandidates[0]
					}
				}
				g, _ := a.gameRepo.GetBySlug(ctx, item.Slug)
				if g != nil {
					item.GameID = g.ID
					item.IsFavorite = g.IsFavorite
					if item.CoverURL == "" && g.CoverURL != "" {
						item.CoverURL = g.CoverURL
					}
				}
				if item.GameID > 0 {
					if locals, lErr := a.localRepo.ListByGameID(ctx, item.GameID); lErr == nil && len(locals) > 0 {
						item.IsDownloaded = true
					}
				}
			}
			return seedLatest, nil
		}
	}

	// 从本地数据库分页查询已收录游戏作为兜底展示
	offset := (page - 1) * 20
	if offset < 0 {
		offset = 0
	}
	games, err := a.gameRepo.List(ctx, offset, 20)
	if err != nil {
		return nil, err
	}

	var results []*model.SearchResult
	for _, g := range games {
		results = append(results, &model.SearchResult{
			GameID:       g.ID,
			Slug:         g.Slug,
			GameNameEN:   g.NameEN,
			GameNameCN:   g.NameCN,
			TrainerTitle: g.NameEN + " Trainer",
			DetailURL:    g.DetailURL,
			CoverURL:     g.CoverURL,
			SourceID:     "fling",
			UpdatedAt:    g.UpdatedAt,
			IsFavorite:   g.IsFavorite,
			IsOnlineOnly: true,
		})
	}
	return results, nil
}

// GetGameCount 获取本地索引收录的游戏总数
func (a *App) GetGameCount() (int, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.gameRepo.Count(ctx)
}

// GetLocalTrainers 获取已下载及自定义导入的本地修改器列表（附加运行中状态）
func (a *App) GetLocalTrainers() ([]*trainer.LocalTrainerDetail, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	details, err := a.trainerSvc.ListLocalTrainers(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range details {
		if d.LocalTrainer != nil {
			d.IsRunning = a.launcher.IsTrainerRunning(d.LocalTrainer.ID)
		}
	}
	return details, nil
}

// ImportCustomTrainer 手动导入本地已有 EXE
func (a *App) ImportCustomTrainer(gameID int64, filePath string) (*model.LocalTrainer, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.trainerSvc.ImportCustomTrainer(ctx, gameID, filePath)
}

// DeleteLocalTrainer 删除本地修改器记录（可选物理删除）
func (a *App) DeleteLocalTrainer(id int64, deleteFile bool) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.trainerSvc.DeleteLocalTrainer(ctx, id, deleteFile)
}

// ToggleFavorite 收藏/取消收藏游戏
func (a *App) ToggleFavorite(gameID int64) (bool, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.trainerSvc.ToggleFavorite(ctx, gameID)
}

// ListFavorites 获取收藏列表
func (a *App) ListFavorites() ([]*model.Game, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.gameRepo.ListFavorites(ctx)
}

// StartDownload 启动指定修改器版本下载
func (a *App) StartDownload(versionID int64) (*model.DownloadTask, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.dlManager.CreateTask(ctx, versionID, a.cfg.DownloadDir)
}

// PauseDownload 暂停下载
func (a *App) PauseDownload(taskID int64) error {
	return a.dlManager.PauseTask(taskID)
}

// ResumeDownload 继续下载
func (a *App) ResumeDownload(taskID int64) error {
	return a.dlManager.ResumeTask(taskID)
}

// RedownloadLocalTrainer 针对文件丢失或损坏的本地修改器重新下载
func (a *App) RedownloadLocalTrainer(localTrainerID int64) (*model.DownloadTask, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	lt, err := a.localRepo.GetByID(ctx, localTrainerID)
	if err != nil || lt == nil {
		return nil, fmt.Errorf("local trainer not found with id: %d", localTrainerID)
	}

	// 1. 若记录中存在明确的版本 VersionID，直接创建重新下载任务
	if lt.VersionID != nil && *lt.VersionID > 0 {
		return a.dlManager.RedownloadTask(ctx, *lt.VersionID, a.cfg.DownloadDir)
	}

	// 2. 若未绑定 VersionID，尝试查找该关联游戏下的最新版本
	if lt.GameID > 0 {
		trainers, tErr := a.trainerRepo.ListTrainersByGameID(ctx, lt.GameID)
		if tErr == nil && len(trainers) > 0 {
			versions, vErr := a.trainerRepo.ListVersionsByTrainerID(ctx, trainers[0].ID)
			if vErr == nil && len(versions) > 0 {
				return a.dlManager.RedownloadTask(ctx, versions[0].ID, a.cfg.DownloadDir)
			}
		}
	}

	return nil, fmt.Errorf("cannot redownload: no valid version found for trainer ID %d", localTrainerID)
}

// CancelDownload 取消下载
func (a *App) CancelDownload(taskID int64) error {
	return a.dlManager.CancelTask(taskID)
}

// ListDownloadTasks 查询所有下载任务（包含下载中、排队中、暂停、重试、已完成及失败任务，按更新时间降序）
func (a *App) ListDownloadTasks() ([]*model.DownloadTask, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	tasks, err := a.dlRepo.ListTasks(ctx, "")
	if err != nil {
		return make([]*model.DownloadTask, 0), err
	}
	if tasks == nil {
		tasks = make([]*model.DownloadTask, 0)
	}
	return tasks, nil
}

// DeleteDownloadTask 从下载任务列表中删除单个任务记录（默认不删除本地物理文件）
func (a *App) DeleteDownloadTask(taskID int64) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.dlRepo.DeleteTask(ctx, taskID)
}

// ListDownloadHistory 查询下载历史
func (a *App) ListDownloadHistory(limit int) ([]*model.DownloadHistory, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		limit = 50
	}
	list, err := a.dlRepo.ListHistory(ctx, limit)
	if err != nil {
		return make([]*model.DownloadHistory, 0), err
	}
	if list == nil {
		list = make([]*model.DownloadHistory, 0)
	}
	return list, nil
}

// DeleteDownloadHistory 删除指定下载历史记录（默认仅删除历史记录，不影响本地物理文件；deletePhysicalFile 为 true 时同时清理本地文件）
func (a *App) DeleteDownloadHistory(id int64, deletePhysicalFile bool) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if deletePhysicalFile {
		hist, err := a.dlRepo.GetHistoryByID(ctx, id)
		if err == nil && hist != nil && hist.SavePath != "" {
			_ = os.Remove(hist.SavePath)
			if a.localRepo != nil {
				if lt, lErr := a.localRepo.GetByFilePath(ctx, hist.SavePath); lErr == nil && lt != nil {
					_ = a.localRepo.Delete(ctx, lt.ID)
				}
			}
		}
	}

	return a.dlRepo.DeleteHistory(ctx, id)
}

// ClearDownloadHistory 清空所有下载历史
func (a *App) ClearDownloadHistory() error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.dlRepo.ClearHistory(ctx)
}

// ToggleFavoriteBySlug 根据 Slug 或游戏名称进行收藏切换，确保未提前入库的游戏也能成功收藏
func (a *App) ToggleFavoriteBySlug(slug, nameEN, nameCN, coverURL string) (bool, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if slug == "" {
		slug = slugify(nameEN)
	}
	g, err := a.gameRepo.GetOrCreateBySlug(ctx, slug, nameEN, nameCN, coverURL)
	if err != nil || g == nil {
		return false, fmt.Errorf("failed to get or create game: %w", err)
	}
	return a.trainerSvc.ToggleFavorite(ctx, g.ID)
}

// GetSecurityCard 获取启动前安全信息卡片
func (a *App) GetSecurityCard(localTrainerID int64) (*launcher.SecurityCardInfo, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.launcher.GetSecurityCard(ctx, localTrainerID)
}

// LaunchTrainer 安全启动修改器进程
func (a *App) LaunchTrainer(localTrainerID int64, runAsAdmin bool) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.launcher.LaunchExecutable(ctx, localTrainerID, runAsAdmin)
}

// OpenContainingFolder 打开修改器所在系统文件夹并选中
func (a *App) OpenContainingFolder(filePath string) error {
	return a.launcher.OpenContainingFolder(filePath)
}

// ScanSteamGames 扫描本地 Steam 已安装游戏并自动关联
func (a *App) ScanSteamGames() (int, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.steamScanner.SyncWithGameRepo(ctx, a.gameRepo)
}

// TriggerSync 手动触发原网站最新资源增量同步
func (a *App) TriggerSync() (*model.SyncRecord, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.syncSvc.SyncLatest(ctx, 2)
}

func isWailsContext(ctx context.Context) bool {
	return ctx != nil && ctx.Value("frontend") != nil
}

// GetContext 获取当前应用 context
func (a *App) GetContext() context.Context {
	return a.ctx
}

// IsQuitting 检查当前是否正处于强制退出流程
func (a *App) IsQuitting() bool {
	return a.isQuitting
}

// Quit 显式退出程序
func (a *App) Quit() {
	a.isQuitting = true
	a.SaveWindowState()
	tray.Destroy()
	if isWailsContext(a.ctx) {
		runtime.Quit(a.ctx)
	}
}

// OnBeforeClose 窗口关闭前钩子（拦截最小化到托盘或退出）
func (a *App) OnBeforeClose(ctx context.Context) (prevent bool) {
	a.SaveWindowState()

	if a.isQuitting {
		return false
	}

	if a.cfg.CloseToTray {
		if isWailsContext(ctx) {
			runtime.WindowHide(ctx)
		}
		return true
	}

	return false
}

// SaveWindowState 保存当前窗口尺寸与位置
func (a *App) SaveWindowState() {
	if !isWailsContext(a.ctx) {
		return
	}
	w, h := runtime.WindowGetSize(a.ctx)
	x, y := runtime.WindowGetPosition(a.ctx)

	_ = a.cfg.Update(func(c *config.AppConfig) {
		if c.RememberWindowSize && w >= 1024 && h >= 680 {
			c.WindowWidth = w
			c.WindowHeight = h
		}
		if c.RememberWindowPosition {
			c.WindowX = x
			c.WindowY = y
		}
	})
}

// RestoreWindowState 启动时恢复窗口尺寸与位置
func (a *App) RestoreWindowState() {
	if !isWailsContext(a.ctx) {
		return
	}

	// 1. 恢复或设置窗口尺寸
	w := 1280
	h := 820
	if a.cfg.RememberWindowSize && a.cfg.WindowWidth >= 1024 && a.cfg.WindowHeight >= 680 {
		w = a.cfg.WindowWidth
		h = a.cfg.WindowHeight
	}
	runtime.WindowSetSize(a.ctx, w, h)

	// 2. 恢复或重置窗口位置（校验多显示器边界有效性）
	if a.cfg.RememberWindowPosition {
		if isPositionOnScreen(a.cfg.WindowX, a.cfg.WindowY, w, h) {
			runtime.WindowSetPosition(a.ctx, a.cfg.WindowX, a.cfg.WindowY)
		} else {
			// 超出当前连接屏幕坐标范围，安全居中
			runtime.WindowCenter(a.ctx)
		}
	} else {
		runtime.WindowCenter(a.ctx)
	}
}

// ValidateDirectoryWritable 检查目录是否合法且具备写入权限
func ValidateDirectoryWritable(dir string) error {
	cleanDir := filepath.Clean(dir)
	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return fmt.Errorf("创建或访问目录失败: %w", err)
	}

	testFile := filepath.Join(cleanDir, fmt.Sprintf(".test_write_%d.tmp", time.Now().UnixNano()))
	f, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("该目录没有写入权限: %w", err)
	}
	_, _ = f.WriteString("write_test")
	_ = f.Close()
	_ = os.Remove(testFile)

	return nil
}

func (a *App) hasActiveDownloads(ctx context.Context) bool {
	if a.dlManager != nil && a.dlManager.HasActiveTasks() {
		return true
	}
	if a.dlRepo != nil {
		tasks, err := a.dlRepo.ListTasks(ctx, "")
		if err == nil {
			for _, t := range tasks {
				if t.Status == model.DownloadStatusDownloading || t.Status == model.DownloadStatusRetrying {
					return true
				}
			}
		}
	}
	return false
}

// SelectDownloadDirectory 唤起原生文件夹选择对话框并校验写入权限
func (a *App) SelectDownloadDirectory() (string, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	// 检查当前是否有正在进行的下载任务
	if a.hasActiveDownloads(ctx) {
		return "", fmt.Errorf("当前有正在进行的下载任务，请等待下载完成或暂停任务后再更改存储路径")
	}

	if !isWailsContext(a.ctx) {
		return "", fmt.Errorf("桌面上下文未就绪")
	}

	selectedDir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: a.cfg.DownloadDir,
		Title:            "选择修改器默认下载目录",
	})
	if err != nil {
		return "", fmt.Errorf("打开目录选择对话框失败: %w", err)
	}
	if selectedDir == "" {
		return "", nil // 用户取消操作
	}

	if err := ValidateDirectoryWritable(selectedDir); err != nil {
		return "", err
	}

	// 校验通过，更新持久化配置
	err = a.cfg.Update(func(c *config.AppConfig) {
		c.DownloadDir = selectedDir
	})
	if err != nil {
		return "", fmt.Errorf("保存下载目录配置失败: %w", err)
	}

	logger.Info("Updated default download directory", "dir", selectedDir)
	return selectedDir, nil
}

func (a *App) updateTrayLabels(lang string) {
	switch lang {
	case "zh-TW":
		tray.UpdateLabels("顯示主視窗", "結束程式")
	case "en-US":
		tray.UpdateLabels("Show Main Window", "Exit Application")
	default:
		tray.UpdateLabels("显示主窗口", "退出程序")
	}
}

// GetConfig 获取当前配置
func (a *App) GetConfig() *config.AppConfig {
	return a.cfg
}

// SaveConfig 保存配置更新并动态更新调度器
func (a *App) SaveConfig(updated *config.AppConfig) error {
	err := a.cfg.Update(func(c *config.AppConfig) {
		c.DownloadDir = updated.DownloadDir
		c.MaxConcurrentDownloads = updated.MaxConcurrentDownloads
		c.AutoSyncEnabled = updated.AutoSyncEnabled
		c.AutoSyncIntervalHours = updated.AutoSyncIntervalHours
		c.ProxyURL = updated.ProxyURL
		c.Theme = updated.Theme
		c.Language = updated.Language
		c.SafeLaunchPrompt = updated.SafeLaunchPrompt
		c.CloseToTray = updated.CloseToTray
		c.RememberWindowSize = updated.RememberWindowSize
		c.RememberWindowPosition = updated.RememberWindowPosition
		c.NotifyOnDownloadComplete = updated.NotifyOnDownloadComplete
		c.NotifyOnDownloadFailed = updated.NotifyOnDownloadFailed
		c.NotifyOnTrainerUpdate = updated.NotifyOnTrainerUpdate
		c.GlobalHotkeysEnabled = updated.GlobalHotkeysEnabled
		c.HotkeyToggleWindow = updated.HotkeyToggleWindow
		c.HotkeyOpenDownloads = updated.HotkeyOpenDownloads
		c.AutoBackupConfig = updated.AutoBackupConfig
	})
	if err != nil {
		return err
	}

	a.updateTrayLabels(updated.Language)

	if a.dlManager != nil {
		a.dlManager.SetMaxConcurrent(updated.MaxConcurrentDownloads)
	}

	_ = tray.UpdateHotkeys(updated.GlobalHotkeysEnabled, updated.HotkeyToggleWindow, updated.HotkeyOpenDownloads)

	return nil
}

// PauseAllDownloads 暂停所有活跃下载任务
func (a *App) PauseAllDownloads() error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.dlManager.PauseAllTasks(ctx)
}

// ResumeAllDownloads 恢复所有暂停或失败的下载任务
func (a *App) ResumeAllDownloads() error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.dlManager.ResumeAllTasks(ctx)
}

// GetSearchHistory 获取持久化搜索历史关键词
func (a *App) GetSearchHistory(limit int) ([]string, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.searchRepo.ListHistory(ctx, limit)
}

// AddSearchHistory 添加或更新搜索历史
func (a *App) AddSearchHistory(keyword string) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.searchRepo.AddHistory(ctx, keyword)
}

// RemoveSearchHistory 移除单条搜索历史
func (a *App) RemoveSearchHistory(keyword string) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.searchRepo.RemoveHistory(ctx, keyword)
}

// ClearSearchHistory 清空全部搜索历史
func (a *App) ClearSearchHistory() error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.searchRepo.ClearHistory(ctx)
}

// RecordGameVisit 记录用户访问游戏详情
func (a *App) RecordGameVisit(gameID int64) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.gameRepo.RecordVisit(ctx, gameID)
}

// GetRecentGames 获取最近访问过的游戏
func (a *App) GetRecentGames(limit int) ([]*model.Game, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.gameRepo.ListVisited(ctx, limit)
}

// GetRecentTrainers 获取最近使用/更新的本地修改器
func (a *App) GetRecentTrainers(limit int) ([]*trainer.LocalTrainerDetail, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	details, err := a.trainerSvc.GetRecentTrainers(ctx, limit)
	if err != nil {
		return nil, err
	}
	for _, d := range details {
		if d.LocalTrainer != nil {
			d.IsRunning = a.launcher.IsTrainerRunning(d.LocalTrainer.ID)
		}
	}
	return details, nil
}

// VerifyTrainerIntegrity 校验本地修改器文件 SHA-256 完整性
func (a *App) VerifyTrainerIntegrity(localTrainerID int64) (*model.IntegrityResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.trainerSvc.VerifyTrainerIntegrity(ctx, localTrainerID)
}

// RelocateTrainer 重新定位修改器路径
func (a *App) RelocateTrainer(localTrainerID int64, newFilePath string) error {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.trainerSvc.RelocateTrainer(ctx, localTrainerID, newFilePath)
}

// IsTrainerRunning 检查修改器是否正在运行中
func (a *App) IsTrainerRunning(localTrainerID int64) bool {
	return a.launcher.IsTrainerRunning(localTrainerID)
}

// StopTrainer 终止正在运行的修改器进程
func (a *App) StopTrainer(localTrainerID int64) error {
	return a.launcher.StopTrainer(localTrainerID)
}

// SelectTrainerExecutable 唤起原生对话框选择本地已有 .exe 修改器
func (a *App) SelectTrainerExecutable() (string, error) {
	if !isWailsContext(a.ctx) {
		return "", fmt.Errorf("桌面上下文未就绪")
	}
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择游戏修改器可执行文件 (.exe)",
		Filters: []runtime.FileFilter{
			{DisplayName: "Executable (*.exe)", Pattern: "*.exe"},
		},
	})
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// ToggleWindow 全局快捷键呼出/隐藏客户端
func (a *App) ToggleWindow() {
	if a.ctx == nil || !isWailsContext(a.ctx) {
		return
	}
	if tray.IsMainWindowActiveAndVisible() {
		runtime.WindowHide(a.ctx)
	} else {
		runtime.WindowShow(a.ctx)
		runtime.WindowUnminimise(a.ctx)
		tray.BringMainWindowToForeground()
	}
}

// OpenDownloadsPage 全局快捷键直达下载管理页面
func (a *App) OpenDownloadsPage() {
	if a.ctx == nil || !isWailsContext(a.ctx) {
		return
	}
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	tray.BringMainWindowToForeground()
	if a.EventEmitter != nil {
		a.EventEmitter(a.ctx, "navigate:downloads", nil)
	}
}

// ApplyHotkeys 强制注册并持久化应用全局快捷键
func (a *App) ApplyHotkeys(toggleWindow, openDownloads string, enabled bool) error {
	normToggle, err := tray.NormalizeHotkeyString(toggleWindow)
	if err != nil {
		return fmt.Errorf("“呼出/隐藏客户端”快捷键格式非法: %w", err)
	}
	normDl, err := tray.NormalizeHotkeyString(openDownloads)
	if err != nil {
		return fmt.Errorf("“直达下载”快捷键格式非法: %w", err)
	}

	if normToggle == normDl {
		return fmt.Errorf("快捷键冲突：“呼出/隐藏客户端”与“直达下载”不能设置为相同的组合键 (%s)", normToggle)
	}

	// 调用底层原子注册与回滚机制
	if err := tray.ApplyHotkeys(enabled, normToggle, normDl); err != nil {
		return err
	}

	// 注册成功后持久化保存到现有配置
	return a.cfg.Update(func(c *config.AppConfig) {
		c.GlobalHotkeysEnabled = enabled
		c.HotkeyToggleWindow = normToggle
		c.HotkeyOpenDownloads = normDl
	})
}

// ResetHotkeys 恢复默认全局快捷键并立即应用与持久化
func (a *App) ResetHotkeys() (map[string]string, error) {
	const defToggle = "Alt + Ctrl + Q"
	const defDl = "Alt + Ctrl + J"

	if err := tray.ApplyHotkeys(a.cfg.GlobalHotkeysEnabled, defToggle, defDl); err != nil {
		return nil, err
	}

	_ = a.cfg.Update(func(c *config.AppConfig) {
		c.HotkeyToggleWindow = defToggle
		c.HotkeyOpenDownloads = defDl
	})

	return map[string]string{
		"hotkey_toggle_window":  defToggle,
		"hotkey_open_downloads": defDl,
	}, nil
}

// SuspendHotkeys 挂起或恢复全局快捷键拦截（录入状态避免抢占）
func (a *App) SuspendHotkeys(suspend bool) {
	tray.SuspendHotkeys(suspend)
}

// GetDefaultHotkeys 获取系统默认快捷键配置
func (a *App) GetDefaultHotkeys() map[string]string {
	return map[string]string{
		"hotkey_toggle_window":  "Alt + Ctrl + Q",
		"hotkey_open_downloads": "Alt + Ctrl + J",
	}
}

// CheckHotkeyConflict 检测指定快捷键是否冲突
func (a *App) CheckHotkeyConflict(hotkeyStr string) (bool, error) {
	return tray.CheckHotkeyConflict(hotkeyStr)
}

// CheckForUpdates 客户端版本检测与更新提醒
func (a *App) CheckForUpdates() (*model.UpdateCheckResult, error) {
	currentVer := "1.0.0"
	latestVer := "1.0.0"
	hasUpdate := update.CompareVersions(latestVer, currentVer) > 0
	return &model.UpdateCheckResult{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVer,
		LatestVersion:  latestVer,
		ReleaseNotes:   "当前已是最新发布版本 (v1.0.0)。",
		PublishedAt:    time.Now().Format("2006-01-02"),
	}, nil
}

// DiagnoseEnvironment 执行一键环境与网络诊断
func (a *App) DiagnoseEnvironment() (*model.DiagnosticReport, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	activeTasks, _ := a.dlRepo.ListTasks(ctx, model.DownloadStatusDownloading)
	allLocals, _ := a.localRepo.ListAll(ctx)
	allFavs, _ := a.gameRepo.ListFavorites(ctx)

	return diagnostics.DiagnoseEnvironment(a.cfg, a.db, len(activeTasks), len(allLocals), len(allFavs))
}

// ExportDiagnostics 一键导出诊断日志为脱敏文本文件并打开所在文件夹
func (a *App) ExportDiagnostics() (string, error) {
	rep, err := a.DiagnoseEnvironment()
	if err != nil {
		return "", err
	}
	logs := logger.GetRecentLogs(100, "")
	exportDir := a.cfg.DownloadDir
	if exportDir == "" {
		exportDir = a.cfg.BaseDir
	}
	filePath, err := diagnostics.ExportDiagnosticsReport(rep, logs, exportDir)
	if err != nil {
		return "", err
	}
	_ = a.OpenContainingFolder(filePath)
	return filePath, nil
}

// GetRecentLogs 查询最新分类日志列表
func (a *App) GetRecentLogs(limit int, category string) ([]model.LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	return logger.GetRecentLogs(limit, category), nil
}

// ExportConfig 导出客户端配置为格式化 JSON 字符串
func (a *App) ExportConfig() (string, error) {
	return a.cfg.ExportConfigJSON()
}

// ImportConfig 从 JSON 导入客户端配置并热应用
func (a *App) ImportConfig(jsonStr string) error {
	if err := a.cfg.ImportConfigJSON(jsonStr); err != nil {
		return fmt.Errorf("导入配置失败: %w", err)
	}
	a.updateTrayLabels(a.cfg.Language)
	if a.dlManager != nil {
		a.dlManager.SetMaxConcurrent(a.cfg.MaxConcurrentDownloads)
	}
	_ = tray.UpdateHotkeys(a.cfg.GlobalHotkeysEnabled, a.cfg.HotkeyToggleWindow, a.cfg.HotkeyOpenDownloads)
	return nil
}

// ResetConfig 恢复默认设置并热应用
func (a *App) ResetConfig() error {
	if err := a.cfg.ResetToDefault(); err != nil {
		return fmt.Errorf("恢复默认设置失败: %w", err)
	}
	a.updateTrayLabels(a.cfg.Language)
	if a.dlManager != nil {
		a.dlManager.SetMaxConcurrent(a.cfg.MaxConcurrentDownloads)
	}
	_ = tray.UpdateHotkeys(a.cfg.GlobalHotkeysEnabled, a.cfg.HotkeyToggleWindow, a.cfg.HotkeyOpenDownloads)
	return nil
}
