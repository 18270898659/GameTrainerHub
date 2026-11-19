package download

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/source"
)

type DownloadManager struct {
	dlRepo        *repository.DownloadRepository
	trainerRepo   *repository.TrainerRepository
	localRepo     *repository.LocalTrainerRepository
	httpClient    *http.Client
	adapter       source.SourceAdapter
	maxConcurrent int

	activeTasks map[int64]context.CancelFunc
	taskQueue   chan int64
	listeners   []ProgressCallback
	mu          sync.RWMutex
	condMu      sync.Mutex
	cond        *sync.Cond
	activeCount int
	ctx         context.Context
	cancel      context.CancelFunc

	// 任务生命周期回调（用于全局通知与统计）
	OnTaskComplete func(task *model.DownloadTask)
	OnTaskFail     func(task *model.DownloadTask, err error)
}

func NewDownloadManager(
	dlRepo *repository.DownloadRepository,
	trainerRepo *repository.TrainerRepository,
	localRepo *repository.LocalTrainerRepository,
	adapter source.SourceAdapter,
	maxConcurrent int,
) *DownloadManager {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}

	ctx, cancel := context.WithCancel(context.Background())

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	dm := &DownloadManager{
		dlRepo:        dlRepo,
		trainerRepo:   trainerRepo,
		localRepo:     localRepo,
		adapter:       adapter,
		httpClient:    &http.Client{Transport: transport, Timeout: 0},
		maxConcurrent: maxConcurrent,
		activeTasks:   make(map[int64]context.CancelFunc),
		taskQueue:     make(chan int64, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
	dm.cond = sync.NewCond(&dm.condMu)

	// 启动统一任务分发调度循环 (支持动态并发调整)
	go dm.dispatchLoop()

	return dm
}

// AddListener 注册下载事件监听器 (供 Wails 前端推流)
func (m *DownloadManager) AddListener(cb ProgressCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, cb)
}

func (m *DownloadManager) notifyListeners(task *model.DownloadTask) {
	m.mu.RLock()
	listeners := make([]ProgressCallback, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, cb := range listeners {
		cb(task)
	}
}

// RecoverInterruptedTasks 启动时恢复中断的下载任务为暂停状态
func (m *DownloadManager) RecoverInterruptedTasks(ctx context.Context) error {
	tasks, err := m.dlRepo.ListTasks(ctx, model.DownloadStatusDownloading)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		t.Status = model.DownloadStatusPaused
		t.SpeedBytesPerSec = 0
		_ = m.dlRepo.UpdateTask(ctx, t)
		logger.Info("Recovered interrupted download task to paused", "task_id", t.ID, "file", t.FileName)
	}
	return nil
}

// CreateTask 创建新下载任务并入队（包含重复任务检测与状态恢复）
func (m *DownloadManager) CreateTask(ctx context.Context, versionID int64, downloadDir string) (*model.DownloadTask, error) {
	// 0. 重复任务检测与状态自愈
	existingTasks, err := m.dlRepo.ListTasks(ctx, "")
	if err == nil {
		for _, ext := range existingTasks {
			if ext.VersionID == versionID {
				switch ext.Status {
				case model.DownloadStatusDownloading, model.DownloadStatusPending:
					logger.Info("Download task already active, returning existing", "task_id", ext.ID)
					return ext, nil
				case model.DownloadStatusPaused, model.DownloadStatusRetrying:
					_ = m.ResumeTask(ext.ID)
					return ext, nil
				case model.DownloadStatusCompleted:
					if _, statErr := os.Stat(ext.SavePath); statErr == nil {
						logger.Info("Trainer version already completed and file exists", "task_id", ext.ID, "path", ext.SavePath)
						return ext, nil
					}
					// 磁盘文件已被删除，重置任务状态并从头重新下载
					ext.DownloadedBytes = 0
					ext.Progress = 0
					ext.Status = model.DownloadStatusPending
					ext.ErrorMsg = ""
					_ = m.dlRepo.UpdateTask(ctx, ext)
					m.taskQueue <- ext.ID
					m.notifyListeners(ext)
					return ext, nil
				case model.DownloadStatusFailed:
					_ = m.ResumeTask(ext.ID)
					return ext, nil
				}
			}
		}
	}

	version, err := m.trainerRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trainer version: %w", err)
	}

	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download dir: %w", err)
	}

	fileName := version.FileName
	if fileName == "" {
		fileName = fmt.Sprintf("trainer_%d.zip", versionID)
	}
	savePath := filepath.Join(downloadDir, fileName)

	gameName := ""
	trainer, _ := m.trainerRepo.GetTrainerByID(ctx, version.TrainerID)
	if trainer != nil {
		if trainer.TitleCN != "" {
			gameName = trainer.TitleCN
		} else {
			gameName = trainer.TitleEN
		}
	}

	task := &model.DownloadTask{
		VersionID:       versionID,
		GameName:        gameName,
		VersionName:     version.VersionName,
		FileName:        fileName,
		SavePath:        savePath,
		TotalBytes:      version.FileSize,
		DownloadedBytes: 0,
		Status:          model.DownloadStatusPending,
	}

	if err := m.dlRepo.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save download task: %w", err)
	}

	// 入队调度
	m.taskQueue <- task.ID
	m.notifyListeners(task)
	return task, nil
}

