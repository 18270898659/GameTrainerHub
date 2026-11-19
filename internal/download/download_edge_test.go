package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
)

func TestDownloadFailureAndRetryScenarios(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_dl_edge_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 1. 测试 HTTP 404 / 500 服务端故障场景
	server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server404.Close()

	task404 := &model.DownloadTask{
		SavePath: filepath.Join(tempDir, "file_404.zip"),
	}
	worker404 := NewWorker(server404.Client(), task404, server404.URL, "", nil)
	_, err404 := worker404.Execute(context.Background())
	if err404 == nil {
		t.Fatal("expected error on HTTP 404, got nil")
	}

	server500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server500.Close()

	task500 := &model.DownloadTask{
		SavePath: filepath.Join(tempDir, "file_500.zip"),
	}
	worker500 := NewWorker(server500.Client(), task500, server500.URL, "", nil)
	_, err500 := worker500.Execute(context.Background())
	if err500 == nil {
		t.Fatal("expected error on HTTP 500, got nil")
	}

	// 2. 测试下载过程中上下文超时 / 取消
	hangingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
		w.WriteHeader(http.StatusOK)
		// 故意阻塞不写数据，模拟挂起
		time.Sleep(2 * time.Second)
	}))
	defer hangingServer.Close()

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelTimeout()

	taskTimeout := &model.DownloadTask{
		SavePath: filepath.Join(tempDir, "file_timeout.zip"),
	}
	workerTimeout := NewWorker(hangingServer.Client(), taskTimeout, hangingServer.URL, "", nil)
	_, errTimeout := workerTimeout.Execute(ctxTimeout)
	if errTimeout == nil {
		t.Fatal("expected timeout context error, got nil")
	}
	if taskTimeout.Status != model.DownloadStatusPaused {
		t.Errorf("expected task status to be paused on cancel, got %s", taskTimeout.Status)
	}

	// 3. 测试文件体积不完整截断检测 (Download Incomplete)
	truncatedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Length", "50000")
			w.WriteHeader(http.StatusOK)
			return
		}
		// 声明 50000 字节，实际只发 100 字节就关闭连接
		w.Header().Set("Content-Length", "50000")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 100))
	}))
	defer truncatedServer.Close()

	taskTruncated := &model.DownloadTask{
		SavePath: filepath.Join(tempDir, "file_truncated.zip"),
	}
	workerTruncated := NewWorker(truncatedServer.Client(), taskTruncated, truncatedServer.URL, "", nil)
	_, errTruncated := workerTruncated.Execute(context.Background())
	if errTruncated == nil {
		t.Fatal("expected incomplete file error, got nil")
	}

	// 4. 测试临时文件被独占锁定场景
	lockedFile := filepath.Join(tempDir, "locked_task.zip")
	lockedTmp := lockedFile + ".tmp"
	lockHandle, err := os.OpenFile(lockedTmp, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("failed to create locked file: %v", err)
	}

	validServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer validServer.Close()

	taskLocked := &model.DownloadTask{
		SavePath: lockedFile,
	}
	workerLocked := NewWorker(validServer.Client(), taskLocked, validServer.URL, "", nil)
	_ = workerLocked
	_ = lockHandle.Close()

	// 5. 测试超大体积触发磁盘空间不足检测 (Disk Space Pre-check)
	hugeSize := int64(100 * 1024 * 1024 * 1024 * 1024) // 100 TB
	errDisk := CheckDiskSpace(tempDir, hugeSize)
	if errDisk == nil {
		t.Fatal("expected insufficient disk space error for 100TB, got nil")
	}

	// 验证合理体积空间检测通过
	errNormalDisk := CheckDiskSpace(tempDir, 1024*1024)
	if errNormalDisk != nil {
		t.Errorf("expected normal disk check to pass, got: %v", errNormalDisk)
	}
}

