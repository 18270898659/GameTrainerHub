package download

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"GameTrainerHub/internal/logger"
)

// ExtractTrainerZip 解压 ZIP 修改器包，返回主可执行文件 (.exe) 的路径
func ExtractTrainerZip(zipPath string, destDir string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	destDirClean := filepath.Clean(destDir)
	if err := os.MkdirAll(destDirClean, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination dir: %w", err)
	}

	var mainExePath string
	var totalExtractedBytes int64
	const maxFileSize = 500 * 1024 * 1024       // 单文件上限 500MB，防 Zip 炸弹
	const maxTotalSize = 2 * 1024 * 1024 * 1024 // 总体积上限 2GB

	for _, file := range reader.File {
		// 严格防止 Zip Slip 路径穿越漏洞
		targetPath := filepath.Join(destDirClean, file.Name)
		cleanTarget := filepath.Clean(targetPath)

		// 目标路径必须严格位于 destDirClean 目录下，禁止跳出
		if cleanTarget != destDirClean && !strings.HasPrefix(cleanTarget, destDirClean+string(os.PathSeparator)) {
			logger.Warn("Blocked illegal zip slip path traversal", "entry", file.Name, "target", cleanTarget)
			continue
		}

		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(cleanTarget, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
			return "", err
		}

		srcFile, err := file.Open()
		if err != nil {
			return "", err
		}

		destFile, err := os.OpenFile(cleanTarget, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return "", err
		}

		// 使用 LimitReader 防范 Zip Bomb 耗尽磁盘
		limitedReader := io.LimitReader(srcFile, maxFileSize)
		n, copyErr := io.Copy(destFile, limitedReader)
		srcFile.Close()
		destFile.Close()
		if copyErr != nil {
			return "", copyErr
		}

		totalExtractedBytes += n
		if totalExtractedBytes > maxTotalSize {
			return "", fmt.Errorf("zip decompression exceeded max total size limit (%d bytes)", maxTotalSize)
		}

		// 识别主要修改器可执行程序
		if strings.HasSuffix(strings.ToLower(cleanTarget), ".exe") {
			if mainExePath == "" || strings.Contains(strings.ToLower(cleanTarget), "trainer") {
				mainExePath = cleanTarget
			}
		}
	}

	return mainExePath, nil
}