// PauseAllTasks 暂停所有进行中或排队中的下载任务
func (m *DownloadManager) PauseAllTasks(ctx context.Context) error {
	tasks, err := m.dlRepo.ListTasks(ctx, "")
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.Status == model.DownloadStatusDownloading || t.Status == model.DownloadStatusPending || t.Status == model.DownloadStatusRetrying {
			_ = m.PauseTask(t.ID)
		}
	}
	return nil
}

// ResumeAllTasks 继续所有暂停或失败的下载任务
func (m *DownloadManager) ResumeAllTasks(ctx context.Context) error {
	tasks, err := m.dlRepo.ListTasks(ctx, "")
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.Status == model.DownloadStatusPaused || t.Status == model.DownloadStatusFailed {
			_ = m.ResumeTask(t.ID)
		}
	}
	return nil
}

// PauseTask 暂停正在执行的下载任务
func (m *DownloadManager) PauseTask(taskID int64) error {
	m.mu.Lock()
	cancel, ok := m.activeTasks[taskID]
	if ok {
		cancel()
		delete(m.activeTasks, taskID)
	}
	m.mu.Unlock()

	ctx := context.Background()
	task, err := m.dlRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = model.DownloadStatusPaused
	task.SpeedBytesPerSec = 0
	_ = m.dlRepo.UpdateTask(ctx, task)
	m.notifyListeners(task)
	return nil
}

// ResumeTask 恢复暂停的下载任务
func (m *DownloadManager) ResumeTask(taskID int64) error {
	ctx := context.Background()
	task, err := m.dlRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = model.DownloadStatusPending
	task.ErrorMsg = ""
	task.Retries = 0
	_ = m.dlRepo.UpdateTask(ctx, task)

	m.taskQueue <- task.ID
	m.notifyListeners(task)
	return nil
}

// CancelTask 取消并删除下载任务及临时文件
func (m *DownloadManager) CancelTask(taskID int64) error {
	_ = m.PauseTask(taskID)

	ctx := context.Background()
	task, err := m.dlRepo.GetTaskByID(ctx, taskID)
	if err == nil && task != nil {
		_ = os.Remove(task.SavePath + ".tmp")
		_ = os.Remove(task.SavePath)
	}
	return m.dlRepo.DeleteTask(ctx, taskID)
}

// dispatchLoop 统一任务分发调度循环
func (m *DownloadManager) dispatchLoop() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case taskID := <-m.taskQueue:
			m.condMu.Lock()
			for m.activeCount >= m.maxConcurrent && m.ctx.Err() == nil {
				m.cond.Wait()
			}
			if m.ctx.Err() != nil {
				m.condMu.Unlock()
				return
			}
			m.activeCount++
			m.condMu.Unlock()

			go func(tID int64) {
				defer func() {
					m.condMu.Lock()
					m.activeCount--
					m.cond.Signal()
					m.condMu.Unlock()
				}()
				m.processTask(tID)
			}(taskID)
		}
	}
}

// SetMaxConcurrent 动态调整最大并发下载任务数并广播唤醒等待中的任务
func (m *DownloadManager) SetMaxConcurrent(n int) {
	if n < 1 {
		n = 1
	}
	if n > 10 {
		n = 10
	}
	m.condMu.Lock()
	m.maxConcurrent = n
	m.cond.Broadcast()
	m.condMu.Unlock()
	logger.Info("Dynamically updated max concurrent downloads", "maxConcurrent", n)
}

// HasActiveTasks 检查当前是否存在下载中或调度中的活跃任务
func (m *DownloadManager) HasActiveTasks() bool {
	m.mu.RLock()
	hasActiveMap := len(m.activeTasks) > 0
	m.mu.RUnlock()

	m.condMu.Lock()
	hasCount := m.activeCount > 0
	m.condMu.Unlock()

	return hasActiveMap || hasCount
}

