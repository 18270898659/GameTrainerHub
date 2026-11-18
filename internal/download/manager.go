package download

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	ctx         context.Context
	cancel      context.CancelFunc
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

	dm := &DownloadManager{
		dlRepo:        dlRepo,
		trainerRepo:   trainerRepo,
		localRepo:     localRepo,
		adapter:       adapter,
		httpClient:    &http.Client{Timeout: 60 * time.Second},
		maxConcurrent: maxConcurrent,
		activeTasks:   make(map[int64]context.CancelFunc),
		taskQueue:     make(chan int64, 100),
		ctx:           ctx,
		cancel:        cancel,
	}

	// 启动 Worker Pool
	for i := 0; i < dm.maxConcurrent; i++ {
		go dm.workerLoop(i)
	}

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

// CreateTask 创建新下载任务并入队
func (m *DownloadManager) CreateTask(ctx context.Context, versionID int64, downloadDir string) (*model.DownloadTask, error) {
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

	task := &model.DownloadTask{
		VersionID:       versionID,
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

func (m *DownloadManager) workerLoop(workerID int) {
	for {
		select {
		case <-m.ctx.Done():
			return
		case taskID := <-m.taskQueue:
			m.processTask(taskID)
		}
	}
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

	// 解析真实下载直链
	realDownloadURL := version.DownloadURL
	if m.adapter != nil {
		dlInfo, resErr := m.adapter.ResolveDownload(ctx, version.DownloadURL, "")
		if resErr == nil && dlInfo != nil && dlInfo.FinalURL != "" {
			realDownloadURL = dlInfo.FinalURL
			if dlInfo.FileSize > 0 && task.TotalBytes == 0 {
				task.TotalBytes = dlInfo.FileSize
			}
		}
	}

	task.Status = model.DownloadStatusDownloading
	_ = m.dlRepo.UpdateTask(ctx, task)
	m.notifyListeners(task)

	// 创建并执行 Worker
	lastDBSave := time.Now()
	worker := NewWorker(m.httpClient, task, realDownloadURL, "", func(t *model.DownloadTask) {
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
			task.Status = model.DownloadStatusFailed
			task.ErrorMsg = err.Error()
			logger.Error("Download task failed", "task_id", task.ID, "err", err)
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
	trainer, _ := m.trainerRepo.GetTrainerByID(context.Background(), version.TrainerID)
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
}

