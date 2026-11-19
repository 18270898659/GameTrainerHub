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
	ftsRepo := fts.NewFTSRepository(db)

	// 业务领域层
	trans := translate.NewTranslator()
	flingAdp := fling.NewFlingAdapter(nil)

	searchSvc := search.NewSearchService(ftsRepo, gameRepo, trainerRepo, trans)
	searchSvc.RegisterAdapter(flingAdp)

	dlMgr := download.NewDownloadManager(dlRepo, trainerRepo, localRepo, flingAdp, cfg.MaxConcurrentDownloads)
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
	}

	// 恢复上次异常退出的未完成任务
	_ = a.dlManager.RecoverInterruptedTasks(ctx)

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

		game, _ := a.gameRepo.GetBySlug(ctx, slug)
		var gameID int64
		if game == nil {
			var gameNameCN string
			cnCandidates := a.translator.TranslateEnglishToChinese(gameNameEN)
			if len(cnCandidates) > 0 {
				gameNameCN = cnCandidates[0]
			}
			newGame := &model.Game{
				Slug:   slug,
				NameEN: gameNameEN,
				NameCN: gameNameCN,
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
			if game.NameCN != "" && trainer.TitleCN == "" {
				trainer.TitleCN = game.NameCN + " 修改器"
			}
		}

		trainer.GameID = gameID
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

func slugify(name string) string {
	lower := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}

// GetLatestTrainers 获取原网最近更新修改器
func (a *App) GetLatestTrainers(page int) ([]*model.SearchResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.flingAdapter.GetLatest(ctx, page)
}

// GetLocalTrainers 获取已下载及自定义导入的本地修改器列表
func (a *App) GetLocalTrainers() ([]*trainer.LocalTrainerDetail, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.trainerSvc.ListLocalTrainers(ctx)
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

// CancelDownload 取消下载
func (a *App) CancelDownload(taskID int64) error {
	return a.dlManager.CancelTask(taskID)
}

// ListDownloadTasks 查询当前所有下载任务
func (a *App) ListDownloadTasks() ([]*model.DownloadTask, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.dlRepo.ListTasks(ctx, "")
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
	return a.dlRepo.ListHistory(ctx, limit)
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
	})
	if err != nil {
		return err
	}

	a.updateTrayLabels(updated.Language)

	if a.dlManager != nil {
		a.dlManager.SetMaxConcurrent(updated.MaxConcurrentDownloads)
	}
	return nil
}