func (m *DownloadManager) processTask(taskID int64) {
	ctx, cancel := context.WithCancel(m.ctx)

	m.mu.Lock()
	m.activeTasks[taskID] = cancel
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.activeTasks, taskID)
		m.mu.Unlock()
	}()

	task, err := m.dlRepo.GetTaskByID(ctx, taskID)
	if err != nil || task == nil {
		return
	}

	version, err := m.trainerRepo.GetVersionByID(ctx, task.VersionID)
	if err != nil || version == nil {
		task.Status = model.DownloadStatusFailed
		task.ErrorMsg = "Version not found"
		_ = m.dlRepo.UpdateTask(ctx, task)
		m.notifyListeners(task)
		return
	}

	// 查找关联修改器以获取真实 DetailURL 作为 Referer (解决防盗链 403 异常)
	var referer string
	trainer, _ := m.trainerRepo.GetTrainerByID(ctx, version.TrainerID)
	if trainer != nil && trainer.DetailURL != "" {
		referer = trainer.DetailURL
	} else if strings.Contains(version.DownloadURL, "flingtrainer.com") {
		referer = "https://flingtrainer.com/"
	}

	// 解析真实下载直链
	realDownloadURL := version.DownloadURL
	if m.adapter != nil {
		dlInfo, resErr := m.adapter.ResolveDownload(ctx, version.DownloadURL, referer)
		if resErr == nil && dlInfo != nil && dlInfo.FinalURL != "" {
			realDownloadURL = dlInfo.FinalURL
			if dlInfo.FileSize > 0 && task.TotalBytes == 0 {
				task.TotalBytes = dlInfo.FileSize
			}
			if dlInfo.FileName != "" && task.FileName != dlInfo.FileName {
				task.FileName = dlInfo.FileName
				task.SavePath = filepath.Join(filepath.Dir(task.SavePath), dlInfo.FileName)
			}
		} else if resErr != nil {
			logger.Warn("Failed to resolve download URL, using direct URL", "download_url", version.DownloadURL, "err", resErr)
		}
	}

	task.Status = model.DownloadStatusDownloading
	task.ErrorMsg = ""
	_ = m.dlRepo.UpdateTask(ctx, task)
	m.notifyListeners(task)

	// 创建并执行 Worker
	lastDBSave := time.Now()
	worker := NewWorker(m.httpClient, task, realDownloadURL, referer, func(t *model.DownloadTask) {
		m.notifyListeners(t)
		// 每隔 1 秒持久化一次进度到 SQLite
		if time.Since(lastDBSave) >= time.Second {
			_ = m.dlRepo.UpdateTask(context.Background(), t)
			lastDBSave = time.Now()
		}
	})

	res, err := worker.Execute(ctx)
	if err != nil {
		if ctx.Err() == context.Canceled {
			task.Status = model.DownloadStatusPaused
		} else {
			const maxRetries = 3
			if task.Retries < maxRetries {
				task.Retries++
				task.Status = model.DownloadStatusRetrying
				task.ErrorMsg = fmt.Sprintf("下载遇到异常: %v (正在进行第 %d/%d 次重试...)", err, task.Retries, maxRetries)
				logger.Warn("Download task failed, scheduling auto retry", "task_id", task.ID, "retry", task.Retries, "err", err)
				_ = m.dlRepo.UpdateTask(context.Background(), task)
				m.notifyListeners(task)

				// 异步退避重试 (500ms, 1s, 2s)
				go func(tID int64, attempt int) {
					backoff := time.Duration(1<<uint(attempt-1)) * 500 * time.Millisecond
					time.Sleep(backoff)
					m.taskQueue <- tID
				}(task.ID, task.Retries)
				return
			}

			task.Status = model.DownloadStatusFailed
			task.ErrorMsg = fmt.Sprintf("下载失败: %v", err)
			logger.Error("Download task failed permanently after max retries", "task_id", task.ID, "err", err)
			if m.OnTaskFail != nil {
				m.OnTaskFail(task, err)
			}
		}
		_ = m.dlRepo.UpdateTask(context.Background(), task)
		m.notifyListeners(task)
		return
	}

	// 下载完成！
	task.Status = model.DownloadStatusCompleted
	task.DownloadedBytes = res.FileSize
	task.TotalBytes = res.FileSize
	task.Progress = 100.0
	task.SpeedBytesPerSec = 0
	_ = m.dlRepo.UpdateTask(context.Background(), task)
	m.notifyListeners(task)
	if m.OnTaskComplete != nil {
		m.OnTaskComplete(task)
	}

	// 1. 记入下载历史表
	hist := &model.DownloadHistory{
		TaskID:    task.ID,
		TrainerID: version.TrainerID,
		VersionID: version.ID,
		GameName:  task.GameName,
		FileName:  filepath.Base(res.FinalPath),
		FileSize:  res.FileSize,
		SHA256:    res.SHA256,
		SavePath:  res.FinalPath,
	}
	_ = m.dlRepo.AddHistory(context.Background(), hist)

	// 2. 自动注册入本地修改器库 (local_trainers)
	trainer, _ = m.trainerRepo.GetTrainerByID(context.Background(), version.TrainerID)
	var gameID int64
	if trainer != nil {
		gameID = trainer.GameID
	}

	local := &model.LocalTrainer{
		GameID:     gameID,
		VersionID:  &version.ID,
		FilePath:   res.FinalPath,
		FileSize:   res.FileSize,
		SHA256:     res.SHA256,
		SourceName: "FLiNG",
		IsCustom:   false,
	}
	_ = m.localRepo.Create(context.Background(), local)

	logger.Info("Download task successfully completed and registered", "file", res.FinalPath, "sha256", res.SHA256)
}

// Close 关闭下载管理器
func (m *DownloadManager) Close() {
	m.cancel()
	m.condMu.Lock()
	m.cond.Broadcast()
	m.condMu.Unlock()
}

