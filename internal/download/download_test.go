package download

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
)

// 创建模拟测试数据
func generateTestData(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i % 251)
	}
	return data
}

// 模拟支持 Range 的 HTTP 下载服务器
func newMockServer(testData []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Type", "application/octet-stream")

		totalSize := len(testData)
		rangeHeader := r.Header.Get("Range")

		if rangeHeader == "" {
			w.Header().Set("Content-Length", strconv.Itoa(totalSize))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(testData)
			return
		}

		// 解析 Range: bytes=start-end 或 bytes=start-
		var start, end int64
		rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
		parts := strings.Split(rangeStr, "-")

		start, _ = strconv.ParseInt(parts[0], 10, 64)
		if len(parts) > 1 && parts[1] != "" {
			end, _ = strconv.ParseInt(parts[1], 10, 64)
		} else {
			end = int64(totalSize - 1)
		}

		if start >= int64(totalSize) {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}

		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, totalSize))
		w.Header().Set("Content-Length", strconv.Itoa(int(end-start+1)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(testData[start : end+1])
	}))
}

func TestDownloadFullAndHash(t *testing.T) {
	data := generateTestData(512 * 1024) // 512KB
	expectedHashBytes := sha256.Sum256(data)
	expectedHash := hex.EncodeToString(expectedHashBytes[:])

	server := newMockServer(data)
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "gth_dl_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	savePath := filepath.Join(tempDir, "test_file.bin")
	task := &model.DownloadTask{
		SavePath:   savePath,
		TotalBytes: int64(len(data)),
		Status:     model.DownloadStatusPending,
	}

	worker := NewWorker(server.Client(), task, server.URL, "", nil)
	ctx := context.Background()

	res, err := worker.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if res.FileSize != int64(len(data)) {
		t.Errorf("expected size %d, got %d", len(data), res.FileSize)
	}
	if res.SHA256 != expectedHash {
		t.Errorf("expected hash %s, got %s", expectedHash, res.SHA256)
	}

	// 验证最终文件存在且无 .tmp 后缀
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Errorf("expected final file %s to exist", savePath)
	}
	if _, err := os.Stat(savePath + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("expected tmp file to be removed after completion")
	}
}

func TestDownloadRangeResumption(t *testing.T) {
	data := generateTestData(1024 * 1024) // 1MB
	expectedHashBytes := sha256.Sum256(data)
	expectedHash := hex.EncodeToString(expectedHashBytes[:])

	server := newMockServer(data)
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "gth_range_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	savePath := filepath.Join(tempDir, "resume_file.bin")
	tmpPath := savePath + ".tmp"

	// 模拟已提前下载了前 300KB
	initialOffset := 300 * 1024
	if err := os.WriteFile(tmpPath, data[:initialOffset], 0644); err != nil {
		t.Fatalf("failed to write initial temp file: %v", err)
	}

	task := &model.DownloadTask{
		SavePath:        savePath,
		TotalBytes:      int64(len(data)),
		DownloadedBytes: int64(initialOffset),
		Status:          model.DownloadStatusPending,
	}

	worker := NewWorker(server.Client(), task, server.URL, "", nil)
	ctx := context.Background()

	// 断点续传执行剩余 724KB
	res, err := worker.Execute(ctx)
	if err != nil {
		t.Fatalf("Resume execution failed: %v", err)
	}

	if res.FileSize != int64(len(data)) {
		t.Errorf("expected total size %d after resume, got %d", len(data), res.FileSize)
	}
	if res.SHA256 != expectedHash {
		t.Errorf("resumed file hash mismatch! expected %s, got %s", expectedHash, res.SHA256)
	}

	// 校验磁盘内容是否与原始数据完全一致
	diskData, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("failed to read disk file: %v", err)
	}
	if !bytes.Equal(diskData, data) {
		t.Fatal("resumed disk file bytes do not match expected data!")
	}
}

func TestZipExtraction(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_zip_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "trainer.zip")

	// 构建一个包含 exe 的测试 zip 文件
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	f, err := zipWriter.Create("Elden Ring Trainer.exe")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = f.Write([]byte("MZ_MOCK_EXE_BINARY_DATA"))
	_ = zipWriter.Close()

	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write zip file: %v", err)
	}

	destDir := filepath.Join(tempDir, "extracted")
	mainExe, err := ExtractTrainerZip(zipPath, destDir)
	if err != nil {
		t.Fatalf("ExtractTrainerZip failed: %v", err)
	}

	if !strings.HasSuffix(mainExe, "Elden Ring Trainer.exe") {
		t.Errorf("expected extracted exe path, got '%s'", mainExe)
	}
	if _, err := os.Stat(mainExe); os.IsNotExist(err) {
		t.Errorf("extracted exe does not exist at %s", mainExe)
	}
}

func TestDownloadManagerIntegration(t *testing.T) {
	data := generateTestData(256 * 1024)
	server := newMockServer(data)
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "gth_mgr_test_*")
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
	trainerRepo := repository.NewTrainerRepository(db)
	dlRepo := repository.NewDownloadRepository(db)
	localRepo := repository.NewLocalTrainerRepository(db)

	game := &model.Game{Slug: "cyberpunk", NameEN: "Cyberpunk 2077"}
	_ = gameRepo.Create(ctx, game)
	trainer := &model.Trainer{GameID: game.ID, SourceID: "fling", ExternalID: "cp-tr", TitleEN: "Cyberpunk Trainer"}
	_ = trainerRepo.CreateTrainer(ctx, trainer)
	ver := &model.TrainerVersion{
		TrainerID:   trainer.ID,
		VersionName: "v2.0",
		DownloadURL: server.URL,
		FileName:    "cyberpunk_trainer.bin",
		FileSize:    int64(len(data)),
	}
	_ = trainerRepo.CreateVersion(ctx, ver)

	mgr := NewDownloadManager(dlRepo, trainerRepo, localRepo, nil, 2)
	defer mgr.Close()

	var wg sync.WaitGroup
	var onceDone sync.Once
	wg.Add(1)
	var completedTask *model.DownloadTask

	mgr.AddListener(func(task *model.DownloadTask) {
		if task.Status == model.DownloadStatusCompleted {
			completedTask = task
			onceDone.Do(func() {
				wg.Done()
			})
		}
	})

	downloadDir := filepath.Join(tempDir, "downloads")
	task, err := mgr.CreateTask(ctx, ver.ID, downloadDir)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.Status != model.DownloadStatusPending {
		t.Errorf("expected status pending, got %s", task.Status)
	}

	// 等待下载完成通知
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		t.Log("Download completed successfully via Manager!")
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for download task completion")
	}

	if completedTask == nil || completedTask.Progress < 99.9 {
		t.Fatalf("expected 100%% completed task, got %+v", completedTask)
	}

	// 验证本地修改器库已自动注册
	localList, err := localRepo.ListByGameID(ctx, game.ID)
	if err != nil || len(localList) != 1 {
		t.Fatalf("expected 1 local trainer registered for game, got %d, err: %v", len(localList), err)
	}

	// 验证下载历史记录
	histList, err := dlRepo.ListHistory(ctx, 10)
	if err != nil || len(histList) != 1 {
		t.Fatalf("expected 1 download history item, got %d, err: %v", len(histList), err)
	}
}
