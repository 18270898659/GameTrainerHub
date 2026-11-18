package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"
)

type ProgressCallback func(task *model.DownloadTask)

type Worker struct {
	client           *http.Client
	task             *model.DownloadTask
	downloadURL      string
	referer          string
	progressCallback ProgressCallback
}

func NewWorker(client *http.Client, task *model.DownloadTask, downloadURL string, referer string, cb ProgressCallback) *Worker {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Worker{
		client:           client,
		task:             task,
		downloadURL:      downloadURL,
		referer:          referer,
		progressCallback: cb,
	}
}

type DownloadResult struct {
	FinalPath string
	SHA256    string
	FileSize  int64
}

// Execute 执行单任务断点续传下载
func (w *Worker) Execute(ctx context.Context) (*DownloadResult, error) {
	tmpPath := w.task.SavePath + ".tmp"
	if err := os.MkdirAll(filepath.Dir(tmpPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 1. 检查本地是否已有部分下载的 .tmp 文件
	var localOffset int64
	if fi, err := os.Stat(tmpPath); err == nil {
		localOffset = fi.Size()
	}

	// 2. 探测目标文件总大小与 Range 支持
	totalSize, supportsRange, err := w.probeResource(ctx)
	if err != nil {
		logger.Warn("Probe failed, falling back to direct GET", "err", err)
	}
	if totalSize > 0 {
		w.task.TotalBytes = totalSize
	}

	// 3. 构建 GET 请求（若支持 Range 且已有部分文件，则请求 Range）
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if w.referer != "" {
		req.Header.Set("Referer", w.referer)
	}

	isPartial := false
	if localOffset > 0 && supportsRange && (totalSize == 0 || localOffset < totalSize) {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", localOffset))
		isPartial = true
		w.task.DownloadedBytes = localOffset
		logger.Info("Resuming download via HTTP Range", "offset", localOffset, "total", totalSize)
	} else {
		// 重新从头下载
		localOffset = 0
		w.task.DownloadedBytes = 0
	}

	resp, err := w.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			w.task.Status = model.DownloadStatusPaused
		}
		return nil, err
	}
	defer resp.Body.Close()

	if isPartial && resp.StatusCode != http.StatusPartialContent {
		// 服务端未返回 206，说明 Range 未生效，重置为从头写入
		logger.Warn("Server ignored Range header, restarting from zero", "status", resp.StatusCode)
		localOffset = 0
		w.task.DownloadedBytes = 0
	} else if !isPartial && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	// 更新总大小
	if w.task.TotalBytes == 0 && resp.ContentLength > 0 {
		w.task.TotalBytes = localOffset + resp.ContentLength
	}

	// 4. 打开目标 .tmp 文件
	fileFlag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if localOffset == 0 {
		fileFlag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	outFile, err := os.OpenFile(tmpPath, fileFlag, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}
	defer outFile.Close()

	// 5. 循环读取流数据并统计速度与进度
	buf := make([]byte, 64*1024) // 64KB 缓冲区
	lastSpeedTick := time.Now()
	var bytesSinceLastTick int64
	lastCallbackTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			w.task.Status = model.DownloadStatusPaused
			return nil, ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
				return nil, fmt.Errorf("file write error: %w", writeErr)
			}

			w.task.DownloadedBytes += int64(n)
			bytesSinceLastTick += int64(n)

			now := time.Now()
			// 速度计算平滑窗口（每 500ms 计算一次）
			timeSinceLastTick := now.Sub(lastSpeedTick)
			if timeSinceLastTick >= 500*time.Millisecond {
				w.task.SpeedBytesPerSec = int64(float64(bytesSinceLastTick) / timeSinceLastTick.Seconds())
				bytesSinceLastTick = 0
				lastSpeedTick = now
			}

			if w.task.TotalBytes > 0 {
				w.task.Progress = float64(w.task.DownloadedBytes) / float64(w.task.TotalBytes) * 100
			}

			// 进度回调推送（每 200ms 最多一次，避免高频刷屏）
			if now.Sub(lastCallbackTime) >= 200*time.Millisecond {
				if w.progressCallback != nil {
					w.progressCallback(w.task)
				}
				lastCallbackTime = now
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, readErr
		}
	}

	_ = outFile.Sync()
	_ = outFile.Close()

	// 6. 校验文件大小
	finalFI, err := os.Stat(tmpPath)
	if err != nil {
		return nil, err
	}
	if w.task.TotalBytes > 0 && finalFI.Size() < w.task.TotalBytes {
		return nil, fmt.Errorf("download incomplete: got %d bytes, expected %d", finalFI.Size(), w.task.TotalBytes)
	}

	// 7. 计算 SHA-256
	hashStr, err := calculateSHA256(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate sha256: %w", err)
	}

	// 8. 重命名 .tmp 文件为正式文件
	finalPath := w.task.SavePath
	_ = os.Remove(finalPath) // 覆盖已有旧文件
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return nil, fmt.Errorf("failed to rename temp file: %w", err)
	}

	// 9. 若为 ZIP 包，安全解压出核心 EXE 修改器
	if strings.HasSuffix(strings.ToLower(finalPath), ".zip") {
		destDir := filepath.Join(filepath.Dir(finalPath), strings.TrimSuffix(filepath.Base(finalPath), ".zip"))
		extractedExe, extractErr := ExtractTrainerZip(finalPath, destDir)
		if extractErr == nil && extractedExe != "" {
			finalPath = extractedExe
		}
	}

	w.task.Status = model.DownloadStatusCompleted
	w.task.Progress = 100.0
	w.task.SpeedBytesPerSec = 0
	if w.progressCallback != nil {
		w.progressCallback(w.task)
	}

	return &DownloadResult{
		FinalPath: finalPath,
		SHA256:    hashStr,
		FileSize:  finalFI.Size(),
	}, nil
}

func (w *Worker) probeResource(ctx context.Context) (totalSize int64, supportsRange bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, w.downloadURL, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if w.referer != "" {
		req.Header.Set("Referer", w.referer)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			w.task.Status = model.DownloadStatusPaused
		}
		return 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
		supportsRange = (resp.Header.Get("Accept-Ranges") == "bytes")
		if resp.ContentLength > 0 {
			totalSize = resp.ContentLength
		}
		// 如果 HEAD 请求没有 Content-Length，尝试 Range: bytes=0-0 探测
		if totalSize == 0 {
			return w.probeWithRange(ctx)
		}
		return totalSize, supportsRange, nil
	}

	return 0, false, fmt.Errorf("HEAD returned status %d", resp.StatusCode)
}

func (w *Worker) probeWithRange(ctx context.Context) (int64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.downloadURL, nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	if w.referer != "" {
		req.Header.Set("Referer", w.referer)
	}
	req.Header.Set("Range", "bytes=0-0")

	resp, err := w.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			w.task.Status = model.DownloadStatusPaused
		}
		return 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPartialContent {
		contentRange := resp.Header.Get("Content-Range")
		if contentRange != "" {
			parts := strings.Split(contentRange, "/")
			if len(parts) == 2 {
				total, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err == nil {
					return total, true, nil
				}
			}
		}
		return 0, true, nil
	}
	return resp.ContentLength, false, nil
}

func calculateSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