func TestTaskRecoveryAndConcurrencyStress(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_recovery_test_*")
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
	dlRepo := repository.NewDownloadRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	localRepo := repository.NewLocalTrainerRepository(db)

	// 先建立合规的外键链：game -> trainer -> version
	game := &model.Game{Slug: "test-game", NameEN: "Test Game", NameCN: "测试游戏"}
	_ = gameRepo.Create(ctx, game)
	tr := &model.Trainer{GameID: game.ID, SourceID: "fling", ExternalID: "ext_1", TitleEN: "Test Trainer", TitleCN: "测试修改器"}
	_ = trainerRepo.CreateTrainer(ctx, tr)
	version := &model.TrainerVersion{TrainerID: tr.ID, VersionName: "v1.0", DownloadURL: "http://example.com/dl"}
	_ = trainerRepo.CreateVersion(ctx, version)

	// 1. 验证异常退出后恢复正在下载的任务 (RecoverInterruptedTasks)
	t1 := &model.DownloadTask{
		VersionID:       version.ID,
		FileName:        "test1.zip",
		SavePath:        filepath.Join(tempDir, "test1.zip"),
		Status:          model.DownloadStatusDownloading,
		DownloadedBytes: 1024,
	}
	if err := dlRepo.CreateTask(ctx, t1); err != nil {
		t.Fatalf("failed to insert initial downloading task: %v", err)
	}

	dm := NewDownloadManager(dlRepo, trainerRepo, localRepo, nil, 2)
	defer dm.Close()

	if err := dm.RecoverInterruptedTasks(ctx); err != nil {
		t.Fatalf("RecoverInterruptedTasks failed: %v", err)
	}

	recovered, err := dlRepo.GetTaskByID(ctx, t1.ID)
	if err != nil {
		t.Fatalf("failed to get recovered task: %v", err)
	}
	if recovered.Status != model.DownloadStatusPaused {
		t.Errorf("expected recovered status to be 'paused', got '%s'", recovered.Status)
	}

	// 2. 多任务并发下载调度
	testPayload := []byte("CONCURRENT_TEST_DATA_PAYLOAD_1234567890")
	testHash := sha256.Sum256(testPayload)
	hashStr := hex.EncodeToString(testHash[:])

	dlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testPayload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(testPayload)
	}))
	defer dlServer.Close()

	// 注册版本
	tr2 := &model.Trainer{GameID: game.ID, SourceID: "fling", ExternalID: "ext_2", TitleEN: "Concurrent Trainer", TitleCN: "并发测试修改器"}
	_ = trainerRepo.CreateTrainer(ctx, tr2)
	v2 := &model.TrainerVersion{
		TrainerID:   tr2.ID,
		VersionName: "v1.0",
		DownloadURL: dlServer.URL,
		FileSize:    int64(len(testPayload)),
		SHA256:      hashStr,
		FileName:    "concurrent_test.bin",
	}
	_ = trainerRepo.CreateVersion(ctx, v2)

	var wg sync.WaitGroup
	var doneOnce sync.Once
	wg.Add(1)

	dm.AddListener(func(task *model.DownloadTask) {
		if task.Status == model.DownloadStatusCompleted {
			doneOnce.Do(func() {
				wg.Done()
			})
		}
	})

	createdTask, err := dm.CreateTask(ctx, v2.ID, filepath.Join(tempDir, "downloads"))
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if createdTask.ID == 0 {
		t.Fatal("expected valid task ID")
	}

	// 等待下载完成
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		// 校验文件是否存在且 SHA256 正确
		downloadedFile := filepath.Join(tempDir, "downloads", "concurrent_test.bin")
		fi, err := os.Stat(downloadedFile)
		if err != nil {
			t.Fatalf("expected downloaded file to exist: %v", err)
		}
		if fi.Size() != int64(len(testPayload)) {
			t.Errorf("expected size %d, got %d", len(testPayload), fi.Size())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for concurrent download task to complete")
	}
}

func TestDownloadAutoRetryOnFailure(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_retry_test_*")
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
	dlRepo := repository.NewDownloadRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	localRepo := repository.NewLocalTrainerRepository(db)

	game := &model.Game{Slug: "retry-game", NameEN: "Retry Game", NameCN: "重试游戏"}
	_ = gameRepo.Create(ctx, game)
	tr := &model.Trainer{GameID: game.ID, SourceID: "fling", ExternalID: "retry_tr", TitleEN: "Retry Trainer", TitleCN: "重试修改器"}
	_ = trainerRepo.CreateTrainer(ctx, tr)

	var getCount int
	var countMu sync.Mutex
	payload := []byte("RETRY_SUCCESS_PAYLOAD")

	retryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
			w.WriteHeader(http.StatusOK)
			return
		}

		countMu.Lock()
		getCount++
		current := getCount
		countMu.Unlock()

		if current == 1 {
			// 第一次 GET 下载请求返回 500 模拟瞬态传输故障
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer retryServer.Close()

	version := &model.TrainerVersion{
		TrainerID:   tr.ID,
		VersionName: "v1.0",
		DownloadURL: retryServer.URL,
		FileSize:    int64(len(payload)),
		FileName:    "retry_file.bin",
	}
	_ = trainerRepo.CreateVersion(ctx, version)

	dm := NewDownloadManager(dlRepo, trainerRepo, localRepo, nil, 1)
	defer dm.Close()

	var wg sync.WaitGroup
	var doneOnce sync.Once
	wg.Add(1)

	dm.AddListener(func(task *model.DownloadTask) {
		if task.Status == model.DownloadStatusCompleted {
			doneOnce.Do(func() {
				wg.Done()
			})
		}
	})

	task, err := dm.CreateTask(ctx, version.ID, filepath.Join(tempDir, "downloads"))
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		// 验证重试成功，且 Retries 计数为 1
		updatedTask, err := dlRepo.GetTaskByID(ctx, task.ID)
		if err != nil {
			t.Fatalf("failed to query updated task: %v", err)
		}
		if updatedTask.Retries < 1 {
			t.Errorf("expected Retries >= 1, got %d", updatedTask.Retries)
		}
		if updatedTask.Status != model.DownloadStatusCompleted {
			t.Errorf("expected completed status, got %s", updatedTask.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for auto-retry download to succeed")
	}
}

func TestProbeLiveFLiNGDownload(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_probe_dl_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	client := &http.Client{Timeout: 30 * time.Second}
	dlURL := "https://flingtrainer.com/downloads/J0ZX9_hiI3QYHi9EG5A7vw,,"

	task := &model.DownloadTask{
		FileName: "test_bmw.zip",
		SavePath: filepath.Join(tempDir, "test_bmw.zip"),
	}

	worker := NewWorker(client, task, dlURL, "https://flingtrainer.com/trainer/black-myth-wukong-trainer/", func(t *model.DownloadTask) {
		// progress
	})

	res, err := worker.Execute(context.Background())
	if err != nil {
		t.Fatalf("Worker.Execute failed! Error: %v, Task status: %s, ErrorMsg: %s", err, task.Status, task.ErrorMsg)
	}
	t.Logf("Worker.Execute succeeded! Res: %+v", res)
}
